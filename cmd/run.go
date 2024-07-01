/*
Copyright © 2024 Pierguido Lambri <olambri@redhat.com>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"

	"log/slog"

	"github.com/spf13/cobra"
)

var (
	envVariables []string
	image        string
	base_path    string = "/tmp/containers/"
)

type IDmapping struct {
	InsideID  int
	OutsideID int
	Len       int
}

func mount_virtfs(path string) syscall.Errno {
	proc_target := path + "/proc"
	sys_target := path + "/sys"
	dev_target := path + "/dev"
	err := Mount("proc", proc_target, "proc", 0, "")
	if err != 0 {
		log.Printf("Error mounting proc on the directory %v: %v", proc_target, err)
		return err
	}
	err = Mount("sys", sys_target, "sysfs", 0, "")
	if err != 0 {
		log.Printf("Error mounting sys on the directory %v: %v", sys_target, err)
		return err
	}
	err = Mount("devtmpfs", dev_target, "devtmpfs", MS_MGC_VAL, "")
	if err != 0 {
		log.Printf("Error mounting dev on the directory %v: %v", dev_target, err)
		return err
	}
	return 0
}

func umount_virtfs(path string) syscall.Errno {
	proc_target := path + "/proc"
	sys_target := path + "/sys"
	dev_target := path + "/dev"
	err := Umount(proc_target, 0)
	if err != 0 {
		log.Printf("Error umounting proc on the directory %v: %v", proc_target, err)
		return err
	}
	err = Umount(sys_target, 0)
	if err != 0 {
		log.Printf("error umounting sys on the directory %v: %v", sys_target, err)
		return err
	}
	err = Umount(dev_target, 0)
	if err != 0 {
		log.Printf("error umounting dev on the directory %v: %v", sys_target, err)
		return err
	}
	return 0
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func WriteMaps(dest string, mapping IDmapping) error {
	idmapf, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		slog.Debug("writeMaps error opening file", "dest", dest)
		return err
	}
	m := strconv.Itoa(mapping.InsideID) + " " + strconv.Itoa(mapping.OutsideID) + " " + strconv.Itoa(mapping.Len)
	_, err = idmapf.Write([]byte(m))
	if err != nil {
		slog.Debug("writeMaps error writing to ", "dest", dest, "m", m)
		return err
	}
	defer idmapf.Close()
	return nil
}

// This function should basically do all the work for the child process.
// It should not be able to return, but only execute the process invoked.
//
//go:noinline
//go:norace
//go:nocheckptr
func runFork1(base_path, image string, args []string) (int, syscall.Errno) {
	slog.Debug("runFork", "base_path", base_path, "image", image, "args", args)
	// Untar the container image into a predefined root
	// For now let's use hardocded paths
	con, errc := SetContainer(image, base_path)
	if errc != nil {
		log.Fatal("Error trying to setup container ", ": ", errc)
	}
	cargs := CloneArgs{
		flags: CLONE_VFORK | CLONE_FILES | CLONE_NEWPID | CLONE_NEWNET | CLONE_INTO_CGROUP | CLONE_NEWUSER,
	}
	slog.Debug("runFork", "base_path", base_path, "image", image, "cargs flags", cargs.flags, "cargs cg fd", cargs.cgroup)
	// Before doing anything with the cgroups, let's make sure some controllers
	// are enabled in the root cgroup
	errctrl := enableControllers(nil, "/sys/fs/cgroup/")
	if errctrl != nil {
		fmt.Printf("Error enabling controllers: %v", errctrl)
		return -1, syscall.EIO
	}
	cgroup, errCG := PrepareCgroup(con, &cargs)
	if errCG != nil {
		log.Fatal("Error while setting up cgroups: ", "id", cgroup.Id, "err", errCG)
	}
	// We need now to get the os.File reference to get then a file descriptor
	// and to pass this to the clone3 syscall
	// This is necessary to avoid the file closure too early and getting EBAFD
	// from clone3.
	// Set also the limits at this stage.
	cgFile, errfd := cgroup.GetCGFd()
	if errfd != nil {
		log.Fatal("Error while getting cgroup fd ", "id", cgroup.Id, "err", errfd)
	}
	cargs.cgroup = uint64(cgFile.Fd())
	cgroup.SetCGLimits()
	defer cgFile.Fd()
	// Let's create the child process
	pid, err := Fork(&cargs)
	if err != 0 {
		fmt.Printf("Error forking: %v", int(err))
		return -1, err
	}
	//Parent
	if pid != 0 {
		return int(pid), 0
	}

	// There should be some sort of communication between parent/child to
	// continue with the execution after the parent has updated the maps.
	// For now let's just wait a second
	time.Sleep(time.Second * 1)
	slog.Debug("Child", "pid", pid, "pid thread", os.Getpid(), "pid parent", os.Getppid())
	slog.Debug("Child", "exec", args[0], "options", args)

	err = Unshare(CLONE_NEWNS | CLONE_NEWUTS)
	if err != 0 {
		log.Fatal("Error trying to unshare ", ": ", err)
	}
	err = SetMount("/", MS_REC|MS_PRIVATE)
	if err != 0 {
		log.Fatal("Error trying to switch root as private: ", err)
	}
	err = SetHostname(con.id, len(con.id))
	if err != 0 {
		log.Println("Error trying to set the hostname ", err)
	}
	// There should be some sort of communication between parent/child to
	// continue with the execution after the parent has updated the maps.
	// For now let's just wait a second
	time.Sleep(time.Second * 60)
	mergepath := con.Path + "/overlay/merge"
	err = Mount("overlay", mergepath, "overlay", MS_MGC_VAL, "lowerdir="+con.Path+"/image_root/,upperdir="+con.Path+"/overlay/upper,workdir="+con.Path+"/overlay/work")
	if err != 0 {
		log.Printf("Error mounting overlay on the directory %v: %v", mergepath, err)
		return -1, err
	}
	log.Println("Created a new root fs for our container :", con.Path+"/overlay/")
	//// This is to temporally have a mountpoint for pivot_root
	//err = Mount(path, path, "", MS_BIND)
	//if err != 0 {
	//	log.Fatal("Error bind mount ", path, "on ", path, ": ", err)
	//}

	err = mount_virtfs(mergepath)
	defer umount_virtfs(mergepath)
	if err != 0 {
		return -1, err
	}
	err = Chdir(mergepath)
	if err != 0 {
		log.Fatal("Error trying to chdir into ", mergepath, ": ", err)
	}
	err = PivotRoot(".", ".")
	if err != 0 {
		log.Fatal("Error trying to pivot_root into ", con.Path, ": ", err)
	}
	err = Umount(".", syscall.MNT_DETACH)
	if err != 0 {
		log.Fatal("Error trying to umount '.'", err)
	}
	// Exec
	a := ExecArgs{
		Exe:     args[0],
		Exeargs: args,
	}
	a.Env = os.Environ()
	for _, entry := range envVariables {
		a.Env = append(a.Env, entry)
	}
	err = Exec(&a)
	if err != 0 {
		log.Fatal("Error executing ", args[0], ": ", err)
	}
	// Clean up everything before returning
	defer func() {
		er := os.RemoveAll(con.Path)
		if er != nil {
			log.Println("Error removing ", con.Path)
		}
	}()
	return 0, 0
}

func runFork(args []string) int {
	ppid := os.Getpid()
	slog.Debug("Forking", "pid thread", ppid, "user", os.Getuid())
	if len(args) == 0 {
		fmt.Printf("You need to specify a program to run\n")
		return -1
	}
	childpid, err := runFork1(base_path, image, args)
	if err != 0 {
		log.Printf("There was an error while forking: %v", err)
		return -1
	}
	return childpid
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a process",
	Run: func(cmd *cobra.Command, args []string) {
		childpid := runFork(args)
		// Write the id maps
		childmap := IDmapping{
			InsideID:  0,
			OutsideID: 15603,
			Len:       1,
		}
		err := WriteMaps("/proc/"+strconv.Itoa(childpid)+"/uid_map", childmap)
		if err != nil {
			log.Printf("There was an error when writing ID maps: %v", err)
		}
		// Wait
		Wait(childpid)
		log.Printf("%v exited\n", childpid)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringArrayVarP(&envVariables, "env", "e", nil, "Sets environment variables. It can be repeated")
	runCmd.Flags().StringVarP(&image, "image", "i", "Fedora", "Use the container image")
	runCmd.MarkFlagRequired("image")
}
