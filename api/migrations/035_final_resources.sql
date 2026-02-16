-- Migration 035: Final FHIR R4 resources — ImmunizationEvaluation
-- AuditEvent table already exists in migration 001

CREATE TABLE immunization_evaluation (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id                     VARCHAR(64) UNIQUE NOT NULL,
    status                      VARCHAR(20) NOT NULL DEFAULT 'completed',
    patient_id                  UUID NOT NULL REFERENCES patient(id),
    date                        TIMESTAMPTZ,
    authority_reference         VARCHAR(255),
    target_disease_code         VARCHAR(50) NOT NULL,
    target_disease_display      VARCHAR(255),
    immunization_event_reference VARCHAR(255) NOT NULL,
    dose_status_code            VARCHAR(50) NOT NULL,
    dose_status_display         VARCHAR(255),
    dose_status_reason_code     VARCHAR(50),
    dose_status_reason_display  VARCHAR(255),
    series                      VARCHAR(255),
    dose_number                 VARCHAR(20),
    series_doses                VARCHAR(20),
    description                 TEXT,
    version_id                  INTEGER NOT NULL DEFAULT 1,
    created_at                  TIMESTAMPTZ DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_immunization_evaluation_fhir_id ON immunization_evaluation(fhir_id);
CREATE INDEX idx_immunization_evaluation_patient ON immunization_evaluation(patient_id);
CREATE INDEX idx_immunization_evaluation_status ON immunization_evaluation(status);
