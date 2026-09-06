import { describe, expect, it } from 'vitest'
import { enrichmentAncestors } from './ancestors'
import type { FlowNode } from '../types/soar.types'

const enrich = (executor: string): FlowNode => ({
  kind: 'enrichment',
  executor,
  onSuccess: ['child'],
})

describe('enrichmentAncestors static fields', () => {
  const nodes: Record<string, FlowNode> = {
    llm1: enrich('llm_enrich'),
    geo: enrich('http'),
    child: { kind: 'executor', executor: 'shell', command: 'echo' },
  }

  it('advertises result for llm_enrich parents (backend-guaranteed shape)', () => {
    const out = enrichmentAncestors(nodes, 'child')
    expect(out.find((a) => a.nodeId === 'llm1')?.fields).toEqual(['result'])
  })

  it('leaves http parents runtime-dependent (no static fields)', () => {
    const out = enrichmentAncestors(nodes, 'child')
    expect(out.find((a) => a.nodeId === 'geo')?.fields).toEqual([])
  })
})
