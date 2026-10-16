import type { AgentStep, PanelRequest } from './types'

export function panelForStep(step: AgentStep): PanelRequest | null {
  if (step.status !== 'done' || !step.payload) return null

  const payload = step.payload
  switch (payload.kind) {
    case 'details':
      return {
        kind: 'entity',
        title: payload.details.attributes.value || payload.id,
        ref: payload.id,
      }
    case 'lookup':
      return payload.entity ? { kind: 'entity', title: payload.value, ref: payload.entity.id } : null
    case 'relations':
      return { kind: 'graph', title: rootValue(payload), ref: payload.id }
    case 'geo':
      return payload.locations.length ? { kind: 'geo', title: payload.id, ref: payload.id } : null
    case 'search':
      return payload.results.length ? { kind: 'search', title: payload.query, ref: payload.query } : null
  }
}

function rootValue(payload: Extract<import('./types').ToolPayload, { kind: 'relations' }>): string {
  return payload.graph.nodes.find((n) => n.id === payload.id)?.value || payload.id
}
