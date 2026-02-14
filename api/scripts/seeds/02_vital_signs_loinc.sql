-- Seed: LOINC Reference Codes (vital signs + common labs)
-- Source: LOINC (Logical Observation Identifiers Names and Codes)

CREATE TABLE IF NOT EXISTS reference_loinc (
    code VARCHAR(20) PRIMARY KEY,
    display VARCHAR(255) NOT NULL,
    component VARCHAR(100),
    property VARCHAR(50),
    time_aspect VARCHAR(20),
    system_uri VARCHAR(255) DEFAULT 'http://loinc.org',
    category VARCHAR(50)
);

INSERT INTO reference_loinc (code, display, component, property, time_aspect, category) VALUES
-- Vital Signs
('8310-5',  'Body temperature',                         'Body temperature',     'Temp',   'Pt', 'vital-signs'),
('8867-4',  'Heart rate',                               'Heart rate',           'NRat',   'Pt', 'vital-signs'),
('9279-1',  'Respiratory rate',                         'Respiratory rate',     'NRat',   'Pt', 'vital-signs'),
('85354-9', 'Blood pressure panel',                     'Blood pressure',       'Pres',   'Pt', 'vital-signs'),
('8480-6',  'Systolic blood pressure',                  'Systolic BP',          'Pres',   'Pt', 'vital-signs'),
('8462-4',  'Diastolic blood pressure',                 'Diastolic BP',         'Pres',   'Pt', 'vital-signs'),
('2708-6',  'Oxygen saturation in arterial blood',      'SpO2',                 'SFr',    'Pt', 'vital-signs'),
('59408-5', 'Oxygen saturation by pulse oximetry',      'SpO2 Pulse Ox',        'SFr',    'Pt', 'vital-signs'),
('29463-7', 'Body weight',                              'Body weight',          'Mass',   'Pt', 'vital-signs'),
('8302-2',  'Body height',                              'Body height',          'Len',    'Pt', 'vital-signs'),
('39156-5', 'Body mass index',                          'BMI',                  'RelMRat','Pt', 'vital-signs'),
('8287-5',  'Head circumference',                       'Head circ',            'Len',    'Pt', 'vital-signs'),
('3141-9',  'Body weight measured',                     'Body weight measured',  'Mass',   'Pt', 'vital-signs'),
('8328-7',  'Axillary temperature',                     'Temp axillary',        'Temp',   'Pt', 'vital-signs'),

-- CBC - Complete Blood Count
('6690-2',  'WBC count',                                'Leukocytes',           'NCnc',   'Pt', 'laboratory'),
('789-8',   'RBC count',                                'Erythrocytes',         'NCnc',   'Pt', 'laboratory'),
('718-7',   'Hemoglobin',                               'Hemoglobin',           'MCnc',   'Pt', 'laboratory'),
('4544-3',  'Hematocrit',                               'Hematocrit',           'VFr',    'Pt', 'laboratory'),
('787-2',   'MCV',                                      'MCV',                  'EntVol', 'Pt', 'laboratory'),
('785-6',   'MCH',                                      'MCH',                  'EntMass','Pt', 'laboratory'),
('786-4',   'MCHC',                                     'MCHC',                 'MCnc',   'Pt', 'laboratory'),
('788-0',   'RDW',                                      'RDW',                  'RelMCnc','Pt', 'laboratory'),
('777-3',   'Platelet count',                           'Platelets',            'NCnc',   'Pt', 'laboratory'),
('770-8',   'Neutrophils %',                            'Neutrophils',          'NFr',    'Pt', 'laboratory'),
('736-9',   'Lymphocytes %',                            'Lymphocytes',          'NFr',    'Pt', 'laboratory'),
('5905-5',  'Monocytes %',                              'Monocytes',            'NFr',    'Pt', 'laboratory'),
('713-8',   'Eosinophils %',                            'Eosinophils',          'NFr',    'Pt', 'laboratory'),
('706-2',   'Basophils %',                              'Basophils',            'NFr',    'Pt', 'laboratory'),

-- BMP - Basic Metabolic Panel
('2345-7',  'Glucose',                                  'Glucose',              'MCnc',   'Pt', 'laboratory'),
('3094-0',  'BUN (Blood urea nitrogen)',                'Urea nitrogen',        'MCnc',   'Pt', 'laboratory'),
('2160-0',  'Creatinine',                               'Creatinine',           'MCnc',   'Pt', 'laboratory'),
('2951-2',  'Sodium',                                   'Sodium',               'SCnc',   'Pt', 'laboratory'),
('2823-3',  'Potassium',                                'Potassium',            'SCnc',   'Pt', 'laboratory'),
('2075-0',  'Chloride',                                 'Chloride',             'SCnc',   'Pt', 'laboratory'),
('1963-8',  'Bicarbonate (CO2)',                        'Bicarbonate',          'SCnc',   'Pt', 'laboratory'),
('17861-6', 'Calcium',                                  'Calcium',              'MCnc',   'Pt', 'laboratory'),

-- CMP additions (beyond BMP)
('1751-7',  'Albumin',                                  'Albumin',              'MCnc',   'Pt', 'laboratory'),
('2885-2',  'Total protein',                            'Protein',              'MCnc',   'Pt', 'laboratory'),
('1975-2',  'Total bilirubin',                          'Bilirubin',            'MCnc',   'Pt', 'laboratory'),
('6768-6',  'Alkaline phosphatase',                     'ALP',                  'CCnc',   'Pt', 'laboratory'),
('1742-6',  'ALT (SGPT)',                               'ALT',                  'CCnc',   'Pt', 'laboratory'),
('1920-8',  'AST (SGOT)',                               'AST',                  'CCnc',   'Pt', 'laboratory'),
('33914-3', 'eGFR',                                     'eGFR',                 'ArVRat', 'Pt', 'laboratory'),

