# @ehr/ui — Implementation Plan

> Beautiful. Modern. Accessible. Fast. Zero lock-in.
> Concrete decisions. No hand-waving.

---

## Tech Stack Decisions (With Rationale)

### Why Each Choice — No Lock-In Test

Every tool choice passes one test: **"Can a developer replace this
without rewriting our components?"** If no, we don't use it.

```
+-------------------------------------------------------------------+
| Decision          | Choice           | Lock-in? | Replaceable?     |
|-------------------|------------------|----------|------------------|
| Monorepo          | pnpm workspaces  | No       | Any pkg manager  |
| Build             | tsup             | No       | Any bundler      |
| Types             | TypeScript 5     | No*      | Industry std     |
| Testing           | Vitest           | No       | Jest-compatible  |
| Component tests   | Testing Library  | No       | Industry std     |
| Accessibility     | axe-core         | No       | Industry std     |
| Docs              | Storybook 8      | No       | Any doc tool     |
| Styling           | CSS custom props | NO       | Pure CSS std     |
| CSS delivery      | Plain .css files | NO       | Works everywhere |
| Theming           | CSS variables    | NO       | Browser native   |
| Framework         | React 18/19      | Yes*     | React is the bet |
| Linting           | Biome            | No       | ESLint-compat    |
| CI                | GitHub Actions   | No       | Any CI           |
+-------------------------------------------------------------------+

  * TypeScript: industry standard, not a lock-in
  * React: the library IS React components, this is inherent
```

### Styling: The Critical Decision

**We use plain CSS with CSS custom properties. Not Tailwind. Not CSS-in-JS. Not Mantine.**

Why:

```
  Tailwind   → requires build pipeline config, PostCSS, class conflicts
  CSS-in-JS  → runtime cost, SSR complexity, dying ecosystem (Emotion/styled)
  Mantine    → framework lock-in (Medplum's mistake)
  MUI        → framework lock-in, heavy bundle

  Plain CSS + CSS custom properties:
    → Zero runtime cost
    → Works with Next.js, Vite, Remix, CRA, Astro, anything
    → No build config needed (just import the .css file)
    → Themeable via CSS variable overrides
    → No class name conflicts (scoped with data attributes)
    → SSR works out of the box (no hydration mismatch)
    → Developers override with their own CSS, Tailwind, whatever
    → Browser-native. Will never be deprecated.
```

How it works:

```css
/* @ehr/react/styles.css — import once */

/* Tokens as CSS custom properties */
:root, [data-ehr-theme] {
  --ehr-color-primary: #2563eb;
  --ehr-color-danger: #dc2626;
  --ehr-radius-md: 6px;
  --ehr-font-sans: Inter, system-ui, sans-serif;
  --ehr-space-4: 1rem;
  /* ... 200 tokens */
}

/* Component styles — scoped by data attribute */
[data-ehr="patient-banner"] {
  display: flex;
  align-items: center;
  gap: var(--ehr-space-4);
  padding: var(--ehr-space-4);
  border: 1px solid var(--ehr-border-default);
  border-radius: var(--ehr-radius-lg);
  font-family: var(--ehr-font-sans);
}

[data-ehr="patient-banner"] [data-ehr="name"] {
  font-size: var(--ehr-text-xl);
  font-weight: var(--ehr-weight-semibold);
  color: var(--ehr-fg-primary);
}
```

Developer overrides:

```css
/* Your app.css — override any token */
:root {
  --ehr-color-primary: #0055b8;  /* your brand blue */
  --ehr-radius-md: 8px;          /* rounder corners */
  --ehr-font-sans: Roboto, sans-serif;
}

/* Or override a specific component */
[data-ehr="patient-banner"] {
  background: #f0f4ff;
}
```

```tsx
// Or use className for one-off overrides
<PatientBanner className="my-custom-banner" patientId="123" />
```

```tsx
// Tailwind users can override with utility classes:
<PatientBanner className="bg-blue-50 rounded-2xl shadow-lg" patientId="123" />
```

