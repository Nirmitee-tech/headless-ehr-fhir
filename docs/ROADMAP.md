# Roadmap

Phased execution plan from current state to complete platform. Each phase produces a working, shippable increment. Phases are sequential but individual work streams within a phase can run in parallel.

---

## Current State (Phase 0) — Complete

**Backend:** 38 domains, 5,000+ tests, FHIR R4 CRUD, SMART auth, CDS Hooks, Bulk Export, HL7v2, C-CDA, Subscriptions, Terminology, OpenTelemetry.

**Frontend:** Monorepo scaffolded with 6 packages. Design tokens (colors, spacing, typography, clinical semantics). 4 primitives (Box, Stack, Text, Badge). 50 tests passing. Storybook configured.

---

## Phase 1 — FHIR Data Components + Core Hooks

**Goal:** A developer can install `@ehr/react` and render any FHIR resource with one line of code.

### packages/ — Component Library

**@ehr/fhir-hooks (8 hooks)**
| Hook | Description |
|------|-------------|
| `FhirProvider` | Context provider for FHIR server configuration |
| `useFhirClient` | Access the configured FHIR client |
| `useFhirRead` | Read a resource by type and ID |
| `useFhirSearch` | Search with FHIR search parameters |
| `useFhirCreate` | Create a resource |
| `useFhirUpdate` | Update an existing resource |
| `useFhirDelete` | Delete a resource |
| `useFhirOperation` | Execute named operations ($validate, $everything) |

**@ehr/react — Phase 1 Components (20)**
| Component | FHIR Type | Description |
|-----------|-----------|-------------|
| HumanName | HumanName | Formatted name with prefix/suffix |
| Address | Address | Multi-line address display |
| ContactPoint | ContactPoint | Phone/email with click actions |
| Identifier | Identifier | MRN, SSN with type labels |
| CodeableConcept | CodeableConcept | Code display with system tooltip |
| Reference | Reference | Linked resource reference |
| PatientBanner | Patient | Header bar: name, DOB, age, MRN, gender, photo, flags |
| MedicationList | MedicationRequest | Active/historical meds with dose, frequency, status |
| AllergyList | AllergyIntolerance | Allergies with severity badges and reactions |
| ProblemList | Condition | Active conditions with onset date and status |
| LabResults | Observation | Lab values with reference ranges, H/L/HH/LL flags |
| VitalsPanel | Observation | Vital signs grid (BP, HR, Temp, RR, SpO2, Weight) |
| ImmunizationList | Immunization | Vaccine history with dates and doses |
| TimelineEvent | Various | Single event for clinical timeline display |
| ClinicalNote | DocumentReference | Note display with author, date, type |
| CarePlanCard | CarePlan | Care plan summary with goals and activities |
| ReferralCard | ServiceRequest | Referral display with status and provider |
| DocumentViewer | DocumentReference | PDF/image viewer for clinical documents |
| DataTable | Any | Sortable, filterable table for clinical data |
| EmptyState | — | Placeholder for empty lists |

**@ehr/fhir-types — Expand resource types**
Add TypeScript types for: Observation, Condition, MedicationRequest, AllergyIntolerance, Immunization, Encounter, Procedure, DiagnosticReport, ServiceRequest, CarePlan, DocumentReference

**@ehr/test-utils — Expand mocks**
Add mock data for every new resource type (full, minimal, empty variants)

**Testing target:** 400+ tests across all packages

### api/ — Backend Gaps

| Feature | Description |
|---------|-------------|
| Secure Messaging API | Message, MessageHeader resources; inbox/outbox/thread queries |
| Document Storage | Binary resource with S3-compatible object storage |
| Batch/Transaction | Bundle processing for atomic multi-resource operations |

### Milestone Deliverable
```bash
pnpm add @ehr/react @ehr/fhir-hooks @ehr/tokens
```
```tsx
import { FhirProvider, useFhirRead } from '@ehr/fhir-hooks'
import { PatientBanner, MedicationList } from '@ehr/react'
// Working clinical UI in 10 lines of code
```

---

## Phase 2 — Provider Portal MVP

**Goal:** A clinician can log in, search for a patient, view a chart, and read clinical data.

### web/ — Reference Application

**Framework:** Next.js 14+ (App Router), authenticated via SMART on FHIR

**Provider Portal — Read-Only MVP:**
| Screen | Description |
|--------|-------------|
| Login | SMART on FHIR launch, provider selection |
| Patient Search | Search by name, MRN, DOB; recent patients list |
| Patient Chart | Tabbed view: Summary, Problems, Medications, Allergies, Labs, Vitals, Notes, Timeline |
| Schedule | Read-only day view of today's appointments |

