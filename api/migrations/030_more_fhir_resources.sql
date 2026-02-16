-- 030_more_fhir_resources.sql
-- Adds 10 new FHIR R4 resource tables: CommunicationRequest,
-- ObservationDefinition, Linkage, Basic, VerificationResult,
-- EventDefinition, GraphDefinition, MolecularSequence,
-- BiologicallyDerivedProduct, CatalogEntry

-- ============================================================
-- Communication / Messaging Tables
-- ============================================================

-- CommunicationRequest (request for communication)
CREATE TABLE IF NOT EXISTS communication_request (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    patient_id      UUID REFERENCES patient(id),
    encounter_id    UUID REFERENCES encounter(id),
    requester_id    UUID REFERENCES practitioner(id),
    recipient_id    UUID REFERENCES practitioner(id),
    sender_id       UUID REFERENCES practitioner(id),
    category_code   VARCHAR(50),
    category_display VARCHAR(255),
    priority        VARCHAR(20),
    medium_code     VARCHAR(50),
    medium_display  VARCHAR(255),
    payload_text    TEXT,
    occurrence_date TIMESTAMPTZ,
    authored_on     TIMESTAMPTZ,
    reason_code     VARCHAR(50),
    reason_display  VARCHAR(255),
    note            TEXT
);

CREATE INDEX IF NOT EXISTS idx_comm_req_patient ON communication_request (patient_id);
CREATE INDEX IF NOT EXISTS idx_comm_req_encounter ON communication_request (encounter_id);
CREATE INDEX IF NOT EXISTS idx_comm_req_status ON communication_request (status);
CREATE INDEX IF NOT EXISTS idx_comm_req_requester ON communication_request (requester_id);
CREATE INDEX IF NOT EXISTS idx_comm_req_recipient ON communication_request (recipient_id);
CREATE INDEX IF NOT EXISTS idx_comm_req_priority ON communication_request (priority);
CREATE INDEX IF NOT EXISTS idx_comm_req_authored ON communication_request (authored_on);

-- ============================================================
-- Observation / Definition Tables
-- ============================================================

-- ObservationDefinition (observation type definitions)
CREATE TABLE IF NOT EXISTS observation_definition (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    category_code   VARCHAR(50),
    category_display VARCHAR(255),
    code_code       VARCHAR(50) NOT NULL,
    code_system     VARCHAR(255),
    code_display    VARCHAR(255),
    permitted_data_type VARCHAR(50),
    multiple_results_allowed BOOLEAN NOT NULL DEFAULT false,
    method_code     VARCHAR(50),
    method_display  VARCHAR(255),
    preferred_report_name VARCHAR(255),
    unit_code       VARCHAR(50),
    unit_display    VARCHAR(100),
    normal_value_low NUMERIC,
    normal_value_high NUMERIC
);

CREATE INDEX IF NOT EXISTS idx_obs_def_status ON observation_definition (status);
CREATE INDEX IF NOT EXISTS idx_obs_def_code ON observation_definition (code_code);
CREATE INDEX IF NOT EXISTS idx_obs_def_category ON observation_definition (category_code);

-- ============================================================
-- Infrastructure / Linkage Tables
-- ============================================================

