package patch

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/grinderz/gocpio"
	"github.com/xi2/xz"
	"io"
	"log"
	"os"
	"path/filepath"
)

const kMaxMagicSize = 6
const kCpioZeroFooterLen = 108

type headerType int

const (
	kUnknown headerType = 0
	kCpio    headerType = 1
	kXZ      headerType = 2
	kGZ      headerType = 3
)

var (
	cpioMagic = []byte{
		0x30, 0x37, 0x30, 0x37, 0x30, 0x31,
	}

	xzMagic = []byte{
		0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00,
	}

	gzMagic = []byte{
		0x1F, 0x8B,
	}
)

type Result struct {
	Path  string
	Count int
	Err   error
}

func newResult(p *Patcher, count int) Result {
	return Result{p.path, count, nil}
}

func newError(p *Patcher, err error) Result {
	return Result{p.path, 0, err}
}

type Patcher struct {
	temp   string
	path   string
	name   string
	result chan<- Result
}

func NewPathcer(temp, path string, result chan<- Result) *Patcher {
	name := filepath.Base(path)
	return &Patcher{
		temp:   temp,
		path:   path,
		name:   name,
		result: result,
	}
}

func (p *Patcher) seekCpio(f *os.File) (int64, error) {
	rdr := cpio.NewReader(f)

	var hdr *cpio.Header
	var err error
	for {
		hdr, err = rdr.Next()
		if err != nil {
			return 0, err
		}

		if hdr.Name == "TRAILER!!!" {
			break
		}
	}
	f.Seek(0, 0)
	return rdr.Pos(), nil

}

func (p *Patcher) cutCpio(dst io.Writer, r *os.File) error {
	sizeBuff := make([]byte, 8)
	if _, err := r.ReadAt(sizeBuff, 54); err != nil {
		return err
	}

	if _, err := r.Seek(0, 0); err != nil {
		return err
	}

	i, err := p.seekCpio(r)
	if err != nil {
		return err
	}

	if _, err = io.CopyN(dst, r, i); err != nil {
		return err
	}

	return nil
}

func (p *Patcher) replaceBytes(f *os.File, offsets []int64, replace []byte) (int, error) {
	totalReplaced := 0
	for _, o := range offsets {
		replaced, err := f.WriteAt(replace, o)
		if err != nil {
			return 0, err
		}
		totalReplaced += replaced
	}

	if err := f.Sync(); err != nil {
		return 0, err
	}

	return totalReplaced, nil
}

func (p *Patcher) searchBytes(f io.Reader, find []byte) ([]int64, error) {
	result := make([]int64, 0, 2)

	buff := make([]byte, 8192)
	r := bufio.NewReader(f)
	findLen := len(find)
	totalRead := int64(0)

	var index int
	var n int

	var err error
	for {
		if n, err = r.Read(buff); err != nil && err != io.EOF {
			return nil, err
		}

		for i, b := range buff {
			if b != find[index] {
				index = 0
				continue
			}

			index++
			if index == findLen {
				result = append(result, totalRead-int64(index)+int64(i)+1)
				index = 0
			}
		}

		totalRead += int64(n)
		if err == io.EOF {
			break
		}
	}
	return result, nil
}

func (p *Patcher) getType(r io.Reader) (headerType, error) {
	buff := make([]byte, kMaxMagicSize)
	if _, err := io.ReadFull(r, buff); err != nil {
		return kUnknown, err
	}

	if bytes.Equal(buff, cpioMagic) {
		return kCpio, nil
	}
	if bytes.Equal(buff, xzMagic) {
		return kXZ, nil
	}
	if bytes.Equal(buff[:len(gzMagic)], gzMagic) {
		return kGZ, nil
	}
	return kUnknown, fmt.Errorf("unsupported format %x", buff)
}

func (p *Patcher) unpackXZ(dst io.Writer, r io.Reader) error {
	xzr, err := xz.NewReader(r, 0)
	if err != nil {
		return err
	}

	if _, err = io.Copy(dst, xzr); err != nil {
		return err
	}
	return nil
}

func (p *Patcher) unpackGZ(dst io.Writer, r io.Reader) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()

	if _, err = io.Copy(dst, gz); err != nil {
		return err
	}

	return nil
}

