# Vision

## One Repo. Three Products. One Command to Run.

This project is the open-source answer to Epic, Cerner, and every proprietary EHR that locks healthcare into closed ecosystems. It is simultaneously:

1. **A headless EHR platform** — a FHIR R4 API and component library that developers use to build any healthcare application
2. **A deployable EHR application** — a reference implementation that hospitals and clinics can self-host and use out of the box
3. **A developer ecosystem** — documentation, tooling, and patterns that make healthcare software as accessible as building a web app

```
headless-ehr-fhir/
├── api/              ← Headless EHR Platform (Go)
├── packages/         ← Component Library (React)
├── web/              ← Reference EHR Application (Next.js)
├── deploy/           ← Docker / Kubernetes deployment
└── docs/             ← Developer documentation site
```

---

## Who This Is For

### Health Tech Startups
Install `@ehr/react` and `@ehr/fhir-hooks`, point them at any FHIR R4 server, and ship a clinical application in weeks instead of years. No vendor lock-in. No per-seat licensing. No waiting for API access.

### Hospital IT Teams
Clone the repo, run `docker compose up`, and deploy a complete EHR with provider portal, patient portal, and admin console. Customize workflows by modifying the reference app or building new modules on the platform.

### Digital Health Companies
Use the headless API to power telehealth platforms, remote patient monitoring, clinical trials, population health tools, or any product that touches clinical data. The platform handles FHIR compliance, terminology, CDS, and interoperability — you handle the product.

### Open Source Contributors
A well-documented, well-tested codebase with clear architecture. Every domain is isolated. Every component is independently testable. The contribution path is obvious: pick a domain, write tests, implement, submit a PR.

---

## The Five Layers

### Layer 1: Headless EHR Platform (`api/`)

A Go backend that implements the full surface area of a modern EHR through FHIR R4 APIs.

**What exists today:**
- 38 clinical domains (Patient, Encounter, Observation, MedicationRequest, AllergyIntolerance, Condition, Procedure, Immunization, CarePlan, CareTeam, Goal, Claim, Coverage, and 25 more)
- FHIR R4 server with full CRUD, search, versioning, and operations
- SMART on FHIR authorization (standalone + EHR launch)
- CDS Hooks 2.0 server with clinical decision support services
- Bulk Data Export ($export for system, patient, and group-level)
- HL7v2 interface engine with MLLP listener
- C-CDA document generation and parsing
- Topic-based FHIR Subscriptions with webhook delivery
- Terminology services ($lookup, $validate-code, $translate, $subsumes)
- PlanDefinition/$apply for clinical protocol automation
- SQL-on-FHIR ViewDefinition execution
- Patient/$match probabilistic matching
- OpenTelemetry observability (traces, metrics, logs)
- Auto-provenance tracking middleware
- 5,000+ tests across 40+ packages

**What remains:**

| Capability | Why It Matters |
|------------|---------------|
| Clinical Documentation API | Structured SOAP notes, templates, co-sign workflows |
| Order Entry API (CPOE) | Lab, imaging, and medication orders with CDS integration |
| e-Prescribing (NCPDP SCRIPT) | Electronic prescribing is mandatory for US EHRs |
| Secure Messaging API | Provider-patient and provider-provider communication |
| Role-Based Access Control | Multi-tenant authorization with clinical role scoping |
| Audit Log Query API | Searchable access logs for HIPAA compliance |
| Reliable Webhook Delivery | At-least-once delivery with retry and dead letter queue |
| Batch/Transaction Support | Atomic multi-resource operations |
| Document Storage | Binary/DocumentReference with S3-compatible storage |

### Layer 2: Component Library (`packages/`)

A React component library purpose-built for healthcare UIs. FHIR-native, backend-agnostic, accessible, and themeable.

