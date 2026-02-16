# Testing & Documentation Strategy

> TDD from line 1. Every component starts as a failing test.
> Storybook yes — but lean. Not a blocker.

---

## Storybook: Yes, But Not How You Think

### The honest answer

Storybook is heavy. Slow to start. Config rabbit holes. But for an
open-source component library aiming to be industry standard, there
is no alternative. It's where developers EXPECT to find components.

**The rule: Storybook is documentation, not development.**

```
  Development happens in tests     → vitest + testing-library
  Visual checking happens in tests → vitest + playwright screenshots
  Documentation happens in Storybook → static site for users

  We do NOT:
    - Develop components inside Storybook
    - Use Storybook as our dev server
    - Block PRs on Storybook stories
    - Write stories before tests

  We DO:
    - Write a story for every shipped component
    - Use the a11y addon to catch violations
    - Deploy Storybook as our public docs site
    - Include live FHIR data examples in stories
```

### Storybook setup: Minimal, day 1

```
Addons we use (5 only — no bloat):
  @storybook/addon-essentials    — controls, actions, viewport, docs
  @storybook/addon-a11y          — axe-core violations in real-time
  @storybook/addon-themes        — theme switcher (light/dark/clinical)
  @storybook/addon-interactions  — play functions for interaction demos
  storybook-addon-performance    — render time display

Addons we DO NOT use:
  @storybook/addon-backgrounds   — themes handle this
  @storybook/addon-links         — unnecessary complexity
  @storybook/addon-storysource   — bloat
  chromatic                       — too expensive early on, add later
```

### Story structure: One story file per component

```tsx
// patient-banner.stories.tsx

import type { Meta, StoryObj } from '@storybook/react'
import { PatientBanner } from './patient-banner'
import { mockPatient, mockPatientMinimal, mockPatientAllergies } from '@ehr/test-utils'

const meta: Meta<typeof PatientBanner> = {
  title: 'Resources/PatientBanner',
  component: PatientBanner,
  tags: ['autodocs'],
  parameters: {
    docs: {
      description: {
        component: `
## Clinical Context
The Patient Banner appears at the top of every patient-facing screen.
Clinicians glance at it to confirm patient identity and check critical
safety information (allergies, flags) before any clinical action.

## FHIR Resources
- Patient (primary)
- AllergyIntolerance (allergy badges)
- Flag (clinical flags)

## Usage
\`\`\`tsx
<PatientBanner patientId="123" />
\`\`\`
        `,
      },
    },
  },
}

export default meta
type Story = StoryObj<typeof PatientBanner>

// Default — full patient data
export const Default: Story = {
  args: { patient: mockPatient },
}

// Compact — single line
export const Compact: Story = {
  args: { patient: mockPatient, compact: true },
}

// With allergies
export const WithAllergies: Story = {
  args: { patient: mockPatientAllergies },
}

// Minimal data — handles missing fields gracefully
export const MinimalData: Story = {
  args: { patient: mockPatientMinimal },
}

// Loading state
export const Loading: Story = {
  args: { patientId: '123' },
  parameters: {
    mockData: { delay: 999999 },  // never resolves = loading forever
  },
}

// Error state
export const Error: Story = {
  args: { patientId: 'nonexistent' },
  parameters: {
    mockData: { error: true },
  },
}

// Dark mode (shown via theme switcher addon)
// High contrast (shown via theme switcher addon)
// Compact density (shown via density control)
```

### When stories get written

```
  Phase 0 (scaffolding):  Storybook configured, one Box story works
  Phase 1 (tokens):       Token showcase page (all colors, spacing, type)
  Phase 3 (primitives):   Story written WITH the component (same PR)
  Phase 4 (FHIR prims):   Story written WITH the component (same PR)
  Phase 5 (resources):    Story written WITH the component (same PR)
  Phase 6 (launch):       Polish stories, add clinical context docs

  Rule: component + test + story ship together. Always.
  But: test is written FIRST, story is written LAST.
