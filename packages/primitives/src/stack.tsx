import { forwardRef, type ReactNode, type HTMLAttributes } from 'react'

export interface StackProps extends HTMLAttributes<HTMLDivElement> {
  /** Stack direction */
  direction?: 'vertical' | 'horizontal'
  /** Gap between items — maps to CSS custom property */
  gap?: '1' | '2' | '3' | '4' | '5' | '6' | '8' | '10' | '12'
  /** Align items on the cross axis */
  align?: 'start' | 'center' | 'end' | 'stretch' | 'baseline'
  /** Justify items on the main axis */
  justify?: 'start' | 'center' | 'end' | 'between' | 'around'
  /** Wrap items */
  wrap?: boolean
  /** Additional CSS class names */
  className?: string
  children?: ReactNode
}

/**
 * Stack — vertical or horizontal flex layout with consistent spacing.
 *
 * Uses CSS custom properties for gap values (--ehr-space-*).
 */
export const Stack = forwardRef<HTMLDivElement, StackProps>(function Stack(
  { direction = 'vertical', gap = '4', align, justify, wrap, className, style, children, ...rest },
  ref,
) {
  const justifyMap: Record<string, string> = {
    start: 'flex-start',
    center: 'center',
    end: 'flex-end',
    between: 'space-between',
    around: 'space-around',
  }

  const alignMap: Record<string, string> = {
    start: 'flex-start',
    center: 'center',
    end: 'flex-end',
    stretch: 'stretch',
    baseline: 'baseline',
  }

  return (
    <div
      ref={ref}
      data-ehr="stack"
      data-direction={direction}
      className={className}
      style={{
        display: 'flex',
        flexDirection: direction === 'horizontal' ? 'row' : 'column',
        gap: `var(--ehr-space-${gap})`,
        alignItems: align ? alignMap[align] : undefined,
        justifyContent: justify ? justifyMap[justify] : undefined,
        flexWrap: wrap ? 'wrap' : undefined,
        ...style,
      }}
      {...rest}
    >
      {children}
    </div>
  )
})
