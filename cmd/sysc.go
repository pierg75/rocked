/*
Copyright © 2024 Pierguido Lambri <plambri@redhat.com>
*/
package cmd

import (
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
)

type cloneArgs struct {
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

type ForkExecArgs struct {
	cloneargs  *cloneArgs
	cloneflags int
	exe        string
	exeargs    []string
	env        []string
}

func ForkExec(args *ForkExecArgs) (p uintptr, err syscall.Errno) {
	if args.cloneflags == 0 {
		args.cloneflags = syscall.CLONE_VFORK | syscall.CLONE_FILES
	}
	if args.cloneargs == nil {
		args.cloneargs = &cloneArgs{
			flags:      uint64(args.cloneflags),
			exitSignal: uint64(syscall.SIGCHLD),
		}
	}
	if args.env == nil {
		args.env = os.Environ()
	}
	pid, _, error := syscall.RawSyscall(CLONE3, uintptr(unsafe.Pointer(args.cloneargs)), unsafe.Sizeof(*args.cloneargs), 0)
	if error != 0 {
		slog.Debug("Error cloning: %v", error)
		return pid, err
	}
	if int(pid) == 0 {
		slog.Debug("Child", "pid", pid, "pid thread", os.Getpid())
		slog.Debug("Child", "exec", args.exe, "options", args.exeargs)
		arg0p, e := syscall.BytePtrFromString(args.exe)
		if e != nil {
			log.Fatal("Error converting process name to pointer")
		}
		arglp, e := syscall.SlicePtrFromStrings(args.exeargs)
		if e != nil {
			log.Fatal("Error converting arguments to pointer")
		}
		envlp, e := syscall.SlicePtrFromStrings(args.env)
		if e != nil {
			log.Fatal("Error converting arguments to pointer")
		}
		_, _, error = syscall.RawSyscall(EXECVE, uintptr(unsafe.Pointer(arg0p)), uintptr(unsafe.Pointer(&arglp[0])), uintptr(unsafe.Pointer(&envlp[0])))
		if error != 0 {
			log.Fatal("** Error calling execv")
		}
		return pid, syscall.Errno(int(error))
	} else {
		slog.Debug("Parent", "child pid", pid, "pid thread", os.Getpid())
		return pid, 0
	}
}

func Wait(pid int) (status int) {
	var s int
	_, _, err := syscall.RawSyscall(WAIT4, uintptr(pid), uintptr(s), 0)
	if err != 0 {
		slog.Debug("wait4", "error", err)
	}
	return s
}