```

---

## TDD: The Exact Process

### The cycle for every component

```
  1. Write the test          → it fails (component doesn't exist)
  2. Write minimal component → it passes
  3. Write a11y test         → it fails (missing ARIA)
  4. Add ARIA/keyboard       → it passes
  5. Write edge case tests   → some fail
  6. Handle edge cases       → all pass
  7. Write the story         → visual verification
  8. Check size budget       → must be under limit
  9. Ship it
```

### Real example: Building `<HumanName>` with TDD

**Step 1: Write the tests first**

```tsx
// human-name.test.tsx

import { render, screen } from '@testing-library/react'
import { axe } from 'vitest-axe'
import { HumanName } from './human-name'

// ─── Rendering ────────────────────────────────────────────

describe('HumanName', () => {
  it('renders a complete name', () => {
    render(
      <HumanName
        name={{
          use: 'official',
          prefix: ['Dr.'],
          given: ['John', 'Andrew'],
          family: 'Smith',
          suffix: ['Jr.'],
        }}
      />
    )
    expect(screen.getByText('Dr. John Andrew Smith Jr.')).toBeInTheDocument()
  })

  it('renders family-first format', () => {
    render(
      <HumanName
        name={{ given: ['John'], family: 'Smith' }}
        format="family-first"
      />
    )
    expect(screen.getByText('Smith, John')).toBeInTheDocument()
  })

  it('renders short format (given + family only)', () => {
    render(
      <HumanName
        name={{
          prefix: ['Dr.'],
          given: ['John', 'Andrew'],
          family: 'Smith',
          suffix: ['Jr.'],
        }}
        format="short"
      />
    )
    expect(screen.getByText('John Smith')).toBeInTheDocument()
  })

  it('shows use label when showUse is true', () => {
    render(
      <HumanName
        name={{ use: 'official', given: ['John'], family: 'Smith' }}
        showUse
      />
    )
    expect(screen.getByText('official')).toBeInTheDocument()
  })

  it('renders multiple names', () => {
    render(
      <HumanName
        name={[
          { use: 'official', given: ['John'], family: 'Smith' },
          { use: 'nickname', given: ['Johnny'] },
        ]}
        showUse
      />
    )
    expect(screen.getByText('John Smith')).toBeInTheDocument()
    expect(screen.getByText('Johnny')).toBeInTheDocument()
  })

  // ─── Edge Cases ───────────────────────────────────────────

  it('handles null name', () => {
    const { container } = render(<HumanName name={null as any} />)
    expect(container.textContent).toBe('')
  })

  it('handles undefined name', () => {
    const { container } = render(<HumanName name={undefined as any} />)
    expect(container.textContent).toBe('')
  })

  it('handles empty object', () => {
    const { container } = render(<HumanName name={{}} />)
    expect(container.textContent).toBe('')
  })

  it('handles name with only family', () => {
    render(<HumanName name={{ family: 'Smith' }} />)
    expect(screen.getByText('Smith')).toBeInTheDocument()
  })

  it('handles name with only given', () => {
    render(<HumanName name={{ given: ['John'] }} />)
    expect(screen.getByText('John')).toBeInTheDocument()
  })

  it('handles empty given array', () => {
    render(<HumanName name={{ given: [], family: 'Smith' }} />)
    expect(screen.getByText('Smith')).toBeInTheDocument()
  })

  it('handles text fallback when no structured name', () => {
    render(<HumanName name={{ text: 'Dr. John Smith' }} />)
    expect(screen.getByText('Dr. John Smith')).toBeInTheDocument()
  })

  // ─── Accessibility ───────────────────────────────────────

  it('has no accessibility violations', async () => {
    const { container } = render(
      <HumanName name={{ given: ['John'], family: 'Smith' }} />
    )
    expect(await axe(container)).toHaveNoViolations()
  })

  it('uses semantic HTML (span, not div)', () => {
    const { container } = render(
      <HumanName name={{ given: ['John'], family: 'Smith' }} />
    )
    expect(container.querySelector('span')).toBeInTheDocument()
  })

  // ─── Styling ──────────────────────────────────────────────

  it('accepts className prop', () => {
    const { container } = render(
      <HumanName
        name={{ given: ['John'], family: 'Smith' }}
        className="custom-class"
      />
    )
    expect(container.firstChild).toHaveClass('custom-class')
  })

  it('renders with data-ehr attribute for CSS targeting', () => {
    const { container } = render(
      <HumanName name={{ given: ['John'], family: 'Smith' }} />
    )
    expect(container.querySelector('[data-ehr="human-name"]')).toBeInTheDocument()
  })
})
```

**That's 16 tests. Written before a single line of component code exists.**
**Every test fails. Now we build the component to make them pass.**

**Step 2: Build the minimal component**

```tsx
// human-name.tsx

import type { HumanName as FHIRHumanName } from '@ehr/fhir-types'

export interface HumanNameProps {
  name: FHIRHumanName | FHIRHumanName[] | null | undefined
  format?: 'full' | 'short' | 'family-first'
  showUse?: boolean
  className?: string
}

export function HumanName({ name, format = 'full', showUse, className }: HumanNameProps) {
  if (!name) return null

  const names = Array.isArray(name) ? name : [name]
  const nonEmpty = names.filter(n => n && (n.given?.length || n.family || n.text))

  if (nonEmpty.length === 0) return null

  return (
    <span data-ehr="human-name" className={className}>
      {nonEmpty.map((n, i) => (
        <span key={i} data-ehr="human-name-entry">
          {formatName(n, format)}
          {showUse && n.use && (
            <>
              {' '}
              <span data-ehr="human-name-use">{n.use}</span>
            </>
          )}
        </span>
      ))}
    </span>
  )
}

function formatName(name: FHIRHumanName, format: string): string {
  // text fallback
  if (!name.given?.length && !name.family && name.text) {
    return name.text
  }

  const given = name.given?.join(' ') ?? ''
  const family = name.family ?? ''

  switch (format) {
    case 'short':
      return [given.split(' ')[0], family].filter(Boolean).join(' ')

    case 'family-first':
      return family && given ? `${family}, ${given}` : family || given

    case 'full':
    default: {
      const parts = [
        ...(name.prefix ?? []),
        given,
        family,
        ...(name.suffix ?? []),
      ].filter(Boolean)
      return parts.join(' ')
    }
  }
}
```

**Step 3: Run the tests**

```
$ pnpm test packages/react/src/fhir-primitives/human-name.test.tsx

 ✓ HumanName > renders a complete name
 ✓ HumanName > renders family-first format
 ✓ HumanName > renders short format
 ✓ HumanName > shows use label when showUse is true
 ✓ HumanName > renders multiple names
 ✓ HumanName > handles null name
 ✓ HumanName > handles undefined name
 ✓ HumanName > handles empty object
 ✓ HumanName > handles name with only family
 ✓ HumanName > handles name with only given
 ✓ HumanName > handles empty given array
 ✓ HumanName > handles text fallback
 ✓ HumanName > has no accessibility violations
 ✓ HumanName > uses semantic HTML
 ✓ HumanName > accepts className prop
 ✓ HumanName > renders with data-ehr attribute

 16/16 passed
```

**Step 4: Write the story (last)**

```tsx
// human-name.stories.tsx — written AFTER tests pass
```

---

### Real example: Building `<LabResults>` with TDD (complex component)

This is the hardest component. 15+ edge cases for abnormal flagging.
TDD is essential here — the logic is where bugs hide.

**Step 1: Test the pure logic FIRST (no React)**

```tsx
// lab-utils.test.ts — pure function tests, no rendering

import { flagLabValue, groupByPanel, calculateTrend } from './lab-utils'

describe('flagLabValue', () => {
  // ─── Normal values ──────────────────────────────────────

  it('flags value within range as normal', () => {
    expect(flagLabValue(85, { low: 70, high: 100 })).toEqual({
      flag: 'normal',
      severity: 'normal',
    })
  })

  it('flags value exactly at low boundary as normal', () => {
    expect(flagLabValue(70, { low: 70, high: 100 })).toEqual({
      flag: 'normal',
      severity: 'normal',
    })
  })

  it('flags value exactly at high boundary as normal', () => {
    expect(flagLabValue(100, { low: 70, high: 100 })).toEqual({
      flag: 'normal',
      severity: 'normal',
    })
  })

  // ─── Abnormal high ─────────────────────────────────────

  it('flags value above high as abnormal high', () => {
    expect(flagLabValue(145, { low: 70, high: 100 })).toEqual({
      flag: 'H',
      severity: 'abnormal_high',
    })
  })

  // ─── Abnormal low ──────────────────────────────────────

  it('flags value below low as abnormal low', () => {
    expect(flagLabValue(55, { low: 70, high: 100 })).toEqual({
      flag: 'L',
      severity: 'abnormal_low',
    })
  })

  // ─── Critical values ───────────────────────────────────

  it('flags critically high value', () => {
    expect(flagLabValue(180, { low: 70, high: 100 }, { low: 40, high: 160 })).toEqual({
      flag: 'HH',
      severity: 'critical_high',
    })
  })

  it('flags critically low value', () => {
    expect(flagLabValue(30, { low: 70, high: 100 }, { low: 40, high: 160 })).toEqual({
      flag: 'LL',
      severity: 'critical_low',
    })
  })

  // ─── Edge cases ─────────────────────────────────────────

  it('handles no reference range', () => {
    expect(flagLabValue(145, undefined)).toEqual({
      flag: null,
      severity: 'unknown',
    })
  })

  it('handles range with only high (e.g., HbA1c < 5.7)', () => {
    expect(flagLabValue(7.2, { high: 5.7 })).toEqual({
      flag: 'H',
      severity: 'abnormal_high',
    })
  })

  it('handles range with only low (e.g., eGFR > 60)', () => {
    expect(flagLabValue(45, { low: 60 })).toEqual({
      flag: 'L',
      severity: 'abnormal_low',
    })
  })

  it('handles zero value', () => {
    expect(flagLabValue(0, { low: 0, high: 100 })).toEqual({
      flag: 'normal',
      severity: 'normal',
    })
  })

  it('handles negative value (e.g., temperature in Celsius can be negative in some scales)', () => {
    expect(flagLabValue(-1, { low: 0, high: 100 })).toEqual({
      flag: 'L',
      severity: 'abnormal_low',
    })
  })

  it('handles string value "positive"', () => {
    expect(flagLabValue('positive', undefined)).toEqual({
      flag: 'abnormal',
      severity: 'abnormal_high',
    })
  })

  it('handles string value "negative"', () => {
    expect(flagLabValue('negative', undefined)).toEqual({
      flag: null,
      severity: 'normal',
    })
  })

  it('uses interpretation code when provided', () => {
    expect(flagLabValue(145, { low: 70, high: 100 }, undefined, 'H')).toEqual({
      flag: 'H',
      severity: 'abnormal_high',
    })
  })

  it('interpretation code overrides calculated flag', () => {
    // Value is in range but interpretation says high
    // (lab-specific reference ranges may differ)
    expect(flagLabValue(95, { low: 70, high: 100 }, undefined, 'H')).toEqual({
      flag: 'H',
      severity: 'abnormal_high',
    })
  })
})

describe('groupByPanel', () => {
  it('groups CBC components together', () => {
    const observations = [
      mockObs('2093-3', 'Hemoglobin', 14.5),
      mockObs('2571-8', 'Triglycerides', 175),
      mockObs('6690-2', 'WBC', 7.2),
      mockObs('789-8', 'RBC', 4.5),
      mockObs('2085-9', 'HDL', 45),
    ]

    const groups = groupByPanel(observations)

    expect(groups).toHaveLength(2)
    expect(groups[0].name).toBe('CBC')
    expect(groups[0].observations).toHaveLength(3) // Hgb, WBC, RBC
    expect(groups[1].name).toBe('Lipid Panel')
    expect(groups[1].observations).toHaveLength(2) // Triglycerides, HDL
  })

  it('puts ungrouped labs in "Other" panel', () => {
    const observations = [
      mockObs('unknown-code', 'Some obscure test', 42),
    ]

    const groups = groupByPanel(observations)
    expect(groups[0].name).toBe('Other')
  })

  it('handles empty array', () => {
    expect(groupByPanel([])).toEqual([])
  })
})

describe('calculateTrend', () => {
  it('returns "rising" when last 3 values increase', () => {
    const values = [
      { value: 100, date: '2024-01-01' },
      { value: 110, date: '2024-04-01' },
      { value: 120, date: '2024-07-01' },
    ]
    expect(calculateTrend(values)).toBe('rising')
  })

  it('returns "falling" when last 3 values decrease', () => {
    const values = [
      { value: 120, date: '2024-01-01' },
      { value: 110, date: '2024-04-01' },
      { value: 100, date: '2024-07-01' },
    ]
    expect(calculateTrend(values)).toBe('falling')
  })

  it('returns "stable" when values are within 5%', () => {
    const values = [
      { value: 100, date: '2024-01-01' },
      { value: 102, date: '2024-04-01' },
      { value: 99, date: '2024-07-01' },
    ]
    expect(calculateTrend(values)).toBe('stable')
  })

  it('returns null when fewer than 2 data points', () => {
    expect(calculateTrend([{ value: 100, date: '2024-01-01' }])).toBeNull()
  })

  it('returns null for empty array', () => {
    expect(calculateTrend([])).toBeNull()
  })
})
```

**Step 2: Test the React component**

```tsx
// lab-results.test.tsx

import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { axe } from 'vitest-axe'
import { LabResults } from './lab-results'
import { FHIRProvider } from '@ehr/fhir-hooks'
import { mockLabBundle, mockLabEmpty, mockLabCritical } from '@ehr/test-utils'

const wrapper = ({ children }) => (
  <FHIRProvider serverUrl="http://mock" mockData={mockLabBundle}>
    {children}
  </FHIRProvider>
)

describe('LabResults', () => {
  // ─── Rendering ──────────────────────────────────────────

  it('renders lab results grouped by panel', () => {
    render(<LabResults observations={mockLabBundle.entry.map(e => e.resource)} />)

    expect(screen.getByText('Basic Metabolic Panel')).toBeInTheDocument()
    expect(screen.getByText('Glucose')).toBeInTheDocument()
    expect(screen.getByText('145')).toBeInTheDocument()
    expect(screen.getByText('mg/dL')).toBeInTheDocument()
  })

  it('shows reference range column', () => {
    render(<LabResults observations={mockLabBundle.entry.map(e => e.resource)} />)
    expect(screen.getByText('70 - 100')).toBeInTheDocument()
  })

  it('highlights abnormal values with H flag', () => {
    render(<LabResults observations={mockLabBundle.entry.map(e => e.resource)} />)

    const glucoseRow = screen.getByText('Glucose').closest('tr')!
    expect(within(glucoseRow).getByText('H')).toBeInTheDocument()
    expect(within(glucoseRow).getByText('H')).toHaveAttribute(
      'data-ehr-severity', 'abnormal_high'
    )
  })

  it('highlights critical values with HH flag', () => {
    render(<LabResults observations={mockLabCritical} />)

    const potassiumRow = screen.getByText('Potassium').closest('tr')!
    expect(within(potassiumRow).getByText('HH')).toBeInTheDocument()
    expect(within(potassiumRow).getByText('HH')).toHaveAttribute(
      'data-ehr-severity', 'critical_high'
    )
  })

  it('does not flag normal values', () => {
    render(<LabResults observations={mockLabBundle.entry.map(e => e.resource)} />)

    const bunRow = screen.getByText('BUN').closest('tr')!
    expect(within(bunRow).queryByText('H')).not.toBeInTheDocument()
    expect(within(bunRow).queryByText('L')).not.toBeInTheDocument()
  })

  // ─── Empty & Loading States ─────────────────────────────

  it('shows empty state when no results', () => {
    render(<LabResults observations={[]} />)
    expect(screen.getByText('No lab results found')).toBeInTheDocument()
  })

  it('shows loading skeleton when loading', () => {
    render(<LabResults patientId="123" />, { wrapper })
    // Loading state shows skeleton rows
    expect(screen.getAllByRole('status')).toHaveLength(1) // spinner or skeleton
  })

  it('shows error state on fetch failure', () => {
    render(
      <FHIRProvider serverUrl="http://mock" mockData={{ error: true }}>
        <LabResults patientId="123" />
      </FHIRProvider>
    )
    expect(screen.getByRole('alert')).toBeInTheDocument()
  })

  // ─── Interaction ────────────────────────────────────────

  it('expands a panel when clicked', async () => {
    render(<LabResults observations={mockLabBundle.entry.map(e => e.resource)} />)

    const bmpHeader = screen.getByText('Basic Metabolic Panel')
    await userEvent.click(bmpHeader)

    // Panel should be expanded, individual tests visible
    expect(screen.getByText('Glucose')).toBeVisible()
    expect(screen.getByText('BUN')).toBeVisible()
  })

  it('collapses a panel when clicked again', async () => {
    render(<LabResults observations={mockLabBundle.entry.map(e => e.resource)} />)

    const bmpHeader = screen.getByText('Basic Metabolic Panel')
    await userEvent.click(bmpHeader)  // expand
    await userEvent.click(bmpHeader)  // collapse

    // Individual tests should be hidden
    // (depends on implementation — may use display:none or not render)
  })

  // ─── Keyboard Navigation ───────────────────────────────

  it('panels are keyboard navigable', async () => {
    render(<LabResults observations={mockLabBundle.entry.map(e => e.resource)} />)

    await userEvent.tab()  // focus first panel header
    expect(screen.getByText('Basic Metabolic Panel').closest('[role="button"]')).toHaveFocus()

    await userEvent.keyboard('{Enter}')  // expand
    await userEvent.keyboard('{Escape}')  // collapse
  })

  it('table rows are navigable with arrow keys', async () => {
    render(<LabResults observations={mockLabBundle.entry.map(e => e.resource)} />)

    // Expand panel first
    await userEvent.click(screen.getByText('Basic Metabolic Panel'))

    // Tab into the table
    await userEvent.tab()
    await userEvent.tab()

    // Arrow down through rows
    await userEvent.keyboard('{ArrowDown}')
  })

  // ─── Accessibility ─────────────────────────────────────

  it('has no accessibility violations', async () => {
    const { container } = render(
      <LabResults observations={mockLabBundle.entry.map(e => e.resource)} />
    )
    expect(await axe(container)).toHaveNoViolations()
  })

  it('table has proper ARIA attributes', () => {
    render(<LabResults observations={mockLabBundle.entry.map(e => e.resource)} />)

    const table = screen.getByRole('table')
    expect(table).toHaveAttribute('aria-label')
  })

  it('abnormal flags have aria-label for screen readers', () => {
    render(<LabResults observations={mockLabBundle.entry.map(e => e.resource)} />)

    const flag = screen.getByText('H')
    expect(flag).toHaveAttribute('aria-label', 'Abnormal high')
  })

  it('critical flags use role="alert" for screen reader announcement', () => {
    render(<LabResults observations={mockLabCritical} />)

    const flag = screen.getByText('HH')
    expect(flag.closest('[role="alert"]')).toBeInTheDocument()
  })

  // ─── Styling ────────────────────────────────────────────

  it('accepts className prop', () => {
    const { container } = render(
      <LabResults observations={mockLabBundle.entry.map(e => e.resource)} className="my-class" />
    )
    expect(container.firstChild).toHaveClass('my-class')
  })

  it('uses clinical.lab tokens for severity colors', () => {
    render(<LabResults observations={mockLabBundle.entry.map(e => e.resource)} />)

    const flag = screen.getByText('H')
    expect(flag).toHaveAttribute('data-ehr-severity', 'abnormal_high')
    // CSS applies: [data-ehr-severity="abnormal_high"] {
    //   color: var(--ehr-lab-abnormal-high-fg);
    //   background: var(--ehr-lab-abnormal-high-bg);
    // }
  })

  // ─── Performance ────────────────────────────────────────

  it('renders 100 observations in under 50ms', () => {
    const manyObs = Array.from({ length: 100 }, (_, i) =>
      mockObs(`code-${i}`, `Test ${i}`, Math.random() * 200)
    )

    const start = performance.now()
    render(<LabResults observations={manyObs} />)
    const duration = performance.now() - start

    expect(duration).toBeLessThan(50)
  })
})
```

**That's 20+ tests for LabResults. All written before the component.**

---

## Test Categories

Every component has tests in these 7 categories:

```
+-------------------------------------------------------------------+
| Category        | What It Tests                | Required?         |
|-----------------|------------------------------|-------------------|
| Rendering       | Does it render correct data? | YES — every comp  |
| Edge cases      | null, undefined, empty, max  | YES — every comp  |
| Empty state     | No data → meaningful message | YES — data comps  |
| Loading state   | Loading → skeleton/spinner   | YES — async comps |
| Error state     | Error → error message        | YES — async comps |
| Interaction     | Click, type, expand, select  | YES — interactive |
| Keyboard        | Tab, arrow, enter, escape    | YES — interactive |
| Accessibility   | axe-core zero violations     | YES — every comp  |
| ARIA            | Roles, labels, live regions  | YES — every comp  |
| Styling         | className, data-ehr attrs    | YES — every comp  |
| Performance     | Render time under budget     | YES — list comps  |
+-------------------------------------------------------------------+
```

---

## Coverage Targets

```
+-------------------------------------------------------------------+
| Package          | Statements | Branches | Functions | Lines      |
|------------------|------------|----------|-----------|------------|
| @ehr/tokens      |    100%    |   100%   |    100%   |   100%     |
| @ehr/fhir-types  |    N/A     |   N/A    |    N/A    |   N/A      |
| @ehr/fhir-hooks  |     95%    |    90%   |     95%   |    95%     |
| @ehr/primitives  |     95%    |    90%   |     95%   |    95%     |
| @ehr/react       |     90%    |    85%   |     90%   |    90%     |
| @ehr/smart-auth  |     90%    |    85%   |     90%   |    90%     |
|------------------|------------|----------|-----------|------------|
| OVERALL          |     92%    |    88%   |     92%   |    92%     |
+-------------------------------------------------------------------+

  Why not 100% everywhere?
  - 100% coverage on tokens: they're pure data, easy to cover
  - 95% on hooks: some error paths in WebSocket are hard to test
  - 90% on react components: some visual edge cases need visual tests
  - fhir-types: no runtime code, just TypeScript interfaces, N/A

  Branch coverage is lower because:
  - FHIR resources have many optional fields
  - Every combination of null/undefined isn't worth testing
  - We test the important branches (clinical logic, edge cases)
  - 85-90% branch coverage catches real bugs without test bloat
```

### Enforcement

```ts
// vitest.config.ts

export default defineConfig({
  test: {
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      thresholds: {
        statements: 90,
        branches: 85,
        functions: 90,
        lines: 90,
      },
      // Fail the build if coverage drops below thresholds
      check: true,
    },
  },
})
```

```yaml
# .github/workflows/ci.yml

