package commands

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func Test_runKeygen_generatesValidKeyFile(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName := "test-key"

	err := runKeygen(keyName)
	if err != nil {
		t.Fatalf("runKeygen() error = %v", err)
	}

	keyPath := filepath.Join(tempDir, configDir, keyName+".ok")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("key file was not created at %s", keyPath)
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read key file: %v", err)
	}

	expectedSize := baseKeySize + len(keyName)
	if len(keyData) != expectedSize {
		t.Errorf("key file size = %d, want %d", len(keyData), expectedSize)
	}

	embeddedName := string(keyData[baseKeySize:])
	if embeddedName != keyName {
		t.Errorf("embedded key name = %q, want %q", embeddedName, keyName)
	}

	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("failed to stat key file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("key file permissions = %o, want 0600", perm)
	}
}

func Test_runKeygen_createsConfigDirectory(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName := "new-key"

	err := runKeygen(keyName)
	if err != nil {
		t.Fatalf("runKeygen() error = %v", err)
	}

	configPath := filepath.Join(tempDir, configDir)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("config directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("config path is not a directory")
	}

	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("config directory permissions = %o, want 0700", perm)
	}
}

func Test_runKeygen_generatesDifferentKeys(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName1 := "key1"
	keyName2 := "key2"

	err := runKeygen(keyName1)
	if err != nil {
		t.Fatalf("first runKeygen() error = %v", err)
	}

	err = runKeygen(keyName2)
	if err != nil {
		t.Fatalf("second runKeygen() error = %v", err)
	}

	keyPath1 := filepath.Join(tempDir, configDir, keyName1+".ok")
	keyPath2 := filepath.Join(tempDir, configDir, keyName2+".ok")

	keyData1, err := os.ReadFile(keyPath1)
	if err != nil {
		t.Fatalf("failed to read first key file: %v", err)
	}

	keyData2, err := os.ReadFile(keyPath2)
	if err != nil {
		t.Fatalf("failed to read second key file: %v", err)
	}

	if string(keyData1[:baseKeySize]) == string(keyData2[:baseKeySize]) {
		t.Error("generated keys should be different")
	}
}

func Test_getKeyPath(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName := "test"
	expected := filepath.Join(tempDir, configDir, keyName+".ok")

	result, err := getKeyPath(keyName)
	if err != nil {
		t.Fatalf("getKeyPath() error = %v", err)
	}

	if result != expected {
		t.Errorf("getKeyPath() = %q, want %q", result, expected)
	}
}

func Test_keygenKeysetIntegration(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName := "integration-test"

	err := runKeygen(keyName)
	if err != nil {
		t.Fatalf("runKeygen() error = %v", err)
	}

	keyPath := filepath.Join(tempDir, configDir, keyName+".ok")
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read generated key: %v", err)
	}

	encodedKey := base64.StdEncoding.EncodeToString(keyData)

	if err := os.Remove(keyPath); err != nil {
		t.Fatalf("failed to remove key file: %v", err)
	}

	err = runKeyset(encodedKey)
	if err != nil {
		t.Fatalf("runKeyset() error = %v", err)
	}

	importedKeyData, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read imported key: %v", err)
	}

	if string(keyData) != string(importedKeyData) {
		t.Error("imported key data does not match original")
	}

	embeddedName := string(importedKeyData[baseKeySize:])
	if embeddedName != keyName {
		t.Errorf("imported key name = %q, want %q", embeddedName, keyName)
	}
}
