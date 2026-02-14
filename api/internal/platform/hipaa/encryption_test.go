package hipaa

import (
	"crypto/rand"
	"testing"
)

func generateTestKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generate test key: %v", err)
	}
	return key
}

func TestNewPHIEncryptor(t *testing.T) {
	t.Run("valid 32-byte key", func(t *testing.T) {
		key := generateTestKey(t)
		enc, err := NewPHIEncryptor(key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if enc == nil {
			t.Fatal("expected non-nil encryptor")
		}
	})

	t.Run("key too short", func(t *testing.T) {
		key := make([]byte, 16)
		_, err := NewPHIEncryptor(key)
		if err == nil {
			t.Fatal("expected error for 16-byte key")
		}
	})

	t.Run("key too long", func(t *testing.T) {
		key := make([]byte, 64)
		_, err := NewPHIEncryptor(key)
		if err == nil {
			t.Fatal("expected error for 64-byte key")
		}
	})

	t.Run("empty key", func(t *testing.T) {
		_, err := NewPHIEncryptor([]byte{})
		if err == nil {
			t.Fatal("expected error for empty key")
		}
	})
}

func TestEncryptDecrypt(t *testing.T) {
	key := generateTestKey(t)
	enc, err := NewPHIEncryptor(key)
	if err != nil {
		t.Fatalf("create encryptor: %v", err)
	}

	cases := []string{
		"John Doe",
		"SSN: 123-45-6789",
		"Patient has a history of hypertension and diabetes mellitus type 2.",
		"MRN-00012345",
		"\x00\x01\x02binary data\xff\xfe",
	}

	for _, plaintext := range cases {
		t.Run(plaintext, func(t *testing.T) {
			ciphertext, err := enc.Encrypt(plaintext)
			if err != nil {
				t.Fatalf("encrypt: %v", err)
			}

			if ciphertext == plaintext {
				t.Fatal("ciphertext should differ from plaintext")
			}

			decrypted, err := enc.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("decrypt: %v", err)
			}

			if decrypted != plaintext {
				t.Errorf("roundtrip failed: got %q, want %q", decrypted, plaintext)
			}
		})
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	key := generateTestKey(t)
	enc, err := NewPHIEncryptor(key)
	if err != nil {
		t.Fatalf("create encryptor: %v", err)
	}

	plaintext := "Patient: Jane Smith, DOB: 1985-03-15"
	ct1, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt 1: %v", err)
	}

	ct2, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encrypt 2: %v", err)
	}

	if ct1 == ct2 {
		t.Error("encrypting same plaintext twice should produce different ciphertexts due to unique nonces")
	}

	// Both should still decrypt to the same plaintext.
	d1, _ := enc.Decrypt(ct1)
	d2, _ := enc.Decrypt(ct2)
	if d1 != plaintext || d2 != plaintext {
		t.Error("both ciphertexts should decrypt to the original plaintext")
	}
}

func TestDecryptInvalidInput(t *testing.T) {
	key := generateTestKey(t)
	enc, err := NewPHIEncryptor(key)
	if err != nil {
		t.Fatalf("create encryptor: %v", err)
	}

	t.Run("not base64", func(t *testing.T) {
		_, err := enc.Decrypt("not-valid-base64!!!")
		if err == nil {
			t.Fatal("expected error for invalid base64")
		}
	})

	t.Run("too short ciphertext", func(t *testing.T) {
		_, err := enc.Decrypt("AQID") // 3 bytes, shorter than nonce
		if err == nil {
			t.Fatal("expected error for short ciphertext")
		}
	})

	t.Run("corrupted ciphertext", func(t *testing.T) {
		ciphertext, err := enc.Encrypt("sensitive data")
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}

		// Flip a byte in the ciphertext to corrupt it.
		corrupted := []byte(ciphertext)
		if len(corrupted) > 10 {
			corrupted[10] ^= 0xff
		}

		_, err = enc.Decrypt(string(corrupted))
		if err == nil {
			t.Fatal("expected error for corrupted ciphertext")
		}
	})

	t.Run("wrong key", func(t *testing.T) {
		ciphertext, err := enc.Encrypt("secret PHI data")
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}

		otherKey := generateTestKey(t)
		otherEnc, err := NewPHIEncryptor(otherKey)
		if err != nil {
			t.Fatalf("create other encryptor: %v", err)
		}

		_, err = otherEnc.Decrypt(ciphertext)
		if err == nil {
			t.Fatal("expected error when decrypting with wrong key")
		}
	})
}

func TestEmptyString(t *testing.T) {
	key := generateTestKey(t)
	enc, err := NewPHIEncryptor(key)
	if err != nil {
		t.Fatalf("create encryptor: %v", err)
	}

	ciphertext, err := enc.Encrypt("")
	if err != nil {
		t.Fatalf("encrypt empty string: %v", err)
	}

	if ciphertext == "" {
		t.Fatal("ciphertext for empty string should not be empty (contains nonce + auth tag)")
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decrypt empty string: %v", err)
	}

	if decrypted != "" {
		t.Errorf("expected empty string, got %q", decrypted)
	}
}
