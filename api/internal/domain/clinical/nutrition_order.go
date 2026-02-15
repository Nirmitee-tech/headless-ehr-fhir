package clinical

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

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
	ID                      uuid.UUID        `json:"id"`
	PatientID               uuid.UUID        `json:"patient_id"`
	EncounterID             *uuid.UUID       `json:"encounter_id,omitempty"`
	Orderer                 *uuid.UUID       `json:"orderer,omitempty"`
	Status                  string           `json:"status"`
	Intent                  string           `json:"intent"`
	DateTime                time.Time        `json:"date_time"`
	OralDiet                *OralDiet        `json:"oral_diet,omitempty"`
	Supplement              []Supplement     `json:"supplement,omitempty"`
	EnteralFormula          *EnteralFormula  `json:"enteral_formula,omitempty"`
	AllergyIntolerances     []string         `json:"allergy_intolerances,omitempty"`
	FoodPreferenceModifiers []string         `json:"food_preference_modifiers,omitempty"`
	ExcludeFoodModifiers    []string         `json:"exclude_food_modifiers,omitempty"`
	Note                    string           `json:"note,omitempty"`
	CreatedAt               time.Time        `json:"created_at"`
	UpdatedAt               time.Time        `json:"updated_at"`
}

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
	result := map[string]interface{}{
		"resourceType": "NutritionOrder",
		"id":           n.ID.String(),
		"status":       n.Status,
		"intent":       n.Intent,
		"dateTime":     n.DateTime.Format(time.RFC3339),
		"patient":      map[string]string{"reference": "Patient/" + n.PatientID.String()},
		"meta": map[string]interface{}{
			"lastUpdated": n.UpdatedAt.Format(time.RFC3339),
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
	Update(ctx context.Context, order *NutritionOrder) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPatient(ctx context.Context, patientID uuid.UUID, limit, offset int) ([]*NutritionOrder, int, error)
	ListByEncounter(ctx context.Context, encounterID uuid.UUID, limit, offset int) ([]*NutritionOrder, int, error)
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
	now := time.Now().UTC()
	order.CreatedAt = now
	order.UpdatedAt = now

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

// RegisterNutritionOrderRoutes registers the NutritionOrder routes on the given Echo group.
func (h *NutritionOrderHandler) RegisterNutritionOrderRoutes(api *echo.Group) {
	api.GET("/nutrition-orders", h.ListNutritionOrders)
	api.POST("/nutrition-orders", h.CreateNutritionOrder)
	api.GET("/nutrition-orders/:id", h.GetNutritionOrder)
	api.PUT("/nutrition-orders/:id", h.UpdateNutritionOrder)
	api.DELETE("/nutrition-orders/:id", h.DeleteNutritionOrder)
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
