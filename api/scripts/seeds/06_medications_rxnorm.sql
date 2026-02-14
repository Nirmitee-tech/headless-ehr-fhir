-- Seed: Common Medications with RxNorm Codes (100+)
-- Source: RxNorm (National Library of Medicine)

CREATE TABLE IF NOT EXISTS reference_medication (
    rxnorm_code VARCHAR(20) PRIMARY KEY,
    display VARCHAR(500) NOT NULL,
    generic_name VARCHAR(255),
    drug_class VARCHAR(100),
    route VARCHAR(50),
    form VARCHAR(50),
    system_uri VARCHAR(255) DEFAULT 'http://www.nlm.nih.gov/research/umls/rxnorm'
);

INSERT INTO reference_medication (rxnorm_code, display, generic_name, drug_class, route, form) VALUES
-- Diabetes
('860975',  'Metformin 500 mg oral tablet',                 'Metformin',          'Biguanide',                    'oral',   'tablet'),
('860981',  'Metformin 1000 mg oral tablet',                'Metformin',          'Biguanide',                    'oral',   'tablet'),
('861007',  'Metformin ER 500 mg oral tablet',              'Metformin ER',       'Biguanide',                    'oral',   'tablet'),
('1373463', 'Empagliflozin 10 mg oral tablet',              'Empagliflozin',      'SGLT2 Inhibitor',              'oral',   'tablet'),
('1373464', 'Empagliflozin 25 mg oral tablet',              'Empagliflozin',      'SGLT2 Inhibitor',              'oral',   'tablet'),
('897122',  'Sitagliptin 100 mg oral tablet',               'Sitagliptin',        'DPP-4 Inhibitor',              'oral',   'tablet'),
('1551291', 'Semaglutide 1 mg/dose injection pen',          'Semaglutide',        'GLP-1 RA',                     'subcut', 'injection'),
('311040',  'Glipizide 5 mg oral tablet',                   'Glipizide',          'Sulfonylurea',                 'oral',   'tablet'),
('351761',  'Insulin Glargine 100 units/mL injection',      'Insulin Glargine',   'Long-Acting Insulin',          'subcut', 'injection'),
('847187',  'Insulin Lispro 100 units/mL injection',        'Insulin Lispro',     'Rapid-Acting Insulin',         'subcut', 'injection'),

-- Cardiovascular - ACE Inhibitors / ARBs
('314076',  'Lisinopril 10 mg oral tablet',                 'Lisinopril',         'ACE Inhibitor',                'oral',   'tablet'),
('314077',  'Lisinopril 20 mg oral tablet',                 'Lisinopril',         'ACE Inhibitor',                'oral',   'tablet'),
('979480',  'Losartan 50 mg oral tablet',                   'Losartan',           'ARB',                          'oral',   'tablet'),
('979485',  'Losartan 100 mg oral tablet',                  'Losartan',           'ARB',                          'oral',   'tablet'),
('349199',  'Valsartan 160 mg oral tablet',                 'Valsartan',          'ARB',                          'oral',   'tablet'),
('898687',  'Olmesartan 20 mg oral tablet',                 'Olmesartan',         'ARB',                          'oral',   'tablet'),

-- Cardiovascular - Beta-Blockers
('866924',  'Metoprolol succinate 25 mg oral tablet ER',    'Metoprolol ER',      'Beta-Blocker',                 'oral',   'tablet'),
('866932',  'Metoprolol succinate 50 mg oral tablet ER',    'Metoprolol ER',      'Beta-Blocker',                 'oral',   'tablet'),
('200031',  'Carvedilol 12.5 mg oral tablet',               'Carvedilol',         'Beta-Blocker',                 'oral',   'tablet'),
('200033',  'Carvedilol 25 mg oral tablet',                 'Carvedilol',         'Beta-Blocker',                 'oral',   'tablet'),
('104375',  'Atenolol 25 mg oral tablet',                   'Atenolol',           'Beta-Blocker',                 'oral',   'tablet'),
('104376',  'Atenolol 50 mg oral tablet',                   'Atenolol',           'Beta-Blocker',                 'oral',   'tablet'),
('856422',  'Propranolol 20 mg oral tablet',                'Propranolol',        'Beta-Blocker',                 'oral',   'tablet'),

