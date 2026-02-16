# Why @ehr/react — The Real Value Proposition

> Not marketing. Not hype. The exact reasons a developer
> installs this package today and never removes it.

---

## The Honest Problem

A developer gets hired to build a healthcare app. Day 1, they search npm:

```
npm search fhir react
```

They find:
- **@medplum/react** — 120 components, but forces Mantine UI + Medplum server
- **fhir-react** — 35 display components, no TypeScript, class components, stale
- **material-fhir-ui** — abandoned 5 years ago
- **fhir-ui** — abandoned 3 years ago
- **Terra UI** — archived by Oracle

So they do what every healthcare dev does: **build from scratch.** They spend
3 weeks writing a patient banner. 2 weeks on a terminology picker. 1 month on
a questionnaire renderer. 6 weeks on lab result display with abnormal flagging.

This happens at every healthcare company. Every team. Every time.

**That is the problem we solve.**

---

## The Exact Value — Package by Package

### Package 1: `@ehr/fhir-types` — Install in 10 seconds, never remove

**What it is:** TypeScript type definitions for every FHIR R4 resource.

**Why someone installs it today:**

```
Before @ehr/fhir-types:

  const patient = await fetch('/fhir/Patient/123')
  const data = await patient.json()
  // data is `any`
  // What fields does Patient have?
  // Is it data.name or data.names?
  // Is name a string or an object?
  // What shape is the object?
  // *opens hl7.org/fhir in another tab*
  // *spends 20 minutes reading spec*
  console.log(data.name[0].given[0])  // runtime error if name is empty
```

```
After @ehr/fhir-types:

  import type { Patient } from '@ehr/fhir-types'

  const patient: Patient = await fetch('/fhir/Patient/123').then(r => r.json())
  //     ^ autocomplete shows every field
  //       name?: HumanName[]
  //       birthDate?: string
  //       telecom?: ContactPoint[]
  //       ... full IDE support

  patient.name?.[0]?.given?.[0]  // TypeScript catches nullability
```

**Differentiator vs alternatives:**
- `@types/fhir` exists but is auto-generated, bloated, has wrong nullability
- `@medplum/fhirtypes` exists but pulls in Medplum's module system
- Ours: standalone, zero dependencies, correct optionality, tree-shakeable
  per resource type (`import type { Patient } from '@ehr/fhir-types/Patient'`)

**Adoption hook:** This is the gateway drug. Zero risk to install.
Works with any project. No opinions. Just types. Developers install it
for the autocomplete and never remove it.

**Metric to prove value:** "How many minutes did you spend on hl7.org
looking up field names this week?" Answer should be zero.

---

### Package 2: `@ehr/fhir-hooks` — The 500 lines you always rewrite

**What it is:** React hooks for FHIR data operations.

**Why someone installs it today:**

Every React app talking to a FHIR server writes these same patterns:

```
Before @ehr/fhir-hooks:

  // Every component, every time:
  const [patient, setPatient] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    setLoading(true)
    fetch(`${FHIR_BASE}/Patient/${id}`, {
      headers: { Authorization: `Bearer ${token}` }
    })
      .then(r => {
        if (!r.ok) throw new Error(r.statusText)
        return r.json()
      })
      .then(data => { setPatient(data); setLoading(false) })
      .catch(e => { setError(e); setLoading(false) })
  }, [id])

  // Then for search:
  // - pagination with Bundle.link
  // - _include resolution
  // - debounced search
  // - cache invalidation
  // - error handling
  // ... 200 more lines per component
```

```
After @ehr/fhir-hooks:

  import { useResource, useSearch } from '@ehr/fhir-hooks'

  // One line. Typed. Cached. Error-handled.
  const { data: patient, loading, error } = useResource<Patient>('Patient', id)

  // Search with pagination built in:
  const { bundle, hasNext, next } = useSearch<Observation>('Observation', {
    patient: id,
    category: 'laboratory',
    _sort: '-date',
    _count: 20,
  })
```

**The hooks that don't exist anywhere else:**

