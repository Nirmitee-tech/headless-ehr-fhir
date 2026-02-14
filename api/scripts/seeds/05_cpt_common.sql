-- Seed: Common CPT Codes (50+)
-- Source: CPT (Current Procedural Terminology) -- AMA
-- Note: CPT codes are copyrighted by AMA. These are generic code identifiers
-- used for reference/mapping purposes in the EHR system.

CREATE TABLE IF NOT EXISTS reference_cpt (
    code VARCHAR(10) PRIMARY KEY,
    display VARCHAR(500) NOT NULL,
    category VARCHAR(100),
    subcategory VARCHAR(100),
    system_uri VARCHAR(255) DEFAULT 'http://www.ama-assn.org/go/cpt'
);

INSERT INTO reference_cpt (code, display, category, subcategory) VALUES
-- E&M - Office/Outpatient Visit (New Patient)
('99201', 'Office visit, new patient, minimal',                             'E&M', 'Office Visit - New'),
('99202', 'Office visit, new patient, straightforward',                     'E&M', 'Office Visit - New'),
('99203', 'Office visit, new patient, low complexity',                      'E&M', 'Office Visit - New'),
('99204', 'Office visit, new patient, moderate complexity',                 'E&M', 'Office Visit - New'),
('99205', 'Office visit, new patient, high complexity',                     'E&M', 'Office Visit - New'),

-- E&M - Office/Outpatient Visit (Established Patient)
('99211', 'Office visit, established patient, minimal',                     'E&M', 'Office Visit - Established'),
('99212', 'Office visit, established patient, straightforward',             'E&M', 'Office Visit - Established'),
('99213', 'Office visit, established patient, low complexity',              'E&M', 'Office Visit - Established'),
('99214', 'Office visit, established patient, moderate complexity',         'E&M', 'Office Visit - Established'),
('99215', 'Office visit, established patient, high complexity',             'E&M', 'Office Visit - Established'),

-- E&M - Hospital Inpatient
('99221', 'Initial hospital care, straightforward/low complexity',          'E&M', 'Hospital Inpatient'),
('99222', 'Initial hospital care, moderate complexity',                     'E&M', 'Hospital Inpatient'),
('99223', 'Initial hospital care, high complexity',                         'E&M', 'Hospital Inpatient'),
('99231', 'Subsequent hospital care, straightforward',                      'E&M', 'Hospital Inpatient'),
('99232', 'Subsequent hospital care, moderate complexity',                  'E&M', 'Hospital Inpatient'),
('99233', 'Subsequent hospital care, high complexity',                      'E&M', 'Hospital Inpatient'),
('99238', 'Hospital discharge day management, 30 min or less',              'E&M', 'Hospital Inpatient'),
('99239', 'Hospital discharge day management, more than 30 min',            'E&M', 'Hospital Inpatient'),

-- E&M - Emergency Department
('99281', 'ED visit, self-limited or minor problem',                        'E&M', 'Emergency Dept'),
('99282', 'ED visit, low to moderate severity',                             'E&M', 'Emergency Dept'),
('99283', 'ED visit, moderate severity',                                    'E&M', 'Emergency Dept'),
('99284', 'ED visit, high severity without threat to life',                 'E&M', 'Emergency Dept'),
('99285', 'ED visit, high severity with threat to life',                    'E&M', 'Emergency Dept'),

-- E&M - Critical Care
('99291', 'Critical care, first 30-74 minutes',                             'E&M', 'Critical Care'),
('99292', 'Critical care, each additional 30 minutes',                      'E&M', 'Critical Care'),

-- E&M - Consultation
('99241', 'Office consultation, straightforward',                           'E&M', 'Consultation'),
('99242', 'Office consultation, straightforward',                           'E&M', 'Consultation'),
('99243', 'Office consultation, low complexity',                            'E&M', 'Consultation'),
('99244', 'Office consultation, moderate complexity',                       'E&M', 'Consultation'),
('99245', 'Office consultation, high complexity',                           'E&M', 'Consultation'),

-- E&M - Preventive / Wellness
('99385', 'Initial preventive visit, 18-39 years',                          'E&M', 'Preventive'),
('99386', 'Initial preventive visit, 40-64 years',                          'E&M', 'Preventive'),
('99387', 'Initial preventive visit, 65+ years',                            'E&M', 'Preventive'),
('99395', 'Periodic preventive visit, 18-39 years',                         'E&M', 'Preventive'),
('99396', 'Periodic preventive visit, 40-64 years',                         'E&M', 'Preventive'),
('99397', 'Periodic preventive visit, 65+ years',                           'E&M', 'Preventive'),

-- Common Procedures
('10060', 'Incision and drainage of abscess, simple',                       'Surgery',    'Integumentary'),
('12001', 'Simple repair of wound, 2.5 cm or less',                        'Surgery',    'Integumentary'),
('20610', 'Arthrocentesis, aspiration or injection, major joint',           'Surgery',    'Musculoskeletal'),
('29881', 'Arthroscopy, knee, surgical with meniscectomy',                  'Surgery',    'Musculoskeletal'),
('36415', 'Venipuncture for collection of blood specimen',                  'Pathology',  'Specimen Collection'),
('43239', 'Upper GI endoscopy with biopsy',                                'Surgery',    'Digestive'),
('45378', 'Colonoscopy, diagnostic',                                        'Surgery',    'Digestive'),
('45380', 'Colonoscopy with biopsy',                                        'Surgery',    'Digestive'),
('47562', 'Laparoscopic cholecystectomy',                                   'Surgery',    'Digestive'),
('49505', 'Repair of inguinal hernia',                                      'Surgery',    'Digestive'),
('93000', 'Electrocardiogram, routine, 12-lead',                           'Medicine',   'Cardiovascular'),
('93306', 'Echocardiography, transthoracic, complete',                     'Medicine',   'Cardiovascular'),
('71046', 'Chest X-ray, 2 views',                                          'Radiology',  'Diagnostic'),
('74177', 'CT abdomen and pelvis with contrast',                            'Radiology',  'Diagnostic'),
('70553', 'MRI brain without and with contrast',                            'Radiology',  'Diagnostic'),
('90471', 'Immunization administration, 1st vaccine',                       'Medicine',   'Immunization')
ON CONFLICT (code) DO NOTHING;
