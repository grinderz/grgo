package libio

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/xi2/xz"
)

func CloneReader(reader io.Reader, dst string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("clone reader create dst failed: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, reader); err != nil {
		return fmt.Errorf("clone reader copy failed: %w", err)
	}

	if err = dstFile.Sync(); err != nil {
		return fmt.Errorf("clone reader sync dst failed: %w", err)
	}

	return nil
}

func UnpackXZ(dst io.Writer, reader io.Reader) error {
	xzReader, err := xz.NewReader(reader, 0)
	if err != nil {
		return fmt.Errorf("unpack xz reader failed: %w", err)
	}

	if _, err = io.Copy(dst, xzReader); err != nil {
		return fmt.Errorf("unpack xz copy failed: %w", err)
	}

	return nil
}

func UnpackGZ(dst io.Writer, reader io.Reader, maxDecompressBytes int64) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("unpack gz reader failed: %w", err)
	}

	defer gzReader.Close()

	written, err := io.CopyN(dst, gzReader, maxDecompressBytes)
	if err != nil {
		return fmt.Errorf("unpack gz copy failed: %w", err)
	}

	if written == maxDecompressBytes {
		return ErrUnpackMaxDecompressLimitReached
	}

	return nil
}

func PackGZ(dst io.Writer, reader io.Reader) error {
	gzWriter := gzip.NewWriter(dst)
	defer gzWriter.Close()

	if _, err := io.Copy(gzWriter, reader); err != nil {
		return fmt.Errorf("pack gz copy failed: %w", err)
	}

	return nil
}
