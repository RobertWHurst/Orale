package orale

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	decryptConfigDir = ".config/orale"
	saltSize         = 32
	keySize          = 32
	hmacSize         = 32
	baseKeySize      = saltSize + keySize + hmacSize
)

var encryptedValuePattern = regexp.MustCompile(`^ENC\[([A-Za-z0-9+/=]+)\]$`)

func isEncrypted(value string) bool {
	return encryptedValuePattern.MatchString(value)
}

func decryptValue(value string) (string, error) {
	matches := encryptedValuePattern.FindStringSubmatch(value)
	if matches == nil {
		return value, nil
	}

	encryptedData := matches[1]

	keyFiles, err := loadAllKeyFiles()
	if err != nil {
		return "", err
	}

	if len(keyFiles) == 0 {
		return "", fmt.Errorf("no encryption keys found in ~/.config/orale/\nRun 'orale keygen <name>' to generate a new key")
	}

	var lastErr error
	for _, keyData := range keyFiles {
		if len(keyData) < baseKeySize {
			continue
		}

		encryptionKey := keyData[saltSize : saltSize+keySize]
		hmacKey := keyData[saltSize+keySize : baseKeySize]

		decrypted, err := decrypt(encryptedData, encryptionKey, hmacKey)
		if err == nil {
			return decrypted, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return "", fmt.Errorf("failed to decrypt value with any available key: %w", lastErr)
	}

	return "", fmt.Errorf("failed to decrypt value: no valid keys found")
}

func loadAllKeyFiles() ([][]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, decryptConfigDir)

	entries, err := os.ReadDir(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return [][]byte{}, nil
		}
		return nil, fmt.Errorf("failed to read config directory: %w", err)
	}

	var keyFiles [][]byte
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

		keyFiles = append(keyFiles, keyData)
	}

	return keyFiles, nil
}

func decrypt(encodedData string, encryptionKey, hmacKey []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", fmt.Errorf("invalid base64 encoding: %w", err)
	}

	if len(data) < ivSize+sha256.Size {
		return "", errors.New("encrypted data too short")
	}

	macSize := sha256.Size
	iv := data[:ivSize]
	ciphertext := data[ivSize : len(data)-macSize]
	mac := data[len(data)-macSize:]

	h := hmac.New(sha256.New, hmacKey)
	h.Write(iv)
	h.Write(ciphertext)
	expectedMac := h.Sum(nil)

	if !hmac.Equal(mac, expectedMac) {
		return "", errors.New("authentication failed: HMAC verification failed")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of block size")
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	plaintext, err = pkcs7Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to unpad plaintext: %w", err)
	}

	return string(plaintext), nil
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}

	padding := int(data[len(data)-1])
	if padding > aes.BlockSize || padding == 0 {
		return nil, errors.New("invalid padding")
	}

	if len(data) < padding {
		return nil, errors.New("invalid padding size")
	}

	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			return nil, errors.New("invalid padding bytes")
		}
	}

	return data[:len(data)-padding], nil
}

// ensureDecrypted checks if a string value is encrypted and decrypts it if needed.
// If the value is not encrypted, it returns the value unchanged.
// This ensures all string values from any source are properly decrypted.
func ensureDecrypted(value string) (string, error) {
	if !isEncrypted(value) {
		return value, nil
	}
	return decryptValue(value)
}
