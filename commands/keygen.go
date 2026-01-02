package commands

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	configDir     = ".config/orale"
	saltSize      = 32 // Salt for potential password-based key derivation
	keySize       = 32 // AES-256
	hmacSize      = 32 // HMAC-SHA256
	baseKeySize   = saltSize + keySize + hmacSize
)

// <name>.ok file format (binary):
// Bytes 0-31:      Salt (32 bytes)
// Bytes 32-63:     Encryption Key (32 bytes)
// Bytes 64-95:     HMAC Key (32 bytes)
// Bytes 96-end:    Key Name (UTF-8 string, variable length)

var keygenCmd = &cobra.Command{
	Use:               "keygen <name>",
	Short:             "Generate a new encryption key",
	Long:              "Generate a new encryption key and save it to ~/.config/orale/<name>.ok for team sharing.",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: cobra.NoFileCompletions,
	Run: func(_ *cobra.Command, args []string) {
		if err := runKeygen(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runKeygen(keyName string) error {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	encryptionKey := make([]byte, keySize)
	if _, err := rand.Read(encryptionKey); err != nil {
		return fmt.Errorf("failed to generate encryption key: %w", err)
	}

	hmacKey := make([]byte, hmacSize)
	if _, err := rand.Read(hmacKey); err != nil {
		return fmt.Errorf("failed to generate HMAC key: %w", err)
	}

	keyData := make([]byte, 0, baseKeySize+len(keyName))
	keyData = append(keyData, salt...)
	keyData = append(keyData, encryptionKey...)
	keyData = append(keyData, hmacKey...)
	keyData = append(keyData, []byte(keyName)...)

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

	encodedKey := base64.StdEncoding.EncodeToString(keyData)

	fmt.Printf("Generated new encryption key: %s\n", keyName)
	fmt.Printf("Saved to: %s\n\n", keyPath)

	fmt.Println("Share this key with your team:")
	fmt.Println("----------------------------------------")
	fmt.Println(encodedKey)
	fmt.Println("----------------------------------------")
	fmt.Printf("\nTeam members should run:\n  orale keyset %s\n", encodedKey)
	fmt.Println("\nKeep it secure and don't commit it to version control!")

	return nil
}

func getKeyPath(keyName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(home, configDir, keyName+".ok"), nil
}
