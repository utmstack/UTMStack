import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'
import type { FlowNode } from '../types/soar.types'
import { NodeInspector } from './NodeInspector'

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: { address?: string }) =>
      key === 'soar.editor.canvas.mail.addressInvalid'
        ? `Invalid email address: ${options?.address}`
        : key,
  }),
}))

describe('NodeInspector mail validation', () => {
  test('does not save a mail node with an invalid recipient', () => {
    const node: FlowNode = {
      kind: 'executor',
      executor: 'mail',
      params: { to: 'invalid-address', cc: '', subject: 'Alert', body: '' },
    }
    const onSave = vi.fn(() => true)

    render(
      <NodeInspector
        nodeId="mail"
        node={node}
        nodes={{ mail: node }}
        onSave={onSave}
        onDelete={vi.fn()}
        onClose={vi.fn()}
      />,
    )

    const saveButtons = screen.getAllByRole('button', { name: 'soar.editor.save' })
    expect(saveButtons).toHaveLength(2)
    saveButtons.forEach((save) => {
      expect(save).toBeDisabled()
      fireEvent.click(save)
    })
    expect(onSave).not.toHaveBeenCalled()
  })
})