-- Linkage (resource linking)
CREATE TABLE IF NOT EXISTS linkage (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active          BOOLEAN NOT NULL DEFAULT true,
    author_id       UUID REFERENCES practitioner(id),
    source_type     VARCHAR(50) NOT NULL,
    source_reference VARCHAR(255) NOT NULL,
    alternate_type  VARCHAR(50),
    alternate_reference VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_linkage_active ON linkage (active);
CREATE INDEX IF NOT EXISTS idx_linkage_author ON linkage (author_id);
CREATE INDEX IF NOT EXISTS idx_linkage_source ON linkage (source_type, source_reference);

-- Basic (generic FHIR resource)
CREATE TABLE IF NOT EXISTS basic (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    code_code       VARCHAR(50) NOT NULL,
    code_system     VARCHAR(255),
    code_display    VARCHAR(255),
    subject_type    VARCHAR(50),
    subject_reference VARCHAR(255),
    author_id       UUID REFERENCES practitioner(id),
    author_date     DATE
);

CREATE INDEX IF NOT EXISTS idx_basic_code ON basic (code_code);
CREATE INDEX IF NOT EXISTS idx_basic_subject ON basic (subject_type, subject_reference);
CREATE INDEX IF NOT EXISTS idx_basic_author ON basic (author_id);

-- ============================================================
-- Verification / Attestation Tables
-- ============================================================

-- VerificationResult (data verification)
CREATE TABLE IF NOT EXISTS verification_result (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'attested',
    target_type     VARCHAR(50),
    target_reference VARCHAR(255),
    need_code       VARCHAR(50),
    need_display    VARCHAR(100),
    status_date     TIMESTAMPTZ,
    validation_type_code VARCHAR(50),
    validation_type_display VARCHAR(255),
    validation_process_code VARCHAR(50),
    validation_process_display VARCHAR(255),
    frequency_value INT,
    frequency_unit  VARCHAR(50),
    last_performed  TIMESTAMPTZ,
    next_scheduled  DATE,
    failure_action_code VARCHAR(50),
    failure_action_display VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_verif_result_status ON verification_result (status);
CREATE INDEX IF NOT EXISTS idx_verif_result_target ON verification_result (target_type, target_reference);
CREATE INDEX IF NOT EXISTS idx_verif_result_status_date ON verification_result (status_date);
CREATE INDEX IF NOT EXISTS idx_verif_result_next_scheduled ON verification_result (next_scheduled);

-- ============================================================
-- Workflow / Event Tables
-- ============================================================

-- EventDefinition (workflow event triggers)
CREATE TABLE IF NOT EXISTS event_definition (
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
    purpose         TEXT,
    trigger_type    VARCHAR(50) NOT NULL,
    trigger_name    VARCHAR(255),
    trigger_condition TEXT
);

CREATE INDEX IF NOT EXISTS idx_event_def_status ON event_definition (status);
CREATE INDEX IF NOT EXISTS idx_event_def_url ON event_definition (url);
CREATE INDEX IF NOT EXISTS idx_event_def_name ON event_definition (name);
CREATE INDEX IF NOT EXISTS idx_event_def_trigger_type ON event_definition (trigger_type);

-- ============================================================
-- Graph / Query Definition Tables
-- ============================================================

-- GraphDefinition (graph query definitions)
CREATE TABLE IF NOT EXISTS graph_definition (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    url             VARCHAR(500),
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    publisher       VARCHAR(255),
    date            DATE,
    start_type      VARCHAR(50) NOT NULL,
    profile         VARCHAR(500)
);

CREATE INDEX IF NOT EXISTS idx_graph_def_status ON graph_definition (status);
CREATE INDEX IF NOT EXISTS idx_graph_def_url ON graph_definition (url);
CREATE INDEX IF NOT EXISTS idx_graph_def_name ON graph_definition (name);
CREATE INDEX IF NOT EXISTS idx_graph_def_start_type ON graph_definition (start_type);

-- ============================================================
-- Genomics / Molecular Tables
-- ============================================================

-- MolecularSequence (genomics data)
CREATE TABLE IF NOT EXISTS molecular_sequence (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    type            VARCHAR(20) NOT NULL,
    patient_id      UUID REFERENCES patient(id),
    specimen_id     UUID REFERENCES specimen(id),
    device_id       UUID REFERENCES device(id),
    performer_id    UUID REFERENCES organization(id),
    coordinate_system INT NOT NULL DEFAULT 0,
    observed_seq    TEXT,
    reference_seq_id VARCHAR(100),
    reference_seq_strand VARCHAR(20),
    window_start    INT,
    window_end      INT
);

CREATE INDEX IF NOT EXISTS idx_mol_seq_type ON molecular_sequence (type);
CREATE INDEX IF NOT EXISTS idx_mol_seq_patient ON molecular_sequence (patient_id);
CREATE INDEX IF NOT EXISTS idx_mol_seq_specimen ON molecular_sequence (specimen_id);
CREATE INDEX IF NOT EXISTS idx_mol_seq_performer ON molecular_sequence (performer_id);
CREATE INDEX IF NOT EXISTS idx_mol_seq_ref_seq ON molecular_sequence (reference_seq_id);

-- ============================================================
-- Biological Products / Blood Bank Tables
-- ============================================================

-- BiologicallyDerivedProduct (blood bank / tissue tracking)
CREATE TABLE IF NOT EXISTS biologically_derived_product (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    product_category VARCHAR(20),
    product_code_code VARCHAR(50),
    product_code_display VARCHAR(255),
    status          VARCHAR(20),
    request_id      UUID,
    quantity        INT,
    parent_id       UUID REFERENCES biologically_derived_product(id),
    collection_source_type VARCHAR(50),
    collection_source_reference VARCHAR(255),
    collection_collected_date TIMESTAMPTZ,
    processing_description TEXT,
    storage_temperature_code VARCHAR(50),
    storage_duration VARCHAR(100)
);

CREATE INDEX IF NOT EXISTS idx_bio_product_category ON biologically_derived_product (product_category);
CREATE INDEX IF NOT EXISTS idx_bio_product_status ON biologically_derived_product (status);
CREATE INDEX IF NOT EXISTS idx_bio_product_code ON biologically_derived_product (product_code_code);
CREATE INDEX IF NOT EXISTS idx_bio_product_parent ON biologically_derived_product (parent_id);
CREATE INDEX IF NOT EXISTS idx_bio_product_collection_date ON biologically_derived_product (collection_collected_date);

-- ============================================================
-- Catalog / Registry Tables
-- ============================================================

-- CatalogEntry (catalog items)
CREATE TABLE IF NOT EXISTS catalog_entry (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    type            VARCHAR(20),
    orderable       BOOLEAN NOT NULL DEFAULT true,
    referenced_item_type VARCHAR(50) NOT NULL,
    referenced_item_reference VARCHAR(255) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    effective_period_start TIMESTAMPTZ,
    effective_period_end TIMESTAMPTZ,
    additional_identifier VARCHAR(255),
    classification_code VARCHAR(50),
    classification_display VARCHAR(255),
    validity_period_start TIMESTAMPTZ,
    validity_period_end TIMESTAMPTZ,
    last_updated    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_catalog_entry_type ON catalog_entry (type);
CREATE INDEX IF NOT EXISTS idx_catalog_entry_status ON catalog_entry (status);
CREATE INDEX IF NOT EXISTS idx_catalog_entry_orderable ON catalog_entry (orderable);
CREATE INDEX IF NOT EXISTS idx_catalog_entry_ref_item ON catalog_entry (referenced_item_type, referenced_item_reference);
CREATE INDEX IF NOT EXISTS idx_catalog_entry_classification ON catalog_entry (classification_code);
