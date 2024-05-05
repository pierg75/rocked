/*
Copyright © 2024 Pierguido Lambri <plambri@redhat.com>
*/
package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"rocked/utils"
	"syscall"
	"unsafe"
)

var (
	EXECVE uintptr = 59
	WAIT4  uintptr = 61
	CHDIR  uintptr = 80
	CLONE3 uintptr = 435
	CHROOT uintptr = 161
	MOUNT  uintptr = 165
	UMOUNT uintptr = 166
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

// Fork executes the clone3 syscalls.
// It accepts a struct Cloneargs, which mirrors the clone_args struct (see man 2 clone)
// If no CloneArgs is passed, it will use default flags (CLONE_VFORK | CLONE_FILES)
// and default exit signal (SIGCHLD).
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

// Exec executes the execve syscalls.
// It accepts a struct ExecArgs, which takes the executable, the arguments and the environment.
// If no environment is passed, it will copy the current OS env
// and default exit signal (SIGCHLD).
func Exec(args *ExecArgs) (err syscall.Errno) {
	if args == nil {
		log.Fatalf("ExecArgs is nil")
	}
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

// Wait waits for the pid to change state.
// It takes the PID of the process to wait on and returns the pid of the process.
func Wait(pid int) (status int) {
	var wstatus uintptr
	p, _, err := syscall.RawSyscall(WAIT4, uintptr(pid), wstatus, 0)
	slog.Debug("wait4 returns", "status", wstatus, "pid", p, "err", err)
	if err != 0 {
		slog.Debug("wait4", "error", err)
	}
	return int(pid)
}

// Chroot changes the root directory of the process to the specified path.
// It returns an error or zero if all is fine.
func Chroot(path string) (err syscall.Errno) {
	if utils.IsRoot() == false {
		return syscall.Errno(syscall.EACCES)
	}
	if len(path) == 0 {
		return syscall.Errno(syscall.EINVAL)
	}
	pathp, e := syscall.BytePtrFromString(path)
	if e != nil {
		log.Fatal("Error converting process name to pointer")
	}
	_, _, error := syscall.RawSyscall(CHROOT, uintptr(unsafe.Pointer(pathp)), 0, 0)
	return error
}

// Mount file systems. For now this is used only for bind mounting
func Mount(source, dest, fstype string) (err syscall.Errno) {
	slog.Debug("Mount", "pid", os.Getpid(), "user", os.Geteuid(), "source", source, "dest", dest, "type", fstype)
	if len(source) == 0 || len(dest) == 0 {
		return syscall.Errno(syscall.EINVAL)
	}

	if !utils.IsVirtual(source) {
		if utils.PathExists(source) == false || utils.PathExists(dest) == false {
			slog.Debug("Mount returns EINVAL")
			return syscall.Errno(syscall.EINVAL)
		}

	}
	sourcep, e := syscall.BytePtrFromString(source)
	if e != nil {
		log.Fatal("Error converting source path to pointer")
	}
	destp, e := syscall.BytePtrFromString(dest)
	if e != nil {
		log.Fatal("Error converting dest path to pointer")
	}

	typep, e := syscall.BytePtrFromString(fstype)
	if e != nil {
		log.Fatal("Error converting type to pointer")
	}
	_, _, error := syscall.RawSyscall6(MOUNT, uintptr(unsafe.Pointer(sourcep)), uintptr(unsafe.Pointer(destp)), uintptr(unsafe.Pointer(typep)), 0, 0, 0)
	slog.Debug("Mount returns ", "error", error)
	return error

}

// Umount file systems
func Umount(target string, flags int) (err syscall.Errno) {
	slog.Debug("Umount", "pid", os.Getpid(), "user", os.Geteuid(), "target", target)
	if len(target) == 0 {
		return syscall.Errno(syscall.EINVAL)
	}

	targetp, e := syscall.BytePtrFromString(target)
	if e != nil {
		log.Fatal("Error converting target path to pointer")
	}
	_, _, error := syscall.RawSyscall(UMOUNT, uintptr(unsafe.Pointer(targetp)), uintptr(flags), 0)

	slog.Debug("Umount returns ", "error", error)
	return error

}

// Change directory
func Chdir(dir string) (err syscall.Errno) {
	slog.Debug("Chdir", "pid", os.Getpid(), "user", os.Geteuid(), "dir", dir)
	if len(dir) == 0 {
		return syscall.Errno(syscall.EINVAL)
	}

	dirp, e := syscall.BytePtrFromString(dir)
	if e != nil {
		log.Fatal("Error converting directory path to pointer")
	}
	_, _, error := syscall.RawSyscall(CHDIR, uintptr(unsafe.Pointer(dirp)), 0, 0)

	slog.Debug("Chdir returns ", "error", error)
	return error

}
