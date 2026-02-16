# @ehr/design-system — Healthcare Design System Architecture

> Tokens. Themes. Primitives. Everything composes from the foundation.
> Developers customize, extend, override — the core just works.

---

## Table of Contents

1. [Philosophy](#philosophy)
2. [Design Token Architecture](#design-token-architecture)
3. [Theme System](#theme-system)
4. [Core Primitives](#core-primitives)
5. [Component Architecture Patterns](#component-architecture-patterns)
6. [Clinical Design Language](#clinical-design-language)
7. [Extending & Customizing](#extending--customizing)
8. [Package Map](#package-map)

---

## Philosophy

```
+-------------------------------------------------------------------+
|                                                                   |
|   "The design system is not the components.                       |
|    The design system is the shared language                       |
|    that makes components inevitable."                             |
|                                                                   |
+-------------------------------------------------------------------+

  Tokens  -->  Theme  -->  Primitives  -->  Components  -->  Workflows
  (atoms)    (decisions)   (building      (compositions)   (experiences)
                            blocks)
```

### Principles

1. **Tokens are the single source of truth** — Every color, spacing value,
   font size, shadow, radius, motion curve is a token. Nothing is hardcoded.

2. **Themes compose tokens into decisions** — A "danger" color isn't red.
   It's `token.color.danger.500`. The theme decides what red that maps to.
   Clinical severity, alert levels, status colors — all semantic tokens.

3. **Core primitives are unstyled logic** — `<Button>`, `<Input>`, `<Select>`,
   `<Dialog>`, `<Popover>` — these handle behavior, accessibility, keyboard
   navigation. They consume tokens. They render nothing visual by default.

4. **Styled components apply the theme** — The pre-built Tailwind layer
   maps tokens to Tailwind utilities. But swap it out for Material,
   Chakra, Ant, or your own CSS. The logic doesn't change.

5. **Developers extend, never fork** — Custom tokens merge into the system.
   Custom themes extend the defaults. Custom components compose from primitives.
   The escape hatch is built into the architecture, not bolted on.

---

## Design Token Architecture

### Token Hierarchy

```
+-------------------------------------------------------------------+
|                        TOKEN LAYERS                               |
+-------------------------------------------------------------------+
|                                                                   |
|  Layer 1: GLOBAL TOKENS (raw values)                              |
|  +-------------------------------------------------------------+  |
|  |  color.blue.500 = #3B82F6                                   |  |
|  |  color.red.600 = #DC2626                                    |  |
|  |  spacing.4 = 1rem                                            |  |
|  |  font.size.sm = 0.875rem                                     |  |
|  |  radius.md = 0.375rem                                        |  |
|  |  shadow.sm = 0 1px 2px rgba(0,0,0,0.05)                     |  |
|  +-------------------------------------------------------------+  |
|           |                                                       |
|           v                                                       |
|  Layer 2: SEMANTIC TOKENS (design decisions)                      |
|  +-------------------------------------------------------------+  |
|  |  color.primary = {color.blue.500}                            |  |
|  |  color.danger = {color.red.600}                              |  |
|  |  color.surface = {color.white}                               |  |
|  |  spacing.component.gap = {spacing.3}                         |  |
|  |  font.body = {font.size.sm}                                  |  |
|  |  radius.control = {radius.md}                                |  |
|  +-------------------------------------------------------------+  |
|           |                                                       |
|           v                                                       |
|  Layer 3: COMPONENT TOKENS (per-component)                        |
|  +-------------------------------------------------------------+  |
|  |  button.bg = {color.primary}                                 |  |
|  |  button.radius = {radius.control}                            |  |
|  |  button.padding.x = {spacing.4}                              |  |
|  |  input.border = {color.border}                               |  |
|  |  alert.danger.bg = {color.danger.50}                         |  |
|  +-------------------------------------------------------------+  |
|           |                                                       |
|           v                                                       |
|  Layer 4: CLINICAL TOKENS (healthcare-specific)                   |
|  +-------------------------------------------------------------+  |
|  |  severity.critical = {color.red.700}                         |  |
|  |  severity.high = {color.orange.600}                          |  |
|  |  severity.moderate = {color.yellow.500}                      |  |
|  |  severity.low = {color.blue.500}                             |  |
|  |  status.active = {color.green.600}                           |  |
|  |  status.resolved = {color.gray.400}                          |  |
|  |  lab.normal = {color.green.600}                              |  |
|  |  lab.abnormal.high = {color.red.600}                         |  |
|  |  lab.abnormal.low = {color.blue.600}                         |  |
|  |  lab.critical = {color.red.800}                              |  |
|  +-------------------------------------------------------------+  |
|                                                                   |
+-------------------------------------------------------------------+
```

---

### Global Tokens (Layer 1)

These are the raw design primitives. Pure values. No opinion.

#### Colors

```
+-------------------------------------------------------------------+
|  COLOR PALETTE                                                    |
+-------------------------------------------------------------------+
|                                                                   |
|  Gray (Neutral)                                                   |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|  |50 |100|150|200|300|400|500|600|700|800|900|950|                 |
|  +---+---+---+---+---+---+---+---+---+---+---+---+               |
|  [___][___][___][___][___][___][###][###][###][###]                |
|                                                                   |
|  Blue (Primary)                                                   |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|  |50 |100|200|300|400|500|600|700|800|900|950|                    |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|  [___][___][___][___][___][###][###][###][###][###]                |
|                                                                   |
|  Red (Danger / Critical)                                          |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|  |50 |100|200|300|400|500|600|700|800|900|950|                    |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|                                                                   |
|  Orange (Warning)                                                 |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|  |50 |100|200|300|400|500|600|700|800|900|950|                    |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|                                                                   |
|  Yellow (Caution)                                                 |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|  |50 |100|200|300|400|500|600|700|800|900|950|                    |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|                                                                   |
|  Green (Success / Normal)                                         |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|  |50 |100|200|300|400|500|600|700|800|900|950|                    |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|                                                                   |
|  Teal (Clinical Info)                                             |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|  |50 |100|200|300|400|500|600|700|800|900|950|                    |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|                                                                   |
|  Purple (Specialty / Research)                                    |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|  |50 |100|200|300|400|500|600|700|800|900|950|                    |
|  +---+---+---+---+---+---+---+---+---+---+---+                   |
|                                                                   |
+-------------------------------------------------------------------+
```

```ts
// tokens/colors.ts

export const colors = {
  // Grayscale
  white:    '#FFFFFF',
  black:    '#000000',
  gray: {
    50:  '#F9FAFB',   100: '#F3F4F6',   150: '#ECEDF0',
    200: '#E5E7EB',   300: '#D1D5DB',   400: '#9CA3AF',
    500: '#6B7280',   600: '#4B5563',   700: '#374151',
    800: '#1F2937',   900: '#111827',   950: '#030712',
  },

  // Primary — Blue
  blue: {
    50:  '#EFF6FF',   100: '#DBEAFE',   200: '#BFDBFE',
    300: '#93C5FD',   400: '#60A5FA',   500: '#3B82F6',
    600: '#2563EB',   700: '#1D4ED8',   800: '#1E40AF',
    900: '#1E3A8A',   950: '#172554',
  },

  // Danger — Red
  red: {
    50:  '#FEF2F2',   100: '#FEE2E2',   200: '#FECACA',
    300: '#FCA5A5',   400: '#F87171',   500: '#EF4444',
    600: '#DC2626',   700: '#B91C1C',   800: '#991B1B',
    900: '#7F1D1D',   950: '#450A0A',
  },

  // Warning — Orange
  orange: {
    50:  '#FFF7ED',   100: '#FFEDD5',   200: '#FED7AA',
    300: '#FDBA74',   400: '#FB923C',   500: '#F97316',
    600: '#EA580C',   700: '#C2410C',   800: '#9A3412',
    900: '#7C2D12',   950: '#431407',
  },

  // Caution — Yellow/Amber
  yellow: {
    50:  '#FFFBEB',   100: '#FEF3C7',   200: '#FDE68A',
    300: '#FCD34D',   400: '#FBBF24',   500: '#F59E0B',
    600: '#D97706',   700: '#B45309',   800: '#92400E',
    900: '#78350F',   950: '#451A03',
  },

  // Success — Green
  green: {
    50:  '#F0FDF4',   100: '#DCFCE7',   200: '#BBF7D0',
    300: '#86EFAC',   400: '#4ADE80',   500: '#22C55E',
    600: '#16A34A',   700: '#15803D',   800: '#166534',
    900: '#14532D',   950: '#052E16',
  },

  // Clinical Info — Teal
  teal: {
    50:  '#F0FDFA',   100: '#CCFBF1',   200: '#99F6E4',
    300: '#5EEAD4',   400: '#2DD4BF',   500: '#14B8A6',
    600: '#0D9488',   700: '#0F766E',   800: '#115E59',
    900: '#134E4A',   950: '#042F2E',
  },

  // Specialty — Purple
  purple: {
    50:  '#FAF5FF',   100: '#F3E8FF',   200: '#E9D5FF',
    300: '#D8B4FE',   400: '#C084FC',   500: '#A855F7',
    600: '#9333EA',   700: '#7E22CE',   800: '#6B21A8',
    900: '#581C87',   950: '#3B0764',
  },
} as const
```

#### Spacing

```ts
// tokens/spacing.ts

export const spacing = {
  px:  '1px',
  0:   '0',
  0.5: '0.125rem',   //  2px
  1:   '0.25rem',    //  4px
  1.5: '0.375rem',   //  6px
  2:   '0.5rem',     //  8px
  2.5: '0.625rem',   // 10px
  3:   '0.75rem',    // 12px
  3.5: '0.875rem',   // 14px
  4:   '1rem',       // 16px   <-- base unit
  5:   '1.25rem',    // 20px
  6:   '1.5rem',     // 24px
  7:   '1.75rem',    // 28px
  8:   '2rem',       // 32px
  9:   '2.25rem',    // 36px
  10:  '2.5rem',     // 40px
  11:  '2.75rem',    // 44px
  12:  '3rem',       // 48px
  14:  '3.5rem',     // 56px
  16:  '4rem',       // 64px
  20:  '5rem',       // 80px
  24:  '6rem',       // 96px
  28:  '7rem',       // 112px
  32:  '8rem',       // 128px
  36:  '9rem',       // 144px
  40:  '10rem',      // 160px
  44:  '11rem',      // 176px
  48:  '12rem',      // 192px
  52:  '13rem',      // 208px
  56:  '14rem',      // 224px
  60:  '15rem',      // 240px
  64:  '16rem',      // 256px
  72:  '18rem',      // 288px
  80:  '20rem',      // 320px
  96:  '24rem',      // 384px
} as const

//  Spacing scale visual:
//
//  0.5  [==]
//  1    [====]
//  2    [========]
//  3    [============]
//  4    [================]               <-- base (16px)
//  6    [========================]
//  8    [================================]
//  12   [================================================]
//  16   [================================================================]
```

#### Typography

```ts
// tokens/typography.ts

export const fontFamily = {
  sans:  'Inter, system-ui, -apple-system, sans-serif',
  mono:  'JetBrains Mono, Menlo, Monaco, monospace',
  clinical: 'Inter, system-ui, sans-serif',  // alias — clinical = sans
} as const

export const fontSize = {
  xs:   ['0.75rem',   { lineHeight: '1rem'    }],   // 12px — fine print
  sm:   ['0.875rem',  { lineHeight: '1.25rem' }],   // 14px — body small
  base: ['1rem',      { lineHeight: '1.5rem'  }],   // 16px — body
  lg:   ['1.125rem',  { lineHeight: '1.75rem' }],   // 18px — body large
  xl:   ['1.25rem',   { lineHeight: '1.75rem' }],   // 20px — heading 4
  '2xl':['1.5rem',    { lineHeight: '2rem'    }],   // 24px — heading 3
  '3xl':['1.875rem',  { lineHeight: '2.25rem' }],   // 30px — heading 2
  '4xl':['2.25rem',   { lineHeight: '2.5rem'  }],   // 36px — heading 1
  '5xl':['3rem',      { lineHeight: '1'       }],   // 48px — display
} as const

export const fontWeight = {
  normal:   '400',
  medium:   '500',
  semibold: '600',
  bold:     '700',
} as const

export const letterSpacing = {
  tighter: '-0.05em',
  tight:   '-0.025em',
  normal:  '0em',
  wide:    '0.025em',
  wider:   '0.05em',
} as const

//  Typography scale visual:
//
//  xs     The quick brown fox       (12px / fine print, labels)
//  sm     The quick brown fox       (14px / body text, tables)
//  base   The quick brown fox       (16px / default body)
//  lg     The quick brown fox       (18px / emphasis)
//  xl     The quick brown fox    (20px / section heading)
//  2xl    The quick brown fox  (24px / page heading)
//  3xl    The quick brown    (30px / title)
//  4xl    The quick bro   (36px / display)
```

#### Radius, Shadow, Z-Index, Motion

```ts
// tokens/radius.ts
export const radius = {
  none: '0',
  sm:   '0.125rem',   // 2px  — subtle
  md:   '0.375rem',   // 6px  — default controls
  lg:   '0.5rem',     // 8px  — cards
  xl:   '0.75rem',    // 12px — large cards
  '2xl':'1rem',       // 16px — modals
  full: '9999px',     // pill shape — badges, avatars
} as const

// tokens/shadow.ts
export const shadow = {
  none: 'none',
  xs:   '0 1px 2px 0 rgba(0,0,0,0.05)',
  sm:   '0 1px 3px 0 rgba(0,0,0,0.1), 0 1px 2px -1px rgba(0,0,0,0.1)',
  md:   '0 4px 6px -1px rgba(0,0,0,0.1), 0 2px 4px -2px rgba(0,0,0,0.1)',
  lg:   '0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -4px rgba(0,0,0,0.1)',
  xl:   '0 20px 25px -5px rgba(0,0,0,0.1), 0 8px 10px -6px rgba(0,0,0,0.1)',
  inner:'inset 0 2px 4px 0 rgba(0,0,0,0.05)',
} as const

// tokens/z-index.ts
export const zIndex = {
  hide:      -1,
  base:       0,
  dropdown:  10,      // Autocomplete dropdowns, select menus
  sticky:    20,      // Sticky headers, patient banner
  overlay:   30,      // Modal backdrop
  modal:     40,      // Modal content
  popover:   50,      // Tooltips, popovers
  toast:     60,      // Toast notifications
  alert:     70,      // Critical clinical alerts (drug interactions)
  max:       99,      // Emergency (code blue) — nothing covers this
} as const

// tokens/motion.ts
export const duration = {
  instant:  '0ms',
  fast:     '100ms',
  normal:   '200ms',
  slow:     '300ms',
  slower:   '500ms',
} as const

export const easing = {
  default:   'cubic-bezier(0.4, 0, 0.2, 1)',
  in:        'cubic-bezier(0.4, 0, 1, 1)',
  out:       'cubic-bezier(0, 0, 0.2, 1)',
  inOut:     'cubic-bezier(0.4, 0, 0.2, 1)',
  spring:    'cubic-bezier(0.175, 0.885, 0.32, 1.275)',
} as const
```

#### Breakpoints

```ts
// tokens/breakpoints.ts
export const breakpoints = {
  sm:  '640px',    // Mobile landscape
  md:  '768px',    // Tablet portrait
  lg:  '1024px',   // Tablet landscape / small desktop
  xl:  '1280px',   // Desktop
  '2xl':'1536px',  // Large desktop / clinical workstation
} as const

//  Breakpoint visual:
//
//  Mobile    Tablet      Desktop        Workstation
//  |--sm--|---md---|---lg---|----xl----|-----2xl------|
//  640    768      1024     1280       1536
//
//  Clinical note: Most EHR workstations are 1920x1080 (xl-2xl range).
//  Tablets at bedside are typically md-lg range.
//  Patient portals must work from sm upward.
```

---

### Semantic Tokens (Layer 2)

These map raw values to design decisions. This is where theming happens.

```ts
// tokens/semantic.ts

export const semantic = {

  // ------ SURFACES ------
  color: {
    // Backgrounds
    bg: {
      primary:    '{color.white}',
      secondary:  '{color.gray.50}',
      tertiary:   '{color.gray.100}',
      inverse:    '{color.gray.900}',
      disabled:   '{color.gray.100}',
    },

    // Foregrounds (text)
    fg: {
      primary:    '{color.gray.900}',
      secondary:  '{color.gray.600}',
      tertiary:   '{color.gray.400}',
      inverse:    '{color.white}',
      disabled:   '{color.gray.400}',
      link:       '{color.blue.600}',
    },

    // Borders
    border: {
      default:    '{color.gray.200}',
      strong:     '{color.gray.300}',
      focus:      '{color.blue.500}',
      error:      '{color.red.500}',
      disabled:   '{color.gray.200}',
    },

    // Interactive states
    interactive: {
      default:    '{color.blue.600}',
      hover:      '{color.blue.700}',
      active:     '{color.blue.800}',
      disabled:   '{color.gray.400}',
    },

    // Feedback
    success: {
      fg:         '{color.green.700}',
      bg:         '{color.green.50}',
      border:     '{color.green.200}',
      icon:       '{color.green.600}',
    },
    warning: {
      fg:         '{color.orange.700}',
      bg:         '{color.orange.50}',
      border:     '{color.orange.200}',
      icon:       '{color.orange.600}',
    },
    danger: {
      fg:         '{color.red.700}',
      bg:         '{color.red.50}',
      border:     '{color.red.200}',
      icon:       '{color.red.600}',
    },
    info: {
      fg:         '{color.blue.700}',
      bg:         '{color.blue.50}',
      border:     '{color.blue.200}',
      icon:       '{color.blue.600}',
    },
  },

  // ------ SPACING ------
  spacing: {
    page:    { x: '{spacing.6}',  y: '{spacing.6}'  },
    section: { x: '{spacing.4}',  y: '{spacing.5}'  },
    card:    { x: '{spacing.4}',  y: '{spacing.4}'  },
    inline:  { gap: '{spacing.2}' },
    stack:   { gap: '{spacing.3}' },
  },

  // ------ TYPOGRAPHY ------
  text: {
    heading: {
      h1: { size: '{fontSize.3xl}', weight: '{fontWeight.bold}',     tracking: '{letterSpacing.tight}' },
      h2: { size: '{fontSize.2xl}', weight: '{fontWeight.semibold}', tracking: '{letterSpacing.tight}' },
      h3: { size: '{fontSize.xl}',  weight: '{fontWeight.semibold}', tracking: '{letterSpacing.normal}'},
      h4: { size: '{fontSize.lg}',  weight: '{fontWeight.medium}',   tracking: '{letterSpacing.normal}'},
    },
    body: {
      default:  { size: '{fontSize.sm}',   weight: '{fontWeight.normal}' },
      large:    { size: '{fontSize.base}',  weight: '{fontWeight.normal}' },
      small:    { size: '{fontSize.xs}',   weight: '{fontWeight.normal}' },
      emphasis: { size: '{fontSize.sm}',   weight: '{fontWeight.medium}' },
    },
    label: {
      default:  { size: '{fontSize.sm}',   weight: '{fontWeight.medium}' },
      small:    { size: '{fontSize.xs}',   weight: '{fontWeight.medium}' },
    },
    code: {
      default:  { size: '{fontSize.sm}',   family: '{fontFamily.mono}' },
    },
  },

  // ------ CONTROLS ------
  control: {
    height: {
      sm: '{spacing.8}',     // 32px — compact tables, dense UIs
      md: '{spacing.10}',    // 40px — default inputs, buttons
      lg: '{spacing.12}',    // 48px — touch targets, mobile
    },
    radius: '{radius.md}',
    border: '1px solid {color.border.default}',
    focus: {
      ring: '0 0 0 2px {color.white}, 0 0 0 4px {color.blue.500}',
    },
  },
}
```

---

### Clinical Tokens (Layer 4)

Healthcare-specific design decisions. This is what makes this library
different from every general-purpose design system.

```ts
// tokens/clinical.ts

export const clinical = {

  // ------ SEVERITY LEVELS ------
  // Used in: AllergyList, DrugInteractionAlert, CDSHooksCard, alerts
  severity: {
    critical: {
      fg:     '{color.white}',
      bg:     '{color.red.700}',
      border: '{color.red.800}',
      icon:   'alert-triangle',
      label:  'Critical',
    },
    high: {
      fg:     '{color.red.700}',
      bg:     '{color.red.50}',
      border: '{color.red.300}',
      icon:   'alert-circle',
      label:  'High',
    },
    moderate: {
      fg:     '{color.orange.700}',
      bg:     '{color.orange.50}',
      border: '{color.orange.300}',
      icon:   'alert-circle',
      label:  'Moderate',
    },
    low: {
      fg:     '{color.yellow.700}',
      bg:     '{color.yellow.50}',
      border: '{color.yellow.300}',
      icon:   'info',
      label:  'Low',
    },
    info: {
      fg:     '{color.blue.700}',
      bg:     '{color.blue.50}',
      border: '{color.blue.200}',
      icon:   'info',
      label:  'Info',
    },
  },

  //  Severity visual scale:
  //
  //  [!!!! CRITICAL ]  white on red     — anaphylaxis, code blue
  //  [!!!  HIGH     ]  red on red-50    — severe allergy, drug interaction
  //  [!!   MODERATE ]  orange on org-50 — abnormal lab, warning
  //  [!    LOW      ]  yellow on yel-50 — mildly abnormal, FYI
  //  [i    INFO     ]  blue on blue-50  — preventive care due, info

  // ------ CLINICAL STATUS ------
  // Used in: ProblemList, MedicationList, CarePlan, Task
  status: {
    active: {
      fg:     '{color.green.700}',
      bg:     '{color.green.50}',
      border: '{color.green.200}',
      dot:    '{color.green.500}',
    },
    inactive: {
      fg:     '{color.gray.500}',
      bg:     '{color.gray.50}',
      border: '{color.gray.200}',
      dot:    '{color.gray.400}',
    },
    resolved: {
      fg:     '{color.gray.500}',
      bg:     '{color.gray.50}',
      border: '{color.gray.200}',
      dot:    '{color.gray.400}',
    },
    entered_in_error: {
      fg:     '{color.red.600}',
      bg:     '{color.red.50}',
      border: '{color.red.200}',
      dot:    '{color.red.500}',
      strikethrough: true,
    },
    draft: {
      fg:     '{color.yellow.700}',
      bg:     '{color.yellow.50}',
      border: '{color.yellow.200}',
      dot:    '{color.yellow.500}',
    },
    on_hold: {
      fg:     '{color.orange.600}',
      bg:     '{color.orange.50}',
      border: '{color.orange.200}',
      dot:    '{color.orange.500}',
    },
    completed: {
      fg:     '{color.blue.600}',
      bg:     '{color.blue.50}',
      border: '{color.blue.200}',
      dot:    '{color.blue.500}',
    },
    cancelled: {
      fg:     '{color.gray.500}',
      bg:     '{color.gray.50}',
      border: '{color.gray.200}',
      dot:    '{color.gray.400}',
      strikethrough: true,
    },
  },

  //  Status badge visual:
  //
  //  (* active   )  green dot + text
  //  (* draft    )  yellow dot + text
  //  (* on-hold  )  orange dot + text
  //  (* completed)  blue dot + text
  //  (* resolved )  gray dot + text
  //  (* cancelled)  gray dot + strikethrough
  //  (* error    )  red dot + strikethrough

  // ------ LAB RESULT FLAGS ------
  // Used in: LabResults, VitalsPanel, LabSparkline
  lab: {
    normal: {
      fg:     '{color.fg.primary}',
      bg:     'transparent',
      weight: '{fontWeight.normal}',
    },
    abnormal_high: {
      fg:     '{color.red.700}',
      bg:     '{color.red.50}',
      weight: '{fontWeight.semibold}',
      flag:   'H',
    },
    abnormal_low: {
      fg:     '{color.blue.700}',
      bg:     '{color.blue.50}',
      weight: '{fontWeight.semibold}',
      flag:   'L',
    },
    critical_high: {
      fg:     '{color.white}',
      bg:     '{color.red.700}',
      weight: '{fontWeight.bold}',
      flag:   'HH',
    },
    critical_low: {
      fg:     '{color.white}',
      bg:     '{color.blue.800}',
      weight: '{fontWeight.bold}',
      flag:   'LL',
    },
  },

  //  Lab flag visual:
  //
  //  Normal value:     140 mg/dL        (plain text)
  //  High:            *145 mg/dL*  [H]  (red, bold)
  //  Low:             * 65 mg/dL*  [L]  (blue, bold)
  //  Critical high:   [*180 mg/dL* HH]  (white on red)
  //  Critical low:    [* 40 mg/dL* LL]  (white on blue)

  // ------ TASK PRIORITY ------
  // Used in: TaskCard, inbox, worklist
  priority: {
    stat: {
      fg:     '{color.white}',
      bg:     '{color.red.600}',
      label:  'STAT',
    },
    urgent: {
      fg:     '{color.orange.700}',
      bg:     '{color.orange.50}',
      label:  'Urgent',
    },
    routine: {
      fg:     '{color.blue.600}',
      bg:     '{color.blue.50}',
      label:  'Routine',
    },
    elective: {
      fg:     '{color.gray.600}',
      bg:     '{color.gray.100}',
      label:  'Elective',
    },
  },

  // ------ ENCOUNTER TYPE COLORS ------
  // Used in: ClinicalTimeline, Scheduler, EncounterSummary
  encounter: {
    office_visit:    '{color.blue.500}',
    telehealth:      '{color.teal.500}',
    emergency:       '{color.red.600}',
    inpatient:       '{color.purple.600}',
    observation:     '{color.orange.500}',
    procedure:       '{color.green.600}',
    imaging:         '{color.yellow.600}',
    lab:             '{color.teal.600}',
  },

  // ------ RESOURCE TYPE ICONS & COLORS ------
  // Used in: ClinicalTimeline, ResourceSearch, navigation
  resourceType: {
    Patient:            { icon: 'user',          color: '{color.blue.500}'   },
    Encounter:          { icon: 'clipboard',     color: '{color.blue.600}'   },
    Condition:          { icon: 'activity',       color: '{color.orange.500}' },
    Observation:        { icon: 'bar-chart-2',   color: '{color.teal.500}'   },
    MedicationRequest:  { icon: 'pill',          color: '{color.green.600}'  },
    AllergyIntolerance: { icon: 'alert-triangle',color: '{color.red.500}'    },
    Procedure:          { icon: 'scissors',      color: '{color.purple.500}' },
    DiagnosticReport:   { icon: 'file-text',     color: '{color.teal.600}'   },
    Immunization:       { icon: 'shield',        color: '{color.green.500}'  },
    CarePlan:           { icon: 'target',        color: '{color.blue.500}'   },
    Appointment:        { icon: 'calendar',      color: '{color.blue.400}'   },
    DocumentReference:  { icon: 'file',          color: '{color.gray.500}'   },
    Task:               { icon: 'check-square',  color: '{color.yellow.600}' },
    Claim:              { icon: 'dollar-sign',   color: '{color.green.700}'  },
    Composition:        { icon: 'edit-3',        color: '{color.blue.700}'   },
  },

  // ------ PATIENT BANNER ------
  // Specific tokens for the patient demographics banner
  patientBanner: {
    bg:             '{color.white}',
    border:         '{color.gray.200}',
    name: {
      size:         '{fontSize.xl}',
      weight:       '{fontWeight.semibold}',
    },
    demographics: {
      size:         '{fontSize.sm}',
      color:        '{color.fg.secondary}',
    },
    allergyBadge: {
      bg:           '{color.red.100}',
      fg:           '{color.red.800}',
      border:       '{color.red.300}',
    },
    flagBadge: {
      bg:           '{color.yellow.100}',
      fg:           '{color.yellow.800}',
      border:       '{color.yellow.300}',
    },
    avatar: {
      size:         '{spacing.16}',
      radius:       '{radius.lg}',
      bg:           '{color.blue.100}',
      fg:           '{color.blue.700}',
    },
  },
}
```

---

## Theme System

### How Theming Works

```
+-------------------------------------------------------------------+
|                        THEME FLOW                                 |
+-------------------------------------------------------------------+
|                                                                   |
|  1. Default tokens ship with the library                          |
|     (you get a beautiful default theme out of the box)            |
|                                                                   |
|  2. Developer creates a custom theme by overriding tokens         |
|     (partial overrides — only change what you need)               |
|                                                                   |
|  3. ThemeProvider merges custom tokens with defaults              |
|     (deep merge — your overrides win)                             |
|                                                                   |
|  4. Tokens become CSS custom properties on :root                  |
|     (every component reads from CSS vars)                         |
|                                                                   |
|  5. Components render using the resolved tokens                   |
|     (no hardcoded colors/spacing anywhere)                        |
|                                                                   |
+-------------------------------------------------------------------+

  Developer's custom theme:            Resolved theme:
  {                                    {
    color: {                             color: {
      primary: '#0F52BA'   ----merge-->    primary: '#0F52BA' (custom)
    },                                     danger: '#DC2626'  (default)
    clinical: {                            ...
      severity: {                        },
        critical: {                      clinical: {
          bg: '#8B0000'    ----merge-->     severity: {
        }                                    critical: {
      }                                        bg: '#8B0000'  (custom)
    }                                          fg: '#FFFFFF'  (default)
  }                                          }
                                           }
                                         }
                                       }
```

### ThemeProvider

```tsx
// Usage: Wrap your app in ThemeProvider

import { ThemeProvider, createTheme } from '@ehr/react'

// Create a custom theme (only override what you need)
const myTheme = createTheme({
  // Override global tokens
  colors: {
    blue: {
      500: '#0F52BA',     // Change primary blue
      600: '#0A3D8F',
    },
  },

  // Override semantic tokens
  semantic: {
    color: {
      interactive: {
        default: '#0F52BA',
        hover:   '#0A3D8F',
      },
    },
    control: {
      radius: '0.5rem',    // Rounder controls
    },
  },

  // Override clinical tokens
  clinical: {
    severity: {
      critical: {
        bg: '#8B0000',     // Darker critical red
      },
    },
    patientBanner: {
      bg: '#F0F4FF',       // Light blue banner background
    },
  },

  // Override typography
  typography: {
    fontFamily: {
      sans: 'Roboto, system-ui, sans-serif',
    },
  },
})

function App() {
  return (
    <ThemeProvider theme={myTheme}>
      <FHIRProvider serverUrl="https://your-ehr.com/fhir">
        <YourApp />
      </FHIRProvider>
    </ThemeProvider>
  )
}
```

### CSS Custom Properties Output

```css
/* What ThemeProvider generates on :root */

:root {
  /* Global tokens */
  --ehr-color-blue-500: #3B82F6;
  --ehr-color-red-600: #DC2626;
  --ehr-color-gray-50: #F9FAFB;
  --ehr-spacing-4: 1rem;
  --ehr-font-size-sm: 0.875rem;
  --ehr-radius-md: 0.375rem;
  --ehr-shadow-sm: 0 1px 3px 0 rgba(0,0,0,0.1);

  /* Semantic tokens */
  --ehr-color-bg-primary: var(--ehr-color-white);
  --ehr-color-fg-primary: var(--ehr-color-gray-900);
  --ehr-color-border-default: var(--ehr-color-gray-200);
  --ehr-color-interactive: var(--ehr-color-blue-600);
  --ehr-control-height-md: var(--ehr-spacing-10);
  --ehr-control-radius: var(--ehr-radius-md);

  /* Clinical tokens */
  --ehr-severity-critical-bg: var(--ehr-color-red-700);
  --ehr-severity-critical-fg: var(--ehr-color-white);
  --ehr-status-active-dot: var(--ehr-color-green-500);
  --ehr-lab-abnormal-high-fg: var(--ehr-color-red-700);
  --ehr-lab-abnormal-high-bg: var(--ehr-color-red-50);
}

/* Dark theme override — just swap the vars */
[data-theme="dark"] {
  --ehr-color-bg-primary: var(--ehr-color-gray-900);
  --ehr-color-fg-primary: var(--ehr-color-gray-50);
  --ehr-color-border-default: var(--ehr-color-gray-700);
  --ehr-color-bg-secondary: var(--ehr-color-gray-800);
}
```

### Pre-Built Themes

```
+-------------------------------------------------------------------+
|  SHIPPED THEMES                                                   |
+-------------------------------------------------------------------+
|                                                                   |
|  1. "default"      — Clean, modern, blue primary                  |
|     Light background, Inter font, rounded controls                |
|     For: General-purpose healthcare apps                          |
|                                                                   |
|  2. "clinical"     — High-contrast, dense information display     |
|     Tighter spacing, larger text for readability,                 |
|     higher contrast ratios for clinical workstations              |
|     For: EHR clinical workflows, provider-facing apps             |
|                                                                   |
|  3. "patient"      — Friendly, accessible, larger touch targets   |
|     Larger fonts, more spacing, simpler layouts,                  |
|     mobile-optimized, warm color palette                          |
|     For: Patient portals, patient-facing apps                     |
|                                                                   |
|  4. "dark"         — Dark mode for all themes                     |
|     Reduced eye strain for night shifts, ICU monitoring           |
|     For: Night mode, low-light clinical environments              |
|                                                                   |
|  5. "high-contrast" — WCAG AAA, maximum contrast                  |
|     For: Vision-impaired users, accessibility compliance          |
|                                                                   |
+-------------------------------------------------------------------+
```

```tsx
import { ThemeProvider, themes } from '@ehr/react'

// Use a pre-built theme
<ThemeProvider theme={themes.clinical}>

// Or extend one
const myTheme = createTheme({
  extends: themes.clinical,
  colors: {
    blue: { 500: '#0055B8' },  // Your brand blue
  },
})
```

### Dark Mode

```
+-------------------------------------------------------------------+
|  LIGHT vs DARK                                                    |
+-------------------------------------------------------------------+
|                                                                   |
|  LIGHT (default):                                                 |
|  +-------------------------------------------------------------+  |
|  |  bg: white                                                  |  |
|  |  fg: gray-900                                               |  |
|  |  card: white, border gray-200, shadow-sm                    |  |
|  |  input: white bg, gray-300 border                           |  |
|  |  severity.critical: red-700 bg, white fg                    |  |
|  |  lab.abnormal: red-50 bg, red-700 fg                        |  |
|  +-------------------------------------------------------------+  |
|                                                                   |
|  DARK:                                                            |
|  +-------------------------------------------------------------+  |
|  |  bg: gray-950                                               |  |
|  |  fg: gray-50                                                |  |
|  |  card: gray-900, border gray-700, no shadow                 |  |
|  |  input: gray-800 bg, gray-600 border                        |  |
|  |  severity.critical: red-600 bg, white fg                    |  |
|  |  lab.abnormal: red-950 bg, red-400 fg                       |  |
|  +-------------------------------------------------------------+  |
|                                                                   |
+-------------------------------------------------------------------+
```

```tsx
// Automatic dark mode support
<ThemeProvider theme={myTheme} colorMode="system">
  {/* Follows OS preference */}
</ThemeProvider>

<ThemeProvider theme={myTheme} colorMode="dark">
  {/* Force dark */}
</ThemeProvider>

// Programmatic toggle
const { colorMode, toggleColorMode } = useTheme()
```

---

## Core Primitives

> These are the unstyled building blocks. Every clinical component
> is built from these. They handle behavior, accessibility, keyboard
> navigation. They render semantic HTML with CSS custom properties.
> They are the "Radix" layer of the design system.

### Primitive Inventory

```
+-------------------------------------------------------------------+
|  CORE PRIMITIVES (unstyled, logic-only)                           |
+-------------------------------------------------------------------+
|                                                                   |
|  Layout:                                                          |
|  +-----+  +-------+  +------+  +--------+  +----------+          |
|  | Box |  | Stack |  | Grid |  | Inline |  | Divider  |          |
|  +-----+  +-------+  +------+  +--------+  +----------+          |
|                                                                   |
|  Typography:                                                      |
|  +------+  +---------+  +------+  +------+                       |
|  | Text |  | Heading |  | Code |  | Link |                       |
|  +------+  +---------+  +------+  +------+                       |
|                                                                   |
|  Forms:                                                           |
|  +-------+  +--------+  +--------+  +----------+                 |
|  | Input |  | Select |  | Switch |  | Checkbox |                 |
|  +-------+  +--------+  +--------+  +----------+                 |
|  +----------+  +--------+  +------+  +----------+                |
|  | Textarea |  | Radio  |  | Slider| | Combobox |                |
|  +----------+  +--------+  +------+  +----------+                |
|  +-----------+  +----------+  +-------+                           |
|  | DatePicker|  | TimePicker|  | Form  |                          |
|  +-----------+  +----------+  +-------+                           |
|                                                                   |
|  Feedback:                                                        |
|  +-------+  +-------+  +----------+  +---------+                 |
|  | Alert |  | Badge |  | Progress |  | Spinner |                 |
|  +-------+  +-------+  +----------+  +---------+                 |
|  +-------+  +----------+                                         |
|  | Toast |  | Skeleton |                                         |
|  +-------+  +----------+                                         |
|                                                                   |
|  Overlay:                                                         |
|  +--------+  +---------+  +---------+  +---------+               |
|  | Dialog |  | Popover |  | Tooltip |  | Drawer  |               |
|  +--------+  +---------+  +---------+  +---------+               |
|  +--------+  +-----------+                                       |
|  | Sheet  |  | DropdownMenu|                                     |
|  +--------+  +-----------+                                       |
|                                                                   |
|  Navigation:                                                      |
|  +------+  +--------+  +-----------+  +--------+                 |
|  | Tabs |  | Sidebar|  | Breadcrumb|  | NavBar |                 |
|  +------+  +--------+  +-----------+  +--------+                 |
|                                                                   |
|  Data Display:                                                    |
|  +-------+  +------+  +---------+  +--------+  +----------+     |
|  | Table |  | List |  | Avatar  |  | Card   |  | Accordion|     |
|  +-------+  +------+  +---------+  +--------+  +----------+     |
|  +-----------+  +--------+                                       |
|  | DataTable |  | Empty  |                                       |
|  +-----------+  +--------+                                       |
|                                                                   |
|  TOTAL: ~42 core primitives                                       |
+-------------------------------------------------------------------+
```

### Primitive Example: `<Box>`

The most fundamental primitive. Every component builds on Box.

```tsx
// Box — the atomic layout primitive
// Consumes spacing/color tokens directly via props

interface BoxProps {
  as?: React.ElementType           // Render as any HTML element
  // Spacing (maps to spacing tokens)
  p?: SpacingToken                 // padding
  px?: SpacingToken                // padding-inline
  py?: SpacingToken                // padding-block
  m?: SpacingToken                 // margin
  mx?: SpacingToken
  my?: SpacingToken
  gap?: SpacingToken               // flex/grid gap
  // Colors (maps to semantic tokens)
  bg?: ColorToken
  color?: ColorToken
  borderColor?: ColorToken
  // Layout
  display?: 'flex' | 'grid' | 'block' | 'inline' | 'inline-flex' | 'none'
  direction?: 'row' | 'column'
  align?: 'start' | 'center' | 'end' | 'stretch' | 'baseline'
  justify?: 'start' | 'center' | 'end' | 'between' | 'around'
  wrap?: boolean
  // Sizing
  w?: SizeToken | string
  h?: SizeToken | string
  minW?: SizeToken | string
  maxW?: SizeToken | string
  // Border
  border?: boolean
  rounded?: RadiusToken
  shadow?: ShadowToken
  // Responsive
  sm?: Partial<BoxProps>
  md?: Partial<BoxProps>
  lg?: Partial<BoxProps>
  // Standard HTML
  className?: string
  style?: React.CSSProperties
  children?: React.ReactNode
}

// Usage:
<Box p="4" bg="bg.primary" rounded="lg" shadow="sm" border>
  <Box as="h2" color="fg.primary" mb="2">Patient Summary</Box>
  <Box color="fg.secondary">Content here</Box>
</Box>

// Renders as:
<div
  style={{
    padding: 'var(--ehr-spacing-4)',
    backgroundColor: 'var(--ehr-color-bg-primary)',
    borderRadius: 'var(--ehr-radius-lg)',
    boxShadow: 'var(--ehr-shadow-sm)',
    border: '1px solid var(--ehr-color-border-default)',
  }}
>
  ...
</div>
```

### Primitive Example: `<Stack>`

```tsx
// Vertical stack with consistent spacing

<Stack gap="4">              // gap = spacing token
  <PatientBanner />
  <ProblemList />
  <MedicationList />
</Stack>

// Renders as:
<div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--ehr-spacing-4)' }}>
  ...
</div>
```

### Primitive Example: `<DataTable>`

The table primitive — used by LabResults, ProblemList, MedicationList, AuditLog.

```tsx
interface DataTableProps<T> {
  data: T[]
  columns: Column<T>[]
  sortable?: boolean
  filterable?: boolean
  selectable?: boolean
  paginated?: boolean
  pageSize?: number
  emptyState?: React.ReactNode
  onRowClick?: (row: T) => void
  stickyHeader?: boolean
  compact?: boolean               // Dense mode for clinical data
  striped?: boolean
  highlightRow?: (row: T) => 'danger' | 'warning' | 'success' | null
}

interface Column<T> {
  key: string
  header: string
  render?: (value: any, row: T) => React.ReactNode
  sortable?: boolean
  width?: string
  align?: 'left' | 'center' | 'right'
}
```

**ASCII — DataTable rendered:**

```
+------------------------------------------------------------------------+
|  [checkbox]  | Name             | DOB        | MRN      | Status       |
|  [ ]  all    | [sortable ^v]    | [sort ^v]  |          | [filter v]   |
|--------------|------------------|------------|----------|--------------|
|  [x]         | John Smith       | 03/15/1978 | 12345678 | (* active )  |
|  [ ]         | Jane Doe         | 11/22/1985 | 87654321 | (* active )  |
|  [ ]  >>>>>> | Bob Wilson       | 06/01/1945 | 11223344 | (* inactive) |  << highlighted row (warning: age > 75)
|  [ ]         | Alice Brown      | 09/10/2001 | 55667788 | (* active )  |
|--------------|------------------|------------|----------|--------------|
|  Showing 1-4 of 156             | [< Prev]  [1] [2] [3] ... [Next >]  |
+------------------------------------------------------------------------+
```

### Primitive Example: `<Combobox>` (Autocomplete)

The foundation for `<TerminologySearch>`, `<PatientSearch>`, `<ReferenceInput>`.

```tsx
interface ComboboxProps<T> {
  items: T[]
  onSearch: (query: string) => void
  onSelect: (item: T) => void
  renderItem: (item: T) => React.ReactNode
  renderSelected?: (item: T) => React.ReactNode
  placeholder?: string
  loading?: boolean
  multiple?: boolean              // Multi-select mode
  creatable?: boolean             // Allow creating new items
  debounce?: number               // Debounce search input (ms)
  emptyMessage?: string
  groupBy?: (item: T) => string   // Group items by category
  // Accessibility
  'aria-label': string
  'aria-describedby'?: string
}
```

**ASCII — Combobox rendered:**

```
+--------------------------------------------------+
|  [Search icon]  Type to search...           [X]  |
+--------------------------------------------------+
|                                                  |
|  GROUP: Recently Used                            |
|  +----------------------------------------------+|
|  |  [>] Essential hypertension          59621000 ||
|  |  [>] Type 2 diabetes                 E11.9   ||
|  +----------------------------------------------+|
|                                                  |
|  GROUP: Search Results                           |
|  +----------------------------------------------+|
|  |  [>] Hypertensive heart disease      I11.9   ||
|  |  [>] Hypertensive crisis             70272006||
|  |  [>] Hypertension in pregnancy       O13     ||
|  +----------------------------------------------+|
|                                                  |
|  5 results  |  Powered by ValueSet/$expand       |
+--------------------------------------------------+
```

---

## Component Architecture Patterns

### Pattern 1: Compound Components

Clinical components expose sub-components for maximum flexibility.

```tsx
// Compound component pattern — developer controls layout

<PatientBanner patient={patient}>
  <PatientBanner.Avatar />
  <PatientBanner.Name />
  <PatientBanner.Demographics />
  <PatientBanner.Identifiers />
  <PatientBanner.Allergies />
  <PatientBanner.Flags />
  <PatientBanner.Actions>
    <Button>Check In</Button>
  </PatientBanner.Actions>
</PatientBanner>

// OR use the default layout (no sub-components needed):
<PatientBanner patient={patient} />
```

**ASCII — Compound Component anatomy:**

```
+------------------------------------------------------------------------+
|  <PatientBanner>                                                       |
|  +--------+  +--------------------------------------------------+     |
|  |<Avatar>|  | <Name>         <Demographics>                     |     |
|  |        |  | John Smith     DOB: 03/15/1978 (47y)  Sex: Male   |     |
|  |  (JS)  |  +--------------------------------------------------+     |
|  |        |  | <Identifiers>                                     |     |
|  +--------+  | MRN: 12345678                                     |     |
|              +--------------------------------------------------+     |
|  <Allergies>                                                          |
|  [! Penicillin - SEVERE] [! Sulfa drugs]                              |
|                                                                        |
|  <Flags>                          <Actions>                            |
|  [FALL RISK] [DNR]                [Check In] [Message]                |
+------------------------------------------------------------------------+
```

### Pattern 2: Slots (Render Props)

Developers can replace any part of a component.

```tsx
// Slot pattern — override rendering of specific parts

<LabResults
  observations={labs}
  slots={{
    // Override how abnormal flags render
    flag: ({ flag, value }) => (
      <MyCustomBadge severity={flag}>{value}</MyCustomBadge>
    ),
    // Override the trend sparkline
    trend: ({ data }) => (
      <MyChartLibrary data={data} type="sparkline" />
    ),
    // Override empty state
    empty: () => <MyEmptyState message="No lab results found" />,
  }}
/>
```

### Pattern 3: Variants via `className` / `class` Override

Every component accepts `className` for Tailwind/CSS overrides.

```tsx
// Override styling with className (Tailwind)
<PatientBanner
  patient={patient}
  className="bg-blue-50 border-blue-200 rounded-2xl"
/>

// Override sub-component styling
<PatientBanner patient={patient}>
  <PatientBanner.Name className="text-2xl font-bold text-blue-900" />
  <PatientBanner.Allergies className="bg-red-100 p-2 rounded-lg" />
</PatientBanner>
```

### Pattern 4: Headless Hook + Styled Component

Every component exists as both a hook (logic) and a component (styled).

```tsx
// OPTION A: Use the pre-styled component
import { MedicationList } from '@ehr/react'
<MedicationList patientId="123" />

// OPTION B: Use the headless hook + your own UI
import { useMedicationList } from '@ehr/react-core'

function MyMedicationList({ patientId }) {
  const {
    medications,        // MedicationRequest[]
    loading,
    error,
    filter,             // 'active' | 'stopped' | 'all'
    setFilter,
    interactions,       // DrugInteraction[]
    sort,
    setSort,
  } = useMedicationList(patientId)

  // Render however you want with YOUR design system
  return (
    <YourCard>
      {medications.map(med => (
        <YourListItem key={med.id}>
          {med.medicationCodeableConcept?.text}
          {/* your custom rendering */}
        </YourListItem>
      ))}
    </YourCard>
  )
}
```

---

## Clinical Design Language

### Information Density Modes

Clinical UIs need to show MORE data than consumer apps. The design
system supports three density modes that adjust spacing, font size,
and layout automatically.

```
+-------------------------------------------------------------------+
|  DENSITY MODES                                                    |
+-------------------------------------------------------------------+
|                                                                   |
|  COMFORTABLE (default — patient portals, admin UIs)               |
|  +-------------------------------------------------------------+  |
|  |                                                             |  |
|  |  spacing.stack.gap = spacing.4 (16px)                       |  |
|  |  font.body = fontSize.base (16px)                           |  |
|  |  control.height = spacing.12 (48px)                         |  |
|  |  table.row.height = spacing.12 (48px)                       |  |
|  |                                                             |  |
|  |  +--------+                                                 |  |
|  |  |        |  John Smith                                     |  |
|  |  |  (JS)  |  DOB: 03/15/1978 (47y)                         |  |
|  |  |        |                                                 |  |
|  |  +--------+  MRN: 12345678                                  |  |
|  |                                                             |  |
|  +-------------------------------------------------------------+  |
|                                                                   |
|  COMPACT (clinical workflows — provider-facing EHR)               |
|  +-------------------------------------------------------------+  |
|  |  spacing.stack.gap = spacing.2 (8px)                        |  |
|  |  font.body = fontSize.sm (14px)                             |  |
|  |  control.height = spacing.9 (36px)                          |  |
|  |  table.row.height = spacing.9 (36px)                        |  |
|  |                                                             |  |
|  |  +------+                                                   |  |
|  |  | (JS) | John Smith  47M  MRN:12345678                     |  |
|  |  +------+ DOB:03/15/1978  [!PCN] [FALL]                    |  |
|  +-------------------------------------------------------------+  |
|                                                                   |
|  DENSE (ICU flowsheets, large data tables)                        |
|  +-------------------------------------------------------------+  |
|  |  spacing.stack.gap = spacing.1 (4px)                        |  |
|  |  font.body = fontSize.xs (12px)                             |  |
|  |  control.height = spacing.7 (28px)                          |  |
|  |  table.row.height = spacing.7 (28px)                        |  |
|  |                                                             |  |
|  |  (JS) John Smith 47M MRN:12345678 [!PCN][FALL]             |  |
|  +-------------------------------------------------------------+  |
|                                                                   |
+-------------------------------------------------------------------+
```

```tsx
import { DensityProvider, useDensity } from '@ehr/react'

// Set density for a section
<DensityProvider density="compact">
  <LabResults observations={labs} />
  <VitalsFlowsheet observations={vitals} />
</DensityProvider>

// Or per-component
<LabResults observations={labs} density="dense" />

// Or read density in custom components
function MyComponent() {
  const { density, spacing, fontSize } = useDensity()
  // density = 'comfortable' | 'compact' | 'dense'
}
```

### Clinical Color Semantics

```
+-------------------------------------------------------------------+
|  CLINICAL COLOR USAGE GUIDE                                       |
+-------------------------------------------------------------------+
|                                                                   |
|  These are NOT arbitrary colors. Each color has a specific        |
|  clinical meaning that clinicians rely on for quick scanning.    |
|  Deviating from these semantics causes clinical errors.          |
|                                                                   |
|  RED = Danger / Critical / Allergy / Abnormal High               |
|  +-----+  [!!! CRITICAL]  Anaphylaxis risk                       |
|  |     |  [! ALLERGY]     Penicillin - Severe                    |
|  | RED |  [HH] 180 mg/dL  Critical high lab value               |
|  |     |  [H]  145 mg/dL  Abnormal high lab value                |
|  +-----+  (* error)       Entered in error status                |
|                                                                   |
|  ORANGE = Warning / Moderate / Attention needed                   |
|  +--------+  [!! WARNING]   Drug interaction - moderate          |
|  | ORANGE |  (* on-hold)    Medication on hold                   |
|  +--------+  [! MODERATE]   Moderately abnormal                  |
|                                                                   |
|  YELLOW = Caution / Low priority / Draft                          |
|  +--------+  [! LOW]        Mildly abnormal                      |
|  | YELLOW |  (* draft)      Draft note / pending review          |
|  +--------+  [DUE]          Preventive care overdue              |
|                                                                   |
|  GREEN = Normal / Active / Success / Completed                    |
|  +-------+  (* active)     Active problem/medication              |
|  | GREEN |  Normal value   Lab within reference range             |
|  +-------+  [done]         Immunization completed                 |
|                                                                   |
|  BLUE = Info / Primary action / Selected / Completed task         |
|  +------+  [i INFO]       Informational CDS card                  |
|  | BLUE |  (* completed)  Completed task                          |
|  +------+  [L] 65 mg/dL   Abnormal low lab value                 |
|                                                                   |
|  PURPLE = Specialty / Research / Inpatient                        |
|  +--------+  Inpatient encounter color                            |
|  | PURPLE |  Research protocol indicator                          |
|  +--------+  Specialty consult marker                             |
|                                                                   |
|  TEAL = Clinical info / Telehealth / Diagnostics                  |
|  +------+  Telehealth encounter color                             |
|  | TEAL |  Lab/diagnostic result category color                   |
|  +------+  Clinical observation marker                            |
|                                                                   |
|  GRAY = Inactive / Resolved / Cancelled / Disabled                |
|  +------+  (* inactive)   Inactive problem                        |
|  | GRAY |  (* resolved)   Resolved condition                      |
|  +------+  (* cancelled)  Cancelled appointment                   |
|                                                                   |
+-------------------------------------------------------------------+
```

---

## Extending & Customizing

### Adding Custom Tokens

```tsx
// Extend the token system with your organization's tokens

const myTheme = createTheme({
  extend: {
    // Add custom colors for your organization
    colors: {
      brand: {
        50:  '#F0F7FF',
        500: '#0055B8',   // Your brand blue
        900: '#002D5F',
      },
    },
    // Add custom clinical tokens
    clinical: {
      // Custom triage colors (your ED uses different colors)
      triage: {
        immediate:  { bg: '#FF0000', fg: '#FFFFFF', label: 'Immediate' },
        emergent:   { bg: '#FF6600', fg: '#FFFFFF', label: 'Emergent' },
        urgent:     { bg: '#FFFF00', fg: '#000000', label: 'Urgent' },
        less_urgent:{ bg: '#00CC00', fg: '#FFFFFF', label: 'Less Urgent' },
        non_urgent: { bg: '#0000FF', fg: '#FFFFFF', label: 'Non-Urgent' },
      },
    },
    // Add custom component tokens
    components: {
      myCustomCard: {
        bg: '{color.brand.50}',
        border: '{color.brand.500}',
        headerBg: '{color.brand.500}',
        headerFg: '{color.white}',
      },
    },
  },
})
```

### Creating Custom Components from Primitives

```tsx
// Build a custom component using core primitives + tokens

import { Box, Stack, Badge, Text, useThemeTokens } from '@ehr/react'
import { useMedicationList } from '@ehr/react-core'

function TriageCard({ patientId, triageLevel }) {
  const tokens = useThemeTokens()
  const triage = tokens.clinical.triage[triageLevel]

  return (
    <Box
      p="4"
      rounded="lg"
      border
      style={{
        borderColor: triage.bg,
        borderLeftWidth: '4px',
      }}
    >
      <Stack gap="2">
        <Badge style={{ backgroundColor: triage.bg, color: triage.fg }}>
          {triage.label}
        </Badge>
        <PatientBanner patientId={patientId} compact />
        <MedicationList patientId={patientId} density="compact" />
      </Stack>
    </Box>
  )
}
```

### Overriding Component Defaults

```tsx
// Override defaults for all instances of a component

const myTheme = createTheme({
  components: {
    PatientBanner: {
      defaultProps: {
        showPhoto: true,
        showAllergies: true,
        showFlags: true,
        compact: false,
      },
      styles: {
        root: 'bg-blue-50 border-blue-200',
        name: 'text-xl font-bold',
        allergyBadge: 'bg-red-200 text-red-900',
      },
    },
    LabResults: {
      defaultProps: {
        showTrends: true,
        highlightAbnormal: true,
        density: 'compact',
      },
    },
    DataTable: {
      defaultProps: {
        striped: true,
        stickyHeader: true,
        pageSize: 25,
      },
    },
  },
})
```

---

## Package Map

### Final Package Architecture

```
@ehr/
  tokens/                  # Design token definitions
    src/
      colors.ts            # Color palette (8 scales x 11 steps)
      spacing.ts           # Spacing scale (0 - 96)
      typography.ts        # Font families, sizes, weights
      radius.ts            # Border radius scale
      shadow.ts            # Box shadow scale
      z-index.ts           # Z-index layers
      motion.ts            # Duration + easing curves
      breakpoints.ts       # Responsive breakpoints
      semantic.ts          # Semantic tokens (Layer 2)
      clinical.ts          # Clinical tokens (Layer 4)
      index.ts             # Exports everything
    package.json

  themes/                  # Pre-built themes
    src/
      default.ts           # Default theme
      clinical.ts          # High-density clinical theme
      patient.ts           # Patient-facing theme
      dark.ts              # Dark mode overrides
      high-contrast.ts     # WCAG AAA theme
      create-theme.ts      # Theme creation utility
      index.ts
    package.json

  primitives/              # Core UI primitives (unstyled)
    src/
      box.tsx              # Base layout primitive
      stack.tsx            # Vertical/horizontal stack
      grid.tsx             # CSS grid layout
      inline.tsx           # Inline flex layout
      text.tsx             # Text rendering
      heading.tsx          # Heading levels
      button.tsx           # Button (all variants)
      input.tsx            # Text input
      select.tsx           # Select / dropdown
      checkbox.tsx         # Checkbox
      radio.tsx            # Radio group
      switch.tsx           # Toggle switch
      textarea.tsx         # Multi-line input
      combobox.tsx         # Autocomplete/search input
      date-picker.tsx      # Date selection
      time-picker.tsx      # Time selection
      slider.tsx           # Range slider
      dialog.tsx           # Modal dialog
      drawer.tsx           # Side drawer
      popover.tsx          # Popover
      tooltip.tsx          # Tooltip
      dropdown-menu.tsx    # Context/dropdown menu
      tabs.tsx             # Tab navigation
      accordion.tsx        # Collapsible sections
      table.tsx            # Base table
      data-table.tsx       # Sortable/filterable table
      badge.tsx            # Badge/chip
      alert.tsx            # Alert banner
      toast.tsx            # Toast notification
      progress.tsx         # Progress bar
      spinner.tsx          # Loading spinner
      skeleton.tsx         # Loading skeleton
      avatar.tsx           # User/resource avatar
      card.tsx             # Card container
      divider.tsx          # Horizontal/vertical divider
      breadcrumb.tsx       # Breadcrumb navigation
      sidebar.tsx          # Sidebar navigation
      form.tsx             # Form wrapper with validation
      form-field.tsx       # Form field with label/error
      empty-state.tsx      # Empty state placeholder
      list.tsx             # List view
      visually-hidden.tsx  # Screen reader only content
      focus-trap.tsx       # Focus trap for modals
      index.ts             # ~42 primitives
    package.json

  fhir-types/              # FHIR R4 TypeScript definitions
  fhir-hooks/              # React hooks for FHIR operations
  react-core/              # Headless clinical components (logic only)
  react/                   # Pre-styled clinical components (Tailwind)
  smart-auth/              # SMART on FHIR authentication
```

### Dependency Graph

```
                    @ehr/tokens
                        |
              +---------+---------+
              |                   |
         @ehr/themes        @ehr/fhir-types
              |                   |
              v                   v
        @ehr/primitives     @ehr/fhir-hooks
              |                   |
              +---------+---------+
                        |
                  @ehr/react-core
                        |
                  @ehr/react         @ehr/smart-auth
                  (pre-styled)       (auth layer)
```

### What Ships to npm

```
+-------------------------------------------------------------------+
| Package            | Purpose              | Size    | Dependencies |
|--------------------|----------------------|---------|--------------|
| @ehr/tokens        | Token definitions     | ~8 KB   | none         |
| @ehr/themes        | Theme presets         | ~12 KB  | tokens       |
| @ehr/primitives    | 42 unstyled components| ~45 KB  | tokens       |
| @ehr/fhir-types    | FHIR R4 TS types     | ~80 KB  | none         |
| @ehr/fhir-hooks    | 20 React hooks        | ~25 KB  | fhir-types   |
| @ehr/react-core    | Headless clinical     | ~60 KB  | primitives,  |
|                    | components            |         | fhir-hooks   |
| @ehr/react         | Pre-styled components | ~100 KB | react-core,  |
|                    |                      |         | themes       |
| @ehr/smart-auth    | SMART on FHIR auth   | ~15 KB  | fhir-hooks   |
|--------------------|----------------------|---------|--------------|
| TOTAL (all)        |                      | ~345 KB | tree-shaken  |
| Typical app import |                      | ~80 KB  | after shake  |
+-------------------------------------------------------------------+
```

---

## Summary

```
+===================================================================+
|                                                                   |
|   This is not a component library.                                |
|   This is a design system.                                        |
|                                                                   |
|   Tokens  -->  Themes  -->  Primitives  -->  Components           |
|                                                                   |
|   Developers don't consume our opinions.                          |
|   They compose from our foundations.                              |
|                                                                   |
|   The core handles:                                               |
|     - FHIR data complexity (types, hooks, reference resolution)   |
|     - Clinical semantics (severity, status, lab flags, priority)  |
|     - Accessibility (ARIA, keyboard, screen reader, contrast)     |
|     - Theming (tokens, CSS vars, dark mode, density)              |
|     - Behavior (compound components, slots, headless hooks)       |
|                                                                   |
|   Developers add:                                                 |
|     - Their brand colors, fonts, spacing preferences              |
|     - Their design system integration (Tailwind, MUI, custom)     |
|     - Their custom clinical tokens (triage colors, org rules)     |
|     - Their custom components built from our primitives           |
|     - Their layouts, workflows, and views                         |
|                                                                   |
|   The result: Every healthcare app looks different.               |
|   But every healthcare app works the same way underneath.         |
|   That's what makes it an industry standard.                      |
|                                                                   |
+===================================================================+
```