**Layout:**
- AppShell with collapsible sidebar navigation
- Top bar with provider name, notifications bell, settings
- Patient context bar (sticky banner when inside a chart)

### packages/ — Application Shell Components

| Component | Description |
|-----------|-------------|
| AppShell | Page layout with sidebar, header, content area |
| Sidebar | Collapsible navigation with icons and labels |
| PatientSearch | Search input with typeahead results |
| UserMenu | Provider menu with logout |

### deploy/ — Local Development

| File | Description |
|------|-------------|
| `docker-compose.yml` | API + PostgreSQL + Next.js web app + MinIO |
| `Dockerfile.api` | Multi-stage Go build |
| `Dockerfile.web` | Next.js standalone build |
| `seed/` | Demo data loader: 5 providers, 50 patients |

### Milestone Deliverable
```bash
docker compose up
# Open localhost:3000 → Login → Search "Smith" → View patient chart
```

---

## Phase 3 — Clinical Workflows + Write Operations

**Goal:** A clinician can document an encounter, place orders, and manage medications.

### web/ — Provider Portal Write Capabilities

| Screen | Description |
|--------|-------------|
| Encounter Form | Start/complete encounters, select type and reason |
| SOAP Note Editor | Structured note with Subjective, Objective, Assessment, Plan sections |
| Order Entry | Place lab, imaging, and medication orders |
| Problem Management | Add/resolve/edit conditions |
| Medication Management | Prescribe, renew, discontinue medications |
| Allergy Management | Record new allergies with severity and reactions |
| Vitals Entry | Record vital signs with auto-BMI calculation |

### packages/ — Clinical Workflow Components (15)

| Component | Description |
|-----------|-------------|
| ClinicalNoteEditor | Rich text SOAP note with template insertion |
| OrderEntry | Order form with CDS Hooks integration |
| PrescriptionForm | Medication prescribing with sig builder |
| MedicationReconciliation | Compare and reconcile medication lists |
| EncounterForm | Start/edit encounter details |
| AssessmentPlan | A&P section with diagnosis and order linking |
| DiagnosisSelector | ICD-10/SNOMED search and select |
| ProcedureLogger | Record procedures with CPT codes |
| FlowSheet | Tabular data entry for serial measurements |
| GrowthChart | Pediatric growth percentile charts |
| FamilyHistory | Family member condition entry |
| SurgicalHistory | Surgical history list management |
| SocialHistory | Social determinants of health entry |
| ReviewOfSystems | Multi-system checklist |
| PhysicalExam | System-by-system exam documentation |

### api/ — Backend for Write Workflows

| Feature | Description |
|---------|-------------|
| Clinical Documentation API | Note templates, structured SOAP storage, co-sign workflow |
| Order Entry API (CPOE) | Order creation with CDS integration, status tracking |
| RBAC | Role-based access control (provider, nurse, admin, patient) |

### Milestone Deliverable
```
Login → Search patient → Start encounter → Write SOAP note →
Order CBC lab → Prescribe amoxicillin → Sign and close encounter
```

---

## Phase 4 — Patient Portal + Admin Console

**Goal:** Patients can view their records and message their provider. Admins can manage the system.

### web/ — Patient Portal

| Screen | Description |
|--------|-------------|
| Login | Patient SMART on FHIR launch |
| Health Summary | Conditions, medications, allergies, immunizations |
| Lab Results | Results with reference ranges, trend sparklines |
| Appointments | View upcoming, book new, cancel existing |
| Messages | Secure messaging threads with care team |
| Intake Forms | Pre-visit questionnaires via FHIR Questionnaire |
| Medication Refills | Request refills from active prescriptions |

### web/ — Admin Console

| Screen | Description |
|--------|-------------|
| User Management | CRUD users, assign roles, reset passwords |
| Organization Setup | Practices, locations, departments |
| Integration Dashboard | HL7v2 feed status, FHIR subscription status, webhook logs |
| Audit Log Viewer | Searchable access log with filters (user, patient, action, date) |
| System Health | OpenTelemetry dashboards (API latency, error rates, queue depth) |

### packages/ — Forms and Input Components (10)

| Component | Description |
|-----------|-------------|
| FhirForm | Auto-generated form from FHIR StructureDefinition |
| SmartFormRenderer | Render FHIR Questionnaire as interactive form |
| ConsentCapture | Digital consent with signature |
| InsuranceCard | Insurance card display and entry |
| DemographicsForm | Patient demographics editor |
| VitalsInput | Vital signs entry with validation |
| PainScale | Visual analog pain scale (0-10) |
| WoundTracker | Wound measurement and photo capture |
| MedicationPicker | Medication search with RxNorm integration |
| DiagnosisPicker | ICD-10/SNOMED diagnosis search |

### packages/ — Remaining Shell Components

