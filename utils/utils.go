package utils

import (
	"os"
)

var (
	VIRTFS = []string{"proc", "sys"}
)

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

func IsVirtual(fs string) bool {
	for _, fstype := range VIRTFS {
		if fstype == fs {
			return true
		}
	}
	return false
}