```ts
// Terminology search — the most rebuilt thing in healthcare
const { concepts, loading } = useTerminology(
  'http://snomed.info/sct',
  { filter: 'hypertension', count: 10 }
)
// This calls ValueSet/$expand. Every team writes this. Nobody shares it.

// Reference resolution — FHIR's biggest UX pain
const { resource } = useReference(observation.performer?.[0])
// Fetches the Practitioner/Organization that the Reference points to.
// Handles caching so the same reference isn't fetched 50 times in a list.

// Real-time subscriptions
useSubscription('Observation?patient=123', (event) => {
  // New lab result came in, update the UI
})
// Handles WebSocket connection, reconnection, cleanup.
```

**Differentiator vs alternatives:**
- SWR/React Query work for generic REST but don't understand FHIR
  Bundle pagination, _include flattening, reference resolution, or
  FHIR operation calls ($expand, $translate, $everything)
- @medplum/react-hooks exists (39K downloads/wk) but requires MedplumClient
  which only works properly with Medplum's server
- Ours: works with ANY FHIR server URL. Pass a URL, get hooks.

**Adoption hook:** Developer is already using `@ehr/fhir-types`.
They're tired of writing fetch+useState+useEffect for every resource.
They install fhir-hooks. Now they have typed, cached, paginated FHIR
data in one line.

---

### Package 3: `@ehr/react-core` — The headless clinical components

**What it is:** Unstyled component logic for clinical UI patterns.

**Why someone installs it today:**

The developer has their own design system (Tailwind, MUI, Chakra, custom).
They don't want our styling. They want the LOGIC that takes weeks to build:

```
The logic @ehr/react-core handles that you don't want to write:

  useMedicationList(patientId)
  - Fetches MedicationRequest + MedicationStatement
  - Merges them into a unified list
  - Separates active vs stopped vs on-hold
  - Resolves medication references to get drug names
  - Groups by medication for deduplication
  - Sorts by status then date
  - Returns structured data ready to render however you want

  useLabResults(patientId)
  - Fetches Observation resources with category=laboratory
  - Groups by panel (CBC, BMP, Lipid, etc.) using LOINC codes
  - Attaches reference ranges from Observation.referenceRange
  - Flags abnormal values (H, L, HH, LL) by comparing to ranges
  - Calculates trends from historical values
  - Returns structured data with flags and trends

  usePatientSummary(patientId)
  - Calls $everything or parallel fetches for 10+ resource types
  - Assembles: demographics, problems, meds, allergies, vitals,
    labs, immunizations, appointments, care team, documents
  - All in one hook. One loading state. One error boundary.
```

**Why this is hard and nobody else does it:**

The logic isn't just "fetch and display." It's healthcare domain knowledge
encoded as code:

```
Lab result flagging logic (simplified):

  Given: Observation.value = 145 mg/dL
         Observation.referenceRange.low = 70
         Observation.referenceRange.high = 100

  Is it abnormal?   145 > 100   → yes, HIGH
  Is it critical?   Is there a critical range?
                    Some labs have Observation.interpretation
                    Some have referenceRange with type=critical
                    Some have neither and you infer from magnitude
  What flag to show? H, HH, L, LL, or normal
  What color?        Red, dark red, blue, dark blue, or default

  This logic has 15+ edge cases:
  - No reference range provided
  - Reference range has only low or only high
  - String values ("positive", "negative", "reactive")
  - Coded values (CodeableConcept)
  - Ratio values (titer 1:256)
  - Component observations (BP has systolic + diastolic)
  - Age/sex-dependent reference ranges
  - Lab-specific reference ranges
```

**Every team that builds a lab results view discovers these edge cases
one by one over weeks. We encode them once. Correctly. Tested.**

**Differentiator:** This is the package no competitor offers.
Medplum has styled components (Mantine-locked). fhir-react has display
components (no logic separation). Nobody offers headless clinical logic
that works with your existing design system.

---

### Package 4: `@ehr/react` — Beautiful defaults, zero config

**What it is:** Pre-styled versions of every component. Tailwind-based.
Import and render. Looks professional immediately.

**Why someone installs it today:**

```tsx
// This is a complete patient chart. 6 lines of code.
import { PatientBanner, ProblemList, MedicationList,
         AllergyList, VitalsPanel, LabResults } from '@ehr/react'

<PatientBanner patientId="123" />
<ProblemList patientId="123" />
<MedicationList patientId="123" />
<AllergyList patientId="123" />
<VitalsPanel patientId="123" />
<LabResults patientId="123" />
```

The developer just built in 30 minutes what normally takes 3 months.

