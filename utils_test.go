package main_test

import (
	"rocked/utils"
	"testing"
)

func TestUntar(t *testing.T) {
	chroot_archive := "utils/chroot.tar"
	err := utils.ExtractImage(chroot_archive, "/tmp/test_chroot")
	if err != nil {
		t.Fatalf("ExtractImage returned an error (%v)", err)
	}
}

func TestUntarNonExistentArchive(t *testing.T) {
	chroot_archive := "utils/chroot2.tar"
	err := utils.ExtractImage(chroot_archive, "/tmp/test_chroot")
	if err == nil {
		t.Fatalf("ExtractImage didn't return an error (%v) when the source archive doesn't exists", err)
	}
}

func TestUntarNonExistentDest(t *testing.T) {
	chroot_archive := "utils/chroot.tar"
	err := utils.ExtractImage(chroot_archive, "/tmp2/test_chroot")
	if err == nil {
		t.Fatalf("ExtractImage didn't return an error (%v) when the destination doesn't exists", err)
	}
}
