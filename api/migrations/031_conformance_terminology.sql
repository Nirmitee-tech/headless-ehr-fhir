-- 031_conformance_terminology.sql
-- Adds 10 new FHIR R4 resource tables: StructureDefinition,
-- SearchParameter, CodeSystem, ValueSet, ConceptMap,
-- ImplementationGuide, CompartmentDefinition, TerminologyCapabilities,
-- StructureMap, TestScript

-- ============================================================
-- Conformance / Profile Tables
-- ============================================================

-- StructureDefinition (profile definitions)
CREATE TABLE IF NOT EXISTS structure_definition (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    url             VARCHAR(500) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    title           VARCHAR(500),
    description     TEXT,
    publisher       VARCHAR(255),
    date            DATE,
    kind            VARCHAR(20) NOT NULL,
    abstract        BOOLEAN NOT NULL DEFAULT false,
    type            VARCHAR(100) NOT NULL,
    base_definition VARCHAR(500),
    derivation      VARCHAR(20),
    context_type    VARCHAR(50)
);

CREATE INDEX IF NOT EXISTS idx_struct_def_status ON structure_definition (status);
CREATE INDEX IF NOT EXISTS idx_struct_def_url ON structure_definition (url);
CREATE INDEX IF NOT EXISTS idx_struct_def_name ON structure_definition (name);
CREATE INDEX IF NOT EXISTS idx_struct_def_kind ON structure_definition (kind);
CREATE INDEX IF NOT EXISTS idx_struct_def_type ON structure_definition (type);
CREATE INDEX IF NOT EXISTS idx_struct_def_base_def ON structure_definition (base_definition);

-- ============================================================
-- Search / Query Parameter Tables
-- ============================================================

-- SearchParameter (custom search parameters)
CREATE TABLE IF NOT EXISTS search_parameter (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    url             VARCHAR(500) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT NOT NULL,
    code            VARCHAR(100) NOT NULL,
    base            VARCHAR(50) NOT NULL,
    type            VARCHAR(20) NOT NULL,
    expression      TEXT,
    xpath           TEXT,
    target          VARCHAR(255),
    modifier        VARCHAR(255),
    comparator      VARCHAR(255),
    publisher       VARCHAR(255),
    date            DATE
);

CREATE INDEX IF NOT EXISTS idx_search_param_status ON search_parameter (status);
CREATE INDEX IF NOT EXISTS idx_search_param_url ON search_parameter (url);
CREATE INDEX IF NOT EXISTS idx_search_param_name ON search_parameter (name);
CREATE INDEX IF NOT EXISTS idx_search_param_code ON search_parameter (code);
CREATE INDEX IF NOT EXISTS idx_search_param_base ON search_parameter (base);
CREATE INDEX IF NOT EXISTS idx_search_param_type ON search_parameter (type);

-- ============================================================
-- Terminology / Code System Tables
-- ============================================================

-- CodeSystem (code system definitions)
CREATE TABLE IF NOT EXISTS code_system (
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
    content         VARCHAR(20) NOT NULL,
    value_set       VARCHAR(500),
    hierarchy_meaning VARCHAR(20),
    compositional   BOOLEAN NOT NULL DEFAULT false,
    version_needed  BOOLEAN NOT NULL DEFAULT false,
    count           INT
);

CREATE INDEX IF NOT EXISTS idx_code_system_status ON code_system (status);
CREATE INDEX IF NOT EXISTS idx_code_system_url ON code_system (url);
CREATE INDEX IF NOT EXISTS idx_code_system_name ON code_system (name);
CREATE INDEX IF NOT EXISTS idx_code_system_content ON code_system (content);

