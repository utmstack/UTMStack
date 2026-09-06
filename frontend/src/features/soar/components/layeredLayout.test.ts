import { describe, expect, test } from 'vitest'
import { computeLayeredLayout, TRIGGER_LAYOUT_ID } from './layeredLayout'
import type { FlowNode } from '../types/soar.types'

const exec = (patch: Partial<FlowNode> = {}): FlowNode => ({ kind: 'executor', executor: 'noop', ...patch })

describe('computeLayeredLayout', () => {
  test('trigger is layer 0 and sits above roots', () => {
    const pos = computeLayeredLayout(['a'], { a: exec() })
    expect(pos[TRIGGER_LAYOUT_ID].y).toBeLessThan(pos.a.y)
  })

  test('each child sits one layer below its parent', () => {
    const pos = computeLayeredLayout(['a'], {
      a: exec({ onSuccess: ['b'] }),
      b: exec({ onSuccess: ['c'] }),
      c: exec(),
    })
    expect(pos.b.y - pos.a.y).toBe(pos.c.y - pos.b.y)
    expect(pos.a.y).toBeLessThan(pos.b.y)
    expect(pos.b.y).toBeLessThan(pos.c.y)
  })

  test('multi-parent join settles at deepest ancestor + 1, not at layer 1', () => {
    // trigger -> a -> b -> d, trigger -> c -> d. d must be below b (deepest).
    const pos = computeLayeredLayout(['a', 'c'], {
      a: exec({ onSuccess: ['b'] }),
      b: exec({ onSuccess: ['d'] }),
      c: exec({ onSuccess: ['d'] }),
      d: exec(),
    })
    expect(pos.d.y).toBeGreaterThan(pos.b.y)
    expect(pos.d.y).toBeGreaterThan(pos.c.y)
  })

  test('siblings in the same layer share y and separate on x', () => {
    const pos = computeLayeredLayout(['a', 'b', 'c'], { a: exec(), b: exec(), c: exec() })
    expect(pos.a.y).toBe(pos.b.y)
    expect(pos.b.y).toBe(pos.c.y)
    expect(pos.a.x).not.toBe(pos.b.x)
    expect(pos.b.x).not.toBe(pos.c.x)
  })

  test('orphan nodes land in a trailing layer, not on top of the trigger', () => {
    const pos = computeLayeredLayout(['a'], { a: exec(), orphan: exec() })
    expect(pos.orphan.y).toBeGreaterThan(pos.a.y)
  })

  test('cycles do not hang the layout', () => {
    const pos = computeLayeredLayout(['a'], {
      a: exec({ onSuccess: ['b'] }),
      b: exec({ onSuccess: ['a'] }),
    })
    expect(pos.a).toBeDefined()
    expect(pos.b).toBeDefined()
  })
})