-- Cardiovascular - Calcium Channel Blockers
('197361',  'Amlodipine 5 mg oral tablet',                  'Amlodipine',         'Calcium Channel Blocker',      'oral',   'tablet'),
('197362',  'Amlodipine 10 mg oral tablet',                 'Amlodipine',         'Calcium Channel Blocker',      'oral',   'tablet'),
('898719',  'Diltiazem ER 120 mg oral capsule',             'Diltiazem ER',       'Calcium Channel Blocker',      'oral',   'capsule'),
('198032',  'Nifedipine ER 30 mg oral tablet',              'Nifedipine ER',      'Calcium Channel Blocker',      'oral',   'tablet'),

-- Cardiovascular - Diuretics
('310798',  'Hydrochlorothiazide 25 mg oral tablet',        'HCTZ',               'Thiazide Diuretic',            'oral',   'tablet'),
('197417',  'Furosemide 20 mg oral tablet',                 'Furosemide',         'Loop Diuretic',                'oral',   'tablet'),
('197418',  'Furosemide 40 mg oral tablet',                 'Furosemide',         'Loop Diuretic',                'oral',   'tablet'),
('104220',  'Spironolactone 25 mg oral tablet',             'Spironolactone',     'K-Sparing Diuretic',           'oral',   'tablet'),

-- Cardiovascular - Statins
('259255',  'Atorvastatin 10 mg oral tablet',               'Atorvastatin',       'HMG-CoA Reductase Inhibitor',  'oral',   'tablet'),
('259256',  'Atorvastatin 20 mg oral tablet',               'Atorvastatin',       'HMG-CoA Reductase Inhibitor',  'oral',   'tablet'),
('259257',  'Atorvastatin 40 mg oral tablet',               'Atorvastatin',       'HMG-CoA Reductase Inhibitor',  'oral',   'tablet'),
('312961',  'Rosuvastatin 10 mg oral tablet',               'Rosuvastatin',       'HMG-CoA Reductase Inhibitor',  'oral',   'tablet'),
('312962',  'Rosuvastatin 20 mg oral tablet',               'Rosuvastatin',       'HMG-CoA Reductase Inhibitor',  'oral',   'tablet'),
('197904',  'Simvastatin 20 mg oral tablet',                'Simvastatin',        'HMG-CoA Reductase Inhibitor',  'oral',   'tablet'),
('197905',  'Simvastatin 40 mg oral tablet',                'Simvastatin',        'HMG-CoA Reductase Inhibitor',  'oral',   'tablet'),
('861643',  'Pravastatin 40 mg oral tablet',                'Pravastatin',        'HMG-CoA Reductase Inhibitor',  'oral',   'tablet'),

-- Cardiovascular - Anticoagulants/Antiplatelets
('855332',  'Warfarin 5 mg oral tablet',                    'Warfarin',           'Anticoagulant',                'oral',   'tablet'),
('1037045', 'Apixaban 5 mg oral tablet',                    'Apixaban',           'DOAC',                         'oral',   'tablet'),
('1114198', 'Rivaroxaban 20 mg oral tablet',                'Rivaroxaban',        'DOAC',                         'oral',   'tablet'),
('309362',  'Clopidogrel 75 mg oral tablet',                'Clopidogrel',        'Antiplatelet',                 'oral',   'tablet'),
('243670',  'Aspirin 81 mg oral tablet (chewable)',         'Aspirin',            'Antiplatelet',                 'oral',   'tablet'),

