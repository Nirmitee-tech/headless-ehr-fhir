package ccda

import "encoding/xml"

// CDA OIDs and template identifiers for C-CDA 2.1 CCD documents.
const (
	// CDA namespace
	CDANamespace = "urn:hl7-org:v3"
	XSINamespace = "http://www.w3.org/2001/XMLSchema-instance"
	SDTCNamespace = "urn:hl7-org:sdtc"

	// Document-level template IDs
	OIDUSRealmHeader = "2.16.840.1.113883.10.20.22.1.1"
	OIDCCDDocument   = "2.16.840.1.113883.10.20.22.1.2"

	// Section-level template IDs
	OIDAllergiesSection     = "2.16.840.1.113883.10.20.22.2.6.1"
	OIDMedicationsSection   = "2.16.840.1.113883.10.20.22.2.1.1"
	OIDProblemsSection      = "2.16.840.1.113883.10.20.22.2.5.1"
	OIDProceduresSection    = "2.16.840.1.113883.10.20.22.2.7.1"
	OIDResultsSection       = "2.16.840.1.113883.10.20.22.2.3.1"
	OIDVitalSignsSection    = "2.16.840.1.113883.10.20.22.2.4.1"
	OIDImmunizationsSection = "2.16.840.1.113883.10.20.22.2.2.1"
	OIDSocialHistorySection = "2.16.840.1.113883.10.20.22.2.17"
	OIDPlanOfCareSection    = "2.16.840.1.113883.10.20.22.2.10"
	OIDEncountersSection    = "2.16.840.1.113883.10.20.22.2.22.1"

	// Entry-level template IDs
	OIDAllergyEntry      = "2.16.840.1.113883.10.20.22.4.30"
	OIDMedicationEntry   = "2.16.840.1.113883.10.20.22.4.16"
	OIDProblemEntry      = "2.16.840.1.113883.10.20.22.4.3"
	OIDProcedureEntry    = "2.16.840.1.113883.10.20.22.4.14"
	OIDResultEntry       = "2.16.840.1.113883.10.20.22.4.1"
	OIDVitalSignEntry    = "2.16.840.1.113883.10.20.22.4.26"
	OIDImmunizationEntry = "2.16.840.1.113883.10.20.22.4.52"
	OIDEncounterEntry    = "2.16.840.1.113883.10.20.22.4.49"

	// LOINC codes for section identification
	LOINCAllergies     = "48765-2"
	LOINCMedications   = "10160-0"
	LOINCProblems      = "11450-4"
	LOINCProcedures    = "47519-4"
	LOINCResults       = "30954-2"
	LOINCVitalSigns    = "8716-3"
	LOINCImmunizations = "11369-6"
	LOINCSocialHistory = "29762-2"
	LOINCPlanOfCare    = "18776-5"
	LOINCEncounters    = "46240-8"

	// Code system OIDs
	OIDLOINC   = "2.16.840.1.113883.6.1"
	OIDSNOMED  = "2.16.840.1.113883.6.96"
	OIDRxNorm  = "2.16.840.1.113883.6.88"
	OIDICD10   = "2.16.840.1.113883.6.90"
	OIDCVX     = "2.16.840.1.113883.12.292"
	OIDAdminGender = "2.16.840.1.113883.5.1"
)

// ClinicalDocument is the root element of a CDA R2 document.
type ClinicalDocument struct {
	XMLName           xml.Name          `xml:"urn:hl7-org:v3 ClinicalDocument"`
	XSI               string            `xml:"xmlns:xsi,attr"`
	SDTC              string            `xml:"xmlns:sdtc,attr,omitempty"`
	RealmCode         *Code             `xml:"realmCode,omitempty"`
	TypeID            *TypeID           `xml:"typeId,omitempty"`
	TemplateIDs       []TemplateID      `xml:"templateId,omitempty"`
	ID                *InstanceID       `xml:"id,omitempty"`
	Code              *Code             `xml:"code,omitempty"`
	Title             string            `xml:"title,omitempty"`
	EffectiveTime     *TimeValue        `xml:"effectiveTime,omitempty"`
	ConfidentialityCode *Code           `xml:"confidentialityCode,omitempty"`
	LanguageCode      *Code             `xml:"languageCode,omitempty"`
	RecordTarget      *RecordTarget     `xml:"recordTarget,omitempty"`
	Author            *Author           `xml:"author,omitempty"`
	Custodian         *Custodian        `xml:"custodian,omitempty"`
	DocumentationOf   *DocumentationOf  `xml:"documentationOf,omitempty"`
	Component         *Component        `xml:"component,omitempty"`
}

