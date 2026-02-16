import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { axe } from 'vitest-axe'
import { Stack } from '../stack'

describe('Stack', () => {
  it('renders children', () => {
    render(
      <Stack>
        <div>Item 1</div>
        <div>Item 2</div>
      </Stack>,
    )
    expect(screen.getByText('Item 1')).toBeInTheDocument()
    expect(screen.getByText('Item 2')).toBeInTheDocument()
  })

  it('renders with data-ehr="stack"', () => {
    const { container } = render(<Stack>content</Stack>)
    expect(container.querySelector('[data-ehr="stack"]')).toBeInTheDocument()
  })

  it('defaults to vertical direction', () => {
    const { container } = render(<Stack>content</Stack>)
    expect(container.firstChild).toHaveStyle({ flexDirection: 'column' })
  })

  it('renders horizontal when specified', () => {
    const { container } = render(<Stack direction="horizontal">content</Stack>)
    expect(container.firstChild).toHaveStyle({ flexDirection: 'row' })
  })

  it('sets data-direction attribute', () => {
    const { container } = render(<Stack direction="horizontal">content</Stack>)
    expect(container.firstChild).toHaveAttribute('data-direction', 'horizontal')
  })

  it('applies gap using CSS custom property', () => {
    const { container } = render(<Stack gap="6">content</Stack>)
    expect(container.firstChild).toHaveStyle({ gap: 'var(--ehr-space-6)' })
  })

  it('defaults to gap 4', () => {
    const { container } = render(<Stack>content</Stack>)
    expect(container.firstChild).toHaveStyle({ gap: 'var(--ehr-space-4)' })
  })

  it('applies align prop', () => {
    const { container } = render(<Stack align="center">content</Stack>)
    expect(container.firstChild).toHaveStyle({ alignItems: 'center' })
  })

  it('applies justify prop', () => {
    const { container } = render(<Stack justify="between">content</Stack>)
    expect(container.firstChild).toHaveStyle({ justifyContent: 'space-between' })
  })

  it('applies wrap prop', () => {
    const { container } = render(<Stack wrap>content</Stack>)
    expect(container.firstChild).toHaveStyle({ flexWrap: 'wrap' })
  })

  it('accepts className', () => {
    const { container } = render(<Stack className="custom">content</Stack>)
    expect(container.firstChild).toHaveClass('custom')
  })

  it('has no accessibility violations', async () => {
    const { container } = render(
      <Stack>
        <div>Item 1</div>
        <div>Item 2</div>
      </Stack>,
    )
    const results = await axe(container)
    expect(results).toHaveNoViolations()
  })
})