func (p *Patcher) backup(r io.Reader) error {
	backupFile, err := os.Create(fmt.Sprintf("%s.bak", p.path))
	if err != nil {
		return err
	}
	defer backupFile.Close()

	if _, err := io.Copy(backupFile, r); err != nil {
		return err
	}

	if err = backupFile.Sync(); err != nil {
		return err
	}
	return nil
}

func (p *Patcher) writeCpio(dst io.Writer, r io.Reader) error {
	if _, err := io.Copy(dst, r); err != nil {
		return err
	}
	if _, err := dst.Write(make([]byte, kCpioZeroFooterLen)); err != nil {
		return err
	}
	return nil
}

func (p *Patcher) packGZ(dst io.Writer, r io.Reader) error {
	gz := gzip.NewWriter(dst)
	defer gz.Close()

	if _, err := io.Copy(gz, r); err != nil {
		return err
	}

	return nil
}

func (p *Patcher) Patch(find, replace []byte, backup bool) {
	r, err := os.OpenFile(p.path, os.O_RDWR, 0644)
	if err != nil {
		p.result <- newError(p, err)
		return
	}
	defer r.Close()

	t, err := p.getType(r)
	if err != nil {
		p.result <- newError(p, err)
		return
	}

	var cpioFile *os.File

	if t == kCpio {
		log.Printf("%s: cut cpio header", p.path)

		cpioFile, err = os.Create(filepath.Join(p.temp, fmt.Sprintf("%s.cpio", p.name)))
		if err != nil {
			p.result <- newError(p, err)
			return
		}
		defer cpioFile.Close()

		if err := p.cutCpio(cpioFile, r); err != nil {
			p.result <- newError(p, err)
			return
		}

		if _, err = r.Seek(kCpioZeroFooterLen, 1); err != nil {
			p.result <- newError(p, err)
			return
		}

		t, err = p.getType(r)
		if err != nil {
			p.result <- newError(p, err)
			return
		}
	}

	if _, err = r.Seek(-kMaxMagicSize, 1); err != nil {
		p.result <- newError(p, err)
		return
	}

	rawFile, err := os.Create(filepath.Join(p.temp, fmt.Sprintf("%s.raw", p.name)))
	if err != nil {
		p.result <- newError(p, err)
		return
	}
	defer rawFile.Close()

	switch t {
	case kXZ:
		log.Printf("%s: unpack xz", p.path)
		if err := p.unpackXZ(rawFile, r); err != nil {
			p.result <- newError(p, err)
			return
		}
	case kGZ:
		log.Printf("%s: unpack gz", p.path)
		if err := p.unpackGZ(rawFile, r); err != nil {
			p.result <- newError(p, err)
			return
		}
	}

	log.Printf("%s: search", p.path)

	if _, err = rawFile.Seek(0, 0); err != nil {
		p.result <- newError(p, err)
		return
	}

	offsets, err := p.searchBytes(rawFile, find)
	if err != nil {
		p.result <- newError(p, err)
		return
	}

	if len(offsets) == 0 {
		p.result <- newResult(p, 0)
		return
	}

	log.Printf("%s: patch", p.path)

	replaced, err := p.replaceBytes(rawFile, offsets, replace)
	if err != nil {
		p.result <- newError(p, err)
		return
	}

	if backup {
		log.Printf("%s: backup", p.path)
		if _, err = r.Seek(0, 0); err != nil {
			p.result <- newError(p, err)
			return
		}
		if err = p.backup(r); err != nil {
			p.result <- newError(p, err)
			return
		}
	}

	if _, err = rawFile.Seek(0, 0); err != nil {
		p.result <- newError(p, err)
		return
	}

	if _, err = r.Seek(0, 0); err != nil {
		p.result <- newError(p, err)
		return
	}
	if err = r.Truncate(0); err != nil {
		p.result <- newError(p, err)
		return
	}

	if cpioFile != nil {
		if _, err = cpioFile.Seek(0, 0); err != nil {
			p.result <- newError(p, err)
			return
		}
		if err = p.writeCpio(r, cpioFile); err != nil {
			p.result <- newError(p, err)
			return
		}
	}

	log.Printf("%s: pack gz", p.path)
	if err = p.packGZ(r, rawFile); err != nil {
		p.result <- newError(p, err)
		return
	}

	p.result <- newResult(p, replaced)

}
