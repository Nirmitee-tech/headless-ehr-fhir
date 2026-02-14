-- ============================================================================
-- T0 CORE TABLES MIGRATION
-- Combined from FHIR-compliant EHR schema files
-- FHIR Version: R4 (4.0.1)
-- Supports: US Healthcare (HIPAA, CMS) + India Healthcare (ABDM, ABHA)
-- ============================================================================

-- ============================================================================
-- 1. ORGANIZATION & DEPARTMENT (from 01_foundation_identity.sql)
-- ============================================================================

CREATE TABLE organization (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    name                VARCHAR(255) NOT NULL,
    type_code           VARCHAR(50),          -- prov | dept | team | govt | ins | pay | edu | reli | crs | cg | other
    type_display        VARCHAR(100),
    active              BOOLEAN DEFAULT TRUE,
    parent_org_id       UUID REFERENCES organization(id),
    -- US Identifiers
    npi_number          VARCHAR(20),          -- US: National Provider Identifier
    tin_number          VARCHAR(20),          -- US: Tax Identification Number
    clia_number         VARCHAR(20),          -- US: Clinical Lab Improvement Amendments
    -- India Identifiers
    rohini_id           VARCHAR(20),          -- India: ROHINI ID
    abdm_facility_id    VARCHAR(50),          -- India: ABDM Health Facility Registry ID
    nabh_accreditation  VARCHAR(30),          -- India: NABH accreditation number
    -- Address
    address_line1       VARCHAR(255),
    address_line2       VARCHAR(255),
    city                VARCHAR(100),
    district            VARCHAR(100),         -- India: District
    state               VARCHAR(100),
    postal_code         VARCHAR(20),
    country             VARCHAR(3) DEFAULT 'US',
    phone               VARCHAR(30),
    email               VARCHAR(255),
    website             VARCHAR(255),
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE department (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id     UUID NOT NULL REFERENCES organization(id),
    name                VARCHAR(255) NOT NULL,
    code                VARCHAR(50),
    description         TEXT,
    head_practitioner_id UUID,                -- FK added after practitioner table
    active              BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 2. PRACTITIONER / PROVIDER (from 02_practitioner_provider.sql)
-- ============================================================================

CREATE TABLE practitioner (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    active              BOOLEAN DEFAULT TRUE,
    prefix              VARCHAR(20),
    first_name          VARCHAR(100) NOT NULL,
    middle_name         VARCHAR(100),
    last_name           VARCHAR(100) NOT NULL,
    suffix              VARCHAR(20),
    gender              VARCHAR(20),          -- male | female | other | unknown
    birth_date          DATE,
    photo_url           TEXT,
    -- US Identifiers
    npi_number          VARCHAR(20) UNIQUE,   -- US: National Provider Identifier
    dea_number          VARCHAR(20),          -- US: DEA Number (controlled substances)
    state_license_num   VARCHAR(50),
    state_license_state VARCHAR(5),
    medicare_ptan       VARCHAR(20),          -- US: Medicare PTAN
    medicaid_id         VARCHAR(20),
    upin                VARCHAR(20),          -- US: Unique Physician Identification Number
    -- India Identifiers
    medical_council_reg VARCHAR(50),          -- India: State/National Medical Council Registration
    abha_id             VARCHAR(20),          -- India: Ayushman Bharat Health Account
    aadhaar_hash        VARCHAR(128),         -- India: Aadhaar hash (NEVER store raw)
    hpr_id              VARCHAR(50),          -- India: Healthcare Professionals Registry ID
    ayush_reg_number    VARCHAR(50),          -- India: AYUSH practitioner registration
    -- Contact
    phone               VARCHAR(30),
    email               VARCHAR(255),
    address_line1       VARCHAR(255),
    city                VARCHAR(100),
    state               VARCHAR(100),
    postal_code         VARCHAR(20),
    country             VARCHAR(3) DEFAULT 'US',
    -- Digital Signature (for e-prescriptions)
    digital_signature_cert TEXT,
    qualification_summary TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- Add FK from department to practitioner
ALTER TABLE department ADD CONSTRAINT fk_dept_head
    FOREIGN KEY (head_practitioner_id) REFERENCES practitioner(id);

CREATE TABLE specialty (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code                VARCHAR(20) NOT NULL UNIQUE,
    display             VARCHAR(255) NOT NULL,
    system_uri          VARCHAR(255),         -- SNOMED CT / custom
    category            VARCHAR(50),          -- medical | surgical | diagnostic | allied_health | ayush | dental | nursing
    country_applicability VARCHAR(10),        -- US | IN | BOTH
    description         TEXT,
    parent_specialty_id UUID REFERENCES specialty(id),
    active              BOOLEAN DEFAULT TRUE
);

CREATE TABLE practitioner_specialty (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practitioner_id     UUID NOT NULL REFERENCES practitioner(id),
    specialty_id        UUID NOT NULL REFERENCES specialty(id),
    is_primary          BOOLEAN DEFAULT FALSE,
    board_certified     BOOLEAN DEFAULT FALSE,
    certification_date  DATE,
    certification_expiry DATE,
    certification_body  VARCHAR(255),
    fellowship_details  VARCHAR(255),
    UNIQUE(practitioner_id, specialty_id)
);

CREATE TABLE practitioner_qualification (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practitioner_id     UUID NOT NULL REFERENCES practitioner(id),
    code                VARCHAR(30),
    display             VARCHAR(255) NOT NULL,
    issuer_name         VARCHAR(255),
    issuer_org_id       UUID REFERENCES organization(id),
    period_start        DATE,
    period_end          DATE
);

CREATE TABLE practitioner_role (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    practitioner_id     UUID NOT NULL REFERENCES practitioner(id),
    organization_id     UUID NOT NULL REFERENCES organization(id),
    department_id       UUID REFERENCES department(id),
    role_code           VARCHAR(50),          -- doctor | nurse | pharmacist | admin | technician | therapist | midwife
    role_display        VARCHAR(100),
    period_start        DATE,
    period_end          DATE,
    active              BOOLEAN DEFAULT TRUE,
    available_days      VARCHAR(20)[],        -- {mon,tue,wed,thu,fri}
    available_start     TIME,
    available_end       TIME,
    telehealth_capable  BOOLEAN DEFAULT FALSE,
    accepting_patients  BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 3. PATIENT (from 03_patient.sql)
-- ============================================================================

CREATE TABLE patient (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    active              BOOLEAN DEFAULT TRUE,
    mrn                 VARCHAR(50) UNIQUE NOT NULL,  -- Medical Record Number
    -- Name
    prefix              VARCHAR(20),
    first_name          VARCHAR(100) NOT NULL,
    middle_name         VARCHAR(100),
    last_name           VARCHAR(100) NOT NULL,
    suffix              VARCHAR(20),
    maiden_name         VARCHAR(100),
    -- Core Demographics
    birth_date          DATE,
    gender              VARCHAR(20),           -- male | female | other | unknown
    deceased_boolean    BOOLEAN DEFAULT FALSE,
    deceased_datetime   TIMESTAMPTZ,
    marital_status      VARCHAR(20),           -- S | M | D | W | L | P | T | U
    multiple_birth      BOOLEAN DEFAULT FALSE,
    multiple_birth_int  INTEGER,
    photo_url           TEXT,
    -- US Identifiers
    ssn_hash            VARCHAR(128),          -- US: SSN hash (NEVER store raw)
    drivers_license     VARCHAR(50),
    medicare_id         VARCHAR(20),
    medicaid_id         VARCHAR(20),
    -- India Identifiers
    abha_id             VARCHAR(20),           -- India: Ayushman Bharat Health Account
    abha_address        VARCHAR(100),          -- India: ABHA Address (PHR)
    aadhaar_hash        VARCHAR(128),          -- India: Aadhaar hash (NEVER store raw)
    pan_hash            VARCHAR(128),          -- India: PAN hash
    ration_card_number  VARCHAR(30),
    voter_id            VARCHAR(30),           -- India: Voter ID / EPIC number
    -- Contact
    phone_home          VARCHAR(30),
    phone_mobile        VARCHAR(30),
    phone_work          VARCHAR(30),
    email               VARCHAR(255),
    -- Address
    address_use         VARCHAR(10),           -- home | work | temp | old | billing
    address_line1       VARCHAR(255),
    address_line2       VARCHAR(255),
    city                VARCHAR(100),
    district            VARCHAR(100),          -- India: District
    state               VARCHAR(100),
    postal_code         VARCHAR(20),
    country             VARCHAR(3) DEFAULT 'US',
    -- US Demographics
    race_code           VARCHAR(20),           -- US: OMB Race categories
    race_display        VARCHAR(100),
    ethnicity_code      VARCHAR(20),           -- US: OMB Ethnicity
    ethnicity_display   VARCHAR(100),
    -- General Demographics
    preferred_language  VARCHAR(10) DEFAULT 'en',
    religion            VARCHAR(50),
    blood_group         VARCHAR(5),
    nationality         VARCHAR(50),
    -- Communication
    communication_lang  VARCHAR(10),
    interpreter_needed  BOOLEAN DEFAULT FALSE,
    -- Primary Care
    primary_care_provider_id UUID REFERENCES practitioner(id),
    managing_org_id     UUID REFERENCES organization(id),
    -- Consent & Privacy
    hipaa_consent       BOOLEAN DEFAULT FALSE,     -- US
    hipaa_consent_date  TIMESTAMPTZ,
    abdm_consent        BOOLEAN DEFAULT FALSE,     -- India
    abdm_consent_date   TIMESTAMPTZ,
    advance_directive   BOOLEAN DEFAULT FALSE,
    -- General
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE patient_contact (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    relationship        VARCHAR(30),           -- emergency | family | guardian | friend | agent | caregiver
    prefix              VARCHAR(20),
    first_name          VARCHAR(100),
    last_name           VARCHAR(100),
    phone               VARCHAR(30),
    email               VARCHAR(255),
    address_line1       VARCHAR(255),
    city                VARCHAR(100),
    state               VARCHAR(100),
    postal_code         VARCHAR(20),
    country             VARCHAR(3),
    gender              VARCHAR(20),
    is_primary_contact  BOOLEAN DEFAULT FALSE,
    period_start        DATE,
    period_end          DATE
);

CREATE TABLE patient_identifier (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    system_uri          VARCHAR(255) NOT NULL,
    value               VARCHAR(100) NOT NULL,
    type_code           VARCHAR(30),           -- MR | SS | DL | PPN | ABHA | AADHAAR
    type_display        VARCHAR(100),
    assigner            VARCHAR(255),
    period_start        DATE,
    period_end          DATE,
    UNIQUE(patient_id, system_uri, value)
);

CREATE TABLE patient_link (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    linked_patient_id   UUID NOT NULL REFERENCES patient(id),
    link_type           VARCHAR(20) NOT NULL,  -- replaced-by | replaces | refer | seealso
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. ENCOUNTER / VISIT (from 04_encounter_visit.sql)
-- ============================================================================

CREATE TABLE encounter (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(30) NOT NULL,  -- planned | arrived | triaged | in-progress | onleave | finished | cancelled | entered-in-error
    class_code          VARCHAR(30) NOT NULL,  -- AMB | EMER | IMP | SS | VR (virtual) | HH (home health) | OBSENC | ACUTE | NONAC | PRENC | FLD
    class_display       VARCHAR(100),
    type_code           VARCHAR(50),
    type_display        VARCHAR(255),
    service_type_code   VARCHAR(30),           -- e.g., cardiology, general-practice
    service_type_display VARCHAR(255),
    priority_code       VARCHAR(20),           -- R (routine) | EM (emergency) | UR (urgent) | EL (elective)
    -- Subject
    patient_id          UUID NOT NULL REFERENCES patient(id),
    -- Participants
    primary_practitioner_id UUID REFERENCES practitioner(id),
    -- Organization
    service_provider_id UUID REFERENCES organization(id),
    department_id       UUID REFERENCES department(id),
    -- Period
    period_start        TIMESTAMPTZ NOT NULL,
    period_end          TIMESTAMPTZ,
    length_minutes      INTEGER,
    -- Location
    location_id         UUID,                  -- FK added after location table
    bed_id              UUID,                  -- FK added after bed table
    -- Admission Details (Inpatient)
    admit_source_code   VARCHAR(30),           -- hosp-trans | emd | outp | born | gp | mp | nursing | psych | rehab | other
    admit_source_display VARCHAR(100),
    discharge_disposition_code VARCHAR(30),    -- home | other-hcf | hosp | long | aadvice | exp | psy | rehab | snf | oth
    discharge_disposition_display VARCHAR(100),
    diet_preference     VARCHAR(50),
    special_arrangement VARCHAR(50),           -- wheelchair | stretcher | interpreter
    re_admission        BOOLEAN DEFAULT FALSE,
    -- Referral
    referral_request_id UUID,
    -- India Specific
    ayushman_bharat_claim_id VARCHAR(50),      -- India: AB-PMJAY claim ID
    uhid                VARCHAR(50),           -- India: Unique Hospital ID
    -- US Specific
    drg_code            VARCHAR(10),           -- US: Diagnosis Related Group
    drg_type            VARCHAR(10),           -- MS-DRG | AP-DRG | APR-DRG
    -- Telehealth
    is_telehealth       BOOLEAN DEFAULT FALSE,
    telehealth_platform VARCHAR(100),
    telehealth_url      TEXT,
    -- Chief Complaint (text)
    reason_text         TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE encounter_participant (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    encounter_id        UUID NOT NULL REFERENCES encounter(id),
    practitioner_id     UUID NOT NULL REFERENCES practitioner(id),
    type_code           VARCHAR(30),           -- ATND (attending) | ADM (admitting) | CON (consultant) | REF (referrer) | SPRF (secondary performer) | PPRF (primary performer) | DIS (discharger)
    type_display        VARCHAR(100),
    period_start        TIMESTAMPTZ,
    period_end          TIMESTAMPTZ
);

CREATE TABLE encounter_diagnosis (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    encounter_id        UUID NOT NULL REFERENCES encounter(id),
    condition_id        UUID,                  -- FK added after condition table
    use_code            VARCHAR(30),           -- AD (admission) | DD (discharge) | CC (chief complaint) | CM (comorbidity) | pre-op | post-op
    rank                INTEGER,               -- 1 = principal diagnosis
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE encounter_status_history (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    encounter_id        UUID NOT NULL REFERENCES encounter(id),
    status              VARCHAR(30) NOT NULL,
    period_start        TIMESTAMPTZ NOT NULL,
    period_end          TIMESTAMPTZ
);

-- ============================================================================
-- 5. LOCATION & BED (from 14_location_bed.sql)
-- ============================================================================

CREATE TABLE location (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    status              VARCHAR(20),           -- active | suspended | inactive
    operational_status  VARCHAR(20),           -- C (closed) | H (housekeeping) | I (isolated) | K (contaminated) | O (occupied) | U (unoccupied)
    name                VARCHAR(255) NOT NULL,
    alias               VARCHAR(255)[],
    description         TEXT,
    mode                VARCHAR(20),           -- instance | kind
    -- Type
    type_code           VARCHAR(30),
    type_display        VARCHAR(255),
    -- Physical Type
    physical_type_code  VARCHAR(20),           -- si | bu | wi | wa | lvl | co | ro | bd | ve | ho | ca | rd | area | jdn
    physical_type_display VARCHAR(100),
    -- Hierarchy
    organization_id     UUID REFERENCES organization(id),
    part_of_location_id UUID REFERENCES location(id),
    managing_org_id     UUID REFERENCES organization(id),
    -- Address
    address_line1       VARCHAR(255),
    address_line2       VARCHAR(255),
    city                VARCHAR(100),
    district            VARCHAR(100),
    state               VARCHAR(100),
    postal_code         VARCHAR(20),
    country             VARCHAR(3),
    -- GPS
    latitude            DECIMAL(10,7),
    longitude           DECIMAL(10,7),
    altitude            DECIMAL(10,3),
    -- Telecom
    phone               VARCHAR(30),
    email               VARCHAR(255),
    -- Hours of Operation
    hours_all_day       BOOLEAN DEFAULT FALSE,
    hours_days_of_week  VARCHAR(10)[],
    hours_opening_time  TIME,
    hours_closing_time  TIME,
    availability_exceptions TEXT,
    -- Endpoint (for FHIR endpoint reference)
    endpoint_url        TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE bed (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id         UUID NOT NULL REFERENCES location(id),
    bed_number          VARCHAR(20) NOT NULL,
    bed_type            VARCHAR(30),           -- regular | icu | nicu | picu | ccu | micu | sicu | isolation | bariatric | electric | stretcher | crib | bassinet
    status              VARCHAR(20) NOT NULL DEFAULT 'available', -- available | occupied | reserved | maintenance | contaminated | housekeeping
    ward_name           VARCHAR(100),
    floor               VARCHAR(10),
    wing                VARCHAR(50),
    room_number         VARCHAR(20),
    -- Current Patient (if occupied)
    current_patient_id  UUID REFERENCES patient(id),
    current_encounter_id UUID REFERENCES encounter(id),
    -- Capabilities
    has_oxygen          BOOLEAN DEFAULT FALSE,
    has_suction         BOOLEAN DEFAULT FALSE,
    has_monitor         BOOLEAN DEFAULT FALSE,
    has_ventilator      BOOLEAN DEFAULT FALSE,
    -- Dates
    last_cleaned        TIMESTAMPTZ,
    occupied_since      TIMESTAMPTZ,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(location_id, bed_number)
);

-- Add deferred FK constraints for encounter -> location/bed
ALTER TABLE encounter ADD CONSTRAINT fk_encounter_location
    FOREIGN KEY (location_id) REFERENCES location(id);
ALTER TABLE encounter ADD CONSTRAINT fk_encounter_bed
    FOREIGN KEY (bed_id) REFERENCES bed(id);

-- ============================================================================
-- 6. SYSTEM USER & ROLE ASSIGNMENT (from 41_system_admin_billing_adt.sql)
-- ============================================================================

CREATE TABLE "system_user" (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username            VARCHAR(100) UNIQUE NOT NULL,
    -- Link to practitioner (if clinical user)
    practitioner_id     UUID REFERENCES practitioner(id),
    -- User Type
    user_type           VARCHAR(20) NOT NULL,  -- provider | nurse | admin | billing | registration | pharmacist | lab-tech | rad-tech | therapist | dietary | social-work | it | interface | system
    -- Status
    status              VARCHAR(20) NOT NULL,  -- active | inactive | locked | pending | terminated
    -- Name
    display_name        VARCHAR(255),
    email               VARCHAR(255),
    phone               VARCHAR(30),
    -- Authentication
    last_login          TIMESTAMPTZ,
    failed_login_count  INTEGER DEFAULT 0,
    password_last_changed TIMESTAMPTZ,
    mfa_enabled         BOOLEAN DEFAULT FALSE,
    -- Department Access
    primary_department_id UUID REFERENCES department(id),
    -- Employment
    employee_id         VARCHAR(30),
    hire_date           DATE,
    termination_date    DATE,
    -- Training
    hipaa_training_date DATE,
    last_compliance_training DATE,
    -- Note
    note                TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE user_role_assignment (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES "system_user"(id),
    role_name           VARCHAR(100) NOT NULL,  -- physician | nurse | nurse-manager | pharmacist | lab-tech | rad-tech | registration | billing | coder | case-manager | social-worker | dietitian | respiratory-therapist | admin | super-admin | read-only | research-coordinator
    -- Scope
    organization_id     UUID REFERENCES organization(id),
    department_id       UUID REFERENCES department(id),
    location_id         UUID REFERENCES location(id),
    -- Period
    start_date          DATE DEFAULT CURRENT_DATE,
    end_date            DATE,
    -- Status
    active              BOOLEAN DEFAULT TRUE,
    -- Granted By
    granted_by_id       UUID REFERENCES "system_user"(id),
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 7. AUDIT & HIPAA ACCESS LOG (from 20_audit_provenance.sql)
-- ============================================================================

CREATE TABLE audit_event (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id             VARCHAR(64) UNIQUE NOT NULL,
    -- Type
    type_code           VARCHAR(30) NOT NULL,  -- rest | export | import | query | application-activity | audit-event | user-authentication | emergency-override | network-activity | security-alert
    type_display        VARCHAR(100),
    subtype_code        VARCHAR(30),           -- create | read | vread | update | patch | delete | history | search | batch | transaction | capabilities | execute
    subtype_display     VARCHAR(100),
    -- Action
    action              VARCHAR(5) NOT NULL,   -- C (create) | R (read) | U (update) | D (delete) | E (execute)
    -- Period & Recorded
    period_start        TIMESTAMPTZ,
    period_end          TIMESTAMPTZ,
    recorded            TIMESTAMPTZ DEFAULT NOW(),
    -- Outcome
    outcome             VARCHAR(5) NOT NULL,   -- 0 (success) | 4 (minor failure) | 8 (serious failure) | 12 (major failure)
    outcome_desc        TEXT,
    -- Agent (who performed the action)
    agent_type_code     VARCHAR(20),           -- human | machine
    agent_type_display  VARCHAR(100),
    agent_who_id        UUID,                  -- practitioner or system user ID
    agent_who_display   VARCHAR(255),
    agent_alt_id        VARCHAR(100),
    agent_name          VARCHAR(255),
    agent_requestor     BOOLEAN DEFAULT TRUE,
    agent_role_code     VARCHAR(30),
    agent_role_display  VARCHAR(100),
    -- Network
    agent_network_address VARCHAR(255),        -- IP address or hostname
    agent_network_type  VARCHAR(5),            -- 1 (machine name) | 2 (IP address) | 3 (phone number) | 4 (email address) | 5 (URI)
    -- Source
    source_site         VARCHAR(255),
    source_observer_id  VARCHAR(100),
    source_observer_display VARCHAR(255),
    source_type_code    VARCHAR(30),           -- user-device | data-transport | application-server | database-server | security-server | network-device | network-router | network-gateway | other
    -- Entity (what was acted upon)
    entity_what_type    VARCHAR(50),           -- Patient | Encounter | Observation | etc.
    entity_what_id      UUID,
    entity_what_display VARCHAR(255),
    entity_type_code    VARCHAR(10),           -- 1 (person) | 2 (system) | 3 (organization) | 4 (other)
    entity_role_code    VARCHAR(20),           -- 1 (patient) | 2 (location) | 3 (report) | 4 (resource) | 6 (user) | etc.
    entity_lifecycle     VARCHAR(20),          -- create | access | update | delete | archive | restore | destroy | deidentify | transmit | export
    entity_name         VARCHAR(255),
    entity_description  TEXT,
    entity_query        TEXT,                  -- Base64 encoded query string
    -- Security Labels
    purpose_of_use_code VARCHAR(30),           -- TREAT | HPAYMT | HOPERAT | HRESCH | CLINTRCH | PUBHLTH | PATRQT | FAMRQT
    purpose_of_use_display VARCHAR(100),
    sensitivity_label   VARCHAR(20),           -- N | R | V | U
    -- Extra context
    user_agent_string   VARCHAR(500),
    session_id          VARCHAR(100),
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE hipaa_access_log (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID NOT NULL REFERENCES patient(id),
    accessed_by_id      UUID NOT NULL,         -- practitioner or user ID
    accessed_by_name    VARCHAR(255),
    accessed_by_role    VARCHAR(50),
    -- What was accessed
    resource_type       VARCHAR(50) NOT NULL,
    resource_id         UUID NOT NULL,
    action              VARCHAR(20) NOT NULL,  -- view | create | update | delete | print | export | fax | email
    -- Context
    reason_code         VARCHAR(30),           -- treatment | payment | operations | break-glass | patient-request
    reason_display      VARCHAR(255),
    -- Break Glass (emergency access override)
    is_break_glass      BOOLEAN DEFAULT FALSE,
    break_glass_reason  TEXT,
    -- Network
    ip_address          INET,
    user_agent          VARCHAR(500),
    session_id          VARCHAR(100),
    -- Timestamp
    accessed_at         TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 8. INDEXES (from 21_indices_performance.sql - T0 tables only)
-- ============================================================================

-- Patient indexes
CREATE INDEX idx_patient_mrn ON patient(mrn);
CREATE INDEX idx_patient_name ON patient(last_name, first_name);
CREATE INDEX idx_patient_dob ON patient(birth_date);
CREATE INDEX idx_patient_fhir ON patient(fhir_id);
CREATE INDEX idx_patient_abha ON patient(abha_id) WHERE abha_id IS NOT NULL;
CREATE INDEX idx_patient_phone ON patient(phone_mobile) WHERE phone_mobile IS NOT NULL;
CREATE INDEX idx_patient_email ON patient(email) WHERE email IS NOT NULL;
CREATE INDEX idx_patient_pcp ON patient(primary_care_provider_id) WHERE primary_care_provider_id IS NOT NULL;
CREATE INDEX idx_patient_org ON patient(managing_org_id) WHERE managing_org_id IS NOT NULL;

-- Practitioner indexes
CREATE INDEX idx_practitioner_fhir ON practitioner(fhir_id);
CREATE INDEX idx_practitioner_npi ON practitioner(npi_number) WHERE npi_number IS NOT NULL;
CREATE INDEX idx_practitioner_name ON practitioner(last_name, first_name);
CREATE INDEX idx_practitioner_hpr ON practitioner(hpr_id) WHERE hpr_id IS NOT NULL;

-- Encounter indexes
CREATE INDEX idx_encounter_fhir ON encounter(fhir_id);
CREATE INDEX idx_encounter_patient ON encounter(patient_id);
CREATE INDEX idx_encounter_practitioner ON encounter(primary_practitioner_id);
CREATE INDEX idx_encounter_status ON encounter(status);
CREATE INDEX idx_encounter_class ON encounter(class_code);
CREATE INDEX idx_encounter_date ON encounter(period_start DESC);
CREATE INDEX idx_encounter_dept ON encounter(department_id) WHERE department_id IS NOT NULL;
CREATE INDEX idx_encounter_patient_date ON encounter(patient_id, period_start DESC);

-- Audit indexes
CREATE INDEX idx_audit_recorded ON audit_event(recorded DESC);
CREATE INDEX idx_audit_entity ON audit_event(entity_what_type, entity_what_id);
CREATE INDEX idx_audit_agent ON audit_event(agent_who_id) WHERE agent_who_id IS NOT NULL;
CREATE INDEX idx_audit_action ON audit_event(action);
CREATE INDEX idx_hipaa_log_patient ON hipaa_access_log(patient_id);
CREATE INDEX idx_hipaa_log_accessed_by ON hipaa_access_log(accessed_by_id);
CREATE INDEX idx_hipaa_log_date ON hipaa_access_log(accessed_at DESC);

-- Location / Bed indexes
CREATE INDEX idx_location_org ON location(organization_id) WHERE organization_id IS NOT NULL;
CREATE INDEX idx_location_type ON location(type_code);
CREATE INDEX idx_bed_status ON bed(status);
CREATE INDEX idx_bed_location ON bed(location_id);
CREATE INDEX idx_bed_patient ON bed(current_patient_id) WHERE current_patient_id IS NOT NULL;
