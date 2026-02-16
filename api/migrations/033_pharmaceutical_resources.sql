-- 033: Pharmaceutical, research, and document management resources

-- ResearchSubject: links patients to research studies
CREATE TABLE IF NOT EXISTS research_subject (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    status VARCHAR(32) NOT NULL DEFAULT 'candidate',
    study_reference VARCHAR(255),
    individual_reference VARCHAR(255),
    consent_reference VARCHAR(255),
    period_start TIMESTAMPTZ,
    period_end TIMESTAMPTZ,
    assigned_arm VARCHAR(255),
    actual_arm VARCHAR(255),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_research_subject_fhir ON research_subject(fhir_id);
CREATE INDEX IF NOT EXISTS idx_research_subject_status ON research_subject(status);

-- DocumentManifest: collection of documents for a purpose
CREATE TABLE IF NOT EXISTS document_manifest (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    status VARCHAR(32) NOT NULL DEFAULT 'current',
    type_code VARCHAR(64),
    type_display VARCHAR(255),
    subject_reference VARCHAR(255),
    created TIMESTAMPTZ,
    author_reference VARCHAR(255),
    recipient_reference VARCHAR(255),
    source_url VARCHAR(1024),
    description TEXT,
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_document_manifest_fhir ON document_manifest(fhir_id);
CREATE INDEX IF NOT EXISTS idx_document_manifest_status ON document_manifest(status);

-- SubstanceSpecification: detailed pharmaceutical substance info
CREATE TABLE IF NOT EXISTS substance_specification (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    status VARCHAR(32),
    type_code VARCHAR(64),
    type_display VARCHAR(255),
    domain_code VARCHAR(64),
    domain_display VARCHAR(255),
    description TEXT,
    source_reference VARCHAR(255),
    comment TEXT,
    molecular_weight_amount NUMERIC,
    molecular_weight_unit VARCHAR(64),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_substance_spec_fhir ON substance_specification(fhir_id);

-- MedicinalProduct: core pharmaceutical product
CREATE TABLE IF NOT EXISTS medicinal_product (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    status VARCHAR(32),
    type_code VARCHAR(64),
    type_display VARCHAR(255),
    domain_code VARCHAR(64),
    domain_display VARCHAR(255),
    description TEXT,
    combined_pharmaceutical_dose_form_code VARCHAR(64),
    combined_pharmaceutical_dose_form_display VARCHAR(255),
    legal_status_of_supply_code VARCHAR(64),
    additional_monitoring BOOLEAN DEFAULT false,
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_medicinal_product_fhir ON medicinal_product(fhir_id);

-- MedicinalProductIngredient: ingredients of a medicinal product
CREATE TABLE IF NOT EXISTS medicinal_product_ingredient (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    role_code VARCHAR(64) NOT NULL,
    role_display VARCHAR(255),
    allergenic_indicator BOOLEAN DEFAULT false,
    substance_code VARCHAR(64),
    substance_display VARCHAR(255),
    strength_numerator_value NUMERIC,
    strength_numerator_unit VARCHAR(64),
    strength_denominator_value NUMERIC,
    strength_denominator_unit VARCHAR(64),
    manufacturer_reference VARCHAR(255),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_med_prod_ingredient_fhir ON medicinal_product_ingredient(fhir_id);

-- MedicinalProductManufactured: manufactured dose form
CREATE TABLE IF NOT EXISTS medicinal_product_manufactured (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    manufactured_dose_form_code VARCHAR(64) NOT NULL,
    manufactured_dose_form_display VARCHAR(255),
    unit_of_presentation_code VARCHAR(64),
    unit_of_presentation_display VARCHAR(255),
    quantity_value NUMERIC,
    quantity_unit VARCHAR(64),
    manufacturer_reference VARCHAR(255),
    ingredient_reference VARCHAR(255),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_med_prod_manufactured_fhir ON medicinal_product_manufactured(fhir_id);

-- MedicinalProductPackaged: packaging of a medicinal product
CREATE TABLE IF NOT EXISTS medicinal_product_packaged (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    subject_reference VARCHAR(255),
    description TEXT,
    legal_status_of_supply_code VARCHAR(64),
    legal_status_of_supply_display VARCHAR(255),
    marketing_status_code VARCHAR(64),
    marketing_status_display VARCHAR(255),
    marketing_authorization_reference VARCHAR(255),
    manufacturer_reference VARCHAR(255),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_med_prod_packaged_fhir ON medicinal_product_packaged(fhir_id);

-- MedicinalProductAuthorization: marketing authorization
CREATE TABLE IF NOT EXISTS medicinal_product_authorization (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    status VARCHAR(32),
    status_date TIMESTAMPTZ,
    subject_reference VARCHAR(255),
    country_code VARCHAR(8),
    country_display VARCHAR(255),
    jurisdiction_code VARCHAR(64),
    jurisdiction_display VARCHAR(255),
    validity_period_start TIMESTAMPTZ,
    validity_period_end TIMESTAMPTZ,
    date_of_first_authorization TIMESTAMPTZ,
    international_birth_date TIMESTAMPTZ,
    holder_reference VARCHAR(255),
    regulator_reference VARCHAR(255),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_med_prod_auth_fhir ON medicinal_product_authorization(fhir_id);
CREATE INDEX IF NOT EXISTS idx_med_prod_auth_status ON medicinal_product_authorization(status);

-- MedicinalProductContraindication: contraindications
CREATE TABLE IF NOT EXISTS medicinal_product_contraindication (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    subject_reference VARCHAR(255),
    disease_code VARCHAR(64),
    disease_display VARCHAR(255),
    disease_status_code VARCHAR(64),
    disease_status_display VARCHAR(255),
    comorbidity_code VARCHAR(64),
    comorbidity_display VARCHAR(255),
    therapeutic_indication_reference VARCHAR(255),
    population_age_low NUMERIC,
    population_age_high NUMERIC,
    population_gender_code VARCHAR(32),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_med_prod_contra_fhir ON medicinal_product_contraindication(fhir_id);

-- MedicinalProductIndication: therapeutic indications
CREATE TABLE IF NOT EXISTS medicinal_product_indication (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    subject_reference VARCHAR(255),
    disease_symptom_procedure_code VARCHAR(64),
    disease_symptom_procedure_display VARCHAR(255),
    disease_status_code VARCHAR(64),
    disease_status_display VARCHAR(255),
    comorbidity_code VARCHAR(64),
    comorbidity_display VARCHAR(255),
    intended_effect_code VARCHAR(64),
    intended_effect_display VARCHAR(255),
    duration_value NUMERIC,
    duration_unit VARCHAR(64),
    undesirable_effect_reference VARCHAR(255),
    population_age_low NUMERIC,
    population_age_high NUMERIC,
    population_gender_code VARCHAR(32),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_med_prod_indication_fhir ON medicinal_product_indication(fhir_id);