**Result: Our components look beautiful out of the box. But the developer
can override ANYTHING with plain CSS, Tailwind, or any method. Zero lock-in.**

### Beauty: The Visual Standard

```
  We target the visual quality of:
    - Linear.app      (clean, fast, modern)
    - Vercel dashboard (minimal, professional)
    - Stripe docs     (readable, structured)
    - shadcn/ui       (composable, neutral, elegant)

  Applied to clinical data:
    - High information density without visual clutter
    - Clear visual hierarchy (patient name > MRN > dates)
    - Subtle animations (expand/collapse, loading states)
    - Thoughtful whitespace (breathable despite dense data)
    - Color used for meaning, not decoration

  Specific visual choices:
    - Font: Inter (most legible screen font, open source, variable)
    - Border radius: 6px default (modern, not childish)
    - Shadows: subtle, layered (depth without heaviness)
    - Colors: muted backgrounds, vibrant only for alerts/status
    - Transitions: 150ms ease-out (responsive, not slow)
    - Icons: Lucide (MIT, consistent, 1000+, tree-shakeable)
```

---

## Performance Targets

```
+-------------------------------------------------------------------+
| Metric                    | Target          | How We Enforce       |
|---------------------------|-----------------|----------------------|
| Bundle size (core)        | < 5 KB gzipped  | tsup + size-limit    |
| Bundle size (full import) | < 45 KB gzipped | tree-shaking + split |
| Single component import   | < 2 KB gzipped  | per-component entry  |
| CSS file                  | < 12 KB gzipped | no utility bloat     |
| First paint (component)   | < 16ms          | no runtime CSS gen   |
| Re-render                 | < 4ms           | memo + stable refs   |
| Hydration cost (SSR)      | Zero CSS mismatch| CSS is static        |
| FHIR fetch (useResource)  | Cached in 1ms   | SWR-style cache      |
| Time to Interactive       | < 100ms         | no large deps        |
| Lighthouse Performance    | > 95             | CI enforcement       |
+-------------------------------------------------------------------+
```

How we achieve these:

```
1. ZERO RUNTIME CSS
   No styled-components, no Emotion, no CSS-in-JS runtime.
   CSS is static files. Parsed once by browser. Costs nothing.

2. TREE-SHAKING
   Each component is a separate entry point:
     import { PatientBanner } from '@ehr/react'
     // Only PatientBanner code + its deps are bundled
     // Not the entire library

   package.json exports map:
   {
     "exports": {
       ".": "./dist/index.js",
       "./PatientBanner": "./dist/PatientBanner.js",
       "./ProblemList": "./dist/ProblemList.js",
       ...
     }
   }

3. LAZY LOADING HEAVY COMPONENTS
   Components with chart rendering (LabSparkline, VitalsFlowsheet)
   lazy-load their chart library:

   const Chart = React.lazy(() => import('./internal/chart'))

   Cost of LabResults without charts: ~1.5 KB
   Chart loads only when trend sparklines are visible.

4. MEMOIZATION
   Every component uses React.memo with custom comparators.
   FHIR resources are compared by versionId, not deep equality.

   Re-render only happens when the resource actually changes.

5. VIRTUALIZATION FOR LARGE LISTS
   ProblemList with 500 conditions? Virtualized.
   AuditLog with 10,000 entries? Virtualized.
   Only visible rows render. @tanstack/virtual (3 KB).

6. SIZE BUDGET CI CHECK
   Every PR runs size-limit:
   - If a component exceeds its size budget, PR is blocked
   - Budget is set per-component in .size-limit.json
```

---

## Accessibility Targets

