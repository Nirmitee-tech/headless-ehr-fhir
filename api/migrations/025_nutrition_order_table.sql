-- 025_nutrition_order_table.sql
-- NutritionOrder resource table

CREATE TABLE IF NOT EXISTS nutrition_order (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fhir_id         TEXT UNIQUE NOT NULL,
    status          VARCHAR(32) NOT NULL DEFAULT 'draft',
    intent          VARCHAR(32) NOT NULL,
    patient_id      UUID NOT NULL REFERENCES patient(id),
    encounter_id    UUID REFERENCES encounter(id),
    orderer_id      UUID REFERENCES practitioner(id),
    date_time       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    oral_diet       JSONB,
    supplement      JSONB,
    enteral_formula JSONB,
    allergy_intolerances      TEXT[],
    food_preference_modifiers TEXT[],
    exclude_food_modifiers    TEXT[],
    note            TEXT,
    version_id      INTEGER NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nutrition_order_patient_id ON nutrition_order(patient_id);
CREATE INDEX IF NOT EXISTS idx_nutrition_order_encounter_id ON nutrition_order(encounter_id);
CREATE INDEX IF NOT EXISTS idx_nutrition_order_status ON nutrition_order(status);
CREATE INDEX IF NOT EXISTS idx_nutrition_order_fhir_id ON nutrition_order(fhir_id);