**Architecture:**
- `@ehr/tokens` — Design tokens as CSS custom properties (colors, spacing, typography, clinical semantics)
- `@ehr/fhir-types` — TypeScript type definitions for FHIR R4 resources
- `@ehr/primitives` — Core UI primitives (Box, Stack, Text, Badge)
- `@ehr/fhir-hooks` — React hooks for FHIR data fetching and operations
- `@ehr/react` — Pre-styled, FHIR-native clinical components
- `@ehr/test-utils` — Mock FHIR data for testing

**Design principles:**
- Zero vendor lock-in: plain CSS custom properties, no Tailwind, no CSS-in-JS runtime
- FHIR-native: pass FHIR resources as props, components handle rendering and formatting
- Backend-agnostic: works with any FHIR R4 server (HAPI, Medplum, Azure, Google, or this one)
- Accessible: WCAG 2.1 AA, axe-core in every test, keyboard navigable, screen reader tested
- Performant: < 2KB per component gzipped, zero runtime CSS overhead
- Themeable: CSS custom properties for full visual customization, dark mode built-in

**Component set (55 total):**

Phase 1 — FHIR Data Display (20 components):
HumanName, Address, ContactPoint, Identifier, CodeableConcept, PatientBanner, MedicationList, AllergyList, ProblemList, LabResults, VitalsPanel, ImmunizationList, TimelineEvent, ClinicalNote, CarePlanCard, ReferralCard, DocumentViewer, Questionnaire, DataTable, EmptyState

Phase 2 — Clinical Workflows (15 components):
OrderEntry, PrescriptionForm, MedicationReconciliation, ClinicalNoteEditor, EncounterForm, AssessmentPlan, DiagnosisSelector, ProcedureLogger, FlowSheet, GrowthChart, FamilyHistory, SurgicalHistory, SocialHistory, ReviewOfSystems, PhysicalExam

Phase 3 — Application Shell (10 components):
AppShell, Sidebar, CommandPalette, PatientSearch, ScheduleCalendar, InboxPanel, TaskBoard, NotificationCenter, AuditTrail, UserMenu

Phase 4 — Forms and Input (10 components):
FhirForm, SmartFormRenderer, ConsentCapture, InsuranceCard, DemographicsForm, VitalsInput, PainScale, WoundTracker, MedicationPicker, DiagnosisPicker

### Layer 3: Reference Application (`web/`)

A Next.js application that proves the platform works end-to-end. Three portals, one codebase.

**Provider Portal** — The daily driver for clinicians:
- Patient Chart: single-page view with banner, problems, medications, allergies, vitals, labs, notes, orders, and clinical timeline
- Clinical Documentation: SOAP note editor with templates, auto-save, voice dictation ready, co-sign workflow
- Order Entry: lab, imaging, and medication orders with inline CDS alerts from the Hooks server
- Schedule: day/week calendar view, appointment types, check-in and checkout flow
- Inbox: results review, patient messages, refill requests, task assignments with priority sorting
- Patient Search: global search by name, MRN, DOB with recent patients list

**Patient Portal** — Self-service for patients:
- Health Summary: conditions, medications, allergies, immunizations in plain language
- Lab Results: results with reference ranges, trend charts, and provider annotations
- Appointments: book, reschedule, cancel appointments; join telehealth sessions
- Messages: secure messaging with care team
- Intake Forms: pre-visit questionnaires and consent forms via FHIR Questionnaire
- Medication Refills: request refills from active prescriptions

**Admin Console** — System configuration and monitoring:
- User Management: create/deactivate users, assign clinical roles, reset credentials
- Organization Setup: practices, locations, departments, provider rosters
- Integration Dashboard: HL7v2 interface status, FHIR subscriptions, webhook delivery logs
- Audit Log Viewer: searchable access logs (who accessed what record, when, from where)
- System Health: API latency, queue depth, error rates from OpenTelemetry data

### Layer 4: Infrastructure (`deploy/`)

Everything needed to go from `git clone` to running in production.