-- GI / Acid Suppression
('311272',  'Omeprazole 20 mg oral capsule',                'Omeprazole',         'Proton Pump Inhibitor',        'oral',   'capsule'),
('311273',  'Omeprazole 40 mg oral capsule',                'Omeprazole',         'Proton Pump Inhibitor',        'oral',   'capsule'),
('828876',  'Pantoprazole 40 mg oral tablet',               'Pantoprazole',       'Proton Pump Inhibitor',        'oral',   'tablet'),
('602731',  'Esomeprazole 40 mg oral capsule',              'Esomeprazole',       'Proton Pump Inhibitor',        'oral',   'capsule'),
('261106',  'Famotidine 20 mg oral tablet',                 'Famotidine',         'H2 Blocker',                   'oral',   'tablet'),
('314231',  'Ondansetron 4 mg oral tablet',                 'Ondansetron',        'Antiemetic',                   'oral',   'tablet'),

-- Respiratory
('245314',  'Albuterol 90 mcg/actuation inhaler',           'Albuterol',          'SABA',                         'inhal',  'inhaler'),
('896188',  'Fluticasone-Salmeterol 250/50 inhaler',        'Fluticasone/Salmet', 'ICS-LABA',                     'inhal',  'inhaler'),
('746763',  'Montelukast 10 mg oral tablet',                'Montelukast',        'Leukotriene Modifier',         'oral',   'tablet'),
('1014585', 'Tiotropium 18 mcg inhalation capsule',         'Tiotropium',         'LAMA',                         'inhal',  'capsule'),
('1049589', 'Budesonide-Formoterol 160/4.5 mcg inhaler',   'Budesonide/Formo',   'ICS-LABA',                     'inhal',  'inhaler'),

-- Pain / Analgesics
('198440',  'Acetaminophen 500 mg oral tablet',             'Acetaminophen',      'Analgesic',                    'oral',   'tablet'),
('197803',  'Ibuprofen 200 mg oral tablet',                 'Ibuprofen',          'NSAID',                        'oral',   'tablet'),
('197805',  'Ibuprofen 400 mg oral tablet',                 'Ibuprofen',          'NSAID',                        'oral',   'tablet'),
('197806',  'Ibuprofen 600 mg oral tablet',                 'Ibuprofen',          'NSAID',                        'oral',   'tablet'),
('198405',  'Naproxen 500 mg oral tablet',                  'Naproxen',           'NSAID',                        'oral',   'tablet'),
('197696',  'Diclofenac 75 mg oral tablet',                 'Diclofenac',         'NSAID',                        'oral',   'tablet'),
('1049221', 'Acetaminophen-Codeine 300/30 mg oral tablet',  'APAP/Codeine',       'Opioid Combination',           'oral',   'tablet'),
('856980',  'Tramadol 50 mg oral tablet',                   'Tramadol',           'Opioid Analgesic',             'oral',   'tablet'),

-- Neurological / Psych
('310384',  'Gabapentin 300 mg oral capsule',               'Gabapentin',         'Anticonvulsant',               'oral',   'capsule'),
('310385',  'Gabapentin 400 mg oral capsule',               'Gabapentin',         'Anticonvulsant',               'oral',   'capsule'),
('312938',  'Pregabalin 75 mg oral capsule',                'Pregabalin',         'Anticonvulsant',               'oral',   'capsule'),
('312940',  'Pregabalin 150 mg oral capsule',               'Pregabalin',         'Anticonvulsant',               'oral',   'capsule'),
('313585',  'Sertraline 50 mg oral tablet',                 'Sertraline',         'SSRI',                         'oral',   'tablet'),
('313586',  'Sertraline 100 mg oral tablet',                'Sertraline',         'SSRI',                         'oral',   'tablet'),
('312938',  'Escitalopram 10 mg oral tablet',               'Escitalopram',       'SSRI',                         'oral',   'tablet'),
('596926',  'Duloxetine 30 mg oral capsule',                'Duloxetine',         'SNRI',                         'oral',   'capsule'),
('596930',  'Duloxetine 60 mg oral capsule',                'Duloxetine',         'SNRI',                         'oral',   'capsule'),
('312087',  'Amitriptyline 25 mg oral tablet',              'Amitriptyline',      'TCA',                          'oral',   'tablet'),
('197320',  'Alprazolam 0.5 mg oral tablet',                'Alprazolam',         'Benzodiazepine',               'oral',   'tablet'),
('311700',  'Lorazepam 1 mg oral tablet',                   'Lorazepam',          'Benzodiazepine',               'oral',   'tablet'),
('835564',  'Trazodone 50 mg oral tablet',                  'Trazodone',          'Antidepressant',               'oral',   'tablet'),
('312246',  'Quetiapine 25 mg oral tablet',                 'Quetiapine',         'Atypical Antipsychotic',       'oral',   'tablet'),
('197694',  'Clonazepam 0.5 mg oral tablet',                'Clonazepam',         'Benzodiazepine',               'oral',   'tablet'),
('311556',  'Zolpidem 10 mg oral tablet',                   'Zolpidem',           'Sedative-Hypnotic',            'oral',   'tablet'),
('259111',  'Bupropion 150 mg oral tablet',                 'Bupropion',          'Aminoketone',                  'oral',   'tablet'),

