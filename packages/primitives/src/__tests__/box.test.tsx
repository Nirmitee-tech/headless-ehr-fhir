import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { axe } from 'vitest-axe'
import { Box } from '../box'

describe('Box', () => {
  // ── Rendering ──────────────────────────────────────────

  it('renders children', () => {
    render(<Box>Hello world</Box>)
    expect(screen.getByText('Hello world')).toBeInTheDocument()
  })

  it('renders as a div by default', () => {
    const { container } = render(<Box>content</Box>)
    expect(container.firstChild?.nodeName).toBe('DIV')
  })

  it('renders as a different element with "as" prop', () => {
    const { container } = render(<Box as="section">content</Box>)
    expect(container.firstChild?.nodeName).toBe('SECTION')
  })

  it('renders as a span', () => {
    const { container } = render(<Box as="span">inline</Box>)
    expect(container.firstChild?.nodeName).toBe('SPAN')
  })

  it('renders as an article', () => {
    const { container } = render(<Box as="article">article content</Box>)
    expect(container.firstChild?.nodeName).toBe('ARTICLE')
  })

  // ── Data Attributes ────────────────────────────────────

  it('renders with data-ehr="box" for CSS targeting', () => {
    const { container } = render(<Box>content</Box>)
    expect(container.querySelector('[data-ehr="box"]')).toBeInTheDocument()
  })

  // ── Styling ────────────────────────────────────────────

  it('accepts className prop', () => {
    const { container } = render(<Box className="custom-class">content</Box>)
    expect(container.firstChild).toHaveClass('custom-class')
  })

  it('accepts style prop', () => {
    const { container } = render(<Box style={{ backgroundColor: 'red' }}>content</Box>)
    expect((container.firstChild as HTMLElement).style.backgroundColor).toBe('red')
  })

  it('merges className with data-ehr', () => {
    const { container } = render(<Box className="my-box">content</Box>)
    const el = container.firstChild as HTMLElement
    expect(el.getAttribute('data-ehr')).toBe('box')
    expect(el.classList.contains('my-box')).toBe(true)
  })

  // ── HTML Attributes ────────────────────────────────────

  it('passes through HTML attributes', () => {
    render(<Box id="test-id" data-testid="my-box" aria-label="test box">content</Box>)
    const el = screen.getByTestId('my-box')
    expect(el).toHaveAttribute('id', 'test-id')
    expect(el).toHaveAttribute('aria-label', 'test box')
  })

  it('passes through onClick handler', () => {
    let clicked = false
    render(<Box onClick={() => { clicked = true }}>click me</Box>)
    screen.getByText('click me').click()
    expect(clicked).toBe(true)
  })

  // ── Edge Cases ─────────────────────────────────────────

  it('renders with no children', () => {
    const { container } = render(<Box />)
    expect(container.firstChild).toBeInTheDocument()
    expect(container.firstChild?.textContent).toBe('')
  })

  it('renders with multiple children', () => {
    render(
      <Box>
        <span>first</span>
        <span>second</span>
      </Box>,
    )
    expect(screen.getByText('first')).toBeInTheDocument()
    expect(screen.getByText('second')).toBeInTheDocument()
  })

  // ── Accessibility ──────────────────────────────────────

  it('has no accessibility violations', async () => {
    const { container } = render(<Box>accessible content</Box>)
    const results = await axe(container)
    expect(results).toHaveNoViolations()
  })

  it('has no a11y violations when rendered as nav with aria-label', async () => {
    const { container } = render(<Box as="nav" aria-label="Main navigation">nav content</Box>)
    const results = await axe(container)
    expect(results).toHaveNoViolations()
  })
})
