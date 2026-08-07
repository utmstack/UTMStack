import { SEVERITY_RANK } from '@/features/alerts/types/alert.types'
import type { AdversaryResponse, Side } from '../types/adversary.types'

function sideLabel(s?: Side | null): string {
  return s?.ip || s?.host || s?.user || s?.domain || ''
}

// Merge two Side records field-wise; existing values win, new values fill gaps.
// Different alerts against the same victim can carry different partial info
// (one has IP+host, another just IP) — this keeps the richest view.
function mergeSide(prev: Side | undefined, next: Side | null | undefined): Side | undefined {
  if (!next) return prev
  if (!prev) return { ...next }
  return {
    ip: prev.ip ?? next.ip,
    host: prev.host ?? next.host,
    user: prev.user ?? next.user,
    port: prev.port ?? next.port,
    domain: prev.domain ?? next.domain,
    mac: prev.mac ?? next.mac,
    geolocation: prev.geolocation ?? next.geolocation,
  }
}

export interface SankeyNode {
  name: string
  label: string
  column: 'adversary' | 'alert' | 'victim'
  maxSeverity: number
  side?: Side
}

export interface SankeyLink {
  source: string
  target: string
  value: number
}

export interface SankeyGraph {
  nodes: SankeyNode[]
  links: SankeyLink[]
}

// Prefixes keep node keys unique across columns — an IP can be both attacker and victim.
const ADV = 'A:'
const ALT = 'E:'
const VIC = 'V:'

export function buildSankey(responses: AdversaryResponse[]): SankeyGraph {
  const nodes = new Map<string, SankeyNode>()
  const linkMap = new Map<string, SankeyLink>()

  const ensureNode = (
    key: string,
    label: string,
    column: SankeyNode['column'],
    severity: number,
    side?: Side | null,
  ) => {
    const cur = nodes.get(key)
    if (cur) {
      if (severity > cur.maxSeverity) cur.maxSeverity = severity
      if (side) cur.side = mergeSide(cur.side, side)
    } else {
      nodes.set(key, { name: key, label, column, maxSeverity: severity, side: side ?? undefined })
    }
  }

  const bumpLink = (source: string, target: string) => {
    const k = `${source}→${target}`
    const cur = linkMap.get(k)
    if (cur) cur.value++
    else linkMap.set(k, { source, target, value: 1 })
  }

  for (const r of responses) {
    const advLabel = sideLabel(r.adversary)
    if (!advLabel || !r.alerts?.length) continue
    const advKey = ADV + advLabel

    for (const wrap of r.alerts) {
      const chain = [wrap.alert, ...(wrap.children ?? [])]
      for (const a of chain) {
        const name = (a.name ?? '').trim()
        const victimLabel = sideLabel(a.target)
        if (!name || !victimLabel) continue
        const altKey = ALT + name
        const vicKey = VIC + victimLabel
        const sev = SEVERITY_RANK[a.severity ?? ''] ?? 0
        ensureNode(advKey, advLabel, 'adversary', sev, r.adversary)
        ensureNode(altKey, name, 'alert', sev)
        ensureNode(vicKey, victimLabel, 'victim', sev, a.target)
        bumpLink(advKey, altKey)
        bumpLink(altKey, vicKey)
      }
    }
  }

  return {
    nodes: [...nodes.values()],
    links: [...linkMap.values()].sort((a, b) => b.value - a.value),
  }
}
