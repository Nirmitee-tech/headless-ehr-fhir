package clinical

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/ehr/ehr/internal/platform/auth"
	"github.com/ehr/ehr/internal/platform/db"
	"github.com/ehr/ehr/internal/platform/fhir"
	"github.com/ehr/ehr/pkg/pagination"
)

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

// CodeableConcept is a simplified code+display structure for NutritionOrder sub-types.
type CodeableConcept struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}

// NutrientComponent represents a single nutrient constraint within an oral diet.
type NutrientComponent struct {
	Modifier CodeableConcept `json:"modifier"`
	Amount   string          `json:"amount,omitempty"`
}

// TextureModifier describes a food texture modification.
type TextureModifier struct {
	Modifier CodeableConcept `json:"modifier"`
	FoodType CodeableConcept `json:"food_type"`
}

// OralDiet describes the oral diet order details.
type OralDiet struct {
	Type             []CodeableConcept  `json:"type,omitempty"`
	Schedule         string             `json:"schedule,omitempty"`
	Nutrients        []NutrientComponent `json:"nutrients,omitempty"`
	TextureModifiers []TextureModifier  `json:"texture_modifiers,omitempty"`
	FluidConsistency []CodeableConcept  `json:"fluid_consistency,omitempty"`
	Instruction      string             `json:"instruction,omitempty"`
}

// Supplement represents a nutritional supplement order.
type Supplement struct {
	Type        CodeableConcept `json:"type"`
	ProductName string          `json:"product_name,omitempty"`
	Schedule    string          `json:"schedule,omitempty"`
	Quantity    string          `json:"quantity,omitempty"`
	Instruction string          `json:"instruction,omitempty"`
}

// EnteralAdministration describes a single enteral formula administration schedule.
type EnteralAdministration struct {
	Schedule string `json:"schedule,omitempty"`
	Rate     string `json:"rate,omitempty"`
	Quantity string `json:"quantity,omitempty"`
}

// EnteralFormula describes an enteral (tube-feeding) formula order.
type EnteralFormula struct {
	BaseFormulaType       CodeableConcept         `json:"base_formula_type"`
	AdditiveType          CodeableConcept         `json:"additive_type"`
	CaloricDensity        string                  `json:"caloric_density,omitempty"`
	RouteOfAdministration CodeableConcept         `json:"route_of_administration"`
	Administration        []EnteralAdministration `json:"administration,omitempty"`
}

// NutritionOrder is the domain model for a FHIR NutritionOrder resource.
type NutritionOrder struct {
	ID                      uuid.UUID        `db:"id" json:"id"`
	FHIRID                  string           `db:"fhir_id" json:"fhir_id"`
	PatientID               uuid.UUID        `db:"patient_id" json:"patient_id"`
	EncounterID             *uuid.UUID       `db:"encounter_id" json:"encounter_id,omitempty"`
	Orderer                 *uuid.UUID       `db:"orderer_id" json:"orderer,omitempty"`
	Status                  string           `db:"status" json:"status"`
	Intent                  string           `db:"intent" json:"intent"`
	DateTime                time.Time        `db:"date_time" json:"date_time"`
	OralDiet                *OralDiet        `json:"oral_diet,omitempty"`
	Supplement              []Supplement     `json:"supplement,omitempty"`
	EnteralFormula          *EnteralFormula  `json:"enteral_formula,omitempty"`
	AllergyIntolerances     []string         `json:"allergy_intolerances,omitempty"`
	FoodPreferenceModifiers []string         `json:"food_preference_modifiers,omitempty"`
	ExcludeFoodModifiers    []string         `json:"exclude_food_modifiers,omitempty"`
	Note                    string           `db:"note" json:"note,omitempty"`
	VersionID               int              `db:"version_id" json:"version_id"`
	CreatedAt               time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time        `db:"updated_at" json:"updated_at"`
}

func (n *NutritionOrder) GetVersionID() int  { return n.VersionID }
func (n *NutritionOrder) SetVersionID(v int) { n.VersionID = v }

// ---------------------------------------------------------------------------
// FHIR Mapping — toFHIR
// ---------------------------------------------------------------------------

func codeableConceptToFHIR(cc CodeableConcept) map[string]interface{} {
	result := map[string]interface{}{}
	coding := map[string]interface{}{}
	if cc.System != "" {
		coding["system"] = cc.System
	}
	if cc.Code != "" {
		coding["code"] = cc.Code
	}
	if cc.Display != "" {
		coding["display"] = cc.Display
	}
	result["coding"] = []map[string]interface{}{coding}
	if cc.Display != "" {
		result["text"] = cc.Display
	}
	return result
}

