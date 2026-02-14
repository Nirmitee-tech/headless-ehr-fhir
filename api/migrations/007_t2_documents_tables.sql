-- ============================================================================
-- T2 DOCUMENTS TABLES MIGRATION
-- Tables: consent, document_reference, clinical_note, composition,
--         composition_section
-- ============================================================================

-- ============================================================================
-- 1. CONSENT (FHIR Consent resource)
-- ============================================================================

CREATE TABLE consent (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(20) NOT NULL,
    scope               VARCHAR(50),
    category_code       VARCHAR(30),
    category_display    VARCHAR(255),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    performer_id        UUID REFERENCES practitioner(id),
    organization_id     UUID REFERENCES organization(id),
    policy_authority    VARCHAR(255),
    policy_uri          VARCHAR(500),
    provision_type      VARCHAR(20),
    provision_start     TIMESTAMPTZ,
    provision_end       TIMESTAMPTZ,
    provision_action    VARCHAR(30),
    hipaa_authorization BOOLEAN DEFAULT FALSE,
    abdm_consent        BOOLEAN DEFAULT FALSE,
    abdm_consent_id     VARCHAR(100),
    signature_type      VARCHAR(30),
    signature_when      TIMESTAMPTZ,
    signature_data      TEXT,
    date_time           TIMESTAMPTZ,
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. DOCUMENT REFERENCE (FHIR DocumentReference resource)
-- ============================================================================

CREATE TABLE document_reference (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(20) NOT NULL,
    doc_status          VARCHAR(20),
    type_code           VARCHAR(30),
    type_display        VARCHAR(255),
    category_code       VARCHAR(30),
    category_display    VARCHAR(255),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    author_id           UUID REFERENCES practitioner(id),
    custodian_id        UUID REFERENCES organization(id),
    encounter_id        UUID REFERENCES encounter(id),
    date                TIMESTAMPTZ,
    description         TEXT,
    security_label      VARCHAR(30),
    content_type        VARCHAR(100),
    content_url         VARCHAR(1000),
    content_size        INTEGER,
    content_hash        VARCHAR(128),
    content_title       VARCHAR(500),
    format_code         VARCHAR(50),
    format_display      VARCHAR(255),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. CLINICAL NOTE (structured clinical notes with SOAP)
-- ============================================================================

CREATE TABLE clinical_note (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    author_id           UUID NOT NULL REFERENCES practitioner(id),
    note_type           VARCHAR(30) NOT NULL,
    status              VARCHAR(20) NOT NULL,
    title               VARCHAR(500),
    subjective          TEXT,
    objective           TEXT,
    assessment          TEXT,
    plan                TEXT,
    note_text           TEXT,
    signed_by           UUID REFERENCES practitioner(id),
    signed_at           TIMESTAMPTZ,
    cosigned_by         UUID REFERENCES practitioner(id),
    cosigned_at         TIMESTAMPTZ,
    amended_by          UUID REFERENCES practitioner(id),
    amended_at          TIMESTAMPTZ,
    amended_reason      TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. COMPOSITION (FHIR Composition resource)
-- ============================================================================

CREATE TABLE composition (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(20) NOT NULL,
    type_code           VARCHAR(30),
    type_display        VARCHAR(255),
    category_code       VARCHAR(30),
    category_display    VARCHAR(255),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    encounter_id        UUID REFERENCES encounter(id),
    date                TIMESTAMPTZ,
    author_id           UUID REFERENCES practitioner(id),
    title               VARCHAR(500),
    confidentiality     VARCHAR(10),
    custodian_id        UUID REFERENCES organization(id),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 5. COMPOSITION SECTION (nested sections of compositions)
-- ============================================================================

CREATE TABLE composition_section (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    composition_id      UUID NOT NULL REFERENCES composition(id) ON DELETE CASCADE,
    title               VARCHAR(500),
    code_value          VARCHAR(30),
    code_display        VARCHAR(255),
    text_status         VARCHAR(20),
    text_div            TEXT,
    mode                VARCHAR(20),
    ordered_by          VARCHAR(30),
    entry_reference     VARCHAR(255),
    sort_order          INTEGER
);

-- ============================================================================
-- 6. INDEXES for T2 documents tables
-- ============================================================================

-- Consent indexes
CREATE INDEX idx_consent_fhir ON consent(fhir_id);
CREATE INDEX idx_consent_patient ON consent(patient_id);
CREATE INDEX idx_consent_status ON consent(status);
CREATE INDEX idx_consent_category ON consent(category_code) WHERE category_code IS NOT NULL;

-- DocumentReference indexes
CREATE INDEX idx_docref_fhir ON document_reference(fhir_id);
CREATE INDEX idx_docref_patient ON document_reference(patient_id);
CREATE INDEX idx_docref_status ON document_reference(status);
CREATE INDEX idx_docref_type ON document_reference(type_code) WHERE type_code IS NOT NULL;
CREATE INDEX idx_docref_encounter ON document_reference(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_docref_date ON document_reference(date DESC) WHERE date IS NOT NULL;

-- ClinicalNote indexes
CREATE INDEX idx_note_patient ON clinical_note(patient_id);
CREATE INDEX idx_note_encounter ON clinical_note(encounter_id) WHERE encounter_id IS NOT NULL;
CREATE INDEX idx_note_author ON clinical_note(author_id);
CREATE INDEX idx_note_status ON clinical_note(status);
CREATE INDEX idx_note_type ON clinical_note(note_type);

-- Composition indexes
CREATE INDEX idx_composition_fhir ON composition(fhir_id);
CREATE INDEX idx_composition_patient ON composition(patient_id);
CREATE INDEX idx_composition_status ON composition(status);
CREATE INDEX idx_composition_type ON composition(type_code) WHERE type_code IS NOT NULL;
CREATE INDEX idx_composition_date ON composition(date DESC) WHERE date IS NOT NULL;

-- CompositionSection indexes
CREATE INDEX idx_comp_section_composition ON composition_section(composition_id);
CREATE INDEX idx_comp_section_sort ON composition_section(composition_id, sort_order);
