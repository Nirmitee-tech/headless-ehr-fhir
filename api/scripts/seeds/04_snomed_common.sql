-- Seed: Common SNOMED CT Codes (procedures and clinical findings, 50+)
-- Source: SNOMED CT (Systematized Nomenclature of Medicine -- Clinical Terms)

CREATE TABLE IF NOT EXISTS reference_snomed (
    code VARCHAR(20) PRIMARY KEY,
    display VARCHAR(500) NOT NULL,
    semantic_tag VARCHAR(50),
    category VARCHAR(50),
    system_uri VARCHAR(255) DEFAULT 'http://snomed.info/sct'
);

INSERT INTO reference_snomed (code, display, semantic_tag, category) VALUES
-- Procedures
('80146002',  'Appendectomy',                                           'procedure', 'surgical'),
('73761001',  'Colonoscopy',                                            'procedure', 'diagnostic'),
('174041007', 'Laparoscopic cholecystectomy',                           'procedure', 'surgical'),
('18286005',  'Catheterization of heart',                               'procedure', 'diagnostic'),
('232717009', 'Coronary artery bypass graft',                           'procedure', 'surgical'),
('11466000',  'Cesarean section',                                       'procedure', 'surgical'),
('287664005', 'Total knee replacement',                                 'procedure', 'surgical'),
('76164006',  'Biopsy',                                                 'procedure', 'diagnostic'),
('40701008',  'Echocardiography',                                       'procedure', 'diagnostic'),
('241615005', 'Magnetic resonance imaging',                             'procedure', 'diagnostic'),
('77477000',  'Computerized axial tomography',                          'procedure', 'diagnostic'),
('363680008', 'X-ray',                                                   'procedure', 'diagnostic'),
('16310003',  'Ultrasonography',                                        'procedure', 'diagnostic'),
('274025005', 'Hip replacement',                                         'procedure', 'surgical'),
('307280005', 'Insertion of cardiac pacemaker',                         'procedure', 'surgical'),
('392021009', 'Lumpectomy',                                              'procedure', 'surgical'),
('173171007', 'Hysterectomy',                                            'procedure', 'surgical'),
('90470006',  'Prostatectomy',                                           'procedure', 'surgical'),
('44578002',  'Tonsillectomy',                                           'procedure', 'surgical'),
('84114007',  'Heart transplant',                                        'procedure', 'surgical'),
('71388002',  'Procedure (generic)',                                     'procedure', 'general'),
('387713003', 'Surgical procedure',                                      'procedure', 'surgical'),
('103693007', 'Diagnostic procedure',                                    'procedure', 'diagnostic'),
('386637004', 'Obstetric procedure',                                     'procedure', 'obstetric'),
('225358003', 'Wound care',                                              'procedure', 'nursing'),
('182813001', 'Emergency treatment',                                     'procedure', 'emergency'),

-- Clinical Findings
('38341003',  'Hypertensive disorder',                                  'finding',   'cardiovascular'),
('73211009',  'Diabetes mellitus',                                       'finding',   'endocrine'),
('44054006',  'Diabetes mellitus type 2',                                'finding',   'endocrine'),
('46635009',  'Diabetes mellitus type 1',                                'finding',   'endocrine'),
('84757009',  'Epilepsy',                                                'finding',   'neurological'),
('195967001', 'Asthma',                                                  'finding',   'respiratory'),
('13645005',  'COPD',                                                    'finding',   'respiratory'),
('22298006',  'Myocardial infarction',                                   'finding',   'cardiovascular'),
('84114007',  'Heart failure',                                           'finding',   'cardiovascular'),
('49436004',  'Atrial fibrillation',                                     'finding',   'cardiovascular'),
('230690007', 'Cerebrovascular accident',                                'finding',   'neurological'),
('36971009',  'Sinusitis',                                               'finding',   'respiratory'),
('68566005',  'Urinary tract infection',                                 'finding',   'genitourinary'),
('233604007', 'Pneumonia',                                               'finding',   'respiratory'),
('25064002',  'Headache',                                                'finding',   'neurological'),
('386661006', 'Fever',                                                   'finding',   'general'),
('271807003', 'Eruption of skin',                                        'finding',   'dermatological'),
('267036007', 'Dyspnea',                                                 'finding',   'respiratory'),
('29857009',  'Chest pain',                                              'finding',   'cardiovascular'),
('21522001',  'Abdominal pain',                                          'finding',   'gastrointestinal'),
('161891005', 'Back pain',                                               'finding',   'musculoskeletal'),
('35489007',  'Depressive disorder',                                     'finding',   'mental'),
('48694002',  'Anxiety disorder',                                        'finding',   'mental'),
('414545008', 'Ischemic heart disease',                                  'finding',   'cardiovascular'),
('90708001',  'Kidney disease',                                          'finding',   'genitourinary')
ON CONFLICT (code) DO NOTHING;