-- ValueSet (value set definitions)
CREATE TABLE IF NOT EXISTS value_set (
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
    immutable       BOOLEAN NOT NULL DEFAULT false,
    purpose         TEXT,
    copyright       TEXT,
    compose_include_system VARCHAR(500),
    compose_include_version VARCHAR(100),
    expansion_identifier VARCHAR(255),
    expansion_timestamp TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_value_set_status ON value_set (status);
CREATE INDEX IF NOT EXISTS idx_value_set_url ON value_set (url);
CREATE INDEX IF NOT EXISTS idx_value_set_name ON value_set (name);

-- ConceptMap (terminology mapping)
CREATE TABLE IF NOT EXISTS concept_map (
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
    source_uri      VARCHAR(500),
    target_uri      VARCHAR(500),
    purpose         TEXT
);

CREATE INDEX IF NOT EXISTS idx_concept_map_status ON concept_map (status);
CREATE INDEX IF NOT EXISTS idx_concept_map_url ON concept_map (url);
CREATE INDEX IF NOT EXISTS idx_concept_map_name ON concept_map (name);
CREATE INDEX IF NOT EXISTS idx_concept_map_source ON concept_map (source_uri);
CREATE INDEX IF NOT EXISTS idx_concept_map_target ON concept_map (target_uri);

-- ============================================================
-- Implementation Guide / Packaging Tables
-- ============================================================

-- ImplementationGuide (IG definitions)
CREATE TABLE IF NOT EXISTS implementation_guide (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    url             VARCHAR(500) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    title           VARCHAR(500),
    description     TEXT,
    publisher       VARCHAR(255),
    date            DATE,
    package_id      VARCHAR(255),
    fhir_version    VARCHAR(20),
    license         VARCHAR(50),
    depends_on_uri  VARCHAR(500),
    global_type     VARCHAR(100),
    global_profile  VARCHAR(500)
);

CREATE INDEX IF NOT EXISTS idx_impl_guide_status ON implementation_guide (status);
CREATE INDEX IF NOT EXISTS idx_impl_guide_url ON implementation_guide (url);
CREATE INDEX IF NOT EXISTS idx_impl_guide_name ON implementation_guide (name);
CREATE INDEX IF NOT EXISTS idx_impl_guide_package ON implementation_guide (package_id);
CREATE INDEX IF NOT EXISTS idx_impl_guide_fhir_version ON implementation_guide (fhir_version);

-- ============================================================
-- Compartment / Access Control Tables
-- ============================================================

-- CompartmentDefinition (compartment definitions)
CREATE TABLE IF NOT EXISTS compartment_definition (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    url             VARCHAR(500) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    publisher       VARCHAR(255),
    date            DATE,
    code            VARCHAR(20) NOT NULL,
    search          BOOLEAN NOT NULL DEFAULT true,
    resource_type   VARCHAR(50),
    resource_param  VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_compartment_def_status ON compartment_definition (status);
CREATE INDEX IF NOT EXISTS idx_compartment_def_url ON compartment_definition (url);
CREATE INDEX IF NOT EXISTS idx_compartment_def_name ON compartment_definition (name);
CREATE INDEX IF NOT EXISTS idx_compartment_def_code ON compartment_definition (code);

-- ============================================================
-- Terminology Server Capabilities Tables
-- ============================================================

-- TerminologyCapabilities (terminology server capabilities)
CREATE TABLE IF NOT EXISTS terminology_capabilities (
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
    kind            VARCHAR(20) NOT NULL DEFAULT 'instance',
    code_search     VARCHAR(20),
    translation     BOOLEAN NOT NULL DEFAULT false,
    closure         BOOLEAN NOT NULL DEFAULT false,
    software_name   VARCHAR(255),
    software_version VARCHAR(100)
);

CREATE INDEX IF NOT EXISTS idx_term_cap_status ON terminology_capabilities (status);
CREATE INDEX IF NOT EXISTS idx_term_cap_url ON terminology_capabilities (url);
CREATE INDEX IF NOT EXISTS idx_term_cap_name ON terminology_capabilities (name);
CREATE INDEX IF NOT EXISTS idx_term_cap_kind ON terminology_capabilities (kind);

-- ============================================================
-- Structure Transform Tables
-- ============================================================

-- StructureMap (transform definitions)
CREATE TABLE IF NOT EXISTS structure_map (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    url             VARCHAR(500) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    title           VARCHAR(500),
    description     TEXT,
    publisher       VARCHAR(255),
    date            DATE,
    structure_url   VARCHAR(500),
    structure_mode  VARCHAR(20),
    import_uri      VARCHAR(500)
);

CREATE INDEX IF NOT EXISTS idx_struct_map_status ON structure_map (status);
CREATE INDEX IF NOT EXISTS idx_struct_map_url ON structure_map (url);
CREATE INDEX IF NOT EXISTS idx_struct_map_name ON structure_map (name);

-- ============================================================
-- Test Automation Tables
-- ============================================================

-- TestScript (test automation definitions)
CREATE TABLE IF NOT EXISTS test_script (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         VARCHAR(64) UNIQUE NOT NULL,
    version_id      INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    url             VARCHAR(500),
    name            VARCHAR(255) NOT NULL,
    title           VARCHAR(500),
    description     TEXT,
    publisher       VARCHAR(255),
    date            DATE,
    purpose         TEXT,
    copyright       TEXT,
    profile_reference VARCHAR(500),
    origin_index    INT,
    destination_index INT
);

CREATE INDEX IF NOT EXISTS idx_test_script_status ON test_script (status);
CREATE INDEX IF NOT EXISTS idx_test_script_url ON test_script (url);
CREATE INDEX IF NOT EXISTS idx_test_script_name ON test_script (name);
