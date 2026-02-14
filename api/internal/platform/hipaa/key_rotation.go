package hipaa

import (
	"fmt"
	"strings"
	"sync"
)

// KeyVersion prefix format: "v{version}:" prepended to ciphertext
const keyVersionPrefix = "v"
const keyVersionSeparator = ":"

// RotatingEncryptor supports encryption key rotation with versioned keys.
type RotatingEncryptor struct {
	mu         sync.RWMutex
	current    *PHIEncryptor
	currentVer int
	previous   map[int]*PHIEncryptor
}

// NewRotatingEncryptor creates a new rotating encryptor with the current key.
func NewRotatingEncryptor(currentKey []byte, currentVersion int) (*RotatingEncryptor, error) {
	enc, err := NewPHIEncryptor(currentKey)
	if err != nil {
		return nil, fmt.Errorf("rotating encryptor: current key: %w", err)
	}
	return &RotatingEncryptor{
		current:    enc,
		currentVer: currentVersion,
		previous:   make(map[int]*PHIEncryptor),
	}, nil
}

// AddPreviousKey adds a previous encryption key for decryption.
func (r *RotatingEncryptor) AddPreviousKey(key []byte, version int) error {
	enc, err := NewPHIEncryptor(key)
	if err != nil {
		return fmt.Errorf("rotating encryptor: previous key v%d: %w", version, err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.previous[version] = enc
	return nil
}

// Encrypt encrypts with the current key and prepends the version prefix.
func (r *RotatingEncryptor) Encrypt(plaintext string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ciphertext, err := r.current.Encrypt(plaintext)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%d%s%s", keyVersionPrefix, r.currentVer, keyVersionSeparator, ciphertext), nil
}

// Decrypt detects the key version and decrypts with the appropriate key.
func (r *RotatingEncryptor) Decrypt(ciphertext string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	version, data, err := parseVersionedCiphertext(ciphertext)
	if err != nil {
		// No version prefix - try current key (legacy data)
		return r.current.Decrypt(ciphertext)
	}

	if version == r.currentVer {
		return r.current.Decrypt(data)
	}

	enc, ok := r.previous[version]
	if !ok {
		return "", fmt.Errorf("no key available for version %d", version)
	}
	return enc.Decrypt(data)
}

// NeedsReEncryption checks if a ciphertext uses an old key version.
func (r *RotatingEncryptor) NeedsReEncryption(ciphertext string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	version, _, err := parseVersionedCiphertext(ciphertext)
	if err != nil {
		return true // No version prefix = legacy data
	}
	return version != r.currentVer
}

// ReEncrypt decrypts with the old key and re-encrypts with the current key.
func (r *RotatingEncryptor) ReEncrypt(ciphertext string) (string, error) {
	plaintext, err := r.Decrypt(ciphertext)
	if err != nil {
		return "", fmt.Errorf("re-encrypt: decrypt: %w", err)
	}
	return r.Encrypt(plaintext)
}

// CurrentVersion returns the current key version.
func (r *RotatingEncryptor) CurrentVersion() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentVer
}

func parseVersionedCiphertext(s string) (int, string, error) {
	if !strings.HasPrefix(s, keyVersionPrefix) {
		return 0, "", fmt.Errorf("no version prefix")
	}

	idx := strings.Index(s, keyVersionSeparator)
	if idx < 0 {
		return 0, "", fmt.Errorf("no version separator")
	}

	versionStr := s[len(keyVersionPrefix):idx]
	var version int
	_, err := fmt.Sscanf(versionStr, "%d", &version)
	if err != nil {
		return 0, "", fmt.Errorf("invalid version: %w", err)
	}

	return version, s[idx+1:], nil
}