- name: Test with coverage
  run: pnpm test -- --coverage

- name: Check coverage thresholds
  run: pnpm coverage:check
  # Fails CI if any package drops below its threshold
```

---

## Test Infrastructure

### Mock FHIR Data: `@ehr/test-utils` (internal package)

```
packages/
  test-utils/
    src/
      mocks/
        patient.ts           # mockPatient, mockPatientMinimal, etc.
        observation.ts       # mockLabBundle, mockVitals, mockLabCritical
        condition.ts         # mockConditions, mockConditionsEmpty
        medication.ts        # mockMedications, mockMedInteraction
        allergy.ts           # mockAllergies, mockAllergyAnaphylaxis
        appointment.ts       # mockAppointments
        questionnaire.ts     # mockPHQ9, mockIntakeForm
        bundle.ts            # mockSearchBundle, mockEmptyBundle
      builders/
        patient-builder.ts   # fluent builder: buildPatient().withName(...).build()
        observation-builder.ts
      server/
        mock-fhir-server.ts  # MSW handlers for FHIR REST API
      index.ts
```

```ts
// Usage in tests:

import { mockPatient, buildPatient, mockFHIRServer } from '@ehr/test-utils'

// Pre-built mock
render(<PatientBanner patient={mockPatient} />)

// Custom mock with builder
const patient = buildPatient()
  .withName('Jane', 'Doe')
  .withBirthDate('1990-05-15')
  .withAllergy('Penicillin', 'severe')
  .build()