// TypeID identifies the CDA R2 schema.
type TypeID struct {
	Root      string `xml:"root,attr"`
	Extension string `xml:"extension,attr"`
}

// TemplateID specifies a template identifier with optional extension.
type TemplateID struct {
	Root      string `xml:"root,attr"`
	Extension string `xml:"extension,attr,omitempty"`
}

// InstanceID is a unique instance identifier.
type InstanceID struct {
	Root      string `xml:"root,attr"`
	Extension string `xml:"extension,attr,omitempty"`
}

// Code represents a coded value with optional code system.
type Code struct {
	Code           string `xml:"code,attr,omitempty"`
	CodeSystem     string `xml:"codeSystem,attr,omitempty"`
	CodeSystemName string `xml:"codeSystemName,attr,omitempty"`
	DisplayName    string `xml:"displayName,attr,omitempty"`
	NullFlavor     string `xml:"nullFlavor,attr,omitempty"`
}

// TimeValue holds a time stamp in HL7 format (YYYYMMDD or YYYYMMDDHHmmss).
type TimeValue struct {
	Value string `xml:"value,attr,omitempty"`
}

// TimeLow represents a low boundary of a time interval.
type TimeLow struct {
	Value string `xml:"value,attr,omitempty"`
}

// TimeHigh represents a high boundary of a time interval.
type TimeHigh struct {
	Value string `xml:"value,attr,omitempty"`
}

// TimeRange represents an effectiveTime interval with low and high.
type TimeRange struct {
	Low  *TimeLow  `xml:"low,omitempty"`
	High *TimeHigh `xml:"high,omitempty"`
}

// RecordTarget holds the patient information in the CDA header.
type RecordTarget struct {
	PatientRole *PatientRole `xml:"patientRole,omitempty"`
}

// PatientRole contains patient identifiers and demographics.
type PatientRole struct {
	IDs     []InstanceID `xml:"id,omitempty"`
	Addr    *Address     `xml:"addr,omitempty"`
	Telecom *Telecom     `xml:"telecom,omitempty"`
	Patient *Patient     `xml:"patient,omitempty"`
}

// Patient holds patient demographic data.
type Patient struct {
	Name            *Name      `xml:"name,omitempty"`
	AdministrativeGenderCode *Code `xml:"administrativeGenderCode,omitempty"`
	BirthTime       *TimeValue `xml:"birthTime,omitempty"`
}

// Name represents a person's name.
type Name struct {
	Given  string `xml:"given,omitempty"`
	Family string `xml:"family,omitempty"`
}

// Address represents a postal address.
type Address struct {
	Use            string `xml:"use,attr,omitempty"`
	StreetAddress  string `xml:"streetAddressLine,omitempty"`
	City           string `xml:"city,omitempty"`
	State          string `xml:"state,omitempty"`
	PostalCode     string `xml:"postalCode,omitempty"`
	Country        string `xml:"country,omitempty"`
}

// Telecom represents a contact point (phone, email, etc.).
type Telecom struct {
	Use   string `xml:"use,attr,omitempty"`
	Value string `xml:"value,attr,omitempty"`
}

// Author holds authoring information in the CDA header.
type Author struct {
	Time             *TimeValue        `xml:"time,omitempty"`
	AssignedAuthor   *AssignedAuthor   `xml:"assignedAuthor,omitempty"`
}

// AssignedAuthor identifies the author entity.
type AssignedAuthor struct {
	ID                 *InstanceID        `xml:"id,omitempty"`
	AssignedAuthoringDevice *AuthoringDevice `xml:"assignedAuthoringDevice,omitempty"`
	RepresentedOrganization *Organization `xml:"representedOrganization,omitempty"`
}

// AuthoringDevice identifies a device as the author.
type AuthoringDevice struct {
	SoftwareName string `xml:"softwareName,omitempty"`
}

// Organization represents a healthcare organization.
type Organization struct {
	IDs   []InstanceID `xml:"id,omitempty"`
	Names []string     `xml:"name,omitempty"`
}

// Custodian holds the custodian organization in the CDA header.
type Custodian struct {
	AssignedCustodian *AssignedCustodian `xml:"assignedCustodian,omitempty"`
}

// AssignedCustodian contains the custodian organization.
type AssignedCustodian struct {
	RepresentedCustodianOrganization *CustodianOrganization `xml:"representedCustodianOrganization,omitempty"`
}

