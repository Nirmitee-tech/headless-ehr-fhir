package main

import (
	"encoding/hex"
	"testing"
)

// ---------------------------------------------------------------------------
// classifyObservation tests (Bug 1: social-history routing)
// ---------------------------------------------------------------------------

func makeObservation(categoryCode string) map[string]interface{} {
	if categoryCode == "" {
		return map[string]interface{}{
			"resourceType": "Observation",
			"id":           "obs-no-cat",
		}
	}
	return map[string]interface{}{
		"resourceType": "Observation",
		"id":           "obs-" + categoryCode,
		"category": []map[string]interface{}{
			{
				"coding": []map[string]interface{}{
					{
						"system": "http://terminology.hl7.org/CodeSystem/observation-category",
						"code":   categoryCode,
					},
				},
			},
		},
	}
}

func TestClassifyObservation_SocialHistory(t *testing.T) {
	obs := makeObservation("social-history")
	if got := classifyObservation(obs); got != "social-history" {
		t.Errorf("classifyObservation(social-history) = %q, want %q", got, "social-history")
	}
}

func TestClassifyObservation_VitalSigns(t *testing.T) {
	obs := makeObservation("vital-signs")
	if got := classifyObservation(obs); got != "vital-signs" {
		t.Errorf("classifyObservation(vital-signs) = %q, want %q", got, "vital-signs")
	}
}

func TestClassifyObservation_Laboratory(t *testing.T) {
	obs := makeObservation("laboratory")
	if got := classifyObservation(obs); got != "laboratory" {
		t.Errorf("classifyObservation(laboratory) = %q, want %q", got, "laboratory")
	}
}

func TestClassifyObservation_NoCategory(t *testing.T) {
	obs := makeObservation("")
	if got := classifyObservation(obs); got != "" {
		t.Errorf("classifyObservation(no category) = %q, want empty string", got)
	}
}

func TestClassifyObservation_NilMap(t *testing.T) {
	if got := classifyObservation(nil); got != "" {
		t.Errorf("classifyObservation(nil) = %q, want empty string", got)
	}
}

func TestClassifyObservation_EmptyCategory(t *testing.T) {
	obs := map[string]interface{}{
		"category": []map[string]interface{}{},
	}
	if got := classifyObservation(obs); got != "" {
		t.Errorf("classifyObservation(empty category) = %q, want empty string", got)
	}
}

func TestClassifyObservation_EmptyCoding(t *testing.T) {
	obs := map[string]interface{}{
		"category": []map[string]interface{}{
			{
				"coding": []map[string]interface{}{},
			},
		},
	}
	if got := classifyObservation(obs); got != "" {
		t.Errorf("classifyObservation(empty coding) = %q, want empty string", got)
	}
}

// Verify that observations are routed to the correct bucket using the same
// switch logic used in FetchPatientData.
func TestObservationRouting(t *testing.T) {
	observations := []map[string]interface{}{
		makeObservation("social-history"),
		makeObservation("vital-signs"),
		makeObservation("laboratory"),
		makeObservation(""),
	}

	var results, vitalSigns, socialHistory []map[string]interface{}

	for _, obs := range observations {
		switch classifyObservation(obs) {
		case "social-history":
			socialHistory = append(socialHistory, obs)
		case "vital-signs":
			vitalSigns = append(vitalSigns, obs)
		default:
			results = append(results, obs)
		}
	}

	if len(socialHistory) != 1 {
		t.Errorf("expected 1 social-history observation, got %d", len(socialHistory))
	}
	if len(vitalSigns) != 1 {
		t.Errorf("expected 1 vital-signs observation, got %d", len(vitalSigns))
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results (laboratory + no-category), got %d", len(results))
	}
}

// ---------------------------------------------------------------------------
// resolveSmartSigningKey tests (Bug 2: insecure default key)
// ---------------------------------------------------------------------------

func TestResolveSmartSigningKey_FromEnv(t *testing.T) {
	// Produce a known 32-byte hex string.
	want := make([]byte, 32)
	for i := range want {
		want[i] = byte(i)
	}
	hexStr := hex.EncodeToString(want)

	key, random, err := resolveSmartSigningKey(hexStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if random {
		t.Error("expected random=false when env var is set")
	}
	if len(key) != 32 {
		t.Errorf("expected 32-byte key, got %d bytes", len(key))
	}
	if hex.EncodeToString(key) != hexStr {
		t.Errorf("key mismatch: got %x, want %x", key, want)
	}
}

func TestResolveSmartSigningKey_InvalidHex(t *testing.T) {
	_, _, err := resolveSmartSigningKey("not-valid-hex!!!")
	if err == nil {
		t.Fatal("expected error for invalid hex, got nil")
	}
}

func TestResolveSmartSigningKey_RandomGeneration(t *testing.T) {
	key, random, err := resolveSmartSigningKey("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !random {
		t.Error("expected random=true when env var is empty")
	}
	if len(key) != 32 {
		t.Errorf("expected 32-byte key, got %d bytes", len(key))
	}

	// Verify randomness by generating a second key and ensuring they differ.
	key2, _, err := resolveSmartSigningKey("")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if hex.EncodeToString(key) == hex.EncodeToString(key2) {
		t.Error("two random keys should not be identical")
	}
}

func TestResolveSmartSigningKey_NoInsecureDefault(t *testing.T) {
	// Ensure the old insecure default is never returned.
	key, _, err := resolveSmartSigningKey("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	insecure := "smart-signing-key-change-in-production"
	if string(key) == insecure {
		t.Error("resolveSmartSigningKey must not return the insecure default key")
	}
}
