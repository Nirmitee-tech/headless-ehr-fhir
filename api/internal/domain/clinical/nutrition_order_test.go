package clinical

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ── Helpers ──

func newTestNutritionOrderService() *NutritionOrderService {
	repo := NewInMemoryNutritionOrderRepo()
	return NewNutritionOrderService(repo)
}

func newTestNutritionOrderHandler() (*NutritionOrderHandler, *echo.Echo) {
	svc := newTestNutritionOrderService()
	h := NewNutritionOrderHandler(svc)
	e := echo.New()
	return h, e
}

func makeNutritionOrder(patientID uuid.UUID) *NutritionOrder {
	return &NutritionOrder{
		PatientID: patientID,
		Status:    "draft",
		Intent:    "order",
		DateTime:  time.Now().UTC().Truncate(time.Second),
	}
}

// ── Repository CRUD Tests ──

func TestNutritionOrderRepo_Create(t *testing.T) {
	repo := NewInMemoryNutritionOrderRepo()
	order := makeNutritionOrder(uuid.New())

	err := repo.Create(context.Background(), order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.ID == uuid.Nil {
		t.Error("expected ID to be assigned")
	}
}

func TestNutritionOrderRepo_GetByID(t *testing.T) {
	repo := NewInMemoryNutritionOrderRepo()
	order := makeNutritionOrder(uuid.New())
	repo.Create(context.Background(), order)

	got, err := repo.GetByID(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != order.ID {
		t.Errorf("got ID %v, want %v", got.ID, order.ID)
	}
}

func TestNutritionOrderRepo_GetByID_NotFound(t *testing.T) {
	repo := NewInMemoryNutritionOrderRepo()
	_, err := repo.GetByID(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected not-found error")
	}
}

func TestNutritionOrderRepo_Update(t *testing.T) {
	repo := NewInMemoryNutritionOrderRepo()
	order := makeNutritionOrder(uuid.New())
	repo.Create(context.Background(), order)

	order.Status = "active"
	order.Note = "updated note"
	err := repo.Update(context.Background(), order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := repo.GetByID(context.Background(), order.ID)
	if got.Status != "active" {
		t.Errorf("status = %v, want active", got.Status)
	}
	if got.Note != "updated note" {
		t.Errorf("note = %v, want updated note", got.Note)
	}
}

func TestNutritionOrderRepo_Update_NotFound(t *testing.T) {
	repo := NewInMemoryNutritionOrderRepo()
	order := &NutritionOrder{ID: uuid.New(), PatientID: uuid.New(), Status: "draft", Intent: "order"}
	err := repo.Update(context.Background(), order)
	if err == nil {
		t.Error("expected not-found error on update")
	}
}

func TestNutritionOrderRepo_Delete(t *testing.T) {
	repo := NewInMemoryNutritionOrderRepo()
	order := makeNutritionOrder(uuid.New())
	repo.Create(context.Background(), order)

	err := repo.Delete(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.GetByID(context.Background(), order.ID)
	if err == nil {
		t.Error("expected not-found error after delete")
	}
}

func TestNutritionOrderRepo_Delete_NotFound(t *testing.T) {
	repo := NewInMemoryNutritionOrderRepo()
	err := repo.Delete(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected not-found error on delete of nonexistent")
	}
}

func TestNutritionOrderRepo_ListByPatient(t *testing.T) {
	repo := NewInMemoryNutritionOrderRepo()
	ctx := context.Background()
	patientA := uuid.New()
	patientB := uuid.New()

	repo.Create(ctx, makeNutritionOrder(patientA))
	repo.Create(ctx, makeNutritionOrder(patientA))
	repo.Create(ctx, makeNutritionOrder(patientB))

	orders, total, err := repo.ListByPatient(ctx, patientA, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(orders) != 2 {
		t.Errorf("len(orders) = %d, want 2", len(orders))
	}
}

func TestNutritionOrderRepo_ListByEncounter(t *testing.T) {
	repo := NewInMemoryNutritionOrderRepo()
	ctx := context.Background()

	encA := uuid.New()
	encB := uuid.New()

	o1 := makeNutritionOrder(uuid.New())
	o1.EncounterID = &encA
	repo.Create(ctx, o1)

	o2 := makeNutritionOrder(uuid.New())
	o2.EncounterID = &encA
	repo.Create(ctx, o2)

	o3 := makeNutritionOrder(uuid.New())
	o3.EncounterID = &encB
	repo.Create(ctx, o3)

	orders, total, err := repo.ListByEncounter(ctx, encA, 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(orders) != 2 {
		t.Errorf("len(orders) = %d, want 2", len(orders))
	}
}

func TestNutritionOrderRepo_ListByPatient_Pagination(t *testing.T) {
	repo := NewInMemoryNutritionOrderRepo()
	ctx := context.Background()
	patientID := uuid.New()

	for i := 0; i < 5; i++ {
		repo.Create(ctx, makeNutritionOrder(patientID))
	}

	orders, total, err := repo.ListByPatient(ctx, patientID, 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
	if len(orders) != 2 {
		t.Errorf("len(orders) = %d, want 2", len(orders))
	}
}

// ── Service Validation Tests ──

func TestNutritionOrderService_Create_RequiresPatientID(t *testing.T) {
	svc := newTestNutritionOrderService()
	order := &NutritionOrder{Status: "draft", Intent: "order", DateTime: time.Now()}
	err := svc.Create(context.Background(), order)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestNutritionOrderService_Create_RequiresIntent(t *testing.T) {
	svc := newTestNutritionOrderService()
	order := &NutritionOrder{PatientID: uuid.New(), Status: "draft", DateTime: time.Now()}
	err := svc.Create(context.Background(), order)
	if err == nil {
		t.Error("expected error for missing intent")
	}
}

func TestNutritionOrderService_Create_InvalidStatus(t *testing.T) {
	svc := newTestNutritionOrderService()
	order := &NutritionOrder{PatientID: uuid.New(), Status: "bogus", Intent: "order", DateTime: time.Now()}
	err := svc.Create(context.Background(), order)
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestNutritionOrderService_Create_InvalidIntent(t *testing.T) {
	svc := newTestNutritionOrderService()
	order := &NutritionOrder{PatientID: uuid.New(), Status: "draft", Intent: "bogus", DateTime: time.Now()}
	err := svc.Create(context.Background(), order)
	if err == nil {
		t.Error("expected error for invalid intent")
	}
}

func TestNutritionOrderService_Create_DefaultStatus(t *testing.T) {
	svc := newTestNutritionOrderService()
	order := &NutritionOrder{PatientID: uuid.New(), Intent: "order", DateTime: time.Now()}
	err := svc.Create(context.Background(), order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Status != "draft" {
		t.Errorf("status = %v, want draft", order.Status)
	}
}

func TestNutritionOrderService_Create_Success(t *testing.T) {
	svc := newTestNutritionOrderService()
	order := makeNutritionOrder(uuid.New())
	err := svc.Create(context.Background(), order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.ID == uuid.Nil {
		t.Error("expected ID to be set")
	}
}

// ── Status Transition Tests ──

func TestNutritionOrderService_StatusTransition_DraftToActive(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()
	order := makeNutritionOrder(uuid.New())
	svc.Create(ctx, order)

	err := svc.UpdateStatus(ctx, order.ID, "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetByID(ctx, order.ID)
	if got.Status != "active" {
		t.Errorf("status = %v, want active", got.Status)
	}
}

func TestNutritionOrderService_StatusTransition_ActiveToCompleted(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()
	order := makeNutritionOrder(uuid.New())
	svc.Create(ctx, order)
	svc.UpdateStatus(ctx, order.ID, "active")

	err := svc.UpdateStatus(ctx, order.ID, "completed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetByID(ctx, order.ID)
	if got.Status != "completed" {
		t.Errorf("status = %v, want completed", got.Status)
	}
}

func TestNutritionOrderService_StatusTransition_ActiveToOnHold(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()
	order := makeNutritionOrder(uuid.New())
	svc.Create(ctx, order)
	svc.UpdateStatus(ctx, order.ID, "active")

	err := svc.UpdateStatus(ctx, order.ID, "on-hold")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetByID(ctx, order.ID)
	if got.Status != "on-hold" {
		t.Errorf("status = %v, want on-hold", got.Status)
	}
}

func TestNutritionOrderService_StatusTransition_OnHoldToActive(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()
	order := makeNutritionOrder(uuid.New())
	svc.Create(ctx, order)
	svc.UpdateStatus(ctx, order.ID, "active")
	svc.UpdateStatus(ctx, order.ID, "on-hold")

	err := svc.UpdateStatus(ctx, order.ID, "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetByID(ctx, order.ID)
	if got.Status != "active" {
		t.Errorf("status = %v, want active", got.Status)
	}
}

func TestNutritionOrderService_StatusTransition_DraftToRevoked(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()
	order := makeNutritionOrder(uuid.New())
	svc.Create(ctx, order)

	err := svc.UpdateStatus(ctx, order.ID, "revoked")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetByID(ctx, order.ID)
	if got.Status != "revoked" {
		t.Errorf("status = %v, want revoked", got.Status)
	}
}

func TestNutritionOrderService_StatusTransition_ActiveToEnteredInError(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()
	order := makeNutritionOrder(uuid.New())
	svc.Create(ctx, order)
	svc.UpdateStatus(ctx, order.ID, "active")

	err := svc.UpdateStatus(ctx, order.ID, "entered-in-error")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := svc.GetByID(ctx, order.ID)
	if got.Status != "entered-in-error" {
		t.Errorf("status = %v, want entered-in-error", got.Status)
	}
}

func TestNutritionOrderService_StatusTransition_InvalidTransition(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()
	order := makeNutritionOrder(uuid.New())
	svc.Create(ctx, order)
	svc.UpdateStatus(ctx, order.ID, "active")
	svc.UpdateStatus(ctx, order.ID, "completed")

	// completed -> active should fail
	err := svc.UpdateStatus(ctx, order.ID, "active")
	if err == nil {
		t.Error("expected error for invalid transition from completed to active")
	}
}

func TestNutritionOrderService_StatusTransition_CompletedToOnHold(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()
	order := makeNutritionOrder(uuid.New())
	svc.Create(ctx, order)
	svc.UpdateStatus(ctx, order.ID, "active")
	svc.UpdateStatus(ctx, order.ID, "completed")

	// completed -> on-hold should fail
	err := svc.UpdateStatus(ctx, order.ID, "on-hold")
	if err == nil {
		t.Error("expected error for invalid transition from completed to on-hold")
	}
}

// ── Oral Diet Tests ──

func TestNutritionOrder_OralDiet_WithNutrientsAndTextures(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()

	order := makeNutritionOrder(uuid.New())
	order.OralDiet = &OralDiet{
		Type: []CodeableConcept{
			{System: "http://snomed.info/sct", Code: "160674003", Display: "Low sodium diet"},
		},
		Schedule: "TID with meals",
		Nutrients: []NutrientComponent{
			{
				Modifier: CodeableConcept{System: "http://snomed.info/sct", Code: "39972003", Display: "Sodium"},
				Amount:   "2 g",
			},
			{
				Modifier: CodeableConcept{System: "http://snomed.info/sct", Code: "88480006", Display: "Potassium"},
				Amount:   "40 mEq",
			},
		},
		TextureModifiers: []TextureModifier{
			{
				Modifier: CodeableConcept{System: "http://snomed.info/sct", Code: "228049004", Display: "Chopped food"},
				FoodType: CodeableConcept{System: "http://snomed.info/sct", Code: "28647000", Display: "Meat"},
			},
		},
		FluidConsistency: []CodeableConcept{
			{System: "http://snomed.info/sct", Code: "439021000124105", Display: "Nectar thick"},
		},
		Instruction: "Patient requires low sodium with chopped texture.",
	}

	err := svc.Create(ctx, order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetByID(ctx, order.ID)
	if got.OralDiet == nil {
		t.Fatal("expected oral diet to be set")
	}
	if len(got.OralDiet.Type) != 1 {
		t.Errorf("oral diet type count = %d, want 1", len(got.OralDiet.Type))
	}
	if len(got.OralDiet.Nutrients) != 2 {
		t.Errorf("nutrient count = %d, want 2", len(got.OralDiet.Nutrients))
	}
	if got.OralDiet.Nutrients[0].Amount != "2 g" {
		t.Errorf("nutrient amount = %v, want 2 g", got.OralDiet.Nutrients[0].Amount)
	}
	if len(got.OralDiet.TextureModifiers) != 1 {
		t.Errorf("texture modifier count = %d, want 1", len(got.OralDiet.TextureModifiers))
	}
	if len(got.OralDiet.FluidConsistency) != 1 {
		t.Errorf("fluid consistency count = %d, want 1", len(got.OralDiet.FluidConsistency))
	}
	if got.OralDiet.Instruction != "Patient requires low sodium with chopped texture." {
		t.Errorf("instruction = %v", got.OralDiet.Instruction)
	}
}

// ── Supplement Tests ──

func TestNutritionOrder_Supplements(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()

	order := makeNutritionOrder(uuid.New())
	order.Supplement = []Supplement{
		{
			Type:        CodeableConcept{System: "http://snomed.info/sct", Code: "443051000124104", Display: "Adult diabetic formula"},
			ProductName: "Glucerna 1.0",
			Schedule:    "BID between meals",
			Quantity:    "240 mL",
			Instruction: "Serve chilled",
		},
		{
			Type:        CodeableConcept{System: "http://snomed.info/sct", Code: "442651000124102", Display: "Standard enteral formula"},
			ProductName: "Ensure Plus",
			Schedule:    "TID with meals",
			Quantity:    "240 mL",
			Instruction: "May add flavor packet",
		},
	}

	err := svc.Create(ctx, order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetByID(ctx, order.ID)
	if len(got.Supplement) != 2 {
		t.Fatalf("supplement count = %d, want 2", len(got.Supplement))
	}
	if got.Supplement[0].ProductName != "Glucerna 1.0" {
		t.Errorf("supplement product = %v, want Glucerna 1.0", got.Supplement[0].ProductName)
	}
	if got.Supplement[1].Quantity != "240 mL" {
		t.Errorf("supplement quantity = %v, want 240 mL", got.Supplement[1].Quantity)
	}
}

// ── Enteral Formula Tests ──

func TestNutritionOrder_EnteralFormula(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()

	order := makeNutritionOrder(uuid.New())
	order.EnteralFormula = &EnteralFormula{
		BaseFormulaType: CodeableConcept{System: "http://snomed.info/sct", Code: "443051000124104", Display: "Adult diabetic formula"},
		AdditiveType:    CodeableConcept{System: "http://terminology.hl7.org/CodeSystem/entformula-additive", Code: "lipid", Display: "Lipid Emulsion"},
		CaloricDensity:  "1 cal/mL",
		RouteOfAdministration: CodeableConcept{
			System:  "http://terminology.hl7.org/CodeSystem/v3-RouteOfAdministration",
			Code:    "GT",
			Display: "Gastrostomy tube",
		},
		Administration: []EnteralAdministration{
			{Schedule: "Continuous", Rate: "60 mL/hr", Quantity: ""},
			{Schedule: "Q4H bolus", Rate: "", Quantity: "240 mL"},
		},
	}

	err := svc.Create(ctx, order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetByID(ctx, order.ID)
	if got.EnteralFormula == nil {
		t.Fatal("expected enteral formula to be set")
	}
	if got.EnteralFormula.BaseFormulaType.Code != "443051000124104" {
		t.Errorf("base formula code = %v", got.EnteralFormula.BaseFormulaType.Code)
	}
	if got.EnteralFormula.CaloricDensity != "1 cal/mL" {
		t.Errorf("caloric density = %v", got.EnteralFormula.CaloricDensity)
	}
	if got.EnteralFormula.RouteOfAdministration.Code != "GT" {
		t.Errorf("route code = %v", got.EnteralFormula.RouteOfAdministration.Code)
	}
	if len(got.EnteralFormula.Administration) != 2 {
		t.Fatalf("administration count = %d, want 2", len(got.EnteralFormula.Administration))
	}
	if got.EnteralFormula.Administration[0].Rate != "60 mL/hr" {
		t.Errorf("admin rate = %v, want 60 mL/hr", got.EnteralFormula.Administration[0].Rate)
	}
}

// ── Food Preference / Exclusion Modifiers Tests ──

func TestNutritionOrder_FoodPreferenceModifiers(t *testing.T) {
	svc := newTestNutritionOrderService()
	ctx := context.Background()

	order := makeNutritionOrder(uuid.New())
	order.FoodPreferenceModifiers = []string{"kosher", "dairy-free"}
	order.ExcludeFoodModifiers = []string{"shellfish", "peanuts", "tree nuts"}
	order.AllergyIntolerances = []string{"AllergyIntolerance/allergy-001", "AllergyIntolerance/allergy-002"}

	err := svc.Create(ctx, order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := svc.GetByID(ctx, order.ID)
	if len(got.FoodPreferenceModifiers) != 2 {
		t.Errorf("food pref count = %d, want 2", len(got.FoodPreferenceModifiers))
	}
	if got.FoodPreferenceModifiers[0] != "kosher" {
		t.Errorf("food pref[0] = %v, want kosher", got.FoodPreferenceModifiers[0])
	}
	if len(got.ExcludeFoodModifiers) != 3 {
		t.Errorf("exclude food count = %d, want 3", len(got.ExcludeFoodModifiers))
	}
	if len(got.AllergyIntolerances) != 2 {
		t.Errorf("allergy ref count = %d, want 2", len(got.AllergyIntolerances))
	}
}

// ── FHIR Mapping Tests ──

func TestNutritionOrder_ToFHIR_BasicFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	patientID := uuid.New()
	encID := uuid.New()
	ordererID := uuid.New()

	order := &NutritionOrder{
		ID:          uuid.New(),
		PatientID:   patientID,
		EncounterID: &encID,
		Orderer:     &ordererID,
		Status:      "active",
		Intent:      "order",
		DateTime:    now,
		Note:        "Test note",
		CreatedAt:   now,
		UpdatedAt:   now,
		FoodPreferenceModifiers: []string{"kosher"},
		ExcludeFoodModifiers:    []string{"shellfish"},
		AllergyIntolerances:     []string{"AllergyIntolerance/ai-001"},
	}

	fhirMap := order.ToFHIR()

	if fhirMap["resourceType"] != "NutritionOrder" {
		t.Errorf("resourceType = %v, want NutritionOrder", fhirMap["resourceType"])
	}
	if fhirMap["id"] != order.ID.String() {
		t.Errorf("id = %v, want %v", fhirMap["id"], order.ID.String())
	}
	if fhirMap["status"] != "active" {
		t.Errorf("status = %v, want active", fhirMap["status"])
	}
	if fhirMap["intent"] != "order" {
		t.Errorf("intent = %v, want order", fhirMap["intent"])
	}
	if fhirMap["dateTime"] != now.Format(time.RFC3339) {
		t.Errorf("dateTime = %v, want %v", fhirMap["dateTime"], now.Format(time.RFC3339))
	}

	// patient reference
	patient, ok := fhirMap["patient"].(map[string]string)
	if !ok {
		t.Fatal("patient is not map[string]string")
	}
	if patient["reference"] != "Patient/"+patientID.String() {
		t.Errorf("patient.reference = %v", patient["reference"])
	}

	// encounter reference
	enc, ok := fhirMap["encounter"].(map[string]string)
	if !ok {
		t.Fatal("encounter missing")
	}
	if enc["reference"] != "Encounter/"+encID.String() {
		t.Errorf("encounter.reference = %v", enc["reference"])
	}

	// orderer reference
	orderer, ok := fhirMap["orderer"].(map[string]string)
	if !ok {
		t.Fatal("orderer missing")
	}
	if orderer["reference"] != "Practitioner/"+ordererID.String() {
		t.Errorf("orderer.reference = %v", orderer["reference"])
	}

	// note
	notes, ok := fhirMap["note"].([]map[string]string)
	if !ok || len(notes) == 0 {
		t.Fatal("note missing")
	}
	if notes[0]["text"] != "Test note" {
		t.Errorf("note text = %v", notes[0]["text"])
	}

	// foodPreferenceModifier
	fpMods, ok := fhirMap["foodPreferenceModifier"].([]map[string]interface{})
	if !ok || len(fpMods) != 1 {
		t.Fatalf("foodPreferenceModifier missing or wrong length: %v", fhirMap["foodPreferenceModifier"])
	}

	// excludeFoodModifier
	exMods, ok := fhirMap["excludeFoodModifier"].([]map[string]interface{})
	if !ok || len(exMods) != 1 {
		t.Fatalf("excludeFoodModifier missing or wrong length")
	}

	// allergyIntolerance
	aiRefs, ok := fhirMap["allergyIntolerance"].([]map[string]string)
	if !ok || len(aiRefs) != 1 {
		t.Fatalf("allergyIntolerance missing or wrong length")
	}
}

func TestNutritionOrder_ToFHIR_OralDiet(t *testing.T) {
	order := &NutritionOrder{
		ID:        uuid.New(),
		PatientID: uuid.New(),
		Status:    "active",
		Intent:    "order",
		DateTime:  time.Now(),
		UpdatedAt: time.Now(),
		OralDiet: &OralDiet{
			Type: []CodeableConcept{
				{System: "http://snomed.info/sct", Code: "160674003", Display: "Low sodium diet"},
			},
			Schedule: "TID",
			Nutrients: []NutrientComponent{
				{Modifier: CodeableConcept{Code: "39972003", Display: "Sodium"}, Amount: "2 g"},
			},
			TextureModifiers: []TextureModifier{
				{
					Modifier: CodeableConcept{Code: "228049004", Display: "Chopped food"},
					FoodType: CodeableConcept{Code: "28647000", Display: "Meat"},
				},
			},
			FluidConsistency: []CodeableConcept{
				{Code: "439021000124105", Display: "Nectar thick"},
			},
			Instruction: "Low sodium",
		},
	}

	fhirMap := order.ToFHIR()
	oralDiet, ok := fhirMap["oralDiet"].(map[string]interface{})
	if !ok {
		t.Fatal("oralDiet missing from FHIR output")
	}
	if oralDiet["instruction"] != "Low sodium" {
		t.Errorf("oralDiet.instruction = %v", oralDiet["instruction"])
	}

	// Check type array
	dietTypes, ok := oralDiet["type"].([]map[string]interface{})
	if !ok || len(dietTypes) != 1 {
		t.Fatalf("oralDiet.type missing or wrong length")
	}

	// Check nutrients
	nutrients, ok := oralDiet["nutrient"].([]map[string]interface{})
	if !ok || len(nutrients) != 1 {
		t.Fatalf("oralDiet.nutrient missing or wrong length")
	}

	// Check texture
	textures, ok := oralDiet["texture"].([]map[string]interface{})
	if !ok || len(textures) != 1 {
		t.Fatalf("oralDiet.texture missing or wrong length")
	}

	// Check fluidConsistencyType
	fluids, ok := oralDiet["fluidConsistencyType"].([]map[string]interface{})
	if !ok || len(fluids) != 1 {
		t.Fatalf("oralDiet.fluidConsistencyType missing or wrong length")
	}
}

func TestNutritionOrder_ToFHIR_Supplement(t *testing.T) {
	order := &NutritionOrder{
		ID:        uuid.New(),
		PatientID: uuid.New(),
		Status:    "active",
		Intent:    "order",
		DateTime:  time.Now(),
		UpdatedAt: time.Now(),
		Supplement: []Supplement{
			{
				Type:        CodeableConcept{Code: "443051000124104", Display: "Adult diabetic formula"},
				ProductName: "Glucerna",
				Schedule:    "BID",
				Quantity:    "240 mL",
				Instruction: "Serve chilled",
			},
		},
	}

	fhirMap := order.ToFHIR()
	supps, ok := fhirMap["supplement"].([]map[string]interface{})
	if !ok || len(supps) != 1 {
		t.Fatal("supplement missing from FHIR output")
	}
	if supps[0]["productName"] != "Glucerna" {
		t.Errorf("supplement.productName = %v", supps[0]["productName"])
	}
	if supps[0]["instruction"] != "Serve chilled" {
		t.Errorf("supplement.instruction = %v", supps[0]["instruction"])
	}
}

func TestNutritionOrder_ToFHIR_EnteralFormula(t *testing.T) {
	order := &NutritionOrder{
		ID:        uuid.New(),
		PatientID: uuid.New(),
		Status:    "active",
		Intent:    "order",
		DateTime:  time.Now(),
		UpdatedAt: time.Now(),
		EnteralFormula: &EnteralFormula{
			BaseFormulaType:       CodeableConcept{Code: "443051000124104", Display: "Adult diabetic formula"},
			AdditiveType:          CodeableConcept{Code: "lipid", Display: "Lipid Emulsion"},
			CaloricDensity:        "1 cal/mL",
			RouteOfAdministration: CodeableConcept{Code: "GT", Display: "Gastrostomy tube"},
			Administration: []EnteralAdministration{
				{Schedule: "Continuous", Rate: "60 mL/hr"},
			},
		},
	}

	fhirMap := order.ToFHIR()
	ef, ok := fhirMap["enteralFormula"].(map[string]interface{})
	if !ok {
		t.Fatal("enteralFormula missing from FHIR output")
	}
	if ef["caloricDensity"] != "1 cal/mL" {
		t.Errorf("caloricDensity = %v", ef["caloricDensity"])
	}

	admins, ok := ef["administration"].([]map[string]interface{})
	if !ok || len(admins) != 1 {
		t.Fatalf("administration missing or wrong length")
	}
}

func TestNutritionOrder_ToFHIR_OptionalFieldsOmitted(t *testing.T) {
	order := &NutritionOrder{
		ID:        uuid.New(),
		PatientID: uuid.New(),
		Status:    "draft",
		Intent:    "order",
		DateTime:  time.Now(),
		UpdatedAt: time.Now(),
	}

	fhirMap := order.ToFHIR()

	absentKeys := []string{
		"encounter", "orderer", "oralDiet", "supplement",
		"enteralFormula", "note", "foodPreferenceModifier",
		"excludeFoodModifier", "allergyIntolerance",
	}
	for _, key := range absentKeys {
		if _, ok := fhirMap[key]; ok {
			t.Errorf("expected key %q to be absent", key)
		}
	}
}

func TestNutritionOrder_FromFHIR(t *testing.T) {
	patientID := uuid.New()
	encID := uuid.New()
	ordererID := uuid.New()

	fhirJSON := `{
		"resourceType": "NutritionOrder",
		"status": "active",
		"intent": "order",
		"dateTime": "2025-01-15T10:00:00Z",
		"patient": {"reference": "Patient/` + patientID.String() + `"},
		"encounter": {"reference": "Encounter/` + encID.String() + `"},
		"orderer": {"reference": "Practitioner/` + ordererID.String() + `"},
		"foodPreferenceModifier": [{"text": "kosher"}],
		"excludeFoodModifier": [{"text": "shellfish"}],
		"allergyIntolerance": [{"reference": "AllergyIntolerance/ai-001"}],
		"note": [{"text": "Some note"}],
		"oralDiet": {
			"type": [{"coding": [{"system": "http://snomed.info/sct", "code": "160674003", "display": "Low sodium diet"}]}],
			"nutrient": [{"modifier": {"coding": [{"code": "39972003"}]}, "amount": {"value": 2, "unit": "g"}}],
			"texture": [{"modifier": {"coding": [{"code": "228049004"}]}, "foodType": {"coding": [{"code": "28647000"}]}}],
			"fluidConsistencyType": [{"coding": [{"code": "439021000124105"}]}],
			"instruction": "Low sodium diet"
		},
		"supplement": [{
			"type": {"coding": [{"code": "443051000124104", "display": "Adult diabetic formula"}]},
			"productName": "Glucerna",
			"schedule": {"repeat": {"frequency": 2}},
			"quantity": {"value": 240, "unit": "mL"},
			"instruction": "Serve chilled"
		}],
		"enteralFormula": {
			"baseFormulaType": {"coding": [{"code": "443051000124104"}]},
			"additiveType": {"coding": [{"code": "lipid"}]},
			"caloricDensity": {"value": 1, "unit": "cal/mL"},
			"routeofAdministration": {"coding": [{"code": "GT"}]},
			"administration": [{"schedule": {"repeat": {"frequency": 1}}, "rateQuantity": {"value": 60, "unit": "mL/hr"}}]
		}
	}`

	order, err := NutritionOrderFromFHIR([]byte(fhirJSON))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.Status != "active" {
		t.Errorf("status = %v, want active", order.Status)
	}
	if order.Intent != "order" {
		t.Errorf("intent = %v, want order", order.Intent)
	}
	if order.PatientID != patientID {
		t.Errorf("patient_id = %v, want %v", order.PatientID, patientID)
	}
	if order.EncounterID == nil || *order.EncounterID != encID {
		t.Errorf("encounter_id = %v, want %v", order.EncounterID, encID)
	}
	if order.Orderer == nil || *order.Orderer != ordererID {
		t.Errorf("orderer = %v, want %v", order.Orderer, ordererID)
	}
	if order.Note != "Some note" {
		t.Errorf("note = %v, want Some note", order.Note)
	}
	if len(order.FoodPreferenceModifiers) != 1 || order.FoodPreferenceModifiers[0] != "kosher" {
		t.Errorf("foodPreferenceModifiers = %v", order.FoodPreferenceModifiers)
	}
	if len(order.ExcludeFoodModifiers) != 1 || order.ExcludeFoodModifiers[0] != "shellfish" {
		t.Errorf("excludeFoodModifiers = %v", order.ExcludeFoodModifiers)
	}
	if len(order.AllergyIntolerances) != 1 {
		t.Errorf("allergyIntolerances = %v", order.AllergyIntolerances)
	}

	// OralDiet
	if order.OralDiet == nil {
		t.Fatal("expected oral diet")
	}
	if len(order.OralDiet.Type) != 1 {
		t.Errorf("oral diet type count = %d, want 1", len(order.OralDiet.Type))
	}
	if len(order.OralDiet.Nutrients) != 1 {
		t.Errorf("nutrient count = %d, want 1", len(order.OralDiet.Nutrients))
	}

	// Supplement
	if len(order.Supplement) != 1 {
		t.Fatalf("supplement count = %d, want 1", len(order.Supplement))
	}
	if order.Supplement[0].ProductName != "Glucerna" {
		t.Errorf("supplement product = %v", order.Supplement[0].ProductName)
	}

	// EnteralFormula
	if order.EnteralFormula == nil {
		t.Fatal("expected enteral formula")
	}
	if order.EnteralFormula.BaseFormulaType.Code != "443051000124104" {
		t.Errorf("base formula code = %v", order.EnteralFormula.BaseFormulaType.Code)
	}
}

// ── Handler Tests ──

func TestNutritionOrderHandler_Create(t *testing.T) {
	h, e := newTestNutritionOrderHandler()
	body := `{
		"patient_id":"` + uuid.New().String() + `",
		"intent":"order",
		"date_time":"2025-01-15T10:00:00Z"
	}`
	req := httptest.NewRequest(http.MethodPost, "/nutrition-orders", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateNutritionOrder(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d; body: %s", rec.Code, rec.Body.String())
	}
}

func TestNutritionOrderHandler_Create_BadRequest(t *testing.T) {
	h, e := newTestNutritionOrderHandler()
	body := `{"intent":"order"}`
	req := httptest.NewRequest(http.MethodPost, "/nutrition-orders", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateNutritionOrder(c)
	if err == nil {
		t.Error("expected error for missing patient_id")
	}
}

func TestNutritionOrderHandler_Get(t *testing.T) {
	h, e := newTestNutritionOrderHandler()
	order := makeNutritionOrder(uuid.New())
	h.svc.Create(context.Background(), order)

	req := httptest.NewRequest(http.MethodGet, "/nutrition-orders/"+order.ID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(order.ID.String())
	err := h.GetNutritionOrder(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNutritionOrderHandler_Get_NotFound(t *testing.T) {
	h, e := newTestNutritionOrderHandler()
	req := httptest.NewRequest(http.MethodGet, "/nutrition-orders/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(uuid.New().String())
	err := h.GetNutritionOrder(c)
	if err == nil {
		t.Error("expected not-found error")
	}
}

func TestNutritionOrderHandler_Update(t *testing.T) {
	h, e := newTestNutritionOrderHandler()
	order := makeNutritionOrder(uuid.New())
	h.svc.Create(context.Background(), order)

	body := `{"status":"active","intent":"order","note":"updated"}`
	req := httptest.NewRequest(http.MethodPut, "/nutrition-orders/"+order.ID.String(), strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(order.ID.String())
	err := h.UpdateNutritionOrder(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	// Verify the note was updated
	var resp NutritionOrder
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Note != "updated" {
		t.Errorf("note = %v, want updated", resp.Note)
	}
}

func TestNutritionOrderHandler_Delete(t *testing.T) {
	h, e := newTestNutritionOrderHandler()
	order := makeNutritionOrder(uuid.New())
	h.svc.Create(context.Background(), order)

	req := httptest.NewRequest(http.MethodDelete, "/nutrition-orders/"+order.ID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(order.ID.String())
	err := h.DeleteNutritionOrder(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestNutritionOrderHandler_ListByPatient(t *testing.T) {
	h, e := newTestNutritionOrderHandler()
	ctx := context.Background()
	patientID := uuid.New()

	h.svc.Create(ctx, makeNutritionOrder(patientID))
	h.svc.Create(ctx, makeNutritionOrder(patientID))

	req := httptest.NewRequest(http.MethodGet, "/nutrition-orders?patient_id="+patientID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListNutritionOrders(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	total, ok := resp["total"].(float64)
	if !ok || int(total) != 2 {
		t.Errorf("total = %v, want 2", resp["total"])
	}
}

func TestNutritionOrderHandler_ListByEncounter(t *testing.T) {
	h, e := newTestNutritionOrderHandler()
	ctx := context.Background()
	encID := uuid.New()

	o1 := makeNutritionOrder(uuid.New())
	o1.EncounterID = &encID
	h.svc.Create(ctx, o1)

	req := httptest.NewRequest(http.MethodGet, "/nutrition-orders?encounter_id="+encID.String(), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListNutritionOrders(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestNutritionOrderHandler_List_NoQuery(t *testing.T) {
	h, e := newTestNutritionOrderHandler()
	req := httptest.NewRequest(http.MethodGet, "/nutrition-orders", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.ListNutritionOrders(c)
	if err == nil {
		t.Error("expected error when no patient_id or encounter_id provided")
	}
}

func TestNutritionOrderHandler_InvalidJSON(t *testing.T) {
	h, e := newTestNutritionOrderHandler()
	req := httptest.NewRequest(http.MethodPost, "/nutrition-orders", strings.NewReader("{invalid"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.CreateNutritionOrder(c)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestNutritionOrderHandler_Get_InvalidUUID(t *testing.T) {
	h, e := newTestNutritionOrderHandler()
	req := httptest.NewRequest(http.MethodGet, "/nutrition-orders/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("not-a-uuid")
	err := h.GetNutritionOrder(c)
	if err == nil {
		t.Error("expected error for invalid UUID")
	}
}
