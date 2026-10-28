package fhir

// USCoreServerCapabilityURL is the US Core server CapabilityStatement this server declares in
// CapabilityStatement.instantiates. The ONC (g)(10) kit fails us_core_instantiate without it.
const USCoreServerCapabilityURL = "http://hl7.org/fhir/us/core/CapabilityStatement/us-core-server"

// usCoreSupportedProfiles lists the supportedProfile entries per resource type, copied from
// the US Core 6.1.0 server CapabilityStatement (hl7.org/fhir/us/core/STU6.1). Declaring a
// profile tells test tools to check resources against it; it is not proof they conform.
var usCoreSupportedProfiles = map[string][]string{
	"AllergyIntolerance": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-allergyintolerance",
	},
	"CarePlan": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-careplan",
	},
	"CareTeam": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-careteam",
	},
	"Condition": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-condition-encounter-diagnosis",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-condition-problems-health-concerns",
	},
	"Coverage": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-coverage",
	},
	"Device": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-implantable-device",
	},
	"DiagnosticReport": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-diagnosticreport-note",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-diagnosticreport-lab",
	},
	"DocumentReference": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-documentreference",
	},
	"Encounter": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-encounter",
	},
	"Goal": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-goal",
	},
	"Immunization": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-immunization",
	},
	"Location": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-location",
	},
	"Medication": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-medication",
	},
	"MedicationDispense": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-medicationdispense",
	},
	"MedicationRequest": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-medicationrequest",
	},
	"Observation": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-observation-lab",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-observation-pregnancystatus",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-observation-pregnancyintent",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-observation-occupation",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-respiratory-rate",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-simple-observation",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-heart-rate",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-body-temperature",
		"http://hl7.org/fhir/us/core/StructureDefinition/pediatric-weight-for-height",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-pulse-oximetry",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-smokingstatus",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-observation-sexual-orientation",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-head-circumference",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-body-height",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-bmi",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-observation-screening-assessment",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-blood-pressure",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-observation-clinical-result",
		"http://hl7.org/fhir/us/core/StructureDefinition/pediatric-bmi-for-age",
		"http://hl7.org/fhir/us/core/StructureDefinition/head-occipital-frontal-circumference-percentile",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-body-weight",
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-vital-signs",
	},
	"Organization": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-organization",
	},
	"Patient": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient",
	},
	"Practitioner": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-practitioner",
	},
	"PractitionerRole": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-practitionerrole",
	},
	"Procedure": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-procedure",
	},
	"Provenance": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-provenance",
	},
	"QuestionnaireResponse": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-questionnaireresponse",
	},
	"RelatedPerson": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-relatedperson",
	},
	"ServiceRequest": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-servicerequest",
	},
	"Specimen": {
		"http://hl7.org/fhir/us/core/StructureDefinition/us-core-specimen",
	},
}