```
+-------------------------------------------------------------------+
| Standard         | Target    | Enforcement                        |
|------------------|-----------|------------------------------------|
| WCAG version     | 2.1 AA    | axe-core in every test             |
| Contrast ratio   | >= 4.5:1  | automated check on every color     |
| Keyboard nav     | Full      | tab/arrow/enter/escape tests       |
| Screen reader    | Full      | ARIA roles + live regions          |
| Focus management | Visible   | focus-visible + custom indicators  |
| Reduced motion   | Respected | prefers-reduced-motion media query |
| High contrast    | Supported | forced-colors media query          |
| RTL support      | Full      | logical properties (CSS dir)       |
| Font scaling     | Up to 200%| rem units, no px for text          |
| Touch targets    | >= 44px   | minimum interactive size           |
+-------------------------------------------------------------------+
```

How we enforce:

```
1. EVERY COMPONENT TEST includes axe-core:

   import { axe } from 'vitest-axe'

   it('has no accessibility violations', async () => {
     const { container } = render(<PatientBanner patient={mockPatient} />)
     const results = await axe(container)
     expect(results).toHaveNoViolations()
   })

2. EVERY INTERACTIVE ELEMENT has keyboard tests:

   it('can be navigated with keyboard', async () => {
     render(<MedicationList patientId="123" />)
     await userEvent.tab()  // focus first row
     await userEvent.keyboard('{ArrowDown}')  // next row
     await userEvent.keyboard('{Enter}')  // activate
   })

3. EVERY COMPONENT has ARIA roles:

   <table role="table" aria-label="Active medications">
     <thead>
       <tr>
         <th scope="col" aria-sort="ascending">Medication</th>
       </tr>
     </thead>
     <tbody>
       <tr role="row" aria-selected={selected}>
         <td>Lisinopril 10mg</td>
       </tr>
     </tbody>
   </table>

4. DYNAMIC CONTENT uses live regions:

   // When new lab results load:
   <div aria-live="polite" aria-atomic="true">
     3 new lab results loaded
   </div>

   // When a critical alert appears:
   <div role="alert" aria-live="assertive">
     Critical drug interaction detected
   </div>

5. FOCUS TRAP in modals/dialogs:

   // DrugInteractionAlert modal traps focus
   // Escape key closes it
   // Focus returns to trigger element after close

6. STORYBOOK a11y addon shows violations in real-time during development
```

---

## Build Pipeline

```
+-------------------------------------------------------------------+
|                        BUILD PIPELINE                             |
+-------------------------------------------------------------------+
|                                                                   |
|  Source                                                           |
|  packages/                                                        |
|    tokens/src/*.ts                                                |
|    primitives/src/*.tsx                                            |
|    fhir-types/src/*.ts                                            |
|    fhir-hooks/src/*.ts                                            |
|    react-core/src/*.tsx                                           |
|    react/src/*.tsx + *.css                                        |
|        |                                                          |
|        v                                                          |
|  tsup (per-package)                                               |
|    → ESM (.mjs)    — for modern bundlers                          |
|    → CJS (.cjs)    — for Node.js / Jest                           |
|    → Types (.d.ts)  — TypeScript declarations                     |
|    → CSS (.css)     — component styles (react package only)       |
|        |                                                          |
|        v                                                          |
|  Quality Gates (must all pass)                                    |
|    → vitest          — unit + component tests                     |
|    → vitest-axe      — accessibility violations                   |
|    → size-limit      — bundle size budget                         |
|    → tsc --noEmit    — type checking                              |
|    → biome check     — lint + format                              |
|    → publint         — package.json correctness                   |
|        |                                                          |
|        v                                                          |
|  Storybook Build                                                  |
|    → Static site with every component                             |
|    → Deployed to ui.your-ehr.com                                  |
|        |                                                          |
|        v                                                          |
|  npm publish (changesets)                                         |
|    → Semantic versioning                                          |
|    → Changelog generation                                         |
|    → GitHub release                                               |
|                                                                   |
+-------------------------------------------------------------------+
```

### Monorepo Structure

