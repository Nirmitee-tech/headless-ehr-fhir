# @ehr/react — Healthcare React Component Library

> The industry-standard, backend-agnostic, FHIR-native React component library.
> Headless-first architecture. Works with ANY FHIR R4 server.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Package Structure](#package-structure)
3. [Layer 1: FHIR Data Primitives](#layer-1-fhir-data-primitives)
4. [Layer 2: FHIR Resource Components](#layer-2-fhir-resource-components)
5. [Layer 3: Clinical Workflow Components](#layer-3-clinical-workflow-components)
6. [Layer 4: Layout & Infrastructure](#layer-4-layout--infrastructure)
7. [React Hooks Reference](#react-hooks-reference)
8. [Competitive Analysis](#competitive-analysis)
9. [Technical Specifications](#technical-specifications)

---

## Architecture Overview

```
+-------------------------------------------------------------------+
|                        Application Layer                          |
|  (Your App — Next.js, Vite, Remix, CRA, whatever you want)       |
+-------------------------------------------------------------------+
         |              |               |              |
         v              v               v              v
+----------------+ +-----------+ +-------------+ +-----------+
| @ehr/react     | | @ehr/     | | @ehr/       | | @ehr/     |
| Pre-styled     | | react-core| | fhir-hooks  | | smart-auth|
| Tailwind/shadcn| | Headless  | | Data layer  | | OAuth/PKCE|
| components     | | unstyled  | | hooks       | | SMART     |
+----------------+ +-----------+ +-------------+ +-----------+
         |              |               |              |
         v              v               v              v
+-------------------------------------------------------------------+
|                    @ehr/fhir-types                                 |
|  TypeScript types for all 150+ FHIR R4 resources + data types     |
+-------------------------------------------------------------------+
         |
         v
+-------------------------------------------------------------------+
|                    ANY FHIR R4 Server                              |
|  Your Headless EHR | Medplum | HAPI | Smile CDR | Aidbox | etc.  |
+-------------------------------------------------------------------+
```

### Design Principles

1. **Backend-Agnostic** — connects to any FHIR R4 server via standard REST
2. **Headless-First** — logic separated from styling; bring your own design system
3. **Pre-styled Default** — Tailwind CSS + shadcn/ui patterns for instant beauty
4. **Tree-Shakeable** — import only what you use; ESM-first bundles
5. **WCAG 2.1 AA** — accessibility baked into every component from day 1
6. **TypeScript-First** — full type safety for all FHIR resources and operations
7. **Composable** — use individual components or compose into full workflows
8. **i18n-Ready** — RTL support, locale-aware formatting, translation hooks

---

## Package Structure

```
@ehr/
  fhir-types/          # TypeScript interfaces for all FHIR R4 resources
  fhir-hooks/          # React hooks for FHIR data operations
  react-core/          # Headless/unstyled component logic (Radix-style)
  react/               # Pre-styled Tailwind components (the default)
  smart-auth/          # SMART on FHIR OAuth2 launch + auth
```

### Installation

```bash
# Full pre-styled experience (most users)
npm install @ehr/react @ehr/fhir-hooks @ehr/fhir-types

# Headless — bring your own styles
npm install @ehr/react-core @ehr/fhir-hooks @ehr/fhir-types

# Add SMART on FHIR auth
npm install @ehr/smart-auth
```

### Quick Start

```tsx
import { FHIRProvider } from '@ehr/fhir-hooks'
import { PatientBanner, MedicationList } from '@ehr/react'

function App() {
  return (
    <FHIRProvider serverUrl="https://your-ehr.com/fhir">
      <PatientBanner patientId="123" />
      <MedicationList patientId="123" />
    </FHIRProvider>
  )
}
```

---

## Layer 1: FHIR Data Primitives

> 30 display components + 30 input components = 60 total
> These render individual FHIR data types. Zero backend coupling.

---

### 1.1 `<HumanName>` — FHIR HumanName Display

Renders a patient/practitioner name with proper formatting.

```
Props:
  name: HumanName | HumanName[]    # FHIR HumanName object(s)
  format?: 'full' | 'short' | 'family-first' | 'official-only'
  showUse?: boolean                # Show use label (official, usual, etc.)
  className?: string

Variants:     full              short           family-first
            "Dr. John A.      "John Smith"    "Smith, John A."
             Smith Jr."
```

**ASCII Component:**

```
+--------------------------------------------------+
| Full format (default):                           |
|                                                  |
|  Dr. John A. Smith Jr.                           |
|                                                  |
+--------------------------------------------------+

| With use label:                                  |
|                                                  |
|  Dr. John A. Smith Jr.  [official]               |
|                                                  |
+--------------------------------------------------+

| Multiple names:                                  |
|                                                  |
|  John Smith (official)                            |
|  Johnny S. (nickname)                             |
|                                                  |
+--------------------------------------------------+
```

---

### 1.2 `<Address>` — FHIR Address Display

```
Props:
  address: Address | Address[]
  format?: 'full' | 'inline' | 'city-state'
  showType?: boolean              # home, work, temp, etc.
  showMap?: boolean               # Link to Google Maps
```

**ASCII Component:**

```
+--------------------------------------------------+
| Full format:                                     |
|                                                  |
|  [home]  123 Main Street                    [map]|
|          Apt 4B                                   |
|          Springfield, IL 62704                    |
|          United States                            |
|                                                  |
+--------------------------------------------------+

| Inline format:                                   |
|                                                  |
|  123 Main St, Apt 4B, Springfield, IL 62704      |
|                                                  |
+--------------------------------------------------+

| City-state format:                               |
|                                                  |
|  Springfield, IL                                 |
|                                                  |
+--------------------------------------------------+
```

---

### 1.3 `<ContactPoint>` — Phone / Email / Fax

```
Props:
  contact: ContactPoint | ContactPoint[]
  showSystem?: boolean            # phone, email, fax icons
  clickable?: boolean             # tel: / mailto: links
  showUse?: boolean               # home, work, mobile
```

**ASCII Component:**

```
+--------------------------------------------------+
| Default:                                         |
|                                                  |
|  [phone]  (555) 123-4567  mobile                 |
|  [email]  john@example.com  work                 |
|  [fax]    (555) 123-4568  work                   |
|                                                  |
+--------------------------------------------------+

| Compact (single):                                |
|                                                  |
|  [phone] (555) 123-4567                          |
|                                                  |
+--------------------------------------------------+
```

---

### 1.4 `<CodeableConcept>` — Coded Value Display

The most important primitive — used everywhere in FHIR.

```
Props:
  value: CodeableConcept
  showSystem?: boolean            # Show coding system
  showCode?: boolean              # Show code value
  expandable?: boolean            # Expand to show all codings
  tooltip?: boolean               # Hover to see coding details
```

**ASCII Component:**

```
+--------------------------------------------------+
| Default (display text only):                     |
|                                                  |
|  Essential hypertension                          |
|                                                  |
+--------------------------------------------------+

| With tooltip on hover:                           |
|                                                  |
|  Essential hypertension [i]                      |
|  +------------------------------------------+   |
|  | SNOMED-CT: 59621000                       |   |
|  | ICD-10: I10                               |   |
|  | Display: Essential hypertension           |   |
|  +------------------------------------------+   |
|                                                  |
+--------------------------------------------------+

| Expanded (showSystem + showCode):                |
|                                                  |
|  Essential hypertension                          |
|    SNOMED-CT  59621000                           |
|    ICD-10-CM  I10                                |
|                                                  |
+--------------------------------------------------+

| As badge/chip:                                   |
|                                                  |
|  [Essential hypertension]  [Type 2 Diabetes]     |
|                                                  |
+--------------------------------------------------+
```

---

### 1.5 `<Identifier>` — MRN / SSN / NPI Display

```
Props:
  identifier: Identifier | Identifier[]
  showType?: boolean
  showSystem?: boolean
  mask?: boolean                  # Mask sensitive values (SSN)
```

**ASCII Component:**

```
+--------------------------------------------------+
| Single identifier:                               |
|                                                  |
|  MRN  12345678                                   |
|                                                  |
+--------------------------------------------------+

| Multiple identifiers:                            |
|                                                  |
|  MRN   12345678                                  |
|  SSN   ***-**-6789  [show]                       |
|  NPI   1234567890                                |
|                                                  |
+--------------------------------------------------+
```

---

### 1.6 `<Quantity>` — Value + Unit Display

```
Props:
  value: Quantity
  precision?: number
  showComparator?: boolean        # <, <=, >=, >
  showUnit?: boolean
```

**ASCII Component:**

```
+--------------------------------------------------+
| Default:                                         |
|                                                  |
|  120 mg/dL                                       |
|                                                  |
+--------------------------------------------------+

| With comparator:                                 |
|                                                  |
|  >= 120 mg/dL                                    |
|                                                  |
+--------------------------------------------------+

| Unitless:                                        |
|                                                  |
|  7.2                                             |
|                                                  |
+--------------------------------------------------+
```

---

### 1.7 `<Period>` — Date Range

```
Props:
  value: Period
  format?: 'short' | 'long' | 'relative'
  showDuration?: boolean
```

**ASCII Component:**

```
+--------------------------------------------------+
| Long format:                                     |
|                                                  |
|  Jan 15, 2025 - Mar 22, 2025  (66 days)         |
|                                                  |
+--------------------------------------------------+

| Short format:                                    |
|                                                  |
|  01/15/25 - 03/22/25                             |
|                                                  |
+--------------------------------------------------+

| Relative format:                                 |
|                                                  |
|  Started 3 months ago - Ended 1 month ago        |
|                                                  |
+--------------------------------------------------+

| Open-ended:                                      |
|                                                  |
|  Jan 15, 2025 - ongoing                          |
|                                                  |
+--------------------------------------------------+
```

---

### 1.8 `<Reference>` — FHIR Reference Link

```
Props:
  value: Reference
  resolve?: boolean               # Auto-fetch the referenced resource
  onClick?: (resource) => void
  display?: 'link' | 'badge' | 'inline'
```

**ASCII Component:**

```
+--------------------------------------------------+
| Link format (default):                           |
|                                                  |
|  Dr. Sarah Johnson ->                            |
|                                                  |
+--------------------------------------------------+

| Badge format:                                    |
|                                                  |
|  [Practitioner: Dr. Sarah Johnson]               |
|                                                  |
+--------------------------------------------------+

| Resolved with avatar:                            |
|                                                  |
|  (SJ) Dr. Sarah Johnson — Cardiology             |
|                                                  |
+--------------------------------------------------+
```

---

### 1.9 `<Attachment>` — File Preview / Download

```
Props:
  value: Attachment
  preview?: boolean               # Show inline preview for images/PDF
  downloadable?: boolean
  maxPreviewHeight?: number
```

**ASCII Component:**

```
+--------------------------------------------------+
| Image attachment:                                |
|                                                  |
|  +------------------------------------+          |
|  |                                    |          |
|  |          [Image Preview]           |          |
|  |           chest-xray.jpg           |          |
|  |                                    |          |
|  +------------------------------------+          |
|  chest-xray.jpg  2.4 MB  image/jpeg  [download] |
|                                                  |
+--------------------------------------------------+

| PDF attachment:                                  |
|                                                  |
|  [PDF]  lab-report.pdf  156 KB       [download]  |
|                                                  |
+--------------------------------------------------+

| Generic attachment:                              |
|                                                  |
|  [FILE] consent-form.docx  89 KB    [download]   |
|                                                  |
+--------------------------------------------------+
```

---

### 1.10 `<Annotation>` — Clinical Notes

```
Props:
  value: Annotation | Annotation[]
  showAuthor?: boolean
  showTimestamp?: boolean
```

**ASCII Component:**

```
+--------------------------------------------------+
| Single annotation:                               |
|                                                  |
|  "Patient reports improvement in symptoms        |
|   after medication adjustment."                  |
|                                                  |
|   — Dr. Sarah Johnson, Jan 15, 2025 2:30 PM     |
|                                                  |
+--------------------------------------------------+

| Multiple annotations (thread):                  |
|                                                  |
|  [SJ] Jan 15, 2:30 PM                           |
|  Patient reports improvement in symptoms.        |
|                                                  |
|  [MK] Jan 16, 9:15 AM                           |
|  Follow-up labs ordered. Will review in 1 week.  |
|                                                  |
+--------------------------------------------------+
```

---

### 1.11 `<Dosage>` — Medication Dosage Instructions

```
Props:
  value: Dosage | Dosage[]
  format?: 'structured' | 'narrative' | 'compact'
```

**ASCII Component:**

```
+--------------------------------------------------+
| Narrative format:                                |
|                                                  |
|  Take 1 tablet by mouth twice daily with food    |
|                                                  |
+--------------------------------------------------+

| Structured format:                               |
|                                                  |
|  Dose:      10 mg                                |
|  Route:     Oral                                 |
|  Frequency: Twice daily (BID)                    |
|  Timing:    With meals                           |
|  Max dose:  40 mg / 24 hours                     |
|                                                  |
+--------------------------------------------------+

| Compact (inline):                                |
|                                                  |
|  10mg PO BID                                     |
|                                                  |
+--------------------------------------------------+
```

---

### 1.12 `<Money>` — Currency Display

```
+--------------------------------------------------+
|  $1,250.00  USD                                  |
+--------------------------------------------------+
```

### 1.13 `<Range>` — Low-High Range

```
+--------------------------------------------------+
|  70 - 100 mg/dL                                  |
+--------------------------------------------------+
```

### 1.14 `<Ratio>` — Numerator / Denominator

```
+--------------------------------------------------+
|  1 tablet / 8 hours                              |
+--------------------------------------------------+
```

### 1.15 `<Timing>` — Schedule Display

```
+--------------------------------------------------+
| Structured:                                      |
|                                                  |
|  Every 8 hours                                   |
|  Starting Jan 15, 2025                           |
|  For 14 days                                     |
|                                                  |
+--------------------------------------------------+

| Compact:                                         |
|                                                  |
|  Q8H x 14 days                                   |
|                                                  |
+--------------------------------------------------+
```

### 1.16 `<Narrative>` — Safe XHTML Rendering

```
Props:
  value: Narrative
  sanitize?: boolean              # DOMPurify sanitization (default: true)
  maxHeight?: number              # Scrollable container
```

```
+--------------------------------------------------+
|  +----------------------------------------------+|
|  |  <div xmlns="http://www.w3.org/1999/xhtml">  ||
|  |    <p>Patient presented with acute chest      ||
|  |    pain. ECG showed ST elevation in leads     ||
|  |    II, III, aVF...</p>                        ||
|  |  </div>                                       ||
|  +----------------------------------------------+|
|  Generated  [status: generated]                  |
+--------------------------------------------------+
```

---

### 1.17-1.20 Remaining Data Primitives

```
<Age>           "45 years"
<Duration>      "30 minutes"
<Distance>      "2.5 km"
<Count>         "3 doses"
```

---

### Input Variants (1.21 - 1.50)

Every display component above has a matching `<*Input>` variant for forms.

#### `<HumanNameInput>` Example:

```
+--------------------------------------------------+
| Name                                             |
|                                                  |
|  Prefix    First          Middle                 |
|  +------+  +------------+  +----------+          |
|  | Dr.  |  | John       |  | Andrew   |          |
|  +------+  +------------+  +----------+          |
|                                                  |
|  Last            Suffix                          |
|  +-------------+  +------+                       |
|  | Smith       |  | Jr.  |                       |
|  +-------------+  +------+                       |
|                                                  |
|  Use: ( ) Usual  (x) Official  ( ) Nickname      |
|                                                  |
+--------------------------------------------------+
```

#### `<AddressInput>` Example:

```
+--------------------------------------------------+
| Address                                          |
|                                                  |
|  Street                                          |
|  +--------------------------------------------+  |
|  | 123 Main Street                            |  |
|  +--------------------------------------------+  |
|                                                  |
|  Street Line 2                                   |
|  +--------------------------------------------+  |
|  | Apt 4B                                     |  |
|  +--------------------------------------------+  |
|                                                  |
|  City              State      Zip                |
|  +---------------+ +-------+ +----------+        |
|  | Springfield   | | IL  v | | 62704    |        |
|  +---------------+ +-------+ +----------+        |
|                                                  |
|  Country                                         |
|  +--------------------------------------------+  |
|  | United States                         v    |  |
|  +--------------------------------------------+  |
|                                                  |
|  Type: (x) Physical  ( ) Postal  ( ) Both        |
|  Use:  (x) Home  ( ) Work  ( ) Temp              |
|                                                  |
+--------------------------------------------------+
```

#### `<ContactPointInput>` Example:

```
+--------------------------------------------------+
| Contact                                          |
|                                                  |
|  System       Value                   Use        |
|  +--------+   +-------------------+   +-------+  |
|  | Phone v|   | (555) 123-4567    |   | Home v|  |
|  +--------+   +-------------------+   +-------+  |
|                                                  |
|  [+ Add another contact point]                   |
|                                                  |
+--------------------------------------------------+
```

#### `<CodeableConceptInput>` — Terminology-Bound Input

```
+--------------------------------------------------+
| Diagnosis                                        |
|                                                  |
|  +--------------------------------------------+  |
|  | essential hyp_                              |  |
|  +--------------------------------------------+  |
|  | [search] Results from SNOMED-CT             |  |
|  +--------------------------------------------+  |
|  |  Essential hypertension          59621000   |  |
|  |  Essential hypertension (disorder)          |  |
|  |  ----------------------------------------  |  |
|  |  Hypertensive disorder           38341003   |  |
|  |  Hypertensive heart disease      64715009   |  |
|  |  Hypertensive crisis             70272006   |  |
|  +--------------------------------------------+  |
|                                                  |
|  Selected: [Essential hypertension x]            |
|                                                  |
+--------------------------------------------------+
```

#### `<QuantityInput>` Example:

```
+--------------------------------------------------+
| Measurement                                      |
|                                                  |
|  Value          Unit                             |
|  +-----------+  +---------------------------+    |
|  | 120       |  | mg/dL                  v  |    |
|  +-----------+  +---------------------------+    |
|                                                  |
|  Comparator: ( ) <  ( ) <=  (x) =  ( ) >=  ( ) >|
|                                                  |
+--------------------------------------------------+
```

---

## Layer 2: FHIR Resource Components

> ~40 components that render complete FHIR resources with clinical layouts.

---

### 2.1 `<PatientBanner>` — Patient Demographics Header

The most critical component in any EHR. Shown at top of every patient screen.

```
Props:
  patient: Patient | string       # Resource or patient ID
  showPhoto?: boolean
  showAge?: boolean
  showMRN?: boolean
  showAllergies?: boolean         # Show allergy flags
  showFlags?: boolean             # Show clinical flags (fall risk, etc.)
  compact?: boolean
  onClick?: () => void
```

**ASCII Component — Full:**

```
+------------------------------------------------------------------------+
|  +------+                                                              |
|  |      |  John A. Smith                    DOB: 03/15/1978 (47y)     |
|  | PHOTO|  MRN: 12345678                    Sex: Male                  |
|  |      |  SSN: ***-**-6789                 Phone: (555) 123-4567     |
|  +------+                                                              |
|                                                                        |
|  Allergies: [! Penicillin - SEVERE] [! Sulfa drugs] [! Latex]          |
|  Flags:     [FALL RISK] [DNR] [ISOLATION]                              |
|                                                                        |
+------------------------------------------------------------------------+
```

**ASCII Component — Compact:**

```
+------------------------------------------------------------------------+
|  (JS) John Smith  47M  MRN: 12345678  [! Penicillin] [FALL RISK]      |
+------------------------------------------------------------------------+
```

**ASCII Component — Emergency:**

```
+------------------------------------------------------------------------+
| [!!! CRITICAL ALLERGIES]                                               |
|  +------+                                                              |
|  |      |  John A. Smith              DOB: 03/15/1978 (47y)           |
|  | PHOTO|  MRN: 12345678              Blood: O+                       |
|  |      |  Code Status: DNR           Weight: 82 kg                   |
|  +------+                                                              |
|                                                                        |
|  [!!! Penicillin - ANAPHYLAXIS] [! Sulfa] [! Latex]                    |
|  Emergency Contact: Jane Smith (wife) (555) 987-6543                   |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 2.2 `<PatientSummary>` — Clinical Summary Dashboard

Composable dashboard assembling multiple resource views for one patient.

```
Props:
  patientId: string
  sections?: ('demographics' | 'problems' | 'medications' | 'allergies'
              | 'vitals' | 'labs' | 'immunizations' | 'appointments'
              | 'careteam' | 'documents')[]
  layout?: 'single-column' | 'two-column' | 'dashboard'
```

**ASCII Component — Two-Column Layout:**

```
+------------------------------------------------------------------------+
|  [PatientBanner — see 2.1 above]                                       |
+------------------------------------------------------------------------+
|                                |                                       |
|  ACTIVE PROBLEMS               |  CURRENT MEDICATIONS                  |
|  +--------------------------+  |  +-------------------------------+    |
|  | * Essential hypertension |  |  | Lisinopril 10mg PO daily     |    |
|  |   Onset: 2019            |  |  | Metformin 500mg PO BID       |    |
|  | * Type 2 Diabetes        |  |  | Atorvastatin 20mg PO QHS     |    |
|  |   Onset: 2020            |  |  | Aspirin 81mg PO daily        |    |
|  | * Obesity (BMI 32.1)     |  |  |                               |    |
|  |   Onset: 2018            |  |  | [View all 8 medications ->]  |    |
|  +--------------------------+  |  +-------------------------------+    |
|                                |                                       |
|  ALLERGIES                     |  RECENT VITALS (Jan 15, 2025)         |
|  +--------------------------+  |  +-------------------------------+    |
|  | [!!!] Penicillin         |  |  | BP     138/88 mmHg    [!HIGH]|    |
|  |   Reaction: Anaphylaxis  |  |  | HR     72 bpm               |    |
|  |   Severity: Severe       |  |  | Temp   98.6 F               |    |
|  | [!]  Sulfa drugs         |  |  | SpO2   98%                  |    |
|  |   Reaction: Rash         |  |  | Wt     82 kg                |    |
|  | [!]  Latex               |  |  | BMI    32.1         [!HIGH] |    |
|  +--------------------------+  |  +-------------------------------+    |
|                                |                                       |
|  RECENT LABS                   |  IMMUNIZATIONS                        |
|  +--------------------------+  |  +-------------------------------+    |
|  | HbA1c    7.2%   [!HIGH] |  |  | [x] COVID-19 Bivalent 10/23 |    |
|  | Glucose  145    [!HIGH] |  |  | [x] Influenza 2024    09/24 |    |
|  | Cr       1.1            |  |  | [x] Tdap              03/20 |    |
|  | eGFR     78             |  |  | [ ] Pneumococcal      DUE   |    |
|  | Chol     210   [!HIGH]  |  |  | [ ] Shingrix          DUE   |    |
|  | LDL      130   [!HIGH]  |  |  +-------------------------------+    |
|  +--------------------------+  |                                       |
|                                |  UPCOMING APPOINTMENTS                |
|  CARE TEAM                     |  +-------------------------------+    |
|  +--------------------------+  |  | Jan 22 10:00 AM              |    |
|  | (SJ) Dr. Sarah Johnson   |  |  | Follow-up — Dr. Johnson      |    |
|  |   PCP — Internal Med     |  |  |                               |    |
|  | (MK) Dr. Michael Kim     |  |  | Feb 05 2:30 PM               |    |
|  |   Endocrinology          |  |  | Lab draw — Quest Diagnostics |    |
|  | (RL) Rachel Lee, RN      |  |  +-------------------------------+    |
|  |   Care Coordinator       |  |                                       |
|  +--------------------------+  |                                       |
|                                |                                       |
+------------------------------------------------------------------------+
```

---

### 2.3 `<ProblemList>` — Conditions / Diagnoses

```
Props:
  conditions: Condition[] | string    # Resources or patient ID to fetch
  filter?: 'active' | 'resolved' | 'all'
  showOnset?: boolean
  showSeverity?: boolean
  showEvidence?: boolean
  editable?: boolean
  onAdd?: () => void
  onEdit?: (condition: Condition) => void
```

**ASCII Component:**

```
+------------------------------------------------------------------------+
|  Active Problems                                    [+ Add Problem]    |
+------------------------------------------------------------------------+
|  Status   | Condition                  | Onset    | Severity | ICD-10  |
|-----------|----------------------------|----------|----------|---------|
|  [active] | Essential hypertension     | 2019-03  | --       | I10     |
|  [active] | Type 2 diabetes mellitus   | 2020-06  | Moderate | E11.9   |
|  [active] | Obesity, BMI 30-39.9       | 2018-01  | --       | E66.01  |
|  [active] | Hyperlipidemia             | 2019-03  | --       | E78.5   |
+------------------------------------------------------------------------+
|  Resolved Problems                                  [show/hide]        |
+------------------------------------------------------------------------+
|  [resolved] | Acute bronchitis         | 2024-11  | Mild     | J20.9   |
|  [resolved] | Sprained right ankle     | 2023-07  | --       | S93.401 |
+------------------------------------------------------------------------+
```

---

### 2.4 `<MedicationList>` — Current & Past Medications

```
Props:
  medications: MedicationRequest[] | string
  filter?: 'active' | 'stopped' | 'all'
  showPrescriber?: boolean
  showPharmacy?: boolean
  showRefills?: boolean
  showInteractions?: boolean       # Drug-drug interaction warnings
  onPrescribe?: () => void
```

**ASCII Component:**

```
+------------------------------------------------------------------------+
|  Medications (Active)                         [+ New Rx]  [Reconcile]  |
+------------------------------------------------------------------------+
|  Medication          | Dosage       | Frequency | Prescriber  | Refills|
|----------------------|--------------|-----------|-------------|--------|
|  Lisinopril 10mg     | 1 tab PO     | Daily     | Dr. Johnson | 3 left |
|  Metformin 500mg     | 1 tab PO     | BID       | Dr. Kim     | 5 left |
|  Atorvastatin 20mg   | 1 tab PO     | QHS       | Dr. Johnson | 2 left |
|  Aspirin 81mg        | 1 tab PO     | Daily     | Dr. Johnson | OTC    |
|  [!] Amoxicillin 500mg| 1 cap PO    | TID x10d  | Dr. Patel   | 0      |
|      ^ WARNING: Patient allergic to Penicillin — cross-reactivity risk |
+------------------------------------------------------------------------+
|  Recently Stopped                                                      |
|  Glipizide 5mg       | Stopped 12/15/24 | Reason: switched to insulin |
+------------------------------------------------------------------------+
```

---

### 2.5 `<AllergyList>` — Allergy / Intolerance Registry

```
+------------------------------------------------------------------------+
|  Allergies & Intolerances                          [+ Add Allergy]     |
+------------------------------------------------------------------------+
|  Severity | Substance        | Type        | Reaction       | Onset    |
|-----------|------------------|-------------|----------------|----------|
|  [!!! ]   | Penicillin       | Allergy     | Anaphylaxis    | 2005     |
|  [!!  ]   | Sulfa drugs      | Allergy     | Rash, Hives    | 2010     |
|  [!   ]   | Latex            | Allergy     | Contact derm.  | 2015     |
|  [    ]   | Shellfish        | Intolerance | GI upset       | 2018     |
+------------------------------------------------------------------------+
|  Severity: [!!!] High  [!!] Moderate  [!] Low  [ ] Unknown            |
+------------------------------------------------------------------------+
```

---

### 2.6 `<VitalsPanel>` — Vital Signs Display

```
Props:
  observations: Observation[] | string
  showTrends?: boolean             # Sparkline trends
  showReferenceRange?: boolean
  alertOnAbnormal?: boolean
  layout?: 'grid' | 'table' | 'cards'
```

**ASCII Component — Grid Layout:**

```
+------------------------------------------------------------------------+
|  Vital Signs — Jan 15, 2025 2:30 PM          [+ Record Vitals]        |
+------------------------------------------------------------------------+
|                                                                        |
|  +------------------+  +------------------+  +------------------+      |
|  | Blood Pressure   |  | Heart Rate       |  | Temperature      |      |
|  |                  |  |                  |  |                  |      |
|  |   138 / 88       |  |     72           |  |    98.6          |      |
|  |     mmHg         |  |     bpm          |  |     degF          |      |
|  |                  |  |                  |  |                  |      |
|  |  [!HIGH]         |  |  [normal]        |  |  [normal]        |      |
|  |  ___/^^^\_____   |  |  ___/--\___      |  |  ___________     |      |
|  |  (6mo trend)     |  |  (6mo trend)     |  |  (6mo trend)     |      |
|  +------------------+  +------------------+  +------------------+      |
|                                                                        |
|  +------------------+  +------------------+  +------------------+      |
|  | SpO2             |  | Weight           |  | BMI              |      |
|  |                  |  |                  |  |                  |      |
|  |     98           |  |    82.0          |  |    32.1          |      |
|  |      %           |  |     kg           |  |   kg/m2          |      |
|  |                  |  |                  |  |                  |      |
|  |  [normal]        |  |  [stable]        |  |  [!HIGH-Obese]   |      |
|  |  ___________     |  |  ___/--\___      |  |  ___/--\___      |      |
|  |  (6mo trend)     |  |  (6mo trend)     |  |  (6mo trend)     |      |
|  +------------------+  +------------------+  +------------------+      |
|                                                                        |
|  Resp Rate: 16 /min [normal]    Pain: 3/10    Height: 160 cm           |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 2.7 `<VitalsFlowsheet>` — ICU-Style Time-Series Grid

```
+------------------------------------------------------------------------+
|  Vitals Flowsheet — ICU Bed 4                      [1h] [4h] [8h] 24h |
+------------------------------------------------------------------------+
|  Parameter  | 06:00 | 08:00 | 10:00 | 12:00 | 14:00 | 16:00 | 18:00  |
|-------------|-------|-------|-------|-------|-------|-------|---------|
|  Systolic   |  132  |  128  |  135  | *142* |  138  |  130  |  126   |
|  Diastolic  |   84  |   82  |   86  | * 92* |   88  |   84  |   80   |
|  MAP        |  100  |   97  |  102  | *109* |  105  |   99  |   95   |
|  HR         |   78  |   75  |   80  |   82  |   76  |   74  |   72   |
|  SpO2       |   97  |   98  |   97  |   96  |   98  |   98  |   99   |
|  Temp (F)   | 98.8  | 98.6  | 99.1  |*100.2*| 99.8  | 99.2  |  98.8  |
|  RR         |   16  |   14  |   16  |   18  |   16  |   15  |   14   |
|  FiO2       |  21%  |  21%  |  30%  |  30%  |  21%  |  21%  |  21%   |
|  Urine (mL) |  150  |  120  |  180  |  100  |  160  |  140  |  130   |
+------------------------------------------------------------------------+
|  * = abnormal value (highlighted red)                                  |
+------------------------------------------------------------------------+
```

---

### 2.8 `<LabResults>` — Laboratory Results Display

```
Props:
  results: Observation[] | DiagnosticReport | string
  groupByPanel?: boolean           # Group into CBC, BMP, etc.
  showTrends?: boolean
  showReferenceRange?: boolean
  highlightAbnormal?: boolean
  timeRange?: { start: Date, end: Date }
```

**ASCII Component:**

```
+------------------------------------------------------------------------+
|  Laboratory Results                              [Date: Jan 15, 2025]  |
+------------------------------------------------------------------------+
|                                                                        |
|  Basic Metabolic Panel (BMP) — Jan 15, 2025                            |
|  +------------------------------------------------------------------+  |
|  | Test           | Result  | Flag | Ref Range    | Units  | Trend  |  |
|  |----------------|---------|------|--------------|--------|--------|  |
|  | Glucose        | *145*   |  H   | 70-100       | mg/dL  | /^^^   |  |
|  | BUN            |   18    |      | 7-20         | mg/dL  | ---    |  |
|  | Creatinine     |  1.1    |      | 0.7-1.3      | mg/dL  | ---    |  |
|  | Sodium         |  140    |      | 136-145      | mEq/L  | ---    |  |
|  | Potassium      |  4.2    |      | 3.5-5.0      | mEq/L  | ---    |  |
|  | Chloride       |  102    |      | 98-106       | mEq/L  | ---    |  |
|  | CO2            |   24    |      | 22-29        | mEq/L  | ---    |  |
|  | Calcium        |  9.5    |      | 8.5-10.5     | mg/dL  | ---    |  |
|  | eGFR           |   78    |      | >60          | mL/min | \__    |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  HbA1c — Jan 15, 2025                                                  |
|  +------------------------------------------------------------------+  |
|  | Hemoglobin A1c | *7.2*   |  H   | <5.7         | %      | /^^    |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  Lipid Panel — Jan 15, 2025                                            |
|  +------------------------------------------------------------------+  |
|  | Total Chol     | *210*   |  H   | <200         | mg/dL  | ---    |  |
|  | LDL            | *130*   |  H   | <100         | mg/dL  | /^     |  |
|  | HDL            |   45    |      | >40          | mg/dL  | ---    |  |
|  | Triglycerides  |  175    |      | <150         | mg/dL  | /^     |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  * = abnormal (bold red)     Trend: /^^ rising  \__ falling  --- flat  |
+------------------------------------------------------------------------+
```

---

### 2.9 `<LabSparkline>` — Mini Trend Chart

```
+-------------------------------------------+
|  HbA1c Trend (12 months)                  |
|                                           |
|  8.0 |          *                         |
|  7.5 |    *          *                    |
|  7.0 |         .  .  .  .  *    *        |
|  6.5 | *                                  |
|  6.0 |-----------------------------------  |
|  5.7 |- - - - - - - - - (target) - - - -  |
|      | J  F  M  A  M  J  J  A  S  O  N  D |
|                                           |
|  Latest: 7.2%  Goal: <7.0%  Trend: UP    |
+-------------------------------------------+
```

---

### 2.10 `<ImmunizationRecord>` — Vaccine History

```
+------------------------------------------------------------------------+
|  Immunizations                                     [+ Record Vaccine]  |
+------------------------------------------------------------------------+
|  Status | Vaccine                | Date       | Dose  | Site     | Lot |
|---------|------------------------|------------|-------|----------|-----|
|  [done] | COVID-19 Bivalent      | 2024-10-15 | Boost | L deltoid| ... |
|  [done] | Influenza 2024-25      | 2024-09-20 | Annual| R deltoid| ... |
|  [done] | Tdap                   | 2020-03-10 | Boost | L deltoid| ... |
|  [done] | Hepatitis B            | 2019-01-05 | #3    | R deltoid| ... |
|---------|------------------------|------------|-------|----------|-----|
|  [DUE]  | Pneumococcal (PCV20)   | --         | #1    | --       | --  |
|  [DUE]  | Shingrix (RZV)         | --         | #1    | --       | --  |
|  [late] | Influenza 2025-26      | overdue    | Annual| --       | --  |
+------------------------------------------------------------------------+
```

---

### 2.11 `<ClinicalTimeline>` — Patient Event Timeline

```
Props:
  patientId: string
  resourceTypes?: string[]         # Filter to specific types
  timeRange?: { start: Date, end: Date }
  groupBy?: 'day' | 'encounter' | 'type'
  searchable?: boolean
```

**ASCII Component:**

```
+------------------------------------------------------------------------+
|  Clinical Timeline                                                     |
|  [All] [Encounters] [Labs] [Meds] [Procedures] [Notes]    [Search...] |
+------------------------------------------------------------------------+
|                                                                        |
|  TODAY — Jan 15, 2025                                                  |
|  |                                                                     |
|  *--[Encounter] Office Visit — Dr. Johnson                             |
|  |   Follow-up for HTN and DM management                              |
|  |   Diagnoses: Essential HTN, Type 2 DM                              |
|  |                                                                     |
|  *--[Lab] Basic Metabolic Panel                                        |
|  |   Glucose: 145 [H]  |  Creatinine: 1.1  |  eGFR: 78              |
|  |                                                                     |
|  *--[Lab] HbA1c: 7.2% [H]                                             |
|  |                                                                     |
|  *--[Medication] Metformin increased 500mg -> 1000mg BID               |
|  |                                                                     |
|  Jan 10, 2025                                                          |
|  |                                                                     |
|  *--[Note] Phone call — Patient reported increased thirst              |
|  |                                                                     |
|  Dec 20, 2024                                                          |
|  |                                                                     |
|  *--[Immunization] Influenza 2024-25 vaccine administered              |
|  |                                                                     |
|  Dec 15, 2024                                                          |
|  |                                                                     |
|  *--[Encounter] Urgent Care Visit — Dr. Patel                          |
|  |   Diagnosis: Acute bronchitis                                       |
|  |   Rx: Amoxicillin 500mg TID x 10 days                             |
|  |                                                                     |
|  Nov 28, 2024                                                          |
|  |                                                                     |
|  *--[Procedure] Colonoscopy — Dr. Lee                                  |
|  |   Finding: Normal, no polyps                                       |
|  |   Next screening: 10 years                                         |
|  |                                                                     |
|  [Load more...]                                                        |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 2.12 `<EncounterSummary>` — Visit Details

```
+------------------------------------------------------------------------+
|  Encounter: Office Visit                           Jan 15, 2025        |
+------------------------------------------------------------------------+
|                                                                        |
|  Provider:  Dr. Sarah Johnson, MD — Internal Medicine                  |
|  Location:  Springfield Medical Center, Room 204                       |
|  Duration:  2:00 PM - 2:45 PM (45 min)                                |
|  Status:    [finished]                                                 |
|                                                                        |
|  Reason for Visit:                                                     |
|  Follow-up for hypertension and type 2 diabetes management             |
|                                                                        |
|  Diagnoses:                                                            |
|  1. Essential hypertension (I10) — ongoing                             |
|  2. Type 2 diabetes mellitus (E11.9) — ongoing                        |
|                                                                        |
|  Orders:                                                               |
|  [Rx] Metformin 1000mg PO BID (increased from 500mg)                  |
|  [Lab] HbA1c in 3 months                                              |
|  [Ref] Ophthalmology — diabetic eye exam                              |
|                                                                        |
|  Follow-up: 3 months                                                   |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 2.13 `<CareTeamCard>`

```
+----------------------------------------+
|  Care Team                             |
+----------------------------------------+
|  (SJ) Dr. Sarah Johnson               |
|       Primary Care Physician           |
|       Internal Medicine                |
|       [phone] [message]                |
|                                        |
|  (MK) Dr. Michael Kim                  |
|       Specialist                       |
|       Endocrinology                    |
|       [phone] [message]                |
|                                        |
|  (RL) Rachel Lee, RN                   |
|       Care Coordinator                 |
|       [phone] [message]                |
|                                        |
|  (TP) Tom Park, PharmD                 |
|       Clinical Pharmacist              |
|       [phone] [message]                |
+----------------------------------------+
```

---

### 2.14 `<CarePlanView>` — Goals & Activities

```
+------------------------------------------------------------------------+
|  Care Plan: Diabetes Management                    Status: [active]    |
+------------------------------------------------------------------------+
|                                                                        |
|  Goals:                                                                |
|  +------------------------------------------------------------------+  |
|  | [=====>          ] 45%   HbA1c < 7.0%                            |  |
|  |   Current: 7.2%    Target: <7.0%    Due: Jun 2025               |  |
|  |                                                                  |  |
|  | [========>       ] 60%   Weight loss 5%                          |  |
|  |   Current: 82kg    Target: 78kg     Due: Jun 2025               |  |
|  |                                                                  |  |
|  | [==>             ] 20%   BP < 130/80                             |  |
|  |   Current: 138/88  Target: <130/80  Due: Mar 2025               |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  Activities:                                                           |
|  [x] Daily blood glucose monitoring                                    |
|  [x] Metformin 1000mg BID                                             |
|  [ ] 30 min walking 5x/week                                           |
|  [ ] Dietitian consultation (scheduled Feb 1)                          |
|  [ ] Diabetic eye exam (referral sent)                                 |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 2.15 `<AppointmentCard>`

```
+----------------------------------------+
|  Upcoming Appointment                  |
+----------------------------------------+
|  Wed, Jan 22, 2025                     |
|  10:00 AM - 10:30 AM                   |
|                                        |
|  Follow-up Visit                       |
|  Dr. Sarah Johnson                     |
|  Springfield Medical Center            |
|  Room 204                              |
|                                        |
|  [Check In]  [Reschedule]  [Cancel]    |
+----------------------------------------+
```

---

### 2.16 `<DocumentViewer>` — Clinical Document Viewer

```
+------------------------------------------------------------------------+
|  Document: Discharge Summary                       Jan 10, 2025        |
+------------------------------------------------------------------------+
|  Type: Discharge Summary                                               |
|  Author: Dr. Sarah Johnson                                            |
|  Status: [final]                                                       |
|                                                                        |
|  +------------------------------------------------------------------+  |
|  |                                                                  |  |
|  |                    [Document Preview Area]                       |  |
|  |                                                                  |  |
|  |  Renders: PDF | CDA/C-CDA | Images | FHIR Narrative            |  |
|  |                                                                  |  |
|  |  Supports zoom, scroll, multi-page navigation                   |  |
|  |                                                                  |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  [<< Prev] Page 1 of 3 [Next >>]    [Download] [Print] [Share]       |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 2.17 `<ClaimSummary>` — Insurance Claim Status

```
+------------------------------------------------------------------------+
|  Claim #CLM-2025-00142                             Status: [processed] |
+------------------------------------------------------------------------+
|                                                                        |
|  Service Date:  Jan 15, 2025                                           |
|  Provider:      Dr. Sarah Johnson — Springfield Medical                |
|  Payer:         Blue Cross Blue Shield                                 |
|  Patient:       John Smith  (MRN: 12345678)                            |
|                                                                        |
|  +------------------------------------------------------------------+  |
|  | Service                      | Billed  | Allowed | You Owe       |  |
|  |------------------------------|---------|---------|---------------|  |
|  | Office visit (99214)         | $200.00 | $150.00 | $30.00 copay  |  |
|  | HbA1c (83036)                |  $45.00 |  $35.00 | $0.00         |  |
|  | BMP (80048)                  |  $65.00 |  $50.00 | $0.00         |  |
|  | Lipid panel (80061)          |  $55.00 |  $40.00 | $0.00         |  |
|  |------------------------------|---------|---------|---------------|  |
|  | TOTAL                        | $365.00 | $275.00 | $30.00        |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  Insurance paid: $245.00    Your responsibility: $30.00                |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 2.18 `<TaskCard>` — Clinical Task

```
+----------------------------------------+
|  [!! HIGH]  Task                       |
+----------------------------------------+
|  Review abnormal lab results           |
|                                        |
|  Patient: John Smith                   |
|  Assigned: Dr. Sarah Johnson           |
|  Due: Jan 16, 2025                     |
|  Status: [requested]                   |
|                                        |
|  Context: HbA1c 7.2%, Glucose 145     |
|                                        |
|  [Mark Complete]  [Reassign]  [Defer]  |
+----------------------------------------+
```

---

### 2.19 `<ConsentBanner>` — Active Consent Display

```
+------------------------------------------------------------------------+
|  [shield] Active Consents for John Smith                               |
+------------------------------------------------------------------------+
|  [active]  General Treatment Consent           Signed: Jan 1, 2025     |
|  [active]  HIPAA Privacy Notice                Signed: Jan 1, 2025     |
|  [active]  Research Participation (Study #42)  Signed: Nov 15, 2024    |
|  [denied]  Telehealth Consent                  Denied: Dec 20, 2024    |
|                                                                        |
|  [Request New Consent]  [View Details]                                 |
+------------------------------------------------------------------------+
```

---

## Layer 3: Clinical Workflow Components

> ~25 components that compose multiple FHIR resources into real clinical workflows.

---

### 3.1 `<TerminologySearch>` — Code Search Autocomplete

The single most-rebuilt component in healthcare. Searches ValueSet/$expand.

```
Props:
  valueSetUrl?: string             # e.g., "http://hl7.org/fhir/ValueSet/condition-code"
  system?: string                  # e.g., "http://snomed.info/sct"
  onChange: (coding: Coding) => void
  multiple?: boolean               # Allow multi-select
  placeholder?: string
  recentCodes?: Coding[]           # Show recent selections
  favoriteCodes?: Coding[]         # Show pinned favorites
```

**ASCII Component:**

```
+------------------------------------------------------------------------+
|  Search Diagnosis (SNOMED-CT + ICD-10)                                 |
|                                                                        |
|  +------------------------------------------------------------------+  |
|  | essential hyper_                                            [X]  |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  Recent:                                                               |
|  [Essential hypertension] [Type 2 diabetes] [Obesity]                  |
|                                                                        |
|  Results:                                                              |
|  +------------------------------------------------------------------+  |
|  |  [SNOMED] Essential hypertension (disorder)                      |  |
|  |           59621000  |  Also: ICD-10 I10                          |  |
|  |  ----------------------------------------------------------------|  |
|  |  [SNOMED] Essential hypertension in pregnancy                    |  |
|  |           48194001                                               |  |
|  |  ----------------------------------------------------------------|  |
|  |  [SNOMED] Hypertensive disorder, systemic arterial               |  |
|  |           38341003                                               |  |
|  |  ----------------------------------------------------------------|  |
|  |  [SNOMED] Renovascular hypertension                              |  |
|  |           123799005                                              |  |
|  |  ----------------------------------------------------------------|  |
|  |  [ICD-10] I10 — Essential (primary) hypertension                 |  |
|  |  ----------------------------------------------------------------|  |
|  |  [ICD-10] I11.9 — Hypertensive heart disease without HF         |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  Showing 6 of 23 results  [Load more]                                  |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.2 `<MedicationReconciliation>` — Med Reconciliation Workflow

```
+------------------------------------------------------------------------+
|  Medication Reconciliation — John Smith                                |
|  Comparing: [Admission List] vs [Home Medications]                     |
+------------------------------------------------------------------------+
|                                                                        |
|  Home Medications            |  Admission Orders        | Action      |
|  ----------------------------|--------------------------|-------------|
|  Lisinopril 10mg daily       |  Lisinopril 10mg daily   | [Continue]  |
|  Metformin 500mg BID         |  Metformin 1000mg BID    | [! Changed] |
|  Atorvastatin 20mg QHS       |  Atorvastatin 20mg QHS   | [Continue]  |
|  Aspirin 81mg daily          |  Aspirin 81mg daily      | [Continue]  |
|  Ibuprofen 400mg PRN         |  --                      | [Hold]      |
|  --                          |  Heparin 5000u SQ Q8H    | [+ New]     |
|  --                          |  Omeprazole 20mg daily   | [+ New]     |
|                                                                        |
|  Discrepancies: 3                                                      |
|  [!] Metformin dose changed (500mg -> 1000mg) — verify with patient   |
|  [!] Ibuprofen held — renal risk with current Cr                      |
|  [+] Heparin added — DVT prophylaxis                                  |
|                                                                        |
|  [Save Reconciliation]  [Mark All Reviewed]  [Print]                   |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.3 `<DrugInteractionAlert>` — Drug-Drug / Drug-Allergy

```
+------------------------------------------------------------------------+
|  [!!! CRITICAL]  Drug Interaction Alert                                |
+------------------------------------------------------------------------+
|                                                                        |
|  Amoxicillin 500mg <-> Penicillin Allergy                              |
|  +------------------------------------------------------------------+  |
|  |  [!!!] Cross-reactivity risk                                     |  |
|  |                                                                  |  |
|  |  Amoxicillin is a penicillin-class antibiotic. Patient has       |  |
|  |  documented SEVERE allergy to Penicillin with history of         |  |
|  |  ANAPHYLAXIS (2005).                                             |  |
|  |                                                                  |  |
|  |  Cross-reactivity rate: ~1-2% (historically cited higher)       |  |
|  |                                                                  |  |
|  |  Recommendation: Use alternative antibiotic class                |  |
|  |  Alternatives: Azithromycin, Doxycycline, Fluoroquinolone       |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  [Cancel Order]  [Override with Reason v]                              |
|                                                                        |
|  Override reasons:                                                     |
|  ( ) Patient tolerates — previously administered without reaction      |
|  ( ) Benefits outweigh risks — no alternative available                |
|  ( ) Allergy verified inaccurate — updated allergy record             |
|  ( ) Other: ____________                                               |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.4 `<SOAPNote>` — Clinical Documentation Editor

```
Props:
  encounterId?: string
  patientId: string
  template?: 'soap' | 'progress' | 'h&p' | 'procedure' | 'custom'
  autoPopulate?: boolean          # Pre-fill from encounter data
  onSave: (composition: Composition) => void
```

**ASCII Component:**

```
+------------------------------------------------------------------------+
|  Clinical Note — Office Visit                      [Template: SOAP v]  |
|  Patient: John Smith  |  Provider: Dr. Johnson  |  Jan 15, 2025       |
+------------------------------------------------------------------------+
|                                                                        |
|  SUBJECTIVE                                                [expand/v]  |
|  +------------------------------------------------------------------+  |
|  | Chief Complaint:                                                 |  |
|  | Follow-up for hypertension and diabetes management               |  |
|  |                                                                  |  |
|  | HPI:                                                             |  |
|  | 47-year-old male presents for routine follow-up. Reports         |  |
|  | increased thirst and urination over past 2 weeks. Denies         |  |
|  | chest pain, shortness of breath, visual changes. Home BP         |  |
|  | readings averaging 135-140/85-90. Compliant with medications.    |  |
|  |                                                                  |  |
|  | ROS: [+ Polyuria] [+ Polydipsia] [- Chest pain] [- SOB]        |  |
|  | [.symptoms]  <-- SmartPhrase expansion                          |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  OBJECTIVE                                                 [expand/v]  |
|  +------------------------------------------------------------------+  |
|  | Vitals: (auto-populated from today's vitals)                     |  |
|  | BP 138/88  HR 72  Temp 98.6  SpO2 98%  Wt 82kg  BMI 32.1      |  |
|  |                                                                  |  |
|  | Physical Exam:                                                   |  |
|  | General: NAD, well-appearing                                     |  |
|  | HEENT: PERRLA, EOMI, no JVD                                     |  |
|  | Cardiac: RRR, no murmurs                                        |  |
|  | Lungs: CTAB                                                      |  |
|  | Extremities: No edema, pulses 2+ bilaterally                    |  |
|  | Neuro: A&O x3                                                    |  |
|  | [.exam-normal]  <-- SmartPhrase expansion                       |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  ASSESSMENT                                                [expand/v]  |
|  +------------------------------------------------------------------+  |
|  | 1. Essential hypertension (I10) — suboptimally controlled        |  |
|  |    [Search diagnosis...]                                        |  |
|  | 2. Type 2 diabetes (E11.9) — worsening, HbA1c 7.2%             |  |
|  | 3. Obesity (E66.01) — stable, BMI 32.1                          |  |
|  | [+ Add diagnosis]                                                |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  PLAN                                                      [expand/v]  |
|  +------------------------------------------------------------------+  |
|  | 1. Increase Metformin 500mg -> 1000mg BID            [+ Order]  |  |
|  | 2. Continue Lisinopril 10mg daily                               |  |
|  | 3. Repeat HbA1c in 3 months                          [+ Order]  |  |
|  | 4. Refer to Ophthalmology for diabetic eye exam      [+ Refer]  |  |
|  | 5. Dietitian referral for medical nutrition therapy  [+ Refer]  |  |
|  | 6. Follow-up in 3 months                             [+ Appt]   |  |
|  | [+ Add plan item]                                                |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  [Save Draft]  [Sign & Lock]  [Co-sign Required: Dr. Kim]             |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.5 `<QuestionnaireForm>` — FHIR SDC Form Renderer

```
Props:
  questionnaire: Questionnaire | string    # Resource or canonical URL
  subject?: Reference                      # Patient reference
  prePopulate?: boolean                    # Use $populate
  onSubmit: (response: QuestionnaireResponse) => void
  readOnly?: boolean
```

**ASCII Component:**

```
+------------------------------------------------------------------------+
|  PHQ-9 Depression Screening                                            |
|  Over the last 2 weeks, how often have you been bothered by:           |
+------------------------------------------------------------------------+
|                                                                        |
|  1. Little interest or pleasure in doing things?                       |
|     ( ) Not at all  (x) Several days  ( ) More than half  ( ) Nearly  |
|                                                         every day      |
|                                                                        |
|  2. Feeling down, depressed, or hopeless?                              |
|     (x) Not at all  ( ) Several days  ( ) More than half  ( ) Nearly  |
|                                                         every day      |
|                                                                        |
|  3. Trouble falling or staying asleep, or sleeping too much?           |
|     ( ) Not at all  ( ) Several days  (x) More than half  ( ) Nearly  |
|                                                         every day      |
|                                                                        |
|  ... (questions 4-9 continue)                                          |
|                                                                        |
|  +------------------------------------------------------------------+  |
|  |  Score: 8 / 27                                                   |  |
|  |  Interpretation: [Mild Depression]                                |  |
|  |  Recommendation: Monitor, consider follow-up in 2-4 weeks       |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  If you answered "Several days" or more to question 9 (self-harm):    |
|  +------------------------------------------------------------------+  |
|  |  [!! SAFETY ALERT] This section is conditionally displayed       |  |
|  |  How difficult have these problems made it for you?              |  |
|  |  ( ) Not difficult  ( ) Somewhat  ( ) Very  ( ) Extremely        |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  [Save Draft]  [Submit Response]                                       |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.6 `<CDSHooksCard>` — Decision Support Card

```
Props:
  card: CDSCard                    # CDS Hooks card response
  onAccept?: (suggestion) => void
  onDismiss?: (reason: string) => void
  onLaunchSMART?: (url: string) => void
```

**ASCII Component:**

```
+------------------------------------------------------------------------+
|  Clinical Decision Support                                             |
+------------------------------------------------------------------------+
|                                                                        |
|  [!!! CRITICAL]                                                        |
|  +------------------------------------------------------------------+  |
|  |  Opioid Risk Assessment Required                                 |  |
|  |                                                                  |  |
|  |  Patient has active opioid prescription (Oxycodone 5mg).        |  |
|  |  CDC Guideline recommends PDMP check and risk assessment.       |  |
|  |                                                                  |  |
|  |  Source: CDC Opioid Prescribing Guidelines                       |  |
|  |                                                                  |  |
|  |  [Check PDMP]  [Order UDS]  [Open Risk Tool ->]  [Dismiss v]    |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  [!! WARNING]                                                          |
|  +------------------------------------------------------------------+  |
|  |  Diabetes: Annual Eye Exam Overdue                               |  |
|  |                                                                  |  |
|  |  Patient has Type 2 DM (diagnosed 2020). No ophthalmology       |  |
|  |  visit in past 14 months.                                        |  |
|  |                                                                  |  |
|  |  [Create Referral]  [Already Scheduled]  [Dismiss v]             |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  [i INFO]                                                              |
|  +------------------------------------------------------------------+  |
|  |  Pneumococcal Vaccine Recommended                                |  |
|  |                                                                  |  |
|  |  Patient age 47 with diabetes. PCV20 recommended per            |  |
|  |  ACIP guidelines.                                                |  |
|  |                                                                  |  |
|  |  [Order Vaccine]  [Defer 3 months]  [Dismiss v]                 |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.7 `<Scheduler>` — Appointment Booking

```
+------------------------------------------------------------------------+
|  Schedule Appointment                                                  |
+------------------------------------------------------------------------+
|                                                                        |
|  Provider: [Dr. Sarah Johnson      v]                                  |
|  Type:     [Follow-up Visit        v]                                  |
|  Duration: [30 min                 v]                                  |
|  Location: [Springfield Medical    v]                                  |
|                                                                        |
|  January 2025                                                          |
|  +------------------------------------------------------------------+  |
|  | Mon    | Tue    | Wed    | Thu    | Fri    |                      |  |
|  |--------|--------|--------|--------|--------|                      |  |
|  |   20   |   21   | * 22 * |   23   |   24   |                      |  |
|  | --     | 9:00   | 9:00   | --     | 10:00  |                      |  |
|  | --     | 9:30   | 9:30   | --     | 10:30  |                      |  |
|  | --     | 10:00  | 10:00  | --     | 11:00  |                      |  |
|  | --     | 10:30  |[10:30] | --     | --     |                      |  |
|  | --     | 14:00  | 14:00  | --     | 14:00  |                      |  |
|  | --     | 14:30  | 14:30  | --     | 14:30  |                      |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  Selected: Wednesday, Jan 22 at 10:30 AM                               |
|  Provider: Dr. Sarah Johnson                                           |
|  Type:     30-min Follow-up Visit                                      |
|                                                                        |
|  Notes: ____________________________________________                   |
|                                                                        |
|  [Confirm Booking]  [Cancel]                                           |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.8 `<PatientSearch>` — Patient Lookup

```
+------------------------------------------------------------------------+
|  Patient Search                                                        |
|                                                                        |
|  +------------------------------------------------------------------+  |
|  | John Smith                                                  [X]  |  |
|  +------------------------------------------------------------------+  |
|  | [Name] [MRN] [DOB] [Phone] [SSN]   <-- search mode tabs         |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  +------------------------------------------------------------------+  |
|  | (JS) John A. Smith         DOB: 03/15/1978  MRN: 12345678       |  |
|  |      47M  Springfield, IL  Phone: (555) 123-4567                |  |
|  |--------------------------------------------------------------|  |
|  | (JS) John B. Smith         DOB: 11/22/1985  MRN: 87654321       |  |
|  |      39M  Chicago, IL      Phone: (312) 555-0199                |  |
|  |--------------------------------------------------------------|  |
|  | (JS) Jonathan Smith         DOB: 06/01/1990  MRN: 11223344       |  |
|  |      34M  Decatur, IL      Phone: (217) 555-0142                |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  3 results found  [+ Register New Patient]                             |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.9 `<ResourceSearch>` — Generic FHIR Search

```
+------------------------------------------------------------------------+
|  Search: [Observation        v]                      [+ Add Filter]    |
+------------------------------------------------------------------------+
|  Filters:                                                              |
|  [patient = John Smith x] [category = laboratory x] [date >= 2024 x]  |
+------------------------------------------------------------------------+
|                                                                        |
|  | Resource         | Patient      | Date       | Category | Value    |
|  |------------------|--------------|------------|----------|----------|
|  | Glucose          | John Smith   | 2025-01-15 | lab      | 145 mg/dL|
|  | HbA1c            | John Smith   | 2025-01-15 | lab      | 7.2 %   |
|  | Creatinine       | John Smith   | 2025-01-15 | lab      | 1.1 mg/dL|
|  | BUN              | John Smith   | 2025-01-15 | lab      | 18 mg/dL |
|  | Total Cholesterol| John Smith   | 2025-01-15 | lab      | 210 mg/dL|
|                                                                        |
|  Showing 1-5 of 142  [< Prev] [1] [2] [3] ... [29] [Next >]          |
|                                                                        |
|  [Export CSV]  [Export NDJSON]  [Bulk Export $export]                   |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.10 `<ResourceForm>` — Auto-Generated FHIR Resource Form

```
+------------------------------------------------------------------------+
|  Create: Patient                                   [JSON] [Form]       |
+------------------------------------------------------------------------+
|                                                                        |
|  Name                                               [+ Add Name]      |
|  +------------------------------------------------------------------+  |
|  | Use: [Official v]                                                |  |
|  | Family: [____________]  Given: [____________]                    |  |
|  | Prefix: [____]  Suffix: [____]                                   |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  Gender:   (x) Male  ( ) Female  ( ) Other  ( ) Unknown               |
|  Birth Date: [____/____/________]                                      |
|                                                                        |
|  Telecom                                            [+ Add Contact]   |
|  +------------------------------------------------------------------+  |
|  | System: [Phone v]  Value: [____________]  Use: [Mobile v]        |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  Address                                            [+ Add Address]   |
|  +------------------------------------------------------------------+  |
|  | [AddressInput — see Layer 1]                                     |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  Identifier                                                            |
|  +------------------------------------------------------------------+  |
|  | System: [MRN       v]  Value: [auto-generated]                   |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  [Create Patient]  [Cancel]                                            |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.11 `<AuditLog>` — HIPAA Audit Trail Viewer

```
+------------------------------------------------------------------------+
|  Audit Log                                         [Export]  [Filter]  |
+------------------------------------------------------------------------+
|  Patient: John Smith (MRN: 12345678)               Last 30 days       |
+------------------------------------------------------------------------+
|                                                                        |
|  Timestamp          | User              | Action  | Resource          |
|  --------------------|-------------------|---------|-------------------|
|  2025-01-15 14:32:05| Dr. Sarah Johnson | READ    | Patient/123       |
|  2025-01-15 14:32:08| Dr. Sarah Johnson | READ    | Condition?patient |
|  2025-01-15 14:32:10| Dr. Sarah Johnson | READ    | MedicationRequest |
|  2025-01-15 14:33:15| Dr. Sarah Johnson | READ    | Observation/lab-1 |
|  2025-01-15 14:35:22| Dr. Sarah Johnson | UPDATE  | MedicationRequest |
|  2025-01-15 14:36:01| System (CDS)      | READ    | AllergyIntolerance|
|  2025-01-15 14:40:00| Dr. Sarah Johnson | CREATE  | Encounter/enc-55  |
|  2025-01-15 14:42:30| Lab Interface      | CREATE  | Observation/lab-2 |
|                                                                        |
|  Showing 1-8 of 234  [< Prev] [Next >]                                |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.12 `<BulkExportStatus>` — $export Job Dashboard

```
+------------------------------------------------------------------------+
|  Bulk Data Export                                   [+ New Export]      |
+------------------------------------------------------------------------+
|                                                                        |
|  Active Jobs:                                                          |
|  +------------------------------------------------------------------+  |
|  | Job #1  System Export                   Started: 14:00            |  |
|  | [=========>                   ] 35%     ETA: ~8 min              |  |
|  | 5/14 resource types processed                                    |  |
|  | Patient, Observation, Condition, MedicationRequest, Encounter    |  |
|  | [Cancel]                                                         |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
|  Completed:                                                            |
|  +------------------------------------------------------------------+  |
|  | Job #0  Patient/$export (ID: abc-123)  Completed: 13:45          |  |
|  | 14 files, 2.3 MB total, 1,247 resources                         |  |
|  | [Download All]  [Individual Files v]  [Delete]                   |  |
|  |                                                                  |  |
|  | Files:                                                           |  |
|  | Patient.ndjson          12 KB    45 resources    [download]      |  |
|  | Observation.ndjson     890 KB   423 resources    [download]      |  |
|  | Condition.ndjson        56 KB    89 resources    [download]      |  |
|  | MedicationRequest.ndjson 34 KB   67 resources    [download]     |  |
|  | ... 10 more files                                                |  |
|  +------------------------------------------------------------------+  |
|                                                                        |
+------------------------------------------------------------------------+
```

---

### 3.13-3.17 Additional Workflow Components

**`<OrderEntry>`** — Create medication/lab/imaging orders with terminology binding

**`<ReferralWorkflow>`** — Create and track specialist referrals

**`<DischargeChecklist>`** — Discharge planning with med reconciliation, follow-up scheduling

**`<HandoffSummary>`** — I-PASS clinical handoff format

**`<InsuranceVerification>`** — Coverage eligibility check and display

---

## Layer 4: Layout & Infrastructure

---

### 4.1 `<FHIRProvider>` — Root Provider

```tsx
import { FHIRProvider } from '@ehr/fhir-hooks'

<FHIRProvider
  serverUrl="https://your-ehr.com/fhir"       // Any FHIR R4 server
  auth={{
    type: 'bearer',                             // or 'smart', 'basic', 'none'
    token: 'eyJ...',
  }}
  defaultHeaders={{ 'X-Tenant': 'clinic-1' }}  // Multi-tenant support
  onError={(error) => console.error(error)}
>
  {children}
</FHIRProvider>
```

---

### 4.2 `<EHRShell>` — Application Shell

```
+------------------------------------------------------------------------+
|  [logo] EHR System          [search...]        [alerts(3)] [user v]   |
+------------------------------------------------------------------------+
|         |                                                              |
| Sidebar |  Main Content Area                                          |
|         |                                                              |
| [home]  |  +--------------------------------------------------------+  |
| Patients|  |                                                        |  |
| Schedule|  |    [PatientBanner]                                     |  |
| Inbox(5)|  |                                                        |  |
| Tasks(3)|  |    [Tab: Summary] [Tab: Chart] [Tab: Orders] [Tab: .] |  |
| Reports |  |                                                        |  |
| Admin   |  |    [Active content based on route]                     |  |
|         |  |                                                        |  |
|         |  |                                                        |  |
|         |  +--------------------------------------------------------+  |
|         |                                                              |
+------------------------------------------------------------------------+
|  [status bar]  Connected to: your-ehr.com  |  User: Dr. Johnson       |
+------------------------------------------------------------------------+
```

---

### 4.3 `<PatientContext>` — Patient State Provider

```tsx
<PatientContext patientId="123">
  {/* All children automatically have access to patient data */}
  <PatientBanner />        {/* No need to pass patientId */}
  <ProblemList />          {/* Automatically fetches for this patient */}
  <MedicationList />       {/* Automatically fetches for this patient */}
</PatientContext>
```

---

### 4.4 `<RBACGate>` — Role-Based Access Control

```tsx
<RBACGate requiredRole="physician" fallback={<AccessDenied />}>
  <SOAPNote />  {/* Only visible to physicians */}
</RBACGate>

<RBACGate requiredPermission="prescribe" fallback={<ReadOnlyView />}>
  <OrderEntry />  {/* Only visible to users who can prescribe */}
</RBACGate>
```

---

### 4.5 `<PHIGuard>` — Audit-Logged PHI Access

```tsx
<PHIGuard
  resource="Patient/123"
  action="read"
  reason="treatment"
  onAccess={(event) => auditLog.record(event)}
>
  <PatientSummary patientId="123" />
</PHIGuard>
```

---

### 4.6 `<SMARTLaunch>` — SMART on FHIR App Launch

```tsx
import { SMARTLaunch, useSMARTAuth } from '@ehr/smart-auth'

// Handles EHR launch and standalone launch flows
<SMARTLaunch
  clientId="my-app"
  scope="launch patient/*.read"
  redirectUri="/callback"
  iss="https://ehr.example.com/fhir"
>
  <App />
</SMARTLaunch>
```

---

## React Hooks Reference

### Data Fetching Hooks

```tsx
// Fetch a single resource
const { data, loading, error } = useResource<Patient>('Patient', '123')

// FHIR search with params
const { bundle, loading, refetch } = useSearch('Observation', {
  patient: '123',
  category: 'laboratory',
  _sort: '-date',
  _count: 20,
})

// Patient with $everything
const { data, loading } = usePatient('123', { everything: true })

// Resolve a FHIR Reference
const { resource, loading } = useReference({ reference: 'Practitioner/456' })

// Resource version history
const { versions, loading } = useResourceHistory('Patient', '123')

// Server capabilities
const { statement, loading } = useCapabilityStatement()

// Bundle pagination
const { page, hasNext, hasPrev, next, prev } = usePagination(bundle)
```

### Terminology Hooks

```tsx
// ValueSet expansion with search
const { concepts, loading } = useTerminology(
  'http://hl7.org/fhir/ValueSet/condition-code',
  { filter: 'hypertension', count: 10 }
)

// ConceptMap translation
const { translation } = useConceptTranslate(
  'http://snomed.info/sct', '59621000',
  'http://hl7.org/fhir/sid/icd-10'
)

// Load a ValueSet for dropdowns
const { options, loading } = useValueSet(
  'http://hl7.org/fhir/ValueSet/observation-category'
)
```

### Clinical Workflow Hooks

```tsx
// CDS Hooks call
const { cards, loading } = useCDSHooks('patient-view', {
  patientId: '123',
  userId: 'Practitioner/456',
})

// FHIR Subscription (real-time)
useSubscription('Observation?patient=123', (event) => {
  console.log('New observation:', event.resource)
})

// Questionnaire state management
const { answers, setAnswer, score, submit } = useQuestionnaire(
  'http://example.com/Questionnaire/phq9'
)

// FHIRPath evaluation
const { result } = useFHIRPath(patient, "name.where(use='official').given")

// Bulk export management
const { job, progress, download } = useBulkExport({
  type: 'system',
  _type: ['Patient', 'Observation', 'Condition'],
})

// SMART auth state
const { user, patient, token, logout } = useSMARTAuth()

// RBAC
const { roles, hasRole, hasPermission } = useRBAC()

// Audit logging
const logAccess = useAuditLog()
logAccess({ action: 'read', resource: 'Patient/123', reason: 'treatment' })
```

---

## Competitive Analysis

### Component Count Comparison

```
+-------------------------------------------------------------------+
|  Library                  | Components | Backend  | Style    | TS  |
|---------------------------|------------|----------|----------|-----|
|  @ehr/react (THIS)        | ~160       | ANY      | Tailwind | Yes |
|  @medplum/react           | ~120       | Medplum  | Mantine  | Yes |
|  Terra UI (Cerner)        | ~78        | N/A      | Custom   | Yes |
|  nhsuk-react-components  | ~45        | None     | NHS CSS  | Yes |
|  @cmsgov/design-system    | ~37        | None     | Custom   | Yes |
|  fhir-react (1upHealth)   | ~35 res    | ANY      | Custom   | No  |
|  Ottehr                   | ~20        | Oystehr  | Custom   | Yes |
|  fhir-ui                  | ~5         | N/A      | MUI      | No  |
+-------------------------------------------------------------------+
```

### Feature Matrix

```
+-------------------------------------------------------------------+
| Feature                 | @ehr | Medplum | fhir-react | Terra     |
|-------------------------|------|---------|------------|-----------|
| Backend-agnostic        |  YES |  Partial|    YES     |    N/A    |
| TypeScript-first        |  YES |  YES    |    NO      |    YES    |
| Headless/unstyled option|  YES |  NO     |    NO      |    NO     |
| Pre-styled default      |  YES |  YES    |    YES     |    YES    |
| FHIR R4 type-safe       |  YES |  YES    |    Partial |    NO     |
| Terminology search      |  YES |  Limited|    NO      |    NO     |
| Clinical workflows      |  YES |  Partial|    NO      |    NO     |
| SDC Questionnaire       |  YES |  Basic  |    NO      |    NO     |
| CDS Hooks renderer      |  YES |  NO     |    NO      |    NO     |
| SMART on FHIR           |  YES |  Demo   |    NO      |    NO     |
| WCAG 2.1 AA             |  YES |  Partial|    NO      |    YES    |
| i18n / RTL              |  YES |  NO     |    NO      |    YES    |
| React 18/19             |  YES |  YES    |    NO      |  ARCHIVED |
| Tree-shakeable ESM      |  YES |  YES    |    NO      |    NO     |
| Storybook docs          |  YES |  YES    |    NO      |    YES    |
| Clinical context docs   |  YES |  NO     |    NO      |    NO     |
+-------------------------------------------------------------------+
```

### Why This Wins

1. **Only library that is BOTH backend-agnostic AND comprehensive** — Medplum has 120 components but locks you to their server. fhir-react works with any server but only has 35 display-only components.

2. **Headless-first is the modern pattern** — Radix, Headless UI, React Aria proved this model. No healthcare library does it. Teams with existing design systems can adopt immediately.

3. **Clinical workflow components are the moat** — SOAP notes, med reconciliation, terminology search, CDS cards, order entry. These take months to build from scratch. No competitor offers them as composable components.

4. **Matches YOUR API perfectly** — 40+ FHIR resource types, ValueSet/$expand, ConceptMap/$translate, CDS Hooks, SMART auth, Subscriptions, $export, $everything. The React library is the frontend that unlocks all of it.

5. **Terra UI is dead** — Oracle archived it in May 2024. The healthcare industry needs a replacement. This IS that replacement.

---

## Technical Specifications

### Build & Bundle

```
Framework:    React 18/19 (concurrent mode ready)
Language:     TypeScript 5.x (strict mode)
Styling:      Tailwind CSS 4.x (pre-styled) / CSS variables (headless)
Bundler:      tsup (ESM + CJS dual output)
Testing:      Vitest + React Testing Library + Axe (a11y)
Docs:         Storybook 8 + MDX
Linting:      ESLint + Prettier + eslint-plugin-jsx-a11y
Size target:  <50KB gzipped for core bundle (tree-shakeable)
```

### Accessibility Standards

```
Standard:     WCAG 2.1 AA (required by Section 504, May 2026 deadline)
Testing:      axe-core automated + manual screen reader testing
Features:
  - ARIA roles on all interactive elements
  - Full keyboard navigation (Tab, Arrow, Enter, Escape)
  - Screen reader announcements for dynamic content
  - High contrast mode support
  - Focus management for modals/dialogs
  - Reduced motion support
  - Minimum 4.5:1 contrast ratios
  - Visible focus indicators
```

### FHIR Compliance

```
FHIR Version:  R4 (4.0.1)
Resources:     All 150+ resource type definitions
Operations:    $everything, $export, $validate, $expand, $translate,
               $subsumes, $validate-code, $match, $apply, $document
Search:        _include, _revinclude, _has, _filter, chained params
Auth:          SMART on FHIR (EHR launch + standalone), OAuth2 + PKCE
Subscriptions: R4 + R5-style Topic-Based
Bulk Data:     Bulk Data Access IG 2.0
```

### Monorepo Structure

```
packages/
  fhir-types/
    src/
      resources/           # One file per resource type
        patient.ts
        observation.ts
        condition.ts
        ... (150+ files)
      datatypes/           # FHIR data types
        human-name.ts
        codeable-concept.ts
        ... (40+ files)
      index.ts
    package.json

  fhir-hooks/
    src/
      providers/
        fhir-provider.tsx
        patient-context.tsx
      hooks/
        use-resource.ts
        use-search.ts
        use-terminology.ts
        use-subscription.ts
        use-cds-hooks.ts
        use-smart-auth.ts
        use-fhirpath.ts
        use-bulk-export.ts
        use-rbac.ts
        use-audit-log.ts
        ... (20+ hooks)
      index.ts
    package.json

  react-core/              # Headless components
    src/
      primitives/          # Layer 1 headless
      resources/           # Layer 2 headless
      workflows/           # Layer 3 headless
      index.ts
    package.json

  react/                   # Pre-styled Tailwind
    src/
      primitives/          # Layer 1 styled
        human-name.tsx
        address.tsx
        codeable-concept.tsx
        ... (30 display + 30 input)
      resources/           # Layer 2 styled
        patient-banner.tsx
        problem-list.tsx
        medication-list.tsx
        vitals-panel.tsx
        lab-results.tsx
        clinical-timeline.tsx
        ... (40 components)
      workflows/           # Layer 3 styled
        terminology-search.tsx
        soap-note.tsx
        questionnaire-form.tsx
        cds-hooks-card.tsx
        scheduler.tsx
        ... (25 components)
      layout/              # Layer 4 styled
        ehr-shell.tsx
        rbac-gate.tsx
        phi-guard.tsx
        ... (15 components)
      index.ts
    package.json

  smart-auth/
    src/
      smart-launch.tsx
      use-smart-auth.ts
      oauth-callback.tsx
    package.json

apps/
  storybook/               # Interactive documentation
  playground/              # Live component playground
```

---

## Summary: Total Component Inventory

```
+-------------------------------------------------------------------+
| Layer                          | Display | Input | Workflow | Total|
|--------------------------------|---------|-------|----------|------|
| L1: FHIR Data Primitives      |    20   |   20  |     --   |   40 |
| L2: FHIR Resource Components  |    40   |   --  |     --   |   40 |
| L3: Clinical Workflows        |    --   |   --  |     25   |   25 |
| L4: Layout & Infrastructure   |    --   |   --  |     15   |   15 |
| React Hooks                   |    --   |   --  |     20   |   20 |
|--------------------------------|---------|-------|----------|------|
| TOTAL                          |    60   |   20  |     60   |  140 |
+-------------------------------------------------------------------+

vs. Medplum: ~120 components (Mantine-locked, Medplum-coupled)
vs. fhir-react: ~35 components (display-only, no TypeScript)
vs. Terra UI: ~78 components (ARCHIVED)
```

**This library fills every gap in the healthcare React ecosystem and becomes the
industry standard by being the ONLY library that is simultaneously:**

1. Backend-agnostic (works with any FHIR server)
2. Headless-first (bring your own design system)
3. Pre-styled (beautiful defaults with Tailwind)
4. Clinically complete (workflows, not just display)
5. Accessible (WCAG 2.1 AA, legally mandated)
6. TypeScript-first (full FHIR type safety)
7. Actively maintained (backed by a complete headless EHR)
