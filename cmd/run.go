/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a process",
	Run: func(cmd *cobra.Command, args []string) {
		var wstatus syscall.WaitStatus
		ppid := os.Getpid()
		fmt.Printf("go pid:%v user:%v\n", ppid, os.Getuid())
		fmt.Printf("Forking...\n")
		attr := syscall.ProcAttr{
			Dir:   "/tmp",
			Env:   os.Environ(),
			Files: []uintptr{uintptr(syscall.Stdin), uintptr(syscall.Stdout), uintptr(syscall.Stderr)},
			Sys:   nil,
		}
		pid, err := syscall.ForkExec(args[0], args[1:], &attr)
		if err != nil {
			log.Fatal(err)
		}
		_, err = syscall.Wait4(pid, &wstatus, 0, nil)
		for syscall.EINTR == err {
			_, err = syscall.Wait4(pid, &wstatus, 0, nil)
		}
		if pid != ppid {
			fmt.Printf("child pid:%v\n", pid)
			//fmt.Printf("child pid:%v attr:%#v\n", pid, attr)
		} else {
			fmt.Printf("parent pid:%v user:%v\n", os.Getpid(), os.Getuid())
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