**Who uses this vs react-core:**
- Startups building fast → `@ehr/react` (pre-styled, ship today)
- Enterprises with design systems → `@ehr/react-core` (headless, own styles)
- Both stay in the ecosystem either way

---

## The Actual Differentiators (Be Specific)

### Differentiator 1: Backend-Agnostic (The Only One)

```
+-------------------------------------------------------------------+
| "Works with any FHIR server" is the single biggest differentiator |
+-------------------------------------------------------------------+

  @medplum/react     → requires MedplumClient → requires Medplum server
  Ottehr             → requires Oystehr account → requires Oystehr API
  fhir-react         → any server, but display-only, no hooks, no TS

  @ehr/react         → pass a FHIR server URL, everything works

  <FHIRProvider serverUrl="https://anything.com/fhir">
    {/* Every component works with any FHIR R4 server */}
  </FHIRProvider>

  This matters because:
  - Hospitals use HAPI FHIR, Smile CDR, Aidbox, or custom servers
  - Startups switch FHIR servers as they grow
  - Consultancies build for different clients with different servers
  - Developers evaluating products don't want vendor lock-in
```

### Differentiator 2: Headless-First (Nobody Does This for Healthcare)

```
+-------------------------------------------------------------------+
| Every competitor forces a UI framework. We don't.                 |
+-------------------------------------------------------------------+

  Medplum → forces Mantine 8. Your app uses MUI? Too bad.
  Ottehr  → custom styles. Your app uses Tailwind? Too bad.
  NHS     → NHS CSS. Your app uses Chakra? Too bad.

  @ehr/react-core → zero styling. Bring your own.

  This matters because:
  - 70% of React apps already have a design system
  - Nobody wants two competing style systems
  - Enterprise clients mandate their own design system
  - Headless UI, Radix, React Aria proved this model works
```

### Differentiator 3: Clinical Domain Knowledge as Code

```
+-------------------------------------------------------------------+
| We encode healthcare rules that take months to learn              |
+-------------------------------------------------------------------+

  Things that seem simple but aren't:

  1. "Display a patient name"
     → Which name? official? usual? There can be multiple.
     → Prefix/suffix handling (Dr., Jr., III)
     → Cultural ordering (family name first in some locales)

  2. "Show if a lab is abnormal"
     → 15+ edge cases for reference ranges (see above)

  3. "List active medications"
     → MedicationRequest vs MedicationStatement vs MedicationDispense
     → Which status codes mean "active"? (active, on-hold, completed?)
     → PRN vs scheduled vs one-time
     → Drug name from medicationCodeableConcept vs contained Medication
     → Dosage instruction formatting (sig generation)

  4. "Search for a diagnosis"
     → SNOMED-CT has 350,000+ concepts
     → ICD-10-CM has 70,000+ codes
     → Users type "heart attack" and expect "Myocardial infarction"
     → Fuzzy matching, synonyms, hierarchies
     → ValueSet/$expand with proper parameters

  5. "Show allergy severity"
     → FHIR has criticality AND reaction.severity — which to show?
     → Some allergies have no reaction recorded
     → Cross-reactivity (penicillin → amoxicillin)
     → "Intolerance" vs "allergy" distinction matters clinically

  We get these right. Once. With tests. Developers don't have to
  become FHIR experts to build healthcare apps.
```

### Differentiator 4: Matched to a Real Backend

```
+-------------------------------------------------------------------+
| Every hook maps to a real API endpoint that we built and tested   |
+-------------------------------------------------------------------+

  useTerminology()     → calls ValueSet/$expand    → we built this endpoint
  useCDSHooks()        → calls CDS Hooks service   → we built this service
  useSubscription()    → uses FHIR Subscriptions    → we built this system
  useBulkExport()      → calls $export              → we built this pipeline
  useQuestionnaire()   → calls $populate/$extract   → we built these ops
  useSMARTAuth()       → uses SMART on FHIR         → we built this server

  No other React library can say "we built both sides."
  Medplum can, but their frontend is Mantine-locked.
  HAPI FHIR has the backend but zero frontend.
  fhir-react has frontend but no backend.

  We have both. End to end. Tested together.
```

---

## The Adoption Strategy (Long-Term Game)

### Phase 1: The Gateway — `@ehr/fhir-types` (Month 1-3)

