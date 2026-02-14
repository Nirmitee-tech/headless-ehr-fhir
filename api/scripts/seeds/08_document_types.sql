-- Seed: Standard Clinical Document Types
-- Source: LOINC Document Type codes used in FHIR DocumentReference / Composition

CREATE TABLE IF NOT EXISTS reference_document_type (
    code VARCHAR(20) PRIMARY KEY,
    display VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    description TEXT,
    system_uri VARCHAR(255) DEFAULT 'http://loinc.org'
);

INSERT INTO reference_document_type (code, display, category, description) VALUES
-- Clinical Notes
('11488-4',  'Consultation note',                         'Clinical Note',     'Specialist consultation documentation'),
('18842-5',  'Discharge summary',                         'Clinical Note',     'Hospital discharge summary'),
('34117-2',  'History and physical note',                  'Clinical Note',     'Initial H&P documentation'),
('11506-3',  'Progress note',                              'Clinical Note',     'Follow-up or daily progress note'),
('28570-0',  'Procedure note',                             'Clinical Note',     'Procedural documentation'),
('34133-9',  'Summary of episode note',                    'Clinical Note',     'Summary note for episode of care'),
('57133-1',  'Referral note',                              'Clinical Note',     'Referral documentation'),
('34111-5',  'Emergency department note',                  'Clinical Note',     'ED encounter documentation'),
('34108-1',  'Outpatient note',                            'Clinical Note',     'Ambulatory visit documentation'),
('18761-7',  'Transfer summary note',                      'Clinical Note',     'Patient transfer documentation'),
('34900-1',  'General medicine note',                      'Clinical Note',     'General medicine encounter note'),
('15508-5',  'Labor and delivery records',                 'Clinical Note',     'Obstetric delivery documentation'),

-- Operative / Surgical
('11504-8',  'Surgical operation note',                    'Operative',         'Operative report documentation'),
('28573-4',  'Anesthesia record',                          'Operative',         'Anesthesia procedure record'),
('59258-4',  'Emergency department discharge summary',     'Operative',         'ED discharge summary'),

-- Diagnostic Reports
('24323-8',  'Comprehensive metabolic panel',              'Diagnostic Report', 'CMP lab result report'),
('58410-2',  'Complete blood count (CBC)',                  'Diagnostic Report', 'CBC lab result report'),
('24331-1',  'Lipid panel',                                'Diagnostic Report', 'Lipid panel result report'),
('18723-7',  'Hematology studies',                         'Diagnostic Report', 'Hematology lab report'),
('18719-5',  'Chemistry studies',                          'Diagnostic Report', 'Chemistry lab report'),
('18725-2',  'Microbiology studies',                       'Diagnostic Report', 'Microbiology lab report'),
('18727-8',  'Serology studies',                           'Diagnostic Report', 'Serology lab report'),
('18717-9',  'Blood bank studies',                         'Diagnostic Report', 'Blood bank lab report'),

-- Radiology
('18748-4',  'Diagnostic imaging study',                   'Radiology',         'Radiology report'),
('30954-2',  'Relevant diagnostic tests and/or lab data',  'Radiology',         'Relevant diagnostic data'),
('18782-3',  'Radiology study (observation)',               'Radiology',         'Radiological observation report'),

-- Pathology
('11526-1',  'Pathology study',                            'Pathology',         'Pathology report'),
('60568-3',  'Pathology synoptic report',                  'Pathology',         'Synoptic pathology report'),

-- Patient Records / Administrative
('47420-5',  'Functional status assessment note',          'Assessment',        'Functional status assessment'),
('51847-2',  'Assessment and plan',                        'Assessment',        'Assessment and plan documentation'),
('34748-4',  'Telephone encounter note',                   'Clinical Note',     'Telehealth/phone encounter note'),
('68608-9',  'Summary note',                               'Clinical Note',     'Patient care summary note'),
('48765-2',  'Allergy list',                               'Patient List',      'Allergy and adverse reaction list'),
('10160-0',  'History of medication use',                   'Patient List',      'Medication history list'),
('11369-6',  'History of immunization',                     'Patient List',      'Immunization history'),
('57024-2',  'Health Quality Measure document',             'Administrative',    'Quality reporting document'),

-- Consent / Legal
('59284-0',  'Patient consent',                            'Legal',             'Consent for treatment document'),
('64293-4',  'Procedure consent',                          'Legal',             'Procedure-specific consent'),
('57016-8',  'Privacy policy acknowledgment',              'Legal',             'Privacy/HIPAA acknowledgment'),
('57017-6',  'Privacy policy organization document',       'Legal',             'Organization privacy policy'),

-- Advance Directives
('42348-3',  'Advance directive',                          'Advance Directive', 'Advance healthcare directive'),
('64298-3',  'Power of attorney for healthcare',           'Advance Directive', 'Healthcare proxy document'),
('75320-2',  'Advance directive - goals, preferences',     'Advance Directive', 'Goals and preferences'),

-- Education
('69730-0',  'Patient instructions',                       'Education',         'Patient education instructions'),
('34895-3',  'Education note',                             'Education',         'Patient education documentation')
ON CONFLICT (code) DO NOTHING;
