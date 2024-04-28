package main_test

import (
	"rocked/cmd"

	"os"
	"testing"
)

func TestForkSuccess(t *testing.T) {
	pid, err := cmd.Fork(nil)
	if int(err) != 0 {
		t.Fatalf("ForkExec failed with an error (%v)", err)
	}
	if pid < 0 {
		t.Fatalf("ForkExec didn't return a valid pid (%v)", pid)
	}
	p := cmd.Wait(int(pid))
	if p != int(pid) {
		t.Fatalf("Wait failed (returned pid (%v) is different from the passed pid (%v)", p, pid)
	}
}

func TestExecSuccess(t *testing.T) {
	bin := "/usr/bin/echo"
	a := cmd.ExecArgs{
		Exe:     string(bin),
		Exeargs: []string{bin, "test"},
	}
	a.Env = os.Environ()
	err := cmd.Exec(&a)
	if err != 0 {
		t.Fatalf("Error executing %v", a.Exe)
	}
}