render(<PatientBanner patient={patient} />)

// Full mock server (MSW)
beforeAll(() => mockFHIRServer.listen())
afterAll(() => mockFHIRServer.close())
render(<PatientBanner patientId="123" />) // fetches from mock server
```

### Why MSW for mock server:

```
  MSW (Mock Service Worker) intercepts fetch requests at the network level.
  No custom test client needed. The component makes real fetch calls.
  MSW catches them and returns mock FHIR data.

  This means our tests verify:
  - The component makes the correct FHIR request URL
  - The component handles the response correctly
  - The component handles errors correctly
  - All without a real FHIR server running
```

### Visual Regression (Phase 6+, not day 1)

```
  We do NOT use Chromatic (expensive, external dependency).

  Instead, Playwright screenshot tests:

  // patient-banner.visual.test.ts

  test('PatientBanner matches snapshot', async ({ page }) => {
    await page.goto('/storybook/iframe.html?id=resources-patientbanner--default')
    await expect(page).toHaveScreenshot('patient-banner-default.png')
  })

  test('PatientBanner dark mode matches snapshot', async ({ page }) => {
    await page.goto('/storybook/iframe.html?id=resources-patientbanner--default&theme=dark')
    await expect(page).toHaveScreenshot('patient-banner-dark.png')
  })

  When: Added in Phase 6 (polish), not Phase 0.
  Why not earlier: Screenshots are brittle during rapid iteration.
  Once components are stable, screenshots catch regressions.
