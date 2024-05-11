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
	EXECVE  uintptr = 59
	WAIT4   uintptr = 61
	CHDIR   uintptr = 80
	CLONE3  uintptr = 435
	CHROOT  uintptr = 161
	MOUNT   uintptr = 165
	UMOUNT  uintptr = 166
	UNSHARE uintptr = 272
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

// Fork flags
var (
	CLONE_VM             uintptr = 0x00000100 /* set if VM shared between processes */
	CLONE_FS             uintptr = 0x00000200 /* set if fs info shared between processes */
	CLONE_FILES          uintptr = 0x00000400 /* set if open files shared between processes */
	CLONE_SIGHAND        uintptr = 0x00000800 /* set if signal handlers and blocked signals shared */
	CLONE_PIDFD          uintptr = 0x00001000 /* set if a pidfd should be placed in parent */
	CLONE_PTRACE         uintptr = 0x00002000 /* set if we want to let tracing continue on the child too */
	CLONE_VFORK          uintptr = 0x00004000 /* set if the parent wants the child to wake it up on mm_release */
	CLONE_PARENT         uintptr = 0x00008000 /* set if we want to have the same parent as the cloner */
	CLONE_THREAD         uintptr = 0x00010000 /* Same thread group? */
	CLONE_NEWNS          uintptr = 0x00020000 /* New mount namespace group */
	CLONE_SYSVSEM        uintptr = 0x00040000 /* share system V SEM_UNDO semantics */
	CLONE_SETTLS         uintptr = 0x00080000 /* create a new TLS for the child */
	CLONE_PARENT_SETTID  uintptr = 0x00100000 /* set the TID in the parent */
	CLONE_CHILD_CLEARTID uintptr = 0x00200000 /* clear the TID in the child */
	CLONE_DETACHED       uintptr = 0x00400000 /* Unused, ignored */
	CLONE_UNTRACED       uintptr = 0x00800000 /* set if the tracing process can't force CLONE_PTRACE on this clone */
	CLONE_CHILD_SETTID   uintptr = 0x01000000 /* set the TID in the child */
	CLONE_NEWCGROUP      uintptr = 0x02000000 /* New cgroup namespace */
	CLONE_NEWUTS         uintptr = 0x04000000 /* New utsname namespace */
	CLONE_NEWIPC         uintptr = 0x08000000 /* New ipc namespace */
	CLONE_NEWUSER        uintptr = 0x10000000 /* New user namespace */
	CLONE_NEWPID         uintptr = 0x20000000 /* New pid namespace */
	CLONE_NEWNET         uintptr = 0x40000000 /* New network namespace */
	CLONE_IO             uintptr = 0x80000000 /* Clone io context */
)

var (
	MS_RDONLY      uintptr = 1    /* Mount read-only */
	MS_NOSUID      uintptr = 2    /* Ignore suid and sgid bits */
	MS_NODEV       uintptr = 4    /* Disallow access to device special files */
	MS_NOEXEC      uintptr = 8    /* Disallow program execution */
	MS_SYNCHRONOUS uintptr = 16   /* Writes are synced at once */
	MS_REMOUNT     uintptr = 32   /* Alter flags of a mounted FS */
	MS_MANDLOCK    uintptr = 64   /* Allow mandatory locks on an FS */
	MS_DIRSYNC     uintptr = 128  /* Directory modifications are synchronous */
	MS_NOATIME     uintptr = 1024 /* Do not update access times. */
	MS_NODIRATIME  uintptr = 2048 /* Do not update directory access times */
	MS_BIND        uintptr = 4096
	MS_MOVE        uintptr = 8192
	MS_REC         uintptr = 16384
	MS_VERBOSE     uintptr = 32768 /* War is peace. Verbosity is silence.
	   MS_VERBOSE is deprecated. */
	MS_SILENT      uintptr = 32768
	MS_POSIXACL    uintptr = (1 << 16) /* VFS does not apply the umask */
	MS_UNBINDABLE  uintptr = (1 << 17) /* change to unbindable */
	MS_PRIVATE     uintptr = (1 << 18) /* change to private */
	MS_SLAVE       uintptr = (1 << 19) /* change to slave */
	MS_SHARED      uintptr = (1 << 20) /* change to shared */
	MS_RELATIME    uintptr = (1 << 21) /* Update atime relative to mtime/ctime. */
	MS_KERNMOUNT   uintptr = (1 << 22) /* this is a kern_mount call */
	MS_I_VERSION   uintptr = (1 << 23) /* Update inode I_version field */
	MS_STRICTATIME uintptr = (1 << 24) /* Always perform atime updates */
	MS_LAZYTIME    uintptr = (1 << 25) /* Update the on-disk [acm]times lazily */

	/* These sb flags are internal to the kernel */
	MS_SUBMOUNT     uintptr = (1 << 26)
	MS_NOREMOTELOCK uintptr = (1 << 27)
	MS_NOSEC        uintptr = (1 << 28)
	MS_BORN         uintptr = (1 << 29)
	MS_ACTIVE       uintptr = (1 << 30)
	MS_NOUSER       uintptr = (1 << 31)

	/*
	 * Superblock flags that can be altered by MS_REMOUNT
	 */
	MS_RMT_MASK uintptr = MS_RDONLY | MS_SYNCHRONOUS | MS_MANDLOCK | MS_I_VERSION | MS_LAZYTIME

	/*
	 * Old magic mount flag and mask
	 */
	MS_MGC_VAL uintptr = 0xC0ED0000
	MS_MGC_MSK uintptr = 0xffff0000
)

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

// Mount file systems.
// source and destination must be valid paths.
// It takes an integer flags
func Mount(source, dest, fstype string, flags uintptr) (err syscall.Errno) {
	slog.Debug("Mount", "pid", os.Getpid(), "user", os.Geteuid(), "source", source, "dest", dest, "type", fstype)
	if len(source) == 0 || len(dest) == 0 {
		return syscall.Errno(syscall.EINVAL)
	}

	if !utils.IsVirtual(source) {
		if utils.PathExists(source) == false {
			log.Printf("Invalid source (does the directory exist?)")
			return syscall.Errno(syscall.ENOENT)
		}
		if utils.PathExists(dest) == false {
			log.Printf("Invalid destination (does the directory exist?)")
			return syscall.Errno(syscall.ENOENT)
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
	_, _, error := syscall.RawSyscall6(MOUNT, uintptr(unsafe.Pointer(sourcep)), uintptr(unsafe.Pointer(destp)), uintptr(unsafe.Pointer(typep)), uintptr(flags), 0, 0)
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

// Disassociate parts of the process context
func Unshare(flags uintptr) (err syscall.Errno) {
	slog.Debug("Unshare", "pid", os.Getpid(), "user", os.Geteuid(), "flags", flags)
	if flags == 0 {
		return
	}

	_, _, error := syscall.RawSyscall(UNSHARE, flags, 0, 0)

	slog.Debug("Unshare returns ", "error", error)
	return error

}
