-- 034: Substance subtypes and remaining MedicinalProduct resources

-- MedicinalProductInteraction: drug interactions
CREATE TABLE IF NOT EXISTS medicinal_product_interaction (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    subject_reference VARCHAR(255),
    description TEXT,
    type_code VARCHAR(64),
    type_display VARCHAR(255),
    effect_code VARCHAR(64),
    effect_display VARCHAR(255),
    incidence_code VARCHAR(64),
    incidence_display VARCHAR(255),
    management_code VARCHAR(64),
    management_display VARCHAR(255),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_med_prod_interaction_fhir ON medicinal_product_interaction(fhir_id);

-- MedicinalProductUndesirableEffect: side effects
CREATE TABLE IF NOT EXISTS medicinal_product_undesirable_effect (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    subject_reference VARCHAR(255),
    symptom_condition_effect_code VARCHAR(64),
    symptom_condition_effect_display VARCHAR(255),
    classification_code VARCHAR(64),
    classification_display VARCHAR(255),
    frequency_of_occurrence_code VARCHAR(64),
    frequency_of_occurrence_display VARCHAR(255),
    population_age_low NUMERIC,
    population_age_high NUMERIC,
    population_gender_code VARCHAR(32),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_med_prod_undesirable_fhir ON medicinal_product_undesirable_effect(fhir_id);

-- MedicinalProductPharmaceutical: pharmaceutical dose form characteristics
CREATE TABLE IF NOT EXISTS medicinal_product_pharmaceutical (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    administrable_dose_form_code VARCHAR(64) NOT NULL,
    administrable_dose_form_display VARCHAR(255),
    unit_of_presentation_code VARCHAR(64),
    unit_of_presentation_display VARCHAR(255),
    ingredient_reference VARCHAR(255),
    device_reference VARCHAR(255),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_med_prod_pharma_fhir ON medicinal_product_pharmaceutical(fhir_id);

-- SubstancePolymer: polymer substance details
CREATE TABLE IF NOT EXISTS substance_polymer (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    class_code VARCHAR(64),
    class_display VARCHAR(255),
    geometry_code VARCHAR(64),
    geometry_display VARCHAR(255),
    copolymer_connectivity_code VARCHAR(64),
    copolymer_connectivity_display VARCHAR(255),
    modification TEXT,
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_substance_polymer_fhir ON substance_polymer(fhir_id);

-- SubstanceProtein: protein substance details
CREATE TABLE IF NOT EXISTS substance_protein (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    sequence_type_code VARCHAR(64),
    sequence_type_display VARCHAR(255),
    number_of_subunits INT,
    disulfide_linkage TEXT,
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_substance_protein_fhir ON substance_protein(fhir_id);

-- SubstanceNucleicAcid: nucleic acid substance details
CREATE TABLE IF NOT EXISTS substance_nucleic_acid (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    sequence_type_code VARCHAR(64),
    sequence_type_display VARCHAR(255),
    number_of_subunits INT,
    area_of_hybridisation TEXT,
    oligo_nucleotide_type_code VARCHAR(64),
    oligo_nucleotide_type_display VARCHAR(255),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_substance_nucleic_acid_fhir ON substance_nucleic_acid(fhir_id);

-- SubstanceSourceMaterial: source material of a substance
CREATE TABLE IF NOT EXISTS substance_source_material (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    source_material_class_code VARCHAR(64),
    source_material_class_display VARCHAR(255),
    source_material_type_code VARCHAR(64),
    source_material_type_display VARCHAR(255),
    source_material_state_code VARCHAR(64),
    source_material_state_display VARCHAR(255),
    organism_id VARCHAR(255),
    organism_name VARCHAR(255),
    country_of_origin_code VARCHAR(8),
    country_of_origin_display VARCHAR(255),
    geographical_location TEXT,
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_substance_source_fhir ON substance_source_material(fhir_id);

-- SubstanceReferenceInformation: reference information for substances
CREATE TABLE IF NOT EXISTS substance_reference_information (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id VARCHAR(64) NOT NULL UNIQUE DEFAULT gen_random_uuid()::text,
    comment TEXT,
    gene_element_type_code VARCHAR(64),
    gene_element_type_display VARCHAR(255),
    gene_element_source_reference VARCHAR(255),
    classification_code VARCHAR(64),
    classification_display VARCHAR(255),
    classification_domain_code VARCHAR(64),
    classification_domain_display VARCHAR(255),
    target_type_code VARCHAR(64),
    target_type_display VARCHAR(255),
    version_id INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_substance_ref_info_fhir ON substance_reference_information(fhir_id);
