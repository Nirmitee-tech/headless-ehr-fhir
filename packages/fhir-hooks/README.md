# @ehr/fhir-hooks

FHIR-aware React hooks for data fetching, search, and operations. Backend-agnostic — works with any FHIR R4 server.

> **Status:** Stub package. Hooks are coming in Phase 1.

## Installation

```bash
pnpm add @ehr/fhir-hooks
```

## Planned Hooks (Phase 1)

| Hook | Description |
|------|-------------|
| `useFhirClient` | Configure FHIR server connection (base URL, auth, headers) |
| `useFhirRead` | Read a single resource by type and ID |
| `useFhirSearch` | Search resources with FHIR search parameters |
| `useFhirCreate` | Create a new resource |
| `useFhirUpdate` | Update an existing resource |
| `useFhirDelete` | Delete a resource |
| `useFhirOperation` | Execute FHIR operations ($validate, $everything, etc.) |
| `useFhirBundle` | Execute batch/transaction bundles |

## Design Principles

- **Backend-agnostic** — Provide a base URL and optional auth; works with HAPI, Medplum, Azure, Google, or any FHIR R4 server
- **Type-safe** — Returns typed FHIR resources using `@ehr/fhir-types`
- **Cache-aware** — Built-in request deduplication and stale-while-revalidate
- **Framework-independent** — No opinion on state management; returns standard React state

## Planned Usage

```tsx
import { FhirProvider, useFhirSearch, useFhirRead } from '@ehr/fhir-hooks'

function App() {
  return (
    <FhirProvider baseUrl="https://fhir.example.com/r4">
      <PatientList />
    </FhirProvider>
  )
}

function PatientList() {
  const { data, loading, error } = useFhirSearch('Patient', {
    name: 'Smith',
    _count: 10,
  })

  if (loading) return <div>Loading...</div>
  if (error) return <div>Error: {error.message}</div>

  return data.entry?.map(e => <div key={e.resource?.id}>{e.resource?.id}</div>)
}

function PatientDetail({ id }: { id: string }) {
  const { data: patient } = useFhirRead('Patient', id)
  return <div>{patient?.name?.[0]?.family}</div>
}
```

## License

Apache-2.0
