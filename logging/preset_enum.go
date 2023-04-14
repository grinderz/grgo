package logging

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=PresetEnum -linecomment -output preset_enum_string.go
type PresetEnum int

const (
	PresetUnknown     PresetEnum = iota // unknown
	PresetDevelopment PresetEnum = iota // development
	PresetProduction  PresetEnum = iota // production
)

func (e *PresetEnum) SetValue(value string) error {
	preset := PresetFromString(value)
	if preset == PresetUnknown {
		return &PresetValueError{
			Value: value,
		}
	}

	*e = preset

	return nil
}

func (e PresetEnum) MarshalText() ([]byte, error) {
	if e == PresetUnknown {
		return nil, &PresetValueError{
			Value: PresetUnknown.String(),
		}
	}

	return []byte(e.String()), nil
}

func (e *PresetEnum) UnmarshalText(text []byte) error {
	return e.SetValue(string(text))
}

func PresetFromString(value string) PresetEnum {
	switch strings.ToLower(value) {
	case "development":
		return PresetDevelopment
	case "production":
		return PresetProduction
	default:
		return PresetUnknown
	}
}

type PresetValueError struct {
	Value string
}

func (e *PresetValueError) Error() string {
	return fmt.Sprintf("preset invalid value: %s", e.Value)
}
