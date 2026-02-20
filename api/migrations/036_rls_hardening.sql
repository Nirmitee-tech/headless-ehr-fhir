-- ============================================================================
-- Migration 036: RLS HARDENING — Defense-in-Depth for Tenant Isolation
--
-- Strengthens Row-Level Security policies introduced in migration 018.
-- The original policies used USING (true), meaning RLS was enabled but
-- provided no actual filtering. This migration adds:
--
--   1. Session-variable helper functions (current_tenant_id / current_user_id)
--   2. Audit-trail columns (created_by / updated_by) on all clinical tables
--   3. Trigger-based auto-population of audit columns from session variables
--   4. Restrictive RLS policies on audit/compliance tables
--   5. Schema-crossing guard on clinical tables
--
-- Design rationale: Tenant isolation is schema-based (search_path). Rather
-- than adding a tenant_id column to every table, we harden RLS by:
--   - Ensuring the session variable matches the schema the query runs in
--   - Making audit tables readable only by compliance/admin roles
--   - Tracking every write with user attribution via triggers
--
-- This migration is IDEMPOTENT — safe to run multiple times.
-- ============================================================================

-- ============================================================================
-- 1. SESSION VARIABLE HELPER FUNCTIONS
-- ============================================================================

-- Returns the current tenant ID set by the application layer.
-- The application sets  SET app.current_tenant_id = '<id>'  on every connection.
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS text AS $$
BEGIN
    RETURN current_setting('app.current_tenant_id', true);
EXCEPTION
    WHEN undefined_object THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- Returns the current user ID set by the application layer.
-- The application sets  SET app.current_user_id = '<id>'  on every connection.
CREATE OR REPLACE FUNCTION current_user_id() RETURNS text AS $$
BEGIN
    RETURN current_setting('app.current_user_id', true);
EXCEPTION
    WHEN undefined_object THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- ============================================================================
-- 2. AUDIT-TRAIL COLUMNS ON CLINICAL TABLES
-- ============================================================================
-- All core clinical tables already have created_at / updated_at.
-- We add created_by / updated_by to record WHO performed the operation.

-- Patient
ALTER TABLE patient ADD COLUMN IF NOT EXISTS created_by TEXT;
ALTER TABLE patient ADD COLUMN IF NOT EXISTS updated_by TEXT;

-- Practitioner
ALTER TABLE practitioner ADD COLUMN IF NOT EXISTS created_by TEXT;
ALTER TABLE practitioner ADD COLUMN IF NOT EXISTS updated_by TEXT;

-- Encounter
ALTER TABLE encounter ADD COLUMN IF NOT EXISTS created_by TEXT;
ALTER TABLE encounter ADD COLUMN IF NOT EXISTS updated_by TEXT;

-- Condition
ALTER TABLE condition ADD COLUMN IF NOT EXISTS created_by TEXT;
ALTER TABLE condition ADD COLUMN IF NOT EXISTS updated_by TEXT;

-- Observation
ALTER TABLE observation ADD COLUMN IF NOT EXISTS created_by TEXT;
ALTER TABLE observation ADD COLUMN IF NOT EXISTS updated_by TEXT;

-- Allergy Intolerance
ALTER TABLE allergy_intolerance ADD COLUMN IF NOT EXISTS created_by TEXT;
ALTER TABLE allergy_intolerance ADD COLUMN IF NOT EXISTS updated_by TEXT;

-- Procedure Record
ALTER TABLE procedure_record ADD COLUMN IF NOT EXISTS created_by TEXT;
ALTER TABLE procedure_record ADD COLUMN IF NOT EXISTS updated_by TEXT;

-- ============================================================================
-- 3. AUDIT-TRAIL TRIGGER FUNCTION
-- ============================================================================
-- Automatically stamps created_by/updated_by from the session variable
-- set by the application. Falls back to current PostgreSQL user if the
-- session variable is not set (e.g., during migrations or manual maintenance).

CREATE OR REPLACE FUNCTION set_audit_user() RETURNS TRIGGER AS $$
DECLARE
    _user TEXT;
