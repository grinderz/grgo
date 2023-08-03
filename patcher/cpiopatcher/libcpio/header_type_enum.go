package libcpio

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

const MaxMagicSize = 6

var (
	cpioMagic = []byte{ //nolint:gochecknoglobals
		0x30, 0x37, 0x30, 0x37, 0x30, 0x31,
	}

	xzMagic = []byte{ //nolint:gochecknoglobals
		0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00,
	}

	gzMagic = []byte{ //nolint:gochecknoglobals
		0x1F, 0x8B,
	}
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=HeaderTypeEnum -linecomment -output header_type_enum_string.go
type HeaderTypeEnum int

const (
	HeaderTypeUnknown HeaderTypeEnum = iota // unknown
	HeaderTypeCPIO    HeaderTypeEnum = iota // cpio
	HeaderTypeXZ      HeaderTypeEnum = iota // xz
	HeaderTypeGZ      HeaderTypeEnum = iota // gz
)

func (ht *HeaderTypeEnum) SetValue(value string) error {
	headerType := HeaderTypeFromString(value)
	if headerType == HeaderTypeUnknown {
		return &HeaderTypeValueError{
			Value: value,
		}
	}

	*ht = headerType

	return nil
}

func (ht HeaderTypeEnum) MarshalText() ([]byte, error) {
	if ht == HeaderTypeUnknown {
		return nil, &HeaderTypeValueError{
			Value: HeaderTypeUnknown.String(),
		}
	}

	return []byte(ht.String()), nil
}

func (ht *HeaderTypeEnum) UnmarshalText(text []byte) error {
	return ht.SetValue(string(text))
}

func HeaderTypeFromString(value string) HeaderTypeEnum {
	switch strings.ToLower(value) {
	case "cpio":
		return HeaderTypeCPIO
	case "xz":
		return HeaderTypeXZ
	case "gz":
		return HeaderTypeGZ
	default:
		return HeaderTypeUnknown
	}
}

func HeaderTypeFromReader(r io.Reader) (HeaderTypeEnum, error) {
	buff := make([]byte, MaxMagicSize)
	if _, err := io.ReadFull(r, buff); err != nil {
		return HeaderTypeUnknown, fmt.Errorf("read reader failed: %w", err)
	}

	if bytes.Equal(buff, cpioMagic) {
		return HeaderTypeCPIO, nil
	}

	if bytes.Equal(buff, xzMagic) {
		return HeaderTypeXZ, nil
	}

	if bytes.Equal(buff[:len(gzMagic)], gzMagic) {
		return HeaderTypeGZ, nil
	}

	return HeaderTypeUnknown, &HeaderTypeUnsupportedFormatError{
		Format: buff,
	}
}

type HeaderTypeValueError struct {
	Value string
}

func (e *HeaderTypeValueError) Error() string {
	return fmt.Sprintf("cpio header type invalid value: %s", e.Value)
}

type HeaderTypeUnsupportedFormatError struct {
	Format []byte
}

func (e *HeaderTypeUnsupportedFormatError) Error() string {
	return fmt.Sprintf("cpio header unsupported format %x", e.Format)
}