func (n *NutritionOrder) ToFHIR() map[string]interface{} {
	fhirID := n.FHIRID
	if fhirID == "" {
		fhirID = n.ID.String()
	}
	result := map[string]interface{}{
		"resourceType": "NutritionOrder",
		"id":           fhirID,
		"status":       n.Status,
		"intent":       n.Intent,
		"dateTime":     n.DateTime.Format(time.RFC3339),
		"patient":      map[string]string{"reference": "Patient/" + n.PatientID.String()},
		"meta": map[string]interface{}{
			"lastUpdated": n.UpdatedAt.Format(time.RFC3339),
			"versionId":   fmt.Sprintf("%d", n.VersionID),
		},
	}

	if n.EncounterID != nil {
		result["encounter"] = map[string]string{"reference": "Encounter/" + n.EncounterID.String()}
	}
	if n.Orderer != nil {
		result["orderer"] = map[string]string{"reference": "Practitioner/" + n.Orderer.String()}
	}

	// OralDiet
	if n.OralDiet != nil {
		od := map[string]interface{}{}
		if len(n.OralDiet.Type) > 0 {
			types := make([]map[string]interface{}, len(n.OralDiet.Type))
			for i, t := range n.OralDiet.Type {
				types[i] = codeableConceptToFHIR(t)
			}
			od["type"] = types
		}
		if n.OralDiet.Schedule != "" {
			od["schedule"] = map[string]interface{}{"repeat": map[string]interface{}{"description": n.OralDiet.Schedule}}
		}
		if len(n.OralDiet.Nutrients) > 0 {
			nutrients := make([]map[string]interface{}, len(n.OralDiet.Nutrients))
			for i, nu := range n.OralDiet.Nutrients {
				nutrients[i] = map[string]interface{}{
					"modifier": codeableConceptToFHIR(nu.Modifier),
				}
				if nu.Amount != "" {
					nutrients[i]["amount"] = map[string]string{"value": nu.Amount}
				}
			}
			od["nutrient"] = nutrients
		}
		if len(n.OralDiet.TextureModifiers) > 0 {
			textures := make([]map[string]interface{}, len(n.OralDiet.TextureModifiers))
			for i, tm := range n.OralDiet.TextureModifiers {
				textures[i] = map[string]interface{}{
					"modifier": codeableConceptToFHIR(tm.Modifier),
					"foodType": codeableConceptToFHIR(tm.FoodType),
				}
			}
			od["texture"] = textures
		}
		if len(n.OralDiet.FluidConsistency) > 0 {
			fluids := make([]map[string]interface{}, len(n.OralDiet.FluidConsistency))
			for i, f := range n.OralDiet.FluidConsistency {
				fluids[i] = codeableConceptToFHIR(f)
			}
			od["fluidConsistencyType"] = fluids
		}
		if n.OralDiet.Instruction != "" {
			od["instruction"] = n.OralDiet.Instruction
		}
		result["oralDiet"] = od
	}

	// Supplement
	if len(n.Supplement) > 0 {
		supps := make([]map[string]interface{}, len(n.Supplement))
		for i, s := range n.Supplement {
			supp := map[string]interface{}{
				"type": codeableConceptToFHIR(s.Type),
			}
			if s.ProductName != "" {
				supp["productName"] = s.ProductName
			}
			if s.Schedule != "" {
				supp["schedule"] = map[string]interface{}{"repeat": map[string]interface{}{"description": s.Schedule}}
			}
			if s.Quantity != "" {
				supp["quantity"] = map[string]string{"value": s.Quantity}
			}
			if s.Instruction != "" {
				supp["instruction"] = s.Instruction
			}
			supps[i] = supp
		}
		result["supplement"] = supps
	}

	// EnteralFormula
	if n.EnteralFormula != nil {
		ef := map[string]interface{}{
			"baseFormulaType":       codeableConceptToFHIR(n.EnteralFormula.BaseFormulaType),
			"additiveType":          codeableConceptToFHIR(n.EnteralFormula.AdditiveType),
			"routeofAdministration": codeableConceptToFHIR(n.EnteralFormula.RouteOfAdministration),
		}
		if n.EnteralFormula.CaloricDensity != "" {
			ef["caloricDensity"] = n.EnteralFormula.CaloricDensity
		}
		if len(n.EnteralFormula.Administration) > 0 {
			admins := make([]map[string]interface{}, len(n.EnteralFormula.Administration))
			for i, a := range n.EnteralFormula.Administration {
				admin := map[string]interface{}{}
				if a.Schedule != "" {
					admin["schedule"] = map[string]interface{}{"repeat": map[string]interface{}{"description": a.Schedule}}
				}
				if a.Rate != "" {
					admin["rateQuantity"] = map[string]string{"value": a.Rate}
				}
				if a.Quantity != "" {
					admin["quantity"] = map[string]string{"value": a.Quantity}
				}
				admins[i] = admin
			}
			ef["administration"] = admins
		}
		result["enteralFormula"] = ef
	}

	// AllergyIntolerance references
	if len(n.AllergyIntolerances) > 0 {
		refs := make([]map[string]string, len(n.AllergyIntolerances))
		for i, ref := range n.AllergyIntolerances {
			refs[i] = map[string]string{"reference": ref}
		}
		result["allergyIntolerance"] = refs
	}

	// FoodPreferenceModifier
	if len(n.FoodPreferenceModifiers) > 0 {
		mods := make([]map[string]interface{}, len(n.FoodPreferenceModifiers))
		for i, m := range n.FoodPreferenceModifiers {
			mods[i] = map[string]interface{}{"text": m}
		}
		result["foodPreferenceModifier"] = mods
	}

	// ExcludeFoodModifier
	if len(n.ExcludeFoodModifiers) > 0 {
		mods := make([]map[string]interface{}, len(n.ExcludeFoodModifiers))
		for i, m := range n.ExcludeFoodModifiers {
			mods[i] = map[string]interface{}{"text": m}
		}
		result["excludeFoodModifier"] = mods
	}

	// Note
	if n.Note != "" {
		result["note"] = []map[string]string{{"text": n.Note}}
	}

	return result
}

// ---------------------------------------------------------------------------
// FHIR Mapping — fromFHIR
// ---------------------------------------------------------------------------

