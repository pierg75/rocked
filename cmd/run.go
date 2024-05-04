/*
Copyright © 2024 Pierguido Lambri <olambri@redhat.com>
*/
package cmd

import (
	"fmt"
	"log"
	"os"

	"log/slog"

	"github.com/spf13/cobra"
)

var (
	envVariables []string
	image        string
)

func run(args []string) {
	ppid := os.Getpid()
	slog.Debug("Forking", "pid thread", ppid, "user", os.Getuid())
	if len(args) == 0 {
		fmt.Printf("You need to specify a program to run\n")
		return
	}
	pid, err := Fork(nil)
	if err != 0 {
		fmt.Printf("Error forking: %v", int(err))
		return
	}
	// For now we'll use a fixed path for the container images
	path := "/tmp/test-chroot/" + image
	//utils.CleanupChrootDir(path, true)
	// utils.ExtractImage("utils/Fedora-minimal-chroot.tar", path)
	if int(pid) == 0 {
		slog.Debug("Child", "pid", pid, "pid thread", os.Getpid(), "pid parent", os.Getppid())
		slog.Debug("Child", "exec", args[0], "options", args)
		// mount proc and sys
		err := Mount("proc", path+"/proc", "proc")
		if err != 0 {
			log.Fatalf("Error mounting proc on the directory %v: %v", path+"/proc", err)
		}
		err = Mount("sys", path+"/sys", "sysfs")
		if err != 0 {
			log.Fatalf("Error mounting sys on the directory %v: %v", path+"/sys", err)
		}
		// Chroot into the new environment
		err = Chroot(path)
		if err != 0 {
			log.Fatal("Error trying to chroot into ", path, ": ", err)
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
