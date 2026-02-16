import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { axe } from 'vitest-axe'
import { Badge } from '../badge'

describe('Badge', () => {
  it('renders children text', () => {
    render(<Badge>Active</Badge>)
    expect(screen.getByText('Active')).toBeInTheDocument()
  })

  it('renders with data-ehr="badge"', () => {
    const { container } = render(<Badge>test</Badge>)
    expect(container.querySelector('[data-ehr="badge"]')).toBeInTheDocument()
  })

  it('renders with role="status"', () => {
    render(<Badge>Active</Badge>)
    expect(screen.getByRole('status')).toBeInTheDocument()
  })

  it('sets data-variant attribute', () => {
    const { container } = render(<Badge variant="danger">Critical</Badge>)
    expect(container.firstChild).toHaveAttribute('data-variant', 'danger')
  })

  it('defaults to variant "default"', () => {
    const { container } = render(<Badge>Default</Badge>)
    expect(container.firstChild).toHaveAttribute('data-variant', 'default')
  })

  it('renders dot indicator when dot prop is true', () => {
    const { container } = render(<Badge dot>Active</Badge>)
    expect(container.querySelector('[data-ehr="badge-dot"]')).toBeInTheDocument()
  })

  it('hides dot from screen readers', () => {
    const { container } = render(<Badge dot>Active</Badge>)
    const dot = container.querySelector('[data-ehr="badge-dot"]')
    expect(dot).toHaveAttribute('aria-hidden', 'true')
  })

  it('does not render dot by default', () => {
    const { container } = render(<Badge>Active</Badge>)
    expect(container.querySelector('[data-ehr="badge-dot"]')).not.toBeInTheDocument()
  })

  it('accepts className', () => {
    const { container } = render(<Badge className="custom">test</Badge>)
    expect(container.firstChild).toHaveClass('custom')
  })

  it('renders all variant types', () => {
    const variants = ['default', 'success', 'warning', 'danger', 'info'] as const
    for (const variant of variants) {
      const { container } = render(<Badge variant={variant}>{variant}</Badge>)
      expect(container.querySelector(`[data-variant="${variant}"]`)).toBeInTheDocument()
    }
  })

  it('has no accessibility violations', async () => {
    const { container } = render(<Badge variant="success" dot>Active</Badge>)
    const results = await axe(container)
    expect(results).toHaveNoViolations()
  })

  it('has no a11y violations for danger variant', async () => {
    const { container } = render(<Badge variant="danger">Critical Allergy</Badge>)
    const results = await axe(container)
    expect(results).toHaveNoViolations()
  })
})
