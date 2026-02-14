-- Seed: Standard Encounter Types
-- Source: FHIR R4 ValueSet v3.ActEncounterCode and HL7 encounter types

CREATE TABLE IF NOT EXISTS reference_encounter_type (
    code VARCHAR(20) PRIMARY KEY,
    display VARCHAR(255) NOT NULL,
    class_code VARCHAR(20) NOT NULL,
    description TEXT,
    system_uri VARCHAR(255) DEFAULT 'http://terminology.hl7.org/CodeSystem/v3-ActCode'
);

INSERT INTO reference_encounter_type (code, display, class_code, description) VALUES
-- Ambulatory
('AMB',     'Ambulatory',                           'AMB',   'Outpatient encounter for evaluation and management'),
('WELLNESS','Wellness Visit',                        'AMB',   'Routine preventive care visit'),
('FOLLOWUP','Follow-up Visit',                       'AMB',   'Follow-up after previous treatment or procedure'),
('WALKIN',  'Walk-in Visit',                         'AMB',   'Unscheduled walk-in visit'),
('CHECKUP', 'Checkup',                               'AMB',   'General health check-up'),
('PRENC',   'Pre-admission',                         'AMB',   'Pre-admission testing and evaluation'),
('POSTOP',  'Post-operative Visit',                  'AMB',   'Post-operative follow-up visit'),

-- Emergency
('EMER',    'Emergency',                              'EMER',  'Emergency department encounter'),
('URGCARE', 'Urgent Care',                            'EMER',  'Urgent care encounter for acute non-emergent conditions'),

-- Inpatient
('IMP',     'Inpatient',                              'IMP',   'Inpatient hospitalization'),
('ACUTE',   'Inpatient Acute',                        'IMP',   'Acute inpatient hospitalization'),
('NONAC',   'Inpatient Non-Acute',                    'IMP',   'Non-acute inpatient stay (rehabilitation, skilled nursing)'),
('OBSENC',  'Observation',                            'IMP',   'Observation stay (outpatient status in hospital)'),
('SS',      'Short Stay',                             'IMP',   'Short stay encounter (e.g., same-day surgery)'),

-- Virtual / Telehealth
('VR',      'Virtual',                                'VR',    'Telehealth / virtual encounter'),
('PHONE',   'Phone Visit',                            'VR',    'Telephone-based clinical encounter'),
('VIDEO',   'Video Visit',                            'VR',    'Video-based telehealth encounter'),
('ASYNC',   'Asynchronous Telehealth',                'VR',    'Store-and-forward or messaging-based encounter'),

-- Home Health
('HH',      'Home Health',                            'HH',    'Home health visit'),
('HOMEVISIT','Home Visit',                            'HH',    'Clinical home visit by provider'),

-- Field / Other
('FLD',     'Field',                                  'FLD',   'Field encounter (outside healthcare facility)'),
('DAYCASE', 'Day Case / Ambulatory Surgery',          'SS',    'Same-day surgical procedure'),
('CONSULT', 'Consultation',                           'AMB',   'Specialist consultation visit'),
('PREOP',   'Pre-Operative Evaluation',               'AMB',   'Pre-operative assessment visit'),
('NEWPAT',  'New Patient Visit',                      'AMB',   'Initial visit for a new patient'),
('REFVISIT','Referral Visit',                         'AMB',   'Visit resulting from a referral'),
('GROUP',   'Group Therapy Session',                  'AMB',   'Group therapy encounter'),
('PROCENC', 'Procedure Visit',                        'AMB',   'Encounter for a specific procedure'),

-- Lab / Diagnostic
('LABENC',  'Laboratory Encounter',                   'AMB',   'Encounter for laboratory specimen collection'),
('RADENC',  'Radiology Encounter',                    'AMB',   'Encounter for imaging or radiology study')
ON CONFLICT (code) DO NOTHING;
