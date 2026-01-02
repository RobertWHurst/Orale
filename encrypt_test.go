package orale

import (
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func Test_Encrypt_successfully(t *testing.T) {
	encryptionKey := make([]byte, 32)
	hmacKey := make([]byte, 32)

	if _, err := rand.Read(encryptionKey); err != nil {
		t.Fatalf("failed to generate encryption key: %v", err)
	}
	if _, err := rand.Read(hmacKey); err != nil {
		t.Fatalf("failed to generate HMAC key: %v", err)
	}

	plaintext := []byte("test-secret-value")

	encrypted, err := Encrypt(plaintext, encryptionKey, hmacKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if encrypted == "" {
		t.Error("Encrypt() returned empty string")
	}

	_, err = base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Errorf("Encrypt() output is not valid base64: %v", err)
	}
}

func Test_Encrypt_differentOutputsForSameInput(t *testing.T) {
	encryptionKey := make([]byte, 32)
	hmacKey := make([]byte, 32)

	if _, err := rand.Read(encryptionKey); err != nil {
		t.Fatalf("failed to generate encryption key: %v", err)
	}
	if _, err := rand.Read(hmacKey); err != nil {
		t.Fatalf("failed to generate HMAC key: %v", err)
	}

	plaintext := []byte("same-input")

	encrypted1, err := Encrypt(plaintext, encryptionKey, hmacKey)
	if err != nil {
		t.Fatalf("first Encrypt() error = %v", err)
	}

	encrypted2, err := Encrypt(plaintext, encryptionKey, hmacKey)
	if err != nil {
		t.Fatalf("second Encrypt() error = %v", err)
	}

	if encrypted1 == encrypted2 {
		t.Error("Encrypt() produced same output for same input (IV should be random)")
	}
}

func Test_Encrypt_emptyInput(t *testing.T) {
	encryptionKey := make([]byte, 32)
	hmacKey := make([]byte, 32)

	if _, err := rand.Read(encryptionKey); err != nil {
		t.Fatalf("failed to generate encryption key: %v", err)
	}
	if _, err := rand.Read(hmacKey); err != nil {
		t.Fatalf("failed to generate HMAC key: %v", err)
	}

	plaintext := []byte("")

	encrypted, err := Encrypt(plaintext, encryptionKey, hmacKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if encrypted == "" {
		t.Error("Encrypt() returned empty string for empty input")
	}

	_, err = base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Errorf("Encrypt() output is not valid base64: %v", err)
	}
}

func Test_Encrypt_invalidKeySize(t *testing.T) {
	hmacKey := make([]byte, 32)
	if _, err := rand.Read(hmacKey); err != nil {
		t.Fatalf("failed to generate HMAC key: %v", err)
	}

	plaintext := []byte("test")
	invalidKey := []byte{1, 2, 3}

	_, err := Encrypt(plaintext, invalidKey, hmacKey)
	if err == nil {
		t.Error("Encrypt() should fail with invalid key size")
	}
}

func Test_Encrypt_outputStructure(t *testing.T) {
	encryptionKey := make([]byte, 32)
	hmacKey := make([]byte, 32)

	if _, err := rand.Read(encryptionKey); err != nil {
		t.Fatalf("failed to generate encryption key: %v", err)
	}
	if _, err := rand.Read(hmacKey); err != nil {
		t.Fatalf("failed to generate HMAC key: %v", err)
	}

	plaintext := []byte("test-value")

	encrypted, err := Encrypt(plaintext, encryptionKey, hmacKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("failed to decode base64: %v", err)
	}

	minSize := 16 + 16 + 32
	if len(decoded) < minSize {
		t.Errorf("encrypted output too short: got %d bytes, expected at least %d", len(decoded), minSize)
	}

	if len(decoded) < 32 {
		t.Error("output doesn't contain HMAC")
	}
}

func Test_Encrypt_unicodeInput(t *testing.T) {
	encryptionKey := make([]byte, 32)
	hmacKey := make([]byte, 32)

	if _, err := rand.Read(encryptionKey); err != nil {
		t.Fatalf("failed to generate encryption key: %v", err)
	}
	if _, err := rand.Read(hmacKey); err != nil {
		t.Fatalf("failed to generate HMAC key: %v", err)
	}

	plaintext := []byte("Hello 世界 🔐 مرحبا")

	encrypted, err := Encrypt(plaintext, encryptionKey, hmacKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if encrypted == "" {
		t.Error("Encrypt() returned empty string")
	}
}

func Test_pkcs7Pad_emptyData(t *testing.T) {
	result := pkcs7Pad([]byte{}, 16)

	if len(result) != 16 {
		t.Errorf("pkcs7Pad() length = %d, want 16", len(result))
	}

	paddingLen := int(result[len(result)-1])
	if paddingLen != 16 {
		t.Errorf("padding length = %d, want 16", paddingLen)
	}

	for i := 0; i < len(result); i++ {
		if result[i] != 16 {
			t.Errorf("invalid padding byte at position %d: got %d, want 16", i, result[i])
		}
	}
}

func Test_pkcs7Pad_exactBlockSize(t *testing.T) {
	data := make([]byte, 16)
	result := pkcs7Pad(data, 16)

	if len(result) != 32 {
		t.Errorf("pkcs7Pad() length = %d, want 32", len(result))
	}

	for i := 16; i < 32; i++ {
		if result[i] != 16 {
			t.Errorf("padding byte at %d = %d, want 16", i, result[i])
		}
	}
}

func Test_pkcs7Pad_partialBlock(t *testing.T) {
	data := []byte{1, 2, 3}
	result := pkcs7Pad(data, 16)

	if len(result) != 16 {
		t.Errorf("pkcs7Pad() length = %d, want 16", len(result))
	}

	if len(result)%16 != 0 {
		t.Error("pkcs7Pad() result is not multiple of block size")
	}

	paddingLen := int(result[len(result)-1])
	if paddingLen != 13 {
		t.Errorf("padding length = %d, want 13", paddingLen)
	}

	for i := len(result) - paddingLen; i < len(result); i++ {
		if result[i] != byte(paddingLen) {
			t.Errorf("invalid padding byte at position %d: got %d, want %d", i, result[i], paddingLen)
		}
	}
}