```
web/
  package.json               # pnpm workspace root
  pnpm-workspace.yaml
  tsconfig.base.json         # shared TS config
  vitest.workspace.ts        # shared test config
  .size-limit.json           # bundle size budgets

  packages/
    tokens/
      src/
        colors.ts
        spacing.ts
        typography.ts
        radius.ts
        shadows.ts
        z-index.ts
        motion.ts
        breakpoints.ts
        clinical.ts          # healthcare-specific tokens
        index.ts
      package.json
      tsconfig.json
      tsup.config.ts

    fhir-types/
      src/
        r4/
          primitives.ts      # string, boolean, date, etc.
          patient.ts
          observation.ts
          condition.ts
          ... (one file per resource type)
        index.ts
      package.json

    fhir-hooks/
      src/
        provider.tsx         # FHIRProvider context
        client.ts            # FHIR REST client (fetch-based)
        cache.ts             # SWR-style cache
        hooks/
          use-resource.ts
          use-search.ts
          use-terminology.ts
          use-reference.ts
          use-subscription.ts
          use-pagination.ts
          use-capability.ts
        index.ts
      src/__tests__/
        use-resource.test.ts
        use-search.test.ts
        ...
      package.json

    primitives/
      src/
        box.tsx
        stack.tsx
        text.tsx
        button.tsx
        input.tsx
        select.tsx
        combobox.tsx
        dialog.tsx
        data-table.tsx
        badge.tsx
        alert.tsx
        ...
        styles/
          primitives.css     # base primitive styles
        index.ts
      package.json

    react/                   # the main styled package
      src/
        theme/
          provider.tsx       # ThemeProvider
          tokens.css         # all CSS custom properties
          themes/
            default.css
            clinical.css
            dark.css
            high-contrast.css
        fhir-primitives/     # FHIR data type components
          human-name.tsx
          human-name.css
          codeable-concept.tsx
          codeable-concept.css
          identifier.tsx
          quantity.tsx
          reference.tsx
          ...
        resources/           # FHIR resource components
          patient-banner.tsx
          patient-banner.css
          problem-list.tsx
          problem-list.css
          medication-list.tsx
          allergy-list.tsx
          vitals-panel.tsx
          lab-results.tsx
          clinical-timeline.tsx
          ...
        workflows/           # clinical workflow components
          terminology-search.tsx
          questionnaire-form.tsx
          resource-search.tsx
          resource-form.tsx
          ...
        index.ts
        styles.css           # single import for all styles
      package.json

    smart-auth/
      src/
        launch.tsx
        callback.tsx
        use-smart-auth.ts
        pkce.ts
      package.json

  apps/
    storybook/               # Storybook documentation app
      .storybook/
        main.ts
        preview.ts
      stories/
        primitives/
        fhir-primitives/
        resources/
        workflows/

    playground/              # live playground app (like shadcn)
      src/
        app.tsx
        examples/
```

---

## Implementation Phases

### Phase 0: Scaffolding (Week 1)

```
Goal: Monorepo builds, tests run, one component renders.

Tasks:
  [x] pnpm workspace + turbo config
  [x] tsconfig.base.json with strict mode
  [x] tsup configs for each package
  [x] vitest workspace config
  [x] biome config (lint + format)
  [x] size-limit config
  [x] CI pipeline (GitHub Actions)
  [x] Storybook 8 setup
  [x] One component: <Box> renders a div with CSS vars
  [x] One test: Box renders without a11y violations
  [x] One story: Box in Storybook with controls

Deliverable: `pnpm build && pnpm test` works.
Not shipped to npm yet. Internal only.
```

### Phase 1: Tokens + Types + Client (Week 2-3)