```
Goal: 10,000 weekly npm downloads

How:
  - Best FHIR TypeScript types on npm. Period.
  - Zero dependencies. Zero opinions. Zero risk.
  - Developer installs it, gets autocomplete, never removes it.
  - Blog post: "Stop looking up FHIR field names"
  - Tweet: "One import, every FHIR resource fully typed"

Metric: npm install count
Risk: Almost zero. Types don't break anything.

Why this works as entry point:
  - TypeScript types are the lowest-friction package possible
  - No behavior change, no migration, no risk
  - Works with ANY existing codebase
  - Once installed, developer sees @ehr/ in their node_modules
  - Next time they need a hook, they check if @ehr has one
```

### Phase 2: The Hook — `@ehr/fhir-hooks` (Month 3-6)

```
Goal: 5,000 weekly npm downloads

How:
  - Developer is tired of writing fetch+useState for every resource
  - They already have @ehr/fhir-types installed
  - They install @ehr/fhir-hooks and delete 200 lines of boilerplate
  - One blog post: "useResource() — FHIR data in one line"
  - One blog post: "useTerminology() — stop rebuilding code search"

Metric: npm install count, GitHub stars
Risk: Low. Hooks are additive, don't replace existing code.

The terminology hook is the killer feature:
  - Every team builds a SNOMED/ICD-10 search component
  - It takes 2-4 weeks every time
  - useTerminology() does it in one line
  - This alone justifies the install
```

### Phase 3: The Components — `@ehr/react-core` + `@ehr/react` (Month 6-12)

```
Goal: 3,000 weekly npm downloads, 2,000 GitHub stars

How:
  - Developer is using types + hooks already
  - They need a patient banner. Lab results display. Med list.
  - They install @ehr/react and have a patient chart in 30 minutes
  - OR they install @ehr/react-core and use headless hooks with their
    existing design system
  - Storybook site shows every component with live FHIR data
  - "Build a patient portal in 1 hour" tutorial

Metric: npm downloads, GitHub stars, Storybook visitors
Risk: Medium. Components are more opinionated. But headless option
      means low risk for enterprises.
```

### Phase 4: The Ecosystem (Month 12+)

```
Goal: Industry standard. 10,000+ GitHub stars.

How:
  - Community contributes new FHIR resource components
  - Plugin system for custom terminology servers
  - Integration guides for Epic, Cerner, Aidbox, HAPI
  - Conference talks at FHIR DevDays, HIMSS, React Conf
  - Healthcare startups default to @ehr/react like web devs
    default to shadcn/ui
  - Consulting firms recommend it to hospital clients
```

### The Flywheel

```
+-------------------------------------------------------------------+
|                                                                   |
|  Developer finds @ehr/fhir-types                                  |
|       |                                                           |
|       v                                                           |
|  Installs for TypeScript autocomplete (zero risk)                 |
|       |                                                           |
|       v                                                           |
|  Gets tired of writing fetch boilerplate                          |
|       |                                                           |
|       v                                                           |
|  Installs @ehr/fhir-hooks (deletes 200 lines of code)            |
|       |                                                           |
|       v                                                           |
|  Needs a lab results display with abnormal flagging               |
|       |                                                           |
|       v                                                           |
|  Installs @ehr/react (saves 3 weeks of work)                     |
|       |                                                           |
|       v                                                           |
|  Tells their team. Team adopts it.                                |
|       |                                                           |
|       v                                                           |
|  Team member joins new company. Brings @ehr/react.                |
|       |                                                           |
|       v                                                           |
|  New company adopts it. Cycle repeats.                            |
|       |                                                           |
|       v                                                           |
|  Industry standard.                                               |
|                                                                   |
+-------------------------------------------------------------------+
```

---

## What Quick Adoption Actually Looks Like

### The "5-Minute Win" — What Gets Shared on Twitter/Reddit

```tsx
// "I just built a patient chart in 6 lines of React"

import { FHIRProvider, PatientBanner, ProblemList,
         MedicationList, LabResults } from '@ehr/react'

function PatientChart({ patientId }) {
  return (
    <FHIRProvider serverUrl="https://hapi.fhir.org/baseR4">
      <PatientBanner patientId={patientId} />
      <ProblemList patientId={patientId} />
      <MedicationList patientId={patientId} />
      <LabResults patientId={patientId} />
    </FHIRProvider>
  )
}

// That's it. Works with the public HAPI FHIR test server.
// Swap the URL for your server. Everything still works.
```

