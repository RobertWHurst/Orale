package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "orale",
	Short: "Orale - Configuration encryption utilities",
	Long:  "Orale provides utilities for managing encrypted configuration values for local development teams.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, _ []string) {
		if err := cmd.Help(); err != nil {
			fmt.Printf("Error showing help: %s\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(keygenCmd)
	rootCmd.AddCommand(keysetCmd)
	rootCmd.AddCommand(keylistCmd)
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(explainCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
