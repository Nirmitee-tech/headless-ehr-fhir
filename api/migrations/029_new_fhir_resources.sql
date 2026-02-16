-- 029_new_fhir_resources.sql
-- Adds 10 new FHIR R4 resource tables: CoverageEligibilityRequest,
-- CoverageEligibilityResponse, MedicationKnowledge, OrganizationAffiliation,
-- Person, Measure, Library, DeviceDefinition, DeviceMetric, SpecimenDefinition

-- ============================================================
-- Financial / Insurance Eligibility Tables
-- ============================================================

-- CoverageEligibilityRequest (insurance eligibility request)
CREATE TABLE IF NOT EXISTS coverage_eligibility_request (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    patient_id      UUID NOT NULL REFERENCES patient(id),
    provider_id     UUID REFERENCES practitioner(id),
    insurer_id      UUID REFERENCES organization(id),
    purpose         VARCHAR(20) NOT NULL DEFAULT 'benefits',
    serviced_date   DATE,
    created         DATE
);

CREATE INDEX IF NOT EXISTS idx_coverage_elig_req_patient ON coverage_eligibility_request (patient_id);
CREATE INDEX IF NOT EXISTS idx_coverage_elig_req_status ON coverage_eligibility_request (status);
CREATE INDEX IF NOT EXISTS idx_coverage_elig_req_insurer ON coverage_eligibility_request (insurer_id);

-- CoverageEligibilityResponse (insurance eligibility response)
CREATE TABLE IF NOT EXISTS coverage_eligibility_response (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    patient_id      UUID NOT NULL REFERENCES patient(id),
    request_id      UUID REFERENCES coverage_eligibility_request(id),
    insurer_id      UUID REFERENCES organization(id),
    outcome         VARCHAR(20) NOT NULL DEFAULT 'complete',
    disposition     TEXT,
    created         DATE
);

CREATE INDEX IF NOT EXISTS idx_coverage_elig_resp_patient ON coverage_eligibility_response (patient_id);
CREATE INDEX IF NOT EXISTS idx_coverage_elig_resp_status ON coverage_eligibility_response (status);
CREATE INDEX IF NOT EXISTS idx_coverage_elig_resp_request ON coverage_eligibility_response (request_id);

-- ============================================================
-- Medication / Formulary Tables
-- ============================================================

-- MedicationKnowledge (drug database / formulary)
CREATE TABLE IF NOT EXISTS medication_knowledge (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    code_code       VARCHAR(20),
    code_system     VARCHAR(255),
    code_display    VARCHAR(255),
    manufacturer_id UUID REFERENCES organization(id),
    dose_form_code  VARCHAR(20),
    dose_form_display VARCHAR(255),
    amount_value    NUMERIC,
    amount_unit     VARCHAR(50),
    synonym         TEXT,
    description     TEXT
);

CREATE INDEX IF NOT EXISTS idx_med_knowledge_status ON medication_knowledge (status);
CREATE INDEX IF NOT EXISTS idx_med_knowledge_code ON medication_knowledge (code_code);
CREATE INDEX IF NOT EXISTS idx_med_knowledge_manufacturer ON medication_knowledge (manufacturer_id);

-- ============================================================
-- Organization / Person Tables
-- ============================================================

