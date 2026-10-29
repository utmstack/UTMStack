// The Live Operations timeline: every alert of the chosen window as a marker (row by severity),
// the one on stage marked, LIVE to follow new alerts, step back and forth, and replay the window.
import { useEffect, useMemo, useRef, useState } from 'react'
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
  alertsCount,
  win,
  onWin,
  selectedId,
  live,
  page,
  pages,
  onPage,
  onSelect,
  onLive,
  replaying,
  onReplay,
}: {
  t: (k: string, o?: Record<string, unknown>) => string
  alerts: Alert[]
  alertsCount:number,
  win: WindowKey
  onWin: (w: WindowKey) => void
  selectedId: string | null
  live: boolean
  /** 0-based page of the window; the timeline shows only the current page. */
  page: number
  pages: number
  onPage: (p: number) => void
  onSelect: (id: string) => void
  onLive: () => void
  replaying: boolean
  onReplay: () => void
}) {
  const [hover, setHover] = useState<Alert | null>(null)
  const now = Date.now()
  // One page: the window itself. Several pages: the left edge is the oldest alert loaded,
  // so the span covers exactly the visible list, not the whole window.
  const defFrom = useMemo(() => {
    if (pages <= 1 || alerts.length === 0) return now - WINDOW_MS[win]
    return Math.min(...alerts.map((a) => tsMs(a['@timestamp'] ?? '')).filter((x) => Number.isFinite(x)))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [alerts, page, pages, win, Math.floor(now / 60_000)])
  // Mouse panning and zooming move the view; a data change snaps back to the default.
  const [view, setView] = useState<[number, number] | null>(null)
  useEffect(() => setView(null), [defFrom])
  const vFrom = view?.[0] ?? defFrom
  const vTo = view?.[1] ?? now
  const from = vFrom
  const pos = (iso: string) => ((tsMs(iso) - from) / (vTo - from)) * 100
  const idx = alerts.findIndex((a) => a.id === selectedId)

  // Drag to pan (pointer capture-less: window listeners), wheel to zoom at the cursor.
  const laneRef = useRef<HTMLDivElement>(null)
  const viewRef = useRef<{ from: number; to: number; x: number; w: number } | null>(null)
  const draggedRef = useRef(false)
  useEffect(() => {
    const el = laneRef.current
    if (!el) return
    const onWheel = (e: WheelEvent) => {
      e.preventDefault()
      const [f, t] = view ?? [defFrom, now]
      const span = t - f
      const frac = (e.offsetX / el.clientWidth) || 0
      const anchor = f + frac * span
      const factor = Math.exp(e.deltaY * 0.0015)
      let ns = Math.min(Math.max(span * factor, 60_000), now - defFrom)
      setView([anchor - frac * ns, anchor - frac * ns + ns])
    }
    el.addEventListener('wheel', onWheel, { passive: false })
    return () => el.removeEventListener('wheel', onWheel)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [view, defFrom])
  function startPan(e: React.MouseEvent) {
    if (e.button !== 0) return
    e.preventDefault()
    const [f, t] = view ?? [defFrom, now]
    viewRef.current = { from: f, to: t, x: e.clientX, w: laneRef.current?.clientWidth || 1 }
    const move = (ev: MouseEvent) => {
      const v = viewRef.current
      if (!v) return
      const dt = ((v.x - ev.clientX) / v.w) * (v.to - v.from)
      if (Math.abs(ev.clientX - v.x) > 3) draggedRef.current = true
      setView([v.from + dt, v.to + dt])
    }
    const up = () => {
      viewRef.current = null
      window.removeEventListener('mousemove', move)
      window.removeEventListener('mouseup', up)
      window.setTimeout(() => (draggedRef.current = false), 0)
    }
    window.addEventListener('mousemove', move)
    window.addEventListener('mouseup', up)
  }
  const ticks = useMemo(() => {
    const span = vTo - vFrom
    const step = span <= 3_600_000 ? 10 * 60_000 : span <= 6 * 3_600_000 ? 3_600_000 : span <= 24 * 3_600_000 ? 3 * 3_600_000 : 12 * 3_600_000
    const out: { left: number; label: string }[] = []
    for (let tt = Math.ceil(vFrom / step) * step; tt < vTo; tt += step) {
      out.push({ left: ((tt - vFrom) / span) * 100, label: new Date(tt).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' }) })
    }
    return out
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [vFrom, vTo, Math.floor(now / 60_000)])

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
          {t('liveOps.timCount', { count: alertsCount, range: WINDOW_OPTIONS.find((o) => o.key === win)?.label ?? '' })}
          {pages > 1 && <span className="ml-1.5 text-muted-foreground/70">· {t('liveOps.timPage', { p: page + 1, n: pages })}</span>}
          <span className="ml-2 ">
            <button type="button" disabled={page <= 0} onClick={() => onPage(page - 1)} className="rounded-md border border-border p-1 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-30" title={t('liveOps.timPagePrev')}>
              <ChevronLeft size={16} />
            </button>
            <button type="button" disabled={page >= pages - 1} onClick={() => onPage(page + 1)} className="rounded-md border border-border p-1 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-30" title={t('liveOps.timPageNext')}>
              <ChevronRight size={16} />
            </button>
          </span>
        </span>
        <div className="ml-auto flex items-center gap-1">
          <button type="button" disabled={idx <= 0} onClick={() => idx > 0 && onSelect(alerts[idx - 1].id)} className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-30" title={t('liveOps.timPrev')}>
            <ChevronLeft size={16} />
          </button>
          <button type="button" disabled={idx < 0 || idx >= alerts.length - 1} onClick={() => idx >= 0 && idx < alerts.length - 1 && onSelect(alerts[idx + 1].id)} className="rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-30" title={t('liveOps.timNext')}>
            <ChevronRight size={16} />
          </button>
          <button type="button" onClick={onReplay} className={cn('flex items-center gap-1 rounded-md px-2 py-1 text-[11.5px] font-medium', replaying ? 'bg-primary/15 text-primary' : 'text-muted-foreground hover:bg-muted hover:text-foreground')} title={t('liveOps.replayTip')}>
            {replaying ? <Square size={12} /> : <History size={13} />} {replaying ? t('liveOps.timStop') : t('liveOps.timReplay')}
          </button>
          <button type="button" onClick={onLive} className={cn('flex items-center gap-1.5 rounded-md px-2.5 py-1 text-[11.5px] font-bold tracking-wide', live ? 'bg-red-500/15 text-red-600 dark:text-red-400' : 'border border-border text-muted-foreground hover:text-foreground')} title={t('liveOps.liveTip')}>
            <span className={cn('h-2 w-2 rounded-full', live ? 'lo-live-dot bg-red-500' : 'bg-muted-foreground')} /> {t('liveOps.timLive')}
          </button>
        </div>
      </div>

      <div ref={laneRef} className="relative mt-2 h-[58px] cursor-grab select-none active:cursor-grabbing" onMouseDown={startPan} onMouseLeave={() => setHover(null)}>
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
          const off = a['@timestamp'] ? tsMs(a['@timestamp']) < vFrom || tsMs(a['@timestamp']) > vTo : true
          return (
            <button
              key={a.id}
              type="button"
              onClick={() => !draggedRef.current && onSelect(a.id)}
              onMouseEnter={() => setHover(a)}
              className={cn('absolute -translate-x-1/2 -translate-y-1/2 rounded-full transition-transform hover:scale-150', dot(String(a.severity)), selected ? 'h-3.5 w-3.5 ring-2 ring-primary ring-offset-1 ring-offset-card' : 'h-2.5 w-2.5 opacity-85', fresh && 'lo-new-dot')}
              style={{ left: `${pos(a['@timestamp'] ?? '')}%`, top: 10 + (ROW[String(a.severity)] ?? 2) * 16, visibility: off ? 'hidden' : 'visible' }}
              aria-label={a['@timestamp'] ? t('liveOps.alertAt', { name: a.name, time: new Date(tsMs(a['@timestamp'])).toLocaleTimeString() }) : a.name}
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
