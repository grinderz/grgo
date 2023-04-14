package logging

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=OutputEnum -linecomment -output output_enum_string.go
type OutputEnum int

const (
	OutputUnknown OutputEnum = iota // unknown
	OutputStdout  OutputEnum = iota // stdout
	OutputStderr  OutputEnum = iota // stderr
	OutputFile    OutputEnum = iota // file
)

func (e *OutputEnum) SetValue(value string) error {
	output := OutputFromString(value)
	if output == OutputUnknown {
		return &OutputValueError{
			Value: value,
		}
	}

	*e = output

	return nil
}

func (e OutputEnum) MarshalText() ([]byte, error) {
	if e == OutputUnknown {
		return nil, &OutputValueError{
			Value: OutputUnknown.String(),
		}
	}

	return []byte(e.String()), nil
}

func (e *OutputEnum) UnmarshalText(text []byte) error {
	return e.SetValue(string(text))
}

func OutputFromString(value string) OutputEnum {
	switch strings.ToLower(value) {
	case "stdout":
		return OutputStdout
	case "stderr":
		return OutputStderr
	case "file":
		return OutputFile
	default:
		return OutputUnknown
	}
}

type OutputValueError struct {
	Value string
}

func (e *OutputValueError) Error() string {
	return fmt.Sprintf("output invalid value: %s", e.Value)
}
