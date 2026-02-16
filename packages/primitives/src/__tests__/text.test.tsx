import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { axe } from 'vitest-axe'
import { Text } from '../text'

describe('Text', () => {
  it('renders children', () => {
    render(<Text>Hello world</Text>)
    expect(screen.getByText('Hello world')).toBeInTheDocument()
  })

  it('renders as span by default', () => {
    const { container } = render(<Text>content</Text>)
    expect(container.firstChild?.nodeName).toBe('SPAN')
  })

  it('renders as a different element with "as" prop', () => {
    const { container } = render(<Text as="p">paragraph</Text>)
    expect(container.firstChild?.nodeName).toBe('P')
  })

  it('renders with data-ehr="text"', () => {
    const { container } = render(<Text>content</Text>)
    expect(container.querySelector('[data-ehr="text"]')).toBeInTheDocument()
  })

  it('applies size using CSS custom property', () => {
    const { container } = render(<Text size="lg">large</Text>)
    expect(container.firstChild).toHaveStyle({ fontSize: 'var(--ehr-text-lg)' })
  })

  it('applies weight using CSS custom property', () => {
    const { container } = render(<Text weight="bold">bold</Text>)
    expect(container.firstChild).toHaveStyle({ fontWeight: 'var(--ehr-weight-bold)' })
  })

  it('applies color using semantic token', () => {
    const { container } = render(<Text color="secondary">muted</Text>)
    expect(container.firstChild).toHaveStyle({ color: 'var(--ehr-fg-secondary)' })
  })

  it('applies truncation styles', () => {
    const { container } = render(<Text truncate>long text here</Text>)
    expect(container.firstChild).toHaveStyle({
      overflow: 'hidden',
      textOverflow: 'ellipsis',
      whiteSpace: 'nowrap',
    })
  })

  it('accepts className', () => {
    const { container } = render(<Text className="custom">text</Text>)
    expect(container.firstChild).toHaveClass('custom')
  })

  it('passes through HTML attributes', () => {
    render(<Text data-testid="my-text" id="test-id">content</Text>)
    const el = screen.getByTestId('my-text')
    expect(el).toHaveAttribute('id', 'test-id')
  })

  it('has no accessibility violations', async () => {
    const { container } = render(<Text>accessible text</Text>)
    const results = await axe(container)
    expect(results).toHaveNoViolations()
  })
})
