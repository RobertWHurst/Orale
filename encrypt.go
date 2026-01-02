package orale

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const ivSize = 16 // AES block size for CBC mode

// Encrypt encrypts plaintext using AES-256-CBC with HMAC-SHA256 authentication.
func Encrypt(plaintext, encryptionKey, hmacKey []byte) (string, error) {
	plaintext = pkcs7Pad(plaintext, aes.BlockSize)

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	iv := make([]byte, ivSize)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}

	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	h := hmac.New(sha256.New, hmacKey)
	h.Write(iv)
	h.Write(ciphertext)
	mac := h.Sum(nil)

	result := make([]byte, 0, len(iv)+len(ciphertext)+len(mac))
	result = append(result, iv...)
	result = append(result, ciphertext...)
	result = append(result, mac...)

	return base64.StdEncoding.EncodeToString(result), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}