BEGIN
    _user := current_setting('app.current_user_id', true);
    IF _user IS NULL OR _user = '' THEN
        _user := current_user;
    END IF;

    IF TG_OP = 'INSERT' THEN
        NEW.created_by := COALESCE(NEW.created_by, _user);
        NEW.created_at := COALESCE(NEW.created_at, NOW());
    END IF;

    NEW.updated_by := _user;
    NEW.updated_at := NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 4. ATTACH AUDIT TRIGGERS TO CLINICAL TABLES
-- ============================================================================
-- DROP IF EXISTS is not available for triggers, so use a DO block.

DO $$
DECLARE
    _tables TEXT[] := ARRAY[
        'patient',
        'practitioner',
        'encounter',
        'condition',
        'observation',
        'allergy_intolerance',
        'procedure_record'
    ];
    _tbl TEXT;
    _trigger_name TEXT;
BEGIN
    FOREACH _tbl IN ARRAY _tables
    LOOP
        _trigger_name := 'trg_audit_user_' || _tbl;

        -- Drop existing trigger if present (idempotent)
        EXECUTE format(
            'DROP TRIGGER IF EXISTS %I ON %I',
            _trigger_name, _tbl
        );

        -- Create the trigger
        EXECUTE format(
            'CREATE TRIGGER %I
                BEFORE INSERT OR UPDATE ON %I
                FOR EACH ROW EXECUTE FUNCTION set_audit_user()',
            _trigger_name, _tbl
        );
    END LOOP;
END $$;

-- ============================================================================
-- 5. HARDEN RLS POLICIES ON CLINICAL TABLES
-- ============================================================================
-- Replace the permissive USING (true) policies from migration 018 with
-- policies that verify the session tenant variable matches the schema.
--
-- current_setting('search_path') returns something like 'tenant_acme, shared, public'.
-- We extract the first schema component and verify it matches
-- 'tenant_' || current_tenant_id().  If the session variable is not set,
-- no rows are visible — fail closed.

CREATE OR REPLACE FUNCTION rls_tenant_check() RETURNS boolean AS $$
DECLARE
    _tenant TEXT;
    _search_path TEXT;
    _first_schema TEXT;
BEGIN
    _tenant := current_setting('app.current_tenant_id', true);

    -- If no tenant variable is set, deny access (fail closed)
    IF _tenant IS NULL OR _tenant = '' THEN
        RETURN FALSE;
    END IF;

    -- Extract the first schema from search_path
    _search_path := current_setting('search_path', true);
    IF _search_path IS NULL OR _search_path = '' THEN
        RETURN FALSE;
    END IF;

    _first_schema := trim(split_part(_search_path, ',', 1));

    -- Verify the first schema matches the expected tenant schema
    RETURN _first_schema = ('tenant_' || _tenant);
END;
$$ LANGUAGE plpgsql STABLE;


-- Drop old permissive policies and create hardened replacements.
-- The new policies call rls_tenant_check() which validates that the
-- session variable and search_path are consistent.

DO $$
DECLARE
    _clinical_tables TEXT[] := ARRAY[
        'patient',
        'practitioner',
        'encounter',
        'condition',
        'observation',
        'allergy_intolerance',
        'procedure_record'
    ];
    _old_policies TEXT[] := ARRAY[
        'patient_tenant_policy',
        'practitioner_tenant_policy',
        'encounter_tenant_policy',
        'condition_tenant_policy',
        'observation_tenant_policy',
        'allergy_tenant_policy',
        'procedure_tenant_policy'
    ];
    _tbl TEXT;
    _old_policy TEXT;
    _new_policy TEXT;
    _i INT;
BEGIN
    FOR _i IN 1..array_length(_clinical_tables, 1)
    LOOP
        _tbl := _clinical_tables[_i];
        _old_policy := _old_policies[_i];
        _new_policy := _tbl || '_rls_policy';

        -- Drop old permissive policy
        EXECUTE format('DROP POLICY IF EXISTS %I ON %I', _old_policy, _tbl);
        -- Drop new policy too (idempotent re-run)
        EXECUTE format('DROP POLICY IF EXISTS %I ON %I', _new_policy, _tbl);

        -- Create hardened policy
        EXECUTE format(
            'CREATE POLICY %I ON %I
                FOR ALL TO ehr_app
                USING (rls_tenant_check())
                WITH CHECK (rls_tenant_check())',
            _new_policy, _tbl
        );
    END LOOP;
