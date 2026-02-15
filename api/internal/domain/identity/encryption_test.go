package identity

import (
	"testing"

	"github.com/ehr/ehr/internal/platform/hipaa"
)

// strPtr is a helper to get a pointer to a string literal.
func strPtr(s string) *string { return &s }

// newTestEncryptor creates a PHIEncryptor with a fixed 32-byte test key.
func newTestEncryptor(t *testing.T) hipaa.FieldEncryptor {
	t.Helper()
	key := []byte("01234567890123456789012345678901")
	enc, err := hipaa.NewPHIEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create test encryptor: %v", err)
	}
	return enc
}

// newRepoWithEncryptor creates a patientRepoPG with the given encryptor and nil pool.
func newRepoWithEncryptor(enc hipaa.FieldEncryptor) *patientRepoPG {
	return &patientRepoPG{pool: nil, encryptor: enc}
}

// -- encryptField / decryptField tests --

func TestEncryptField_NilEncryptor(t *testing.T) {
	repo := newRepoWithEncryptor(nil)
	original := "sensitive-data"
	val := original

	result, err := repo.encryptField(&val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if *result != original {
		t.Errorf("expected value unchanged %q, got %q", original, *result)
	}
}

func TestEncryptField_NilValue(t *testing.T) {
	enc := newTestEncryptor(t)
	repo := newRepoWithEncryptor(enc)

	result, err := repo.encryptField(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for nil input, got %v", result)
	}
}

func TestEncryptField_EmptyString(t *testing.T) {
	enc := newTestEncryptor(t)
	repo := newRepoWithEncryptor(enc)

	empty := ""
	result, err := repo.encryptField(&empty)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for empty string")
	}
	if *result != "" {
		t.Errorf("expected empty string unchanged, got %q", *result)
	}
}

func TestEncryptField_EncryptsValue(t *testing.T) {
	enc := newTestEncryptor(t)
	repo := newRepoWithEncryptor(enc)

	plaintext := "123-45-6789"
	result, err := repo.encryptField(&plaintext)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if *result == plaintext {
		t.Error("expected encrypted value to differ from plaintext")
	}
	if len(*result) == 0 {
		t.Error("expected non-empty encrypted value")
	}
}

func TestDecryptField_NilEncryptor(t *testing.T) {
	repo := newRepoWithEncryptor(nil)
	original := "some-ciphertext"
	val := original

	result, err := repo.decryptField(&val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if *result != original {
		t.Errorf("expected value unchanged %q, got %q", original, *result)
	}
}

func TestDecryptField_RoundTrip(t *testing.T) {
	enc := newTestEncryptor(t)
	repo := newRepoWithEncryptor(enc)

	plaintext := "my-secret-value"
	encrypted, err := repo.encryptField(&plaintext)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}
	if encrypted == nil {
		t.Fatal("expected non-nil encrypted result")
	}

	decrypted, err := repo.decryptField(encrypted)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}
	if decrypted == nil {
		t.Fatal("expected non-nil decrypted result")
	}
	if *decrypted != plaintext {
		t.Errorf("round-trip failed: expected %q, got %q", plaintext, *decrypted)
	}
}

// -- encryptPatientPHI / decryptPatientPHI tests --

// buildPatientWithAllPHI creates a Patient with all 12 PHI fields populated.
func buildPatientWithAllPHI() *Patient {
	return &Patient{
		FirstName:    "John",
		LastName:     "Doe",
		MRN:          "MRN-ENC",
		SSNHash:      strPtr("123-45-6789"),
		AadhaarHash:  strPtr("1234-5678-9012"),
		PhoneHome:    strPtr("+1-555-0100"),
		PhoneMobile:  strPtr("+1-555-0200"),
		PhoneWork:    strPtr("+1-555-0300"),
		Email:        strPtr("john.doe@example.com"),
		AddressLine1: strPtr("123 Main St"),
		AddressLine2: strPtr("Apt 4B"),
		City:         strPtr("Springfield"),
		District:     strPtr("Sangamon"),
		State:        strPtr("IL"),
		PostalCode:   strPtr("62701"),
	}
}

