-- Seed: Common ICD-10-CM Diagnosis Codes (100+)
-- Source: ICD-10-CM (International Classification of Diseases, 10th Revision, Clinical Modification)

CREATE TABLE IF NOT EXISTS reference_icd10 (
    code VARCHAR(10) PRIMARY KEY,
    display VARCHAR(500) NOT NULL,
    category VARCHAR(100),
    chapter VARCHAR(10),
    system_uri VARCHAR(255) DEFAULT 'http://hl7.org/fhir/sid/icd-10-cm'
);

INSERT INTO reference_icd10 (code, display, category, chapter) VALUES
-- Infectious Diseases (A/B)
('A09',     'Infectious gastroenteritis and colitis, unspecified',        'Infectious',            'I'),
('A41.9',   'Sepsis, unspecified organism',                              'Infectious',            'I'),
('A49.9',   'Bacterial infection, unspecified',                          'Infectious',            'I'),
('B34.9',   'Viral infection, unspecified',                              'Infectious',            'I'),

-- Neoplasms (C/D)
('C34.90',  'Malignant neoplasm of unspecified part of bronchus or lung','Neoplasm',              'II'),
('C50.919', 'Malignant neoplasm of unspecified site of breast',          'Neoplasm',              'II'),
('C61',     'Malignant neoplasm of prostate',                            'Neoplasm',              'II'),
('C18.9',   'Malignant neoplasm of colon, unspecified',                  'Neoplasm',              'II'),
('D64.9',   'Anemia, unspecified',                                       'Blood',                 'III'),

-- Endocrine (E)
('E03.9',   'Hypothyroidism, unspecified',                               'Endocrine',             'IV'),
('E05.90',  'Thyrotoxicosis, unspecified without thyrotoxic crisis',     'Endocrine',             'IV'),
('E11.9',   'Type 2 diabetes mellitus without complications',            'Endocrine',             'IV'),
('E11.65',  'Type 2 diabetes mellitus with hyperglycemia',               'Endocrine',             'IV'),
('E11.40',  'Type 2 diabetes with diabetic neuropathy, unspecified',     'Endocrine',             'IV'),
('E11.319', 'Type 2 diabetes with unspecified diabetic retinopathy',     'Endocrine',             'IV'),
('E11.22',  'Type 2 diabetes with diabetic chronic kidney disease',      'Endocrine',             'IV'),
('E10.9',   'Type 1 diabetes mellitus without complications',            'Endocrine',             'IV'),
('E13.9',   'Other specified diabetes mellitus without complications',   'Endocrine',             'IV'),
('E66.01',  'Morbid (severe) obesity due to excess calories',            'Endocrine',             'IV'),
('E66.9',   'Obesity, unspecified',                                      'Endocrine',             'IV'),
('E78.5',   'Dyslipidemia, unspecified',                                 'Endocrine',             'IV'),
('E78.00',  'Pure hypercholesterolemia, unspecified',                    'Endocrine',             'IV'),
('E55.9',   'Vitamin D deficiency, unspecified',                         'Endocrine',             'IV'),
('E87.6',   'Hypokalemia',                                              'Endocrine',             'IV'),

-- Mental and behavioral (F)
('F10.20',  'Alcohol dependence, uncomplicated',                         'Mental',                'V'),
('F17.210', 'Nicotine dependence, cigarettes, uncomplicated',            'Mental',                'V'),
('F20.9',   'Schizophrenia, unspecified',                                'Mental',                'V'),
('F31.9',   'Bipolar disorder, unspecified',                             'Mental',                'V'),
('F32.9',   'Major depressive disorder, single episode, unspecified',    'Mental',                'V'),
('F33.0',   'Major depressive disorder, recurrent, mild',                'Mental',                'V'),
('F33.1',   'Major depressive disorder, recurrent, moderate',            'Mental',                'V'),
('F41.0',   'Panic disorder without agoraphobia',                        'Mental',                'V'),
('F41.1',   'Generalized anxiety disorder',                              'Mental',                'V'),
('F41.9',   'Anxiety disorder, unspecified',                             'Mental',                'V'),
('F43.10',  'Post-traumatic stress disorder, unspecified',               'Mental',                'V'),
('F90.9',   'Attention-deficit hyperactivity disorder, unspecified type', 'Mental',                'V'),

