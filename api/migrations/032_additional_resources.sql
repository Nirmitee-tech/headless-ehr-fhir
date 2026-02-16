-- 032_additional_resources.sql
-- Additional FHIR resource tables

CREATE TABLE IF NOT EXISTS test_report (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) UNIQUE NOT NULL,
    version_id INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    name VARCHAR(255),
    test_script_reference VARCHAR(500),
    result VARCHAR(20) NOT NULL,
    score NUMERIC,
    tester VARCHAR(255),
    issued TIMESTAMPTZ,
    participant_type VARCHAR(20),
    participant_uri VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_test_report_status ON test_report(status);
CREATE INDEX idx_test_report_name ON test_report(name);
CREATE INDEX idx_test_report_result ON test_report(result);

CREATE TABLE IF NOT EXISTS example_scenario (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) UNIQUE NOT NULL,
    version_id INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    url VARCHAR(500),
    name VARCHAR(255),
    title VARCHAR(500),
    description TEXT,
    publisher VARCHAR(255),
    date DATE,
    purpose TEXT,
    copyright TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_example_scenario_status ON example_scenario(status);
CREATE INDEX idx_example_scenario_url ON example_scenario(url);
CREATE INDEX idx_example_scenario_name ON example_scenario(name);

CREATE TABLE IF NOT EXISTS evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) UNIQUE NOT NULL,
    version_id INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    url VARCHAR(500),
    name VARCHAR(255),
    title VARCHAR(500),
    description TEXT,
    publisher VARCHAR(255),
    date DATE,
    outcome_reference VARCHAR(500),
    exposure_background_reference VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evidence_status ON evidence(status);
CREATE INDEX idx_evidence_url ON evidence(url);
CREATE INDEX idx_evidence_name ON evidence(name);

CREATE TABLE IF NOT EXISTS evidence_variable (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) UNIQUE NOT NULL,
    version_id INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    url VARCHAR(500),
    name VARCHAR(255),
    title VARCHAR(500),
    description TEXT,
    publisher VARCHAR(255),
    date DATE,
    type VARCHAR(20) NOT NULL DEFAULT 'dichotomous',
    characteristic_description TEXT,
    characteristic_definition_reference VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evidence_variable_status ON evidence_variable(status);
CREATE INDEX idx_evidence_variable_url ON evidence_variable(url);
CREATE INDEX idx_evidence_variable_name ON evidence_variable(name);

CREATE TABLE IF NOT EXISTS research_definition (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) UNIQUE NOT NULL,
    version_id INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    url VARCHAR(500),
    name VARCHAR(255),
    title VARCHAR(500),
    description TEXT,
    publisher VARCHAR(255),
    date DATE,
    population_reference VARCHAR(500),
    exposure_reference VARCHAR(500),
    outcome_reference VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_research_definition_status ON research_definition(status);
CREATE INDEX idx_research_definition_url ON research_definition(url);
CREATE INDEX idx_research_definition_name ON research_definition(name);

CREATE TABLE IF NOT EXISTS research_element_definition (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) UNIQUE NOT NULL,
    version_id INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    url VARCHAR(500),
    name VARCHAR(255),
    title VARCHAR(500),
    description TEXT,
    publisher VARCHAR(255),
    date DATE,
    type VARCHAR(20) NOT NULL,
    characteristic_definition_reference VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_research_element_definition_status ON research_element_definition(status);
CREATE INDEX idx_research_element_definition_url ON research_element_definition(url);
CREATE INDEX idx_research_element_definition_name ON research_element_definition(name);

CREATE TABLE IF NOT EXISTS effect_evidence_synthesis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) UNIQUE NOT NULL,
    version_id INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    url VARCHAR(500),
    name VARCHAR(255),
    title VARCHAR(500),
    description TEXT,
    publisher VARCHAR(255),
    date DATE,
    population_reference VARCHAR(500),
    exposure_reference VARCHAR(500),
    outcome_reference VARCHAR(500),
    sample_size_description TEXT,
    result_by_exposure_description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_effect_evidence_synthesis_status ON effect_evidence_synthesis(status);
CREATE INDEX idx_effect_evidence_synthesis_url ON effect_evidence_synthesis(url);
CREATE INDEX idx_effect_evidence_synthesis_name ON effect_evidence_synthesis(name);

CREATE TABLE IF NOT EXISTS risk_evidence_synthesis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) UNIQUE NOT NULL,
    version_id INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    url VARCHAR(500),
    name VARCHAR(255),
    title VARCHAR(500),
    description TEXT,
    publisher VARCHAR(255),
    date DATE,
    population_reference VARCHAR(500),
    outcome_reference VARCHAR(500),
    sample_size_description TEXT,
    risk_estimate_description TEXT,
    risk_estimate_value NUMERIC,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_risk_evidence_synthesis_status ON risk_evidence_synthesis(status);
CREATE INDEX idx_risk_evidence_synthesis_url ON risk_evidence_synthesis(url);
CREATE INDEX idx_risk_evidence_synthesis_name ON risk_evidence_synthesis(name);
