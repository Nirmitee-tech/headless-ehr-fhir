package identity

import (
	"testing"

	"github.com/ehr/ehr/internal/platform/hipaa"
)

// mockFieldEncryptor is a test double for hipaa.FieldEncryptor that applies a
// reversible prefix transformation so we can verify encrypt/decrypt without
// needing real AES keys.
type mockFieldEncryptor struct{}

func (m *mockFieldEncryptor) Encrypt(plaintext string) (string, error) {
	return "ENC:" + plaintext, nil
}

func (m *mockFieldEncryptor) Decrypt(ciphertext string) (string, error) {
	if len(ciphertext) > 4 && ciphertext[:4] == "ENC:" {
		return ciphertext[4:], nil
	}
	return ciphertext, nil
}

// ---------- encryptField / decryptField ----------

func TestPractitionerEncryptField_NilEncryptor(t *testing.T) {
	r := &practRepoPG{} // nil encryptor
	val := "test-value"
	got, err := r.encryptPractitionerField(&val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || *got != "test-value" {
		t.Fatalf("expected unchanged value, got %v", got)
	}
}

func TestPractitionerEncryptField_NilValue(t *testing.T) {
	r := &practRepoPG{encryptor: &mockFieldEncryptor{}}
	got, err := r.encryptPractitionerField(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestPractitionerEncryptField_EncryptsValue(t *testing.T) {
	r := &practRepoPG{encryptor: &mockFieldEncryptor{}}
	val := "secret"
	got, err := r.encryptPractitionerField(&val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if *got == "secret" {
		t.Fatal("encrypted value should differ from original")
	}
	if *got != "ENC:secret" {
		t.Fatalf("expected ENC:secret, got %s", *got)
	}
}

func TestPractitionerDecryptField_RoundTrip(t *testing.T) {
	r := &practRepoPG{encryptor: &mockFieldEncryptor{}}
	original := "my-data"
	encrypted, err := r.encryptPractitionerField(&original)
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}
	decrypted, err := r.decryptPractitionerField(encrypted)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}
	if decrypted == nil || *decrypted != "my-data" {
		t.Fatalf("round trip failed: got %v", decrypted)
	}
}

// ---------- encryptPractitionerPII / decryptPractitionerPII ----------

func TestEncryptPractitionerPII_EncryptsAllFields(t *testing.T) {
	r := &practRepoPG{encryptor: &mockFieldEncryptor{}}
	p := &Practitioner{
		Phone:             ptrStr("555-1234"),
		Email:             ptrStr("doc@example.com"),
		AddressLine1:      ptrStr("123 Main St"),
		City:              ptrStr("Springfield"),
		State:             ptrStr("IL"),
		PostalCode:        ptrStr("62701"),
		Country:           ptrStr("US"),
		NPINumber:         ptrStr("1234567890"),
		DEANumber:         ptrStr("AB1234567"),
		StateLicenseNum:   ptrStr("MD-12345"),
		MedicalCouncilReg: ptrStr("MCR-999"),
		AbhaID:            ptrStr("ABHA-001"),
	}

	if err := r.encryptPractitionerPII(p); err != nil {
		t.Fatalf("encryptPractitionerPII error: %v", err)
	}

	fields := map[string]*string{
		"Phone":             p.Phone,
		"Email":             p.Email,
		"AddressLine1":      p.AddressLine1,
		"City":              p.City,
		"State":             p.State,
		"PostalCode":        p.PostalCode,
		"Country":           p.Country,
		"NPINumber":         p.NPINumber,
		"DEANumber":         p.DEANumber,
		"StateLicenseNum":   p.StateLicenseNum,
		"MedicalCouncilReg": p.MedicalCouncilReg,
		"AbhaID":            p.AbhaID,
	}
	for name, val := range fields {
		if val == nil {
			t.Errorf("%s is nil after encryption", name)
			continue
		}
		if len(*val) < 5 || (*val)[:4] != "ENC:" {
			t.Errorf("%s was not encrypted: got %q", name, *val)
		}
	}
}

func TestDecryptPractitionerPII_DecryptsAllFields(t *testing.T) {
	r := &practRepoPG{encryptor: &mockFieldEncryptor{}}
	p := &Practitioner{
		Phone:             ptrStr("ENC:555-1234"),
		Email:             ptrStr("ENC:doc@example.com"),
		AddressLine1:      ptrStr("ENC:123 Main St"),
		City:              ptrStr("ENC:Springfield"),
		State:             ptrStr("ENC:IL"),
		PostalCode:        ptrStr("ENC:62701"),
		Country:           ptrStr("ENC:US"),
		NPINumber:         ptrStr("ENC:1234567890"),
		DEANumber:         ptrStr("ENC:AB1234567"),
		StateLicenseNum:   ptrStr("ENC:MD-12345"),
		MedicalCouncilReg: ptrStr("ENC:MCR-999"),
		AbhaID:            ptrStr("ENC:ABHA-001"),
	}

	if err := r.decryptPractitionerPII(p); err != nil {
		t.Fatalf("decryptPractitionerPII error: %v", err)
	}

	expected := map[string]struct {
		got  *string
		want string
	}{
		"Phone":             {p.Phone, "555-1234"},
		"Email":             {p.Email, "doc@example.com"},
		"AddressLine1":      {p.AddressLine1, "123 Main St"},
		"City":              {p.City, "Springfield"},
		"State":             {p.State, "IL"},
		"PostalCode":        {p.PostalCode, "62701"},
		"Country":           {p.Country, "US"},
		"NPINumber":         {p.NPINumber, "1234567890"},
		"DEANumber":         {p.DEANumber, "AB1234567"},
		"StateLicenseNum":   {p.StateLicenseNum, "MD-12345"},
		"MedicalCouncilReg": {p.MedicalCouncilReg, "MCR-999"},
		"AbhaID":            {p.AbhaID, "ABHA-001"},
	}
	for name, tc := range expected {
		if tc.got == nil {
			t.Errorf("%s is nil after decryption", name)
			continue
		}
		if *tc.got != tc.want {
			t.Errorf("%s: got %q, want %q", name, *tc.got, tc.want)
		}
	}
}

func TestEncryptDecryptPractitionerPII_RoundTrip(t *testing.T) {
	// Use real AES encryption for a true round-trip test.
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	enc, err := hipaa.NewPHIEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}
	r := &practRepoPG{encryptor: enc}

	p := &Practitioner{
		Phone:             ptrStr("555-1234"),
		Email:             ptrStr("doc@example.com"),
		AddressLine1:      ptrStr("123 Main St"),
		City:              ptrStr("Springfield"),
		State:             ptrStr("IL"),
		PostalCode:        ptrStr("62701"),
		Country:           ptrStr("US"),
		NPINumber:         ptrStr("1234567890"),
		DEANumber:         ptrStr("AB1234567"),
		StateLicenseNum:   ptrStr("MD-12345"),
		MedicalCouncilReg: ptrStr("MCR-999"),
		AbhaID:            ptrStr("ABHA-001"),
	}

	if err := r.encryptPractitionerPII(p); err != nil {
		t.Fatalf("encrypt error: %v", err)
	}
	// Verify fields are actually different after encryption.
	if p.Phone != nil && *p.Phone == "555-1234" {
		t.Fatal("Phone was not encrypted")
	}

	if err := r.decryptPractitionerPII(p); err != nil {
		t.Fatalf("decrypt error: %v", err)
	}

	checks := map[string]struct {
		got  *string
		want string
	}{
		"Phone":             {p.Phone, "555-1234"},
		"Email":             {p.Email, "doc@example.com"},
		"AddressLine1":      {p.AddressLine1, "123 Main St"},
		"City":              {p.City, "Springfield"},
		"State":             {p.State, "IL"},
		"PostalCode":        {p.PostalCode, "62701"},
		"Country":           {p.Country, "US"},
		"NPINumber":         {p.NPINumber, "1234567890"},
		"DEANumber":         {p.DEANumber, "AB1234567"},
		"StateLicenseNum":   {p.StateLicenseNum, "MD-12345"},
		"MedicalCouncilReg": {p.MedicalCouncilReg, "MCR-999"},
		"AbhaID":            {p.AbhaID, "ABHA-001"},
	}
	for name, tc := range checks {
		if tc.got == nil {
			t.Errorf("%s is nil after round trip", name)
			continue
		}
		if *tc.got != tc.want {
			t.Errorf("%s: got %q, want %q", name, *tc.got, tc.want)
		}
	}
}

func TestEncryptPractitionerPII_NilEncryptor(t *testing.T) {
	r := &practRepoPG{} // nil encryptor
	p := &Practitioner{
		Phone:    ptrStr("555-1234"),
		Email:    ptrStr("doc@example.com"),
		NPINumber: ptrStr("1234567890"),
	}

	if err := r.encryptPractitionerPII(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Values should be unchanged.
	if p.Phone == nil || *p.Phone != "555-1234" {
		t.Errorf("Phone changed with nil encryptor: %v", p.Phone)
	}
	if p.Email == nil || *p.Email != "doc@example.com" {
		t.Errorf("Email changed with nil encryptor: %v", p.Email)
	}
	if p.NPINumber == nil || *p.NPINumber != "1234567890" {
		t.Errorf("NPINumber changed with nil encryptor: %v", p.NPINumber)
	}
}

func TestEncryptPractitionerPII_NilFields(t *testing.T) {
	r := &practRepoPG{encryptor: &mockFieldEncryptor{}}
	p := &Practitioner{
		// All PII pointer fields are nil by default.
		FirstName: "John",
		LastName:  "Doe",
	}

	if err := r.encryptPractitionerPII(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Nil fields should remain nil.
	if p.Phone != nil {
		t.Errorf("Phone should be nil, got %v", p.Phone)
	}
	if p.Email != nil {
		t.Errorf("Email should be nil, got %v", p.Email)
	}
	if p.AddressLine1 != nil {
		t.Errorf("AddressLine1 should be nil, got %v", p.AddressLine1)
	}
	if p.City != nil {
		t.Errorf("City should be nil, got %v", p.City)
	}
	if p.State != nil {
		t.Errorf("State should be nil, got %v", p.State)
	}
	if p.PostalCode != nil {
		t.Errorf("PostalCode should be nil, got %v", p.PostalCode)
	}
	if p.Country != nil {
		t.Errorf("Country should be nil, got %v", p.Country)
	}
	if p.NPINumber != nil {
		t.Errorf("NPINumber should be nil, got %v", p.NPINumber)
	}
	if p.DEANumber != nil {
		t.Errorf("DEANumber should be nil, got %v", p.DEANumber)
	}
	if p.StateLicenseNum != nil {
		t.Errorf("StateLicenseNum should be nil, got %v", p.StateLicenseNum)
	}
	if p.MedicalCouncilReg != nil {
		t.Errorf("MedicalCouncilReg should be nil, got %v", p.MedicalCouncilReg)
	}
	if p.AbhaID != nil {
		t.Errorf("AbhaID should be nil, got %v", p.AbhaID)
	}
}