| Artifact | Purpose |
|----------|---------|
| `docker-compose.yml` | One-command local development (API + PostgreSQL + web + MinIO) |
| `docker-compose.prod.yml` | Production configuration with replicas and health checks |
| `Dockerfile.api` | Multi-stage Go build, minimal scratch image |
| `Dockerfile.web` | Next.js standalone output, optimized for container deployment |
| `k8s/` | Kubernetes manifests (Deployment, Service, Ingress, HPA, PDB) |
| `terraform/` | Optional cloud provisioning templates (AWS, GCP) |
| `seed/` | Demo data: 5 providers, 50 patients, realistic clinical records |
| Database migrations | Versioned, rollback-safe, tested in CI |

### Layer 5: Documentation (`docs/`)

A documentation site that serves as the developer portal and learning resource.

- **Getting Started**: five-minute quickstart from git clone to first API call
- **API Reference**: every FHIR endpoint, every operation, with curl and SDK examples
- **Component Storybook**: live, interactive component documentation with code snippets
- **Guides**: building a patient portal, adding a custom resource type, deploying to production, HIPAA compliance checklist
- **Architecture**: system design, data model, security model, integration patterns
- **Contributing**: how to add a domain, build a component, or write an integration

---

## Definition of Done

The project is complete when someone can do this:

```bash
git clone https://github.com/org/headless-ehr-fhir
cd headless-ehr-fhir
docker compose up
```

And within 60 seconds, open a browser to:

| URL | What You See |
|-----|-------------|
| `localhost:3000` | Provider portal with 50 demo patients, full charting |
| `localhost:3001` | Patient portal with health summary and messaging |
| `localhost:3002` | Admin console with user management and system health |
| `localhost:8080/fhir` | FHIR R4 API with full Capability Statement |
| `localhost:6006` | Component Storybook with 55 interactive components |
| `localhost:3003` | Documentation site with guides and API reference |

Demo seed data (5 providers, 50 patients with realistic clinical histories) loads automatically so the system feels alive on first launch.

And independently, a developer can do:

```bash
pnpm add @ehr/react @ehr/fhir-hooks @ehr/tokens
```

And build their own healthcare application from scratch, pointed at any FHIR R4 server, using only the libraries.

---

## What Makes This the Industry Standard

**Completeness.** Not a FHIR server. Not a component library. Not a starter template. A complete, production-grade EHR that works as both a product and a platform.

**No lock-in.** Every layer is independent. Use the API with your own frontend. Use the components with your own backend. Use the reference app as-is. Fork any piece. The platform imposes zero opinions on your architecture beyond FHIR R4 compliance.

**Developer experience.** Healthcare software has historically been built by healthcare companies for healthcare companies. This project makes it accessible to any developer. If you can build a Next.js app, you can build an EHR.

**Open source economics.** Epic charges millions. Cerner charges millions. Athenahealth charges per-encounter. This is free. The business model is services, hosting, and certifications — not the software itself.

**Clinical rigor.** Design tokens for clinical severity levels. Accessibility tested to WCAG 2.1 AA. FHIR R4 spec-compliant. HL7v2 interoperability. CDS Hooks for clinical decision support. This is not a toy — it is built to the same standards as commercial EHRs.

---

## Non-Goals

Things this project intentionally does not do:

- **Medical device certification** — This is a software platform, not a certified medical device. Deployments requiring FDA 510(k) or CE marking must pursue certification independently.
- **Billing/claims processing** — The platform stores FHIR Claim and ExplanationOfBenefit resources, but does not include a clearinghouse or payer integration. Use a dedicated RCM system.
- **AI/ML clinical models** — The CDS Hooks server can invoke external AI services, but no clinical prediction models are included. This avoids liability and allows teams to choose their own ML stack.
- **Mobile native apps** — The reference app is responsive web. Native iOS/Android apps are a separate effort that can use the FHIR API and SMART on FHIR auth.
- **Multi-language i18n** — The initial release targets English. The component library uses externalized strings for future localization but does not ship translations.
