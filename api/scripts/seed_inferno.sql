-- ============================================================================
-- Inferno g(10) Standardized API Test Kit — Seed Data
-- Covers all US Core FHIR R4 profiles for ONC (g)(10) certification testing
--
-- Run:  psql -d ehr -f scripts/seed_inferno.sql
-- Idempotent: uses INSERT ... ON CONFLICT DO NOTHING throughout
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. ORGANIZATION
-- ============================================================================

INSERT INTO organization (
    id, fhir_id, name, type_code, type_display, active,
    npi_number, address_line1, city, state, postal_code, country,
    phone, email, version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000001'::uuid,
    'org-acme-general-hospital',
    'Acme General Hospital',
    'prov', 'Healthcare Provider',
    TRUE,
    '1234567890',
    '123 Main Street', 'Anytown', 'CA', '90210', 'US',
    '555-555-0100', 'info@acmegeneral.example.com',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 2. LOCATION
-- ============================================================================

INSERT INTO location (
    id, fhir_id, status, name, description, mode,
    type_code, type_display,
    physical_type_code, physical_type_display,
    organization_id, managing_org_id,
    address_line1, city, state, postal_code, country,
    phone, version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000002'::uuid,
    'loc-main-campus',
    'active',
    'Main Campus',
    'Acme General Hospital Main Campus',
    'instance',
    'HOSP', 'Hospital',
    'bu', 'Building',
    'a0000000-0000-4000-8000-000000000001'::uuid,
    'a0000000-0000-4000-8000-000000000001'::uuid,
    '123 Main Street', 'Anytown', 'CA', '90210', 'US',
    '555-555-0100',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 3. PRACTITIONERS
-- ============================================================================

-- Dr. Smith (Physician)
INSERT INTO practitioner (
    id, fhir_id, active, prefix, first_name, last_name,
    gender, npi_number,
    phone, email, city, state, postal_code, country,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'practitioner-dr-smith',
    TRUE, 'Dr.', 'Robert', 'Smith',
    'male', '9999999901',
    '555-555-0110', 'robert.smith@acmegeneral.example.com',
    'Anytown', 'CA', '90210', 'US',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Nurse Johnson
INSERT INTO practitioner (
    id, fhir_id, active, prefix, first_name, last_name,
    gender, npi_number,
    phone, email, city, state, postal_code, country,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000011'::uuid,
    'practitioner-nurse-johnson',
    TRUE, NULL, 'Sarah', 'Johnson',
    'female', '9999999902',
    '555-555-0111', 'sarah.johnson@acmegeneral.example.com',
    'Anytown', 'CA', '90210', 'US',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 4. PRACTITIONER ROLES
-- ============================================================================

INSERT INTO practitioner_role (
    id, fhir_id, practitioner_id, organization_id,
    role_code, role_display, active, period_start, version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000012'::uuid,
    'practitionerrole-dr-smith',
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000001'::uuid,
    'doctor', 'Doctor',
    TRUE, '2015-01-01', 1
) ON CONFLICT (fhir_id) DO NOTHING;

INSERT INTO practitioner_role (
    id, fhir_id, practitioner_id, organization_id,
    role_code, role_display, active, period_start, version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000013'::uuid,
    'practitionerrole-nurse-johnson',
    'a0000000-0000-4000-8000-000000000011'::uuid,
    'a0000000-0000-4000-8000-000000000001'::uuid,
    'nurse', 'Nurse',
    TRUE, '2018-06-01', 1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 5. PATIENTS
-- ============================================================================

-- Patient A: John Smith
INSERT INTO patient (
    id, fhir_id, active, mrn,
    prefix, first_name, last_name,
    birth_date, gender,
    phone_home, phone_mobile, email,
    address_use, address_line1, city, state, postal_code, country,
    race_code, race_display, race_text,
    ethnicity_code, ethnicity_display, ethnicity_text,
    birth_sex,
    preferred_language,
    managing_org_id, primary_care_provider_id,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'patient-john-smith',
    TRUE, 'MRN-001',
    'Mr.', 'John', 'Smith',
    '1970-01-15', 'male',
    '555-555-0201', '555-555-0202', 'john.smith@example.com',
    'home', '456 Oak Avenue', 'Anytown', 'CA', '90210', 'US',
    '2106-3', 'White', 'White',
    '2186-5', 'Not Hispanic or Latino', 'Not Hispanic or Latino',
    'M',
    'en',
    'a0000000-0000-4000-8000-000000000001'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Patient B: Maria Garcia
INSERT INTO patient (
    id, fhir_id, active, mrn,
    prefix, first_name, last_name,
    birth_date, gender,
    phone_home, phone_mobile, email,
    address_use, address_line1, city, state, postal_code, country,
    race_code, race_display, race_text,
    ethnicity_code, ethnicity_display, ethnicity_text,
    birth_sex,
    preferred_language,
    managing_org_id, primary_care_provider_id,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'patient-maria-garcia',
    TRUE, 'MRN-002',
    'Mrs.', 'Maria', 'Garcia',
    '1985-06-20', 'female',
    '555-555-0301', '555-555-0302', 'maria.garcia@example.com',
    'home', '789 Elm Street', 'Anytown', 'CA', '90211', 'US',
    '2131-1', 'Other Race', 'Other',
    '2135-2', 'Hispanic or Latino', 'Hispanic or Latino',
    'F',
    'es',
    'a0000000-0000-4000-8000-000000000001'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 6. ENCOUNTERS (one per patient, ambulatory, finished)
-- ============================================================================

-- Encounter for John Smith
INSERT INTO encounter (
    id, fhir_id, status, class_code, class_display,
    type_code, type_display,
    patient_id, primary_practitioner_id, service_provider_id,
    location_id,
    period_start, period_end, length_minutes,
    reason_text,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'encounter-john-smith-1',
    'finished', 'AMB', 'ambulatory',
    '99213', 'Office or other outpatient visit',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000001'::uuid,
    'a0000000-0000-4000-8000-000000000002'::uuid,
    '2025-01-15 09:00:00-08', '2025-01-15 09:30:00-08', 30,
    'Routine follow-up for diabetes management',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Encounter for Maria Garcia
INSERT INTO encounter (
    id, fhir_id, status, class_code, class_display,
    type_code, type_display,
    patient_id, primary_practitioner_id, service_provider_id,
    location_id,
    period_start, period_end, length_minutes,
    reason_text,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000201'::uuid,
    'encounter-maria-garcia-1',
    'finished', 'AMB', 'ambulatory',
    '99213', 'Office or other outpatient visit',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000001'::uuid,
    'a0000000-0000-4000-8000-000000000002'::uuid,
    '2025-02-10 10:00:00-08', '2025-02-10 10:45:00-08', 45,
    'Annual physical exam',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 7. CONDITIONS
-- ============================================================================

-- Patient A: Type 2 Diabetes Mellitus (problem-list-item)
INSERT INTO condition (
    id, fhir_id, patient_id, encounter_id, recorder_id, asserter_id,
    clinical_status, verification_status, category_code,
    code_system, code_value, code_display,
    onset_datetime, recorded_date,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000300'::uuid,
    'condition-john-diabetes',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'active', 'confirmed', 'problem-list-item',
    'http://snomed.info/sct', '44054006', 'Type 2 diabetes mellitus',
    '2018-03-15 00:00:00-08', '2018-03-15 00:00:00-08',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Patient A: Headache (encounter-diagnosis)
INSERT INTO condition (
    id, fhir_id, patient_id, encounter_id, recorder_id, asserter_id,
    clinical_status, verification_status, category_code,
    code_system, code_value, code_display,
    onset_datetime, recorded_date,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000301'::uuid,
    'condition-john-headache',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'active', 'confirmed', 'encounter-diagnosis',
    'http://snomed.info/sct', '25064002', 'Headache',
    '2025-01-15 00:00:00-08', '2025-01-15 00:00:00-08',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Patient B: Essential hypertension (problem-list-item)
INSERT INTO condition (
    id, fhir_id, patient_id, encounter_id, recorder_id, asserter_id,
    clinical_status, verification_status, category_code,
    code_system, code_value, code_display,
    onset_datetime, recorded_date,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000302'::uuid,
    'condition-maria-hypertension',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'active', 'confirmed', 'problem-list-item',
    'http://snomed.info/sct', '59621000', 'Essential hypertension',
    '2020-05-10 00:00:00-08', '2020-05-10 00:00:00-08',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Patient B: Seasonal allergic rhinitis (encounter-diagnosis)
INSERT INTO condition (
    id, fhir_id, patient_id, encounter_id, recorder_id, asserter_id,
    clinical_status, verification_status, category_code,
    code_system, code_value, code_display,
    onset_datetime, recorded_date,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000303'::uuid,
    'condition-maria-allergic-rhinitis',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'active', 'confirmed', 'encounter-diagnosis',
    'http://snomed.info/sct', '367498001', 'Seasonal allergic rhinitis',
    '2025-02-10 00:00:00-08', '2025-02-10 00:00:00-08',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 8. OBSERVATIONS
-- ============================================================================

-- ----- Patient A (John Smith) -----

-- Blood Pressure (panel observation, with components)
INSERT INTO observation (
    id, fhir_id, status, category_code, category_display,
    code_system, code_value, code_display,
    patient_id, encounter_id, performer_id,
    effective_datetime,
    has_member,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000400'::uuid,
    'obs-john-bp',
    'final', 'vital-signs', 'Vital Signs',
    'http://loinc.org', '85354-9', 'Blood pressure panel with all children optional',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'a0000000-0000-4000-8000-000000000011'::uuid,
    '2025-01-15 09:10:00-08',
    FALSE,
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- BP: Systolic component
INSERT INTO observation_component (
    id, observation_id,
    code_system, code_value, code_display,
    value_quantity, value_unit
) VALUES (
    'a0000000-0000-4000-8000-000000000410'::uuid,
    'a0000000-0000-4000-8000-000000000400'::uuid,
    'http://loinc.org', '8480-6', 'Systolic blood pressure',
    120.0000, 'mmHg'
) ON CONFLICT DO NOTHING;

-- BP: Diastolic component
INSERT INTO observation_component (
    id, observation_id,
    code_system, code_value, code_display,
    value_quantity, value_unit
) VALUES (
    'a0000000-0000-4000-8000-000000000411'::uuid,
    'a0000000-0000-4000-8000-000000000400'::uuid,
    'http://loinc.org', '8462-4', 'Diastolic blood pressure',
    80.0000, 'mmHg'
) ON CONFLICT DO NOTHING;

-- Heart Rate
INSERT INTO observation (
    id, fhir_id, status, category_code, category_display,
    code_system, code_value, code_display,
    patient_id, encounter_id, performer_id,
    effective_datetime,
    value_quantity, value_unit, value_system, value_code,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000401'::uuid,
    'obs-john-heart-rate',
    'final', 'vital-signs', 'Vital Signs',
    'http://loinc.org', '8867-4', 'Heart rate',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'a0000000-0000-4000-8000-000000000011'::uuid,
    '2025-01-15 09:10:00-08',
    72.0000, '/min', 'http://unitsofmeasure.org', '/min',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- BMI
INSERT INTO observation (
    id, fhir_id, status, category_code, category_display,
    code_system, code_value, code_display,
    patient_id, encounter_id, performer_id,
    effective_datetime,
    value_quantity, value_unit, value_system, value_code,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000402'::uuid,
    'obs-john-bmi',
    'final', 'vital-signs', 'Vital Signs',
    'http://loinc.org', '39156-5', 'Body mass index (BMI) [Ratio]',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'a0000000-0000-4000-8000-000000000011'::uuid,
    '2025-01-15 09:12:00-08',
    24.5000, 'kg/m2', 'http://unitsofmeasure.org', 'kg/m2',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Hemoglobin A1c (laboratory)
INSERT INTO observation (
    id, fhir_id, status, category_code, category_display,
    code_system, code_value, code_display,
    patient_id, encounter_id, performer_id,
    effective_datetime,
    value_quantity, value_unit, value_system, value_code,
    reference_range_low, reference_range_high, reference_range_unit,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000403'::uuid,
    'obs-john-hba1c',
    'final', 'laboratory', 'Laboratory',
    'http://loinc.org', '4548-4', 'Hemoglobin A1c/Hemoglobin.total in Blood',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    '2025-01-15 09:15:00-08',
    6.8000, '%', 'http://unitsofmeasure.org', '%',
    4.0000, 5.6000, '%',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Smoking Status (social-history)
INSERT INTO observation (
    id, fhir_id, status, category_code, category_display,
    code_system, code_value, code_display,
    patient_id, encounter_id, performer_id,
    effective_datetime,
    value_codeable_code, value_codeable_display,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000404'::uuid,
    'obs-john-smoking',
    'final', 'social-history', 'Social History',
    'http://loinc.org', '72166-2', 'Tobacco smoking status',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    '2025-01-15 09:20:00-08',
    '8517006', 'Former smoker',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ----- Patient B (Maria Garcia) -----

-- Blood Pressure (panel observation, with components)
INSERT INTO observation (
    id, fhir_id, status, category_code, category_display,
    code_system, code_value, code_display,
    patient_id, encounter_id, performer_id,
    effective_datetime,
    has_member,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000405'::uuid,
    'obs-maria-bp',
    'final', 'vital-signs', 'Vital Signs',
    'http://loinc.org', '85354-9', 'Blood pressure panel with all children optional',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    'a0000000-0000-4000-8000-000000000011'::uuid,
    '2025-02-10 10:10:00-08',
    FALSE,
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- BP: Systolic component
INSERT INTO observation_component (
    id, observation_id,
    code_system, code_value, code_display,
    value_quantity, value_unit
) VALUES (
    'a0000000-0000-4000-8000-000000000412'::uuid,
    'a0000000-0000-4000-8000-000000000405'::uuid,
    'http://loinc.org', '8480-6', 'Systolic blood pressure',
    130.0000, 'mmHg'
) ON CONFLICT DO NOTHING;

-- BP: Diastolic component
INSERT INTO observation_component (
    id, observation_id,
    code_system, code_value, code_display,
    value_quantity, value_unit
) VALUES (
    'a0000000-0000-4000-8000-000000000413'::uuid,
    'a0000000-0000-4000-8000-000000000405'::uuid,
    'http://loinc.org', '8462-4', 'Diastolic blood pressure',
    85.0000, 'mmHg'
) ON CONFLICT DO NOTHING;

-- Heart Rate
INSERT INTO observation (
    id, fhir_id, status, category_code, category_display,
    code_system, code_value, code_display,
    patient_id, encounter_id, performer_id,
    effective_datetime,
    value_quantity, value_unit, value_system, value_code,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000406'::uuid,
    'obs-maria-heart-rate',
    'final', 'vital-signs', 'Vital Signs',
    'http://loinc.org', '8867-4', 'Heart rate',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    'a0000000-0000-4000-8000-000000000011'::uuid,
    '2025-02-10 10:10:00-08',
    68.0000, '/min', 'http://unitsofmeasure.org', '/min',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Glucose (laboratory)
INSERT INTO observation (
    id, fhir_id, status, category_code, category_display,
    code_system, code_value, code_display,
    patient_id, encounter_id, performer_id,
    effective_datetime,
    value_quantity, value_unit, value_system, value_code,
    reference_range_low, reference_range_high, reference_range_unit,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000407'::uuid,
    'obs-maria-glucose',
    'final', 'laboratory', 'Laboratory',
    'http://loinc.org', '2345-7', 'Glucose [Mass/volume] in Serum or Plasma',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    '2025-02-10 10:20:00-08',
    95.0000, 'mg/dL', 'http://unitsofmeasure.org', 'mg/dL',
    70.0000, 100.0000, 'mg/dL',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 9. ALLERGY INTOLERANCE
-- ============================================================================

-- Patient A: Penicillin allergy
INSERT INTO allergy_intolerance (
    id, fhir_id, patient_id, encounter_id, recorder_id, asserter_id,
    clinical_status, verification_status,
    type, category, criticality,
    code_system, code_value, code_display,
    onset_datetime, recorded_date,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000500'::uuid,
    'allergy-john-penicillin',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'active', 'confirmed',
    'allergy', '{medication}', 'high',
    'http://www.nlm.nih.gov/research/umls/rxnorm', '7980', 'Penicillin G',
    '2010-05-01 00:00:00-08', '2010-05-01 00:00:00-08',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Penicillin allergy reaction
INSERT INTO allergy_reaction (
    id, allergy_id,
    manifestation_code, manifestation_display,
    severity
) VALUES (
    'a0000000-0000-4000-8000-000000000510'::uuid,
    'a0000000-0000-4000-8000-000000000500'::uuid,
    '271807003', 'Skin rash',
    'moderate'
) ON CONFLICT DO NOTHING;

-- Patient B: Latex allergy
INSERT INTO allergy_intolerance (
    id, fhir_id, patient_id, encounter_id, recorder_id, asserter_id,
    clinical_status, verification_status,
    type, category, criticality,
    code_system, code_value, code_display,
    onset_datetime, recorded_date,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000501'::uuid,
    'allergy-maria-latex',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'active', 'confirmed',
    'allergy', '{environment}', 'low',
    'http://snomed.info/sct', '111088007', 'Latex',
    '2019-11-20 00:00:00-08', '2019-11-20 00:00:00-08',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Latex allergy reaction
INSERT INTO allergy_reaction (
    id, allergy_id,
    manifestation_code, manifestation_display,
    severity
) VALUES (
    'a0000000-0000-4000-8000-000000000511'::uuid,
    'a0000000-0000-4000-8000-000000000501'::uuid,
    '247472004', 'Urticaria',
    'mild'
) ON CONFLICT DO NOTHING;

-- ============================================================================
-- 10. PROCEDURES
-- ============================================================================

-- Patient A: Blood draw (venipuncture)
INSERT INTO procedure_record (
    id, fhir_id, status, patient_id, encounter_id, recorder_id,
    code_system, code_value, code_display,
    performed_datetime,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000600'::uuid,
    'procedure-john-blood-draw',
    'completed',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'a0000000-0000-4000-8000-000000000011'::uuid,
    'http://snomed.info/sct', '82078001', 'Collection of blood specimen for laboratory',
    '2025-01-15 09:05:00-08',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Patient B: Electrocardiogram
INSERT INTO procedure_record (
    id, fhir_id, status, patient_id, encounter_id, recorder_id,
    code_system, code_value, code_display,
    performed_datetime,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000601'::uuid,
    'procedure-maria-ecg',
    'completed',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    'a0000000-0000-4000-8000-000000000011'::uuid,
    'http://snomed.info/sct', '29303009', 'Electrocardiographic procedure',
    '2025-02-10 10:15:00-08',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 11. MEDICATIONS (drug catalog entries)
-- ============================================================================

-- Metformin 500mg tablet
INSERT INTO medication (
    id, fhir_id,
    code_system, code_value, code_display,
    status, form_code, form_display,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000700'::uuid,
    'medication-metformin',
    'http://www.nlm.nih.gov/research/umls/rxnorm', '860975', 'Metformin hydrochloride 500 MG Oral Tablet',
    'active', 'tablet', 'Tablet',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Lisinopril 10mg tablet
INSERT INTO medication (
    id, fhir_id,
    code_system, code_value, code_display,
    status, form_code, form_display,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000701'::uuid,
    'medication-lisinopril',
    'http://www.nlm.nih.gov/research/umls/rxnorm', '314076', 'Lisinopril 10 MG Oral Tablet',
    'active', 'tablet', 'Tablet',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 12. MEDICATION REQUESTS
-- ============================================================================

-- Patient A: Metformin prescription
INSERT INTO medication_request (
    id, fhir_id,
    status, intent, category_code, category_display, priority,
    medication_id, patient_id, encounter_id, requester_id,
    reason_condition_id,
    dosage_text, dosage_timing_code, dosage_timing_display,
    dosage_route_code, dosage_route_display,
    dose_quantity, dose_unit,
    quantity_value, quantity_unit, days_supply, refills_allowed,
    authored_on,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000710'::uuid,
    'medicationrequest-john-metformin',
    'active', 'order', 'outpatient', 'Outpatient', 'routine',
    'a0000000-0000-4000-8000-000000000700'::uuid,
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000300'::uuid,
    'Take 1 tablet by mouth twice daily with meals',
    'BID', 'Twice daily',
    'PO', 'Oral',
    1.0000, 'tablet',
    60.0000, 'tablet', 30, 3,
    '2025-01-15 09:25:00-08',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Patient B: Lisinopril prescription
INSERT INTO medication_request (
    id, fhir_id,
    status, intent, category_code, category_display, priority,
    medication_id, patient_id, encounter_id, requester_id,
    reason_condition_id,
    dosage_text, dosage_timing_code, dosage_timing_display,
    dosage_route_code, dosage_route_display,
    dose_quantity, dose_unit,
    quantity_value, quantity_unit, days_supply, refills_allowed,
    authored_on,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000711'::uuid,
    'medicationrequest-maria-lisinopril',
    'active', 'order', 'outpatient', 'Outpatient', 'routine',
    'a0000000-0000-4000-8000-000000000701'::uuid,
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000302'::uuid,
    'Take 1 tablet by mouth once daily',
    'QD', 'Once daily',
    'PO', 'Oral',
    1.0000, 'tablet',
    30.0000, 'tablet', 30, 5,
    '2025-02-10 10:40:00-08',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 13. IMMUNIZATIONS
-- ============================================================================

-- Patient A: Influenza vaccine
INSERT INTO immunization (
    id, fhir_id, status,
    patient_id, encounter_id,
    vaccine_code_system, vaccine_code, vaccine_display,
    occurrence_datetime,
    primary_source,
    lot_number, site_code, site_display,
    route_code, route_display,
    dose_quantity, dose_unit,
    performer_id
) VALUES (
    'a0000000-0000-4000-8000-000000000800'::uuid,
    'immunization-john-influenza',
    'completed',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'http://hl7.org/fhir/sid/cvx', '158', 'influenza, injectable, quadrivalent, contains preservative',
    '2025-01-15 09:25:00-08',
    TRUE,
    'LOT-2025-FLU-001', 'LA', 'Left arm',
    'IM', 'Intramuscular',
    0.50, 'mL',
    'a0000000-0000-4000-8000-000000000011'::uuid
) ON CONFLICT (fhir_id) DO NOTHING;

-- Patient B: COVID-19 vaccine
INSERT INTO immunization (
    id, fhir_id, status,
    patient_id, encounter_id,
    vaccine_code_system, vaccine_code, vaccine_display,
    occurrence_datetime,
    primary_source,
    lot_number, site_code, site_display,
    route_code, route_display,
    dose_quantity, dose_unit,
    performer_id
) VALUES (
    'a0000000-0000-4000-8000-000000000801'::uuid,
    'immunization-maria-covid19',
    'completed',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    'http://hl7.org/fhir/sid/cvx', '213', 'SARS-COV-2 (COVID-19) vaccine, mRNA, spike protein, LNP, preservative free, 30 mcg/0.3mL dose, tris-sucrose formulation',
    '2025-02-10 10:30:00-08',
    TRUE,
    'LOT-2025-COV-042', 'LA', 'Left arm',
    'IM', 'Intramuscular',
    0.30, 'mL',
    'a0000000-0000-4000-8000-000000000011'::uuid
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 14. DIAGNOSTIC REPORTS
-- ============================================================================

-- Patient A: Metabolic panel report (includes HbA1c)
INSERT INTO diagnostic_report (
    id, fhir_id, patient_id, encounter_id, performer_id,
    status,
    category_code, category_display,
    code_system, code_value, code_display,
    effective_datetime, issued,
    conclusion,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000900'::uuid,
    'diagnosticreport-john-metabolic',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'final',
    'LAB', 'Laboratory',
    'http://loinc.org', '51990-0', 'Basic metabolic panel - Blood',
    '2025-01-15 09:15:00-08', '2025-01-15 11:00:00-08',
    'HbA1c elevated at 6.8%, consistent with controlled type 2 diabetes.',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Link HbA1c observation to diagnostic report
INSERT INTO diagnostic_report_result (diagnostic_report_id, observation_id)
VALUES (
    'a0000000-0000-4000-8000-000000000900'::uuid,
    'a0000000-0000-4000-8000-000000000403'::uuid
) ON CONFLICT DO NOTHING;

-- Patient B: Basic metabolic panel (includes glucose)
INSERT INTO diagnostic_report (
    id, fhir_id, patient_id, encounter_id, performer_id,
    status,
    category_code, category_display,
    code_system, code_value, code_display,
    effective_datetime, issued,
    conclusion,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000000901'::uuid,
    'diagnosticreport-maria-metabolic',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'final',
    'LAB', 'Laboratory',
    'http://loinc.org', '51990-0', 'Basic metabolic panel - Blood',
    '2025-02-10 10:20:00-08', '2025-02-10 12:30:00-08',
    'Glucose within normal limits at 95 mg/dL.',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Link glucose observation to diagnostic report
INSERT INTO diagnostic_report_result (diagnostic_report_id, observation_id)
VALUES (
    'a0000000-0000-4000-8000-000000000901'::uuid,
    'a0000000-0000-4000-8000-000000000407'::uuid
) ON CONFLICT DO NOTHING;

-- ============================================================================
-- 15. DOCUMENT REFERENCES
-- ============================================================================

-- Patient A: Clinical summary document
INSERT INTO document_reference (
    id, fhir_id, status, doc_status,
    type_code, type_display,
    category_code, category_display,
    patient_id, author_id, custodian_id, encounter_id,
    date, description,
    content_type, content_url, content_title,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000001000'::uuid,
    'documentreference-john-summary',
    'current', 'final',
    '34133-9', 'Summary of episode note',
    'clinical-note', 'Clinical Note',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000001'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    '2025-01-15 11:00:00-08',
    'Clinical visit summary for diabetes follow-up on 2025-01-15',
    'text/plain',
    'Binary/doc-john-summary',
    'Visit Summary - John Smith - 2025-01-15',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- Patient B: Clinical summary document
INSERT INTO document_reference (
    id, fhir_id, status, doc_status,
    type_code, type_display,
    category_code, category_display,
    patient_id, author_id, custodian_id, encounter_id,
    date, description,
    content_type, content_url, content_title,
    version_id
) VALUES (
    'a0000000-0000-4000-8000-000000001001'::uuid,
    'documentreference-maria-summary',
    'current', 'final',
    '34133-9', 'Summary of episode note',
    'clinical-note', 'Clinical Note',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'a0000000-0000-4000-8000-000000000001'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    '2025-02-10 13:00:00-08',
    'Clinical visit summary for annual physical on 2025-02-10',
    'text/plain',
    'Binary/doc-maria-summary',
    'Visit Summary - Maria Garcia - 2025-02-10',
    1
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 16. CARE PLANS
-- ============================================================================

-- Patient A: Diabetes management care plan
INSERT INTO care_plan (
    id, fhir_id, status, intent,
    category_code, category_display,
    title, description,
    patient_id, encounter_id,
    period_start, period_end,
    author_id,
    note
) VALUES (
    'a0000000-0000-4000-8000-000000001100'::uuid,
    'careplan-john-diabetes',
    'active', 'plan',
    'assess-plan', 'Assessment and Plan of Treatment',
    'Diabetes Management Plan',
    'Ongoing management plan for Type 2 Diabetes including medication, diet, and exercise.',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'a0000000-0000-4000-8000-000000000200'::uuid,
    '2025-01-15 00:00:00-08', '2025-07-15 00:00:00-08',
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'Continue Metformin 500mg BID. Dietary counseling for carbohydrate counting. Follow-up HbA1c in 3 months.'
) ON CONFLICT (fhir_id) DO NOTHING;

-- Patient B: Hypertension management care plan
INSERT INTO care_plan (
    id, fhir_id, status, intent,
    category_code, category_display,
    title, description,
    patient_id, encounter_id,
    period_start, period_end,
    author_id,
    note
) VALUES (
    'a0000000-0000-4000-8000-000000001101'::uuid,
    'careplan-maria-hypertension',
    'active', 'plan',
    'assess-plan', 'Assessment and Plan of Treatment',
    'Hypertension Management Plan',
    'Ongoing management plan for essential hypertension including medication and lifestyle modifications.',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'a0000000-0000-4000-8000-000000000201'::uuid,
    '2025-02-10 00:00:00-08', '2025-08-10 00:00:00-08',
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'Continue Lisinopril 10mg daily. DASH diet recommended. Recheck BP in 4 weeks.'
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 17. GOALS
-- ============================================================================

-- Patient A: HbA1c target goal
INSERT INTO goal (
    id, fhir_id, lifecycle_status,
    achievement_status,
    category_code, category_display,
    description,
    patient_id,
    target_measure, target_detail_string, target_due_date,
    expressed_by_id,
    note
) VALUES (
    'a0000000-0000-4000-8000-000000001200'::uuid,
    'goal-john-hba1c',
    'active',
    'in-progress',
    'safety', 'Safety',
    'Maintain HbA1c below 7.0%',
    'a0000000-0000-4000-8000-000000000100'::uuid,
    'HbA1c', 'Less than 7.0%', '2025-07-15',
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'Current HbA1c is 6.8%. Target is to keep below 7.0% with diet, exercise, and Metformin.'
) ON CONFLICT (fhir_id) DO NOTHING;

-- Patient B: Blood pressure target goal
INSERT INTO goal (
    id, fhir_id, lifecycle_status,
    achievement_status,
    category_code, category_display,
    description,
    patient_id,
    target_measure, target_detail_string, target_due_date,
    expressed_by_id,
    note
) VALUES (
    'a0000000-0000-4000-8000-000000001201'::uuid,
    'goal-maria-bp',
    'active',
    'in-progress',
    'safety', 'Safety',
    'Achieve blood pressure below 130/80 mmHg',
    'a0000000-0000-4000-8000-000000000101'::uuid,
    'Blood Pressure', 'Below 130/80 mmHg', '2025-06-10',
    'a0000000-0000-4000-8000-000000000010'::uuid,
    'Current BP is 130/85. Target is below 130/80 with Lisinopril and lifestyle changes.'
) ON CONFLICT (fhir_id) DO NOTHING;

-- ============================================================================
-- 18. ENCOUNTER DIAGNOSIS LINKS
-- ============================================================================

-- Link encounter-diagnosis conditions to encounters
INSERT INTO encounter_diagnosis (
    id, encounter_id, condition_id, use_code, rank
) VALUES
    -- John headache
    ('a0000000-0000-4000-8000-000000001300'::uuid,
     'a0000000-0000-4000-8000-000000000200'::uuid,
     'a0000000-0000-4000-8000-000000000301'::uuid,
     'AD', 1),
    -- Maria allergic rhinitis
    ('a0000000-0000-4000-8000-000000001301'::uuid,
     'a0000000-0000-4000-8000-000000000201'::uuid,
     'a0000000-0000-4000-8000-000000000303'::uuid,
     'AD', 1)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 19. ENCOUNTER PARTICIPANTS
-- ============================================================================

INSERT INTO encounter_participant (
    id, encounter_id, practitioner_id,
    type_code, type_display,
    period_start, period_end
) VALUES
    -- John encounter: Dr. Smith as attending
    ('a0000000-0000-4000-8000-000000001400'::uuid,
     'a0000000-0000-4000-8000-000000000200'::uuid,
     'a0000000-0000-4000-8000-000000000010'::uuid,
     'ATND', 'attender',
     '2025-01-15 09:00:00-08', '2025-01-15 09:30:00-08'),
    -- John encounter: Nurse Johnson
    ('a0000000-0000-4000-8000-000000001401'::uuid,
     'a0000000-0000-4000-8000-000000000200'::uuid,
     'a0000000-0000-4000-8000-000000000011'::uuid,
     'PPRF', 'primary performer',
     '2025-01-15 09:00:00-08', '2025-01-15 09:30:00-08'),
    -- Maria encounter: Dr. Smith as attending
    ('a0000000-0000-4000-8000-000000001402'::uuid,
     'a0000000-0000-4000-8000-000000000201'::uuid,
     'a0000000-0000-4000-8000-000000000010'::uuid,
     'ATND', 'attender',
     '2025-02-10 10:00:00-08', '2025-02-10 10:45:00-08'),
    -- Maria encounter: Nurse Johnson
    ('a0000000-0000-4000-8000-000000001403'::uuid,
     'a0000000-0000-4000-8000-000000000201'::uuid,
     'a0000000-0000-4000-8000-000000000011'::uuid,
     'PPRF', 'primary performer',
     '2025-02-10 10:00:00-08', '2025-02-10 10:45:00-08')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 20. PROCEDURE PERFORMERS
-- ============================================================================

INSERT INTO procedure_performer (
    id, procedure_id, practitioner_id,
    role_code, role_display, organization_id
) VALUES
    -- Blood draw by Nurse Johnson
    ('a0000000-0000-4000-8000-000000001500'::uuid,
     'a0000000-0000-4000-8000-000000000600'::uuid,
     'a0000000-0000-4000-8000-000000000011'::uuid,
     'performer', 'Performer',
     'a0000000-0000-4000-8000-000000000001'::uuid),
    -- ECG by Nurse Johnson
    ('a0000000-0000-4000-8000-000000001501'::uuid,
     'a0000000-0000-4000-8000-000000000601'::uuid,
     'a0000000-0000-4000-8000-000000000011'::uuid,
     'performer', 'Performer',
     'a0000000-0000-4000-8000-000000000001'::uuid)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 21. IMPLANTABLE DEVICES (required by US Core / Inferno)
-- ============================================================================

INSERT INTO device (
    id, fhir_id, status, manufacturer_name, model_number,
    device_name, device_name_type, type_code, type_display, type_system,
    serial_number, udi_carrier, patient_id,
    version_id, created_at, updated_at
) VALUES
    -- Cardiac pacemaker for John Smith
    ('a0000000-0000-4000-8000-000000001700'::uuid,
     'device-pacemaker-john', 'active',
     'Medtronic', 'Azure XT DR MRI',
     'Cardiac Pacemaker', 'user-friendly-name',
     '14106009', 'Cardiac pacemaker, device', 'http://snomed.info/sct',
     'SN-12345678', '(01)00844588003288(17)141120(21)SN-12345678',
     'a0000000-0000-4000-8000-000000000100'::uuid,
     1, NOW(), NOW()),
    -- Knee prosthesis for Maria Garcia
    ('a0000000-0000-4000-8000-000000001701'::uuid,
     'device-knee-prosthesis-maria', 'active',
     'Zimmer Biomet', 'NexGen Complete Knee',
     'Knee joint prosthesis', 'user-friendly-name',
     '303533002', 'Knee joint prosthesis', 'http://snomed.info/sct',
     'SN-87654321', '(01)00844588003295(17)220315(21)SN-87654321',
     'a0000000-0000-4000-8000-000000000101'::uuid,
     1, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 22. CARE TEAMS (required by US Core / Inferno)
-- ============================================================================

INSERT INTO care_team (
    id, fhir_id, status, name, patient_id, encounter_id,
    category_code, category_display,
    period_start, managing_organization_id,
    version_id, created_at, updated_at
) VALUES
    -- Diabetes care team for John Smith
    ('a0000000-0000-4000-8000-000000001800'::uuid,
     'careteam-john-diabetes', 'active', 'Diabetes Care Team',
     'a0000000-0000-4000-8000-000000000100'::uuid,
     'a0000000-0000-4000-8000-000000000200'::uuid,
     'longitudinal-care', 'Longitudinal Care Coordination',
     '2024-01-15', 'a0000000-0000-4000-8000-000000000001'::uuid,
     1, NOW(), NOW()),
    -- Primary care team for Maria Garcia
    ('a0000000-0000-4000-8000-000000001801'::uuid,
     'careteam-maria-primary', 'active', 'Primary Care Team',
     'a0000000-0000-4000-8000-000000000101'::uuid,
     'a0000000-0000-4000-8000-000000000201'::uuid,
     'longitudinal-care', 'Longitudinal Care Coordination',
     '2024-02-20', 'a0000000-0000-4000-8000-000000000001'::uuid,
     1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO care_team_participant (
    id, care_team_id, member_id, member_type,
    role_code, role_display, period_start
) VALUES
    -- Dr. Smith on John's team
    ('a0000000-0000-4000-8000-000000001810'::uuid,
     'a0000000-0000-4000-8000-000000001800'::uuid,
     'a0000000-0000-4000-8000-000000000010'::uuid,
     'Practitioner', '223366009', 'Healthcare professional',
     '2024-01-15'),
    -- Nurse Johnson on John's team
    ('a0000000-0000-4000-8000-000000001811'::uuid,
     'a0000000-0000-4000-8000-000000001800'::uuid,
     'a0000000-0000-4000-8000-000000000011'::uuid,
     'Practitioner', '224535009', 'Registered nurse',
     '2024-01-15'),
    -- Dr. Smith on Maria's team
    ('a0000000-0000-4000-8000-000000001812'::uuid,
     'a0000000-0000-4000-8000-000000001801'::uuid,
     'a0000000-0000-4000-8000-000000000010'::uuid,
     'Practitioner', '223366009', 'Healthcare professional',
     '2024-02-20'),
    -- Nurse Johnson on Maria's team
    ('a0000000-0000-4000-8000-000000001813'::uuid,
     'a0000000-0000-4000-8000-000000001801'::uuid,
     'a0000000-0000-4000-8000-000000000011'::uuid,
     'Practitioner', '224535009', 'Registered nurse',
     '2024-02-20')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 23. PROVENANCE (required by Inferno _revinclude=Provenance:target)
-- ============================================================================

INSERT INTO provenance (
    id, fhir_id, target_type, target_id,
    recorded, activity_code, activity_display,
    version_id, created_at, updated_at
) VALUES
    -- Provenance for Patient John Smith
    ('a0000000-0000-4000-8000-000000001900'::uuid,
     'provenance-patient-john', 'Patient', 'patient-john-smith',
     '2024-01-15T10:00:00Z', 'CREATE', 'create',
     1, NOW(), NOW()),
    -- Provenance for Patient Maria Garcia
    ('a0000000-0000-4000-8000-000000001901'::uuid,
     'provenance-patient-maria', 'Patient', 'patient-maria-garcia',
     '2024-02-20T09:00:00Z', 'CREATE', 'create',
     1, NOW(), NOW()),
    -- Provenance for Condition diabetes
    ('a0000000-0000-4000-8000-000000001902'::uuid,
     'provenance-condition-diabetes', 'Condition', 'condition-john-diabetes',
     '2024-01-15T10:30:00Z', 'CREATE', 'create',
     1, NOW(), NOW()),
    -- Provenance for Encounter John
    ('a0000000-0000-4000-8000-000000001903'::uuid,
     'provenance-encounter-john', 'Encounter', 'encounter-john-smith-1',
     '2024-01-15T08:00:00Z', 'CREATE', 'create',
     1, NOW(), NOW()),
    -- Provenance for AllergyIntolerance penicillin
    ('a0000000-0000-4000-8000-000000001904'::uuid,
     'provenance-allergy-penicillin', 'AllergyIntolerance', 'allergy-john-penicillin',
     '2024-01-15T10:15:00Z', 'CREATE', 'create',
     1, NOW(), NOW()),
    -- Provenance for MedicationRequest metformin
    ('a0000000-0000-4000-8000-000000001905'::uuid,
     'provenance-medrx-metformin', 'MedicationRequest', 'medicationrequest-john-metformin',
     '2024-01-15T11:00:00Z', 'CREATE', 'create',
     1, NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO provenance_agent (
    id, provenance_id, type_code, type_display,
    who_type, who_id
) VALUES
    ('a0000000-0000-4000-8000-000000001910'::uuid,
     'a0000000-0000-4000-8000-000000001900'::uuid,
     'author', 'Author', 'Practitioner', 'practitioner-dr-smith'),
    ('a0000000-0000-4000-8000-000000001911'::uuid,
     'a0000000-0000-4000-8000-000000001901'::uuid,
     'author', 'Author', 'Practitioner', 'practitioner-dr-smith'),
    ('a0000000-0000-4000-8000-000000001912'::uuid,
     'a0000000-0000-4000-8000-000000001902'::uuid,
     'author', 'Author', 'Practitioner', 'practitioner-dr-smith'),
    ('a0000000-0000-4000-8000-000000001913'::uuid,
     'a0000000-0000-4000-8000-000000001903'::uuid,
     'author', 'Author', 'Practitioner', 'practitioner-dr-smith'),
    ('a0000000-0000-4000-8000-000000001914'::uuid,
     'a0000000-0000-4000-8000-000000001904'::uuid,
     'author', 'Author', 'Practitioner', 'practitioner-dr-smith'),
    ('a0000000-0000-4000-8000-000000001915'::uuid,
     'a0000000-0000-4000-8000-000000001905'::uuid,
     'author', 'Author', 'Practitioner', 'practitioner-dr-smith')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 24. PATIENT IDENTIFIERS (MRN in identifier table)
-- ============================================================================

INSERT INTO patient_identifier (
    id, patient_id,
    system_uri, value, type_code, type_display, assigner
) VALUES
    ('a0000000-0000-4000-8000-000000001600'::uuid,
     'a0000000-0000-4000-8000-000000000100'::uuid,
     'http://acmegeneral.example.com/mrn', 'MRN-001',
     'MR', 'Medical record number', 'Acme General Hospital'),
    ('a0000000-0000-4000-8000-000000001601'::uuid,
     'a0000000-0000-4000-8000-000000000101'::uuid,
     'http://acmegeneral.example.com/mrn', 'MRN-002',
     'MR', 'Medical record number', 'Acme General Hospital')
ON CONFLICT (patient_id, system_uri, value) DO NOTHING;

COMMIT;

-- ============================================================================
-- Summary of seeded Inferno g(10) test data
-- ============================================================================
-- Organization:       1  (org-acme-general-hospital)
-- Location:           1  (loc-main-campus)
-- Practitioner:       2  (practitioner-dr-smith, practitioner-nurse-johnson)
-- PractitionerRole:   2  (practitionerrole-dr-smith, practitionerrole-nurse-johnson)
-- Patient:            2  (patient-john-smith, patient-maria-garcia)
-- Encounter:          2  (encounter-john-smith-1, encounter-maria-garcia-1)
-- Condition:          4  (2 problem-list-item, 2 encounter-diagnosis)
-- Observation:        8  (3 vital-signs + 1 lab + 1 social-history for John,
--                         2 vital-signs + 1 lab for Maria)
-- ObservationComponent: 4 (systolic+diastolic for each BP)
-- AllergyIntolerance: 2  (penicillin, latex)
-- AllergyReaction:    2  (skin rash, urticaria)
-- Procedure:          2  (blood draw, ECG)
-- Medication:         2  (metformin, lisinopril)
-- MedicationRequest:  2  (metformin Rx, lisinopril Rx)
-- Immunization:       2  (influenza, COVID-19)
-- DiagnosticReport:   2  (metabolic panels)
-- DocumentReference:  2  (clinical summaries)
-- CarePlan:           2  (diabetes plan, hypertension plan)
-- Goal:               2  (HbA1c target, BP target)
-- Device:             2  (pacemaker for John, knee prosthesis for Maria)
-- CareTeam:           2  (diabetes team for John, primary care for Maria)
-- CareTeamParticipant: 4 (Dr. Smith + Nurse Johnson per team)
-- Provenance:         6  (Patient x2, Condition, Encounter, Allergy, MedRequest)
-- ProvenanceAgent:    6  (Dr. Smith as author for each)
-- EncounterDiagnosis: 2  (linking encounter-dx conditions)
-- EncounterParticipant: 4 (2 per encounter)
-- ProcedurePerformer: 2  (one per procedure)
-- PatientIdentifier:  2  (MRN identifiers)
-- ============================================================================
