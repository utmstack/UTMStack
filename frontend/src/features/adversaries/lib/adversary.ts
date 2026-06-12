import type {
  AdversaryKind,
  AdversaryResponse,
  AlertLite,
  AlertRaw,
  Adversary,
  Side,
  ThreatLevel,
} from '../types/adversary.types'

function sideKind(s?: Side | null): AdversaryKind {
  if (s?.ip) return 'ip'
  if (s?.host) return 'host'
  if (s?.user) return 'user'
  return 'ip'
}
function sideLabel(s?: Side | null): string {
  return s?.ip || s?.host || s?.user || s?.domain || '—'
}

export function threatLevelOf(severity: number): ThreatLevel {
  if (severity >= 4) return 'critical'
  if (severity === 3) return 'high'
  if (severity === 2) return 'medium'
  return 'low'
}

/** 24 hourly buckets over the last 24h (most recent on the right). */
function activityHistogram(timestamps: string[]): number[] {
  const buckets = new Array(24).fill(0)
  const now = Date.now()
  for (const ts of timestamps) {
    const t = new Date(ts).getTime()
    if (Number.isNaN(t)) continue
    const hoursAgo = (now - t) / 3_600_000
    if (hoursAgo < 0 || hoursAgo >= 24) continue
    buckets[23 - Math.floor(hoursAgo)]++
  }
  return buckets
}

/** Build the derived Adversary view-model list from the grouped backend response. */
export function toAdversaries(responses: AdversaryResponse[]): Adversary[] {
  const out: Adversary[] = []
  for (const r of responses) {
    const side = r.adversary
    const identifier = sideLabel(side)
    if (identifier === '—' || !r.alerts?.length) continue

    const parents = r.alerts.map((a) => a.alert)
    const all: AlertRaw[] = r.alerts.flatMap((a) => [a.alert, ...(a.children ?? [])])

    const severities = all.map((a) => a.severity ?? 0)
    const maxSeverity = severities.length ? Math.max(...severities) : 0

    const times = all.map((a) => a['@timestamp']).filter(Boolean) as string[]
    const sortedTimes = [...times].sort()
    const firstSeen = sortedTimes[0]
    const lastSeen = sortedTimes[sortedTimes.length - 1]

    // Techniques: count occurrences of the (string) technique across all alerts.
    const techMap = new Map<string, number>()
    for (const a of all) {
      const tch = (a.technique ?? '').trim()
      if (tch) techMap.set(tch, (techMap.get(tch) ?? 0) + 1)
    }
    const techniques = [...techMap.entries()]
      .map(([id, count]) => ({ id, count }))
      .sort((a, b) => b.count - a.count)

    // Targets: aggregate by the victim Side.
    const tgtMap = new Map<string, { label: string; kind: AdversaryKind; hits: number }>()
    for (const a of all) {
      if (!a.target) continue
      const label = sideLabel(a.target)
      if (label === '—') continue
      const key = label
      const cur = tgtMap.get(key) ?? { label, kind: sideKind(a.target), hits: 0 }
      cur.hits++
      tgtMap.set(key, cur)
    }
    const targets = [...tgtMap.values()].sort((a, b) => b.hits - a.hits)

    const tags = Array.from(new Set(all.flatMap((a) => a.tags ?? [])))

    const alerts: AlertLite[] = parents
      .slice()
      .sort((a, b) => (b['@timestamp'] ?? '').localeCompare(a['@timestamp'] ?? ''))
      .slice(0, 30)
      .map((a, i) => ({
        id: a.id || `${identifier}-${i}`,
        timestamp: a['@timestamp'],
        name: a.name || '—',
        category: a.category,
        severity: a.severity ?? 0,
        severityLabel: a.severityLabel,
        technique: a.technique,
        statusLabel: a.statusLabel,
        echoes: r.alerts.find((aw) => aw.alert === a)?.children?.length ?? 0,
        target: a.target ? { label: sideLabel(a.target), kind: sideKind(a.target) } : undefined,
      }))

    out.push({
      id: identifier,
      identifier,
      kind: sideKind(side),
      geo: side?.geolocation,
      threatLevel: threatLevelOf(maxSeverity),
      maxSeverity,
      alertsCount: all.length,
      firstSeen,
      lastSeen,
      techniques,
      targets,
      activity: activityHistogram(times),
      alerts,
      tags,
    })
  }
  // Most dangerous first, then most alerts.
  return out.sort((a, b) => b.maxSeverity - a.maxSeverity || b.alertsCount - a.alertsCount)
}
