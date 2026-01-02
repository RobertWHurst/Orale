package orale

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

func Test_isEncrypted_validEncrypted(t *testing.T) {
	if !isEncrypted("ENC[aGVsbG8=]") {
		t.Error("isEncrypted() should return true for valid encrypted value")
	}
}

func Test_isEncrypted_plainValue(t *testing.T) {
	if isEncrypted("plain-text") {
		t.Error("isEncrypted() should return false for plain text")
	}
}

func Test_isEncrypted_partialMatch(t *testing.T) {
	if isEncrypted("ENC[abc") {
		t.Error("isEncrypted() should return false for partial match")
	}
}

func Test_isEncrypted_noBrackets(t *testing.T) {
	if isEncrypted("ENC") {
		t.Error("isEncrypted() should return false when no brackets")
	}
}

func Test_isEncrypted_empty(t *testing.T) {
	if isEncrypted("") {
		t.Error("isEncrypted() should return false for empty string")
	}
}

func Test_encryptDecryptRoundtrip(t *testing.T) {
	salt := make([]byte, saltSize)
	encryptionKey := make([]byte, keySize)
	hmacKey := make([]byte, hmacSize)

	if _, err := rand.Read(salt); err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}
	if _, err := rand.Read(encryptionKey); err != nil {
		t.Fatalf("failed to generate encryption key: %v", err)
	}
	if _, err := rand.Read(hmacKey); err != nil {
		t.Fatalf("failed to generate HMAC key: %v", err)
	}

	plaintext := "my-secret-password"

	encrypted, err := Encrypt([]byte(plaintext), encryptionKey, hmacKey)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	decrypted, err := decrypt(encrypted, encryptionKey, hmacKey)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("roundtrip failed: got %q, want %q", decrypted, plaintext)
	}
}

func Test_decryptValue_multipleKeys(t *testing.T) {
	tempDir := t.TempDir()
	home := os.Getenv("HOME")
	defer os.Setenv("HOME", home)
	os.Setenv("HOME", tempDir)

	configPath := filepath.Join(tempDir, ".config", "orale")
	if err := os.MkdirAll(configPath, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	keyData1 := make([]byte, baseKeySize)
	keyData2 := make([]byte, baseKeySize)
	keyData3 := make([]byte, baseKeySize)

	if _, err := rand.Read(keyData1); err != nil {
		t.Fatalf("failed to generate key1: %v", err)
	}
	if _, err := rand.Read(keyData2); err != nil {
		t.Fatalf("failed to generate key2: %v", err)
	}
	if _, err := rand.Read(keyData3); err != nil {
		t.Fatalf("failed to generate key3: %v", err)
	}

	keyData1 = append(keyData1, []byte("key1")...)
	keyData2 = append(keyData2, []byte("key2")...)
	keyData3 = append(keyData3, []byte("key3")...)

	if err := os.WriteFile(filepath.Join(configPath, "key1.ok"), keyData1, 0600); err != nil {
		t.Fatalf("failed to write key1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configPath, "key2.ok"), keyData2, 0600); err != nil {
		t.Fatalf("failed to write key2: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configPath, "key3.ok"), keyData3, 0600); err != nil {
		t.Fatalf("failed to write key3: %v", err)
	}

	plaintext := "secret-value"
	encryptionKey := keyData2[saltSize : saltSize+keySize]
	hmacKey := keyData2[saltSize+keySize : baseKeySize]

	encrypted, err := Encrypt([]byte(plaintext), encryptionKey, hmacKey)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	encryptedValue := "ENC[" + encrypted + "]"

	decrypted, err := decryptValue(encryptedValue)
	if err != nil {
		t.Fatalf("decryptValue failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("decryption failed: got %q, want %q", decrypted, plaintext)
	}
}

func Test_decryptValue_noKeys(t *testing.T) {
	tempDir := t.TempDir()
	home := os.Getenv("HOME")
	defer os.Setenv("HOME", home)
	os.Setenv("HOME", tempDir)

	encryptedValue := "ENC[aGVsbG8=]"

	_, err := decryptValue(encryptedValue)
	if err == nil {
		t.Error("expected error when no keys are available, got nil")
	}
}

func Test_decryptValue_invalidHMAC(t *testing.T) {
	tempDir := t.TempDir()
	home := os.Getenv("HOME")
	defer os.Setenv("HOME", home)
	os.Setenv("HOME", tempDir)

	configPath := filepath.Join(tempDir, ".config", "orale")
	if err := os.MkdirAll(configPath, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	keyData := make([]byte, baseKeySize)
	if _, err := rand.Read(keyData); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	keyData = append(keyData, []byte("test")...)

	if err := os.WriteFile(filepath.Join(configPath, "test.ok"), keyData, 0600); err != nil {
		t.Fatalf("failed to write key: %v", err)
	}

	differentKeyData := make([]byte, baseKeySize)
	if _, err := rand.Read(differentKeyData); err != nil {
		t.Fatalf("failed to generate different key: %v", err)
	}

	plaintext := "secret-value"
	encryptionKey := differentKeyData[saltSize : saltSize+keySize]
	hmacKey := differentKeyData[saltSize+keySize : baseKeySize]

	encrypted, err := Encrypt([]byte(plaintext), encryptionKey, hmacKey)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	encryptedValue := "ENC[" + encrypted + "]"

	_, err = decryptValue(encryptedValue)
	if err == nil {
		t.Error("expected error when HMAC doesn't match any key, got nil")
	}
}
