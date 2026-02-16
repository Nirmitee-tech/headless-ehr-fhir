# @ehr/fhir-types

TypeScript type definitions for FHIR R4 resources and data types. Zero runtime overhead, zero dependencies.

## Installation

```bash
pnpm add @ehr/fhir-types
```

## Usage

```ts
import type { Patient, HumanName, Bundle, CodeableConcept } from '@ehr/fhir-types'

const patient: Patient = {
  resourceType: 'Patient',
  id: '123',
  name: [{ use: 'official', family: 'Smith', given: ['John'] }],
  gender: 'male',
  birthDate: '1990-01-15',
}

function displayName(name: HumanName): string {
  const given = name.given?.join(' ') ?? ''
  return `${given} ${name.family ?? ''}`.trim()
}
```

## What's Included

### FHIR Primitive Types

All FHIR R4 primitive types as branded type aliases for type safety:

```ts
import type {
  FHIRString, FHIRBoolean, FHIRInteger, FHIRDecimal,
  FHIRUri, FHIRUrl, FHIRCanonical,
  FHIRDate, FHIRDateTime, FHIRInstant, FHIRTime,
  FHIRCode, FHIRId, FHIROid, FHIRUuid,
  FHIRBase64Binary, FHIRMarkdown,
  FHIRUnsignedInt, FHIRPositiveInt,
} from '@ehr/fhir-types'
```

### FHIR Data Types

Complex data types used across all FHIR resources:

| Type | Description |
|------|-------------|
| `Coding` | Code from a terminology system |
| `CodeableConcept` | Concept with one or more codings and text |
| `HumanName` | Person name with family, given, prefix, suffix |
| `Address` | Physical / postal address |
| `ContactPoint` | Phone, email, fax, etc. |
| `Identifier` | Business identifier (MRN, SSN, etc.) |
| `Reference` | Reference to another resource |
| `Period` | Start/end datetime range |
| `Quantity` | Value with unit and system |
| `Range` | Low/high quantity range |
| `Ratio` | Numerator/denominator quantities |
| `Money` | Currency amount |
| `Annotation` | Text note with author and time |
| `Attachment` | Binary content (documents, images) |
| `Narrative` | XHTML narrative for human display |
| `Dosage` | Medication dosage instructions |
| `Timing` | Repeating event schedule |
| `Meta` | Resource metadata (version, lastUpdated, tags) |
| `Extension` | FHIR extension with typed values |

### Base Types

```ts
import type { Resource, DomainResource, Bundle, BundleEntry, OperationOutcome } from '@ehr/fhir-types'
```

- **`Resource`** — Base for all FHIR resources (`resourceType`, `id`, `meta`)
- **`DomainResource`** — Extends Resource with `contained`, `extension`, `modifierExtension`
- **`Bundle<T>`** — Generic bundle (searchset, transaction, history, document)
- **`BundleEntry<T>`** — Bundle entry with `resource`, `search`, `request`, `response`
- **`OperationOutcome`** — Error/info response with typed issues

### Resources

| Resource | Description |
|----------|-------------|
| `Patient` | Patient demographics, identifiers, contacts |
| `PatientContact` | Emergency contacts and next of kin |
| `PatientCommunication` | Language preferences |
| `PatientLink` | Links between patient records |

## Design Principles

- **Type-only package** — Zero runtime code, zero bundle impact
- **FHIR R4 compliant** — Matches the HL7 FHIR R4 specification
- **Backend-agnostic** — Works with any FHIR server (HAPI, Medplum, Azure, Google, etc.)
- **Strict TypeScript** — All types use exact string literals for code values

## Adding New Resource Types

Resource types are added as individual files under `src/resources/` and re-exported from the index:

```
src/
  primitives.ts      # FHIR primitive types
  datatypes.ts       # Complex data types
  resources/
    patient.ts       # Patient resource
    observation.ts   # (coming in Phase 1)
    condition.ts     # (coming in Phase 1)
  index.ts           # Re-exports everything
```

## License

Apache-2.0
