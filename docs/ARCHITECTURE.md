# Architecture

This document provides a detailed technical overview of the OpenEHR Server architecture, covering system design, domain-driven patterns, multi-tenancy, authentication, HIPAA compliance, and extensibility.

---

## Table of Contents

- [System Overview](#system-overview)
- [Domain-Driven Design](#domain-driven-design)
- [The 5-File Pattern](#the-5-file-pattern)
- [Multi-Tenancy Implementation](#multi-tenancy-implementation)
- [Authentication Flow](#authentication-flow)
- [HIPAA Compliance Features](#hipaa-compliance-features)
- [Extension Mechanisms](#extension-mechanisms)
- [Migration Strategy](#migration-strategy)

---

## System Overview

OpenEHR Server is a headless, API-first Electronic Health Record system written in Go. It uses the Echo web framework for HTTP routing and middleware, PostgreSQL 16 as the primary data store, and follows a layered architecture that separates concerns across well-defined boundaries.

### Key Design Decisions

- **Headless architecture.** The server has no built-in user interface. It exposes two API surfaces (FHIR R4 and operational REST) that any frontend can consume. This allows organizations to build custom UIs while relying on a standards-compliant backend.

- **Dual API surface.** The FHIR R4 API provides interoperability with external systems and SMART on FHIR applications. The operational REST API provides a richer, more ergonomic interface for internal UI development with pagination, nested resources, and domain-specific query parameters.

- **Schema-per-tenant isolation.** Each tenant receives a dedicated PostgreSQL schema. All request-scoped queries run against the tenant schema via `SET search_path`. This provides HIPAA-grade data isolation without the operational overhead of separate database instances.

- **Domain-driven decomposition.** The system is organized into 20 domains, each following an identical 5-file pattern. This makes the codebase predictable, easy to navigate, and straightforward to extend.

- **Repository interface pattern.** All database access flows through Go interfaces. Service layers depend on these interfaces rather than concrete PostgreSQL implementations. This enables in-memory mock testing without a database.

### Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Language | Go 1.22 | Server implementation |
| Web Framework | Echo v4 | HTTP routing, middleware |
| Database | PostgreSQL 16 | Primary data store |
| Database Driver | pgx v5 | PostgreSQL driver, connection pooling |
| Cache | Redis 7 | Session cache, rate limiting |
| Identity Provider | Keycloak 24 | OAuth2/OIDC (development default) |
| Configuration | Viper | Environment variable and file config |
| CLI | Cobra | CLI subcommands (serve, migrate, tenant) |
| Logging | zerolog | Structured JSON logging |
| Authentication | golang-jwt/jwt v5 | JWT parsing and validation |
| Container | Docker, Docker Compose | Deployment and development environment |
| Schema Management | Atlas | Database schema versioning |

### Runtime Architecture

```
                    +---------------------------+
                    |      Load Balancer        |
                    +------------+--------------+
                                 |
              +------------------+------------------+
              |                                     |
    +---------v---------+             +-------------v---------+
    |   EHR Server      |             |   EHR Server          |
    |   Instance 1      |             |   Instance N          |
    |   (Go/Echo)       |             |   (Go/Echo)           |
    +---------+---------+             +-------------+---------+
              |                                     |
              +------------------+------------------+
                                 |
              +------------------+------------------+
              |                                     |
    +---------v---------+             +-------------v---------+
    |   PostgreSQL 16   |             |      Redis 7          |
    |   (Multi-tenant)  |             |      (Cache)          |
    +-------------------+             +-----------------------+
              |
    +---------v---------+
    |   Keycloak / OIDC |
    |   Provider        |
    +-------------------+
```

---

## Domain-Driven Design

### Domain Organization

The 20 domains are organized into five tiers (plus cross-cutting infrastructure) based on clinical priority, dependency relationships, and deployment criticality:

**Tier 0 -- Core Infrastructure**

These domains must exist before any other domain can function:

- `admin` -- Organizations, departments, locations, system users. Provides the organizational hierarchy that all other domains reference.
- `identity` -- Patients and practitioners. The master person index. Every clinical domain references patient and practitioner records.
- `encounter` -- Patient visits and care episodes. Links patients to their clinical activities across all domains.

**Tier 1 -- Primary Clinical**

Core clinical data that most healthcare workflows depend on:

- `clinical` -- Conditions (diagnoses), observations (vitals, lab values), allergies/intolerances, procedures.
- `medication` -- Medication catalog, prescriptions, administration records, dispenses, medication statements.
- `diagnostics` -- Service requests (orders), specimen tracking, diagnostic reports, imaging studies.

**Tier 2 -- Operational Workflows**

Supporting systems that drive daily hospital operations:

- `scheduling` -- Provider schedules, time slots, appointments, waitlist management.
- `billing` -- Insurance coverage, claims, adjudication responses, invoices.
- `documents` -- Patient consents, document references, clinical notes, FHIR compositions.
- `inbox` -- In-basket messaging, co-sign requests, patient lists, clinical handoffs.

**Tier 3 -- Specialty Modules**

Department-specific clinical modules:

- `surgery` -- Operating room management, surgical cases, surgeon preference cards, implant tracking.
- `emergency` -- Triage (ESI), ED patient tracking, trauma activation management.
- `obstetrics` -- Pregnancy records, prenatal visits, labor monitoring, delivery, newborn, postpartum.
- `oncology` -- Cancer diagnosis staging, treatment protocols, chemotherapy, radiation, tumor markers, tumor boards.
- `nursing` -- Flowsheets, nursing assessments, fall risk, skin/pain assessments, lines/drains, restraints, I/O.

**Tier 4 -- Extended Modules**

Specialized systems that extend the core EHR:

- `behavioral` -- Psychiatric assessments, safety plans, legal holds, seclusion/restraint, group therapy.
- `research` -- Clinical trial protocols, patient enrollment, adverse events, protocol deviations.
- `portal` -- Patient portal accounts, portal messaging, questionnaires, check-in workflows.
- `cds` -- Clinical decision support rules, alerts, drug interactions, order sets, clinical pathways, formulary, medication reconciliation.

**Cross-Cutting Infrastructure**

- `subscription` -- FHIR R4 Subscription (rest-hook webhooks). Listens for resource mutations via VersionTracker events, evaluates criteria, delivers webhook notifications with retry. Admin-only.

### Domain Boundaries

Each domain is self-contained within `internal/domain/<name>/`. Domains do not import from other domains. Cross-domain references are handled through:

1. **UUID foreign keys** -- Domains reference entities in other domains by UUID, not by importing types.
2. **Service composition at the application layer** -- When a workflow spans multiple domains, the wiring happens in `main.go` where services from different domains can be composed.
3. **Shared platform packages** -- Common infrastructure (database connections, FHIR types, pagination) lives in `internal/platform/` and `pkg/`.

---

## The 5-File Pattern

Every domain follows an identical five-file structure. This pattern is the fundamental unit of code organization in the project.

### model.go

Defines the domain's data types as Go structs. Each struct maps to a database table via `db` struct tags and to JSON via `json` struct tags. Models that correspond to FHIR resources include a `ToFHIR()` method that converts the internal representation to a FHIR R4 JSON map.

Key conventions:

- Primary key is always `ID uuid.UUID`.
- FHIR-mapped resources include a `FHIRID string` field.
- Nullable fields use pointer types (`*string`, `*time.Time`, `*uuid.UUID`).
- Timestamps follow the pattern: `CreatedAt time.Time`, `UpdatedAt time.Time`.
- Sensitive fields (SSN, Aadhaar) are stored as hashes and excluded from JSON with `json:"-"`.

### repo.go

Defines one or more Go interfaces that describe the persistence contract for the domain. This is the boundary between business logic and data access.

Key conventions:

- Every interface method takes `context.Context` as its first parameter.
- `Create` methods accept a pointer to the model and populate the `ID` field on success.
- `List` methods return `([]*Model, int, error)` where the `int` is the total count for pagination.
- `Search` methods accept `params map[string]string` for flexible query parameters.

### repo_pg.go

Provides the PostgreSQL implementation of the repository interfaces defined in `repo.go`. All implementations follow the same pattern:

1. Accept a `*pgxpool.Pool` in the constructor.
2. Retrieve the tenant-scoped connection from context via `db.ConnFromContext(ctx)`.
3. Execute parameterized queries using pgx.
4. Scan results into model structs.

The tenant-scoped connection is critical. The tenant middleware sets `search_path` on the connection before the request reaches the handler, so all queries automatically target the correct schema.

### service.go

Contains business logic and validation. The service struct holds repository interfaces as fields, injected via the constructor function.

Key conventions:

- Validate required fields and return descriptive errors.
- Set default values (e.g., `Active = true` on creation).
- Never access the database directly; always go through the repository interface.
- Never access HTTP request/response objects; the handler layer handles HTTP concerns.

### handler.go

Translates between HTTP and the service layer. Registers routes and provides both operational REST and FHIR endpoint implementations.

Key conventions:

- `RegisterRoutes(api *echo.Group, fhirGroup *echo.Group)` registers all routes for the domain.
- Operational endpoints use `echo.NewHTTPError()` for errors.
- FHIR endpoints use `fhir.ErrorOutcome()` and `fhir.NotFoundOutcome()` to return FHIR OperationOutcome resources.
- List endpoints use `pagination.FromContext(c)` to extract limit/offset parameters.
- FHIR search endpoints return results wrapped in `fhir.NewSearchBundle()`.

---

## Multi-Tenancy Implementation

### Schema-Per-Tenant Model

Each tenant is assigned a PostgreSQL schema named `tenant_<identifier>`. The schema contains all domain tables for that tenant. A shared schema (`shared`) holds reference data common across tenants. The `public` schema holds infrastructure tables.

```
PostgreSQL Database: ehr
|
|-- public                    # Infrastructure tables (_migrations, etc.)
|-- shared                    # Reference data (code systems, value sets)
|-- tenant_default            # Default tenant schema
|   |-- patient
|   |-- practitioner
|   |-- encounter
|   |-- condition
|   |-- ... (200+ tables)
|-- tenant_acme               # Acme Corp tenant schema
|   |-- patient
|   |-- practitioner
|   |-- ... (identical structure)
|-- tenant_mercy              # Mercy Health tenant schema
|   |-- ...
```

### Tenant Resolution Flow

The tenant middleware (`db.TenantMiddleware`) runs on every request after authentication:

```
1. Extract tenant ID:
   a. Check echo context for "jwt_tenant_id" (set by auth middleware from JWT claim)
   b. Check X-Tenant-ID request header
   c. Check "tenant_id" query parameter
   d. Fall back to DEFAULT_TENANT config value

2. Validate tenant ID against pattern: ^[a-zA-Z0-9_]+$

3. Acquire a database connection from the pool

4. Execute: SET search_path TO tenant_<id>, shared, public

5. Store tenant ID and connection in request context

6. Pass control to the next middleware/handler

7. Release the connection when the request completes
```

### Tenant Provisioning

Creating a new tenant involves two steps:

1. **Create the schema** -- The `tenant create` CLI command calls `db.CreateTenantSchema()`, which executes `CREATE SCHEMA IF NOT EXISTS tenant_<name>`.

2. **Run migrations** -- The `scripts/migrate.sh` script applies all migration files against the new schema, creating the full table structure.

### Row-Level Security

In addition to schema isolation, migration `018_hipaa_row_level_security.sql` applies PostgreSQL Row-Level Security (RLS) policies as a defense-in-depth measure. RLS ensures that even if a query accidentally targets the wrong schema, the database itself enforces access boundaries.

---

## Authentication Flow

### Production Flow

```
Client                         EHR Server                    OIDC Provider
  |                               |                              |
  |  1. Obtain token              |                              |
  |------------------------------------------------------------->|
  |  <--- JWT (with tenant_id, roles, fhir_scopes) -------------|
  |                               |                              |
  |  2. API Request               |                              |
  |  Authorization: Bearer <jwt>  |                              |
  |------------------------------>|                              |
  |                               |  3. Fetch JWKS (cached)      |
  |                               |----------------------------->|
  |                               |  <--- Public keys -----------|
  |                               |                              |
  |                               |  4. Validate JWT             |
  |                               |     - Verify signature       |
  |                               |     - Check issuer           |
  |                               |     - Check audience         |
  |                               |     - Check expiration       |
  |                               |                              |
  |                               |  5. Extract claims           |
  |                               |     - sub -> user_id         |
  |                               |     - tenant_id -> schema    |
  |                               |     - roles -> RBAC          |
  |                               |     - fhir_scopes -> scopes  |
  |                               |                              |
  |  <--- API Response -----------|                              |
```

### OIDC Auto-Discovery

When only `AUTH_ISSUER` is configured (without an explicit `AUTH_JWKS_URL`), the server automatically discovers the JWKS endpoint by fetching:

```
{AUTH_ISSUER}/.well-known/openid-configuration
```

This works with any standards-compliant OIDC provider and eliminates the need to manually configure the JWKS URL.

### JWKS Key Caching

The JWKS keys are cached in memory with a 5-minute TTL. When a JWT arrives with a `kid` (key ID) that is not in the cache, the server fetches fresh keys from the JWKS endpoint. This handles key rotation automatically.

### Development Mode Bypass

When `ENV=development`, the `DevAuthMiddleware` is installed instead of the JWT middleware:

- Requests without an `Authorization` header receive default credentials (admin role, all scopes, default tenant).
- Requests with a token are still processed normally.
- This allows development without running an identity provider.

### RBAC Enforcement

Two middleware functions enforce access control at the route level:

- `auth.RequireRole(roles ...string)` -- Checks if the user has at least one of the specified roles. The `admin` role implicitly satisfies all role checks.
- `auth.RequireScope(resource, operation string)` -- Checks if the user's FHIR scopes cover the requested resource and operation. Supports wildcards (`user/*.*`, `patient/*.read`).

---

## HIPAA Compliance Features

### Audit Logging

The system provides two levels of audit logging:

**1. Request-Level Audit (Middleware)**

The audit middleware (`middleware.Audit`) logs every HTTP request with:

- Request ID, tenant ID
- HTTP method and path
- Remote IP address and user agent
- Response status code

These logs are structured JSON and can be shipped to any log aggregation system.

**2. FHIR AuditEvent (Database)**

The `hipaa.AuditLogger` writes detailed audit events to the `audit_event` table following the FHIR AuditEvent resource structure. Each event records:

- Event type and subtype (rest/read, rest/create, rest/delete, etc.)
- Action code (C/R/U/D/E)
- Outcome and description
- Agent information (who performed the action, their role, IP address)
- Entity information (what resource was accessed)
- Purpose of use (treatment, operations, emergency, etc.)
- Sensitivity label
- Session and user agent details

### PHI Access Logging

The `hipaa_access_log` table provides a focused log of all accesses to Protected Health Information:

- Which patient's data was accessed
- Who accessed it (ID, name, role)
- What resource type and ID
- Whether it was a break-glass override
- IP address and session context

### Break-Glass Support

The `AuditLogger.LogBreakGlass()` method handles emergency access overrides. When a clinician needs to access a patient's record outside their normal authorization (e.g., emergency treatment), the system:

1. Records the PHI access with `is_break_glass = true` and the stated reason
2. Creates a FHIR AuditEvent with type `emergency` and subtype `break-glass`
3. Marks the purpose of use as `ETREAT` (Emergency Treatment)
4. Sets the sensitivity label to `R` (Restricted)

All break-glass events are flagged for post-hoc review by compliance officers.

### PHI Field-Level Encryption

The `hipaa.PHIEncryptor` provides AES-256-GCM encryption for sensitive fields:

- Accepts a 32-byte encryption key
- Generates a random nonce for each encryption operation
- Prepends the nonce to the ciphertext
- Returns base64-encoded output for database storage
- Provides both string and byte-level APIs

This allows individual database columns containing PHI (names, SSNs, addresses) to be encrypted at rest while keeping non-sensitive fields in plaintext for querying.

### Row-Level Security

PostgreSQL Row-Level Security policies (applied in migration `018_hipaa_row_level_security.sql`) provide database-level enforcement of tenant isolation. Even if application-level schema resolution were bypassed, RLS policies would prevent cross-tenant data access.

---

## Extension Mechanisms

### Plugin Architecture

The `plugin.DomainPlugin` interface allows external packages to register new domains without modifying the core server:

```go
type DomainPlugin interface {
    Name() string
    RegisterRoutes(api *echo.Group, fhir *echo.Group)
    Migrate(ctx context.Context, pool *pgxpool.Pool) error
}
```

The `plugin.Registry` manages plugin lifecycle:

- `Register(p DomainPlugin)` -- Registers a plugin
- `RegisterRoutes(api, fhir)` -- Calls `RegisterRoutes` on all registered plugins
- `Migrate(ctx, pool)` -- Runs migrations for all registered plugins

This enables organizations to add custom domains (e.g., institution-specific workflows) as separate Go packages that are compiled into the server.

### Extension Tables

The migration system supports extension tables that add custom fields to existing domains. By adding new migration files that reference existing schemas, organizations can extend the data model without modifying core domain code.

### Custom FHIR Resources

Because the FHIR layer uses `map[string]interface{}` for resource representation, plugins can define custom FHIR resource mappings for non-standard resources or extensions.

### FHIR Operations

The platform supports FHIR-defined operations as platform-level handlers in `internal/platform/fhir/`:

- **Patient/$everything** (`GET /fhir/Patient/:id/$everything`) -- Returns all data for a patient in a single searchset Bundle. Uses a registered-fetcher pattern where each domain registers a `PatientResourceFetcher` function. Supports `_type` (comma-separated resource filter) and `_count` (per-type limit) query parameters. Covers all 29 Patient Compartment resource types.

- **$export** (`POST /fhir/$export`, `POST /fhir/Patient/$export`, `POST /fhir/Patient/:id/$export`, `POST /fhir/Group/:id/$export`) -- Asynchronous bulk data export per the FHIR Bulk Data Access IG. Supports system-level, patient-level, and group-level exports with NDJSON output. Features include `_outputFormat` validation, `Retry-After` headers, `_typeFilter` parameter, progress tracking (`X-Progress`), `requiresAccessToken: true`, max concurrent job limits (429 with Retry-After), automatic job expiration/cleanup, and 29 registered resource type exporters. Uses the `ExportManager` with registered `ResourceExporter` implementations.

- **CDS Hooks** (`GET /cds-services`, `POST /cds-services/:id`, `POST /cds-services/:id/feedback`) -- HL7 CDS Hooks 2.0 interface. Uses a `CDSHooksHandler` with registered `ServiceHandler` functions per hook service. Ships with three built-in services: patient-risk-alerts (patient-view hook), drug-interaction-check (order-select hook), and formulary-check (order-select hook). Routes are registered at the root level (not under /fhir) per the CDS Hooks specification.

- **$validate** (`POST /fhir/$validate`, `POST /fhir/:resourceType/$validate`) -- FHIR resource validation against structure rules, required fields, status values, reference formats, date formats, and business rules (e.g. Patient must have name or identifier, MedicationRequest must have medication[x]). Supports `mode` query parameter (create, update, delete) and returns OperationOutcome with FHIRPath-style issue locations. Validates 30+ registered FHIR R4 resource types.

- **C-CDA 2.1 Generation & Parsing** (`GET /api/v1/patients/:id/ccd`, `POST /api/v1/ccda/parse`) -- Produces and consumes Consolidated Clinical Document Architecture (C-CDA) 2.1 Continuity of Care Documents. The `Generator` maps FHIR resources to 10 CDA sections (Allergies, Medications, Problems, Procedures, Results, Vital Signs, Immunizations, Social History, Plan of Care, Encounters) with proper OIDs, LOINC codes, and human-readable narrative HTML. The `Parser` extracts structured data from incoming C-CDA XML. Located in `internal/platform/ccda/`.

- **SMART on FHIR App Launch v2.0** (`GET /auth/authorize`, `POST /auth/token`, `POST /auth/register`, `POST /auth/launch`, `POST /auth/introspect`, `GET /.well-known/smart-configuration`) -- Full OAuth2 authorization server for SMART app launch. Supports EHR launch (with launch context) and standalone launch, authorization code flow with PKCE (S256 required for public clients), dynamic client registration, refresh tokens, token introspection, and scope negotiation. JWT access tokens include SMART launch context claims (patient, encounter, fhirUser). Located in `internal/platform/auth/smart_launch.go`.

- **HL7v2 Interface Engine** (`POST /api/v1/hl7v2/parse`, `POST /api/v1/hl7v2/generate/adt`, `POST /api/v1/hl7v2/generate/orm`, `POST /api/v1/hl7v2/generate/oru`) -- Parse and generate HL7v2 messages for hospital integrations. Supports ADT (A01 admit, A02 transfer, A03 discharge, A04 register, A08 update), ORM (O01 new order), and ORU (R01 observation result) message types. Full parser handles field separators, components (^), repetitions (~), and HL7 escape sequences. Generator converts FHIR Patient, Encounter, ServiceRequest, DiagnosticReport, and Observation resources to HL7v2 pipe-delimited format. Located in `internal/platform/hl7v2/`.

- **Patient/$match** (`POST /fhir/Patient/$match`) -- Probabilistic patient matching using weighted multi-field scoring. Accepts a FHIR Parameters resource containing a Patient resource and returns a scored Bundle with match-grade extensions (certain/probable/possible). Implements Jaro-Winkler string similarity for fuzzy name matching. Configurable weights across 9 fields: last name (0.15), first name (0.15), birth date (0.20), gender (0.05), MRN (0.20), phone (0.10), email (0.05), address (0.05), SSN last-4 (0.05). Located in `internal/platform/fhir/match_op.go`.

- **ConceptMap/$translate** (`GET/POST /fhir/ConceptMap/$translate`, `GET /fhir/ConceptMap/:id/$translate`, `GET /fhir/ConceptMap`) -- Code system translation between clinical terminologies. Ships with 3 built-in concept maps: SNOMED CT → ICD-10-CM (15 common conditions), ICD-10-CM → SNOMED CT (reverse), LOINC → SNOMED CT (10 lab tests). Returns FHIR Parameters resources with equivalence classification. Supports lookup by source/target system pair or specific ConceptMap URL/ID. Located in `internal/platform/fhir/translate_op.go`.

- **CodeSystem/$subsumes** (`GET/POST /fhir/CodeSystem/$subsumes`) -- Hierarchical subsumption testing between codes within a code system. Supports SNOMED CT (4 clinical hierarchies: diabetes, hypertension, heart disease, respiratory with transitive ancestor walking) and ICD-10 (prefix-based subsumption). Returns outcome: subsumes, subsumed-by, equivalent, or not-subsumed. Located in `internal/platform/fhir/subsumes_op.go`.

- **ValueSet/$validate-code** (`GET/POST /fhir/ValueSet/$validate-code`) -- Validates whether a code is a member of a specified value set. Ships with 10 built-in FHIR R4 required value sets (observation-status, condition-clinical, administrative-gender, encounter-status, medication-request-status, procedure-status, diagnostic-report-status, immunization-status, allergy-intolerance-clinical, care-plan-status). Supports optional system filtering. Located in `internal/platform/fhir/valueset_validate_op.go`.

- **Composition/$document** (`GET /fhir/Composition/:id/$document`, `POST /fhir/Composition/$document`) -- Generates complete FHIR Document Bundles from Composition resources. Walks all references (subject, author, custodian, encounter, attester, section entries) and resolves them into Bundle entries. Composition is always the first entry per FHIR spec. Handles nested sections and deduplicates references. Located in `internal/platform/fhir/document_op.go`.

- **Advanced Search: _has and _filter** -- Library for parsing and SQL generation of advanced FHIR search parameters. `_has` (reverse chaining) generates EXISTS subqueries for finding resources referenced by other resources (e.g., find Patients with specific Observations). `_filter` parses structured filter expressions (eq, ne, gt, lt, ge, le, co, sw, ew operators with and/or combiners) and generates parameterized PostgreSQL WHERE clauses. Located in `internal/platform/fhir/search_advanced.go`.

### Real-Time Event System

The VersionTracker (used by all domain services for version history) supports a listener pattern via `ResourceEventListener`. The `NotificationEngine` registers as a listener and evaluates resource mutations against active FHIR Subscription criteria. Matching events produce notification rows in a PostgreSQL queue, which a background delivery loop POSTs to configured webhook endpoints with retry.

---

## Migration Strategy

### Migration File Convention

Migration files are stored in `migrations/` and follow the naming pattern:

```
NNN_tX_description.sql
```

Where:

- `NNN` -- Three-digit sequence number (001, 002, ...)
- `tX` -- Tier designation (t0, t1, t2, t3, t4)
- `description` -- Human-readable name

Current migration files:

| File | Description |
|------|-------------|
| `001_t0_core_tables.sql` | Admin, identity, encounter tables |
| `002_t1_clinical_tables.sql` | Conditions, observations, allergies, procedures |
| `003_t1_medication_tables.sql` | Medication lifecycle tables |
| `004_t1_diagnostics_tables.sql` | Service requests, specimens, reports, imaging |
| `005_t2_scheduling_tables.sql` | Schedules, slots, appointments, waitlists |
| `006_t2_billing_tables.sql` | Coverage, claims, responses, invoices |
| `007_t2_documents_tables.sql` | Consents, documents, notes, compositions |
| `008_t2_inbox_tables.sql` | Messaging, co-sign, handoffs |
| `009_t3_surgery_tables.sql` | OR rooms, surgical cases, implants |
| `010_t3_emergency_tables.sql` | Triage, ED tracking, trauma |
| `011_t3_obstetrics_tables.sql` | Pregnancy, labor, delivery, newborn |
| `012_t3_oncology_tables.sql` | Cancer diagnosis, chemo, radiation, tumor boards |
| `013_t3_nursing_tables.sql` | Flowsheets, assessments, I/O, restraints |
| `014_t4_behavioral_tables.sql` | Psych assessments, safety plans, legal holds |
| `015_t4_research_tables.sql` | Studies, enrollment, adverse events |
| `016_t4_portal_tables.sql` | Portal accounts, questionnaires, check-in |
| `017_t4_cds_tables.sql` | CDS rules, alerts, order sets, pathways |
| `018_hipaa_row_level_security.sql` | Row-level security policies |
| `022_subscription_tables.sql` | Subscription and notification tables |

### Migration Runner

The `db.Migrator` handles schema migrations:

1. **Loading** -- Reads all `.sql` files from the migrations directory, parses version numbers from filename prefixes, and sorts by version.

2. **Tracking** -- Each schema has a `_migrations` table that records which versions have been applied and when.

3. **Applying** -- Each migration runs in its own database transaction. If a migration fails, the transaction rolls back and no partial state is left behind.

4. **Per-schema execution** -- Migrations are applied per-tenant schema. The migrator sets `search_path` to the target schema before running each migration, so all `CREATE TABLE` statements create tables in the correct schema.

### Adding a New Migration

1. Create a new `.sql` file with the next available sequence number.
2. Use the appropriate tier prefix in the filename.
3. Write standard DDL statements (they will run within the tenant schema context).
4. Apply the migration with `make migrate-up` or `./scripts/migrate.sh <tenant>`.

### Atlas Integration

The `atlas.hcl` configuration file at the project root enables use of the Atlas schema management tool for more advanced migration workflows including schema diffing, lint checks, and declarative schema management.
