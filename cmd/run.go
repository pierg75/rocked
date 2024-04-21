/*
Copyright © 2024 Pierguido Lambri <olambri@redhat.com>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a process",
	Run: func(cmd *cobra.Command, args []string) {
		ppid := os.Getpid()
		fmt.Printf("go pid:%v user:%v\n", ppid, os.Getuid())
		fmt.Printf("Forking...\n")
		if len(args) == 0 {
			fmt.Printf("You need to specify a program to run\n")
			return
		}
		a := ForkExecArgs{
			exe:     args[0],
			exeargs: args,
		}
		pid, err := ForkExec(&a)
		if err != 0 {
			fmt.Printf("Error: %v", int(err))
			return
		}
		Wait(int(pid))
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