// NutritionOrderFromFHIR parses a FHIR NutritionOrder JSON into the domain model.
func NutritionOrderFromFHIR(data []byte) (*NutritionOrder, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	order := &NutritionOrder{}

	// Status
	if v, ok := raw["status"]; ok {
		var s string
		json.Unmarshal(v, &s)
		order.Status = s
	}

	// Intent
	if v, ok := raw["intent"]; ok {
		var s string
		json.Unmarshal(v, &s)
		order.Intent = s
	}

	// DateTime
	if v, ok := raw["dateTime"]; ok {
		var s string
		json.Unmarshal(v, &s)
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			order.DateTime = t
		}
	}

	// Patient
	if v, ok := raw["patient"]; ok {
		order.PatientID = parseReferenceUUID(v, "Patient/")
	}

	// Encounter
	if v, ok := raw["encounter"]; ok {
		id := parseReferenceUUID(v, "Encounter/")
		if id != uuid.Nil {
			order.EncounterID = &id
		}
	}

	// Orderer
	if v, ok := raw["orderer"]; ok {
		id := parseReferenceUUID(v, "Practitioner/")
		if id != uuid.Nil {
			order.Orderer = &id
		}
	}

	// Note
	if v, ok := raw["note"]; ok {
		var notes []struct {
			Text string `json:"text"`
		}
		json.Unmarshal(v, &notes)
		if len(notes) > 0 {
			order.Note = notes[0].Text
		}
	}

	// FoodPreferenceModifier
	if v, ok := raw["foodPreferenceModifier"]; ok {
		var mods []struct {
			Text string `json:"text"`
		}
		json.Unmarshal(v, &mods)
		for _, m := range mods {
			if m.Text != "" {
				order.FoodPreferenceModifiers = append(order.FoodPreferenceModifiers, m.Text)
			}
		}
	}

	// ExcludeFoodModifier
	if v, ok := raw["excludeFoodModifier"]; ok {
		var mods []struct {
			Text string `json:"text"`
		}
		json.Unmarshal(v, &mods)
		for _, m := range mods {
			if m.Text != "" {
				order.ExcludeFoodModifiers = append(order.ExcludeFoodModifiers, m.Text)
			}
		}
	}

	// AllergyIntolerance
	if v, ok := raw["allergyIntolerance"]; ok {
		var refs []struct {
			Reference string `json:"reference"`
		}
		json.Unmarshal(v, &refs)
		for _, r := range refs {
			if r.Reference != "" {
				order.AllergyIntolerances = append(order.AllergyIntolerances, r.Reference)
			}
		}
	}

	// OralDiet
	if v, ok := raw["oralDiet"]; ok {
		order.OralDiet = parseOralDiet(v)
	}

	// Supplement
	if v, ok := raw["supplement"]; ok {
		order.Supplement = parseSupplements(v)
	}

	// EnteralFormula
	if v, ok := raw["enteralFormula"]; ok {
		order.EnteralFormula = parseEnteralFormula(v)
	}

	return order, nil
}

func parseReferenceUUID(data json.RawMessage, prefix string) uuid.UUID {
	var ref struct {
		Reference string `json:"reference"`
	}
	json.Unmarshal(data, &ref)
	if strings.HasPrefix(ref.Reference, prefix) {
		id, err := uuid.Parse(strings.TrimPrefix(ref.Reference, prefix))
		if err == nil {
			return id
		}
	}
	return uuid.Nil
}

func parseFHIRCodeableConcept(data json.RawMessage) CodeableConcept {
	var raw struct {
		Coding []struct {
			System  string `json:"system"`
			Code    string `json:"code"`
			Display string `json:"display"`
		} `json:"coding"`
		Text string `json:"text"`
	}
	json.Unmarshal(data, &raw)
	cc := CodeableConcept{}
	if len(raw.Coding) > 0 {
		cc.System = raw.Coding[0].System
		cc.Code = raw.Coding[0].Code
		cc.Display = raw.Coding[0].Display
	}
	return cc
}

func parseOralDiet(data json.RawMessage) *OralDiet {
	var raw map[string]json.RawMessage
	json.Unmarshal(data, &raw)

	od := &OralDiet{}

	// type
	if v, ok := raw["type"]; ok {
		var types []json.RawMessage
		json.Unmarshal(v, &types)
		for _, t := range types {
			od.Type = append(od.Type, parseFHIRCodeableConcept(t))
		}
	}

	// instruction
	if v, ok := raw["instruction"]; ok {
		var s string
		json.Unmarshal(v, &s)
		od.Instruction = s
	}

	// schedule
	if v, ok := raw["schedule"]; ok {
		var sched struct {
			Repeat struct {
				Description string `json:"description"`
			} `json:"repeat"`
		}
		json.Unmarshal(v, &sched)
		od.Schedule = sched.Repeat.Description
	}

	// nutrient
	if v, ok := raw["nutrient"]; ok {
		var nutrients []struct {
			Modifier json.RawMessage `json:"modifier"`
			Amount   json.RawMessage `json:"amount"`
		}
		json.Unmarshal(v, &nutrients)
		for _, n := range nutrients {
			nc := NutrientComponent{}
			if n.Modifier != nil {
				nc.Modifier = parseFHIRCodeableConcept(n.Modifier)
			}
			if n.Amount != nil {
				var amt struct {
					Value interface{} `json:"value"`
					Unit  string      `json:"unit"`
				}
				json.Unmarshal(n.Amount, &amt)
				if amt.Unit != "" {
					nc.Amount = fmt.Sprintf("%v %s", amt.Value, amt.Unit)
				} else {
					nc.Amount = fmt.Sprintf("%v", amt.Value)
				}
			}
			od.Nutrients = append(od.Nutrients, nc)
		}
	}

	// texture
	if v, ok := raw["texture"]; ok {
		var textures []struct {
			Modifier json.RawMessage `json:"modifier"`
			FoodType json.RawMessage `json:"foodType"`
		}
		json.Unmarshal(v, &textures)
		for _, t := range textures {
			tm := TextureModifier{}
			if t.Modifier != nil {
				tm.Modifier = parseFHIRCodeableConcept(t.Modifier)
			}
			if t.FoodType != nil {
				tm.FoodType = parseFHIRCodeableConcept(t.FoodType)
			}
			od.TextureModifiers = append(od.TextureModifiers, tm)
		}
	}

	// fluidConsistencyType
	if v, ok := raw["fluidConsistencyType"]; ok {
		var fluids []json.RawMessage
		json.Unmarshal(v, &fluids)
		for _, f := range fluids {
			od.FluidConsistency = append(od.FluidConsistency, parseFHIRCodeableConcept(f))
		}
	}

	return od
}