-- Nervous System (G)
('G20',     'Parkinson disease',                                         'Nervous',               'VI'),
('G30.9',   'Alzheimer disease, unspecified',                            'Nervous',               'VI'),
('G40.909', 'Epilepsy, unspecified, not intractable',                    'Nervous',               'VI'),
('G43.909', 'Migraine, unspecified, not intractable',                    'Nervous',               'VI'),
('G47.00',  'Insomnia, unspecified',                                     'Nervous',               'VI'),
('G47.33',  'Obstructive sleep apnea',                                   'Nervous',               'VI'),
('G89.29',  'Other chronic pain',                                        'Nervous',               'VI'),

-- Eye (H)
('H40.10X0','Open-angle glaucoma, unspecified eye',                      'Eye',                   'VII'),
('H26.9',   'Unspecified cataract',                                      'Eye',                   'VII'),

-- Circulatory (I)
('I10',     'Essential (primary) hypertension',                          'Circulatory',           'IX'),
('I11.9',   'Hypertensive heart disease without heart failure',          'Circulatory',           'IX'),
('I20.9',   'Angina pectoris, unspecified',                              'Circulatory',           'IX'),
('I21.9',   'Acute myocardial infarction, unspecified',                  'Circulatory',           'IX'),
('I25.10',  'Atherosclerotic heart disease of native coronary artery',   'Circulatory',           'IX'),
('I25.9',   'Chronic ischemic heart disease, unspecified',               'Circulatory',           'IX'),
('I48.91',  'Unspecified atrial fibrillation',                           'Circulatory',           'IX'),
('I48.0',   'Paroxysmal atrial fibrillation',                            'Circulatory',           'IX'),
('I50.9',   'Heart failure, unspecified',                                'Circulatory',           'IX'),
('I50.22',  'Chronic systolic (congestive) heart failure',               'Circulatory',           'IX'),
('I63.9',   'Cerebral infarction, unspecified',                          'Circulatory',           'IX'),
('I73.9',   'Peripheral vascular disease, unspecified',                  'Circulatory',           'IX'),
('I83.90',  'Asymptomatic varicose veins of unspecified lower extremity','Circulatory',           'IX'),
('I87.2',   'Venous insufficiency (chronic) (peripheral)',               'Circulatory',           'IX'),

-- Respiratory (J)
('J02.9',   'Acute pharyngitis, unspecified',                            'Respiratory',           'X'),
('J06.9',   'Acute upper respiratory infection, unspecified',            'Respiratory',           'X'),
('J18.9',   'Pneumonia, unspecified organism',                           'Respiratory',           'X'),
('J20.9',   'Acute bronchitis, unspecified',                             'Respiratory',           'X'),
('J30.9',   'Allergic rhinitis, unspecified',                            'Respiratory',           'X'),
('J44.1',   'Chronic obstructive pulmonary disease with acute exac',     'Respiratory',           'X'),
('J44.9',   'Chronic obstructive pulmonary disease, unspecified',        'Respiratory',           'X'),
('J45.20',  'Mild intermittent asthma, uncomplicated',                   'Respiratory',           'X'),
('J45.909', 'Unspecified asthma, uncomplicated',                         'Respiratory',           'X'),
('J96.00',  'Acute respiratory failure, unspecified',                    'Respiratory',           'X'),

-- Digestive (K)
('K21.0',   'Gastro-esophageal reflux disease with esophagitis',         'Digestive',             'XI'),
('K21.9',   'Gastro-esophageal reflux disease without esophagitis',      'Digestive',             'XI'),
('K25.9',   'Gastric ulcer, unspecified, without hemorrhage or perf',    'Digestive',             'XI'),
('K29.70',  'Gastritis, unspecified, without bleeding',                  'Digestive',             'XI'),
('K35.80',  'Unspecified acute appendicitis',                            'Digestive',             'XI'),
('K40.90',  'Unilateral inguinal hernia without obstruction or gangrene','Digestive',             'XI'),
('K57.90',  'Diverticulosis of intestine, unspecified',                  'Digestive',             'XI'),
('K58.9',   'Irritable bowel syndrome without diarrhea',                 'Digestive',             'XI'),
('K76.0',   'Fatty (change of) liver, not elsewhere classified',        'Digestive',             'XI'),
('K80.20',  'Calculus of gallbladder without cholecystitis, without obs','Digestive',             'XI'),

