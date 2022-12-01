package patch

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestPatch(t *testing.T) {

}

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func checkError(err error) {
	if err != nil {
		fmt.Println("fatal error ", err.Error())
		os.Exit(1)
	}
}
