package fhir

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

const usCorePatient = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient"

func resourceByType(t *testing.T, cs map[string]interface{}, rt string) map[string]interface{} {
	t.Helper()
	for _, r := range cs["rest"].([]map[string]interface{})[0]["resource"].([]map[string]interface{}) {
		if r["type"] == rt {
			return r
		}
	}
	t.Fatalf("resource %s not in CapabilityStatement", rt)
	return nil
}

func TestCapabilityBuilder_USCoreOffByDefault(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.AddResource("Patient", DefaultInteractions(), nil)
	cs := b.Build()

	for _, u := range cs["instantiates"].([]string) {
		if u == USCoreServerCapabilityURL {
			t.Fatal("US Core declared without DeclareUSCoreServer")
		}
	}
	if _, ok := resourceByType(t, cs, "Patient")["supportedProfile"]; ok {
		t.Fatal("supportedProfile set without DeclareUSCoreServer")
	}
}

func TestCapabilityBuilder_DeclareUSCoreServer(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.AddResource("Patient", DefaultInteractions(), nil)
	b.AddResource("Observation", DefaultInteractions(), nil)
	b.AddResourceWithProfile("Condition", DefaultInteractions(), nil, []string{"http://example.org/StructureDefinition/local-condition"})
	b.AddResource("Basic", DefaultInteractions(), nil) // not a US Core type
	b.DeclareUSCoreServer()
	cs := b.Build()

	t.Run("instantiates keeps bulk data and adds US Core server", func(t *testing.T) {
		got := cs["instantiates"].([]string)
		if len(got) != 2 || got[0] != "http://hl7.org/fhir/uv/bulkdata/CapabilityStatement/bulk-data" || got[1] != USCoreServerCapabilityURL {
			t.Fatalf("instantiates = %v", got)
		}
	})

	t.Run("Patient lists the US Core Patient profile", func(t *testing.T) {
		got := resourceByType(t, cs, "Patient")["supportedProfile"].([]string)
		if len(got) != 1 || got[0] != usCorePatient {
			t.Fatalf("Patient supportedProfile = %v", got)
		}
	})

	t.Run("Observation lists all 22 US Core 6.1.0 profiles", func(t *testing.T) {
		if got := resourceByType(t, cs, "Observation")["supportedProfile"].([]string); len(got) != 22 {
			t.Fatalf("Observation supportedProfile has %d entries, want 22", len(got))
		}
	})

	t.Run("explicit profiles stay first and are not duplicated", func(t *testing.T) {
		got := resourceByType(t, cs, "Condition")["supportedProfile"].([]string)
		if len(got) != 3 || got[0] != "http://example.org/StructureDefinition/local-condition" {
			t.Fatalf("Condition supportedProfile = %v", got)
		}
	})

	t.Run("non US Core type gets no profile", func(t *testing.T) {
		if _, ok := resourceByType(t, cs, "Basic")["supportedProfile"]; ok {
			t.Fatal("Basic should have no supportedProfile")
		}
	})

	t.Run("unregistered US Core types are not added", func(t *testing.T) {
		for _, r := range cs["rest"].([]map[string]interface{})[0]["resource"].([]map[string]interface{}) {
			if r["type"] == "Goal" {
				t.Fatal("Goal was not registered but appears")
			}
		}
	})
}

func TestMergeProfiles(t *testing.T) {
	got := mergeProfiles([]string{"a", "b"}, []string{"b", "c"})
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("mergeProfiles = %v", got)
	}
	if got := mergeProfiles(nil, nil); got != nil {
		t.Fatalf("mergeProfiles(nil, nil) = %v", got)
	}
}

// The HTTP check reads /fhir/metadata the way the (g)(10) kit does: as JSON over the wire.
func TestCapabilityHandler_MetadataDeclaresUSCore(t *testing.T) {
	b := NewCapabilityBuilder("http://localhost:8000/fhir", "0.1.0")
	b.AddResource("Patient", DefaultInteractions(), nil)
	b.AddResource("Encounter", DefaultInteractions(), nil)
	b.DeclareUSCoreServer()

	e := echo.New()
	NewCapabilityHandler(b).RegisterRoutes(e.Group("/fhir"))
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fhir/metadata", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var cs struct {
		Instantiates []string `json:"instantiates"`
		Rest         []struct {
			Resource []struct {
				Type             string   `json:"type"`
				SupportedProfile []string `json:"supportedProfile"`
			} `json:"resource"`
		} `json:"rest"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &cs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	found := false
	for _, u := range cs.Instantiates {
		found = found || u == USCoreServerCapabilityURL
	}
	if !found {
		t.Fatalf("instantiates over HTTP = %v", cs.Instantiates)
	}
	for _, r := range cs.Rest[0].Resource {
		want := map[string]string{"Patient": usCorePatient, "Encounter": "http://hl7.org/fhir/us/core/StructureDefinition/us-core-encounter"}[r.Type]
		if len(r.SupportedProfile) != 1 || r.SupportedProfile[0] != want {
			t.Errorf("%s supportedProfile over HTTP = %v", r.Type, r.SupportedProfile)
		}
	}
}