-- Lipid Panel
('2093-3',  'Total cholesterol',                        'Cholesterol',          'MCnc',   'Pt', 'laboratory'),
('2571-8',  'Triglycerides',                            'Triglycerides',        'MCnc',   'Pt', 'laboratory'),
('2085-9',  'HDL cholesterol',                          'HDL',                  'MCnc',   'Pt', 'laboratory'),
('2089-1',  'LDL cholesterol',                          'LDL',                  'MCnc',   'Pt', 'laboratory'),
('13457-7', 'LDL cholesterol (calculated)',             'LDL calc',             'MCnc',   'Pt', 'laboratory'),

-- Endocrine / Metabolic
('4548-4',  'Hemoglobin A1c',                           'HbA1c',                'MFr',    'Pt', 'laboratory'),
('3016-3',  'TSH',                                      'TSH',                  'ACnc',   'Pt', 'laboratory'),
('3026-2',  'Free T4',                                  'Free T4',              'MCnc',   'Pt', 'laboratory'),
('3024-7',  'Free T3',                                  'Free T3',              'MCnc',   'Pt', 'laboratory'),
('2132-9',  'Vitamin B12',                              'Vitamin B12',          'MCnc',   'Pt', 'laboratory'),
('1989-3',  'Vitamin D (25-OH)',                        'Vitamin D',            'MCnc',   'Pt', 'laboratory'),
('2284-8',  'Folate',                                   'Folate',               'MCnc',   'Pt', 'laboratory'),

-- Coagulation
('5902-2',  'Prothrombin time (PT)',                    'PT',                   'Time',   'Pt', 'laboratory'),
('6301-6',  'INR',                                      'INR',                  'RelTime','Pt', 'laboratory'),
('3173-2',  'aPTT',                                     'aPTT',                 'Time',   'Pt', 'laboratory'),
('3255-7',  'Fibrinogen',                               'Fibrinogen',           'MCnc',   'Pt', 'laboratory'),
('48065-7', 'D-dimer',                                  'D-dimer',              'MCnc',   'Pt', 'laboratory'),

-- Cardiac Markers
('2157-6',  'CK (creatine kinase)',                     'CK',                   'CCnc',   'Pt', 'laboratory'),
('49563-0', 'Troponin I (high sensitivity)',            'hs-Troponin I',        'MCnc',   'Pt', 'laboratory'),
('89579-7', 'Troponin T (high sensitivity)',            'hs-Troponin T',        'MCnc',   'Pt', 'laboratory'),
('30934-4', 'BNP',                                      'BNP',                  'MCnc',   'Pt', 'laboratory'),
('33762-6', 'NT-proBNP',                                'NT-proBNP',            'MCnc',   'Pt', 'laboratory'),

-- Inflammatory Markers
('1988-5',  'C-reactive protein (CRP)',                 'CRP',                  'MCnc',   'Pt', 'laboratory'),
('30522-7', 'High sensitivity CRP',                    'hs-CRP',               'MCnc',   'Pt', 'laboratory'),
('4537-7',  'ESR (sed rate)',                           'ESR',                  'Vel',    'Pt', 'laboratory'),
('33959-8', 'Procalcitonin',                            'Procalcitonin',        'MCnc',   'Pt', 'laboratory'),

-- Urinalysis
('5811-5',  'Urine specific gravity',                   'Sp Gravity Urine',    'Rden',   'Pt', 'laboratory'),
('5803-2',  'Urine pH',                                 'pH Urine',             'LsCnc',  'Pt', 'laboratory'),
('5804-0',  'Urine protein',                            'Protein Urine',        'MCnc',   'Pt', 'laboratory'),
('5792-7',  'Urine glucose',                            'Glucose Urine',        'MCnc',   'Pt', 'laboratory'),

-- Other Common Tests
('2947-0',  'PSA',                                      'PSA',                  'MCnc',   'Pt', 'laboratory'),
('5794-3',  'Urine culture',                            'Culture Urine',        'Prid',   'Pt', 'laboratory'),
('600-7',   'Blood culture',                            'Culture Blood',        'Prid',   'Pt', 'laboratory'),
('2276-4',  'Ferritin',                                 'Ferritin',             'MCnc',   'Pt', 'laboratory'),
('2498-4',  'Iron',                                     'Iron',                 'MCnc',   'Pt', 'laboratory'),
('2502-3',  'Iron saturation',                          'TIBC',                 'MCnc',   'Pt', 'laboratory'),
('14979-9', 'aPTT (partial thromboplastin time)',       'aPTT',                 'Time',   'Pt', 'laboratory'),
('1558-6',  'Fasting glucose',                          'Fasting glucose',      'MCnc',   'Pt', 'laboratory'),
('20570-8', 'Hematocrit (venous)',                      'Hematocrit venous',    'VFr',    'Pt', 'laboratory'),
('2339-0',  'Glucose (blood, random)',                  'Glucose random',       'MCnc',   'Pt', 'laboratory')
ON CONFLICT (code) DO NOTHING;