```
Goal: @ehr/tokens, @ehr/fhir-types, and FHIRClient ship to npm.

@ehr/tokens:
  - All color scales (8 palettes x 11 steps)
  - Spacing scale
  - Typography scale
  - Radius, shadow, z-index, motion
  - Clinical tokens (severity, status, lab flags)
  - CSS file with all custom properties
  - TypeScript token types for type-safe access
  Tests: Token values are valid CSS, no duplicates

@ehr/fhir-types:
  - Top 30 resource types fully typed:
    Patient, Practitioner, Organization, Encounter,
    Condition, Observation, MedicationRequest, MedicationStatement,
    AllergyIntolerance, Procedure, DiagnosticReport, Immunization,
    CarePlan, CareTeam, Appointment, Schedule, Slot,
    DocumentReference, Composition, Consent,
    Claim, ExplanationOfBenefit, Coverage,
    Task, ServiceRequest, Goal, FamilyMemberHistory,
    RelatedPerson, Device, Questionnaire, QuestionnaireResponse
  - Common types: Bundle, Reference, CodeableConcept, HumanName,
    Address, ContactPoint, Identifier, Quantity, Period, Annotation,
    Attachment, Dosage, Timing, Money, Range, Ratio, Narrative
  - Resource type discriminator union
  Tests: Types compile, match FHIR spec structure definitions

FHIR Client (internal, inside fhir-hooks):
  - fetch-based, zero dependencies
  - read(type, id), search(type, params), create, update, delete
  - Bundle pagination (next/prev links)
  - Operation calls ($expand, $translate, $everything, $export)
  - Auth header injection (bearer token, basic, none)
  - Error parsing (OperationOutcome → structured error)
  - Request deduplication (same URL in flight = one request)
  Tests: Client against mock server, error handling, auth

Deliverable: Three packages on npm. Blog post:
  "TypeScript autocomplete for every FHIR resource."
```

### Phase 2: Hooks (Week 3-5)

```
Goal: @ehr/fhir-hooks ships with 8 hooks.

Hooks to build (in order):

  1. FHIRProvider + useFHIRClient
     - Context provider, configures server URL + auth
     - Test: provider renders, client is accessible

  2. useResource<T>(type, id)
     - Fetch single resource, typed, cached
     - Loading/error states
     - Refetch on id change
     - Test: fetches patient, caches, handles 404

  3. useSearch<T>(type, params)
     - FHIR search with typed params
     - Automatic pagination (hasNext, next, prev)
     - _include flattening (included resources accessible)
     - Test: search with pagination, _include resolution

  4. useReference(ref)
     - Resolve a FHIR Reference to its target resource
     - Cache shared across components (same ref = one fetch)
     - Test: resolves reference, caches, handles missing

  5. useTerminology(valueSetUrl, { filter, count })
     - Calls ValueSet/$expand with debounced filter
     - Returns Coding[] for autocomplete
     - Test: expand search, debounce, empty results

  6. usePagination(bundle)
     - Navigate Bundle next/prev links
     - Page number tracking
     - Test: paginate forward/back, boundary conditions

  7. useCapabilityStatement()
     - Fetch CapabilityStatement from /metadata
     - Cache forever (doesn't change)
     - Test: fetches and caches

  8. useSubscription(criteria, callback)
     - WebSocket-based FHIR subscription
     - Auto-reconnect
     - Cleanup on unmount
     - Test: connects, receives event, reconnects, cleans up

Performance:
  - SWR-style: stale-while-revalidate caching
  - Request deduplication
  - Configurable cache TTL
  - No unnecessary re-renders (stable references)

Deliverable: @ehr/fhir-hooks on npm. Blog post:
  "useResource() — typed FHIR data in one line of React."
```

### Phase 3: Primitives (Week 4-6)

```
Goal: @ehr/primitives ships with 20 core primitives.

Build order (each includes: component + CSS + tests + story):

  Batch 1 — Layout (Week 4):
    1.  Box          — base layout, CSS var consumer
    2.  Stack        — vertical/horizontal flex
    3.  Inline       — inline flex with wrapping
    4.  Grid         — CSS grid with token-based columns
    5.  Divider      — horizontal/vertical separator

  Batch 2 — Typography + Feedback (Week 4):
    6.  Text         — body text with size/weight/color
    7.  Heading      — h1-h6 with semantic tokens
    8.  Badge        — status/label chip
    9.  Alert        — info/success/warning/danger banners
    10. Spinner      — loading indicator
    11. Skeleton     — loading placeholder shapes

  Batch 3 — Forms (Week 5):
    12. Button       — primary/secondary/danger/ghost variants
    13. Input        — text input with label/error/hint
    14. Select       — native select with styling
    15. Checkbox     — accessible checkbox
    16. Textarea     — multi-line input

  Batch 4 — Overlay + Data (Week 5-6):
    17. Dialog       — modal with focus trap, escape to close
    18. Popover      — positioned popup
    19. Tooltip      — accessible tooltip
    20. DataTable    — sortable, paginated table

  Each primitive:
    - Pure CSS styling via data attributes + CSS custom properties
    - Full ARIA roles and keyboard navigation
    - axe-core test (zero violations)
    - Keyboard navigation test
    - Storybook story with controls
    - Size budget check

Deliverable: @ehr/primitives on npm.
```

