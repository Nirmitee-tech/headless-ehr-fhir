# @ehr/primitives

Core UI primitives for building healthcare interfaces. Provides foundational layout and typography components that use design tokens from `@ehr/tokens`.

## Installation

```bash
pnpm add @ehr/primitives
```

## Setup

Import the token CSS and (optionally) the primitives base styles at your app root:

```tsx
import '@ehr/tokens/css'
import '@ehr/primitives/css'
```

## Components

### Box

The foundational layout primitive. Renders a `<div>` by default with polymorphic `as` prop.

```tsx
import { Box } from '@ehr/primitives'

<Box>Basic div</Box>
<Box as="section" className="my-section">Rendered as section</Box>
<Box as="nav" aria-label="Main navigation">Nav element</Box>
```

**Props:**

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `as` | `ElementType` | `'div'` | HTML element to render |
| `className` | `string` | — | CSS class name |
| `ref` | `Ref<HTMLElement>` | — | Forwarded ref |
| *...rest* | `HTMLAttributes` | — | All native HTML attributes |

### Stack

Vertical or horizontal flex layout with consistent token-based spacing.

```tsx
import { Stack } from '@ehr/primitives'

{/* Vertical stack (default) */}
<Stack gap="6">
  <div>Item 1</div>
  <div>Item 2</div>
</Stack>

{/* Horizontal with alignment */}
<Stack direction="horizontal" gap="4" align="center" justify="between">
  <span>Left</span>
  <span>Right</span>
</Stack>

{/* Wrapping grid-like layout */}
<Stack direction="horizontal" gap="3" wrap>
  {tags.map(tag => <Badge key={tag}>{tag}</Badge>)}
</Stack>
```

**Props:**

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `direction` | `'vertical' \| 'horizontal'` | `'vertical'` | Flex direction |
| `gap` | `'1'-'12'` | `'4'` | Gap using `--ehr-space-{n}` token |
| `align` | `'start' \| 'center' \| 'end' \| 'stretch' \| 'baseline'` | — | Cross-axis alignment |
| `justify` | `'start' \| 'center' \| 'end' \| 'between' \| 'around'` | — | Main-axis alignment |
| `wrap` | `boolean` | `false` | Enable flex wrap |
| `className` | `string` | — | CSS class name |
| `ref` | `Ref<HTMLDivElement>` | — | Forwarded ref |

### Text

Typographic primitive with token-based sizing, weight, and color.

```tsx
import { Text } from '@ehr/primitives'

<Text>Default span text</Text>
<Text as="h1" size="3xl" weight="bold">Page Title</Text>
<Text size="sm" color="secondary">Muted helper text</Text>
<Text truncate>This very long text will be truncated with an ellipsis...</Text>
```

**Props:**

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `as` | `ElementType` | `'span'` | HTML element to render |
| `size` | `'xs' \| 'sm' \| 'base' \| 'lg' \| 'xl' \| '2xl' \| '3xl'` | — | Font size token |
| `weight` | `'normal' \| 'medium' \| 'semibold' \| 'bold'` | — | Font weight token |
| `color` | `'primary' \| 'secondary' \| 'tertiary' \| 'inverse' \| 'link' \| 'inherit'` | — | Semantic text color |
| `truncate` | `boolean` | `false` | Truncate with ellipsis |
| `className` | `string` | — | CSS class name |
| `ref` | `Ref<HTMLElement>` | — | Forwarded ref |

### Badge

Status indicator chip using clinical semantic tokens. Renders with `role="status"` for accessibility.

```tsx
import { Badge } from '@ehr/primitives'

<Badge>Default</Badge>
<Badge variant="success" dot>Active</Badge>
<Badge variant="danger">Critical Allergy</Badge>
<Badge variant="warning">Pending Review</Badge>
<Badge variant="info" size="md">Lab Result</Badge>
```

**Props:**

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `variant` | `'default' \| 'success' \| 'warning' \| 'danger' \| 'info'` | `'default'` | Visual variant |
| `size` | `'sm' \| 'md'` | `'sm'` | Badge size |
| `dot` | `boolean` | `false` | Show status dot indicator |
| `className` | `string` | — | CSS class name |
| `ref` | `Ref<HTMLSpanElement>` | — | Forwarded ref |

## Data Attributes

Every primitive renders a `data-ehr` attribute for CSS targeting without specificity wars:

```css
[data-ehr="box"]   { /* target all Box components */ }
[data-ehr="stack"] { /* target all Stack components */ }
[data-ehr="text"]  { /* target all Text components */ }
[data-ehr="badge"] { /* target all Badge components */ }

[data-ehr="badge"][data-variant="danger"] {
  /* target only danger badges */
}
```

## Accessibility

- All components forward `ref` and accept `aria-*` attributes
- Badge renders with `role="status"` for screen reader announcements
- Badge dot uses `aria-hidden="true"` (decorative only)
- Focus ring via `[data-ehr]:focus-visible` using `--ehr-focus-ring`
- Reduced motion: all transitions/animations disabled when `prefers-reduced-motion: reduce`

## Testing

50 tests cover rendering, props, data attributes, styling, HTML pass-through, and axe accessibility audits.

```bash
pnpm test
```

## License

Apache-2.0
