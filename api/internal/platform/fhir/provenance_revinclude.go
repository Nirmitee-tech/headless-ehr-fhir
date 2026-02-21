package fhir

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ProvenanceRevIncludeProvider implements RevIncludeProvider by querying the provenance table.
type ProvenanceRevIncludeProvider struct {
	pool *pgxpool.Pool
}

// NewProvenanceRevIncludeProvider creates a new ProvenanceRevIncludeProvider.
func NewProvenanceRevIncludeProvider(pool *pgxpool.Pool) *ProvenanceRevIncludeProvider {
	return &ProvenanceRevIncludeProvider{pool: pool}
}

// FindByTargets returns Provenance resources whose target references match the given refs.
// targetRefs are strings like "Condition/abc-123".
func (p *ProvenanceRevIncludeProvider) FindByTargets(ctx context.Context, targetRefs []string) ([]interface{}, error) {
	if len(targetRefs) == 0 {
		return nil, nil
	}

	// Query provenance rows matching any of the target references.
	query := `
		SELECT p.id, p.fhir_id, p.target_type, p.target_id, p.recorded,
			p.activity_code, p.activity_display, p.reason_code, p.reason_display,
			p.created_at, p.updated_at
		FROM provenance p
		WHERE (p.target_type || '/' || p.target_id) = ANY($1)
		ORDER BY p.recorded DESC
	`

	rows, err := p.pool.Query(ctx, query, targetRefs)
	if err != nil {
		return nil, fmt.Errorf("provenance revinclude query: %w", err)
	}
	defer rows.Close()

	type provenanceRow struct {
		id, fhirID, targetType, targetID string
		recorded                          time.Time
		activityCode, activityDisplay     *string
		reasonCode, reasonDisplay         *string
		createdAt, updatedAt              time.Time
	}

	var provRows []provenanceRow
	var provIDs []string
	for rows.Next() {
		var r provenanceRow
		if err := rows.Scan(&r.id, &r.fhirID, &r.targetType, &r.targetID,
			&r.recorded, &r.activityCode, &r.activityDisplay,
			&r.reasonCode, &r.reasonDisplay,
			&r.createdAt, &r.updatedAt); err != nil {
			continue
		}
		provRows = append(provRows, r)
		provIDs = append(provIDs, r.id)
	}

	if len(provRows) == 0 {
		return nil, nil
	}

	// Load agents for all provenance rows.
	agents := p.loadAgents(ctx, provIDs)

	var results []interface{}
	for _, r := range provRows {
		provenance := map[string]interface{}{
			"resourceType": "Provenance",
			"id":           r.fhirID,
			"target": []Reference{
				{Reference: FormatReference(r.targetType, r.targetID)},
			},
			"recorded": r.recorded.Format(time.RFC3339),
			"meta": Meta{
				LastUpdated: r.updatedAt,
				Profile:     []string{"http://hl7.org/fhir/us/core/StructureDefinition/us-core-provenance"},
			},
		}
		if r.activityCode != nil {
			provenance["activity"] = CodeableConcept{
				Coding: []Coding{{Code: *r.activityCode, Display: ptrStrVal(r.activityDisplay)}},
			}
		}
		if r.reasonCode != nil {
			provenance["reason"] = []CodeableConcept{{
				Coding: []Coding{{Code: *r.reasonCode, Display: ptrStrVal(r.reasonDisplay)}},
			}}
		}
		if agentList, ok := agents[r.id]; ok && len(agentList) > 0 {
			provenance["agent"] = agentList
		}
		results = append(results, provenance)
	}

	return results, nil
}

// loadAgents fetches provenance agents grouped by provenance ID.
func (p *ProvenanceRevIncludeProvider) loadAgents(ctx context.Context, provenanceIDs []string) map[string][]map[string]interface{} {
	result := make(map[string][]map[string]interface{})
	if len(provenanceIDs) == 0 {
		return result
	}

	query := `
		SELECT provenance_id, type_code, type_display, who_type, who_id
		FROM provenance_agent
		WHERE provenance_id = ANY($1)
		ORDER BY provenance_id
	`
	rows, err := p.pool.Query(ctx, query, provenanceIDs)
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var provID, whoType, whoID string
		var typeCode, typeDisplay *string
		if err := rows.Scan(&provID, &typeCode, &typeDisplay, &whoType, &whoID); err != nil {
			continue
		}

		agent := map[string]interface{}{
			"who": Reference{Reference: FormatReference(whoType, whoID)},
		}
		if typeCode != nil {
			agent["type"] = []CodeableConcept{{
				Coding: []Coding{{
					System:  "http://terminology.hl7.org/CodeSystem/provenance-participant-type",
					Code:    *typeCode,
					Display: ptrStrVal(typeDisplay),
				}},
			}}
		}
		result[provID] = append(result[provID], agent)
	}
	return result
}

// ptrStrVal safely dereferences a *string, returning "" if nil.
func ptrStrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
