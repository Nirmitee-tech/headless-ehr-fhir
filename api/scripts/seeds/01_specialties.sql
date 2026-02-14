-- Seed: Medical Specialties (40+)
-- Source: SNOMED CT codes for clinical specialties

INSERT INTO specialty (code, display, system_uri, category, country_applicability, active)
VALUES
    -- Medical Specialties
    ('394579002', 'Cardiology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394802001', 'General Medicine', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394585009', 'Obstetrics', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394584008', 'Gastroenterology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394537008', 'Pediatrics', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394587001', 'Psychiatry', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394596001', 'Anesthesiology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394600006', 'Dermatology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394583002', 'Endocrinology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394589003', 'Nephrology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394591006', 'Neurology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394733009', 'Emergency Medicine', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394807007', 'Pulmonology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394811001', 'Geriatric Medicine', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394812008', 'Hematology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394577000', 'Allergy and Immunology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394578005', 'Oncology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394810000', 'Rheumatology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394592004', 'Clinical Genetics', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394806003', 'Palliative Medicine', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394602003', 'Rehabilitation Medicine', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394603008', 'Infectious Disease', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394604002', 'Urology', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('394821009', 'Occupational Medicine', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394605001', 'Sports Medicine', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('394809005', 'Clinical Pharmacology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('408443003', 'Neonatology', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('418112009', 'Sleep Medicine', 'http://snomed.info/sct', 'medical', 'BOTH', true),
    ('408459003', 'Pain Medicine', 'http://snomed.info/sct', 'medical', 'BOTH', true),

    -- Surgical Specialties
    ('394609007', 'General Surgery', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('394610002', 'Neurosurgery', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('394611003', 'Plastic Surgery', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('394582007', 'ENT', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('394593009', 'Ophthalmology', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('394801008', 'Orthopedics', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('408466002', 'Cardiac Surgery', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('394612005', 'Thoracic Surgery', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('408464004', 'Colorectal Surgery', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('394613000', 'Pediatric Surgery', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('394614006', 'Vascular Surgery', 'http://snomed.info/sct', 'surgical', 'BOTH', true),
    ('408470005', 'Transplant Surgery', 'http://snomed.info/sct', 'surgical', 'BOTH', true),

    -- Diagnostic Specialties
    ('394914008', 'Radiology', 'http://snomed.info/sct', 'diagnostic', 'BOTH', true),
    ('394915009', 'Pathology', 'http://snomed.info/sct', 'diagnostic', 'BOTH', true),
    ('394916005', 'Nuclear Medicine', 'http://snomed.info/sct', 'diagnostic', 'BOTH', true),

    -- Allied Health
    ('394601005', 'Clinical Psychology', 'http://snomed.info/sct', 'allied_health', 'BOTH', true),
    ('310101009', 'Audiology', 'http://snomed.info/sct', 'allied_health', 'BOTH', true),
    ('310080006', 'Pharmacy', 'http://snomed.info/sct', 'allied_health', 'BOTH', true),
    ('3842006',   'Chiropractic', 'http://snomed.info/sct', 'allied_health', 'US', true),
    ('394803006', 'Family Medicine', 'http://snomed.info/sct', 'medical', 'US', true),
    ('394804000', 'Internal Medicine', 'http://snomed.info/sct', 'medical', 'US', true),

    -- India AYUSH
    ('AYUSH001', 'Ayurveda', 'http://nrces.in/ndhm/fhir/r4/CodeSystem/ndhm-ayush-specialty', 'ayush', 'IN', true),
    ('AYUSH002', 'Unani', 'http://nrces.in/ndhm/fhir/r4/CodeSystem/ndhm-ayush-specialty', 'ayush', 'IN', true),
    ('AYUSH003', 'Siddha', 'http://nrces.in/ndhm/fhir/r4/CodeSystem/ndhm-ayush-specialty', 'ayush', 'IN', true),
    ('AYUSH004', 'Homeopathy', 'http://nrces.in/ndhm/fhir/r4/CodeSystem/ndhm-ayush-specialty', 'ayush', 'IN', true),
    ('AYUSH005', 'Yoga and Naturopathy', 'http://nrces.in/ndhm/fhir/r4/CodeSystem/ndhm-ayush-specialty', 'ayush', 'IN', true)
ON CONFLICT (code) DO NOTHING;
