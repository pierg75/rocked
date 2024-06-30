package cmd_test

import (
	"errors"
	"rocked/cmd"

	"os"
	"strings"
	"testing"
)

func TestWriteMap(t *testing.T) {
	t.Run("WriteMap", func(t *testing.T) {
		f := "/tmp/testmap"
		m := cmd.IDmapping{
			InsideID:  0,
			OutsideID: 15603,
			Len:       1,
		}
		expected := "0 15603 1"
		err := cmd.WriteMaps(f, m)
		_, err = os.Stat(f)
		if errors.Is(err, os.ErrNotExist) {
			t.Fatalf("WriteMap failed to create %v (%v)", f, err)
		}
		mapfile, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("Could not read file %v", mapfile)
		}
		if strings.TrimSuffix(string(mapfile), "\n") != expected {
			t.Errorf("got %v want %v", string(mapfile), expected)
		}
	})
}
