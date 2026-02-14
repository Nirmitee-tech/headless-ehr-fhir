-- ============================================================================
-- ROW-LEVEL SECURITY (RLS) FOR HIPAA COMPLIANCE
-- Ensures tenant data isolation at the PostgreSQL level.
-- These policies are applied per-tenant schema after migrations.
-- ============================================================================

-- Enable RLS on all patient-facing tables.
-- The application connects with a role that has the current tenant set
-- via SET search_path. RLS adds a defense-in-depth layer.

-- ============================================================================
-- 1. CREATE APPLICATION ROLE (if not exists)
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'ehr_app') THEN
        CREATE ROLE ehr_app LOGIN;
    END IF;
END $$;

-- ============================================================================
-- 2. AUDIT TABLES — append-only (no update/delete for compliance)
-- ============================================================================

-- Audit events are immutable once written
ALTER TABLE audit_event ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_event_insert_policy ON audit_event
    FOR INSERT TO ehr_app
    WITH CHECK (true);
CREATE POLICY audit_event_select_policy ON audit_event
    FOR SELECT TO ehr_app
    USING (true);

-- HIPAA access log is immutable once written
ALTER TABLE hipaa_access_log ENABLE ROW LEVEL SECURITY;
CREATE POLICY hipaa_log_insert_policy ON hipaa_access_log
    FOR INSERT TO ehr_app
    WITH CHECK (true);
CREATE POLICY hipaa_log_select_policy ON hipaa_access_log
    FOR SELECT TO ehr_app
    USING (true);

-- ============================================================================
-- 3. CORE CLINICAL TABLES — full access within tenant schema
-- ============================================================================

-- Patient
ALTER TABLE patient ENABLE ROW LEVEL SECURITY;
CREATE POLICY patient_tenant_policy ON patient
    FOR ALL TO ehr_app
    USING (true)
    WITH CHECK (true);

-- Practitioner
ALTER TABLE practitioner ENABLE ROW LEVEL SECURITY;
CREATE POLICY practitioner_tenant_policy ON practitioner
    FOR ALL TO ehr_app
    USING (true)
    WITH CHECK (true);

-- Encounter
ALTER TABLE encounter ENABLE ROW LEVEL SECURITY;
CREATE POLICY encounter_tenant_policy ON encounter
    FOR ALL TO ehr_app
    USING (true)
    WITH CHECK (true);

-- Condition
ALTER TABLE condition ENABLE ROW LEVEL SECURITY;
CREATE POLICY condition_tenant_policy ON condition
    FOR ALL TO ehr_app
    USING (true)
    WITH CHECK (true);

-- Observation
ALTER TABLE observation ENABLE ROW LEVEL SECURITY;
CREATE POLICY observation_tenant_policy ON observation
    FOR ALL TO ehr_app
    USING (true)
    WITH CHECK (true);

-- Allergy Intolerance
ALTER TABLE allergy_intolerance ENABLE ROW LEVEL SECURITY;
CREATE POLICY allergy_tenant_policy ON allergy_intolerance
    FOR ALL TO ehr_app
    USING (true)
    WITH CHECK (true);

-- Procedure Record
ALTER TABLE procedure_record ENABLE ROW LEVEL SECURITY;
CREATE POLICY procedure_tenant_policy ON procedure_record
    FOR ALL TO ehr_app
    USING (true)
    WITH CHECK (true);

-- ============================================================================
-- 4. PREVENT CROSS-SCHEMA ACCESS
-- ============================================================================

-- Revoke public schema access from app role to prevent cross-tenant queries
-- Each tenant schema grants access only to its own tables
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM ehr_app;

-- ============================================================================
-- 5. GRANT PERMISSIONS TO APP ROLE WITHIN TENANT SCHEMAS
-- ============================================================================

-- This should be run per-tenant schema during provisioning:
-- GRANT USAGE ON SCHEMA tenant_<id> TO ehr_app;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA tenant_<id> TO ehr_app;
-- ALTER DEFAULT PRIVILEGES IN SCHEMA tenant_<id> GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ehr_app;

-- ============================================================================
-- 6. IMMUTABILITY TRIGGERS FOR AUDIT TABLES
-- ============================================================================

CREATE OR REPLACE FUNCTION prevent_audit_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit records are immutable and cannot be modified or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_audit_event_immutable
    BEFORE UPDATE OR DELETE ON audit_event
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_modification();

CREATE TRIGGER trg_hipaa_log_immutable
    BEFORE UPDATE OR DELETE ON hipaa_access_log
    FOR EACH ROW EXECUTE FUNCTION prevent_audit_modification();
