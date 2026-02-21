-- 037_us_core_patient_extensions.sql: US Core race, ethnicity, birth sex

ALTER TABLE patient ADD COLUMN IF NOT EXISTS race_code VARCHAR(50);
ALTER TABLE patient ADD COLUMN IF NOT EXISTS race_display VARCHAR(200);
ALTER TABLE patient ADD COLUMN IF NOT EXISTS race_text VARCHAR(200);
ALTER TABLE patient ADD COLUMN IF NOT EXISTS ethnicity_code VARCHAR(50);
ALTER TABLE patient ADD COLUMN IF NOT EXISTS ethnicity_display VARCHAR(200);
ALTER TABLE patient ADD COLUMN IF NOT EXISTS ethnicity_text VARCHAR(200);
ALTER TABLE patient ADD COLUMN IF NOT EXISTS birth_sex VARCHAR(10);
