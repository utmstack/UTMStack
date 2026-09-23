import { useState } from 'react'
import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'
import type { FlowNode } from '../types/soar.types'
import { MailParamsEditor } from './MailParamsEditor'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: { address?: string }) =>
      key === 'soar.editor.canvas.mail.addressInvalid'
        ? `Invalid email address: ${options?.address}`
        : key,
  }),
}))

const nodes: Record<string, FlowNode> = {
  mail: { kind: 'executor', executor: 'mail' },
}

function MailHarness() {
  const [params, setParams] = useState({ to: '', cc: '', subject: '', body: '' })
  return (
    <MailParamsEditor
      nodeId="mail"
      nodes={nodes}
      params={params}
      onChange={(next) => setParams({ ...params, ...next })}
    />
  )
}

describe('MailParamsEditor recipient validation', () => {
  test('shows errors for malformed recipients and clears them when corrected', () => {
    render(<MailHarness />)

    const to = screen.getByPlaceholderText('alice@example.com, bob@example.com')
    const cc = screen.getByPlaceholderText('carol@example.com')

    expect(to).toHaveAttribute('aria-invalid', 'false')
    expect(cc).toHaveAttribute('aria-invalid', 'false')

    fireEvent.change(to, { target: { value: 'alice@example.com, invalid-address' } })
    fireEvent.change(cc, { target: { value: 'invalid-cc' } })

    expect(to).toHaveAttribute('aria-invalid', 'true')
    expect(cc).toHaveAttribute('aria-invalid', 'true')
    expect(screen.getByText('Invalid email address: invalid-address')).toBeInTheDocument()
    expect(screen.getByText('Invalid email address: invalid-cc')).toBeInTheDocument()

    fireEvent.change(to, { target: { value: 'alice@example.com, bob@example.com' } })
    fireEvent.change(cc, { target: { value: 'carol@example.com' } })

    expect(to).toHaveAttribute('aria-invalid', 'false')
    expect(cc).toHaveAttribute('aria-invalid', 'false')
    expect(screen.queryByText('Invalid email address: invalid-address')).not.toBeInTheDocument()
    expect(screen.queryByText('Invalid email address: invalid-cc')).not.toBeInTheDocument()
  })
})