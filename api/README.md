# OpenEHR Server

**A headless, API-first Electronic Health Record system built in Go.**

OpenEHR Server provides a complete clinical data platform with dual REST APIs: a standards-compliant FHIR R4 interface for interoperability and an operational REST API for internal UI consumption. It is designed for multi-tenant deployments, HIPAA-grade security, and extensibility across 20 domains.

![Go Version](https://img.shields.io/badge/Go-1.22-blue)
![License](https://img.shields.io/badge/License-Apache%202.0-green)
![Build Status](https://img.shields.io/badge/Build-passing-brightgreen)
![FHIR](https://img.shields.io/badge/FHIR-R4-orange)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue)

---

## Features

- **29 domains** covering 200+ database tables (identity, encounter, clinical, medication, diagnostics, scheduling, billing, documents, inbox, surgery, emergency, obstetrics, oncology, nursing, behavioral, research, portal, admin, CDS, subscription, careplan, careteam, device, immunization, familyhistory, relatedperson, provenance, task, terminology)
- **FHIR R4 REST API** with 35+ resource types and full CRUD, search, history, patch, transaction bundles
- **FHIR $validate** operation for resource validation against structure rules, required fields, and business rules
- **C-CDA 2.1 Generation & Parsing** — produce and consume Continuity of Care Documents (10 clinical sections)
- **SMART on FHIR App Launch v2.0** — full OAuth2 authorization server with EHR launch, standalone launch, PKCE, dynamic client registration
- **FHIR R4 Subscriptions** with REST-hook webhook delivery, criteria matching, retry with exponential backoff
- **Patient/$everything** aggregating 28 resource types from the patient compartment
- **Bulk Data Export** ($export) with 29 resource type exporters, job limits, expiration, progress tracking
- **HL7v2 Interface Engine** — parse and generate ADT (A01-A08), ORM (O01), ORU (R01) messages with FHIR conversion
- **Patient/$match** — probabilistic patient matching with Jaro-Winkler similarity scoring and configurable weights
- **ConceptMap/$translate** — code system translation (SNOMED↔ICD-10, LOINC→SNOMED) with 3 built-in concept maps
- **CodeSystem/$subsumes** — hierarchical subsumption testing for SNOMED CT and ICD-10 code systems
- **ValueSet/$validate-code** — code membership validation against 10 built-in FHIR R4 required value sets
- **Composition/$document** — generate complete FHIR Document Bundles from Compositions with reference resolution
- **_has and _filter** — advanced FHIR search: reverse chaining (_has) and filter expressions (_filter) with SQL generation
- **$process-message** — FHIR Message Bundle processing with dispatch to registered event handlers
- **HL7v2 MLLP Listener** — TCP server for receiving HL7v2 messages over Minimal Lower Layer Protocol
- **Patient Self-Scheduling** — slot search, booking with double-booking prevention, cancellation, and appointment management
- **WebSocket Real-time Updates** — live patient record change notifications with topic-based subscriptions
- **Email/SMS Notifications** — template-based notification system with retry logic and delivery tracking
- **Document/Blob Storage** — S3-compatible file storage for clinical images, scans, and attachments with SHA-256 integrity
- **HTTP Cache/ETag Support** — response caching, conditional requests (If-None-Match, If-Modified-Since), and cache-control headers
- **Audit Trail Search/Export** — query audit logs by date/user/action with CSV and JSON export for compliance reviews
- **FHIR Bulk Import/Edit** — batch import via NDJSON, bulk update/delete with criteria matching, and job tracking
- **FHIR $graphql** — GraphQL query interface for FHIR resources with field selection
- **CodeSystem/$closure** — transitive closure table management for SNOMED CT hierarchies
- **API Key Management** — create, revoke, rotate API keys with SHA-256 hashing and per-client scoping
- **Per-Client Rate Limiting** — tiered rate plans (Free/Starter/Professional/Enterprise) with minute/hour/day limits
- **SMART Backend Services** — client_credentials grant with RS384 JWT assertion, JTI replay protection
- **Webhook Management API** — register endpoints, test connectivity, view delivery logs, retry failures, HMAC-SHA256 signatures
- **API Usage Analytics** — per-endpoint/client/resource metrics, time-series data, P95 latency tracking
- **Sandbox & Synthetic Data** — generate realistic FHIR patients with encounters, observations, conditions, medications
- **Detailed CapabilityStatement** — 38 resource types with search params, 12 operations, custom search parameter API
- **Operational REST API** for internal UI consumption with full CRUD, pagination, and search
- **Schema-per-tenant multi-tenancy** providing HIPAA-grade data isolation via PostgreSQL schemas
- **OAuth2 / SMART on FHIR authentication** compatible with Keycloak, Auth0, Okta, and Azure AD
- **Role-based access control** with 10 roles and per-domain permission enforcement
- **AES-256-GCM field-level encryption** for Protected Health Information (PHI)
- **Immutable HIPAA audit trail** with FHIR AuditEvent logging and break-glass support
- **PostgreSQL Row-Level Security** for defense-in-depth tenant isolation
- **Docker Compose** for instant development environment (PostgreSQL 16, Redis 7, Keycloak 24)
- **Plugin architecture** for extending the system with custom domains

---

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for local development)

### Running with Docker Compose

```bash
git clone https://github.com/openehr/ehr.git && cd ehr

# Copy the environment template
cp .env.example .env

# Start all services (PostgreSQL, Redis, Keycloak, EHR server)
docker compose up -d

# Wait for services to become healthy (about 30 seconds)
sleep 30

# Verify the server is running
curl http://localhost:8000/health

# View the FHIR CapabilityStatement
curl http://localhost:8000/fhir/metadata
```

### Running Locally

```bash
# Copy the environment template
cp .env.example .env

# Start infrastructure only
docker compose up -d postgres redis keycloak

# Build and run the server
make run

# Or build separately
make build
./bin/ehr-server serve
```

### Creating a Tenant

```bash
# Using the CLI
./bin/ehr-server tenant create --name acme

# Using Make
make tenant-create NAME=acme

# Run migrations for the new tenant
./scripts/migrate.sh acme
```

---

## Architecture

### Request Flow

Every API request flows through a layered middleware chain before reaching the domain handler:

```
Client Request
      |
      v
+------------------+
|   Echo Router    |
+------------------+
      |
      v
+------------------+
| Recovery MW      |  Panic recovery, error formatting
+------------------+
      |
      v
+------------------+
| Request ID MW    |  X-Request-ID generation/propagation
+------------------+
      |
      v
+------------------+
| Logger MW        |  Structured request logging (zerolog)
+------------------+
      |
      v
+------------------+
| CORS MW          |  Cross-origin resource sharing
+------------------+
      |
      v
+------------------+
| Auth MW          |  JWT validation (JWKS) or dev-mode bypass
+------------------+
      |
      v
+------------------+
| Tenant MW        |  Schema resolution (JWT -> header -> query -> default)
+------------------+
      |
      v
+------------------+
| Audit MW         |  HIPAA audit event logging
+------------------+
      |
      v
+------------------+
| Domain Handler   |  Route handler (e.g., identity.CreatePatient)
+------------------+
      |
      v
+------------------+
| Domain Service   |  Business logic, validation
+------------------+
      |
      v
+------------------+
| Repository (PG)  |  PostgreSQL queries (tenant-scoped via search_path)
+------------------+
      |
      v
+------------------+
|  PostgreSQL 16   |  Schema-per-tenant data storage
+------------------+
```

### Domain Tiers

Domains are organized into tiers based on clinical priority and dependency order. Migrations follow this ordering:

| Tier | Domains | Description |
|------|---------|-------------|
| **T0** | admin, identity, encounter | Core infrastructure: organizations, patients, practitioners, encounters |
| **T1** | clinical, medication, diagnostics | Primary clinical data: conditions, observations, allergies, medications, lab orders |
| **T2** | scheduling, billing, documents, inbox | Operational workflows: appointments, claims, clinical notes, messaging |
| **T3** | surgery, emergency, obstetrics, oncology, nursing | Specialty modules: OR management, ED tracking, labor/delivery, chemo, nursing assessments |
| **T4** | behavioral, research, portal, cds | Extended modules: psychiatric care, clinical trials, patient portal, decision support |

### Infrastructure Components

```
+-------------+     +-------------+     +--------------+
|  EHR Server |---->| PostgreSQL  |     |   Keycloak   |
|  (Go/Echo)  |     |     16      |     |   (OIDC)     |
|  :8000      |     |   :5433     |     |   :8080      |
+-------------+     +-------------+     +--------------+
      |
      v
+-------------+
|    Redis    |
|     7       |
|   :6380     |
+-------------+
```

---

## API Reference

The server exposes two parallel API surfaces: FHIR R4 endpoints for interoperability and operational REST endpoints for internal use.

### Health Check

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Server health status |

### FHIR R4 Endpoints

All FHIR endpoints are prefixed with `/fhir` and return FHIR R4 JSON.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/metadata` | CapabilityStatement |
| GET | `/fhir/.well-known/smart-configuration` | SMART on FHIR discovery |

**Identity**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/Patient` | Search patients |
| GET | `/fhir/Patient/:id` | Read patient by FHIR ID |
| POST | `/fhir/Patient` | Create patient |
| PUT | `/fhir/Patient/:id` | Update patient |
| DELETE | `/fhir/Patient/:id` | Delete patient |
| GET | `/fhir/Patient/:id/$everything` | Patient $everything |
| GET | `/fhir/Practitioner` | Search practitioners |
| GET | `/fhir/Practitioner/:id` | Read practitioner |
| POST | `/fhir/Practitioner` | Create practitioner |
| PUT | `/fhir/Practitioner/:id` | Update practitioner |
| DELETE | `/fhir/Practitioner/:id` | Delete practitioner |

**Admin**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/Organization` | Search organizations |
| GET | `/fhir/Organization/:id` | Read organization |
| POST | `/fhir/Organization` | Create organization |
| GET | `/fhir/Location` | Search locations |
| GET | `/fhir/Location/:id` | Read location |
| POST | `/fhir/Location` | Create location |

**Encounter**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/Encounter` | Search encounters |
| GET | `/fhir/Encounter/:id` | Read encounter |
| POST | `/fhir/Encounter` | Create encounter |
| PUT | `/fhir/Encounter/:id` | Update encounter |
| DELETE | `/fhir/Encounter/:id` | Delete encounter |

**Clinical**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/Condition` | Search conditions |
| GET | `/fhir/Condition/:id` | Read condition |
| POST | `/fhir/Condition` | Create condition |
| GET | `/fhir/Observation` | Search observations |
| GET | `/fhir/Observation/:id` | Read observation |
| POST | `/fhir/Observation` | Create observation |
| GET | `/fhir/AllergyIntolerance` | Search allergies |
| GET | `/fhir/Procedure` | Search procedures |

**Medication**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/Medication` | Search medications |
| GET | `/fhir/MedicationRequest` | Search prescriptions |
| POST | `/fhir/MedicationRequest` | Create prescription |
| GET | `/fhir/MedicationAdministration` | Search administrations |
| GET | `/fhir/MedicationDispense` | Search dispenses |

**Diagnostics**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/ServiceRequest` | Search service requests |
| POST | `/fhir/ServiceRequest` | Create service request |
| GET | `/fhir/DiagnosticReport` | Search diagnostic reports |
| GET | `/fhir/Specimen` | Search specimens |
| GET | `/fhir/ImagingStudy` | Search imaging studies |

**Scheduling**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/Appointment` | Search appointments |
| POST | `/fhir/Appointment` | Create appointment |
| GET | `/fhir/Schedule` | Search schedules |
| GET | `/fhir/Slot` | Search slots |

**Billing**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/Coverage` | Search coverage |
| GET | `/fhir/Claim` | Search claims |
| POST | `/fhir/Claim` | Create claim |

**Documents**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/Consent` | Search consent records |
| POST | `/fhir/Consent` | Create consent |
| GET | `/fhir/DocumentReference` | Search document references |
| GET | `/fhir/Composition` | Search compositions |

**Messaging**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/Communication` | Search communications |
| POST | `/fhir/Communication` | Create communication |

**Subscription** (Admin only)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/Subscription` | Search subscriptions |
| GET | `/fhir/Subscription/:id` | Read subscription |
| POST | `/fhir/Subscription` | Create subscription |
| PUT | `/fhir/Subscription/:id` | Update subscription |
| DELETE | `/fhir/Subscription/:id` | Delete subscription |
| PATCH | `/fhir/Subscription/:id` | Patch subscription |

**Research**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/ResearchStudy` | Search research studies |
| POST | `/fhir/ResearchStudy` | Create research study |

**Portal**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/Questionnaire` | Search questionnaires |
| GET | `/fhir/QuestionnaireResponse` | Search responses |
| POST | `/fhir/QuestionnaireResponse` | Create response |

**CDS Hooks** (HL7 CDS Hooks 2.0)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/cds-services` | Discovery -- list available CDS services |
| POST | `/cds-services/:id` | Invoke a CDS hook |
| POST | `/cds-services/:id/feedback` | Submit feedback on cards |

**Bulk Data Export** (FHIR Bulk Data Access IG)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/fhir/$export` | System-level bulk export kick-off |
| POST | `/fhir/Patient/$export` | Patient-level bulk export (all patients) |
| POST | `/fhir/Patient/:id/$export` | Patient-level bulk export by patient ID |
| POST | `/fhir/Group/:id/$export` | Group-level bulk export |
| GET | `/fhir/$export-status/:id` | Poll export job status |
| GET | `/fhir/$export-data/:id/:type` | Download NDJSON export data |
| DELETE | `/fhir/$export-status/:id` | Cancel/delete export job |

**FHIR $validate** (Resource Validation)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/fhir/$validate` | Validate a FHIR resource (type from body) |
| POST | `/fhir/:resourceType/$validate` | Validate a resource against type rules |

**C-CDA Generation & Parsing** (Continuity of Care Documents)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/patients/:id/ccd` | Generate CCD XML for a patient |
| POST | `/api/v1/ccda/parse` | Parse an incoming C-CDA document |

**SMART on FHIR App Launch** (OAuth2 Authorization Server)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/auth/authorize` | OAuth2 authorization endpoint |
| POST | `/auth/token` | Token endpoint (auth code + refresh grants) |
| POST | `/auth/register` | Dynamic client registration |
| POST | `/auth/launch` | Create EHR launch context |
| POST | `/auth/introspect` | Token introspection |
| GET | `/.well-known/smart-configuration` | SMART discovery document |

### Operational REST Endpoints

All operational endpoints are prefixed with `/api/v1` and return standard JSON with pagination.

**Identity** -- `/api/v1/patients`, `/api/v1/practitioners`

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/patients` | Create patient |
| GET | `/api/v1/patients` | List patients (paginated) |
| GET | `/api/v1/patients/:id` | Get patient by ID |
| PUT | `/api/v1/patients/:id` | Update patient |
| DELETE | `/api/v1/patients/:id` | Delete patient |
| POST | `/api/v1/patients/:id/contacts` | Add patient contact |
| GET | `/api/v1/patients/:id/contacts` | List patient contacts |
| DELETE | `/api/v1/patients/:id/contacts/:contact_id` | Remove patient contact |
| POST | `/api/v1/patients/:id/identifiers` | Add patient identifier |
| GET | `/api/v1/patients/:id/identifiers` | List patient identifiers |
| POST | `/api/v1/practitioners` | Create practitioner |
| GET | `/api/v1/practitioners` | List practitioners |
| GET | `/api/v1/practitioners/:id` | Get practitioner |
| PUT | `/api/v1/practitioners/:id` | Update practitioner |
| DELETE | `/api/v1/practitioners/:id` | Delete practitioner |
| POST | `/api/v1/practitioners/:id/roles` | Add practitioner role |
| GET | `/api/v1/practitioners/:id/roles` | List practitioner roles |

**Admin** -- `/api/v1/organizations`, `/api/v1/departments`, `/api/v1/locations`, `/api/v1/system-users`

**Encounter** -- `/api/v1/encounters`

**Clinical** -- `/api/v1/conditions`, `/api/v1/observations`, `/api/v1/allergies`, `/api/v1/procedures`

**Medication** -- `/api/v1/medications`, `/api/v1/medication-requests`, `/api/v1/medication-administrations`, `/api/v1/medication-dispenses`, `/api/v1/medication-statements`

**Diagnostics** -- `/api/v1/service-requests`, `/api/v1/specimens`, `/api/v1/diagnostic-reports`, `/api/v1/imaging-studies`

**Scheduling** -- `/api/v1/schedules`, `/api/v1/slots`, `/api/v1/appointments`, `/api/v1/waitlists`

**Billing** -- `/api/v1/coverages`, `/api/v1/claims`, `/api/v1/claim-responses`, `/api/v1/invoices`

**Documents** -- `/api/v1/consents`, `/api/v1/document-references`, `/api/v1/clinical-notes`, `/api/v1/compositions`

**Inbox** -- `/api/v1/message-pools`, `/api/v1/inbox-messages`, `/api/v1/cosign-requests`, `/api/v1/patient-lists`, `/api/v1/handoffs`

**Surgery** -- `/api/v1/or-rooms`, `/api/v1/surgical-cases`, `/api/v1/preference-cards`, `/api/v1/implant-logs`

**Emergency** -- `/api/v1/triage`, `/api/v1/ed-tracking`, `/api/v1/trauma`

**Obstetrics** -- `/api/v1/pregnancies`, `/api/v1/prenatal-visits`, `/api/v1/labor`, `/api/v1/deliveries`, `/api/v1/newborns`, `/api/v1/postpartum`

**Oncology** -- `/api/v1/cancer-diagnoses`, `/api/v1/treatment-protocols`, `/api/v1/chemo-cycles`, `/api/v1/radiation-therapy`, `/api/v1/tumor-markers`, `/api/v1/tumor-boards`

**Nursing** -- `/api/v1/flowsheet-templates`, `/api/v1/flowsheet-entries`, `/api/v1/nursing-assessments`, `/api/v1/fall-risks`, `/api/v1/skin-assessments`, `/api/v1/pain-assessments`, `/api/v1/lines-drains`, `/api/v1/restraints`, `/api/v1/intake-output`

**Behavioral Health** -- `/api/v1/psych-assessments`, `/api/v1/safety-plans`, `/api/v1/legal-holds`, `/api/v1/seclusion-restraints`, `/api/v1/group-therapy`

**Research** -- `/api/v1/studies`, `/api/v1/enrollments`, `/api/v1/adverse-events`, `/api/v1/deviations`

**Portal** -- `/api/v1/portal-accounts`, `/api/v1/portal-messages`, `/api/v1/questionnaires`, `/api/v1/questionnaire-responses`, `/api/v1/patient-checkins`

**CDS** -- `/api/v1/cds-rules`, `/api/v1/cds-alerts`, `/api/v1/drug-interactions`, `/api/v1/order-sets`, `/api/v1/clinical-pathways`, `/api/v1/pathway-enrollments`, `/api/v1/formulary`, `/api/v1/med-reconciliations`

**Subscription** -- `/api/v1/subscriptions`, `/api/v1/subscriptions/:id/notifications`

---

## Authentication

### Overview

OpenEHR Server uses OAuth2 with OpenID Connect (OIDC) for authentication. It supports any OIDC-compliant identity provider.

### Supported Providers

- **Keycloak** (included in Docker Compose for development)
- **Auth0**
- **Okta**
- **Azure AD**
- Any provider with a standard `/.well-known/openid-configuration` discovery endpoint

### Development Mode

When `ENV=development` (the default), the server runs in **dev mode**:

- No token is required for API requests
- Unauthenticated requests receive default credentials:
  - User ID: `dev-user`
  - Roles: `["admin"]`
  - Scopes: `["user/*.*"]`
  - Tenant: `default`
- If a Bearer token is provided, it is still accepted

### Production Mode

When `ENV=production`, the server requires a valid JWT Bearer token on every request:

```
Authorization: Bearer <jwt-token>
```

The token is validated against the JWKS endpoint discovered from the configured issuer. The following JWT claims are extracted:

| Claim | Purpose |
|-------|---------|
| `sub` | User ID |
| `tenant_id` | Tenant identifier (used for schema resolution) |
| `roles` | Array of role names for RBAC enforcement |
| `fhir_scopes` | Array of SMART on FHIR scopes |

### SMART on FHIR

The server implements SMART on FHIR discovery at:

```
GET /fhir/.well-known/smart-configuration
```

Supported scopes include:

- `openid`, `profile`, `fhirUser`
- `launch`, `launch/patient`, `launch/encounter`
- `patient/*.read`, `patient/*.write`
- `user/*.read`, `user/*.write`

Supported capabilities:

- `launch-ehr`, `launch-standalone`
- `client-public`, `client-confidential-symmetric`
- `context-ehr-patient`, `context-ehr-encounter`
- `context-standalone-patient`
- `permission-offline`, `permission-patient`, `permission-user`
- `sso-openid-connect`

### HL7v2 Interface Engine

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/hl7v2/parse` | Parse HL7v2 message to JSON |
| POST | `/api/v1/hl7v2/generate/adt` | Generate ADT message from FHIR Patient + Encounter |
| POST | `/api/v1/hl7v2/generate/orm` | Generate ORM message from FHIR ServiceRequest |
| POST | `/api/v1/hl7v2/generate/oru` | Generate ORU message from FHIR DiagnosticReport + Observations |

Supported ADT events: A01 (Admit), A02 (Transfer), A03 (Discharge), A04 (Register), A08 (Update).

### Patient $match

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/fhir/Patient/$match` | Probabilistic patient matching |

Accepts a FHIR Parameters resource with a Patient resource and returns a scored Bundle. Matching uses Jaro-Winkler similarity with configurable weights across 9 fields (name, DOB, gender, MRN, phone, email, address, SSN).

### ConceptMap $translate

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/fhir/ConceptMap` | List available concept maps |
| GET | `/fhir/ConceptMap/$translate` | Translate code (query params) |
| POST | `/fhir/ConceptMap/$translate` | Translate code (Parameters body) |
| GET | `/fhir/ConceptMap/:id/$translate` | Translate using specific map |

Built-in maps: SNOMED CT ↔ ICD-10-CM (15 conditions), LOINC → SNOMED CT (10 lab tests).

### FHIR $process-message

| Method | Path | Description |
|--------|------|-------------|
| POST | `/fhir/$process-message` | Process a FHIR Message Bundle |

Built-in event handlers: `notification`, `patient-link`, `diagnostic-report`. Custom handlers can be registered at startup.

### HL7v2 MLLP TCP Listener

Set `MLLP_ADDR` environment variable (e.g., `MLLP_ADDR=:2575`) to start the MLLP TCP listener. Receives HL7v2 messages over the Minimal Lower Layer Protocol with automatic ACK generation.

### Patient Self-Scheduling

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/scheduling/slots` | Search available slots (params: `start`, `end`, `service_type`, `schedule_id`) |
| POST | `/api/v1/scheduling/book` | Book an appointment (body: `slot_id`, `patient_id`, `reason`) |
| POST | `/api/v1/scheduling/cancel/:id` | Cancel appointment (body: `patient_id`, `reason`) |
| GET | `/api/v1/scheduling/appointments` | List patient appointments (params: `patient_id`, `status`, `limit`) |
| GET | `/api/v1/scheduling/appointments/:id` | Get appointment by ID (params: `patient_id`) |

Double-booking prevention with slot-level locking. Supports date range search, service type filtering, and automatic slot freeing on cancellation.

### WebSocket Real-time Updates

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/ws` | Upgrade to WebSocket connection for real-time events |

Subscribe to topics (e.g., `Patient/123`, `Encounter/*`) via JSON messages. Supports broadcast by topic and global broadcast.

### Email/SMS Notifications

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/notifications/send` | Send a notification (email or SMS) |
| POST | `/api/v1/notifications/send-template` | Send from a named template |
| GET | `/api/v1/notifications/:id` | Get notification status |
| GET | `/api/v1/notifications` | List notifications by recipient |
| POST | `/api/v1/notifications/:id/retry` | Retry a failed notification |
| GET | `/api/v1/notifications/stats` | Notification delivery statistics |

Built-in templates: appointment-reminder, lab-result-ready, prescription-filled, password-reset, visit-summary.

### Document/Blob Storage

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/blobs/upload` | Upload a file (multipart form) |
| GET | `/api/v1/blobs/:id` | Download file content |
| GET | `/api/v1/blobs/:id/metadata` | Get file metadata |
| DELETE | `/api/v1/blobs/:id` | Delete a file |
| GET | `/api/v1/blobs` | Search files (params: `patient_id`, `category`, `content_type`) |
| GET | `/api/v1/blobs/patient/:patientId` | List files by patient |

100MB max file size, SHA-256 integrity hashing, 10 supported medical MIME types.

### Audit Trail Search/Export

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/audit/search` | Search audit logs (params: `user_id`, `patient_id`, `action`, `start`, `end`) |
| GET | `/api/v1/audit/export/csv` | Export audit logs as CSV |
| GET | `/api/v1/audit/export/json` | Export audit logs as JSON |
| GET | `/api/v1/audit/summary` | Aggregate audit statistics |
| GET | `/api/v1/audit/:id` | Get single audit entry |

### FHIR Bulk Import/Edit

| Method | Path | Description |
|--------|------|-------------|
| POST | `/fhir/$import` | Start bulk import (NDJSON body) |
| GET | `/fhir/$import/:id` | Get import job status |
| GET | `/fhir/$import` | List import jobs |
| POST | `/fhir/$bulk-edit` | Start bulk update (JSON criteria + patch) |
| POST | `/fhir/$bulk-delete` | Start bulk delete (JSON criteria) |
| GET | `/fhir/$bulk-edit/:id` | Get edit job status |
| DELETE | `/fhir/$bulk-edit/:id` | Cancel edit job |

### FHIR $graphql

| Method | Path | Description |
|--------|------|-------------|
| POST | `/fhir/$graphql` | Execute GraphQL query (JSON body) |
| GET | `/fhir/$graphql` | Execute GraphQL query (query param) |

Supports: single resource by ID, list queries with search params, field selection.

### CodeSystem/$closure

| Method | Path | Description |
|--------|------|-------------|
| POST | `/fhir/CodeSystem/$closure` | Initialize or process closure table |

Transitive closure computation for SNOMED CT hierarchies. Initialize with `name`, then add concepts to compute subsumption relationships.

### API Key Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/api-keys` | Create new API key (returns raw key once) |
| GET | `/api/v1/api-keys` | List keys for tenant (never exposes hash) |
| GET | `/api/v1/api-keys/:id` | Get key details |
| DELETE | `/api/v1/api-keys/:id` | Revoke key |
| POST | `/api/v1/api-keys/:id/rotate` | Rotate key (revokes old, returns new) |

Key format: `ehr_k1_<32-hex>`. SHA-256 hashed storage. Supports scopes, rate limits, expiration, and metadata.

### Per-Client Rate Limiting

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/rate-limits/plans` | List all rate plans |
| POST | `/api/v1/admin/rate-limits/plans` | Create/update rate plan |
| GET | `/api/v1/admin/rate-limits/clients/:id` | Get client usage stats |
| PUT | `/api/v1/admin/rate-limits/clients/:id/plan` | Assign plan to client |
| POST | `/api/v1/admin/rate-limits/clients/:id/reset` | Reset client counters |

Default plans: Free (60/min), Starter (300/min), Professional (1K/min), Enterprise (5K/min). Per-minute/hour/day limits with burst and concurrent request tracking.

### SMART Backend Services

| Method | Path | Description |
|--------|------|-------------|
| POST | `/fhir/auth/token` | Token endpoint (grant_type=client_credentials) |
| POST | `/fhir/auth/register-backend` | Register backend service client |
| GET | `/fhir/auth/backend-clients` | List registered clients |
| DELETE | `/fhir/auth/backend-clients/:id` | Remove client |

RS384 JWT assertion verification, JTI replay protection, scope-subset validation.

### Webhook Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/webhooks` | Register webhook endpoint |
| GET | `/api/v1/webhooks` | List endpoints |
| GET | `/api/v1/webhooks/:id` | Get endpoint details |
| PUT | `/api/v1/webhooks/:id` | Update endpoint |
| DELETE | `/api/v1/webhooks/:id` | Delete endpoint |
| POST | `/api/v1/webhooks/:id/test` | Send test event |
| GET | `/api/v1/webhooks/:id/deliveries` | Delivery logs |
| POST | `/api/v1/webhooks/deliveries/:id/retry` | Retry failed delivery |
| POST | `/api/v1/webhooks/:id/pause` | Pause endpoint |
| POST | `/api/v1/webhooks/:id/resume` | Resume endpoint |

HMAC-SHA256 payload signing, wildcard event matching, exponential retry backoff.

### API Usage Analytics

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/analytics/overview` | Overall usage stats |
| GET | `/api/v1/analytics/endpoints` | Top endpoints with latency/error rates |
| GET | `/api/v1/analytics/clients` | Top clients by usage |
| GET | `/api/v1/analytics/clients/:id` | Specific client stats |
| GET | `/api/v1/analytics/resources` | Resource type CRUD breakdown |
| GET | `/api/v1/analytics/timeseries` | Time-bucketed metrics (interval=1m/5m/1h) |

### Sandbox & Synthetic Data

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/admin/sandbox/seed` | Generate synthetic data (accepts SeedConfig) |
| GET | `/api/v1/admin/sandbox/resources/:type` | List generated resources |
| POST | `/api/v1/admin/sandbox/reset` | Clear all synthetic data |
| GET | `/api/v1/admin/sandbox/export/ndjson/:type` | Export as NDJSON |
| GET | `/api/v1/admin/sandbox/export/bundle` | Export as FHIR Transaction Bundle |

Generates realistic patients with ICD-10 conditions, LOINC observations, RxNorm medications, CVX immunizations, CPT procedures, and SNOMED allergies. Reproducible with seed parameter.

### Detailed CapabilityStatement

| Method | Path | Description |
|--------|------|-------------|
| GET | `/fhir/metadata` | Full CapabilityStatement |
| GET | `/fhir/metadata/resources` | List supported resource types |
| GET | `/fhir/metadata/resources/:type` | Capability for specific resource |
| GET | `/fhir/metadata/operations` | List all FHIR operations |
| POST | `/fhir/metadata/search-params` | Register custom search parameter |
| GET | `/fhir/metadata/search-params` | List custom search parameters |
| DELETE | `/fhir/metadata/search-params/:type/:name` | Delete custom search parameter |

38 resource types with detailed search parameters (20+ for Patient), 12 server-level operations, custom search parameter API.

### Role-Based Access Control

Route-level RBAC is enforced via `auth.RequireRole()` middleware. The `admin` role has implicit access to all resources. Scope-level access control follows the SMART on FHIR format and is enforced via `auth.RequireScope()`.

---

## Multi-Tenancy

### Schema-Per-Tenant Isolation

Each tenant gets a dedicated PostgreSQL schema named `tenant_<identifier>`. All queries within a request run against the tenant's schema via `SET search_path TO tenant_<id>, shared, public`. This provides:

- Complete data isolation between tenants at the database level
- No risk of cross-tenant data leakage
- Independent migration state per tenant
- HIPAA-compliant separation of PHI

### Tenant Resolution

The tenant middleware resolves the active tenant using the following priority:

1. **JWT claim** -- The `tenant_id` claim in the authenticated token (highest priority)
2. **HTTP header** -- The `X-Tenant-ID` request header
3. **Query parameter** -- The `tenant_id` query parameter
4. **Default** -- Falls back to the `DEFAULT_TENANT` configuration value

### Creating a Tenant

```bash
# Using the CLI
./bin/ehr-server tenant create --name acme

# This creates the PostgreSQL schema: tenant_acme
# Then run migrations for the new schema:
./scripts/migrate.sh acme
```

Tenant identifiers must be alphanumeric (plus underscores) and are validated against the pattern `^[a-zA-Z0-9_]+$`.

---

## Domain Guide

### admin

Organizations, departments, locations, and system users. Provides the foundational administrative structures that other domains reference. Maps to FHIR Organization and Location resources.

### identity

Patient and practitioner demographics, identifiers, contacts, and roles. Serves as the master person index. Maps to FHIR Patient, Practitioner, and PractitionerRole resources. Supports MRN, NPI, ABHA, HPR, and custom identifier systems.

### encounter

Patient visits, admissions, and care episodes. Tracks encounter status, class (ambulatory, inpatient, emergency), diagnoses, care teams, and discharge summaries. Maps to FHIR Encounter.

### clinical

Core clinical observations including conditions (diagnoses), observations (vitals, labs), allergies/intolerances, and procedures. Maps to FHIR Condition, Observation, AllergyIntolerance, and Procedure.

### medication

Full medication lifecycle: medication catalog, prescriptions (MedicationRequest), administrations (MAR), dispenses, and medication statements. Maps to FHIR Medication, MedicationRequest, MedicationAdministration, MedicationDispense, and MedicationStatement.

### diagnostics

Laboratory and imaging workflows: service requests (orders), specimen tracking, diagnostic reports, and imaging studies. Maps to FHIR ServiceRequest, Specimen, DiagnosticReport, and ImagingStudy.

### scheduling

Appointment scheduling: provider schedules, time slots, appointments, and waitlist management. Maps to FHIR Schedule, Slot, and Appointment.

### billing

Revenue cycle management: insurance coverage, claims submission, claim responses (adjudication/EOBs), and invoices. Maps to FHIR Coverage, Claim, ClaimResponse, and Invoice.

### documents

Clinical documentation: patient consents, document references, clinical notes (progress notes, H&P, discharge summaries), and FHIR Compositions. Maps to FHIR Consent, DocumentReference, and Composition.

### inbox

In-basket messaging system: message pools, inbox messages, co-sign requests, patient lists, and clinical handoffs. Maps to FHIR Communication for messaging interoperability.

### surgery

Operating room management: OR room scheduling, surgical cases, surgeon preference cards, and implant tracking/logging.

### emergency

Emergency department workflows: triage assessments (ESI scoring), ED patient tracking boards, and trauma activation management.

### obstetrics

Maternal health: pregnancy records, prenatal visit tracking, labor monitoring, delivery records, newborn documentation, and postpartum assessments.

### oncology

Cancer care management: cancer diagnosis staging (TNM), treatment protocols, chemotherapy cycles, radiation therapy sessions, tumor marker tracking, and tumor board reviews.

### nursing

Nursing documentation: flowsheet templates and entries, nursing assessments, fall risk assessments, skin assessments (Braden), pain assessments, lines/drains/tubes, restraint monitoring, and intake/output tracking.

### behavioral

Behavioral health: psychiatric assessments, safety plans, legal holds (involuntary commitment), seclusion/restraint episodes, and group therapy sessions.

### research

Clinical research: study protocols, patient enrollment, adverse event reporting, and protocol deviation tracking. Maps to FHIR ResearchStudy.

### portal

Patient portal: portal accounts, patient-provider messaging, questionnaires, questionnaire responses, and patient check-in workflows. Maps to FHIR Questionnaire and QuestionnaireResponse.

### cds

Clinical decision support: CDS rules engine, alerts, drug interaction checking, order sets, clinical pathways with patient enrollment, formulary management, and medication reconciliation.

### subscription

FHIR R4 Subscription management with REST-hook webhook delivery. When resources are created, updated, or deleted, the notification engine evaluates them against active subscription criteria and queues webhook deliveries. Supports criteria filtering (resource type + parameters), exponential retry backoff, subscription expiry, and handshake verification. Admin-only access. Maps to FHIR Subscription.

---

## Configuration

All configuration is managed through environment variables. Copy `.env.example` to `.env` for local development.

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | *(required)* | PostgreSQL connection string |
| `DB_MAX_CONNS` | `20` | Maximum database connection pool size |
| `DB_MIN_CONNS` | `5` | Minimum database connection pool size |
| `REDIS_URL` | -- | Redis connection string |
| `PORT` | `8000` | HTTP server port |
| `ENV` | `development` | Environment (`development` or `production`). Controls auth mode and log format |
| `AUTH_ISSUER` | -- | OIDC issuer URL (e.g., `http://localhost:8080/realms/ehr`) |
| `AUTH_JWKS_URL` | -- | JWKS endpoint URL. Auto-discovered from issuer if not set |
| `AUTH_AUDIENCE` | -- | Expected JWT audience claim (e.g., `ehr-api`) |
| `DEFAULT_TENANT` | `default` | Fallback tenant ID when none is specified in the request |
| `CORS_ORIGINS` | `http://localhost:3000` | Comma-separated list of allowed CORS origins |

---

## Project Structure

```
ehr/
|-- cmd/
|   |-- ehr-server/
|       |-- main.go                 # Entry point, CLI commands, server wiring
|-- internal/
|   |-- config/
|   |   |-- config.go               # Configuration loading (viper)
|   |-- platform/
|   |   |-- auth/
|   |   |   |-- middleware.go        # JWT validation, dev-mode bypass
|   |   |   |-- oidc.go             # OIDC discovery client
|   |   |   |-- rbac.go             # Role and scope enforcement
|   |   |   |-- smart.go            # SMART on FHIR configuration endpoint
|   |   |-- db/
|   |   |   |-- pool.go             # Connection pool (pgxpool)
|   |   |   |-- tenant.go           # Tenant middleware, schema resolution
|   |   |   |-- migrate.go          # SQL migration runner
|   |   |-- fhir/
|   |   |   |-- bundle.go           # FHIR Bundle builder
|   |   |   |-- resource.go         # FHIR resource types, CapabilityStatement
|   |   |-- hipaa/
|   |   |   |-- audit.go            # HIPAA audit logger, break-glass support
|   |   |   |-- encryption.go       # AES-256-GCM PHI field encryption
|   |   |-- middleware/
|   |   |   |-- audit.go            # Request-level audit logging
|   |   |   |-- logger.go           # Structured HTTP request logging
|   |   |   |-- recovery.go         # Panic recovery
|   |   |   |-- requestid.go        # X-Request-ID generation
|   |   |-- plugin/
|   |       |-- host.go             # Plugin registry and lifecycle
|   |-- domain/
|       |-- admin/                   # T0: Organizations, departments, locations, users
|       |-- identity/                # T0: Patients, practitioners, roles
|       |-- encounter/               # T0: Encounters, diagnoses, care teams
|       |-- clinical/                # T1: Conditions, observations, allergies, procedures
|       |-- medication/              # T1: Medications, requests, admin, dispense, statements
|       |-- diagnostics/             # T1: Service requests, specimens, reports, imaging
|       |-- scheduling/              # T2: Schedules, slots, appointments, waitlists
|       |-- billing/                 # T2: Coverage, claims, responses, invoices
|       |-- documents/               # T2: Consents, document refs, notes, compositions
|       |-- inbox/                   # T2: Message pools, inbox, cosign, handoffs
|       |-- surgery/                 # T3: OR rooms, surgical cases, preference cards, implants
|       |-- emergency/               # T3: Triage, ED tracking, trauma activations
|       |-- obstetrics/              # T3: Pregnancy, prenatal, labor, delivery, newborn
|       |-- oncology/                # T3: Cancer Dx, protocols, chemo, radiation, tumor boards
|       |-- nursing/                 # T3: Flowsheets, assessments, I/O, restraints
|       |-- behavioral/              # T4: Psych assessments, safety plans, legal holds
|       |-- research/                # T4: Studies, enrollment, adverse events, deviations
|       |-- portal/                  # T4: Portal accounts, questionnaires, check-in
|       |-- cds/                     # T4: Rules, alerts, order sets, pathways, formulary
|       |-- subscription/              # Subscriptions, webhook notifications, delivery engine
|-- migrations/
|   |-- 001_t0_core_tables.sql       # Admin, identity, encounter schemas
|   |-- 002_t1_clinical_tables.sql   # Clinical domain tables
|   |-- ...                          # One migration file per domain tier
|   |-- 018_hipaa_row_level_security.sql  # Row-level security policies
|-- api/                             # API specifications (OpenAPI, etc.)
|-- pkg/
|   |-- fhirmodels/                  # Shared FHIR data types
|   |-- pagination/                  # Pagination utilities
|-- scripts/
|   |-- migrate.sh                   # Schema migration helper
|   |-- seed.sh                      # Seed data loader
|-- test/
|   |-- e2e/                         # End-to-end tests
|   |-- integration/                 # Integration tests
|-- docker-compose.yml               # Development environment
|-- Dockerfile                       # Multi-stage production build
|-- Makefile                         # Build, test, lint, migrate commands
|-- .env.example                     # Environment variable template
|-- atlas.hcl                        # Atlas schema management config
```

Each domain follows a consistent **5-file pattern**:

| File | Purpose |
|------|---------|
| `model.go` | Domain types (structs), FHIR conversion methods |
| `repo.go` | Repository interface definition |
| `repo_pg.go` | PostgreSQL implementation of the repository |
| `service.go` | Business logic and validation |
| `handler.go` | HTTP handlers, route registration |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on setting up the development environment, code organization conventions, testing practices, and the pull request process.

---

## License

This project is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for the full license text.

---

## Acknowledgments

- [Echo](https://echo.labstack.com/) -- High-performance Go web framework
- [pgx](https://github.com/jackc/pgx) -- PostgreSQL driver and connection pool for Go
- [HL7 FHIR R4](https://hl7.org/fhir/R4/) -- Fast Healthcare Interoperability Resources
- [SMART on FHIR](https://smarthealthit.org/) -- Standards-based application framework for healthcare
- [Keycloak](https://www.keycloak.org/) -- Open-source identity and access management