func parseSupplements(data json.RawMessage) []Supplement {
	var raw []struct {
		Type        json.RawMessage `json:"type"`
		ProductName string          `json:"productName"`
		Schedule    json.RawMessage `json:"schedule"`
		Quantity    json.RawMessage `json:"quantity"`
		Instruction string          `json:"instruction"`
	}
	json.Unmarshal(data, &raw)

	var supps []Supplement
	for _, r := range raw {
		s := Supplement{
			ProductName: r.ProductName,
			Instruction: r.Instruction,
		}
		if r.Type != nil {
			s.Type = parseFHIRCodeableConcept(r.Type)
		}
		if r.Schedule != nil {
			var sched struct {
				Repeat struct {
					Description string      `json:"description"`
					Frequency   interface{} `json:"frequency"`
				} `json:"repeat"`
			}
			json.Unmarshal(r.Schedule, &sched)
			if sched.Repeat.Description != "" {
				s.Schedule = sched.Repeat.Description
			} else if sched.Repeat.Frequency != nil {
				s.Schedule = fmt.Sprintf("%v", sched.Repeat.Frequency)
			}
		}
		if r.Quantity != nil {
			var qty struct {
				Value interface{} `json:"value"`
				Unit  string      `json:"unit"`
			}
			json.Unmarshal(r.Quantity, &qty)
			if qty.Unit != "" {
				s.Quantity = fmt.Sprintf("%v %s", qty.Value, qty.Unit)
			} else {
				s.Quantity = fmt.Sprintf("%v", qty.Value)
			}
		}
		supps = append(supps, s)
	}
	return supps
}

func parseEnteralFormula(data json.RawMessage) *EnteralFormula {
	var raw map[string]json.RawMessage
	json.Unmarshal(data, &raw)

	ef := &EnteralFormula{}

	if v, ok := raw["baseFormulaType"]; ok {
		ef.BaseFormulaType = parseFHIRCodeableConcept(v)
	}
	if v, ok := raw["additiveType"]; ok {
		ef.AdditiveType = parseFHIRCodeableConcept(v)
	}
	if v, ok := raw["caloricDensity"]; ok {
		var cd struct {
			Value interface{} `json:"value"`
			Unit  string      `json:"unit"`
		}
		json.Unmarshal(v, &cd)
		if cd.Unit != "" {
			ef.CaloricDensity = fmt.Sprintf("%v %s", cd.Value, cd.Unit)
		} else {
			// It might be a plain string
			var s string
			if json.Unmarshal(v, &s) == nil {
				ef.CaloricDensity = s
			}
		}
	}
	if v, ok := raw["routeofAdministration"]; ok {
		ef.RouteOfAdministration = parseFHIRCodeableConcept(v)
	}

	if v, ok := raw["administration"]; ok {
		var admins []struct {
			Schedule     json.RawMessage `json:"schedule"`
			RateQuantity json.RawMessage `json:"rateQuantity"`
			Quantity     json.RawMessage `json:"quantity"`
		}
		json.Unmarshal(v, &admins)
		for _, a := range admins {
			ea := EnteralAdministration{}
			if a.Schedule != nil {
				var sched struct {
					Repeat struct {
						Description string      `json:"description"`
						Frequency   interface{} `json:"frequency"`
					} `json:"repeat"`
				}
				json.Unmarshal(a.Schedule, &sched)
				if sched.Repeat.Description != "" {
					ea.Schedule = sched.Repeat.Description
				} else if sched.Repeat.Frequency != nil {
					ea.Schedule = fmt.Sprintf("%v", sched.Repeat.Frequency)
				}
			}
			if a.RateQuantity != nil {
				var rq struct {
					Value interface{} `json:"value"`
					Unit  string      `json:"unit"`
				}
				json.Unmarshal(a.RateQuantity, &rq)
				if rq.Unit != "" {
					ea.Rate = fmt.Sprintf("%v %s", rq.Value, rq.Unit)
				} else {
					ea.Rate = fmt.Sprintf("%v", rq.Value)
				}
			}
			if a.Quantity != nil {
				var qty struct {
					Value interface{} `json:"value"`
					Unit  string      `json:"unit"`
				}
				json.Unmarshal(a.Quantity, &qty)
				if qty.Unit != "" {
					ea.Quantity = fmt.Sprintf("%v %s", qty.Value, qty.Unit)
				} else {
					ea.Quantity = fmt.Sprintf("%v", qty.Value)
				}
			}
			ef.Administration = append(ef.Administration, ea)
		}
	}

	return ef
}

// ---------------------------------------------------------------------------
// Repository Interface
// ---------------------------------------------------------------------------

// NutritionOrderRepository defines the data-access contract for NutritionOrder.
type NutritionOrderRepository interface {
	Create(ctx context.Context, order *NutritionOrder) error
	GetByID(ctx context.Context, id uuid.UUID) (*NutritionOrder, error)
	GetByFHIRID(ctx context.Context, fhirID string) (*NutritionOrder, error)
	Update(ctx context.Context, order *NutritionOrder) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*NutritionOrder, int, error)
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*NutritionOrder, int, error)
	ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*NutritionOrder, int, error)
	Search(ctx context.Context, params map[string]string, limit, offset int) ([]*NutritionOrder, int, error)
}

// ---------------------------------------------------------------------------
// In-Memory Repository
// ---------------------------------------------------------------------------

type inMemoryNutritionOrderRepo struct {
	mu    sync.RWMutex
	store map[uuid.UUID]*NutritionOrder
	order []uuid.UUID // insertion order for deterministic pagination
}

// NewInMemoryNutritionOrderRepo creates a new in-memory NutritionOrder repository.
func NewInMemoryNutritionOrderRepo() NutritionOrderRepository {
	return &inMemoryNutritionOrderRepo{
		store: make(map[uuid.UUID]*NutritionOrder),
	}
}

func (r *inMemoryNutritionOrderRepo) Create(_ context.Context, order *NutritionOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order.ID = uuid.New()
	if order.FHIRID == "" {
		order.FHIRID = order.ID.String()
	}
	now := time.Now().UTC()
	order.CreatedAt = now
	order.UpdatedAt = now
	order.VersionID = 1

	// deep copy to avoid external mutation
	cp := *order
	r.store[cp.ID] = &cp
	r.order = append(r.order, cp.ID)
	return nil
}

