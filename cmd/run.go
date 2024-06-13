/*
Copyright © 2024 Pierguido Lambri <olambri@redhat.com>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	"log/slog"

	"github.com/spf13/cobra"
)

var (
	envVariables []string
	image        string
	base_path    string = "/tmp/containers/"
)

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

// This function should basically do all the work for the child process.
// It should not be able to return, but only execute the process invoked.
//
//go:noinline
//go:norace
//go:nocheckptr
func runFork(base_path, image string, args []string) (int, syscall.Errno) {
	slog.Debug("runFork", "base_path", base_path, "image", image, "args", args)
	// Untar the container image into a predefined root
	// For now let's use hardocded paths
	con, errc := SetContainer(image, base_path)
	if errc != nil {
		log.Fatal("Error trying to setup container ", ": ", errc)
	}
	cargs := CloneArgs{
		flags: CLONE_VFORK | CLONE_FILES | CLONE_NEWPID | CLONE_NEWNET | CLONE_INTO_CGROUP,
	}
	errCgroup := PrepareCgroup(con.id, &cargs)
	if errCgroup != nil {
		log.Fatal("Error while setting up cgroups: ", errCgroup)
	}
	pid, err := Fork(&cargs)
	if err != 0 {
		fmt.Printf("Error forking: %v", int(err))
		return -1, err
	}
	//Parent
	if pid != 0 {
		return int(pid), 0
	}

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

func run(args []string) {
	ppid := os.Getpid()
	slog.Debug("Forking", "pid thread", ppid, "user", os.Getuid())
	if len(args) == 0 {
		fmt.Printf("You need to specify a program to run\n")
		return
	}
	childpid, err := runFork(base_path, image, args)
	if err != 0 {
		log.Printf("There was an error while forking: %v", err)
	}
	// Wait
	Wait(childpid)
	log.Printf("%v exited\n", childpid)
	return
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a process",
	Run: func(cmd *cobra.Command, args []string) {
		run(args)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringArrayVarP(&envVariables, "env", "e", nil, "Sets environment variables. It can be repeated")
	runCmd.Flags().StringVarP(&image, "image", "i", "Fedora", "Use the container image")
	runCmd.MarkFlagRequired("image")
}