END $$;

-- ============================================================================
-- 6. HARDEN RLS ON AUDIT / COMPLIANCE TABLES
-- ============================================================================
-- audit_event and hipaa_access_log should be:
--   - INSERT: any authenticated ehr_app user (the app writes audit entries)
--   - SELECT: only users with 'admin' or 'compliance' role
--
-- Role information is passed via session variable app.current_user_roles
-- which the application sets as a comma-separated list.

CREATE OR REPLACE FUNCTION rls_has_audit_read_role() RETURNS boolean AS $$
DECLARE
    _roles TEXT;
BEGIN
    _roles := current_setting('app.current_user_roles', true);
    IF _roles IS NULL OR _roles = '' THEN
        RETURN FALSE;
    END IF;

    -- Check if the user has admin or compliance role
    RETURN _roles LIKE '%admin%'
        OR _roles LIKE '%compliance%'
        OR _roles LIKE '%auditor%';
END;
$$ LANGUAGE plpgsql STABLE;

-- ---- audit_event ----
DROP POLICY IF EXISTS audit_event_insert_policy ON audit_event;
DROP POLICY IF EXISTS audit_event_select_policy ON audit_event;
DROP POLICY IF EXISTS audit_event_rls_insert ON audit_event;
DROP POLICY IF EXISTS audit_event_rls_select ON audit_event;

-- Any ehr_app connection with a valid tenant may write audit entries
CREATE POLICY audit_event_rls_insert ON audit_event
    FOR INSERT TO ehr_app
    WITH CHECK (rls_tenant_check());

-- Only users with audit-read roles may query audit entries
CREATE POLICY audit_event_rls_select ON audit_event
    FOR SELECT TO ehr_app
    USING (rls_tenant_check() AND rls_has_audit_read_role());

-- ---- hipaa_access_log ----
DROP POLICY IF EXISTS hipaa_log_insert_policy ON hipaa_access_log;
DROP POLICY IF EXISTS hipaa_log_select_policy ON hipaa_access_log;
DROP POLICY IF EXISTS hipaa_log_rls_insert ON hipaa_access_log;
DROP POLICY IF EXISTS hipaa_log_rls_select ON hipaa_access_log;

-- Any ehr_app connection with a valid tenant may write HIPAA log entries
CREATE POLICY hipaa_log_rls_insert ON hipaa_access_log
    FOR INSERT TO ehr_app
    WITH CHECK (rls_tenant_check());

-- Only compliance/admin users may read the HIPAA access log
CREATE POLICY hipaa_log_rls_select ON hipaa_access_log
    FOR SELECT TO ehr_app
    USING (rls_tenant_check() AND rls_has_audit_read_role());

-- ============================================================================
-- 7. INDEXES ON AUDIT COLUMNS
-- ============================================================================
-- Support queries like "show me everything user X modified today".

DO $$
DECLARE
    _tables TEXT[] := ARRAY[
        'patient',
        'practitioner',
        'encounter',
        'condition',
        'observation',
        'allergy_intolerance',
        'procedure_record'
    ];
    _tbl TEXT;
BEGIN
    FOREACH _tbl IN ARRAY _tables
    LOOP
        EXECUTE format(
            'CREATE INDEX IF NOT EXISTS idx_%s_created_by ON %I (created_by) WHERE created_by IS NOT NULL',
            _tbl, _tbl
        );
        EXECUTE format(
            'CREATE INDEX IF NOT EXISTS idx_%s_updated_by ON %I (updated_by) WHERE updated_by IS NOT NULL',
            _tbl, _tbl
        );
    END LOOP;
END $$;

-- ============================================================================
-- DONE. Summary of changes:
--   - current_tenant_id() / current_user_id() helper functions
--   - rls_tenant_check() validates session var matches search_path
--   - rls_has_audit_read_role() gates audit table reads
--   - created_by / updated_by columns + triggers on 7 clinical tables
--   - Hardened RLS: clinical tables require tenant check, audit tables
--     additionally require admin/compliance/auditor role for reads
-- ============================================================================
