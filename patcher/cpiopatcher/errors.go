package cpiopatcher

import "fmt"

type InvalidOffsetsLengthError struct {
	Path          string
	PatternIndex  int
	PatternsCount int
	OffsetsLength int
}

func (e *InvalidOffsetsLengthError) Error() string {
	return fmt.Sprintf(
		"%s: pattern %d invalid offsets length offsets_len[%d] != pattern_count[%d]",
		e.Path,
		e.PatternIndex,
		e.OffsetsLength,
		e.PatternsCount,
	)
}

type PatternNotFoundError struct {
	Path         string
	PatternIndex int
}

func (e *PatternNotFoundError) Error() string {
	return fmt.Sprintf(
		"%s: pattern %d not found",
		e.Path,
		e.PatternIndex,
	)
}
