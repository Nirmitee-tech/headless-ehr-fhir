// Package sandbox provides synthetic FHIR data generation for sandbox/demo
// environments. It produces reproducible, clinically realistic resources
// suitable for integration testing, developer on-boarding, and UI demos.
package sandbox

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// Configuration
// ---------------------------------------------------------------------------

// SeedConfig controls the volume and shape of generated synthetic data.
type SeedConfig struct {
	PatientCount             int    `json:"patientCount"`
	EncountersPerPatient     int    `json:"encountersPerPatient"`
	ObservationsPerEncounter int    `json:"observationsPerEncounter"`
	ConditionsPerPatient     int    `json:"conditionsPerPatient"`
	MedicationsPerPatient    int    `json:"medicationsPerPatient"`
	AllergiesPerPatient      int    `json:"allergiesPerPatient"`
	ProceduresPerPatient     int    `json:"proceduresPerPatient"`
	ImmunizationsPerPatient  int    `json:"immunizationsPerPatient"`
	IncludePractitioners     bool   `json:"includePractitioners"`
	PractitionerCount        int    `json:"practitionerCount"`
	IncludeOrganization      bool   `json:"includeOrganization"`
	TenantID                 string `json:"tenantId,omitempty"`
	Seed                     int64  `json:"seed"`
}

// DefaultSeedConfig returns a SeedConfig with sensible production defaults.
func DefaultSeedConfig() SeedConfig {
	return SeedConfig{
		PatientCount:             100,
		EncountersPerPatient:     5,
		ObservationsPerEncounter: 3,
		ConditionsPerPatient:     2,
		MedicationsPerPatient:    3,
		AllergiesPerPatient:      1,
		ProceduresPerPatient:     1,
		ImmunizationsPerPatient:  4,
		IncludePractitioners:     true,
		PractitionerCount:        20,
		IncludeOrganization:      true,
	}
}

// ---------------------------------------------------------------------------
// SyntheticPatient (FHIR-like struct, kept for documentation/typing)
// ---------------------------------------------------------------------------

// HumanName represents a FHIR HumanName.
type HumanName struct {
	Use    string   `json:"use"`
	Family string   `json:"family"`
	Given  []string `json:"given"`
}

