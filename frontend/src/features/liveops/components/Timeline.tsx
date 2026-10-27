// The Live Operations timeline: every alert of the chosen window as a marker (row by severity),
// the one on stage marked, LIVE to follow new alerts, step back and forth, and replay the window.
import { useMemo, useState } from 'react'
import { ChevronLeft, ChevronRight, History, Square } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { SEV_META } from '@/features/alerts/lib/alert-meta'
import type { Alert } from '@/features/alerts/types/alert.types'
import { tsMs } from '../model'

const ROW: Record<string, number> = { high: 0, medium: 1, low: 2 }
// Severity colours as the Alerts page shows them.
const dot = (severity: string) => (SEV_META[severity as keyof typeof SEV_META] ?? SEV_META.low).bar

/** Time window options, newest first; labels are plain (1h / 6h / 24h / 7d). */
export type WindowKey = '1h' | '6h' | '24h' | '7d'
export const WINDOW_OPTIONS: { key: WindowKey; label: string }[] = [
  { key: '1h', label: '1h' },
  { key: '6h', label: '6h' },
  { key: '24h', label: '24h' },
  { key: '7d', label: '7d' },
]
export const WINDOW_MS: Record<WindowKey, number> = {
  '1h': 3_600_000,
  '6h': 6 * 3_600_000,
  '24h': 24 * 3_600_000,
  '7d': 7 * 24 * 3_600_000,
}

