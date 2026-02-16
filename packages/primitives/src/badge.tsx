import { forwardRef, type ReactNode, type HTMLAttributes } from 'react'

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  /** Visual variant */
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info'
  /** Size */
  size?: 'sm' | 'md'
  /** Show a status dot before the label */
  dot?: boolean
  className?: string
  children?: ReactNode
}

/**
 * Badge — status indicator chip using clinical semantic tokens.
 *
 * Used in: StatusBadge, AllergyList, ProblemList, TaskCard.
 */
export const Badge = forwardRef<HTMLSpanElement, BadgeProps>(function Badge(
  { variant = 'default', size = 'sm', dot, className, children, style, ...rest },
  ref,
) {
  const variantStyles: Record<string, { color: string; bg: string; border: string }> = {
    default: {
      color: 'var(--ehr-fg-secondary)',
      bg: 'var(--ehr-bg-secondary)',
      border: 'var(--ehr-border-default)',
    },
    success: {
      color: 'var(--ehr-success-fg)',
      bg: 'var(--ehr-success-bg)',
      border: 'var(--ehr-success-border)',
    },
    warning: {
      color: 'var(--ehr-warning-fg)',
      bg: 'var(--ehr-warning-bg)',
      border: 'var(--ehr-warning-border)',
    },
    danger: {
      color: 'var(--ehr-danger-fg)',
      bg: 'var(--ehr-danger-bg)',
      border: 'var(--ehr-danger-border)',
    },
    info: {
      color: 'var(--ehr-info-fg)',
      bg: 'var(--ehr-info-bg)',
      border: 'var(--ehr-info-border)',
    },
  }

  const v = variantStyles[variant]
  const sizeStyles = size === 'sm'
    ? { fontSize: 'var(--ehr-text-xs)', padding: '0.125rem 0.5rem' }
    : { fontSize: 'var(--ehr-text-sm)', padding: '0.25rem 0.625rem' }

  return (
    <span
      ref={ref}
      data-ehr="badge"
      data-variant={variant}
      role="status"
      className={className}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '0.375rem',
        fontFamily: 'var(--ehr-font-sans)',
        fontWeight: 'var(--ehr-weight-medium)' as any,
        lineHeight: '1',
        borderRadius: 'var(--ehr-radius-full)',
        border: `1px solid ${v.border}`,
        color: v.color,
        backgroundColor: v.bg,
        whiteSpace: 'nowrap',
        ...sizeStyles,
        ...style,
      }}
      {...rest}
    >
      {dot && (
        <span
          data-ehr="badge-dot"
          style={{
            width: '0.375rem',
            height: '0.375rem',
            borderRadius: '50%',
            backgroundColor: 'currentColor',
            flexShrink: 0,
          }}
          aria-hidden="true"
        />
      )}
      {children}
    </span>
  )
})
