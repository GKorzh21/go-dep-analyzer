package cmd

import (
	"fmt"
	"os"

	"github.com/GKorzh21/go-dep-analyzer/internal/analyzer"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-dep-analyzer <repository-url>",
	Short: "Analyzes Go module dependencies for a given Git repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoURL := args[0]
		return analyzer.Analyze(repoURL)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