| Component | Description |
|-----------|-------------|
| CommandPalette | Cmd+K quick actions and navigation |
| ScheduleCalendar | Day/week/month calendar with drag-and-drop |
| InboxPanel | Unified inbox (results, messages, tasks, refills) |
| TaskBoard | Kanban-style task management |
| NotificationCenter | Real-time notification dropdown |
| AuditTrail | Audit event display with details |

### api/ — Backend Additions

| Feature | Description |
|---------|-------------|
| Audit Log Query API | Searchable audit events with FHIR AuditEvent resource |
| Reliable Webhook Delivery | At-least-once with exponential backoff and dead letter queue |
| Questionnaire Response API | Save and query patient-submitted forms |

### Milestone Deliverable
```bash
docker compose up
# localhost:3000 → Provider portal (full charting)
# localhost:3001 → Patient portal (view records, message, book appointments)
# localhost:3002 → Admin console (manage users, view audit logs)
```

---

## Phase 5 — Production Hardening

**Goal:** Ready for real clinical use in a supervised environment.

### Security and Compliance

| Item | Description |
|------|-------------|
| e-Prescribing (NCPDP SCRIPT) | Electronic prescribing with pharmacy routing |
| Penetration testing | OWASP top 10 audit, FHIR-specific security review |
| HIPAA compliance checklist | BAA template, encryption at rest/transit, access controls |
| Data encryption | AES-256 at rest, TLS 1.3 in transit, field-level encryption for PII |
| Session management | Token rotation, idle timeout, concurrent session limits |
| Rate limiting | Per-user, per-endpoint rate limits with 429 responses |

### Performance

| Item | Description |
|------|-------------|
| Load testing | k6 scripts for API endpoints, target: 1000 concurrent users |
| Database optimization | Query analysis, index tuning, connection pooling |
| CDN configuration | Static asset caching for web applications |
| Component bundle analysis | Verify < 2KB per component, < 50KB total library |

### Infrastructure

| Item | Description |
|------|-------------|
| Kubernetes Helm chart | Parameterized deployment for any K8s cluster |
| Terraform modules | AWS and GCP provisioning templates |
| Backup and restore | Automated database backups with point-in-time recovery |
| Monitoring dashboards | Grafana dashboards for API, database, and application metrics |
| Alerting | PagerDuty/Slack alerts for error rate spikes, latency, disk usage |

### Milestone Deliverable
Deployment guide that takes a team from zero to production on AWS or GCP in under a day. Penetration test report with zero critical findings.

---

## Phase 6 — Ecosystem and Community

**Goal:** A thriving open-source community and commercial ecosystem.

### Documentation Site (`docs/`)

| Section | Description |
|---------|-------------|
| Getting Started | 5-minute quickstart, first API call, first component |
| API Reference | Auto-generated from OpenAPI spec with curl examples |
| Component Storybook | Hosted, versioned, with copy-paste code snippets |
| Guides | 10+ tutorials covering common workflows |
| Architecture | System design, data model, security model |
| Contributing | Issue templates, PR guidelines, code style, domain guide |

### Community

| Item | Description |
|------|-------------|
| GitHub Discussions | Q&A, feature requests, show-and-tell |
| Discord server | Real-time community support |
| Contributing guide | How to add a domain, component, or integration |
| Issue templates | Bug reports, feature requests, security disclosures |
| CI/CD pipeline | GitHub Actions for test, build, release, deploy |
| Automated releases | Semantic versioning, changelogs, npm publishing |
| Security policy | Responsible disclosure process, CVE tracking |

### Ecosystem

| Item | Description |
|------|-------------|
| Plugin architecture | Extension points for custom domains and integrations |
| Marketplace concept | Directory of community-built modules (lab interfaces, pharmacy integrations, specialty workflows) |
| Certification guidance | ONC Health IT certification pathways documentation |
| Commercial support | Documentation for organizations offering paid support, hosting, and implementation services |

### Milestone Deliverable
1,000 GitHub stars. 10+ community contributors. Components published on npm with stable v1.0 semver. Documentation site live and indexed by search engines.

---

## Tracking

Progress is tracked through GitHub milestones matching these phases. Each phase has a milestone with issues for every deliverable. Use GitHub project boards for sprint-level tracking within each phase.

| Phase | Milestone | Status |
|-------|-----------|--------|
| Phase 0 | Foundation | Complete |
| Phase 1 | FHIR Data Components | Not Started |
| Phase 2 | Provider Portal MVP | Not Started |
| Phase 3 | Clinical Workflows | Not Started |
| Phase 4 | Patient Portal + Admin | Not Started |
| Phase 5 | Production Hardening | Not Started |
| Phase 6 | Ecosystem | Not Started |
