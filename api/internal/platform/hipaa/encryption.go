package hipaa

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// PHIEncryptor provides AES-256-GCM field-level encryption and decryption for PHI data.
type PHIEncryptor struct {
	aead cipher.AEAD
}

// NewPHIEncryptor creates a new PHIEncryptor with the given 32-byte AES-256 key.
func NewPHIEncryptor(key []byte) (*PHIEncryptor, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("phi encryptor: key must be 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("phi encryptor: create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("phi encryptor: create GCM: %w", err)
	}

	return &PHIEncryptor{aead: aead}, nil
}

// Encrypt encrypts the plaintext string and returns a base64-encoded ciphertext
// with the nonce prepended.
func (e *PHIEncryptor) Encrypt(plaintext string) (string, error) {
	encrypted, err := e.EncryptBytes([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// Decrypt decodes the base64 ciphertext, extracts the prepended nonce, and decrypts.
func (e *PHIEncryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("phi decrypt: base64 decode: %w", err)
	}

	plaintext, err := e.DecryptBytes(data)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// EncryptBytes encrypts the data and returns the nonce prepended to the ciphertext.
func (e *PHIEncryptor) EncryptBytes(data []byte) ([]byte, error) {
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("phi encrypt: generate nonce: %w", err)
	}

	// Seal appends the ciphertext to nonce, so the result is nonce + ciphertext.
	return e.aead.Seal(nonce, nonce, data, nil), nil
}

// DecryptBytes extracts the nonce from the front of data and decrypts the remainder.
func (e *PHIEncryptor) DecryptBytes(data []byte) ([]byte, error) {
	nonceSize := e.aead.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("phi decrypt: ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("phi decrypt: %w", err)
	}
	return plaintext, nil
}
