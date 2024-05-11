package utils

import (
	"os"
)

var (
	VIRTFS = []string{"proc", "sys", "devtmpfs"}
)

// Checks if a path (either file or directory) exists
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			return true
		}
	}
	return true
}

// Checks if a fstype is a virtual file system.
func IsVirtual(fs string) bool {
	for _, fstype := range VIRTFS {
		if fstype == fs {
			return true
		}
	}
	return false
}
