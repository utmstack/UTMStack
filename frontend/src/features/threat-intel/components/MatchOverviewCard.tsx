import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useTiIocsOverview } from '../hooks/use-ti-iocs-overview'
import type { AdvancedSearchRequest } from '../domain/threat-intel.types'

const TYPE_FAMILIES: Record<'ip' | 'url' | 'domain' | 'signatures', string[]> = {
  ip:         ['ip', 'cidr'],
  url:        ['url', 'link', 'github-organization', 'github-repository'],
  domain:     ['domain', 'hostname'],
  signatures: [
    'md5', 'sha1', 'sha224', 'sha256', 'sha384', 'sha512', 'sha512-224', 'sha512-256',
    'sha3-224', 'sha3-256', 'sha3-384', 'sha3-512',
    'authentihash', 'cdhash', 'malware', 'filename',
    'profile-photo', 'facebook-profile', 'tiktok-profile', 'twitter-profile',
  ],
}

// Show time-of-day for sub-day intervals, date for day+ intervals.
function formatTs(ts: number, interval: string): string {
  const perDay = /min|hour|^[0-9]+[mh]$/i.test(interval)
  return new Date(ts).toLocaleString(undefined, perDay
    ? { hour: '2-digit', minute: '2-digit' }
    : { month: 'short', day: '2-digit' })
}

interface MatchOverviewCardProps {
  body: AdvancedSearchRequest | undefined
  interval: string
}

export function MatchOverviewCard({ body, interval }: MatchOverviewCardProps) {
  const { t } = useTranslation()
  const query = useTiIocsOverview(body, interval)

  const total = useMemo(() => {
    if (query.data?.kind !== 'ok') return 0
    return query.data.value.items
  }, [query.data])

  const points = useMemo(() => {
    if (query.data?.kind !== 'ok') return []
    const buckets = query.data.value.aggregations?.histogram?.buckets ?? []
    return buckets
      .map((b) => ({ ts: Number(b.key), count: b.doc_count }))
      .filter((b) => Number.isFinite(b.ts))
      .sort((a, b) => a.ts - b.ts)
  }, [query.data])

  const familyCounts = useMemo(() => {
    const zero = { ip: 0, url: 0, domain: 0, signatures: 0 }
    if (query.data?.kind !== 'ok') return zero
    const buckets = query.data.value.aggregations?.by_types?.buckets ?? []
    const byType: Record<string, number> = {}
    for (const b of buckets) byType[String(b.key)] = b.doc_count
    const sum = (types: string[]) => types.reduce((n, t) => n + (byType[t] ?? 0), 0)
    return {
      ip:         sum(TYPE_FAMILIES.ip),
      url:        sum(TYPE_FAMILIES.url),
      domain:     sum(TYPE_FAMILIES.domain),
      signatures: sum(TYPE_FAMILIES.signatures),
    }
  }, [query.data])

  const w = 1000
  const h = 100
  const counts = points.map((p) => p.count)
  const max = counts.length > 0 ? Math.max(...counts) * 1.15 : 1
  const xs = counts.map((_, i) => (i * w) / Math.max(counts.length - 1, 1))
  const ys = counts.map((v) => h - (v / max) * h)

  let linePath = ''
  if (counts.length > 0) {
    linePath = counts.reduce((acc, _, i) => {
      if (i === 0) return `M ${xs[i]} ${ys[i]}`
      const prevX = xs[i - 1]
      const prevY = ys[i - 1]
      const cx1 = prevX + (xs[i] - prevX) / 2
      const cx2 = xs[i] - (xs[i] - prevX) / 2
      return `${acc} C ${cx1} ${prevY}, ${cx2} ${ys[i]}, ${xs[i]} ${ys[i]}`
    }, '')
  } else {
    linePath = `M 0 ${h} L ${w} ${h}`
  }

  const areaPath = counts.length > 0 ? `${linePath} L ${xs[xs.length - 1]} ${h} L ${xs[0]} ${h} Z` : linePath

  const startLabel = points.length > 0 ? formatTs(points[0].ts, interval) : ''
  const endLabel   = points.length > 0 ? formatTs(points[points.length - 1].ts, interval) : ''
  const midLabel   = points.length > 1
    ? formatTs(points[Math.floor(points.length / 2)].ts, interval)
    : ''

  if (query.data?.kind === 'not-configured') return null

  return (
    <div className="rounded-xl border border-border bg-card p-6">
      <div className="flex items-baseline justify-between">
        <div>
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
            {t('threatIntel.overview.title')}
          </div>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-3xl font-semibold tabular-nums">
              {query.isPending ? '—' : total.toLocaleString()}
            </span>
            <span className="text-sm text-muted-foreground">{t('threatIntel.overview.totalIndicators')}</span>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {(['ip', 'url', 'domain', 'signatures'] as const).map((family) => (
            <div
              key={family}
              className="rounded-md border border-border bg-background px-3 py-1.5"
            >
              <div className="text-[10px] uppercase tracking-wider text-muted-foreground">
                {t(`threatIntel.overview.families.${family}`, { defaultValue: family })}
              </div>
              <div className="text-lg font-semibold tabular-nums">
                {query.isPending ? '—' : familyCounts[family].toLocaleString()}
              </div>
            </div>
          ))}
        </div>
      </div>

      <svg viewBox={`0 0 ${w} ${h}`} className="mt-4 h-24 w-full" preserveAspectRatio="none">
        <defs>
          <linearGradient id="iocGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="rgb(168 85 247)" stopOpacity="0.28" />
            <stop offset="100%" stopColor="rgb(168 85 247)" stopOpacity="0" />
          </linearGradient>
        </defs>
        {counts.length > 0 && <path d={areaPath} fill="url(#iocGrad)" />}
        <path
          d={linePath}
          fill="none"
          stroke="rgb(168 85 247)"
          strokeOpacity="0.85"
          strokeWidth="1.75"
          strokeLinejoin="round"
          strokeLinecap="round"
        />
      </svg>

      <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
        <span>{startLabel}</span>
        <span>{midLabel}</span>
        <span>{endLabel}</span>
      </div>
    </div>
  )
}