This is the screenshot that gets 500 retweets.
This is the README example that gets GitHub stars.
This is the "holy shit" moment.

### The "Replace 500 Lines" Win — What Gets Internal Adoption

```
Before @ehr/fhir-hooks:                After:

  patient-api.ts        (85 lines)      useResource('Patient', id)     (1 line)
  use-patient.ts        (45 lines)      // deleted
  patient-types.ts      (120 lines)     import { Patient }             (1 line)
  lab-results-api.ts    (95 lines)      useSearch('Observation', ...)   (1 line)
  use-lab-results.ts    (60 lines)      // deleted
  lab-flag-logic.ts     (80 lines)      // built into LabResults
  terminology-search.ts (110 lines)     useTerminology(url, query)     (1 line)
  ──────────────────────────────────    ─────────────────────────────────
  595 lines of custom code              4 lines of imports

  Tech lead sees the PR: "-595 lines, +4 lines"
  Tech lead approves immediately.
```

### The "Saved Us 2 Months" Win — What Gets Executive Buy-In

```
  "We need a patient portal with:
   - Patient demographics display
   - Problem list
   - Medication list with drug interaction alerts
   - Lab results with trend charts
   - Appointment scheduling
   - Immunization record
   - Secure messaging
   - SMART on FHIR launch from Epic"

  Without @ehr/react: 3-4 months, 2 developers
  With @ehr/react:    2-3 weeks, 1 developer

  That's not an estimate. That's the component list:
  PatientBanner + ProblemList + MedicationList +
  DrugInteractionAlert + LabResults + LabSparkline +
  Scheduler + ImmunizationRecord + SMARTLaunch

  Each is a single import. Each works out of the box.
  Customization is Tailwind className overrides.
```

---

## Honest Comparison: Why Not Just Use Medplum?

```
+-------------------------------------------------------------------+
| Question                    | Medplum          | @ehr/react        |
|-----------------------------|------------------|-------------------|
| "I use HAPI FHIR"          | Won't work well  | Works perfectly    |
| "I use Smile CDR"          | Won't work well  | Works perfectly    |
| "I use my own FHIR server" | Won't work well  | Works perfectly    |
| "I use Medplum backend"    | Works perfectly   | Works perfectly    |
|                             |                  |                   |
| "I use Tailwind"           | Conflict w/Mantine| Native Tailwind   |
| "I use Material UI"        | Conflict          | Use react-core    |
| "I use Chakra UI"          | Conflict          | Use react-core    |
| "I use custom CSS"         | Conflict          | Use react-core    |
| "I use Mantine"            | Native            | Works fine too    |
|                             |                  |                   |
| "I need just types"        | Pulls in SDK      | @ehr/fhir-types   |
| "I need just hooks"        | Pulls in Mantine  | @ehr/fhir-hooks   |
| "I need one component"     | Big bundle        | Tree-shakes       |
|                             |                  |                   |
| Total components            | ~120             | ~140+             |
| Backend requirement         | Medplum server   | Any FHIR R4       |
| UI framework lock-in        | Mantine 8        | None (headless)   |
| Tree-shakeable              | Partial          | Full ESM          |
+-------------------------------------------------------------------+

  Medplum is a great product. If you use Medplum's backend AND
  you're fine with Mantine, use @medplum/react. Seriously.

  But most healthcare developers don't use Medplum. They use
  HAPI, or Smile CDR, or Aidbox, or their own server. And most
  React apps already have a design system. That's who we're for.
```

---

## The One-Line Pitch for Each Audience

```
For the developer searching npm:
  "FHIR React components that work with any server."

For the tech lead evaluating libraries:
  "Delete 500 lines of FHIR boilerplate. Replace with 5 imports."

For the CTO choosing a stack:
  "Ship a patient portal in 2 weeks instead of 4 months."

For the open-source community:
  "The shadcn/ui of healthcare — headless, composable, yours to own."

For the FHIR conference:
  "Backend-agnostic. Headless-first. Clinically complete."
```

---

## What We Are NOT

Being specific about what we are means being specific about what we aren't:

