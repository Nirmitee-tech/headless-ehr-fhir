package hipaa

import (
	"strings"
	"testing"
)

func TestRotatingEncryptor_EncryptDecryptCurrentKey(t *testing.T) {
	key := generateTestKey(t)
	re, err := NewRotatingEncryptor(key, 1)
	if err != nil {
		t.Fatalf("create rotating encryptor: %v", err)
	}

	plaintext := "Patient: John Doe, SSN: 123-45-6789"
	ciphertext, err := re.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	// Verify version prefix is present
	if !strings.HasPrefix(ciphertext, "v1:") {
		t.Errorf("expected ciphertext to start with 'v1:', got %q", ciphertext[:10])
	}

	decrypted, err := re.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("roundtrip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestRotatingEncryptor_DecryptWithPreviousKey(t *testing.T) {
	oldKey := generateTestKey(t)
	newKey := generateTestKey(t)

	// Create encryptor with old key (version 1)
	oldEnc, err := NewRotatingEncryptor(oldKey, 1)
	if err != nil {
		t.Fatalf("create old encryptor: %v", err)
	}

	// Encrypt data with old key
	plaintext := "Sensitive medical record"
	oldCiphertext, err := oldEnc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt with old key: %v", err)
	}

	// Create new encryptor with new key (version 2) and add old key as previous
	newEnc, err := NewRotatingEncryptor(newKey, 2)
	if err != nil {
		t.Fatalf("create new encryptor: %v", err)
	}
	if err := newEnc.AddPreviousKey(oldKey, 1); err != nil {
		t.Fatalf("add previous key: %v", err)
	}

	// Decrypt data encrypted with old key using new encryptor
	decrypted, err := newEnc.Decrypt(oldCiphertext)
	if err != nil {
		t.Fatalf("decrypt with previous key: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestRotatingEncryptor_DecryptUnknownVersion(t *testing.T) {
	key := generateTestKey(t)
	re, err := NewRotatingEncryptor(key, 2)
	if err != nil {
		t.Fatalf("create encryptor: %v", err)
	}

	// Try to decrypt a ciphertext with version 99 (no key registered)
	_, err = re.Decrypt("v99:someciphertext")
	if err == nil {
		t.Fatal("expected error for unknown key version")
	}
}

func TestRotatingEncryptor_NeedsReEncryption(t *testing.T) {
	oldKey := generateTestKey(t)
	newKey := generateTestKey(t)

	// Create encryptor with old key (version 1), encrypt some data
	oldEnc, err := NewRotatingEncryptor(oldKey, 1)
	if err != nil {
		t.Fatalf("create old encryptor: %v", err)
	}
	oldCiphertext, err := oldEnc.Encrypt("test data")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	// Create new encryptor with version 2
	newEnc, err := NewRotatingEncryptor(newKey, 2)
	if err != nil {
		t.Fatalf("create new encryptor: %v", err)
	}
	if err := newEnc.AddPreviousKey(oldKey, 1); err != nil {
		t.Fatalf("add previous key: %v", err)
	}

	// Old ciphertext needs re-encryption
	if !newEnc.NeedsReEncryption(oldCiphertext) {
		t.Error("expected old ciphertext to need re-encryption")
	}

	// New ciphertext does not need re-encryption
	newCiphertext, err := newEnc.Encrypt("new data")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if newEnc.NeedsReEncryption(newCiphertext) {
		t.Error("expected new ciphertext to not need re-encryption")
	}
}

func TestRotatingEncryptor_NeedsReEncryption_LegacyData(t *testing.T) {
	key := generateTestKey(t)
	re, err := NewRotatingEncryptor(key, 1)
	if err != nil {
		t.Fatalf("create encryptor: %v", err)
	}

	// Legacy data without version prefix always needs re-encryption
	enc, _ := NewPHIEncryptor(key)
	legacyCiphertext, err := enc.Encrypt("legacy data")
	if err != nil {
		t.Fatalf("encrypt legacy: %v", err)
	}

	if !re.NeedsReEncryption(legacyCiphertext) {
		t.Error("expected legacy ciphertext (no version prefix) to need re-encryption")
	}
}

func TestRotatingEncryptor_ReEncrypt(t *testing.T) {
	oldKey := generateTestKey(t)
	newKey := generateTestKey(t)

	// Create encryptor with old key, encrypt data
	oldEnc, err := NewRotatingEncryptor(oldKey, 1)
	if err != nil {
		t.Fatalf("create old encryptor: %v", err)
	}
	plaintext := "PHI that needs rotation"
	oldCiphertext, err := oldEnc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	// Create new encryptor with version 2, add old key
	newEnc, err := NewRotatingEncryptor(newKey, 2)
	if err != nil {
		t.Fatalf("create new encryptor: %v", err)
	}
	if err := newEnc.AddPreviousKey(oldKey, 1); err != nil {
		t.Fatalf("add previous key: %v", err)
	}

	// Re-encrypt
	newCiphertext, err := newEnc.ReEncrypt(oldCiphertext)
	if err != nil {
		t.Fatalf("re-encrypt: %v", err)
	}

	// New ciphertext should have version 2
	if !strings.HasPrefix(newCiphertext, "v2:") {
		t.Errorf("expected re-encrypted ciphertext to start with 'v2:', got prefix %q", newCiphertext[:5])
	}

	// Should no longer need re-encryption
	if newEnc.NeedsReEncryption(newCiphertext) {
		t.Error("expected re-encrypted ciphertext to not need re-encryption")
	}

	// Should still decrypt to the original plaintext
	decrypted, err := newEnc.Decrypt(newCiphertext)
	if err != nil {
		t.Fatalf("decrypt re-encrypted: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestRotatingEncryptor_LegacyDecrypt(t *testing.T) {
	key := generateTestKey(t)

	// Encrypt with plain PHIEncryptor (no version prefix)
	enc, err := NewPHIEncryptor(key)
	if err != nil {
		t.Fatalf("create PHI encryptor: %v", err)
	}
	plaintext := "legacy encrypted data"
	legacyCiphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt legacy: %v", err)
	}

	// Create rotating encryptor with the same key
	re, err := NewRotatingEncryptor(key, 1)
	if err != nil {
		t.Fatalf("create rotating encryptor: %v", err)
	}

	// Should be able to decrypt legacy data (falls back to current key)
	decrypted, err := re.Decrypt(legacyCiphertext)
	if err != nil {
		t.Fatalf("decrypt legacy: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestRotatingEncryptor_CurrentVersion(t *testing.T) {
	key := generateTestKey(t)
	re, err := NewRotatingEncryptor(key, 42)
	if err != nil {
		t.Fatalf("create encryptor: %v", err)
	}

	if re.CurrentVersion() != 42 {
		t.Errorf("expected current version 42, got %d", re.CurrentVersion())
	}
}

func TestNewRotatingEncryptor_InvalidKey(t *testing.T) {
	_, err := NewRotatingEncryptor([]byte("short"), 1)
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestRotatingEncryptor_AddPreviousKey_InvalidKey(t *testing.T) {
	key := generateTestKey(t)
	re, err := NewRotatingEncryptor(key, 1)
	if err != nil {
		t.Fatalf("create encryptor: %v", err)
	}

	err = re.AddPreviousKey([]byte("short"), 0)
	if err == nil {
		t.Fatal("expected error for invalid previous key")
	}
}

func TestParseVersionedCiphertext(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantVer    int
		wantData   string
		wantErr    bool
	}{
		{"valid v1", "v1:data", 1, "data", false},
		{"valid v2", "v2:encrypted_data_here", 2, "encrypted_data_here", false},
		{"valid v99", "v99:data", 99, "data", false},
		{"no prefix", "data_without_prefix", 0, "", true},
		{"no separator", "v1data", 0, "", true},
		{"invalid version", "vX:data", 0, "", true},
		{"empty string", "", 0, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, data, err := parseVersionedCiphertext(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ver != tt.wantVer {
				t.Errorf("version: got %d, want %d", ver, tt.wantVer)
			}
			if data != tt.wantData {
				t.Errorf("data: got %q, want %q", data, tt.wantData)
			}
		})
	}
}
