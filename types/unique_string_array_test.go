package types

import (
	"strings"
	"testing"
)

func TestUniqueStringArray(t *testing.T) {
	keys := [3]string{"t1", "t2", "t3"}

	arr := make(UniqueStringArray)
	for _, k := range keys {
		arr.Set(k)
		arr.Set(k)
	}

	if len(arr) != len(keys) {
		t.Fatalf("arr len non valid: %d != %d", len(arr), len(keys))
	}

	keysStr := strings.Join(keys[:], "|")
	if len(arr.String()) != len(keysStr) {
		t.Fatalf("str len non valid: %d != %d", len(arr.String()), len(keysStr))
	}

}
