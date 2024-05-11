/*
Copyright © 2024 Pierguido Lambri <olambri@redhat.com>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"log/slog"

	"github.com/spf13/cobra"
)

var (
	envVariables []string
	image        string
	base_path    string = "/tmp/test-chroot/"
)

func mount_virtfs(path string) syscall.Errno {
	proc_target := path + "/proc"
	sys_target := path + "/sys"
	dev_target := path + "/dev"
	err := Mount("proc", proc_target, "proc", 0)
	if err != 0 {
		log.Printf("Error mounting proc on the directory %v: %v", proc_target, err)
		return err
	}
	err = Mount("sys", sys_target, "sysfs", 0)
	if err != 0 {
		log.Printf("Error mounting sys on the directory %v: %v", sys_target, err)
		return err
	}
	err = Mount("devtmpfs", dev_target, "devtmpfs", 0xC0ED0000)
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

func run(args []string) {
	ppid := os.Getpid()
	slog.Debug("Forking", "pid thread", ppid, "user", os.Getuid())
	if len(args) == 0 {
		fmt.Printf("You need to specify a program to run\n")
		return
	}
	path := base_path + image
	err := mount_virtfs(path)
	defer umount_virtfs(path)
	if err != 0 {
		return
	}
	// For now we'll use a fixed path for the container images
	//utils.CleanupChrootDir(path, true)
	// utils.ExtractImage("utils/Fedora-minimal-chroot.tar", path)
	if int(pid) == 0 {
		slog.Debug("Child", "pid", pid, "pid thread", os.Getpid(), "pid parent", os.Getppid())
		slog.Debug("Child", "exec", args[0], "options", args)
		// New namespaces
		err = Unshare(CLONE_NEWNS)
		if err != 0 {
			log.Fatal("Error trying to unshare ", ": ", err)
		}

		// Chroot into the new environment
		err = Chroot(path)
		if err != 0 {
			log.Fatal("Error trying to chroot into ", path, ": ", err)
		}
		err = Chdir("/")
		if err != 0 {
			log.Fatal("Error trying to chdir into root", err)
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
	} else {
		// Wait
		slog.Debug("Parent", "child pid", pid, "pid thread", os.Getpid())
		Wait(int(pid))
		return
	}
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
