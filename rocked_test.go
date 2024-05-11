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

func TestExecArgsNil(t *testing.T) {
	err := cmd.Exec(nil)
	if err != 0 {
		t.Fatalf("Error executing with an empty ExecArgs")
	}
}

func TestChroot(t *testing.T) {
	dir := t.TempDir()
	err := cmd.Chroot(dir)
	if err != 0 {
		t.Fatalf("Error chrooting on the directory %v: %v", dir, err)
	}
}

func TestChrootMount(t *testing.T) {
	dir := t.TempDir()
	target := dir + "/proc"
	os.Mkdir(dir+"/proc", 777)
	t.Cleanup(func() {
		cmd.Umount(dir+"/proc", 0)
	})
	err := cmd.Mount("proc", target, "proc", 0)
	if err != 0 {
		t.Fatalf("Error mounting proc on the directory %v: %v", target, err)
	}
	err = cmd.Chroot(dir)
	if err != 0 {
		t.Fatalf("Error chrooting on the directory %v: %v", dir, err)
	}
}

func TestMountUmountProc(t *testing.T) {
	dir := t.TempDir()
	target := dir + "/proc"
	t.Cleanup(func() {
		cmd.Umount(dir+"/proc", 0)
	})
	os.Mkdir(dir+"/proc", 777)
	err := cmd.Mount("proc", dir+"/proc", "proc", 0)
	if err != 0 {
		t.Fatalf("Error mounting proc on the directory %v: %v", target, err)
	}
	err = cmd.Umount(target, 0)
	if err != 0 {
		t.Fatalf("Error umounting on the directory %v: %v", dir, err)
	}
}

func TestUnshareMount(t *testing.T) {
	err := cmd.Unshare(cmd.CLONE_NEWNS)
	if err != 0 {
		t.Fatalf("Error unsharing mount namespace %v", err)
	}
	bin := "/usr/bin/df"
	a := cmd.ExecArgs{
		Exe: string(bin),
	}
	a.Env = os.Environ()
	err = cmd.Exec(&a)
	if err != 0 {
		t.Fatalf("Error executing %v", a.Exe)
	}
}
