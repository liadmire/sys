package sys

import (
	"os"
	"path/filepath"
	"strings"
)

// SelfPath gets compiled executable file absolute path
func SelfPath() string {
	path, _ := filepath.Abs(os.Args[0])
	return path
}

// SelfDir gets compiled executable file directory
func SelfDir() string {
	return filepath.Dir(SelfPath())
}

// SelfName gets compiled executable name
func SelfName() string {
	return filepath.Base(SelfPath())
}

// SelfNameWithoutExt gets compiled executable name with ext.
func SelfNameWithoutExt() string {
	return strings.TrimSuffix(SelfName(), SelfExt())
}

// SelfExt  gets compiled executable suffix
func SelfExt() string {
	return filepath.Ext(SelfPath())
}

// FileExists reports whether the named file or directory exists.
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func FileSize(name string) int64 {
	fileInfo, err := os.Stat(name)
	if err != nil {
		return -1
	}
	if os.IsNotExist(err) {
		return -1
	}

	return fileInfo.Size()
}
