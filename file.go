package sys

import (
	"os"
	"path/filepath"
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