-- Thyroid
('966222',  'Levothyroxine 50 mcg oral tablet',             'Levothyroxine',      'Thyroid Hormone',              'oral',   'tablet'),
('966225',  'Levothyroxine 100 mcg oral tablet',            'Levothyroxine',      'Thyroid Hormone',              'oral',   'tablet'),

-- Antibiotics
('308182',  'Amoxicillin 500 mg oral capsule',              'Amoxicillin',        'Penicillin',                   'oral',   'capsule'),
('860225',  'Amoxicillin-Clavulanate 875/125 mg tablet',    'Amoxicillin-Clav',   'Penicillin Combination',       'oral',   'tablet'),
('309054',  'Azithromycin 250 mg oral tablet',              'Azithromycin',       'Macrolide',                    'oral',   'tablet'),
('197511',  'Ciprofloxacin 500 mg oral tablet',             'Ciprofloxacin',      'Fluoroquinolone',              'oral',   'tablet'),
('309079',  'Cephalexin 500 mg oral capsule',               'Cephalexin',         'Cephalosporin',                'oral',   'capsule'),
('197595',  'Doxycycline 100 mg oral capsule',              'Doxycycline',        'Tetracycline',                 'oral',   'capsule'),
('847626',  'Sulfamethoxazole-TMP 800/160 mg oral tablet',  'SMX-TMP DS',         'Sulfonamide Combination',      'oral',   'tablet'),
('309308',  'Ceftriaxone 1 g injection',                    'Ceftriaxone',        'Cephalosporin',                'IV',     'injection'),
('197450',  'Clindamycin 300 mg oral capsule',              'Clindamycin',        'Lincosamide',                  'oral',   'capsule'),

-- Misc Common
('310429',  'Prednisone 10 mg oral tablet',                 'Prednisone',         'Corticosteroid',               'oral',   'tablet'),
('312617',  'Prednisone 20 mg oral tablet',                 'Prednisone',         'Corticosteroid',               'oral',   'tablet'),
('1191222', 'Methylprednisolone 4 mg dose pack',            'Methylprednisolone', 'Corticosteroid',               'oral',   'tablet'),
('197381',  'Cetirizine 10 mg oral tablet',                 'Cetirizine',         'Antihistamine',                'oral',   'tablet'),
('998695',  'Loratadine 10 mg oral tablet',                 'Loratadine',         'Antihistamine',                'oral',   'tablet'),
('310436',  'Potassium Chloride 20 mEq oral tablet',        'KCl',                'Electrolyte',                  'oral',   'tablet'),
('197541',  'Cyclobenzaprine 10 mg oral tablet',            'Cyclobenzaprine',    'Muscle Relaxant',              'oral',   'tablet'),
('309043',  'Tamsulosin 0.4 mg oral capsule',               'Tamsulosin',         'Alpha-Blocker',                'oral',   'capsule'),
('860092',  'Finasteride 5 mg oral tablet',                 'Finasteride',        '5-Alpha Reductase Inhibitor',  'oral',   'tablet')
ON CONFLICT (code) DO NOTHING;
