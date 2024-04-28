package main_test

import (
	"fmt"
	"rocked/utils"
	"testing"
)

func TestUntar(t *testing.T) {
	var paths = []string{
		"utils/chroot.tar",
		"utils/Fedora-minimal-chroot.tar",
	}
	for idx, path := range paths {
		dest_path := "/tmp/test_chroot" + fmt.Sprintf("%d", idx)
		t.Run(dest_path, func(t *testing.T) {
			err := utils.ExtractImage(path, dest_path)
			if err != nil {
				t.Fatalf("ExtractImage returned an error (%v)", err)
			}
			if utils.PathExists(dest_path) != true {
				t.Fatalf("Destination directory was not created")
			}
		})

	}
}

func TestUntarNonExistentArchive(t *testing.T) {
	chroot_archive := "utils/chroot2.tar"
	err := utils.ExtractImage(chroot_archive, "/tmp/test_chroot")
	if err == nil {
		t.Fatalf("ExtractImage didn't return an error (%v) when the source archive doesn't exists", err)
	}
}