```

---

## CI Pipeline: What Runs on Every PR

```
+-------------------------------------------------------------------+
|  GitHub Actions CI                            ~3 min total        |
+-------------------------------------------------------------------+
|                                                                   |
|  1. Install (pnpm install --frozen-lockfile)         ~30s         |
|                                                                   |
|  2. Parallel:                                                     |
|     +---------------------------------------------------+         |
|     | a. Lint + Format (biome check)            ~15s    |         |
|     | b. Type Check (tsc --noEmit)              ~20s    |         |
|     | c. Unit Tests (vitest --coverage)          ~45s    |         |
|     | d. Build (tsup all packages)              ~30s    |         |
|     +---------------------------------------------------+         |
|                                                                   |
|  3. After build:                                                  |
|     +---------------------------------------------------+         |
|     | e. Size Limit check                       ~5s     |         |
|     | f. Package validation (publint)            ~5s     |         |
|     | g. Storybook build (static)               ~30s    |         |
|     +---------------------------------------------------+         |
|                                                                   |
|  4. Coverage Report                                               |
|     - Post coverage summary as PR comment                         |
|     - Fail if below thresholds                                    |
|                                                                   |
|  Blockers (PR cannot merge if ANY fail):                          |
|    [x] All tests pass                                             |
|    [x] Zero a11y violations                                       |
|    [x] Coverage above thresholds                                  |
|    [x] Size budget not exceeded                                   |
|    [x] Types compile                                              |
|    [x] Lint clean                                                 |
|    [x] Packages valid (publint)                                   |
|    [x] Storybook builds                                           |
|                                                                   |
+-------------------------------------------------------------------+
```

---

## Test Count Estimates Per Phase

```
+-------------------------------------------------------------------+
| Phase | What                    | Components | Tests  | Coverage  |
|-------|-------------------------|------------|--------|-----------|
|   0   | Scaffolding             |     1      |    5   | 100%      |
|   1   | Tokens + Types + Client |     0      |   45   | 100%/95%  |
|   2   | 8 Hooks                 |     0      |   80   | 95%       |
|   3   | 20 Primitives           |    20      |  200   | 95%       |
|   4   | 18 FHIR Primitives      |    18      |  250   | 95%       |
|   5   | 10 Resource Components  |    10      |  300   | 90%       |
|   6   | Polish + Visual         |     0      |   50   | same      |
|-------|-------------------------|------------|--------|-----------|
| TOTAL |                         |    49      |  930   | 92%+      |
+-------------------------------------------------------------------+

  930 tests for 49 components = ~19 tests per component average.

  Simple components (Badge, Divider): 8-10 tests
  Medium components (PatientBanner, ProblemList): 15-20 tests
  Complex components (LabResults, QuestionnaireForm): 30-40 tests

  Pure function tests (lab-utils, name-formatting): 15-25 tests
  Hook tests: 8-12 tests per hook
```

---

## What "Done" Looks Like

```
  $ pnpm test

  ✓ @ehr/tokens           12 tests passed    100% coverage
  ✓ @ehr/fhir-hooks       80 tests passed     96% coverage
  ✓ @ehr/primitives      200 tests passed     95% coverage
  ✓ @ehr/react           350 tests passed     91% coverage

  Total: 642 tests passed, 0 failed
  Coverage: 92.4% statements, 88.1% branches

  $ pnpm build

  @ehr/tokens       2.1 KB (gzip)   PASS  budget: 5 KB
  @ehr/fhir-types   0 KB (types only)
  @ehr/fhir-hooks   8.3 KB (gzip)   PASS  budget: 15 KB
  @ehr/primitives  14.2 KB (gzip)   PASS  budget: 20 KB
  @ehr/react       31.5 KB (gzip)   PASS  budget: 45 KB
  styles.css        9.8 KB (gzip)   PASS  budget: 15 KB

  $ pnpm lint

  No issues found.

  $ pnpm storybook:build

  Storybook built successfully. 49 stories.

  Ready to ship.
```
