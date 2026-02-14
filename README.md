# OpenEHR — Open Source Electronic Health Record

An open-source, headless EHR system designed for modern healthcare. API-first architecture supporting any frontend (web, mobile, voice).

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

- 19 clinical domains covering 40+ Epic EMR modules
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

The system is organized into 19 clinical domains across 5 tiers:

| Tier | Domain | Description |
|------|--------|-------------|
| T0 | Admin | Organizations, departments, locations, system users |
| T0 | Identity | Patient demographics, matching, and merging |
| T0 | Encounter | Visits, admissions, and transfers |
| T1 | Clinical | Problems, allergies, vitals, and assessments |
| T1 | Medication | Prescriptions, dispensing, and administration |
| T1 | Diagnostics | Lab orders, results, and imaging |
| T2 | Scheduling | Appointments and provider availability |
| T2 | Billing | Claims, charges, and insurance |
| T2 | Documents | Clinical documents and notes |
| T2 | Inbox | Clinical messaging and notifications |
| T3 | Surgery | Surgical cases and procedures |
| T3 | Nursing | Nursing assessments and care plans |
| T3 | Oncology | Cancer treatment protocols |
| T3 | Emergency | ED triage and tracking |
| T3 | Obstetrics | Maternal and prenatal care |
| T4 | Behavioral | Behavioral health assessments |
| T4 | Research | Clinical trials and research protocols |
| T4 | Portal | Patient portal and self-service |
| T4 | CDS | Clinical decision support engine |

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -am 'Add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

Please ensure all tests pass and follow the existing code style.

## License

This project is open source. See the [LICENSE](LICENSE) file for details.
