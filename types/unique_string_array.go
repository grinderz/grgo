package types

import "strings"

type UniqueStringArray map[string]interface{}

func (a *UniqueStringArray) Set(s string) error {
	(*a)[s] = nil
	return nil
}

func (a *UniqueStringArray) String() string {
	keys := make([]string, len(*a))
	var i uint
	for k := range *a {
		keys[i] = k
		i++
	}
	return strings.Join(keys, "|")
}
