# @ehr/tokens

Design tokens for healthcare user interfaces. Provides a complete set of colors, spacing, typography, and clinical semantic values as both JavaScript objects and CSS custom properties.

## Installation

```bash
pnpm add @ehr/tokens
```

## Usage

### CSS Custom Properties (recommended)

Import the CSS file once at your application root. All tokens become available as `--ehr-*` variables.

```tsx
import '@ehr/tokens/css'
```

```css
.my-card {
  padding: var(--ehr-space-4);
  border-radius: var(--ehr-radius-lg);
  background: var(--ehr-bg-primary);
  color: var(--ehr-fg-primary);
  box-shadow: var(--ehr-shadow-sm);
  font-family: var(--ehr-font-sans);
}
```

### JavaScript / TypeScript

Import token objects directly for use in JS, theme configuration, or tooling.

```ts
import { colors, spacing, fontSize, severity, status } from '@ehr/tokens'

colors.blue[500]     // '#3B82F6'
spacing[4]           // '1rem'
fontSize.lg          // '1.125rem'
severity.critical.bg // '#B91C1C'
status.active.dot    // '#22C55E'
```

## Token Categories

### Core Design Tokens

| Category | Tokens | CSS Prefix |
|----------|--------|------------|
| **Colors** | 8 palettes (gray, blue, red, orange, yellow, green, teal, purple) x 11 shades | `--ehr-{color}-{shade}` |
| **Spacing** | 0 to 96 scale in rem | `--ehr-space-{n}` |
| **Typography** | Font families, sizes (xs-5xl), weights, letter-spacing | `--ehr-text-*`, `--ehr-weight-*`, `--ehr-font-*` |
| **Radius** | none, sm, md, lg, xl, 2xl, full | `--ehr-radius-{size}` |
| **Shadows** | xs, sm, md, lg, xl, inner | `--ehr-shadow-{size}` |
| **Z-Index** | hide(-1) to max(99) with named layers | N/A (JS only) |
| **Motion** | Duration (instant-slower), easing curves | `--ehr-duration-*`, `--ehr-easing-*` |
| **Breakpoints** | sm(640px) to 2xl(1536px) | N/A (JS only) |

### Semantic Tokens

Surface and text colors that automatically adapt between light and dark themes:

```css
--ehr-bg-primary      /* Main background */
--ehr-bg-secondary    /* Subtle background */
--ehr-fg-primary      /* Main text */
--ehr-fg-secondary    /* Muted text */
--ehr-border-default  /* Standard borders */
--ehr-interactive     /* Buttons, links */
--ehr-focus-ring      /* Accessible focus indicator */
```

### Feedback Tokens

```css
--ehr-success-fg / bg / border
--ehr-warning-fg / bg / border
--ehr-danger-fg  / bg / border
--ehr-info-fg    / bg / border
```

### Clinical Tokens

Healthcare-specific semantic values for clinical data display:

| Token Group | Use Case | Values |
|-------------|----------|--------|
| **Severity** | Allergy alerts, CDS cards | critical, high, moderate, low, info |
| **Status** | Problem lists, medication status | active, inactive, resolved, draft, on_hold, completed, cancelled, entered_in_error |
| **Lab Flags** | Lab results, vitals | normal, H, L, HH, LL |
| **Priority** | Task cards, orders | stat, urgent, routine, elective |
| **Encounter Colors** | Timelines, calendars | office_visit, telehealth, emergency, inpatient, observation, procedure, imaging, lab |

```ts
import { severity, labFlags, priority } from '@ehr/tokens'

severity.critical  // { fg: '#FFFFFF', bg: '#B91C1C', border: '#991B1B', label: 'Critical' }
labFlags.abnormal_high // { fg: '#B91C1C', bg: '#FEF2F2', weight: '600', flag: 'H' }
priority.stat      // { fg: '#FFFFFF', bg: '#DC2626', label: 'STAT' }
```

## Theming

### Dark Mode

Add `data-ehr-theme="dark"` to any parent element. All semantic tokens invert automatically.

```html
<body data-ehr-theme="dark">
  <!-- All child components use dark theme -->
</body>
```

### Custom Themes

Override any CSS custom property to create a custom theme:

```css
[data-ehr-theme="custom"] {
  --ehr-blue-500: #0066CC;
  --ehr-interactive: #0066CC;
  --ehr-bg-primary: #FAFAFA;
}
```

## TypeScript

All tokens are typed with `as const`. Type helpers are exported:

```ts
import type {
  ColorScale, ColorShade,
  SpacingToken,
  FontSizeToken, FontWeightToken,
  RadiusToken, ShadowToken,
  DurationToken, EasingToken,
  BreakpointToken,
  SeverityLevel, StatusType, LabFlagType, PriorityLevel,
} from '@ehr/tokens'
```

## License

Apache-2.0
