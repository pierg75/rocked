/*
Copyright © 2024 Pierguido Lambri <plambri@redhat.com>
*/
package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"syscall"
	"unsafe"
)

var (
	EXECVE uintptr = 59
	WAIT4  uintptr = 61
	CLONE3 uintptr = 435
	CHROOT uintptr = 161
)

type CloneArgs struct {
	flags      uint64 // Flags bit mask
	pidFD      uint64 // Where to store PID file descriptor (int *)
	childTID   uint64 // Where to store child TID, in child's memory (pid_t *)
	parentTID  uint64 // Where to store child TID, in parent's memory (pid_t *)
	exitSignal uint64 // Signal to deliver to parent on child termination
	stack      uint64 // Pointer to lowest byte of stack
	stackSize  uint64 // Size of stack
	tls        uint64 // Location of new TLS
	setTID     uint64 // Pointer to a pid_t array (since Linux 5.5)
	setTIDSize uint64 // Number of elements in set_tid (since Linux 5.5)
	cgroup     uint64 // File descriptor for target cgroup of child (since Linux 5.7)
}

type ExecArgs struct {
	Exe     string
	Exeargs []string
	Env     []string
}

func Fork(args *CloneArgs) (p uintptr, err syscall.Errno) {
	if args == nil {
		args = &CloneArgs{}
	}
	if args.flags == 0 {
		args.flags = syscall.CLONE_VFORK | syscall.CLONE_FILES
	}
	if args.exitSignal == 0 {
		args.exitSignal = uint64(syscall.SIGCHLD)
	}
	pid, _, error := syscall.RawSyscall(CLONE3, uintptr(unsafe.Pointer(args)), unsafe.Sizeof(*args), 0)
	if error != 0 {
		fmt.Printf("Error cloning: %v\n", error)
		return pid, err
	}
	return pid, error
}

func Exec(args *ExecArgs) (err syscall.Errno) {
	if args.Env == nil {
		args.Env = os.Environ()
	}
	arg0p, e := syscall.BytePtrFromString(args.Exe)
	if e != nil {
		log.Fatal("Error converting process name to pointer")
	}
	arglp, e := syscall.SlicePtrFromStrings(args.Exeargs)
	if e != nil {
		log.Fatal("Error converting arguments to pointer")
	}
	envlp, e := syscall.SlicePtrFromStrings(args.Env)
	if e != nil {
		log.Fatal("Error converting arguments to pointer")
	}
	_, _, error := syscall.RawSyscall(EXECVE, uintptr(unsafe.Pointer(arg0p)), uintptr(unsafe.Pointer(&arglp[0])), uintptr(unsafe.Pointer(&envlp[0])))
	if error != 0 {
		return error
	}
	return syscall.Errno(0)
}

func Wait(pid int) (status int) {
	var s int
	_, _, err := syscall.RawSyscall(WAIT4, uintptr(pid), uintptr(s), 0)
	if err != 0 {
		slog.Debug("wait4", "error", err)
	}
	return s
}

func Chroot(path string) (err syscall.Errno) {

	pathp, e := syscall.BytePtrFromString(path)
	if e != nil {
		log.Fatal("Error converting process name to pointer")
	}
	_, _, error := syscall.RawSyscall(CHROOT, uintptr(unsafe.Pointer(pathp)), 0, 0)

	return error
}
