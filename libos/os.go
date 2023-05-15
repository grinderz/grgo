package libos

import (
	"os"
)

func IsExists(path string) (os.FileInfo, bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		return stat, true, nil
	}

	if os.IsNotExist(err) {
		return stat, false, nil
	}

	return stat, false, err //nolint:wrapcheck
}

func IsDirExists(path string) (bool, error) {
	stat, exists, err := IsExists(path)
	if err != nil || !exists {
		return false, err
	}

	if stat.IsDir() {
		return true, nil
	}

	return false, err
}
