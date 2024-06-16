package cmd_test

import (
	"os"
	"rocked/cmd"
	"strings"

	"testing"
)

func TestNewCgroup(t *testing.T) {
	t.Run("NewCgroupSuccess", func(t *testing.T) {
		id := "6e220a98-c915-4b21-9898-4db49208f6ff"
		cg := cmd.NewCgroup(id)
		gotPath := cg.Path()
		wantPath := "/sys/fs/cgroup/rocked/"
		if gotPath != wantPath {
			t.Errorf("got %v want %v", gotPath, wantPath)
		}
		gotFullPath := cg.CgroupConPath
		wantFulltPath := "/sys/fs/cgroup/rocked/" + id
		if gotFullPath != wantFulltPath {
			t.Errorf("got %v want %v", gotPath, wantPath)
		}
	})
}

func TestCgroupController(t *testing.T) {
	t.Run("NewCgroupSuccess", func(t *testing.T) {
		id := "6e220a98-c915-4b21-9898-4db49208f6ff"
		controllerFile := "/sys/fs/cgroup/rocked/cgroup.subtree_control"
		expected := "cpu io memory pids"
		cg := cmd.NewCgroup(id)
		cg.SetControllers()
		controllerContent, err := os.ReadFile(controllerFile)
		if err != nil {
			t.Errorf("Could not read file %v", controllerFile)
		}
		if strings.TrimSuffix(string(controllerContent), "\n") != expected {
			t.Errorf("got %v want %v", string(controllerContent), expected)
		}
	})
}