-- OrganizationAffiliation (organization relationships)
CREATE TABLE IF NOT EXISTS organization_affiliation (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active          BOOLEAN NOT NULL DEFAULT true,
    organization_id UUID REFERENCES organization(id),
    participating_org_id UUID REFERENCES organization(id),
    period_start    TIMESTAMPTZ,
    period_end      TIMESTAMPTZ,
    code_code       VARCHAR(50),
    code_display    VARCHAR(255),
    specialty_code  VARCHAR(50),
    specialty_display VARCHAR(255),
    location_id     UUID REFERENCES location(id),
    telecom_phone   VARCHAR(50),
    telecom_email   VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_org_affiliation_org ON organization_affiliation (organization_id);
CREATE INDEX IF NOT EXISTS idx_org_affiliation_participating ON organization_affiliation (participating_org_id);
CREATE INDEX IF NOT EXISTS idx_org_affiliation_active ON organization_affiliation (active);

-- Person (person demographics linkage)
CREATE TABLE IF NOT EXISTS person (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active          BOOLEAN NOT NULL DEFAULT true,
    name_family     VARCHAR(100),
    name_given      VARCHAR(100),
    gender          VARCHAR(20),
    birth_date      DATE,
    address_line    VARCHAR(255),
    address_city    VARCHAR(100),
    address_state   VARCHAR(50),
    address_postal_code VARCHAR(20),
    telecom_phone   VARCHAR(50),
    telecom_email   VARCHAR(255),
    managing_org_id UUID REFERENCES organization(id)
);

CREATE INDEX IF NOT EXISTS idx_person_active ON person (active);
CREATE INDEX IF NOT EXISTS idx_person_name ON person (name_family, name_given);
CREATE INDEX IF NOT EXISTS idx_person_managing_org ON person (managing_org_id);

-- ============================================================
-- Quality Measure / Clinical Knowledge Tables
-- ============================================================

-- Measure (quality measure definitions)
CREATE TABLE IF NOT EXISTS measure (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    url             VARCHAR(500),
    name            VARCHAR(255),
    title           VARCHAR(500),
    description     TEXT,
    publisher       VARCHAR(255),
    date            DATE,
    effective_period_start TIMESTAMPTZ,
    effective_period_end TIMESTAMPTZ,
    scoring_code    VARCHAR(50),
    scoring_display VARCHAR(100),
    subject_code    VARCHAR(50),
    subject_display VARCHAR(100),
    approval_date   DATE,
    last_review_date DATE
);

CREATE INDEX IF NOT EXISTS idx_measure_status ON measure (status);
CREATE INDEX IF NOT EXISTS idx_measure_url ON measure (url);
CREATE INDEX IF NOT EXISTS idx_measure_name ON measure (name);

-- Library (clinical knowledge artifacts)
CREATE TABLE IF NOT EXISTS library (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    url             VARCHAR(500),
    name            VARCHAR(255),
    title           VARCHAR(500),
    type_code       VARCHAR(50) NOT NULL,
    type_display    VARCHAR(100),
    description     TEXT,
    publisher       VARCHAR(255),
    date            DATE,
    content_type    VARCHAR(100),
    content_data    BYTEA
);

CREATE INDEX IF NOT EXISTS idx_library_status ON library (status);
CREATE INDEX IF NOT EXISTS idx_library_url ON library (url);
CREATE INDEX IF NOT EXISTS idx_library_type ON library (type_code);

-- ============================================================
-- Device Catalog / Metrics Tables
-- ============================================================

-- DeviceDefinition (device catalog)
CREATE TABLE IF NOT EXISTS device_definition (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    manufacturer_string VARCHAR(255),
    model_number    VARCHAR(255),
    device_name     VARCHAR(255),
    device_name_type VARCHAR(50),
    type_code       VARCHAR(50),
    type_display    VARCHAR(255),
    specialization  TEXT,
    safety_code     VARCHAR(50),
    safety_display  VARCHAR(255),
    owner_id        UUID REFERENCES organization(id),
    parent_device_id UUID REFERENCES device_definition(id),
    description     TEXT
);

CREATE INDEX IF NOT EXISTS idx_device_def_type ON device_definition (type_code);
CREATE INDEX IF NOT EXISTS idx_device_def_owner ON device_definition (owner_id);
CREATE INDEX IF NOT EXISTS idx_device_def_model ON device_definition (model_number);

-- DeviceMetric (device measurements)
CREATE TABLE IF NOT EXISTS device_metric (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    type_code       VARCHAR(50) NOT NULL,
    type_display    VARCHAR(255),
    source_id       UUID REFERENCES device(id),
    parent_id       UUID REFERENCES device(id),
    unit_code       VARCHAR(50),
    unit_display    VARCHAR(100),
    operational_status VARCHAR(20),
    color           VARCHAR(20),
    category        VARCHAR(20) NOT NULL,
    calibration_type VARCHAR(20),
    calibration_state VARCHAR(20),
    calibration_time TIMESTAMPTZ,
    measurement_period_value NUMERIC,
    measurement_period_unit VARCHAR(50)
);

CREATE INDEX IF NOT EXISTS idx_device_metric_source ON device_metric (source_id);
CREATE INDEX IF NOT EXISTS idx_device_metric_type ON device_metric (type_code);
CREATE INDEX IF NOT EXISTS idx_device_metric_category ON device_metric (category);

-- ============================================================
-- Specimen Tables
-- ============================================================

-- SpecimenDefinition (specimen type definitions)
CREATE TABLE IF NOT EXISTS specimen_definition (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    type_code       VARCHAR(50),
    type_display    VARCHAR(255),
    patient_preparation TEXT,
    time_aspect     VARCHAR(100),
    collection_code VARCHAR(50),
    collection_display VARCHAR(255),
    handling_temperature_low NUMERIC,
    handling_temperature_high NUMERIC,
    handling_temperature_unit VARCHAR(20),
    handling_max_duration VARCHAR(50),
    handling_instruction TEXT
);

CREATE INDEX IF NOT EXISTS idx_specimen_def_type ON specimen_definition (type_code);
