package commands

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

func Test_runEncrypt_encryptsValue(t *testing.T) {
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

	keyPath := filepath.Join(tempDir, configDir, keyName+".ok")
	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(keyPath, keyData, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	value := "my-secret-value"

	err := runEncrypt(keyName, value)
	if err != nil {
		t.Fatalf("runEncrypt() error = %v", err)
	}
}

func Test_runEncrypt_keyNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	err := runEncrypt("nonexistent-key", "test-value")
	if err == nil {
		t.Error("runEncrypt() should return error when key file doesn't exist")
	}
}

func Test_runEncrypt_invalidKeyFormat(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName := "bad-key"
	keyPath := filepath.Join(tempDir, configDir, keyName+".ok")
	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	invalidKeyData := []byte("too short")
	if err := os.WriteFile(keyPath, invalidKeyData, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	err := runEncrypt(keyName, "test-value")
	if err == nil {
		t.Error("runEncrypt() should return error for invalid key format")
	}
}

func Test_runEncrypt_emptyValue(t *testing.T) {
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

	keyPath := filepath.Join(tempDir, configDir, keyName+".ok")
	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(keyPath, keyData, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	err := runEncrypt(keyName, "")
	if err != nil {
		t.Fatalf("runEncrypt() should handle empty value, error = %v", err)
	}
}

func Test_runEncrypt_longValue(t *testing.T) {
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

	keyPath := filepath.Join(tempDir, configDir, keyName+".ok")
	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	if err := os.WriteFile(keyPath, keyData, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	longValue := make([]byte, 1024)
	for i := range longValue {
		longValue[i] = byte(i % 256)
	}

	err := runEncrypt(keyName, string(longValue))
	if err != nil {
		t.Fatalf("runEncrypt() error = %v", err)
	}
}

