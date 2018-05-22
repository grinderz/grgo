package types

import (
	"strings"
)

type UniqueStringArray map[string]interface{}

func (a *UniqueStringArray) Set(s string) error {
	(*a)[s] = nil
	return nil
}

func (a *UniqueStringArray) String() string {
	keys := make([]string, 0, len(*a))
	for k := range *a {
		keys = append(keys, k)
	}
	return strings.Join(keys, "|")
}
