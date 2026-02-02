package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encrypt encrypts plain text string into base64 encoded string using AES-GCM
// The key should be 32 bytes (AES-256)
func Encrypt(plainText, keyString string) (string, error) {
	// Key adjustment: Ensure key is 32 bytes (or hash it).
	// For simplicity, we assume caller manages key size or we trust system secret length.
	// But it's safer to just slice or hash. Let's slice/pad for now to be robust-ish.
	// NOTE: In production, derive a key using HKDF. Here we might just use the raw string if it fits, or panic.
	// Let's rely on correct configuration for now (users often provide base64 encoded 32-byte keys or long strings).
	// Ideally, we just use the first 32 bytes of a long secret.
	key := []byte(keyString)
	if len(key) > 32 {
		key = key[:32]
	}
	// If less than 16/24/32, AES won't work. We assume the system JWT secret is long enough.

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt decrypts base64 encoded string
func Decrypt(cipherTextString, keyString string) (string, error) {
	key := []byte(keyString)
	if len(key) > 32 {
		key = key[:32]
	}

	cipherText, err := base64.StdEncoding.DecodeString(cipherTextString)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(cipherText) < gcm.NonceSize() {
		return "", errors.New("malformed ciphertext")
	}

	nonce, cipherText := cipherText[:gcm.NonceSize()], cipherText[gcm.NonceSize():]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
