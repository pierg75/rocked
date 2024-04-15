/*
Copyright © 2024 Pierguido Lambri <plambri@redhat.com>
*/
package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var envVariables []string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "Runs a containers",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("Running command", "cmd", args)
		slog.Debug("Environment Variables", "EnvVars", envVariables)
		c := exec.Command(args[0], args[1:]...)
		c.Env = os.Environ()
		for _, entry := range envVariables {
			c.Env = append(c.Env, entry)
		}
		out, err := c.Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(string(out))
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringArrayVarP(&envVariables, "env", "e", nil, "Sets environment variables. It can be repeated")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
