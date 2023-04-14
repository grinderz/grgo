package libmap

import "strings"

type UniqueStringArray map[string]interface{}

func (a UniqueStringArray) Set(s string) error {
	(a)[s] = nil
	return nil
}

func (a UniqueStringArray) String() string {
	return strings.Join(MapKeysAsStrings(a), "|")
}
