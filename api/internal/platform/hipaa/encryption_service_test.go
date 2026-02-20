package hipaa

import (
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// validHexKey returns a deterministic 64-char hex string encoding 32 random
// bytes suitable for test use.
func validHexKey(t *testing.T) string {
	t.Helper()
	key := generateTestKey(t) // from encryption_test.go
	return hex.EncodeToString(key)
}

// --- NewEncryptionService ---------------------------------------------------

func TestNewEncryptionService_ValidKey(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	hexKey := validHexKey(t)

	svc, err := NewEncryptionService(hexKey, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if !svc.IsEnabled() {
		t.Fatal("expected encryption to be enabled with a valid key")
	}
	if svc.Encryptor() == nil {
		t.Fatal("expected non-nil encryptor when enabled")
	}
}

func TestNewEncryptionService_EmptyKey(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	svc, err := NewEncryptionService("", logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.IsEnabled() {
		t.Fatal("expected encryption to be disabled with empty key")
	}
	if svc.Encryptor() != nil {
		t.Fatal("expected nil encryptor when disabled")
	}
}

func TestNewEncryptionService_InvalidHex(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	_, err := NewEncryptionService("not-valid-hex!", logger)
	if err == nil {
		t.Fatal("expected error for invalid hex key")
	}
	if !strings.Contains(err.Error(), "not valid hex") {
		t.Errorf("error should mention invalid hex, got: %v", err)
	}
}

func TestNewEncryptionService_WrongLength(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	// 16 bytes = 32 hex chars, but we need 32 bytes = 64 hex chars
	shortKey := hex.EncodeToString(make([]byte, 16))
	_, err := NewEncryptionService(shortKey, logger)
	if err == nil {
		t.Fatal("expected error for 16-byte key")
	}
	if !strings.Contains(err.Error(), "32 bytes") {
		t.Errorf("error should mention 32 bytes, got: %v", err)
	}
}

func TestNewEncryptionService_TooLong(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	longKey := hex.EncodeToString(make([]byte, 64))
	_, err := NewEncryptionService(longKey, logger)
	if err == nil {
		t.Fatal("expected error for 64-byte key")
	}
	if !strings.Contains(err.Error(), "32 bytes") {
		t.Errorf("error should mention 32 bytes, got: %v", err)
	}
}

// --- EncryptField / DecryptField round-trip ---------------------------------

func TestEncryptDecryptField_RoundTrip(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	hexKey := validHexKey(t)

	svc, err := NewEncryptionService(hexKey, logger)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	cases := []string{
		"123-45-6789",
		"patient@example.com",
		"+1 (555) 867-5309",
		"123 Main Street, Apt 4B",
		"",
	}

	for _, original := range cases {
		t.Run(original, func(t *testing.T) {
			encrypted, err := svc.EncryptField(original)
			if err != nil {
				t.Fatalf("encrypt: %v", err)
			}

			if original != "" && encrypted == original {
				t.Error("encrypted value should differ from original")
			}

			decrypted, err := svc.DecryptField(encrypted)
			if err != nil {
				t.Fatalf("decrypt: %v", err)
			}

			if decrypted != original {
				t.Errorf("round-trip failed: got %q, want %q", decrypted, original)
			}
		})
	}
}

func TestEncryptField_ProducesDifferentCiphertexts(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	hexKey := validHexKey(t)

	svc, err := NewEncryptionService(hexKey, logger)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	value := "555-12-3456"
	ct1, _ := svc.EncryptField(value)
	ct2, _ := svc.EncryptField(value)

	if ct1 == ct2 {
		t.Error("encrypting the same value twice should produce different ciphertexts (unique nonces)")
	}
}

// --- Disabled mode ----------------------------------------------------------

func TestDisabledMode_ReturnsValuesUnchanged(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	svc, err := NewEncryptionService("", logger)
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	values := []string{
		"SSN: 123-45-6789",
		"patient@example.com",
		"+1 555 867 5309",
		"",
	}

	for _, v := range values {
		encrypted, err := svc.EncryptField(v)
		if err != nil {
			t.Fatalf("encrypt disabled: %v", err)
		}
		if encrypted != v {
			t.Errorf("disabled encrypt: got %q, want %q", encrypted, v)
		}

		decrypted, err := svc.DecryptField(v)
		if err != nil {
			t.Fatalf("decrypt disabled: %v", err)
		}
		if decrypted != v {
			t.Errorf("disabled decrypt: got %q, want %q", decrypted, v)
		}
	}
}

// --- IsEnabled --------------------------------------------------------------

func TestIsEnabled_Enabled(t *testing.T) {
	logger := zerolog.New(os.Stderr)
	hexKey := validHexKey(t)

	svc, _ := NewEncryptionService(hexKey, logger)
	if !svc.IsEnabled() {
		t.Error("expected IsEnabled() == true with valid key")
	}
}

func TestIsEnabled_Disabled(t *testing.T) {
	logger := zerolog.New(os.Stderr)

	svc, _ := NewEncryptionService("", logger)
	if svc.IsEnabled() {
		t.Error("expected IsEnabled() == false with empty key")
	}
}
