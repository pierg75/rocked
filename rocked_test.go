package main_test

import (
	"rocked/cmd"

	"syscall"
	"testing"
)

func TestForkSuccess(t *testing.T) {
	bin := "/usr/bin/echo"
	a := cmd.ForkExecArgs{
		Exe:     string(bin),
		Exeargs: []string{bin, "test"},
	}
	a.Cloneflags = syscall.CLONE_VFORK | syscall.CLONE_FILES
	pid, err := cmd.ForkExec(&a)
	cmd.Wait(int(pid))
	if int(err) != 0 {
		t.Fatalf("ForkExec failed with an error (%v)", err)
	}
	if pid < 0 {
		t.Fatalf("ForkExec didn't return a valid pid (%v)", pid)
	}
}

/*
func TestForkFailureBin(t *testing.T) {
	bin := "/usr/bin/balh"
	a := cmd.ForkExecArgs{
		Exe:     string(bin),
		Exeargs: []string{bin, "test"},
	}
	a.Cloneflags = syscall.CLONE_VFORK | syscall.CLONE_FILES
	pid, err := cmd.ForkExec(&a)
	if int(err) == 0 {
		t.Fatalf("ForkExec should have failed with an error, got (%v)", err)
	}
	if pid >= 0 {
		t.Fatalf("ForkExec shouldn't return a valid pid, got (%v)", pid)
	}
	cmd.Wait(int(pid))
}
*/