### Phase 4: FHIR Primitives — The Data Type Components (Week 6-8)

```
Goal: 12 display + 6 input components for FHIR data types.

Display components (render FHIR data types beautifully):

  1.  <HumanName>          — name formatting with use labels
  2.  <Address>            — address formatting (full/inline/city-state)
  3.  <ContactPoint>       — phone/email/fax with icons
  4.  <CodeableConcept>    — coded value with tooltip details
  5.  <Identifier>         — MRN/NPI with type labels, masking
  6.  <Quantity>           — value + unit + comparator
  7.  <Period>             — date range (full/short/relative)
  8.  <Reference>          — clickable link, auto-resolves display
  9.  <Annotation>         — clinical notes with author/time
  10. <Dosage>             — medication dosage (narrative/structured)
  11. <Attachment>         — file preview/download
  12. <StatusBadge>        — FHIR status with clinical color tokens

Input components (FHIR data type editors):

  1.  <HumanNameInput>     — structured name entry
  2.  <AddressInput>       — structured address entry
  3.  <ContactPointInput>  — phone/email entry with system/use
  4.  <CodeableConceptInput> — terminology-bound autocomplete
  5.  <QuantityInput>      — value + unit picker
  6.  <IdentifierInput>    — type + system + value

Each FHIR primitive:
  - Accepts the FHIR data type directly (e.g., HumanName object)
  - Handles null/undefined/missing fields gracefully
  - Composable — use standalone or inside resource components
  - axe-core tested, keyboard navigable
  - Storybook story with real FHIR data examples
  - Size: < 1 KB per component

Deliverable: FHIR primitives in @ehr/react.
  These are the building blocks for everything in Phase 5.
```

### Phase 5: The 10 That Matter — Resource Components (Week 8-14)