// CustodianOrganization identifies the custodian.
type CustodianOrganization struct {
	IDs   []InstanceID `xml:"id,omitempty"`
	Names []string     `xml:"name,omitempty"`
}

// DocumentationOf records the service event documented.
type DocumentationOf struct {
	ServiceEvent *ServiceEvent `xml:"serviceEvent,omitempty"`
}

// ServiceEvent describes the clinical service documented.
type ServiceEvent struct {
	ClassCode     string     `xml:"classCode,attr,omitempty"`
	EffectiveTime *TimeRange `xml:"effectiveTime,omitempty"`
}

// Component wraps the structured body of the CDA document.
type Component struct {
	StructuredBody *StructuredBody `xml:"structuredBody,omitempty"`
}

// StructuredBody holds the document sections.
type StructuredBody struct {
	Components []SectionComponent `xml:"component,omitempty"`
}

// SectionComponent wraps a single section.
type SectionComponent struct {
	Section *Section `xml:"section,omitempty"`
}

// Section represents a CDA section with template, code, narrative, and entries.
type Section struct {
	TemplateIDs []TemplateID `xml:"templateId,omitempty"`
	Code        *Code        `xml:"code,omitempty"`
	Title       string       `xml:"title,omitempty"`
	Text        *Narrative   `xml:"text,omitempty"`
	Entries     []Entry      `xml:"entry,omitempty"`
}

// Narrative holds the human-readable HTML narrative block for a section.
type Narrative struct {
	Table *NarrativeTable `xml:"table,omitempty"`
	Content string        `xml:",innerxml"`
}

// NarrativeTable is a simplified HTML table for section narratives.
type NarrativeTable struct {
	Thead *NarrativeThead `xml:"thead,omitempty"`
	Tbody *NarrativeTbody `xml:"tbody,omitempty"`
}

// NarrativeThead is a table header.
type NarrativeThead struct {
	Tr *NarrativeTr `xml:"tr,omitempty"`
}

// NarrativeTbody is a table body.
type NarrativeTbody struct {
	Trs []NarrativeTr `xml:"tr,omitempty"`
}

// NarrativeTr is a table row.
type NarrativeTr struct {
	Tds []string `xml:"td,omitempty"`
	Ths []string `xml:"th,omitempty"`
}

// Entry represents a CDA entry element containing clinical data.
type Entry struct {
	TypeCode string `xml:"typeCode,attr,omitempty"`
	Act      *Act   `xml:"act,omitempty"`
	Organizer *Organizer `xml:"organizer,omitempty"`
	SubstanceAdministration *SubstanceAdministration `xml:"substanceAdministration,omitempty"`
	Procedure *ProcedureEntry `xml:"procedure,omitempty"`
	Encounter *EncounterEntry `xml:"encounter,omitempty"`
	Observation *ObservationEntry `xml:"observation,omitempty"`
}

// Act represents a CDA act element.
type Act struct {
	ClassCode   string       `xml:"classCode,attr,omitempty"`
	MoodCode    string       `xml:"moodCode,attr,omitempty"`
	TemplateIDs []TemplateID `xml:"templateId,omitempty"`
	IDs         []InstanceID `xml:"id,omitempty"`
	Code        *Code        `xml:"code,omitempty"`
	StatusCode  *Code        `xml:"statusCode,omitempty"`
	EffectiveTime *TimeRange `xml:"effectiveTime,omitempty"`
	EntryRelationships []EntryRelationship `xml:"entryRelationship,omitempty"`
}

// EntryRelationship links entries together.
type EntryRelationship struct {
	TypeCode    string       `xml:"typeCode,attr,omitempty"`
	Observation *ObservationEntry `xml:"observation,omitempty"`
}

// ObservationEntry represents a CDA observation.
type ObservationEntry struct {
	ClassCode   string       `xml:"classCode,attr,omitempty"`
	MoodCode    string       `xml:"moodCode,attr,omitempty"`
	TemplateIDs []TemplateID `xml:"templateId,omitempty"`
	IDs         []InstanceID `xml:"id,omitempty"`
	Code        *Code        `xml:"code,omitempty"`
	StatusCode  *Code        `xml:"statusCode,omitempty"`
	EffectiveTime *TimeRange `xml:"effectiveTime,omitempty"`
	Value       *Value       `xml:"value,omitempty"`
	Text        string       `xml:"text,omitempty"`
	Participant *Participant `xml:"participant,omitempty"`
	Consumable  *Consumable  `xml:"consumable,omitempty"`
}

