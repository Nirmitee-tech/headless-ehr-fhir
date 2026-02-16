import type { Meta, StoryObj } from '@storybook/react'
import { Box } from './box'

const meta: Meta<typeof Box> = {
  title: 'Primitives/Box',
  component: Box,
  tags: ['autodocs'],
}

export default meta
type Story = StoryObj<typeof Box>

export const Default: Story = {
  args: {
    children: 'A basic box',
    style: { padding: 'var(--ehr-space-4)', background: 'var(--ehr-bg-subtle)' },
  },
}

export const AsSection: Story = {
  args: {
    as: 'section',
    children: 'Rendered as a <section>',
    style: { padding: 'var(--ehr-space-4)', border: '1px solid var(--ehr-border-default)' },
  },
}

export const WithClassName: Story = {
  args: {
    className: 'custom-box',
    children: 'Box with custom class',
  },
}