func TestEncryptPatientPHI_EncryptsAllFields(t *testing.T) {
	enc := newTestEncryptor(t)
	repo := newRepoWithEncryptor(enc)
	p := buildPatientWithAllPHI()

	// Save original values for comparison.
	origSSN := *p.SSNHash
	origAadhaar := *p.AadhaarHash
	origPhoneHome := *p.PhoneHome
	origPhoneMobile := *p.PhoneMobile
	origPhoneWork := *p.PhoneWork
	origEmail := *p.Email
	origAddr1 := *p.AddressLine1
	origAddr2 := *p.AddressLine2
	origCity := *p.City
	origDistrict := *p.District
	origState := *p.State
	origPostal := *p.PostalCode

	err := repo.encryptPatientPHI(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Each field should be non-nil and different from its original plaintext.
	checks := []struct {
		name     string
		field    *string
		original string
	}{
		{"SSNHash", p.SSNHash, origSSN},
		{"AadhaarHash", p.AadhaarHash, origAadhaar},
		{"PhoneHome", p.PhoneHome, origPhoneHome},
		{"PhoneMobile", p.PhoneMobile, origPhoneMobile},
		{"PhoneWork", p.PhoneWork, origPhoneWork},
		{"Email", p.Email, origEmail},
		{"AddressLine1", p.AddressLine1, origAddr1},
		{"AddressLine2", p.AddressLine2, origAddr2},
		{"City", p.City, origCity},
		{"District", p.District, origDistrict},
		{"State", p.State, origState},
		{"PostalCode", p.PostalCode, origPostal},
	}

	for _, c := range checks {
		if c.field == nil {
			t.Errorf("%s: expected non-nil after encryption", c.name)
			continue
		}
		if *c.field == c.original {
			t.Errorf("%s: expected encrypted value to differ from plaintext %q", c.name, c.original)
		}
	}
}

func TestDecryptPatientPHI_DecryptsAllFields(t *testing.T) {
	enc := newTestEncryptor(t)
	repo := newRepoWithEncryptor(enc)
	p := buildPatientWithAllPHI()

	// Save originals.
	originals := map[string]string{
		"SSNHash":      *p.SSNHash,
		"AadhaarHash":  *p.AadhaarHash,
		"PhoneHome":    *p.PhoneHome,
		"PhoneMobile":  *p.PhoneMobile,
		"PhoneWork":    *p.PhoneWork,
		"Email":        *p.Email,
		"AddressLine1": *p.AddressLine1,
		"AddressLine2": *p.AddressLine2,
		"City":         *p.City,
		"District":     *p.District,
		"State":        *p.State,
		"PostalCode":   *p.PostalCode,
	}

	// Encrypt first.
	if err := repo.encryptPatientPHI(p); err != nil {
		t.Fatalf("encrypt error: %v", err)
	}

	// Then decrypt.
	if err := repo.decryptPatientPHI(p); err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	// Verify each field matches its original value.
	fields := map[string]*string{
		"SSNHash":      p.SSNHash,
		"AadhaarHash":  p.AadhaarHash,
		"PhoneHome":    p.PhoneHome,
		"PhoneMobile":  p.PhoneMobile,
		"PhoneWork":    p.PhoneWork,
		"Email":        p.Email,
		"AddressLine1": p.AddressLine1,
		"AddressLine2": p.AddressLine2,
		"City":         p.City,
		"District":     p.District,
		"State":        p.State,
		"PostalCode":   p.PostalCode,
	}

	for name, field := range fields {
		if field == nil {
			t.Errorf("%s: expected non-nil after decryption", name)
			continue
		}
		if *field != originals[name] {
			t.Errorf("%s: expected %q, got %q", name, originals[name], *field)
		}
	}
}

func TestEncryptDecryptPatientPHI_RoundTrip(t *testing.T) {
	enc := newTestEncryptor(t)
	repo := newRepoWithEncryptor(enc)
	p := buildPatientWithAllPHI()

	// Save all original values.
	origSSN := *p.SSNHash
	origAadhaar := *p.AadhaarHash
	origPhoneHome := *p.PhoneHome
	origPhoneMobile := *p.PhoneMobile
	origPhoneWork := *p.PhoneWork
	origEmail := *p.Email
	origAddr1 := *p.AddressLine1
	origAddr2 := *p.AddressLine2
	origCity := *p.City
	origDistrict := *p.District
	origState := *p.State
	origPostal := *p.PostalCode

	// Non-PHI fields that should remain unchanged.
	origFirstName := p.FirstName
	origLastName := p.LastName
	origMRN := p.MRN

	// Encrypt then decrypt.
	if err := repo.encryptPatientPHI(p); err != nil {
		t.Fatalf("encrypt error: %v", err)
	}
	if err := repo.decryptPatientPHI(p); err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	// Verify all 12 PHI fields survived the round-trip.
	type fieldCheck struct {
		name     string
		got      *string
		expected string
	}
	checks := []fieldCheck{
		{"SSNHash", p.SSNHash, origSSN},
		{"AadhaarHash", p.AadhaarHash, origAadhaar},
		{"PhoneHome", p.PhoneHome, origPhoneHome},
		{"PhoneMobile", p.PhoneMobile, origPhoneMobile},
		{"PhoneWork", p.PhoneWork, origPhoneWork},
		{"Email", p.Email, origEmail},
		{"AddressLine1", p.AddressLine1, origAddr1},
		{"AddressLine2", p.AddressLine2, origAddr2},
		{"City", p.City, origCity},
		{"District", p.District, origDistrict},
		{"State", p.State, origState},
		{"PostalCode", p.PostalCode, origPostal},
	}

	for _, c := range checks {
		if c.got == nil {
			t.Errorf("%s: expected non-nil after round-trip", c.name)
			continue
		}
		if *c.got != c.expected {
			t.Errorf("%s: expected %q, got %q", c.name, c.expected, *c.got)
		}
	}

	// Verify non-PHI fields are untouched.
	if p.FirstName != origFirstName {
		t.Errorf("FirstName: expected %q, got %q", origFirstName, p.FirstName)
	}
	if p.LastName != origLastName {
		t.Errorf("LastName: expected %q, got %q", origLastName, p.LastName)
	}
	if p.MRN != origMRN {
		t.Errorf("MRN: expected %q, got %q", origMRN, p.MRN)
	}
}

func TestEncryptPatientPHI_NilEncryptor(t *testing.T) {
	repo := newRepoWithEncryptor(nil)
	p := buildPatientWithAllPHI()

	// Save originals.
	origSSN := *p.SSNHash
	origAadhaar := *p.AadhaarHash
	origPhoneHome := *p.PhoneHome
	origPhoneMobile := *p.PhoneMobile
	origPhoneWork := *p.PhoneWork
	origEmail := *p.Email
	origAddr1 := *p.AddressLine1
	origAddr2 := *p.AddressLine2
	origCity := *p.City
	origDistrict := *p.District
	origState := *p.State
	origPostal := *p.PostalCode

	err := repo.encryptPatientPHI(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All fields should remain unchanged with nil encryptor.
	checks := []struct {
		name     string
		field    *string
		expected string
	}{
		{"SSNHash", p.SSNHash, origSSN},
		{"AadhaarHash", p.AadhaarHash, origAadhaar},
		{"PhoneHome", p.PhoneHome, origPhoneHome},
		{"PhoneMobile", p.PhoneMobile, origPhoneMobile},
		{"PhoneWork", p.PhoneWork, origPhoneWork},
		{"Email", p.Email, origEmail},
		{"AddressLine1", p.AddressLine1, origAddr1},
		{"AddressLine2", p.AddressLine2, origAddr2},
		{"City", p.City, origCity},
		{"District", p.District, origDistrict},
		{"State", p.State, origState},
		{"PostalCode", p.PostalCode, origPostal},
	}

	for _, c := range checks {
		if c.field == nil {
			t.Errorf("%s: expected non-nil", c.name)
			continue
		}
		if *c.field != c.expected {
			t.Errorf("%s: expected %q (unchanged), got %q", c.name, c.expected, *c.field)
		}
	}
}

func TestEncryptPatientPHI_NilFields(t *testing.T) {
	enc := newTestEncryptor(t)
	repo := newRepoWithEncryptor(enc)

	// Patient with all PHI pointer fields left as nil.
	p := &Patient{
		FirstName: "Jane",
		LastName:  "Smith",
		MRN:       "MRN-NIL",
	}

	err := repo.encryptPatientPHI(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All nil pointer fields should remain nil.
	nilChecks := []struct {
		name  string
		field *string
	}{
		{"SSNHash", p.SSNHash},
		{"AadhaarHash", p.AadhaarHash},
		{"PhoneHome", p.PhoneHome},
		{"PhoneMobile", p.PhoneMobile},
		{"PhoneWork", p.PhoneWork},
		{"Email", p.Email},
		{"AddressLine1", p.AddressLine1},
		{"AddressLine2", p.AddressLine2},
		{"City", p.City},
		{"District", p.District},
		{"State", p.State},
		{"PostalCode", p.PostalCode},
	}

	for _, c := range nilChecks {
		if c.field != nil {
			t.Errorf("%s: expected nil, got %q", c.name, *c.field)
		}
	}

	// Non-PHI fields should be untouched.
	if p.FirstName != "Jane" {
		t.Errorf("FirstName: expected Jane, got %s", p.FirstName)
	}
	if p.LastName != "Smith" {
		t.Errorf("LastName: expected Smith, got %s", p.LastName)
	}
}