// Value represents a typed value (physical quantity, coded value, etc.).
type Value struct {
	Type        string `xml:"xsi:type,attr,omitempty"`
	Value       string `xml:"value,attr,omitempty"`
	Unit        string `xml:"unit,attr,omitempty"`
	Code        string `xml:"code,attr,omitempty"`
	CodeSystem  string `xml:"codeSystem,attr,omitempty"`
	DisplayName string `xml:"displayName,attr,omitempty"`
}

// Participant represents a participant in an entry.
type Participant struct {
	TypeCode       string          `xml:"typeCode,attr,omitempty"`
	ParticipantRole *ParticipantRole `xml:"participantRole,omitempty"`
}

// ParticipantRole holds participant role information.
type ParticipantRole struct {
	ClassCode       string           `xml:"classCode,attr,omitempty"`
	PlayingEntity   *PlayingEntity   `xml:"playingEntity,omitempty"`
}

// PlayingEntity holds an entity name and code.
type PlayingEntity struct {
	ClassCode string `xml:"classCode,attr,omitempty"`
	Code      *Code  `xml:"code,omitempty"`
	Name      string `xml:"name,omitempty"`
}

// SubstanceAdministration represents a medication administration entry.
type SubstanceAdministration struct {
	ClassCode   string       `xml:"classCode,attr,omitempty"`
	MoodCode    string       `xml:"moodCode,attr,omitempty"`
	TemplateIDs []TemplateID `xml:"templateId,omitempty"`
	IDs         []InstanceID `xml:"id,omitempty"`
	StatusCode  *Code        `xml:"statusCode,omitempty"`
	EffectiveTime *TimeRange `xml:"effectiveTime,omitempty"`
	Consumable  *Consumable  `xml:"consumable,omitempty"`
	DoseQuantity *Value      `xml:"doseQuantity,omitempty"`
}

// Consumable wraps a manufactured product (medication).
type Consumable struct {
	ManufacturedProduct *ManufacturedProduct `xml:"manufacturedProduct,omitempty"`
}

// ManufacturedProduct holds a medication material.
type ManufacturedProduct struct {
	TemplateIDs        []TemplateID       `xml:"templateId,omitempty"`
	ManufacturedMaterial *ManufacturedMaterial `xml:"manufacturedMaterial,omitempty"`
}

// ManufacturedMaterial holds the medication code.
type ManufacturedMaterial struct {
	Code *Code `xml:"code,omitempty"`
}

// Organizer groups related observations (e.g., lab panels, vital sign sets).
type Organizer struct {
	ClassCode   string       `xml:"classCode,attr,omitempty"`
	MoodCode    string       `xml:"moodCode,attr,omitempty"`
	TemplateIDs []TemplateID `xml:"templateId,omitempty"`
	IDs         []InstanceID `xml:"id,omitempty"`
	Code        *Code        `xml:"code,omitempty"`
	StatusCode  *Code        `xml:"statusCode,omitempty"`
	EffectiveTime *TimeRange `xml:"effectiveTime,omitempty"`
	Components  []OrganizerComponent `xml:"component,omitempty"`
}

// OrganizerComponent wraps an observation inside an organizer.
type OrganizerComponent struct {
	Observation *ObservationEntry `xml:"observation,omitempty"`
}

// ProcedureEntry represents a CDA procedure.
type ProcedureEntry struct {
	ClassCode   string       `xml:"classCode,attr,omitempty"`
	MoodCode    string       `xml:"moodCode,attr,omitempty"`
	TemplateIDs []TemplateID `xml:"templateId,omitempty"`
	IDs         []InstanceID `xml:"id,omitempty"`
	Code        *Code        `xml:"code,omitempty"`
	StatusCode  *Code        `xml:"statusCode,omitempty"`
	EffectiveTime *TimeRange `xml:"effectiveTime,omitempty"`
}

// EncounterEntry represents a CDA encounter.
type EncounterEntry struct {
	ClassCode   string       `xml:"classCode,attr,omitempty"`
	MoodCode    string       `xml:"moodCode,attr,omitempty"`
	TemplateIDs []TemplateID `xml:"templateId,omitempty"`
	IDs         []InstanceID `xml:"id,omitempty"`
	Code        *Code        `xml:"code,omitempty"`
	StatusCode  *Code        `xml:"statusCode,omitempty"`
	EffectiveTime *TimeRange `xml:"effectiveTime,omitempty"`
}
