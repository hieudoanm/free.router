package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fr",
	Short: "Run OpenRouter free models like Ollama — Cursor-compatible",
	Long: `FreeRouter proxies OpenRouter's free models through a local
OpenAI-compatible HTTP server so Cursor (and any OpenAI client) can
use them without any cost.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(statusCmd)
}