// Address represents a FHIR Address.
type Address struct {
	Use        string `json:"use"`
	Line       []string `json:"line"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

// ContactPoint represents a FHIR ContactPoint.
type ContactPoint struct {
	System string `json:"system"`
	Value  string `json:"value"`
	Use    string `json:"use"`
}

// Identifier represents a FHIR Identifier.
type Identifier struct {
	System string `json:"system"`
	Value  string `json:"value"`
	Type   struct {
		Coding []struct {
			System string `json:"system"`
			Code   string `json:"code"`
		} `json:"coding"`
	} `json:"type"`
}

// SyntheticPatient is a typed representation of a FHIR Patient resource.
type SyntheticPatient struct {
	ResourceType  string         `json:"resourceType"`
	ID            string         `json:"id"`
	Name          []HumanName    `json:"name"`
	Gender        string         `json:"gender"`
	BirthDate     string         `json:"birthDate"`
	Address       []Address      `json:"address"`
	Telecom       []ContactPoint `json:"telecom"`
	Identifier    []Identifier   `json:"identifier"`
	MaritalStatus string         `json:"maritalStatus"`
	Active        bool           `json:"active"`
}

// ---------------------------------------------------------------------------
// SeedResult
// ---------------------------------------------------------------------------

// SeedResult summarizes the output of a seed operation.
type SeedResult struct {
	Patients       int           `json:"patients"`
	Practitioners  int           `json:"practitioners"`
	Organizations  int           `json:"organizations"`
	Encounters     int           `json:"encounters"`
	Observations   int           `json:"observations"`
	Conditions     int           `json:"conditions"`
	Medications    int           `json:"medications"`
	Allergies      int           `json:"allergies"`
	Procedures     int           `json:"procedures"`
	Immunizations  int           `json:"immunizations"`
	TotalResources int           `json:"totalResources"`
	Duration       time.Duration `json:"duration"`
}

// ---------------------------------------------------------------------------
// Code pools — clinical terminology
// ---------------------------------------------------------------------------

type codeEntry struct {
	Code    string
	Display string
}

type observationDef struct {
	Code    string
	Display string
	Unit    string
	Low     float64
	High    float64
}

var (
	firstNamesMale = []string{
		"James", "Robert", "John", "Michael", "David", "William", "Richard",
		"Joseph", "Thomas", "Christopher", "Charles", "Daniel", "Matthew",
		"Anthony", "Mark", "Donald", "Steven", "Paul", "Andrew", "Joshua",
		"Kenneth", "Kevin", "Brian", "George", "Timothy", "Ronald", "Edward",
		"Jason", "Jeffrey", "Ryan", "Jacob", "Gary", "Nicholas", "Eric",
		"Jonathan", "Stephen", "Larry", "Justin", "Scott", "Brandon",
		"Benjamin", "Samuel", "Raymond", "Gregory", "Frank", "Alexander",
		"Patrick", "Jack", "Dennis", "Jerry", "Tyler",
	}
	firstNamesFemale = []string{
		"Mary", "Patricia", "Jennifer", "Linda", "Barbara", "Elizabeth",
		"Susan", "Jessica", "Sarah", "Karen", "Lisa", "Nancy", "Betty",
		"Margaret", "Sandra", "Ashley", "Dorothy", "Kimberly", "Emily",
		"Donna", "Michelle", "Carol", "Amanda", "Melissa", "Deborah",
		"Stephanie", "Rebecca", "Sharon", "Laura", "Cynthia", "Kathleen",
		"Amy", "Angela", "Shirley", "Anna", "Brenda", "Pamela", "Emma",
		"Nicole", "Helen", "Samantha", "Katherine", "Christine", "Debra",
		"Rachel", "Carolyn", "Janet", "Catherine", "Maria", "Heather",
		"Diane",
	}
	lastNames = []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia",
		"Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez",
		"Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore",
		"Jackson", "Martin", "Lee", "Perez", "Thompson", "White", "Harris",
		"Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker",
		"Young", "Allen", "King", "Wright", "Scott", "Torres", "Nguyen",
		"Hill", "Flores", "Green", "Adams", "Nelson", "Baker", "Hall",
		"Rivera", "Campbell", "Mitchell", "Carter", "Roberts", "Gomez",
	}

	streets = []string{
		"123 Main St", "456 Oak Ave", "789 Elm St", "321 Pine Rd",
		"654 Maple Dr", "987 Cedar Ln", "147 Birch Blvd", "258 Walnut Way",
		"369 Cherry Ct", "741 Spruce Pl", "852 Willow Rd", "963 Ash St",
	}
	cities = []string{
		"New York", "Los Angeles", "Chicago", "Houston", "Phoenix",
		"Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose",
		"Austin", "Jacksonville", "Fort Worth", "Columbus", "Charlotte",
	}
	states = []string{
		"NY", "CA", "IL", "TX", "AZ", "PA", "FL", "OH", "NC", "GA",
		"MI", "NJ", "VA", "WA", "CO",
	}
	zips = []string{
		"10001", "90001", "60601", "77001", "85001", "19101", "78201",
		"92101", "75201", "95101", "73301", "32201", "76101", "43201", "28201",
	}

	icd10Conditions = []codeEntry{
		{"E11.9", "Type 2 diabetes mellitus without complications"},
		{"I10", "Essential (primary) hypertension"},
		{"J45.909", "Unspecified asthma, uncomplicated"},
		{"E78.5", "Hyperlipidemia, unspecified"},
		{"J06.9", "Acute upper respiratory infection, unspecified"},
		{"M54.5", "Low back pain"},
		{"F32.9", "Major depressive disorder, single episode, unspecified"},
		{"K21.0", "Gastro-esophageal reflux disease with esophagitis"},
		{"N39.0", "Urinary tract infection, site not specified"},
		{"J20.9", "Acute bronchitis, unspecified"},
		{"E03.9", "Hypothyroidism, unspecified"},
		{"G43.909", "Migraine, unspecified, not intractable"},
		{"M79.3", "Panniculitis, unspecified"},
		{"R05.9", "Cough, unspecified"},
		{"L30.9", "Dermatitis, unspecified"},
		{"K58.9", "Irritable bowel syndrome without diarrhea"},
		{"G47.00", "Insomnia, unspecified"},
		{"J30.9", "Allergic rhinitis, unspecified"},
		{"M25.50", "Pain in unspecified joint"},
		{"R10.9", "Unspecified abdominal pain"},
		{"E55.9", "Vitamin D deficiency, unspecified"},
	}

	loincObservations = []observationDef{
		{"8867-4", "Heart rate", "beats/minute", 50, 110},
		{"8310-5", "Body temperature", "degC", 36.0, 38.5},
		{"29463-7", "Body weight", "kg", 40, 150},
		{"8302-2", "Body height", "cm", 140, 200},
		{"85354-9", "Blood pressure panel", "mmHg", 90, 180},
		{"8480-6", "Systolic blood pressure", "mmHg", 90, 180},
		{"8462-4", "Diastolic blood pressure", "mmHg", 50, 110},
		{"2708-6", "Oxygen saturation", "%", 92, 100},
		{"9279-1", "Respiratory rate", "breaths/minute", 10, 25},
		{"2339-0", "Glucose [Mass/volume] in Blood", "mg/dL", 60, 250},
		{"2093-3", "Total Cholesterol", "mg/dL", 120, 300},
		{"2571-8", "Triglycerides", "mg/dL", 50, 400},
		{"718-7", "Hemoglobin [Mass/volume] in Blood", "g/dL", 10, 18},
		{"4548-4", "Hemoglobin A1c", "%", 4.0, 12.0},
		{"33914-3", "Glomerular filtration rate", "mL/min", 30, 120},
		{"2160-0", "Creatinine [Mass/volume] in Serum", "mg/dL", 0.5, 2.5},
	}

	rxnormMedications = []codeEntry{
		{"197361", "Metformin 500 MG Oral Tablet"},
		{"310798", "Lisinopril 10 MG Oral Tablet"},
		{"197381", "Atorvastatin 20 MG Oral Tablet"},
		{"311700", "Omeprazole 20 MG Delayed Release Oral Capsule"},
		{"308136", "Amoxicillin 500 MG Oral Capsule"},
		{"198211", "Levothyroxine Sodium 0.05 MG Oral Tablet"},
		{"314076", "Amlodipine 5 MG Oral Tablet"},
		{"200801", "Hydrochlorothiazide 25 MG Oral Tablet"},
		{"312961", "Sertraline 50 MG Oral Tablet"},
		{"197591", "Albuterol 0.83 MG/ML Inhalation Solution"},
		{"310965", "Losartan Potassium 50 MG Oral Tablet"},
		{"197517", "Gabapentin 300 MG Oral Capsule"},
		{"308056", "Acetaminophen 500 MG Oral Tablet"},
		{"198240", "Montelukast 10 MG Oral Tablet"},
		{"197446", "Furosemide 40 MG Oral Tablet"},
		{"199026", "Prednisone 10 MG Oral Tablet"},
	}

	cvxVaccines = []codeEntry{
		{"08", "Hepatitis B vaccine"},
		{"20", "DTaP vaccine"},
		{"03", "MMR vaccine"},
		{"21", "Varicella vaccine"},
		{"10", "IPV (Polio) vaccine"},
		{"114", "Meningococcal MCV4P vaccine"},
		{"62", "HPV quadrivalent vaccine"},
		{"141", "Influenza, seasonal, injectable"},
		{"133", "Pneumococcal conjugate PCV 13"},
		{"187", "Recombinant zoster vaccine"},
		{"197", "Influenza, high-dose, quadrivalent"},
		{"213", "SARS-CoV-2 (COVID-19) vaccine, mRNA"},
	}

	snomedAllergies = []codeEntry{
		{"387207008", "Ibuprofen allergy"},
		{"91936005", "Penicillin allergy"},
		{"293586001", "Aspirin allergy"},
		{"294499007", "Sulfonamide allergy"},
		{"418689008", "Latex allergy"},
		{"232350006", "Peanut allergy"},
		{"91935009", "Amoxicillin allergy"},
		{"294505004", "Codeine allergy"},
		{"300913006", "Shellfish allergy"},
		{"419199007", "Egg allergy"},
		{"300916003", "Bee venom allergy"},
	}

	cptProcedures = []codeEntry{
		{"99213", "Office or outpatient visit, established patient, low complexity"},
		{"71046", "Radiologic examination, chest, 2 views"},
		{"80053", "Comprehensive metabolic panel"},
		{"85025", "Complete blood count (CBC) with differential"},
		{"93000", "Electrocardiogram, routine ECG"},
		{"36415", "Venipuncture, routine"},
		{"99214", "Office or outpatient visit, established patient, moderate complexity"},
		{"90837", "Psychotherapy, 60 minutes"},
		{"99203", "Office or outpatient visit, new patient, low complexity"},
		{"81001", "Urinalysis, automated, with microscopy"},
		{"87880", "Strep A assay with direct optical observation"},
		{"99395", "Preventive visit, established patient, 18-39 years"},
	}

	organizationNames = []string{
		"Mercy General Hospital", "St. Luke's Medical Center",
		"Community Health Partners", "Regional Medical Associates",
		"Valley Care Medical Group", "Northside Health System",
		"Lakewood Family Medicine", "Summit Healthcare Network",
		"Riverside Community Hospital", "Beacon Health Alliance",
	}

	maritalStatuses = []string{
		"S", "M", "D", "W", "A",
	}

	encounterClasses = []string{
		"AMB", "EMER", "IMP", "OBSENC", "SS",
	}

	encounterTypes = []codeEntry{
		{"185349003", "Encounter for check up"},
		{"185345009", "Encounter for symptom"},
		{"185347001", "Encounter for problem"},
		{"270427003", "Patient-initiated encounter"},
		{"390906007", "Follow-up encounter"},
	}
)

// ---------------------------------------------------------------------------
// DataGenerator
// ---------------------------------------------------------------------------

// DataGenerator produces deterministic synthetic FHIR resources.
type DataGenerator struct {
	rng     *rand.Rand
	counter uint64
}

// NewDataGenerator returns a generator seeded for reproducibility. If seed is
// 0 a time-based seed is chosen.
func NewDataGenerator(seed int64) *DataGenerator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &DataGenerator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

func (g *DataGenerator) nextID(prefix string) string {
	g.counter++
	return fmt.Sprintf("%s-%08x-%04x", prefix, g.rng.Uint32(), g.counter)
}

func (g *DataGenerator) pick(pool []string) string {
	return pool[g.rng.Intn(len(pool))]
}

func (g *DataGenerator) pickCode(pool []codeEntry) codeEntry {
	return pool[g.rng.Intn(len(pool))]
}

func (g *DataGenerator) pickObs() observationDef {
	return loincObservations[g.rng.Intn(len(loincObservations))]
}

func (g *DataGenerator) randomDate(minYear, maxYear int) string {
	y := minYear + g.rng.Intn(maxYear-minYear+1)
	m := 1 + g.rng.Intn(12)
	d := 1 + g.rng.Intn(28) // safe for all months
	return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
}

func (g *DataGenerator) randomPhone() string {
	return fmt.Sprintf("(%03d) %03d-%04d",
		200+g.rng.Intn(800),
		200+g.rng.Intn(800),
		g.rng.Intn(10000),
	)
}

// GeneratePatient produces a FHIR Patient resource as a generic map.
func (g *DataGenerator) GeneratePatient() map[string]interface{} {
	isMale := g.rng.Intn(2) == 0
	var firstName string
	var gender string
	if isMale {
		firstName = g.pick(firstNamesMale)
		gender = "male"
	} else {
		firstName = g.pick(firstNamesFemale)
		gender = "female"
	}
	lastName := g.pick(lastNames)
	id := g.nextID("pat")

	mrn := fmt.Sprintf("MRN-%08d", g.rng.Intn(100000000))

	return map[string]interface{}{
		"resourceType": "Patient",
		"id":           id,
		"active":       true,
		"name": []interface{}{
			map[string]interface{}{
				"use":    "official",
				"family": lastName,
				"given":  []interface{}{firstName},
			},
		},
		"gender":    gender,
		"birthDate": g.randomDate(1940, 2010),
		"address": []interface{}{
			map[string]interface{}{
				"use":        "home",
				"line":       []interface{}{g.pick(streets)},
				"city":       g.pick(cities),
				"state":      g.pick(states),
				"postalCode": g.pick(zips),
				"country":    "US",
			},
		},
		"telecom": []interface{}{
			map[string]interface{}{
				"system": "phone",
				"value":  g.randomPhone(),
				"use":    "home",
			},
			map[string]interface{}{
				"system": "email",
				"value":  fmt.Sprintf("%s.%s@example.com", firstName, lastName),
				"use":    "home",
			},
		},
		"identifier": []interface{}{
			map[string]interface{}{
				"type": map[string]interface{}{
					"coding": []interface{}{
						map[string]interface{}{
							"system": "http://terminology.hl7.org/CodeSystem/v2-0203",
							"code":   "MR",
						},
					},
				},
				"system": "urn:oid:1.2.36.146.595.217.0.1",
				"value":  mrn,
			},
		},
		"maritalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/v3-MaritalStatus",
					"code":   g.pick(maritalStatuses),
				},
			},
		},
	}
}

// GeneratePractitioner produces a FHIR Practitioner resource.
func (g *DataGenerator) GeneratePractitioner() map[string]interface{} {
	isMale := g.rng.Intn(2) == 0
	var firstName string
	var gender string
	if isMale {
		firstName = g.pick(firstNamesMale)
		gender = "male"
	} else {
		firstName = g.pick(firstNamesFemale)
		gender = "female"
	}
	lastName := g.pick(lastNames)
	id := g.nextID("prac")
	npi := fmt.Sprintf("%010d", g.rng.Int63n(10000000000))

	return map[string]interface{}{
		"resourceType": "Practitioner",
		"id":           id,
		"active":       true,
		"name": []interface{}{
			map[string]interface{}{
				"use":    "official",
				"family": lastName,
				"given":  []interface{}{firstName},
				"prefix": []interface{}{"Dr."},
			},
		},
		"gender": gender,
		"identifier": []interface{}{
			map[string]interface{}{
				"system": "http://hl7.org/fhir/sid/us-npi",
				"value":  npi,
			},
		},
		"telecom": []interface{}{
			map[string]interface{}{
				"system": "phone",
				"value":  g.randomPhone(),
				"use":    "work",
			},
		},
	}
}

// GenerateOrganization produces a FHIR Organization resource.
func (g *DataGenerator) GenerateOrganization() map[string]interface{} {
	id := g.nextID("org")
	return map[string]interface{}{
		"resourceType": "Organization",
		"id":           id,
		"active":       true,
		"name":         g.pick(organizationNames),
		"type": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://terminology.hl7.org/CodeSystem/organization-type",
						"code":    "prov",
						"display": "Healthcare Provider",
					},
				},
			},
		},
		"telecom": []interface{}{
			map[string]interface{}{
				"system": "phone",
				"value":  g.randomPhone(),
				"use":    "work",
			},
		},
		"address": []interface{}{
			map[string]interface{}{
				"line":       []interface{}{g.pick(streets)},
				"city":       g.pick(cities),
				"state":      g.pick(states),
				"postalCode": g.pick(zips),
				"country":    "US",
			},
		},
	}
}

// GenerateEncounter produces a FHIR Encounter resource.
func (g *DataGenerator) GenerateEncounter(patientID, practitionerID string) map[string]interface{} {
	id := g.nextID("enc")
	encType := g.pickCode(encounterTypes)
	return map[string]interface{}{
		"resourceType": "Encounter",
		"id":           id,
		"status":       "finished",
		"class": map[string]interface{}{
			"system":  "http://terminology.hl7.org/CodeSystem/v3-ActCode",
			"code":    g.pick(encounterClasses),
			"display": "ambulatory",
		},
		"type": []interface{}{
			map[string]interface{}{
				"coding": []interface{}{
					map[string]interface{}{
						"system":  "http://snomed.info/sct",
						"code":    encType.Code,
						"display": encType.Display,
					},
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/" + patientID,
		},
		"participant": []interface{}{
			map[string]interface{}{
				"individual": map[string]interface{}{
					"reference": "Practitioner/" + practitionerID,
				},
			},
		},
		"period": map[string]interface{}{
			"start": g.randomDate(2020, 2025) + "T09:00:00Z",
			"end":   g.randomDate(2020, 2025) + "T10:00:00Z",
		},
	}
}

// GenerateObservation produces a FHIR Observation with a LOINC code and numeric value.
func (g *DataGenerator) GenerateObservation(patientID, encounterID string) map[string]interface{} {
	id := g.nextID("obs")
	obs := g.pickObs()
	value := obs.Low + g.rng.Float64()*(obs.High-obs.Low)
	// Round to 1 decimal
	value = float64(int(value*10)) / 10

	return map[string]interface{}{
		"resourceType": "Observation",
		"id":           id,
		"status":       "final",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://loinc.org",
					"code":    obs.Code,
					"display": obs.Display,
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/" + patientID,
		},
		"encounter": map[string]interface{}{
			"reference": "Encounter/" + encounterID,
		},
		"effectiveDateTime": g.randomDate(2020, 2025) + "T10:00:00Z",
		"valueQuantity": map[string]interface{}{
			"value":  value,
			"unit":   obs.Unit,
			"system": "http://unitsofmeasure.org",
			"code":   obs.Unit,
		},
	}
}

// GenerateCondition produces a FHIR Condition with an ICD-10 code.
func (g *DataGenerator) GenerateCondition(patientID string) map[string]interface{} {
	id := g.nextID("cond")
	cond := g.pickCode(icd10Conditions)
	return map[string]interface{}{
		"resourceType":   "Condition",
		"id":             id,
		"clinicalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/condition-clinical",
					"code":   "active",
				},
			},
		},
		"verificationStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/condition-ver-status",
					"code":   "confirmed",
				},
			},
		},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://hl7.org/fhir/sid/icd-10-cm",
					"code":    cond.Code,
					"display": cond.Display,
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/" + patientID,
		},
		"onsetDateTime": g.randomDate(2018, 2025) + "T00:00:00Z",
	}
}

// GenerateMedicationRequest produces a FHIR MedicationRequest with an RxNorm code.
func (g *DataGenerator) GenerateMedicationRequest(patientID, practitionerID string) map[string]interface{} {
	id := g.nextID("medreq")
	med := g.pickCode(rxnormMedications)
	return map[string]interface{}{
		"resourceType": "MedicationRequest",
		"id":           id,
		"status":       "active",
		"intent":       "order",
		"medicationCodeableConcept": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://www.nlm.nih.gov/research/umls/rxnorm",
					"code":    med.Code,
					"display": med.Display,
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/" + patientID,
		},
		"requester": map[string]interface{}{
			"reference": "Practitioner/" + practitionerID,
		},
		"authoredOn": g.randomDate(2020, 2025),
		"dosageInstruction": []interface{}{
			map[string]interface{}{
				"text": "Take as directed",
				"timing": map[string]interface{}{
					"repeat": map[string]interface{}{
						"frequency": 1,
						"period":    1,
						"periodUnit": "d",
					},
				},
			},
		},
	}
}

// GenerateAllergyIntolerance produces a FHIR AllergyIntolerance with a SNOMED code.
func (g *DataGenerator) GenerateAllergyIntolerance(patientID string) map[string]interface{} {
	id := g.nextID("allergy")
	allergy := g.pickCode(snomedAllergies)
	return map[string]interface{}{
		"resourceType":   "AllergyIntolerance",
		"id":             id,
		"clinicalStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/allergyintolerance-clinical",
					"code":   "active",
				},
			},
		},
		"verificationStatus": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system": "http://terminology.hl7.org/CodeSystem/allergyintolerance-verification",
					"code":   "confirmed",
				},
			},
		},
		"type":     "allergy",
		"category": []interface{}{"medication"},
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://snomed.info/sct",
					"code":    allergy.Code,
					"display": allergy.Display,
				},
			},
		},
		"patient": map[string]interface{}{
			"reference": "Patient/" + patientID,
		},
		"recordedDate": g.randomDate(2018, 2025),
	}
}

// GenerateProcedure produces a FHIR Procedure with a CPT code.
func (g *DataGenerator) GenerateProcedure(patientID, encounterID string) map[string]interface{} {
	id := g.nextID("proc")
	proc := g.pickCode(cptProcedures)
	return map[string]interface{}{
		"resourceType": "Procedure",
		"id":           id,
		"status":       "completed",
		"code": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://www.ama-assn.org/go/cpt",
					"code":    proc.Code,
					"display": proc.Display,
				},
			},
		},
		"subject": map[string]interface{}{
			"reference": "Patient/" + patientID,
		},
		"encounter": map[string]interface{}{
			"reference": "Encounter/" + encounterID,
		},
		"performedDateTime": g.randomDate(2020, 2025) + "T11:00:00Z",
	}
}

// GenerateImmunization produces a FHIR Immunization with a CVX code.
func (g *DataGenerator) GenerateImmunization(patientID string) map[string]interface{} {
	id := g.nextID("imm")
	vax := g.pickCode(cvxVaccines)
	return map[string]interface{}{
		"resourceType": "Immunization",
		"id":           id,
		"status":       "completed",
		"vaccineCode": map[string]interface{}{
			"coding": []interface{}{
				map[string]interface{}{
					"system":  "http://hl7.org/fhir/sid/cvx",
					"code":    vax.Code,
					"display": vax.Display,
				},
			},
		},
		"patient": map[string]interface{}{
			"reference": "Patient/" + patientID,
		},
		"occurrenceDateTime": g.randomDate(2015, 2025),
		"primarySource":      true,
	}
}

// ---------------------------------------------------------------------------
// Seeder — orchestrates full data generation
// ---------------------------------------------------------------------------

// Seeder orchestrates generation of a complete set of synthetic resources.
type Seeder struct {
	generator *DataGenerator
	config    SeedConfig
	mu        sync.RWMutex
	resources map[string][]map[string]interface{}
}

// NewSeeder creates a new Seeder with the given config.
func NewSeeder(config SeedConfig) *Seeder {
	seed := config.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Seeder{
		generator: NewDataGenerator(seed),
		config:    config,
		resources: make(map[string][]map[string]interface{}),
	}
}

// Generate creates all synthetic resources according to config.
func (s *Seeder) Generate() (*SeedResult, error) {
	start := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reset
	s.resources = make(map[string][]map[string]interface{})

	result := &SeedResult{}

	// Organizations
	if s.config.IncludeOrganization {
		org := s.generator.GenerateOrganization()
		s.resources["Organization"] = append(s.resources["Organization"], org)
		result.Organizations = 1
	}

	// Practitioners
	var practitionerIDs []string
	if s.config.IncludePractitioners {
		for i := 0; i < s.config.PractitionerCount; i++ {
			prac := s.generator.GeneratePractitioner()
			s.resources["Practitioner"] = append(s.resources["Practitioner"], prac)
			practitionerIDs = append(practitionerIDs, prac["id"].(string))
		}
		result.Practitioners = s.config.PractitionerCount
	}

	// Patients and their clinical data
	for i := 0; i < s.config.PatientCount; i++ {
		patient := s.generator.GeneratePatient()
		patientID := patient["id"].(string)
		s.resources["Patient"] = append(s.resources["Patient"], patient)

		// Pick a practitioner for this patient (round-robin if available)
		practID := ""
		if len(practitionerIDs) > 0 {
			practID = practitionerIDs[i%len(practitionerIDs)]
		}

		// Encounters
		var encounterIDs []string
		for j := 0; j < s.config.EncountersPerPatient; j++ {
			enc := s.generator.GenerateEncounter(patientID, practID)
			s.resources["Encounter"] = append(s.resources["Encounter"], enc)
			encounterIDs = append(encounterIDs, enc["id"].(string))

			// Observations per encounter
			for k := 0; k < s.config.ObservationsPerEncounter; k++ {
				obs := s.generator.GenerateObservation(patientID, enc["id"].(string))
				s.resources["Observation"] = append(s.resources["Observation"], obs)
				result.Observations++
			}
		}
		result.Encounters += s.config.EncountersPerPatient

		// Conditions
		for j := 0; j < s.config.ConditionsPerPatient; j++ {
			cond := s.generator.GenerateCondition(patientID)
			s.resources["Condition"] = append(s.resources["Condition"], cond)
		}
		result.Conditions += s.config.ConditionsPerPatient

		// Medications
		for j := 0; j < s.config.MedicationsPerPatient; j++ {
			med := s.generator.GenerateMedicationRequest(patientID, practID)
			s.resources["MedicationRequest"] = append(s.resources["MedicationRequest"], med)
		}
		result.Medications += s.config.MedicationsPerPatient

		// Allergies
		for j := 0; j < s.config.AllergiesPerPatient; j++ {
			allergy := s.generator.GenerateAllergyIntolerance(patientID)
			s.resources["AllergyIntolerance"] = append(s.resources["AllergyIntolerance"], allergy)
		}
		result.Allergies += s.config.AllergiesPerPatient

		// Procedures (use first encounter if available)
		for j := 0; j < s.config.ProceduresPerPatient; j++ {
			encID := ""
			if len(encounterIDs) > 0 {
				encID = encounterIDs[j%len(encounterIDs)]
			}
			proc := s.generator.GenerateProcedure(patientID, encID)
			s.resources["Procedure"] = append(s.resources["Procedure"], proc)
		}
		result.Procedures += s.config.ProceduresPerPatient

		// Immunizations
		for j := 0; j < s.config.ImmunizationsPerPatient; j++ {
			imm := s.generator.GenerateImmunization(patientID)
			s.resources["Immunization"] = append(s.resources["Immunization"], imm)
		}
		result.Immunizations += s.config.ImmunizationsPerPatient
	}

	result.Patients = s.config.PatientCount
	result.TotalResources = result.Patients + result.Practitioners + result.Organizations +
		result.Encounters + result.Observations + result.Conditions +
		result.Medications + result.Allergies + result.Procedures + result.Immunizations
	result.Duration = time.Since(start)

	return result, nil
}

// GetResources returns generated resources of the given FHIR type.
func (s *Seeder) GetResources(resourceType string) []map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.resources[resourceType]
}

// ExportNDJSON writes resources of the given type as newline-delimited JSON.
func (s *Seeder) ExportNDJSON(w io.Writer, resourceType string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := s.resources[resourceType]
	enc := json.NewEncoder(w)
	for _, r := range resources {
		if err := enc.Encode(r); err != nil {
			return fmt.Errorf("encoding %s: %w", resourceType, err)
		}
	}
	return nil
}

// ExportBundle writes all generated resources as a FHIR Transaction Bundle.
func (s *Seeder) ExportBundle(w io.Writer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var entries []interface{}

	// Deterministic ordering: Organization, Practitioner, Patient, then clinical
	typeOrder := []string{
		"Organization", "Practitioner", "Patient", "Encounter",
		"Observation", "Condition", "MedicationRequest",
		"AllergyIntolerance", "Procedure", "Immunization",
	}

	for _, rt := range typeOrder {
		for _, r := range s.resources[rt] {
			resType := r["resourceType"].(string)
			resID := r["id"].(string)
			entries = append(entries, map[string]interface{}{
				"fullUrl":  fmt.Sprintf("urn:uuid:%s", resID),
				"resource": r,
				"request": map[string]interface{}{
					"method": "POST",
					"url":    resType,
				},
			})
		}
	}

	bundle := map[string]interface{}{
		"resourceType": "Bundle",
		"type":         "transaction",
		"entry":        entries,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(bundle)
}

// ---------------------------------------------------------------------------
// SeedHandler — Echo HTTP handlers
// ---------------------------------------------------------------------------

// SeedHandler provides HTTP endpoints for sandbox data management.
type SeedHandler struct {
	seeder *Seeder
	mu     sync.Mutex
}

// NewSeedHandler creates a new handler with no pre-seeded data.
func NewSeedHandler() *SeedHandler {
	return &SeedHandler{}
}

// RegisterRoutes registers sandbox routes on the given Echo group.
func (h *SeedHandler) RegisterRoutes(g *echo.Group) {
	g.POST("/seed", h.handleSeed)
	g.GET("/resources/:type", h.handleListResources)
	g.POST("/reset", h.handleReset)
	g.GET("/export/ndjson/:type", h.handleExportNDJSON)
	g.GET("/export/bundle", h.handleExportBundle)
}

func (h *SeedHandler) handleSeed(c echo.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var cfg SeedConfig
	if err := c.Bind(&cfg); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Apply defaults for zero values
	if cfg.PatientCount == 0 {
		cfg.PatientCount = 10
	}

	h.seeder = NewSeeder(cfg)
	result, err := h.seeder.Generate()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}

func (h *SeedHandler) handleListResources(c echo.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.seeder == nil {
		return c.JSON(http.StatusOK, []interface{}{})
	}

	resourceType := c.Param("type")
	resources := h.seeder.GetResources(resourceType)
	if resources == nil {
		resources = []map[string]interface{}{}
	}
	return c.JSON(http.StatusOK, resources)
}

func (h *SeedHandler) handleReset(c echo.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.seeder != nil {
		h.seeder.mu.Lock()
		h.seeder.resources = make(map[string][]map[string]interface{})
		h.seeder.mu.Unlock()
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "reset"})
}

func (h *SeedHandler) handleExportNDJSON(c echo.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.seeder == nil {
		return c.String(http.StatusOK, "")
	}

	resourceType := c.Param("type")
	c.Response().Header().Set(echo.HeaderContentType, "application/x-ndjson")
	c.Response().WriteHeader(http.StatusOK)

	return h.seeder.ExportNDJSON(c.Response().Writer, resourceType)
}

func (h *SeedHandler) handleExportBundle(c echo.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.seeder == nil {
		empty := map[string]interface{}{
			"resourceType": "Bundle",
			"type":         "transaction",
			"entry":        []interface{}{},
		}
		return c.JSON(http.StatusOK, empty)
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)
	return h.seeder.ExportBundle(c.Response().Writer)
}
