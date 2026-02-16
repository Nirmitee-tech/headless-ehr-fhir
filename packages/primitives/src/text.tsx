import { forwardRef, type ElementType, type ReactNode, type HTMLAttributes } from 'react'

export interface TextProps extends HTMLAttributes<HTMLElement> {
  /** Render as a different HTML element */
  as?: ElementType
  /** Text size — maps to --ehr-text-* token */
  size?: 'xs' | 'sm' | 'base' | 'lg' | 'xl' | '2xl' | '3xl'
  /** Font weight — maps to --ehr-weight-* token */
  weight?: 'normal' | 'medium' | 'semibold' | 'bold'
  /** Text color — maps to --ehr-fg-* token */
  color?: 'primary' | 'secondary' | 'tertiary' | 'inverse' | 'link' | 'inherit'
  /** Truncate with ellipsis */
  truncate?: boolean
  className?: string
  children?: ReactNode
}

/**
 * Text — typographic primitive using design tokens.
 *
 * Renders a `<span>` by default. Use `as` for semantic elements.
 */
export const Text = forwardRef<HTMLElement, TextProps>(function Text(
  { as: Component = 'span', size, weight, color, truncate, className, style, children, ...rest },
  ref,
) {
  const colorMap: Record<string, string> = {
    primary: 'var(--ehr-fg-primary)',
    secondary: 'var(--ehr-fg-secondary)',
    tertiary: 'var(--ehr-fg-tertiary)',
    inverse: 'var(--ehr-fg-inverse)',
    link: 'var(--ehr-fg-link)',
    inherit: 'inherit',
  }

  return (
    <Component
      ref={ref}
      data-ehr="text"
      className={className}
      style={{
        fontSize: size ? `var(--ehr-text-${size})` : undefined,
        fontWeight: weight ? `var(--ehr-weight-${weight})` : undefined,
        color: color ? colorMap[color] : undefined,
        ...(truncate
          ? { overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' as const }
          : {}),
        ...style,
      }}
      {...rest}
    >
      {children}
    </Component>
  )
})
