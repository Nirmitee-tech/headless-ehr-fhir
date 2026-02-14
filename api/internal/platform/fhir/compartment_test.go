package fhir

import (
	"testing"
)

func TestGetCompartmentParam(t *testing.T) {
	tests := []struct {
		name         string
		compartment  *CompartmentDefinition
		resourceType string
		want         string
	}{
		{
			name:         "resource with linking param",
			compartment:  &PatientCompartment,
			resourceType: "Observation",
			want:         "patient",
		},
		{
			name:         "resource with empty params (Medication)",
			compartment:  &PatientCompartment,
			resourceType: "Medication",
			want:         "",
		},
		{
			name:         "resource with empty params (Slot)",
			compartment:  &PatientCompartment,
			resourceType: "Slot",
			want:         "",
		},
		{
			name:         "resource not in compartment",
			compartment:  &PatientCompartment,
			resourceType: "Device",
			want:         "",
		},
		{
			name:         "another resource with linking param (Encounter)",
			compartment:  &PatientCompartment,
			resourceType: "Encounter",
			want:         "patient",
		},
		{
			name: "custom compartment with multi-param resource returns first",
			compartment: &CompartmentDefinition{
				Type: "Custom",
				Resources: map[string][]string{
					"Foo": {"alpha", "beta"},
				},
			},
			resourceType: "Foo",
			want:         "alpha",
		},
		{
			name: "custom compartment with nil param slice",
			compartment: &CompartmentDefinition{
				Type: "Custom",
				Resources: map[string][]string{
					"Bar": nil,
				},
			},
			resourceType: "Bar",
			want:         "",
		},
		{
			name: "empty resources map",
			compartment: &CompartmentDefinition{
				Type:      "Empty",
				Resources: map[string][]string{},
			},
			resourceType: "Anything",
			want:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCompartmentParam(tt.compartment, tt.resourceType)
			if got != tt.want {
				t.Errorf("GetCompartmentParam(%q) = %q, want %q", tt.resourceType, got, tt.want)
			}
		})
	}
}

func TestIsInCompartment(t *testing.T) {
	tests := []struct {
		name         string
		compartment  *CompartmentDefinition
		resourceType string
		want         bool
	}{
		{
			name:         "resource with linking param is in compartment",
			compartment:  &PatientCompartment,
			resourceType: "Observation",
			want:         true,
		},
		{
			name:         "resource with empty params is still in compartment",
			compartment:  &PatientCompartment,
			resourceType: "Medication",
			want:         true,
		},
		{
			name:         "resource with empty params (ResearchStudy)",
			compartment:  &PatientCompartment,
			resourceType: "ResearchStudy",
			want:         true,
		},
		{
			name:         "resource not in compartment",
			compartment:  &PatientCompartment,
			resourceType: "Device",
			want:         false,
		},
		{
			name:         "another missing resource",
			compartment:  &PatientCompartment,
			resourceType: "Organization",
			want:         false,
		},
		{
			name: "empty resources map returns false",
			compartment: &CompartmentDefinition{
				Type:      "Empty",
				Resources: map[string][]string{},
			},
			resourceType: "Anything",
			want:         false,
		},
		{
			name: "custom compartment hit",
			compartment: &CompartmentDefinition{
				Type: "Custom",
				Resources: map[string][]string{
					"MyResource": {"link"},
				},
			},
			resourceType: "MyResource",
			want:         true,
		},
		{
			name: "custom compartment miss",
			compartment: &CompartmentDefinition{
				Type: "Custom",
				Resources: map[string][]string{
					"MyResource": {"link"},
				},
			},
			resourceType: "OtherResource",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInCompartment(tt.compartment, tt.resourceType)
			if got != tt.want {
				t.Errorf("IsInCompartment(%q) = %v, want %v", tt.resourceType, got, tt.want)
			}
		})
	}
}
