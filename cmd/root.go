package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var (
	verbose bool
	quiet   bool
	jsonOut bool
)

var rootCmd = &cobra.Command{
	Use:   "gitall",
	Short: "Manage multiple GitHub repositories in one command",
	Long: `GitAll is a CLI tool for batch-managing GitHub repositories.
Clone, pull, fetch, and check status across multiple user or
organisation accounts with a single command.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output in JSON format")
	rootCmd.MarkFlagsMutuallyExclusive("verbose", "quiet")
	rootCmd.MarkFlagsMutuallyExclusive("json", "quiet")

	rootCmd.Version = version
}
