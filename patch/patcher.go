package patch

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/grinderz/gocpio"
	"github.com/grinderz/grgo/logging"
	"github.com/xi2/xz"
)

const kMaxMagicSize = 6

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

type Pattern struct {
	Description string
	Count       int
	Search      []byte
	Replace     []byte
}

type Result struct {
	Path         string
	BytesPatched int
	Err          error
}

func newResult(p *Patcher, bytesPatched int) Result {
	return Result{p.path, bytesPatched, nil}
}

func newError(p *Patcher, err error) Result {
	return Result{p.path, 0, err}
}

type Patcher struct {
	temp              string
	path              string
	name              string
	cpioZeroFooterLen int64
	result            chan<- Result
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

func (p *Patcher) findCpioZeroFooterLen(f *os.File) (int64, error) {
	buff := make([]byte, 8192)
	totalRead := int64(0)

	var index int64
	var n int

	var err error
	for {
		if n, err = f.Read(buff); err != nil && err != io.EOF {
			return 0, err
		}
		totalRead += int64(n)

		for _, b := range buff {
			if b != 0x00 {
				if _, err := f.Seek(-totalRead+index, 1); err != nil {
					return 0, err
				}
				return index, nil
			}
			index++
		}

		if err == io.EOF {
			return 0, errors.New("EOF detected")
		}
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

	if _, err := f.Seek(0, 0); err != nil {
		return 0, err
	}
	return rdr.Pos(), nil
}

func (p *Patcher) cutCpio(dst io.Writer, src *os.File) error {
	if _, err := src.Seek(0, 0); err != nil {
		return err
	}
	i, err := p.seekCpio(src)
	if err != nil {
		return err
	}
	if _, err = io.CopyN(dst, src, i); err != nil {
		return err
	}
	return nil
}

func (p *Patcher) replaceBytes(f *os.File, offsets []int64, replace []byte) (int, error) {
	totalReplaced := 0
	for _, offset := range offsets {
		replaced, err := f.WriteAt(replace, offset)
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
	reader := bufio.NewReader(f)
	findLen := len(find)
	totalRead := int64(0)

	var index int
	var readCounter int

	var err error
	for {
		if readCounter, err = reader.Read(buff); err != nil && err != io.EOF {
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

		totalRead += int64(readCounter)
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

func (p *Patcher) unpackXZ(dst io.Writer, reader io.Reader) error {
	xzReader, err := xz.NewReader(reader, 0)
	if err != nil {
		return err
	}
	if _, err = io.Copy(dst, xzReader); err != nil {
		return err
	}
	return nil
}

func (p *Patcher) unpackGZ(dst io.Writer, reader io.Reader) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer func(gzReader *gzip.Reader) {
		if err := gzReader.Close(); err != nil {
			logging.Log.Errorln(err)
		}
	}(gzReader)

	if _, err = io.Copy(dst, gzReader); err != nil {
		return err
	}
	return nil
}

func (p *Patcher) backup(reader io.Reader) error {
	backupFile, err := os.Create(fmt.Sprintf("%s.bak", p.path))
	if err != nil {
		return err
	}
	defer func(backupFile *os.File) {
		if err := backupFile.Close(); err != nil {
			logging.Log.Errorln(err)
		}
	}(backupFile)

	if _, err := io.Copy(backupFile, reader); err != nil {
		return err
	}
	if err = backupFile.Sync(); err != nil {
		return err
	}
	return nil
}

func (p *Patcher) writeCpio(dst io.Writer, reader io.Reader) error {
	if _, err := io.Copy(dst, reader); err != nil {
		return err
	}
	if _, err := dst.Write(make([]byte, p.cpioZeroFooterLen)); err != nil {
		return err
	}
	return nil
}

func (p *Patcher) packGZ(dst io.Writer, reader io.Reader) error {
	gzWriter := gzip.NewWriter(dst)
	defer func(gzWriter *gzip.Writer) {
		if err := gzWriter.Close(); err != nil {
			logging.Log.Errorln(err)
		}
	}(gzWriter)

	if _, err := io.Copy(gzWriter, reader); err != nil {
		return err
	}
	return nil
}

func (p *Patcher) Patch(patterns []Pattern, backup bool) {
	inFile, err := os.OpenFile(p.path, os.O_RDWR, 0644)
	if err != nil {
		p.result <- newError(p, err)
		return
	}
	defer func(inFile *os.File) {
		if err := inFile.Close(); err != nil {
			logging.Log.Errorln(err)
		}
	}(inFile)

	fileType, err := p.getType(inFile)
	if err != nil {
		p.result <- newError(p, err)
		return
	}

	var cpioFile *os.File
	if fileType == kCpio {
		logging.Log.Printf("%s: cut cpio header", p.path)

		cpioFile, err = os.Create(filepath.Join(p.temp, fmt.Sprintf("%s.cpio", p.name)))
		if err != nil {
			p.result <- newError(p, err)
			return
		}
		defer func(cpioFile *os.File) {
			if err := cpioFile.Close(); err != nil {
				logging.Log.Errorln(err)
			}
		}(cpioFile)

		if err := p.cutCpio(cpioFile, inFile); err != nil {
			p.result <- newError(p, err)
			return
		}

		p.cpioZeroFooterLen, err = p.findCpioZeroFooterLen(inFile)
		if err != nil {
			p.result <- newError(p, err)
			return
		}

		fileType, err = p.getType(inFile)
		if err != nil {
			p.result <- newError(p, err)
			return
		}
	}

	if _, err = inFile.Seek(-kMaxMagicSize, 1); err != nil {
		p.result <- newError(p, err)
		return
	}

	rawFile, err := os.Create(filepath.Join(p.temp, fmt.Sprintf("%s.raw", p.name)))
	if err != nil {
		p.result <- newError(p, err)
		return
	}
	defer func(rawFile *os.File) {
		if err := rawFile.Close(); err != nil {
			logging.Log.Errorln(err)
		}
	}(rawFile)

	switch fileType {
	case kXZ:
		logging.Log.Printf("%s: unpack xz", p.path)
		if err := p.unpackXZ(rawFile, inFile); err != nil {
			p.result <- newError(p, err)
			return
		}
	case kGZ:
		logging.Log.Printf("%s: unpack gz", p.path)
		if err := p.unpackGZ(rawFile, inFile); err != nil {
			p.result <- newError(p, err)
			return
		}
	}

	replaced := 0
	for patternIndex, pattern := range patterns {
		logging.Log.Printf("%s: search %d [%s]", p.path, patternIndex, pattern.Description)

		if _, err = rawFile.Seek(0, 0); err != nil {
			p.result <- newError(p, err)
			return
		}

		offsets, err := p.searchBytes(rawFile, pattern.Search)
		if err != nil {
			p.result <- newError(p, err)
			return
		}
		if len(offsets) == 0 {
			p.result <- newError(p, fmt.Errorf("%s: pattern %d not found", p.path, patternIndex))
			return
		}

		if len(offsets) != pattern.Count {
			p.result <- newError(p, fmt.Errorf(
				"%s: pattern %d invalid offset count offsets_len[%d] != pattern_count[%d]",
				p.path,
				patternIndex,
				len(offsets),
				pattern.Count,
			),
			)
			return
		}
		logging.Log.Printf("%s: patch %d", p.path, patternIndex)

		r, err := p.replaceBytes(rawFile, offsets, pattern.Replace)
		if err != nil {
			p.result <- newError(p, err)
			return
		}

		replaced += r
	}

	if replaced == 0 {
		p.result <- newResult(p, 0)
		return
	}

	if backup {
		logging.Log.Printf("%s: backup", p.path)
		if _, err = inFile.Seek(0, 0); err != nil {
			p.result <- newError(p, err)
			return
		}
		if err = p.backup(inFile); err != nil {
			p.result <- newError(p, err)
			return
		}
	}

	if _, err = rawFile.Seek(0, 0); err != nil {
		p.result <- newError(p, err)
		return
	}

	if _, err = inFile.Seek(0, 0); err != nil {
		p.result <- newError(p, err)
		return
	}
	if err = inFile.Truncate(0); err != nil {
		p.result <- newError(p, err)
		return
	}

	if cpioFile != nil {
		if _, err = cpioFile.Seek(0, 0); err != nil {
			p.result <- newError(p, err)
			return
		}
		if err = p.writeCpio(inFile, cpioFile); err != nil {
			p.result <- newError(p, err)
			return
		}
	}

	logging.Log.Printf("%s: pack gz", p.path)
	if err = p.packGZ(inFile, rawFile); err != nil {
		p.result <- newError(p, err)
		return
	}

	p.result <- newResult(p, replaced)
}
