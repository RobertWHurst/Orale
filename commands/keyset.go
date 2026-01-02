package commands

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var keysetCmd = &cobra.Command{
	Use:               "keyset <key>",
	Short:             "Set the team encryption key",
	Long:              "Set a base64-encoded encryption key shared by a team member and save it to ~/.config/orale/<name>.ok.",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: cobra.NoFileCompletions,
	Run: func(_ *cobra.Command, args []string) {
		if err := runKeyset(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runKeyset(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return fmt.Errorf("no key provided")
	}

	keyData, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return fmt.Errorf("invalid base64 encoding: %w", err)
	}

	if len(keyData) < baseKeySize {
		return fmt.Errorf("invalid key size: minimum %d bytes, got %d", baseKeySize, len(keyData))
	}

	keyName := string(keyData[baseKeySize:])
	if keyName == "" {
		return fmt.Errorf("key name not found in key data")
	}

	keyPath, err := getKeyPath(keyName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(keyPath, keyData, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	fmt.Printf("Successfully imported key '%s' to: %s\n", keyName, keyPath)

	return nil
}
