package hipaa

import (
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog"
)

// EncryptionService provides field-level PHI encryption for the application.
// It wraps a FieldEncryptor and adds a disabled mode for development environments
// where no encryption key is configured.
type EncryptionService struct {
	encryptor FieldEncryptor
	enabled   bool
}

// NewEncryptionService creates a new encryption service.
//
// If key is empty, encryption is disabled (development mode) and a warning is
// logged. All Encrypt/Decrypt calls become no-ops that return the value as-is.
//
// If key is non-empty, it must be a valid 64-character hex string encoding a
// 32-byte AES-256 key. An invalid key causes a fatal error so the application
// refuses to start with a misconfigured key.
func NewEncryptionService(key string, logger zerolog.Logger) (*EncryptionService, error) {
	if key == "" {
		logger.Warn().Msg("PHI encryption disabled: HIPAA_ENCRYPTION_KEY is not set")
		return &EncryptionService{
			encryptor: nil,
			enabled:   false,
		}, nil
	}

	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("HIPAA_ENCRYPTION_KEY is not valid hex: %w", err)
	}

	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("HIPAA_ENCRYPTION_KEY must be 32 bytes (64 hex chars), got %d bytes", len(keyBytes))
	}

	enc, err := NewPHIEncryptor(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("create PHI encryptor: %w", err)
	}

	logger.Info().Msg("PHI field-level encryption enabled")
	return &EncryptionService{
		encryptor: enc,
		enabled:   true,
	}, nil
}

// Encryptor returns the underlying FieldEncryptor, or nil if encryption is
// disabled. This is useful for passing the encryptor to repositories that
// accept an optional FieldEncryptor.
func (s *EncryptionService) Encryptor() FieldEncryptor {
	return s.encryptor
}

// EncryptField encrypts a single PHI field value. Returns the original value
// unchanged if encryption is disabled.
func (s *EncryptionService) EncryptField(value string) (string, error) {
	if !s.enabled {
		return value, nil
	}
	return s.encryptor.Encrypt(value)
}

// DecryptField decrypts a single PHI field value. Returns the original value
// unchanged if encryption is disabled.
func (s *EncryptionService) DecryptField(value string) (string, error) {
	if !s.enabled {
		return value, nil
	}
	return s.encryptor.Decrypt(value)
}

// IsEnabled returns true if encryption is active.
func (s *EncryptionService) IsEnabled() bool {
	return s.enabled
}
