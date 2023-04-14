package cpiopatcher

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/grinderz/grgo/libio"
	"github.com/grinderz/grgo/patcher"
	"github.com/grinderz/grgo/patcher/cpiopatcher/libcpio"
)

const (
	bufferSize         = 8192
	filePerm           = 0644
	maxDecompressBytes = 524_288_000
)

type Patcher struct {
	tempDir            string
	path               string
	fileName           string
	cpioZeroFooterSize int64
	result             chan<- patcher.Result
	logger             *zap.Logger
}

func New(temp, path string, result chan<- patcher.Result, logger *zap.Logger) *Patcher {
	return &Patcher{
		tempDir:  temp,
		path:     path,
		fileName: filepath.Base(path),
		result:   result,
		logger:   logger,
	}
}

func (p *Patcher) Patch(patterns []*patcher.Pattern, backup bool) {
	var inFile, cpioFile, rawFile *os.File

	inFile, err := os.OpenFile(p.path, os.O_RDWR, filePerm)
	if err != nil {
		p.result <- patcher.NewError(p.path, err)
		return
	}

	defer inFile.Close()

	fileType, err := libcpio.HeaderTypeFromReader(inFile)
	if err != nil {
		p.result <- patcher.NewError(p.path, err)
		return
	}

	if fileType == libcpio.HeaderTypeCPIO {
		p.logger.Info(fmt.Sprintf("%s: cut cpio header", p.path))

		cpioFile, err := os.Create(filepath.Join(p.tempDir, fmt.Sprintf("%s.cpio", p.fileName)))
		if err != nil {
			p.result <- patcher.NewError(p.path, fmt.Errorf("create cpio file failed: %w", err))
			return
		}

		defer cpioFile.Close()

		if fileType, p.cpioZeroFooterSize, err = libcpio.CutHeader(inFile, cpioFile, bufferSize); err != nil {
			p.result <- patcher.NewError(p.path, err)
			return
		}
	}

	if rawFile, err = os.Create(filepath.Join(p.tempDir, fmt.Sprintf("%s.raw", p.fileName))); err != nil {
		p.result <- patcher.NewError(p.path, err)
		return
	}

	defer rawFile.Close()

	if err := p.unpack(rawFile, inFile, fileType); err != nil {
		p.result <- patcher.NewError(p.path, err)
		return
	}

	replaced, err := p.patch(rawFile, patterns)
	if err != nil {
		p.result <- patcher.NewError(p.path, err)
		return
	}

	if replaced == 0 {
		p.result <- patcher.NewResult(p.path, 0)
		return
	}

	if err := p.pack(rawFile, inFile, cpioFile, backup); err != nil {
		p.result <- patcher.NewError(p.path, err)
		return
	}

	p.result <- patcher.NewResult(p.path, replaced)
}

func (p *Patcher) backup(inFile *os.File) error {
	p.logger.Info(fmt.Sprintf("%s: backup", p.path))

	if _, err := inFile.Seek(0, 0); err != nil {
		return fmt.Errorf("file seek failed: %w", err)
	}

	return libio.CloneReader(inFile, fmt.Sprintf("%s.bak", p.path))
}

func (p *Patcher) unpack(rawFile, inFile *os.File, fileType libcpio.HeaderTypeEnum) error {
	if _, err := inFile.Seek(-libcpio.MaxMagicSize, 1); err != nil {
		return fmt.Errorf("in file seek failed: %w", err)
	}

	switch fileType {
	case libcpio.HeaderTypeXZ:
		p.logger.Info(fmt.Sprintf("%s: unpack xz", p.path))

		if err := libio.UnpackXZ(rawFile, inFile); err != nil {
			return err
		}
	case libcpio.HeaderTypeGZ:
		p.logger.Info(fmt.Sprintf("%s: unpack gz", p.path))

		if err := libio.UnpackGZ(rawFile, inFile, maxDecompressBytes); err != nil {
			return err
		}
	case libcpio.HeaderTypeCPIO, libcpio.HeaderTypeUnknown:
		return &libcpio.HeaderTypeValueError{
			Value: fileType.String(),
		}
	}

	return nil
}

func (p *Patcher) patch(rawFile *os.File, patterns []*patcher.Pattern) (int, error) {
	var replaced int

	for patternIndex, pattern := range patterns {
		p.logger.Info(fmt.Sprintf("%s: search %d [%s]", p.path, patternIndex, pattern.Description))

		if _, err := rawFile.Seek(0, 0); err != nil {
			return 0, fmt.Errorf("raw seek failed: %w", err)
		}

		offsets, err := patcher.SearchBytes(rawFile, pattern.Search, bufferSize, pattern.Count)
		if err != nil {
			return 0, err
		}

		if len(offsets) == 0 {
			return 0, &PatternNotFoundError{
				Path:         p.path,
				PatternIndex: patternIndex,
			}
		}

		if len(offsets) != pattern.Count {
			return 0, &InvalidOffsetsLengthError{
				Path:          p.path,
				PatternIndex:  patternIndex,
				PatternsCount: pattern.Count,
				OffsetsLength: len(offsets),
			}
		}

		p.logger.Info(fmt.Sprintf("%s: patch %d", p.path, patternIndex))

		rbs, err := patcher.ReplaceBytes(rawFile, offsets, pattern.Replace)
		if err != nil {
			return 0, err
		}

		replaced += rbs
	}

	return replaced, nil
}

func (p *Patcher) pack(rawFile, inFile, cpioFile *os.File, backup bool) error {
	if backup {
		if err := p.backup(inFile); err != nil {
			return err
		}
	}

	if _, err := rawFile.Seek(0, 0); err != nil {
		return fmt.Errorf("raw file seek failed: %w", err)
	}

	if _, err := inFile.Seek(0, 0); err != nil {
		return fmt.Errorf("in file seek failed: %w", err)
	}

	if err := inFile.Truncate(0); err != nil {
		return fmt.Errorf("in file truncate failed: %w", err)
	}

	if cpioFile != nil {
		if _, err := cpioFile.Seek(0, 0); err != nil {
			return fmt.Errorf("cpio file seek failed: %w", err)
		}

		if err := libcpio.WriteHeader(inFile, cpioFile, p.cpioZeroFooterSize); err != nil {
			return err
		}
	}

	p.logger.Info(fmt.Sprintf("%s: pack gz", p.path))

	return libio.PackGZ(inFile, rawFile)
}
