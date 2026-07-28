import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { useThemeContext } from '@/app/providers/ThemeProvider'
import { EChartsRenderer } from '@/features/dashboard/components/EChartsRenderer'
import { flagEmoji } from '@/features/alerts/lib/alert-meta'
import { buildSankey, type SankeyNode } from '../lib/sankey-data'
import type { AdversaryResponse, Side } from '../types/adversary.types'

const COLUMN_COLOR = {
  adversary: '#ef4444',
  alert: '#f59e0b',
  victim: '#3b82f6',
} as const

const SEVERITY_COLOR: Record<number, string> = {
  4: '#b91c1c',
  3: '#ef4444',
  2: '#f59e0b',
  1: '#38bdf8',
  0: '#64748b',
}

function nodeColor(n: SankeyNode): string {
  if (n.column === 'adversary') return SEVERITY_COLOR[Math.min(n.maxSeverity, 4)] ?? COLUMN_COLOR.adversary
  return COLUMN_COLOR[n.column]
}

const escapeHtml = (s: string) =>
  s.replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[c] as string)

function renderSideRows(side: Side, labels: Record<string, string>): string {
  const rows: [string, string | undefined][] = [
    [labels.ip, side.ip],
    [labels.host, side.host],
    [labels.user, side.user],
    [labels.domain, side.domain],
    [labels.port, side.port != null ? String(side.port) : undefined],
    [labels.mac, side.mac],
  ]
  const g = side.geolocation
  if (g) {
    const country = g.country ? `${flagEmoji(g.countryCode)} ${g.country}`.trim() : undefined
    rows.push(
      [labels.country, country],
      [labels.city, g.city],
      [labels.asn, g.asn != null ? String(g.asn) : undefined],
      [labels.aso, g.aso],
    )
  }
  return rows
    .filter(([, v]) => v)
    .map(
      ([k, v]) =>
        `<tr><td style="padding:2px 8px 2px 0;color:#94a3b8;font-size:11px;">${escapeHtml(k)}</td>` +
        `<td style="padding:2px 0;font-family:ui-monospace,SFMono-Regular,monospace;font-size:11px;">${escapeHtml(v!)}</td></tr>`,
    )
    .join('')
}

export function AdversariesSankey({ data }: { data: AdversaryResponse[] }) {
  const { t } = useTranslation()
  const { theme } = useThemeContext()
  const labelColor = theme === 'dark' ? '#f8fafc' : '#0f172a'
  const graph = useMemo(() => buildSankey(data), [data])

  const fieldLabels = useMemo(
    () => ({
      ip: t('adversaries.field.ip'),
      host: t('adversaries.field.host'),
      user: t('adversaries.field.user'),
      port: t('adversaries.field.port'),
      domain: t('adversaries.field.domain'),
      mac: t('adversaries.field.mac'),
      country: t('adversaries.field.country'),
      city: t('adversaries.field.city'),
      asn: t('adversaries.field.asn'),
      aso: t('adversaries.field.aso'),
    }),
    [t],
  )

  const option = useMemo(
    () => ({
      tooltip: {
        trigger: 'item',
        triggerOn: 'mousemove',
        extraCssText: 'max-width:620px;',
        formatter: (p: {
          dataType: string
          name: string
          data: { source?: string; target?: string; value?: number; side?: Side; label?: string }
        }) => {
          if (p.dataType === 'edge') {
            const src = p.data.source?.slice(2) ?? ''
            const tgt = p.data.target?.slice(2) ?? ''
            return `${escapeHtml(src)} → ${escapeHtml(tgt)}<br/><b>${p.data.value ?? 0}</b>`
          }
          const label = p.data.label ?? p.name?.slice(2) ?? ''
          const header = `<div style="font-weight:600;margin-bottom:4px;">${escapeHtml(label)}</div>`
          if (!p.data.side) return header
          const rows = renderSideRows(p.data.side, fieldLabels)
          if (!rows) return header
          return `${header}<table style="border-collapse:collapse;">${rows}</table>`
        },
      },
      series: [
        {
          type: 'sankey',
          left: 50,
          right: 50,
          top: 24,
          bottom: 16,
          nodeAlign: 'justify',
          nodeGap: 12,
          nodeWidth: 17,
          layoutIterations: 32,
          emphasis: { focus: 'adjacency' },
          lineStyle: { color: 'gradient', curveness: 0.8, opacity: 0.75 },
          label: {
            color: labelColor,
            show: true,
            fontSize: 16,
            formatter: (p: { data: { label?: string }; name: string }) => (p.data.label ?? p.name.slice(2)) ,
          },
          data: graph.nodes.map((n) => ({
            name: n.name,
            label: n.label,
            side: n.side,
            itemStyle: { color: nodeColor(n) },
          })),
          links: graph.links,
          levels: [
            {
              depth: 0,
              label: { position: 'left', align: 'right' },
              itemStyle: { borderColor: COLUMN_COLOR.adversary },
            },
            {
              depth: 1,
              label: { position: 'top', align: 'center', width: 380 },
              itemStyle: { borderColor: COLUMN_COLOR.alert },
            },
            {
              depth: 2,
              label: { position: 'right', align: 'left', width: 340 },
              itemStyle: { borderColor: COLUMN_COLOR.victim },
            },
          ],
        },
      ],
    }),
    [graph, fieldLabels, labelColor],
  )

  if (!graph.nodes.length || !graph.links.length) {
    return (
      <div className="flex h-full items-center justify-center rounded-xl border border-border bg-card text-sm text-muted-foreground">
        {t('adversaries.empty')}
      </div>
    )
  }

  const headers: { key: string; color: string }[] = [
    { key: 'adversaries.col.adversary', color: COLUMN_COLOR.adversary },
    { key: 'adversaries.col.alerts', color: COLUMN_COLOR.alert },
    { key: 'adversaries.col.victims', color: COLUMN_COLOR.victim },
  ]


  return (
    <div
      style={{
        height:`${Math.min(Math.max(data.length,50)*1.7,300)}%`
      }}
      className={`flex  w-full flex-col rounded-xl border border-border bg-card p-3`}>
      <div className="grid grid-cols-3 gap-2 pb-2 text-xs font-medium">
        {headers.map((h, i) => (
          <div
            key={h.key}
            className={cn(
              'flex items-center gap-2',
              i === 0 && 'justify-start',
              i === 1 && 'justify-center',
              i === 2 && 'justify-end',
            )}
          >
            <span className="inline-block h-2 w-2 rounded-full" style={{ background: h.color }} />
            <span className="uppercase tracking-wide text-muted-foreground">{t(h.key)}</span>
          </div>
        ))}
      </div>
      <div className="min-h-0 flex-1">
        <EChartsRenderer option={option} />
      </div>
    </div>
  )
}