```
  We are NOT a FHIR server.
    → Use our headless EHR backend, or HAPI, or Smile CDR, or Aidbox.

  We are NOT an EHR application.
    → We give you components. You build the application.
    → Ottehr is an app. We are a library.

  We are NOT a design system from scratch.
    → We build on Tailwind (styled) or your system (headless).
    → We add healthcare semantics on top.

  We are NOT vendor-locked.
    → Switch FHIR servers. Switch CSS frameworks. We still work.

  We are NOT comprehensive on day 1.
    → We ship the 20 components that matter most first.
    → Community and usage tell us what to build next.
```

---

## The 20 Components That Matter Most (Ship First)

Not 140. Not 40. The 20 that solve the most pain:

```
+-------------------------------------------------------------------+
| #  | Component            | Why it's top 20                      |
|----|----------------------|--------------------------------------|
|  1 | FHIRProvider          | Foundation. Everything needs this.   |
|  2 | useResource           | Replaces 50 lines per component.    |
|  3 | useSearch             | Replaces 80 lines per search page.  |
|  4 | PatientBanner         | Every EHR screen has this.           |
|  5 | HumanName             | Used inside 10+ other components.    |
|  6 | CodeableConcept       | Used in every clinical component.    |
|  7 | Reference             | Used everywhere FHIR links exist.    |
|  8 | Identifier            | MRN display, patient lookup.         |
|  9 | ProblemList           | Core clinical component.             |
| 10 | MedicationList        | Core clinical component.             |
| 11 | AllergyList           | Patient safety — always visible.     |
| 12 | LabResults            | Most complex, highest value.         |
| 13 | VitalsPanel           | Used in every encounter.             |
| 14 | useTerminology        | The hook everyone rebuilds.          |
| 15 | TerminologySearch     | The component everyone rebuilds.     |
| 16 | ResourceSearch        | Generic FHIR search with filters.    |
| 17 | QuestionnaireForm     | SDC is too hard to implement alone.  |
| 18 | ClinicalTimeline      | Patient history visualization.       |
| 19 | ResourceForm          | Auto-form for any resource type.     |
| 20 | StatusBadge           | Active/resolved/draft everywhere.    |
+-------------------------------------------------------------------+

  These 20 cover 80% of use cases.
  Ship these. Get adoption. Then build the rest.
```

---

## Long-Term Game: What Makes This Unkillable

```
  Year 1:  Ship the foundation. Get 5,000 GitHub stars.
           Become the default answer to "FHIR React components."

  Year 2:  Community contributes components. Plugin ecosystem grows.
           Consulting firms build on top of it. Training courses emerge.

  Year 3:  Healthcare startups can't imagine NOT using it.
           Like how web devs can't imagine NOT using React Query or shadcn.
           Conference talks reference @ehr components by name.

  Year 5:  It's in the HL7 implementation guides as a reference
           implementation. Epic/Oracle publish SMART app templates
           using @ehr/react. Medical schools teach with it.

  The moat isn't code. Code can be copied.
  The moat is:
    1. Community (contributors, blog posts, Stack Overflow answers)
    2. Domain expertise (clinical edge cases handled correctly)
    3. Integration testing (verified against 5+ FHIR servers)
    4. Trust (hospitals don't adopt untested software)
    5. Network effects (devs bring it to new jobs)
```

---

## Summary: Why @ehr/react Wins

```
+===================================================================+
|                                                                   |
|  1. ZERO LOCK-IN                                                  |
|     Works with any FHIR server + any CSS framework.               |
|     This alone eliminates every competitor.                       |
|                                                                   |
|  2. PROGRESSIVE ADOPTION                                          |
|     Start with types (zero risk).                                 |
|     Add hooks (delete boilerplate).                               |
|     Add components (ship faster).                                 |
|     Never forced to adopt everything at once.                     |
|                                                                   |
|  3. CLINICAL DOMAIN EXPERTISE AS CODE                             |
|     Lab flagging edge cases. Medication status logic.             |
|     Name formatting rules. Terminology search patterns.           |
|     Stuff that takes months to learn, encoded in tested code.     |
|                                                                   |
|  4. THE "DELETE 500 LINES" MOMENT                                 |
|     The PR that replaces custom code with library imports.        |
|     This is the moment adoption becomes irreversible.             |
|                                                                   |
|  5. MATCHED BACKEND                                               |
|     We built both sides. Every hook has a tested API behind it.   |
|     No other project can say "we built the server AND the UI."    |
|     (And our UI still works with other servers.)                  |
|                                                                   |
+===================================================================+
```
