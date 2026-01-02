package commands

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func Test_runKeyset_importsValidKey(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName := "imported-key"
	keyData := make([]byte, baseKeySize)
	for i := range keyData {
		keyData[i] = byte(i % 256)
	}
	keyData = append(keyData, []byte(keyName)...)

	encodedKey := base64.StdEncoding.EncodeToString(keyData)

	err := runKeyset(encodedKey)
	if err != nil {
		t.Fatalf("runKeyset() error = %v", err)
	}

	keyPath := filepath.Join(tempDir, configDir, keyName+".ok")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Errorf("key file was not created at %s", keyPath)
	}

	importedData, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read imported key: %v", err)
	}

	if string(importedData) != string(keyData) {
		t.Error("imported key data does not match original")
	}
}

func Test_runKeyset_extractsKeyName(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName := "production"
	keyData := make([]byte, baseKeySize)
	keyData = append(keyData, []byte(keyName)...)
	encodedKey := base64.StdEncoding.EncodeToString(keyData)

	err := runKeyset(encodedKey)
	if err != nil {
		t.Fatalf("runKeyset() error = %v", err)
	}

	expectedPath := filepath.Join(tempDir, configDir, keyName+".ok")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("key file was not created at expected path: %s", expectedPath)
	}
}

func Test_runKeyset_invalidBase64(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	invalidKey := "this is not valid base64!!!"

	err := runKeyset(invalidKey)
	if err == nil {
		t.Error("runKeyset() should return error for invalid base64")
	}
}

func Test_runKeyset_keyTooShort(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	shortKey := make([]byte, 10)
	encodedKey := base64.StdEncoding.EncodeToString(shortKey)

	err := runKeyset(encodedKey)
	if err == nil {
		t.Error("runKeyset() should return error for key that is too short")
	}
}

func Test_runKeyset_emptyKeyName(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyData := make([]byte, baseKeySize)
	encodedKey := base64.StdEncoding.EncodeToString(keyData)

	err := runKeyset(encodedKey)
	if err == nil {
		t.Error("runKeyset() should return error when key name is empty")
	}
}

func Test_runKeyset_handlesWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName := "whitespace-test"
	keyData := make([]byte, baseKeySize)
	keyData = append(keyData, []byte(keyName)...)
	encodedKey := base64.StdEncoding.EncodeToString(keyData)

	encodedKeyWithWhitespace := "  " + encodedKey + "\n\t"

	err := runKeyset(encodedKeyWithWhitespace)
	if err != nil {
		t.Fatalf("runKeyset() should handle whitespace, error = %v", err)
	}

	keyPath := filepath.Join(tempDir, configDir, keyName+".ok")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("key file was not created")
	}
}

func Test_runKeyset_overwritesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName := "overwrite-test"

	keyData1 := make([]byte, baseKeySize)
	for i := range keyData1 {
		keyData1[i] = 1
	}
	keyData1 = append(keyData1, []byte(keyName)...)
	encodedKey1 := base64.StdEncoding.EncodeToString(keyData1)

	err := runKeyset(encodedKey1)
	if err != nil {
		t.Fatalf("first runKeyset() error = %v", err)
	}

	keyData2 := make([]byte, baseKeySize)
	for i := range keyData2 {
		keyData2[i] = 2
	}
	keyData2 = append(keyData2, []byte(keyName)...)
	encodedKey2 := base64.StdEncoding.EncodeToString(keyData2)

	err = runKeyset(encodedKey2)
	if err != nil {
		t.Fatalf("second runKeyset() error = %v", err)
	}

	keyPath := filepath.Join(tempDir, configDir, keyName+".ok")
	finalData, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read final key: %v", err)
	}

	if string(finalData) != string(keyData2) {
		t.Error("key file was not overwritten with new data")
	}
}

func Test_runKeyset_setsCorrectPermissions(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	keyName := "perms-test"
	keyData := make([]byte, baseKeySize)
	keyData = append(keyData, []byte(keyName)...)
	encodedKey := base64.StdEncoding.EncodeToString(keyData)

	err := runKeyset(encodedKey)
	if err != nil {
		t.Fatalf("runKeyset() error = %v", err)
	}

	keyPath := filepath.Join(tempDir, configDir, keyName+".ok")
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("failed to stat key file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("key file permissions = %o, want 0600", perm)
	}
}
