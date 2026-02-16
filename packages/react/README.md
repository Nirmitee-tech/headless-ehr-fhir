# @ehr/react

Pre-styled, FHIR-native React components for healthcare user interfaces. Built on `@ehr/primitives` and `@ehr/tokens` for consistent, accessible clinical UIs.

> **Status:** Stub package. Components are coming in Phase 1.

## Installation

```bash
pnpm add @ehr/react
```

## Planned Components (Phase 1)

### Clinical Data Display

| Component | Description | FHIR Resource |
|-----------|-------------|---------------|
| `HumanName` | Renders patient/practitioner names with proper formatting | HumanName datatype |
| `Address` | Formatted address display | Address datatype |
| `ContactPoint` | Phone, email, fax with click-to-action | ContactPoint datatype |
| `Identifier` | MRN, SSN, insurance ID display | Identifier datatype |
| `CodeableConcept` | Code display with tooltip for system/code | CodeableConcept datatype |
| `PatientBanner` | Standard patient header (name, DOB, MRN, photo, allergies) | Patient |

### Clinical Workflows

| Component | Description | FHIR Resource |
|-----------|-------------|---------------|
| `MedicationList` | Active/historical medications with status | MedicationRequest |
| `AllergyList` | Allergy display with severity badges | AllergyIntolerance |
| `ProblemList` | Active problems / conditions | Condition |
| `LabResults` | Lab values with reference ranges and flags (H/L/HH/LL) | Observation |
| `VitalsPanel` | Vital signs grid with sparklines | Observation |
| `TimelineEvent` | Clinical event for timeline views | Various |

### Layout Components

| Component | Description |
|-----------|-------------|
| `Card` | Content container with header, body, footer |
| `DataTable` | Sortable, filterable clinical data table |
| `Sidebar` | Collapsible navigation sidebar |
| `Tabs` | Tab navigation with lazy panel loading |
| `Modal` | Accessible dialog with focus trap |

## Design Principles

- **FHIR-native** — Pass FHIR resources directly as props; components handle rendering
- **Backend-agnostic** — Works with any FHIR R4 server; no vendor lock-in
- **Accessible** — WCAG 2.1 AA compliant, keyboard navigable, screen reader tested
- **Themeable** — All styling via CSS custom properties from `@ehr/tokens`
- **Composable** — Compound component patterns for maximum flexibility

## Planned Usage

```tsx
import { PatientBanner, MedicationList, AllergyList } from '@ehr/react'
import '@ehr/tokens/css'

function PatientChart({ patient, medications, allergies }) {
  return (
    <div>
      <PatientBanner patient={patient} />
      <AllergyList allergies={allergies} />
      <MedicationList
        medications={medications}
        columns={['medication', 'dose', 'frequency', 'status']}
        onSelect={(med) => console.log(med)}
      />
    </div>
  )
}
```

## License

Apache-2.0