-- Skin (L)
('L03.90',  'Cellulitis, unspecified',                                   'Skin',                  'XII'),
('L30.9',   'Dermatitis, unspecified',                                   'Skin',                  'XII'),
('L40.0',   'Psoriasis vulgaris',                                        'Skin',                  'XII'),
('L50.9',   'Urticaria, unspecified',                                    'Skin',                  'XII'),
('L70.0',   'Acne vulgaris',                                             'Skin',                  'XII'),

-- Musculoskeletal (M)
('M06.9',   'Rheumatoid arthritis, unspecified',                         'Musculoskeletal',       'XIII'),
('M10.9',   'Gout, unspecified',                                         'Musculoskeletal',       'XIII'),
('M17.9',   'Osteoarthritis of knee, unspecified',                       'Musculoskeletal',       'XIII'),
('M19.90',  'Unspecified osteoarthritis, unspecified site',              'Musculoskeletal',       'XIII'),
('M25.50',  'Pain in unspecified joint',                                 'Musculoskeletal',       'XIII'),
('M54.2',   'Cervicalgia',                                               'Musculoskeletal',       'XIII'),
('M54.5',   'Low back pain',                                             'Musculoskeletal',       'XIII'),
('M54.9',   'Dorsalgia, unspecified',                                    'Musculoskeletal',       'XIII'),
('M62.830', 'Muscle spasm of back',                                      'Musculoskeletal',       'XIII'),
('M79.3',   'Panniculitis, unspecified',                                 'Musculoskeletal',       'XIII'),
('M81.0',   'Age-related osteoporosis without current pathological fracture','Musculoskeletal',   'XIII'),

-- Genitourinary (N)
('N18.9',   'Chronic kidney disease, unspecified',                       'Genitourinary',         'XIV'),
('N18.3',   'Chronic kidney disease, stage 3 (moderate)',                'Genitourinary',         'XIV'),
('N39.0',   'Urinary tract infection, site not specified',               'Genitourinary',         'XIV'),
('N40.0',   'Benign prostatic hyperplasia without lower urinary sx',     'Genitourinary',         'XIV'),

-- Pregnancy (O)
('O09.90',  'Supervision of high risk pregnancy, unspecified',           'Pregnancy',             'XV'),

-- Perinatal (P)
('P07.39',  'Other low birth weight newborn',                            'Perinatal',             'XVI'),

-- Symptoms / Signs (R)
('R00.0',   'Tachycardia, unspecified',                                  'Symptoms',              'XVIII'),
('R05.9',   'Cough, unspecified',                                        'Symptoms',              'XVIII'),
('R06.02',  'Shortness of breath',                                       'Symptoms',              'XVIII'),
('R07.9',   'Chest pain, unspecified',                                   'Symptoms',              'XVIII'),
('R10.9',   'Unspecified abdominal pain',                                'Symptoms',              'XVIII'),
('R11.0',   'Nausea',                                                    'Symptoms',              'XVIII'),
('R42',     'Dizziness and giddiness',                                   'Symptoms',              'XVIII'),
('R50.9',   'Fever, unspecified',                                        'Symptoms',              'XVIII'),
('R51.9',   'Headache, unspecified',                                     'Symptoms',              'XVIII'),
('R53.83',  'Other fatigue',                                             'Symptoms',              'XVIII'),
('R73.03',  'Prediabetes',                                               'Symptoms',              'XVIII'),

-- Injury (S/T)
('S62.509A','Unspecified fracture of unspecified wrist, initial enc',    'Injury',                'XIX'),
('T78.40XA','Allergy, unspecified, initial encounter',                   'Injury',                'XIX'),

-- External Causes (Z)
('Z00.00',  'Encounter for general adult medical examination',           'Factors',               'XXI'),
('Z00.129', 'Encounter for routine child health exam without abnormal',  'Factors',               'XXI'),
('Z12.31',  'Encounter for screening mammogram for malignant neoplasm', 'Factors',               'XXI'),
('Z23',     'Encounter for immunization',                                'Factors',               'XXI'),
('Z79.4',   'Long term (current) use of insulin',                       'Factors',               'XXI'),
('Z79.899', 'Other long term (current) drug therapy',                   'Factors',               'XXI'),
('Z87.891', 'Personal history of nicotine dependence',                   'Factors',               'XXI'),
('Z95.1',   'Presence of aortocoronary bypass graft',                   'Factors',               'XXI')
ON CONFLICT (code) DO NOTHING;