export function Timeline({
  t,
  alerts,
  win,
  onWin,
  selectedId,
  live,
  onSelect,
  onLive,
  replaying,
  onReplay,
}: {
  t: (k: string, o?: Record<string, unknown>) => string
  alerts: Alert[]
  win: WindowKey
  onWin: (w: WindowKey) => void
  selectedId: string | null
  live: boolean
  onSelect: (id: string) => void
  onLive: () => void
  replaying: boolean
  onReplay: () => void
}) {
  const [hover, setHover] = useState<Alert | null>(null)
  const now = Date.now()
  const from = now - WINDOW_MS[win]
  const pos = (iso: string) => ((tsMs(iso) - from) / (now - from)) * 100
  const idx = alerts.findIndex((a) => a.id === selectedId)
  const ticks = useMemo(() => {
    const span = WINDOW_MS[win]
    const step = span <= 3_600_000 ? 10 * 60_000 : span <= 6 * 3_600_000 ? 3_600_000 : span <= 24 * 3_600_000 ? 3 * 3_600_000 : 12 * 3_600_000
    const out: { left: number; label: string }[] = []
    for (let tt = Math.ceil(from / step) * step; tt < now; tt += step) {
      out.push({ left: ((tt - from) / (now - from)) * 100, label: new Date(tt).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' }) })
    }
    return out
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [win, Math.floor(now / 60_000)])

  return (
    <div className="lo-panel rounded-xl px-3 pb-2 pt-2">
      <div className="flex items-center gap-2">
        <div className="flex rounded-md border border-border p-0.5">
          {WINDOW_OPTIONS.map((o) => (
            <button key={o.key} type="button" onClick={() => onWin(o.key)} className={cn('rounded px-2 py-0.5 text-[11.5px] font-medium', o.key === win ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground')}>
              {o.label}
            </button>
          ))}
        </div>
        <span className="text-[11.5px] text-muted-foreground">
          {t('liveOps.timCount', { count: alerts.length, range: WINDOW_OPTIONS.find((o) => o.key === win)?.label ?? '' })}
        </span>
        <div className="ml-auto flex items-center gap-1">
          <button type="button" disabled={idx <= 0} onClick={() => idx > 0 && onSelect(alerts[idx - 1].id)} className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-30" title={t('liveOps.timPrev')}>
            <ChevronLeft size={16} />
          </button>
          <button type="button" disabled={idx < 0 || idx >= alerts.length - 1} onClick={() => idx >= 0 && idx < alerts.length - 1 && onSelect(alerts[idx + 1].id)} className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-30" title={t('liveOps.timNext')}>
            <ChevronRight size={16} />
          </button>
          <button type="button" onClick={onReplay} className={cn('flex items-center gap-1 rounded-md px-2 py-1 text-[11.5px] font-medium', replaying ? 'bg-primary/15 text-primary' : 'text-muted-foreground hover:bg-muted hover:text-foreground')} title="Play every alert of this window in order">
            {replaying ? <Square size={12} /> : <History size={13} />} {replaying ? t('liveOps.timStop') : t('liveOps.timReplay')}
          </button>
          <button type="button" onClick={onLive} className={cn('flex items-center gap-1.5 rounded-md px-2.5 py-1 text-[11.5px] font-bold tracking-wide', live ? 'bg-red-500/15 text-red-600 dark:text-red-400' : 'border border-border text-muted-foreground hover:text-foreground')} title="Follow new alerts as they arrive">
            <span className={cn('h-2 w-2 rounded-full', live ? 'lo-live-dot bg-red-500' : 'bg-muted-foreground')} /> {t('liveOps.timLive')}
          </button>
        </div>
      </div>

      <div className="relative mt-2 h-[58px] select-none" onMouseLeave={() => setHover(null)}>
        {[0, 1, 2].map((r) => (
          <div key={r} className="absolute left-0 right-0 border-t border-dashed border-border/70" style={{ top: 10 + r * 16 }} />
        ))}
        {ticks.map((t, i) => (
          <div key={i} className="absolute top-0 h-[50px] border-l border-border/60" style={{ left: `${t.left}%` }}>
            {t.left < 95 && <span className="absolute top-[46px] -translate-x-1/2 whitespace-nowrap text-[10px] text-muted-foreground">{t.label}</span>}
          </div>
        ))}
        {selectedId && idx >= 0 && <div className="absolute top-0 h-[52px] w-px bg-primary" style={{ left: `${pos(alerts[idx]['@timestamp'] ?? '')}%` }} />}
        {alerts.map((a) => {
          const selected = a.id === selectedId
          const fresh = !!a['@timestamp'] && now - tsMs(a['@timestamp']) < 5*60_000
          return (
            <button
              key={a.id}
              type="button"
              onClick={() => onSelect(a.id)}
              onMouseEnter={() => setHover(a)}
              className={cn('absolute -translate-x-1/2 -translate-y-1/2 rounded-full transition-transform hover:scale-150', dot(String(a.severity)), selected ? 'h-3.5 w-3.5 ring-2 ring-primary ring-offset-1 ring-offset-card' : 'h-2.5 w-2.5 opacity-85', fresh && 'lo-new-dot')}
              style={{ left: `${pos(a['@timestamp'] ?? '')}%`, top: 10 + (ROW[String(a.severity)] ?? 2) * 16 }}
              aria-label={`${a.name} at ${a['@timestamp'] ? new Date(tsMs(a['@timestamp'])).toLocaleTimeString() : ''}`}
            />
          )
        })}
        <div className="absolute right-0 top-0 h-[52px] border-l-2 border-red-500/60">
          <span className="absolute -right-0.5 top-[46px] text-[10px] font-semibold text-red-500">{t('liveOps.timNow')}</span>
        </div>
        {hover && (
          <div className="pointer-events-none absolute -top-9 z-10 max-w-[360px] -translate-x-1/2 truncate rounded-md border border-border bg-popover px-2 py-1 text-[11.5px] text-popover-foreground shadow-lg" style={{ left: `${Math.min(88, Math.max(12, pos(hover['@timestamp'] ?? '')))}%` }}>
            <span className="font-mono text-muted-foreground">{hover['@timestamp'] ? new Date(tsMs(hover['@timestamp'])).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' }) : ''}</span> {hover.name}
          </div>
        )}
      </div>
    </div>
  )
}