func (r *inMemoryNutritionOrderRepo) GetByID(_ context.Context, id uuid.UUID) (*NutritionOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.store[id]
	if !ok {
		return nil, fmt.Errorf("nutrition order not found: %s", id)
	}
	cp := *o
	return &cp, nil
}

func (r *inMemoryNutritionOrderRepo) GetByFHIRID(_ context.Context, fhirID string) (*NutritionOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, o := range r.store {
		if o.FHIRID == fhirID {
			cp := *o
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("nutrition order not found: %s", fhirID)
}

func (r *inMemoryNutritionOrderRepo) List(_ context.Context, limit, offset int) ([]*NutritionOrder, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := len(r.order)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	var result []*NutritionOrder
	for _, id := range r.order[offset:end] {
		cp := *r.store[id]
		result = append(result, &cp)
	}
	return result, total, nil
}

func (r *inMemoryNutritionOrderRepo) Search(_ context.Context, params map[string]string, limit, offset int) ([]*NutritionOrder, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*NutritionOrder
	for _, id := range r.order {
		o := r.store[id]
		if p, ok := params["patient"]; ok && o.PatientID.String() != p {
			continue
		}
		if s, ok := params["status"]; ok && o.Status != s {
			continue
		}
		cp := *o
		all = append(all, &cp)
	}
	total := len(all)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (r *inMemoryNutritionOrderRepo) Update(_ context.Context, order *NutritionOrder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[order.ID]; !ok {
		return fmt.Errorf("nutrition order not found: %s", order.ID)
	}
	order.UpdatedAt = time.Now().UTC()
	cp := *order
	r.store[cp.ID] = &cp
	return nil
}

func (r *inMemoryNutritionOrderRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[id]; !ok {
		return fmt.Errorf("nutrition order not found: %s", id)
	}
	delete(r.store, id)
	for i, oid := range r.order {
		if oid == id {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
	return nil
}

func (r *inMemoryNutritionOrderRepo) ListByPatient(_ context.Context, patientID uuid.UUID, limit, offset int) ([]*NutritionOrder, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*NutritionOrder
	for _, id := range r.order {
		o := r.store[id]
		if o.PatientID == patientID {
			cp := *o
			all = append(all, &cp)
		}
	}
	total := len(all)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (r *inMemoryNutritionOrderRepo) ListByEncounter(_ context.Context, encounterID uuid.UUID, limit, offset int) ([]*NutritionOrder, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*NutritionOrder
	for _, id := range r.order {
		o := r.store[id]
		if o.EncounterID != nil && *o.EncounterID == encounterID {
			cp := *o
			all = append(all, &cp)
		}
	}
	total := len(all)
	if offset >= total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

// ---------------------------------------------------------------------------
// Service
// ---------------------------------------------------------------------------

var validNutritionOrderStatuses = map[string]bool{
	"draft": true, "active": true, "on-hold": true,
	"revoked": true, "completed": true, "entered-in-error": true,
}

var validNutritionOrderIntents = map[string]bool{
	"proposal": true, "plan": true, "directive": true, "order": true,
	"original-order": true, "reflex-order": true, "filler-order": true, "instance-order": true,
}

// allowedStatusTransitions defines which target statuses are reachable from each source status.
var allowedStatusTransitions = map[string]map[string]bool{
	"draft": {
		"active":           true,
		"on-hold":          true,
		"revoked":          true,
		"entered-in-error": true,
	},
	"active": {
		"on-hold":          true,
		"completed":        true,
		"revoked":          true,
		"entered-in-error": true,
	},
	"on-hold": {
		"active":           true,
		"revoked":          true,
		"entered-in-error": true,
	},
	"revoked":          {"entered-in-error": true},
	"completed":        {"entered-in-error": true},
	"entered-in-error": {},
}

// NutritionOrderService provides business logic for NutritionOrder.
type NutritionOrderService struct {
	repo NutritionOrderRepository
}

// NewNutritionOrderService creates a new NutritionOrderService.
func NewNutritionOrderService(repo NutritionOrderRepository) *NutritionOrderService {
	return &NutritionOrderService{repo: repo}
}

func (s *NutritionOrderService) Create(ctx context.Context, order *NutritionOrder) error {
	if order.PatientID == uuid.Nil {
		return fmt.Errorf("patient_id is required")
	}
	if order.Intent == "" {
		return fmt.Errorf("intent is required")
	}
	if !validNutritionOrderIntents[order.Intent] {
		return fmt.Errorf("invalid intent: %s", order.Intent)
	}
	if order.Status == "" {
		order.Status = "draft"
	}
	if !validNutritionOrderStatuses[order.Status] {
		return fmt.Errorf("invalid status: %s", order.Status)
	}
	if order.DateTime.IsZero() {
		order.DateTime = time.Now().UTC()
	}
	return s.repo.Create(ctx, order)
}

func (s *NutritionOrderService) GetByID(ctx context.Context, id uuid.UUID) (*NutritionOrder, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *NutritionOrderService) Update(ctx context.Context, order *NutritionOrder) error {
	existing, err := s.repo.GetByID(ctx, order.ID)
	if err != nil {
		return err
	}

	// If status is changing, validate the transition.
	if order.Status != "" && order.Status != existing.Status {
		if !validNutritionOrderStatuses[order.Status] {
			return fmt.Errorf("invalid status: %s", order.Status)
		}
		allowed := allowedStatusTransitions[existing.Status]
		if !allowed[order.Status] {
			return fmt.Errorf("invalid status transition from %s to %s", existing.Status, order.Status)
		}
	}
	if order.Intent != "" && !validNutritionOrderIntents[order.Intent] {
		return fmt.Errorf("invalid intent: %s", order.Intent)
	}

	// Merge: keep existing values for zero fields.
	if order.PatientID == uuid.Nil {
		order.PatientID = existing.PatientID
	}
	if order.Status == "" {
		order.Status = existing.Status
	}
	if order.Intent == "" {
		order.Intent = existing.Intent
	}
	if order.DateTime.IsZero() {
		order.DateTime = existing.DateTime
	}
	order.CreatedAt = existing.CreatedAt

	return s.repo.Update(ctx, order)
}

// UpdateStatus is a convenience method for status-only transitions.
func (s *NutritionOrderService) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !validNutritionOrderStatuses[newStatus] {
		return fmt.Errorf("invalid status: %s", newStatus)
	}
	allowed := allowedStatusTransitions[existing.Status]
	if !allowed[newStatus] {
		return fmt.Errorf("invalid status transition from %s to %s", existing.Status, newStatus)
	}
	existing.Status = newStatus
	return s.repo.Update(ctx, existing)
}

func (s *NutritionOrderService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *NutritionOrderService) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*NutritionOrder, int, error) {
	return s.repo.ListByPatient(ctx, patientID, limit, offset)
}

func (s *NutritionOrderService) ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*NutritionOrder, int, error) {
	return s.repo.ListByEncounter(ctx, encounterID, limit, offset)
}

func (s *NutritionOrderService) GetByFHIRID(ctx context.Context, fhirID string) (*NutritionOrder, error) {
	return s.repo.GetByFHIRID(ctx, fhirID)
}

func (s *NutritionOrderService) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*NutritionOrder, int, error) {
	return s.repo.Search(ctx, params, limit, offset)
}

// ---------------------------------------------------------------------------
// Postgres Repository
// ---------------------------------------------------------------------------

type nutritionOrderRepoPG struct{ pool *pgxpool.Pool }

// NewNutritionOrderRepoPG creates a Postgres-backed NutritionOrderRepository.
func NewNutritionOrderRepoPG(pool *pgxpool.Pool) NutritionOrderRepository {
	return &nutritionOrderRepoPG{pool: pool}
}

func (r *nutritionOrderRepoPG) conn(ctx context.Context) queryable {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	if c := db.ConnFromContext(ctx); c != nil {
		return c
	}
	return r.pool
}

const nutritionOrderCols = `id, fhir_id, status, intent, patient_id, encounter_id,
	orderer_id, date_time, oral_diet, supplement, enteral_formula,
	allergy_intolerances, food_preference_modifiers, exclude_food_modifiers,
	note, version_id, created_at, updated_at`

func (r *nutritionOrderRepoPG) scanOrder(row pgx.Row) (*NutritionOrder, error) {
	var n NutritionOrder
	var oralDietJSON, supplementJSON, enteralFormulaJSON []byte
	err := row.Scan(&n.ID, &n.FHIRID, &n.Status, &n.Intent, &n.PatientID, &n.EncounterID,
		&n.Orderer, &n.DateTime, &oralDietJSON, &supplementJSON, &enteralFormulaJSON,
		&n.AllergyIntolerances, &n.FoodPreferenceModifiers, &n.ExcludeFoodModifiers,
		&n.Note, &n.VersionID, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if oralDietJSON != nil {
		var od OralDiet
		if json.Unmarshal(oralDietJSON, &od) == nil {
			n.OralDiet = &od
		}
	}
	if supplementJSON != nil {
		json.Unmarshal(supplementJSON, &n.Supplement)
	}
	if enteralFormulaJSON != nil {
		var ef EnteralFormula
		if json.Unmarshal(enteralFormulaJSON, &ef) == nil {
			n.EnteralFormula = &ef
		}
	}
	return &n, nil
}

func (r *nutritionOrderRepoPG) Create(ctx context.Context, n *NutritionOrder) error {
	n.ID = uuid.New()
	if n.FHIRID == "" {
		n.FHIRID = n.ID.String()
	}
	oralDietJSON, _ := json.Marshal(n.OralDiet)
	supplementJSON, _ := json.Marshal(n.Supplement)
	enteralFormulaJSON, _ := json.Marshal(n.EnteralFormula)
	if n.OralDiet == nil {
		oralDietJSON = nil
	}
	if n.EnteralFormula == nil {
		enteralFormulaJSON = nil
	}
	if len(n.Supplement) == 0 {
		supplementJSON = nil
	}

	_, err := r.conn(ctx).Exec(ctx, `
		INSERT INTO nutrition_order (id, fhir_id, status, intent, patient_id, encounter_id,
			orderer_id, date_time, oral_diet, supplement, enteral_formula,
			allergy_intolerances, food_preference_modifiers, exclude_food_modifiers, note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		n.ID, n.FHIRID, n.Status, n.Intent, n.PatientID, n.EncounterID,
		n.Orderer, n.DateTime, oralDietJSON, supplementJSON, enteralFormulaJSON,
		n.AllergyIntolerances, n.FoodPreferenceModifiers, n.ExcludeFoodModifiers, n.Note)
	if err != nil {
		return err
	}
	n.VersionID = 1
	n.CreatedAt = time.Now().UTC()
	n.UpdatedAt = n.CreatedAt
	return nil
}

func (r *nutritionOrderRepoPG) GetByID(ctx context.Context, id uuid.UUID) (*NutritionOrder, error) {
	return r.scanOrder(r.conn(ctx).QueryRow(ctx, `SELECT `+nutritionOrderCols+` FROM nutrition_order WHERE id = $1`, id))
}

func (r *nutritionOrderRepoPG) GetByFHIRID(ctx context.Context, fhirID string) (*NutritionOrder, error) {
	return r.scanOrder(r.conn(ctx).QueryRow(ctx, `SELECT `+nutritionOrderCols+` FROM nutrition_order WHERE fhir_id = $1`, fhirID))
}

func (r *nutritionOrderRepoPG) Update(ctx context.Context, n *NutritionOrder) error {
	oralDietJSON, _ := json.Marshal(n.OralDiet)
	supplementJSON, _ := json.Marshal(n.Supplement)
	enteralFormulaJSON, _ := json.Marshal(n.EnteralFormula)
	if n.OralDiet == nil {
		oralDietJSON = nil
	}
	if n.EnteralFormula == nil {
		enteralFormulaJSON = nil
	}
	if len(n.Supplement) == 0 {
		supplementJSON = nil
	}

	_, err := r.conn(ctx).Exec(ctx, `
		UPDATE nutrition_order SET status=$2, intent=$3, patient_id=$4, encounter_id=$5,
			orderer_id=$6, date_time=$7, oral_diet=$8, supplement=$9, enteral_formula=$10,
			allergy_intolerances=$11, food_preference_modifiers=$12, exclude_food_modifiers=$13,
			note=$14, version_id=version_id+1, updated_at=NOW()
		WHERE id = $1`,
		n.ID, n.Status, n.Intent, n.PatientID, n.EncounterID,
		n.Orderer, n.DateTime, oralDietJSON, supplementJSON, enteralFormulaJSON,
		n.AllergyIntolerances, n.FoodPreferenceModifiers, n.ExcludeFoodModifiers, n.Note)
	return err
}

func (r *nutritionOrderRepoPG) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.conn(ctx).Exec(ctx, `DELETE FROM nutrition_order WHERE id = $1`, id)
	return err
}

func (r *nutritionOrderRepoPG) List(ctx context.Context, limit, offset int) ([]*NutritionOrder, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM nutrition_order`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+nutritionOrderCols+` FROM nutrition_order ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*NutritionOrder
	for rows.Next() {
		n, err := r.scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, total, nil
}

func (r *nutritionOrderRepoPG) ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*NutritionOrder, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM nutrition_order WHERE patient_id = $1`, patientID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+nutritionOrderCols+` FROM nutrition_order WHERE patient_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, patientID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*NutritionOrder
	for rows.Next() {
		n, err := r.scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, total, nil
}

func (r *nutritionOrderRepoPG) ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*NutritionOrder, int, error) {
	var total int
	if err := r.conn(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM nutrition_order WHERE encounter_id = $1`, encounterID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.conn(ctx).Query(ctx, `SELECT `+nutritionOrderCols+` FROM nutrition_order WHERE encounter_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, encounterID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*NutritionOrder
	for rows.Next() {
		n, err := r.scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, total, nil
}

func (r *nutritionOrderRepoPG) Search(ctx context.Context, params map[string]string, limit, offset int) ([]*NutritionOrder, int, error) {
	query := `SELECT ` + nutritionOrderCols + ` FROM nutrition_order WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM nutrition_order WHERE 1=1`
	var args []interface{}
	idx := 1

	if p, ok := params["patient"]; ok {
		query += fmt.Sprintf(` AND patient_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND patient_id = $%d`, idx)
		args = append(args, p)
		idx++
	}
	if s, ok := params["status"]; ok {
		query += fmt.Sprintf(` AND status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND status = $%d`, idx)
		args = append(args, s)
		idx++
	}

	var total int
	if err := r.conn(ctx).QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.conn(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []*NutritionOrder
	for rows.Next() {
		n, err := r.scanOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, n)
	}
	return items, total, nil
}

// ---------------------------------------------------------------------------
// Handler
// ---------------------------------------------------------------------------

// NutritionOrderHandler handles HTTP requests for NutritionOrder resources.
type NutritionOrderHandler struct {
	svc *NutritionOrderService
}

// NewNutritionOrderHandler creates a new NutritionOrderHandler.
func NewNutritionOrderHandler(svc *NutritionOrderService) *NutritionOrderHandler {
	return &NutritionOrderHandler{svc: svc}
}

// RegisterNutritionOrderRoutes registers the NutritionOrder REST and FHIR routes.
func (h *NutritionOrderHandler) RegisterNutritionOrderRoutes(api *echo.Group, fhirGroup *echo.Group) {
	// REST endpoints
	read := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	read.GET("/nutrition-orders", h.ListNutritionOrders)
	read.GET("/nutrition-orders/:id", h.GetNutritionOrder)

	write := api.Group("", auth.RequireRole("admin", "physician", "nurse"))
	write.POST("/nutrition-orders", h.CreateNutritionOrder)
	write.PUT("/nutrition-orders/:id", h.UpdateNutritionOrder)
	write.DELETE("/nutrition-orders/:id", h.DeleteNutritionOrder)

	// FHIR read endpoints
	fr := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse"))
	fr.GET("/NutritionOrder", h.SearchNutritionOrdersFHIR)
	fr.GET("/NutritionOrder/:id", h.GetNutritionOrderFHIR)
	fr.POST("/NutritionOrder/_search", h.SearchNutritionOrdersFHIR)
	fr.GET("/NutritionOrder/:id/_history/:vid", h.VreadNutritionOrderFHIR)
	fr.GET("/NutritionOrder/:id/_history", h.HistoryNutritionOrderFHIR)

	// FHIR write endpoints
	fw := fhirGroup.Group("", auth.RequireRole("admin", "physician", "nurse"))
	fw.POST("/NutritionOrder", h.CreateNutritionOrderFHIR)
	fw.PUT("/NutritionOrder/:id", h.UpdateNutritionOrderFHIR)
	fw.DELETE("/NutritionOrder/:id", h.DeleteNutritionOrderFHIR)
	fw.PATCH("/NutritionOrder/:id", h.PatchNutritionOrderFHIR)
}

func (h *NutritionOrderHandler) CreateNutritionOrder(c echo.Context) error {
	var order NutritionOrder
	if err := c.Bind(&order); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body: "+err.Error())
	}
	if err := h.svc.Create(c.Request().Context(), &order); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, order)
}

func (h *NutritionOrderHandler) GetNutritionOrder(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	order, err := h.svc.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, order)
}

func (h *NutritionOrderHandler) UpdateNutritionOrder(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var order NutritionOrder
	if err := c.Bind(&order); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body: "+err.Error())
	}
	order.ID = id
	if err := h.svc.Update(c.Request().Context(), &order); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, order)
}

func (h *NutritionOrderHandler) DeleteNutritionOrder(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *NutritionOrderHandler) ListNutritionOrders(c echo.Context) error {
	p := pagination.FromContext(c)
	ctx := c.Request().Context()

	patientIDStr := c.QueryParam("patient_id")
	encounterIDStr := c.QueryParam("encounter_id")

	if patientIDStr != "" {
		patientID, err := uuid.Parse(patientIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid patient_id")
		}
		orders, total, err := h.svc.ListByPatient(ctx, patientID, p.Limit, p.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(orders, total, p.Limit, p.Offset))
	}

	if encounterIDStr != "" {
		encounterID, err := uuid.Parse(encounterIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid encounter_id")
		}
		orders, total, err := h.svc.ListByEncounter(ctx, encounterID, p.Limit, p.Offset)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, pagination.NewResponse(orders, total, p.Limit, p.Offset))
	}

	return echo.NewHTTPError(http.StatusBadRequest, "patient_id or encounter_id query parameter is required")
}

// ---------------------------------------------------------------------------
// FHIR Handlers
// ---------------------------------------------------------------------------

func (h *NutritionOrderHandler) SearchNutritionOrdersFHIR(c echo.Context) error {
	pg := pagination.FromContext(c)
	params := map[string]string{}
	for _, k := range []string{"patient", "status"} {
		if v := c.QueryParam(k); v != "" {
			params[k] = v
		}
	}
	items, total, err := h.svc.Search(c.Request().Context(), params, pg.Limit, pg.Offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	resources := make([]interface{}, len(items))
	for i, item := range items {
		resources[i] = item.ToFHIR()
	}
	return c.JSON(http.StatusOK, fhir.NewSearchBundle(resources, total, "/fhir/NutritionOrder"))
}

func (h *NutritionOrderHandler) GetNutritionOrderFHIR(c echo.Context) error {
	n, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NutritionOrder", c.Param("id")))
	}
	return c.JSON(http.StatusOK, n.ToFHIR())
}

func (h *NutritionOrderHandler) CreateNutritionOrderFHIR(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	order, err := NutritionOrderFromFHIR(body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	if err := h.svc.Create(c.Request().Context(), order); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	c.Response().Header().Set("Location", "/fhir/NutritionOrder/"+order.FHIRID)
	return c.JSON(http.StatusCreated, order.ToFHIR())
}

func (h *NutritionOrderHandler) UpdateNutritionOrderFHIR(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}
	order, err := NutritionOrderFromFHIR(body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	existing, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NutritionOrder", c.Param("id")))
	}
	order.ID = existing.ID
	order.FHIRID = existing.FHIRID
	if err := h.svc.Update(c.Request().Context(), order); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, order.ToFHIR())
}

func (h *NutritionOrderHandler) DeleteNutritionOrderFHIR(c echo.Context) error {
	existing, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NutritionOrder", c.Param("id")))
	}
	if err := h.svc.Delete(c.Request().Context(), existing.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, fhir.ErrorOutcome(err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *NutritionOrderHandler) PatchNutritionOrderFHIR(c echo.Context) error {
	contentType := c.Request().Header.Get("Content-Type")
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("failed to read request body"))
	}

	existing, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NutritionOrder", c.Param("id")))
	}
	currentResource := existing.ToFHIR()

	var patched map[string]interface{}
	if strings.Contains(contentType, "json-patch+json") {
		ops, err := fhir.ParseJSONPatch(body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
		}
		patched, err = fhir.ApplyJSONPatch(currentResource, ops)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else if strings.Contains(contentType, "merge-patch+json") {
		var mergePatch map[string]interface{}
		if err := json.Unmarshal(body, &mergePatch); err != nil {
			return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome("invalid merge patch JSON: "+err.Error()))
		}
		patched, err = fhir.ApplyMergePatch(currentResource, mergePatch)
		if err != nil {
			return c.JSON(http.StatusUnprocessableEntity, fhir.ErrorOutcome(err.Error()))
		}
	} else {
		return c.JSON(http.StatusUnsupportedMediaType, fhir.ErrorOutcome(
			"PATCH requires Content-Type: application/json-patch+json or application/merge-patch+json"))
	}

	// Apply patched fields back to model
	if v, ok := patched["status"].(string); ok {
		existing.Status = v
	}
	if v, ok := patched["intent"].(string); ok {
		existing.Intent = v
	}
	if err := h.svc.Update(c.Request().Context(), existing); err != nil {
		return c.JSON(http.StatusBadRequest, fhir.ErrorOutcome(err.Error()))
	}
	return c.JSON(http.StatusOK, existing.ToFHIR())
}

func (h *NutritionOrderHandler) VreadNutritionOrderFHIR(c echo.Context) error {
	n, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NutritionOrder", c.Param("id")))
	}
	fhir.SetVersionHeaders(c, 1, n.UpdatedAt.Format("2006-01-02T15:04:05Z"))
	return c.JSON(http.StatusOK, n.ToFHIR())
}

func (h *NutritionOrderHandler) HistoryNutritionOrderFHIR(c echo.Context) error {
	n, err := h.svc.GetByFHIRID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, fhir.NotFoundOutcome("NutritionOrder", c.Param("id")))
	}
	raw, _ := json.Marshal(n.ToFHIR())
	entry := &fhir.HistoryEntry{
		ResourceType: "NutritionOrder", ResourceID: n.FHIRID, VersionID: 1,
		Resource: raw, Action: "create", Timestamp: n.CreatedAt,
	}
	return c.JSON(http.StatusOK, fhir.NewHistoryBundle([]*fhir.HistoryEntry{entry}, 1, "/fhir"))
}
