import { describe, expect, it, vi } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useBackdropDismiss } from './useBackdropDismiss'
import type { MouseEvent } from 'react'

const backdrop = {} as EventTarget
const inner = {} as EventTarget

function fire(handlers: ReturnType<typeof useBackdropDismiss>, type: 'onMouseDown' | 'onMouseUp', target: EventTarget) {
  handlers[type]({ target, currentTarget: backdrop } as unknown as MouseEvent)
}

describe('useBackdropDismiss', () => {
  it('closes on a genuine backdrop click (mousedown and mouseup both on the backdrop)', () => {
    const onClose = vi.fn()
    const { result } = renderHook(() => useBackdropDismiss(onClose))
    fire(result.current, 'onMouseDown', backdrop)
    fire(result.current, 'onMouseUp', backdrop)
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('does not close when a text-selection drag starts inside the modal and is released on the backdrop', () => {
    const onClose = vi.fn()
    const { result } = renderHook(() => useBackdropDismiss(onClose))
    fire(result.current, 'onMouseDown', inner) // drag starts inside the modal card
    fire(result.current, 'onMouseUp', backdrop) // released outside it
    expect(onClose).not.toHaveBeenCalled()
  })

  it('does not close when the drag starts on the backdrop but ends inside the modal', () => {
    const onClose = vi.fn()
    const { result } = renderHook(() => useBackdropDismiss(onClose))
    fire(result.current, 'onMouseDown', backdrop)
    fire(result.current, 'onMouseUp', inner)
    expect(onClose).not.toHaveBeenCalled()
  })
})
