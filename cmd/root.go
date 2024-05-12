/*
Copyright © 2024 Pierguido Lambri <plambri@redhat.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"log/slog"
)

var (
	Verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rocked",
	Short: "Handles containers",
	Long:  `A simple utility to handle container.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if Verbose {
			var programLevel = new(slog.LevelVar)
			h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
			slog.SetDefault(slog.New(h))
			programLevel.Set(slog.LevelDebug)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose logging")
}
