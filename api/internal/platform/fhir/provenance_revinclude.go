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
	// The provenance table stores target_type and target_id separately.
	query := `
		SELECT id, fhir_id, target_type, target_id, recorded,
			activity_code, activity_display, reason_code, reason_display,
			created_at, updated_at
		FROM provenance
		WHERE (target_type || '/' || target_id) = ANY($1)
		ORDER BY recorded DESC
	`

	rows, err := p.pool.Query(ctx, query, targetRefs)
	if err != nil {
		return nil, fmt.Errorf("provenance revinclude query: %w", err)
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var (
			id, fhirID, targetType, targetID string
			recorded                          time.Time
			activityCode, activityDisplay     *string
			reasonCode, reasonDisplay         *string
			createdAt, updatedAt              time.Time
		)
		if err := rows.Scan(&id, &fhirID, &targetType, &targetID,
			&recorded, &activityCode, &activityDisplay,
			&reasonCode, &reasonDisplay,
			&createdAt, &updatedAt); err != nil {
			continue
		}

		provenance := map[string]interface{}{
			"resourceType": "Provenance",
			"id":           fhirID,
			"target": []Reference{
				{Reference: FormatReference(targetType, targetID)},
			},
			"recorded": recorded.Format(time.RFC3339),
			"meta": Meta{
				LastUpdated: updatedAt,
				Profile:     []string{"http://hl7.org/fhir/StructureDefinition/Provenance"},
			},
		}
		if activityCode != nil {
			provenance["activity"] = CodeableConcept{
				Coding: []Coding{{Code: *activityCode, Display: ptrStrVal(activityDisplay)}},
			}
		}
		if reasonCode != nil {
			provenance["reason"] = []CodeableConcept{{
				Coding: []Coding{{Code: *reasonCode, Display: ptrStrVal(reasonDisplay)}},
			}}
		}
		results = append(results, provenance)
	}

	return results, nil
}

// ptrStrVal safely dereferences a *string, returning "" if nil.
func ptrStrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
