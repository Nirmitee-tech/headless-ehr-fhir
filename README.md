# Headless EHR — Open Source Electronic Health Record

An open-source, headless EHR system designed for modern healthcare. API-first architecture supporting any frontend (web, mobile, voice).

## 🏥 ONC Inferno Conformance — 47/47 tests passing

This server is tested against the **ONC Certification (g)(10) Standardized API Test Kit** (Inferno) — the same test suite used to certify US EHRs — on the **SMART App Launch (STU2) — Standalone Patient App** scenario.

**Latest run (June 11, 2026): all 47 tests passing — zero failures, zero skips.**

The run was executed over HTTPS (TLS-terminating proxy in front of the server) with the client authenticating via **asymmetric `private_key_jwt` (RS384 client assertions)** — so the `client-confidential-asymmetric` capability is exercised end-to-end, not just declared. Coverage includes: full OAuth 2.0 + PKCE standalone launch (GET and POST authorize), OpenID Connect (ID token, JWKS, `fhirUser` resolution), token refresh, SMART v1 **and** v2 scope grammars, patient context, and TLS verification.

📋 **[Full test-by-test results →](api/inferno-results/2026-06-11/results.md)**

### US Core Single Patient API (g10) — 86/94 functional tests passing

Beyond SMART App Launch, the server is tested against the **US Core 3.1.1 Single
Patient API** group: **86 of 94 functional tests pass** (up from 40), covering
search by patient, token (`system|code`), GET/POST `_search` parity, `_include`,
and reference integrity across all US Core resource types. Every resource the
server emits validates clean when sent to the HL7 FHIR validator directly.
The remaining items require a CI-grade HL7 validator (the local Docker validator
cannot resolve US Core 3.1.1 profiles under an orchestrated run) plus HTTPS.
📋 **[US Core status & details →](api/inferno-results/US-CORE-STATUS.md)**

### Reproduce it yourself

```bash
# 1. This server
cd api && docker compose up -d postgres redis
go run ./cmd/ehr-server migrate up && make seed
# load Inferno test fixtures
(echo "SET search_path TO tenant_default, public;"; cat scripts/seed_inferno.sql) | \
  docker exec -i api-postgres-1 psql -U ehr -d ehr
go build -o bin/ehr-server ./cmd/ehr-server
SMART_ISSUER=http://host.docker.internal:8000 AUTH_MODE=standalone ./bin/ehr-server serve

# 2. Inferno (ONC g10 kit)
git clone https://github.com/onc-healthit/onc-certification-g10-test-kit
cd onc-certification-g10-test-kit && docker compose run --rm inferno bundle exec inferno migrate && docker compose up -d

# 3. Run the SMART Standalone Patient App suite against http://host.docker.internal:8000/fhir
./api/scripts/inferno-test.sh
```

Don't take our word for it — run the tests.

## Architecture

```
ehr/
├── api/          # Go backend (FHIR R4 + REST API)
├── web/          # Frontend (coming soon)
├── docs/         # Documentation
└── deploy/       # Deployment configs
```

## Quick Start

```bash
cd api
docker compose up -d          # Start Postgres + Redis + Keycloak
make migrate-up               # Run database migrations
make seed                     # Load reference data
make dev                      # Full setup (all above)
make build                    # Build server binary
./bin/ehr-server              # Start server
```

## Features

- 28 clinical domains covering 70+ FHIR R4 resources
- FHIR R4 compliant REST API
- SMART on FHIR app launch framework
- Schema-per-tenant multi-tenancy
- HIPAA audit logging and PHI encryption
- Role-based access control (RBAC)
- Clinical Decision Support (CDS)
- Terminology service (ICD-10, LOINC, SNOMED, RxNorm, CPT)
- Real-time audit trail with break-glass support

## API Documentation

The API server exposes a RESTful interface organized by clinical domain. Each domain follows FHIR R4 resource conventions where applicable.

Once the server is running, API documentation is available at:
- `GET /api/v1/` — API root with available endpoints
- `GET /fhir/metadata` — FHIR CapabilityStatement

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for a detailed system design overview.

## Domains Overview

The system is organized into 28 clinical domains across 5 tiers:

| Tier | Domain | Description |
|------|--------|-------------|
| T0 | Admin | Organizations, departments, locations, system users |
| T0 | Identity | Patient demographics, matching, and merging |
| T0 | Encounter | Visits, admissions, and transfers |
| T1 | Clinical | Problems, allergies, vitals, assessments, flags, detected issues, adverse events, clinical impressions, risk assessments |
| T1 | Medication | Prescriptions, dispensing, and administration |
| T1 | Diagnostics | Lab orders, results, and imaging |
| T1 | FHIR List | Curated resource lists |
| T2 | Scheduling | Appointments and provider availability |
| T2 | Billing | Claims, charges, and insurance |
| T2 | Documents | Clinical documents and notes |
| T2 | Inbox | Clinical messaging and notifications |
| T2 | Episode of Care | Longitudinal care tracking |
| T2 | Healthcare Service | Service catalog and availability |
| T2 | Measure Report | Quality measure reporting |
| T3 | Surgery | Surgical cases and procedures |
| T3 | Nursing | Nursing assessments and care plans |
| T3 | Oncology | Cancer treatment protocols |
| T3 | Emergency | ED triage and tracking |
| T3 | Obstetrics | Maternal and prenatal care |
| T3 | Financial | Accounts, insurance plans, payments, charges, contracts, enrollments |
| T3 | Workflow | Activity definitions, request groups, guidance responses |
| T3 | Supply | Supply requests and deliveries |
| T4 | Behavioral | Behavioral health assessments |
| T4 | Research | Clinical trials and research protocols |
| T4 | Portal | Patient portal and self-service |
| T4 | CDS | Clinical decision support engine |
| T4 | Conformance | Naming systems, operation definitions, message definitions |
| T4 | Vision Prescription | Optometry prescriptions and lens specifications |
| T4 | Terminology | ICD-10, LOINC, SNOMED CT, RxNorm, CPT code systems |

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -am 'Add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

Please ensure all tests pass and follow the existing code style.

## License

This project is open source. See the [LICENSE](LICENSE) file for details.