```
Goal: The 10 components that solve 80% of healthcare UI needs.

Build order (highest impact first):

  1. <PatientBanner>  (Week 8)
     - Full + compact variants
     - Photo, name, DOB/age, sex, MRN, phone
     - Allergy badges (severity-colored)
     - Clinical flags (fall risk, DNR, isolation)
     - Uses: HumanName, Identifier, ContactPoint, Badge
     - Tests: renders patient, handles missing fields,
       allergy severity colors, a11y, keyboard
     - Size: < 3 KB

  2. <ProblemList>  (Week 9)
     - Active/resolved filter tabs
     - Condition rows: status dot, name, onset, severity, ICD-10
     - Click to expand details
     - Uses: CodeableConcept, StatusBadge, DataTable
     - Data: useSearch('Condition', { patient, _sort: '-onset-date' })
     - Tests: filter active/resolved, sort, empty state, a11y
     - Size: < 2.5 KB

  3. <MedicationList>  (Week 9)
     - Active/stopped filter
     - Rows: drug name, dosage, frequency, prescriber, refills
     - Drug interaction warning indicator
     - Uses: CodeableConcept, Dosage, Reference, Badge
     - Data: useSearch('MedicationRequest', { patient, status: 'active' })
     - Tests: filter, dosage formatting, interaction flags, a11y
     - Size: < 2.5 KB

  4. <AllergyList>  (Week 10)
     - Severity badges with clinical token colors
     - Substance, type (allergy/intolerance), reaction, onset
     - Critical allergies highlighted
     - Uses: CodeableConcept, StatusBadge, clinical.severity tokens
     - Data: useSearch('AllergyIntolerance', { patient })
     - Tests: severity coloring, critical highlighting, a11y
     - Size: < 2 KB

  5. <VitalsPanel>  (Week 10)
     - Grid layout: BP, HR, Temp, SpO2, Weight, BMI
     - Abnormal value highlighting (clinical.lab tokens)
     - Reference range comparison
     - Uses: Quantity, Badge, Grid
     - Data: useSearch('Observation', { patient, category: 'vital-signs' })
     - Tests: abnormal flagging, missing values, a11y
     - Size: < 2.5 KB

  6. <LabResults>  (Week 11-12) — MOST COMPLEX
     - Group by panel (CBC, BMP, Lipid, etc.)
     - Reference ranges with abnormal flags (H, L, HH, LL)
     - Flag coloring from clinical.lab tokens
     - Trend sparkline (lazy-loaded, < 2 KB)
     - Expandable panel rows
     - Uses: Quantity, CodeableConcept, Badge, DataTable, Accordion
     - Data: useSearch('Observation', { patient, category: 'laboratory' })
     - Logic: groupByPanel(), flagAbnormal(), calculateTrend()
     - Tests: flag logic (15+ edge cases), grouping, trends, a11y
     - Size: < 4 KB (+ 2 KB lazy chart)

  7. <TerminologySearch>  (Week 12)
     - Autocomplete backed by useTerminology hook
     - Recent codes, favorites
     - Multi-system display (SNOMED + ICD-10 together)
     - Group by code system
     - Uses: Combobox, CodeableConcept, Badge
     - Tests: search, debounce, select, recent, empty, a11y
     - Size: < 3 KB

  8. <ClinicalTimeline>  (Week 13)
     - Multi-resource patient event timeline
     - Resource type filter tabs
     - Date grouping (today, this week, this month, older)
     - Resource-type-specific rendering (lab shows value, med shows dose)
     - Uses: all FHIR primitives, Badge, StatusBadge
     - Data: useSearch for multiple resource types
     - Tests: filtering, date grouping, resource rendering, a11y
     - Size: < 4 KB

  9. <ResourceSearch>  (Week 13-14)
     - Generic FHIR search with filter chips
     - Column configuration
     - Sorting by any column
     - Pagination with usePagination
     - CSV/NDJSON export
     - Uses: DataTable, Badge, Input, Select
     - Tests: filter add/remove, sort, paginate, export, a11y
     - Size: < 3.5 KB

  10. <QuestionnaireForm>  (Week 14)
     - FHIR Questionnaire renderer
     - Item types: string, integer, decimal, boolean, date,
       choice, open-choice, group, display
     - enableWhen conditional display
     - Required field validation
     - Scoring (calculated fields)
     - Produces QuestionnaireResponse on submit
     - Uses: Input, Select, Checkbox, Radio, Textarea, Button, Form
     - Tests: all item types, conditionals, validation, scoring, a11y
     - Size: < 5 KB

Deliverable: @ehr/react with 10 resource components + 18 FHIR primitives
  + 20 core primitives = 48 production-ready components.
  Blog post: "Build a patient chart in 6 lines of React."
```

### Phase 6: Polish, Docs, Launch (Week 14-16)

```
Goal: Production-ready v1.0 launch.

  Storybook site:
    - Every component with interactive controls
    - Real FHIR data examples (from public HAPI FHIR server)
    - Clinical context documentation (what is this component FOR)
    - Code snippets for copy-paste
    - Accessibility audit results shown per component
    - Theme switcher (default / clinical / dark / high-contrast)
    - Density toggle (comfortable / compact / dense)

  README:
    - "5-minute quickstart" with copy-paste code
    - Screenshot/GIF of PatientBanner + LabResults rendering
    - "Why @ehr/react" section (the value prop, 3 sentences)
    - Link to Storybook, API docs, contributing guide

  Performance audit:
    - size-limit checks pass for every component
    - Lighthouse > 95 for Storybook pages
    - No runtime CSS generation confirmed
    - Tree-shaking verified (import one component, check bundle)

  Cross-server testing:
    - Verified against: public HAPI FHIR, our EHR backend
    - Document any server-specific quirks

  v1.0 npm publish:
    - @ehr/tokens@1.0.0
    - @ehr/fhir-types@1.0.0
    - @ehr/fhir-hooks@1.0.0
    - @ehr/primitives@1.0.0
    - @ehr/react@1.0.0
```

