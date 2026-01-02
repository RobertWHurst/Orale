package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/RobertWHurst/orale"
	"github.com/spf13/cobra"
)

var encryptCmd = &cobra.Command{
	Use:               "encrypt <name> [value]",
	Short:             "Encrypt a value using the specified key",
	Long:              "Encrypt a value using the encryption key from ~/.config/orale/<name>.ok. If value is not provided, reads from stdin.",
	Args:              cobra.RangeArgs(1, 2),
	ValidArgsFunction: cobra.NoFileCompletions,
	Run: func(_ *cobra.Command, args []string) {
		keyName := args[0]
		var value string

		if len(args) == 2 {
			value = args[1]
		} else {
			reader := bufio.NewReader(os.Stdin)
			valueBytes, err := io.ReadAll(reader)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
			value = string(valueBytes)
		}

		if err := runEncrypt(keyName, value); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func runEncrypt(keyName, value string) error {
	keyPath, err := getKeyPath(keyName)
	if err != nil {
		return err
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("key file not found at %s\nRun 'orale keygen %s' to generate a new key", keyPath, keyName)
		}
		return fmt.Errorf("failed to read key file: %w", err)
	}

	if len(keyData) < baseKeySize {
		return fmt.Errorf("invalid key file format: expected at least %d bytes, got %d", baseKeySize, len(keyData))
	}

	encryptionKey := keyData[saltSize : saltSize+keySize]
	hmacKey := keyData[saltSize+keySize : baseKeySize]

	encrypted, err := orale.Encrypt([]byte(value), encryptionKey, hmacKey)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	fmt.Printf("ENC[%s]\n", encrypted)

	return nil
}
