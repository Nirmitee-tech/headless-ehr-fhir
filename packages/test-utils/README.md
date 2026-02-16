# @ehr/test-utils

Mock FHIR data and test utilities for the EHR component library. Provides realistic, type-safe FHIR resources for unit testing.

## Installation

This is a private, workspace-internal package. It is not published to npm.

```json
{
  "devDependencies": {
    "@ehr/test-utils": "workspace:*"
  }
}
```

## Usage

```tsx
import { mockPatient, mockPatientMinimal, mockPatientEmpty } from '@ehr/test-utils'
import { render } from '@testing-library/react'
import { PatientBanner } from '@ehr/react'

describe('PatientBanner', () => {
  it('renders full patient data', () => {
    render(<PatientBanner patient={mockPatient} />)
  })

  it('handles minimal data gracefully', () => {
    render(<PatientBanner patient={mockPatientMinimal} />)
  })

  it('handles empty patient', () => {
    render(<PatientBanner patient={mockPatientEmpty} />)
  })
})
```

## Available Mocks

### `mockPatient`

A fully populated Patient resource with:

- MRN identifier (12345678)
- Official name: Dr. John Andrew Smith Jr.
- Phone: (555) 123-4567 (mobile)
- Email: john.smith@example.com
- Gender: male
- DOB: 1978-03-15
- Address: 123 Main Street, Apt 4B, Springfield, IL 62704
- Meta: version 1, lastUpdated 2025-01-15

### `mockPatientMinimal`

A Patient with only required/common fields:

- Name: Jane Doe
- Gender: female

### `mockPatientEmpty`

A bare Patient with only `resourceType` and `id`. Use this to test null-safe rendering and graceful degradation.

## Adding New Mocks

Add mock files under `src/mocks/` and re-export from `src/index.ts`:

```
src/
  mocks/
    patient.ts        # Patient mocks
    observation.ts    # (add new resource mocks here)
    condition.ts
  index.ts            # Re-exports all mocks
```

Each mock file should export:
1. A **full** mock with all fields populated
2. A **minimal** mock with only common fields
3. An **empty** mock for edge case testing

## License

Apache-2.0
