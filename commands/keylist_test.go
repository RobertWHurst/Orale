package commands

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

func Test_runKeylist_noKeys(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	err := runKeylist()
	if err != nil {
		t.Fatalf("runKeylist() should not error when no keys exist, error = %v", err)
	}
}

func Test_runKeylist_singleKey(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName := "test-key"
	keyData := make([]byte, baseKeySize)
	if _, err := rand.Read(keyData); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	keyData = append(keyData, []byte(keyName)...)

	configPath := filepath.Join(tempDir, configDir)
	if err := os.MkdirAll(configPath, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	keyPath := filepath.Join(configPath, keyName+".ok")
	if err := os.WriteFile(keyPath, keyData, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	err := runKeylist()
	if err != nil {
		t.Fatalf("runKeylist() error = %v", err)
	}
}

func Test_runKeylist_multipleKeys(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyNames := []string{"dev", "staging", "production"}

	configPath := filepath.Join(tempDir, configDir)
	if err := os.MkdirAll(configPath, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	for _, keyName := range keyNames {
		keyData := make([]byte, baseKeySize)
		if _, err := rand.Read(keyData); err != nil {
			t.Fatalf("failed to generate key for %s: %v", keyName, err)
		}
		keyData = append(keyData, []byte(keyName)...)

		keyPath := filepath.Join(configPath, keyName+".ok")
		if err := os.WriteFile(keyPath, keyData, 0600); err != nil {
			t.Fatalf("failed to write key file for %s: %v", keyName, err)
		}
	}

	err := runKeylist()
	if err != nil {
		t.Fatalf("runKeylist() error = %v", err)
	}
}

func Test_runKeylist_ignoresInvalidFiles(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	configPath := filepath.Join(tempDir, configDir)
	if err := os.MkdirAll(configPath, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	validKeyName := "valid-key"
	validKeyData := make([]byte, baseKeySize)
	if _, err := rand.Read(validKeyData); err != nil {
		t.Fatalf("failed to generate valid key: %v", err)
	}
	validKeyData = append(validKeyData, []byte(validKeyName)...)

	if err := os.WriteFile(filepath.Join(configPath, validKeyName+".ok"), validKeyData, 0600); err != nil {
		t.Fatalf("failed to write valid key: %v", err)
	}

	if err := os.WriteFile(filepath.Join(configPath, "not-a-key.txt"), []byte("ignore"), 0600); err != nil {
		t.Fatalf("failed to write txt file: %v", err)
	}

	if err := os.WriteFile(filepath.Join(configPath, "too-short.ok"), []byte("short"), 0600); err != nil {
		t.Fatalf("failed to write invalid key: %v", err)
	}

	invalidKeyData := make([]byte, baseKeySize)
	if err := os.WriteFile(filepath.Join(configPath, "no-name.ok"), invalidKeyData, 0600); err != nil {
		t.Fatalf("failed to write key without name: %v", err)
	}

	err := runKeylist()
	if err != nil {
		t.Fatalf("runKeylist() error = %v", err)
	}
}

func Test_runKeylist_ignoresDirectories(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	configPath := filepath.Join(tempDir, configDir)
	if err := os.MkdirAll(configPath, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(configPath, "subdir.ok"), 0700); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	keyName := "real-key"
	keyData := make([]byte, baseKeySize)
	if _, err := rand.Read(keyData); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	keyData = append(keyData, []byte(keyName)...)

	if err := os.WriteFile(filepath.Join(configPath, keyName+".ok"), keyData, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	err := runKeylist()
	if err != nil {
		t.Fatalf("runKeylist() error = %v", err)
	}
}
