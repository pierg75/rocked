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
	"rocked/utils"

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
	err = Mount("devtmpfs", dev_target, "devtmpfs", MS_MGC_VAL)
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
	pid, err := Fork(nil)
	if err != 0 {
		fmt.Printf("Error forking: %v", int(err))
		return
	}
	if int(pid) == 0 {
		slog.Debug("Child", "pid", pid, "pid thread", os.Getpid(), "pid parent", os.Getppid())
		slog.Debug("Child", "exec", args[0], "options", args)
		// Untar the container image into a predefined root
		// For now let's use hardocded paths
		var defaultContainerImage string = "/tmp/containers/" + image
		var defaultSourceContainersImages string = "blobs/container_images/" + image + ".tar"
		utils.ExtractImage(defaultSourceContainersImages, defaultContainerImage)
		con := NewContainer(defaultContainerImage)
		errcon := con.LoadConfigJson()
		slog.Debug("Child", "Manifests", con.Index.Manifests)
		for _, manifest := range con.Index.Manifests {
			exists := con.BlobExists(manifest.Digest.Algorithm(), manifest.Digest.Encoded())
			slog.Debug("Child", "manifest", manifest.Annotations["org.opencontainers.image.ref.name"], "exists", exists)
		}
		if errcon != nil {
			log.Fatalf("Error opening container index json")
		}
		err = Unshare(CLONE_NEWNS)
		if err != 0 {
			log.Fatal("Error trying to unshare ", ": ", err)
		}
		err = SetMount("/", MS_REC|MS_PRIVATE)
		if err != 0 {
			log.Fatal("Error trying to switch root as private: ", err)
		}
		// This is to temporally have a mountpoint for pivot_root
		err = Mount(path, path, "", MS_BIND)
		if err != 0 {
			log.Fatal("Error bind mount ", path, "on ", path, ": ", err)
		}

		err := mount_virtfs(path)
		defer umount_virtfs(path)
		if err != 0 {
			return
		}
		err = Chdir(path)
		if err != 0 {
			log.Fatal("Error trying to chdir into ", path, ": ", err)
		}
		err = PivotRoot(".", ".")
		if err != 0 {
			log.Fatal("Error trying to pivot_root into ", path, ": ", err)
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
}
