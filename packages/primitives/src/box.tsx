import { forwardRef, type ElementType, type ReactNode } from 'react'

type BoxOwnProps<C extends ElementType = 'div'> = {
  /** Render as a different HTML element */
  as?: C
  /** Additional CSS class names */
  className?: string
  /** Child elements */
  children?: ReactNode
}

type BoxProps<C extends ElementType = 'div'> = BoxOwnProps<C> &
  Omit<React.ComponentPropsWithoutRef<C>, keyof BoxOwnProps>

/**
 * Box — the foundational layout primitive.
 *
 * Renders a `div` by default. Use the `as` prop to render as any HTML element.
 * Applies `data-ehr="box"` for CSS targeting via design tokens.
 * Accepts all native HTML attributes of the rendered element.
 */
export const Box = forwardRef<HTMLElement, BoxProps>(
  function Box({ as, className, children, ...rest }, ref) {
    const Component = (as || 'div') as ElementType

    return (
      <Component
        ref={ref}
        data-ehr="box"
        className={className}
        {...rest}
      >
        {children}
      </Component>
    )
  },
)

export type { BoxProps }
