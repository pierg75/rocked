package cmd_test

import (
	"errors"
	"rocked/cmd"

	"os"
	"testing"
)

func TestCreateOverlayDirsSuccess(t *testing.T) {
	t.Run("CreateOverlayDirs", func(t *testing.T) {
		err := cmd.CreateOverlayDirs("/tmp/")
		if err != nil {
			t.Fatalf("CreateOverlayDirsSuccess failed with an error (%v)", err)
		}
		_, err = os.Stat("/tmp/overlay/work")

		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("CreateOverlayDirsSuccess failed with an error (%v)", err)
		}
	})
}