---

## Phase 7+ (Post-Launch Roadmap)

```
Month 4-5:
  - @ehr/smart-auth (SMART on FHIR launch flows)
  - 5 more resource components (EncounterSummary, ImmunizationRecord,
    CareTeamCard, CarePlanView, AppointmentCard)
  - Dark mode theme
  - Mobile-responsive variants

Month 5-6:
  - SOAPNote (clinical documentation editor)
  - CDSHooksCard (decision support renderer)
  - MedicationReconciliation
  - DrugInteractionAlert
  - VitalsFlowsheet (ICU time-series grid)

Month 6-8:
  - ResourceForm (auto-generated from StructureDefinition)
  - QuestionnaireBuilder (drag-and-drop designer)
  - BulkExportStatus ($export dashboard)
  - AuditLog viewer

Month 8-12:
  - Scheduler (appointment booking with slot availability)
  - PatientSearch (demographics matching)
  - EHRShell (full application layout)
  - PatientSummary (composable dashboard)
  - Community contributions
```

---

## Component Quality Checklist

Every component must pass ALL of these before merge:

```
+-------------------------------------------------------------------+
| #  | Check                                    | Tool              |
|----|------------------------------------------|-------------------|
|  1 | Renders without errors                   | vitest            |
|  2 | Zero accessibility violations            | vitest-axe        |
|  3 | Full keyboard navigation                 | testing-library   |
|  4 | Screen reader announcements              | manual + ARIA     |
|  5 | Handles null/undefined/empty data        | vitest            |
|  6 | Handles loading state                    | vitest            |
|  7 | Handles error state                      | vitest            |
|  8 | Under size budget                        | size-limit        |
|  9 | TypeScript strict (no any)               | tsc               |
| 10 | Storybook story with controls            | storybook         |
| 11 | Storybook story with real FHIR data      | storybook         |
| 12 | Dark mode renders correctly              | visual + story    |
| 13 | High contrast renders correctly          | visual + story    |
| 14 | Compact density renders correctly        | visual + story    |
| 15 | className override works                 | vitest            |
| 16 | CSS custom property override works       | vitest            |
| 17 | Server-side renders (no hydration error) | vitest + next     |
| 18 | No runtime CSS generation                | bundle audit      |
| 19 | Responsive (works at 320px - 2560px)     | storybook viewport|
| 20 | Biome lint + format passes               | biome             |
+-------------------------------------------------------------------+
```

---

## Timeline Summary

```
+===================================================================+
|                                                                   |
|  Week  1      : Scaffolding (monorepo, build, CI)                 |
|  Week  2-3    : Tokens + FHIR types + FHIR client                |
|  Week  3-5    : 8 React hooks                                     |
|  Week  4-6    : 20 core primitives                                |
|  Week  6-8    : 18 FHIR data type components                     |
|  Week  8-14   : 10 resource/workflow components                   |
|  Week  14-16  : Polish, docs, Storybook, v1.0 launch             |
|                                                                   |
|  Total: 16 weeks to v1.0                                          |
|                                                                   |
|  v1.0 ships: 48 components + 8 hooks + tokens + types             |
|  Bundle: < 45 KB gzipped total, < 2 KB per component             |
|  Tests: 100% a11y pass, keyboard nav, dark mode, responsive      |
|  Docs: Full Storybook with live FHIR data                         |
|  Servers: Verified against HAPI FHIR + our EHR backend            |
|                                                                   |
+===================================================================+
```
