package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var keylistCmd = &cobra.Command{
	Use:               "keylist",
	Short:             "List available encryption keys",
	Long:              "List all encryption keys available in ~/.config/orale/",
	Args:              cobra.NoArgs,
	ValidArgsFunction: cobra.NoFileCompletions,
	Run: func(_ *cobra.Command, _ []string) {
		if err := runKeylist(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runKeylist() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, configDir)

	entries, err := os.ReadDir(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No encryption keys found")
			fmt.Println("\nRun 'orale keygen <name>' to generate a new key")
			return nil
		}
		return fmt.Errorf("failed to read config directory: %w", err)
	}

	var keyNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".ok" {
			continue
		}

		keyPath := filepath.Join(configPath, entry.Name())
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			continue
		}

		if len(keyData) < baseKeySize {
			continue
		}

		keyName := string(keyData[baseKeySize:])
		if keyName != "" {
			keyNames = append(keyNames, keyName)
		}
	}

	if len(keyNames) == 0 {
		fmt.Println("No encryption keys found")
		fmt.Println("\nRun 'orale keygen <name>' to generate a new key")
		return nil
	}

	fmt.Println("Available encryption keys:")
	for _, name := range keyNames {
		fmt.Printf("  - %s\n", name)
	}

	return nil
}
