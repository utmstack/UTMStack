import { useEffect, useMemo, useRef, useState } from 'react'
import { cn } from '@/shared/lib/utils'

/* ─── Animation hooks ──────────────────────────────────────────────────── */

/** Flips true on the frame after mount — used to trigger CSS enter transitions. */
export function useMounted(): boolean {
  const [m, setM] = useState(false)
  useEffect(() => {
    const id = requestAnimationFrame(() => setM(true))
    return () => cancelAnimationFrame(id)
  }, [])
  return m
}

/** Eased count-up to `value` (re-animates from the previous value on change). */
export function useCountUp(value: number, duration = 900): number {
  const [n, setN] = useState(value)
  const fromRef = useRef(0)
  useEffect(() => {
    const from = fromRef.current
    const start = performance.now()
    let raf = 0
    const tick = (now: number) => {
      const p = Math.min(1, (now - start) / duration)
      const eased = 1 - Math.pow(1 - p, 3)
      setN(from + (value - from) * eased)
      if (p < 1) raf = requestAnimationFrame(tick)
      else fromRef.current = value
    }
    raf = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(raf)
  }, [value, duration])
  return n
}

/* ─── Formatting ───────────────────────────────────────────────────────── */

export function compact(n: number): string {
  const v = Math.round(n)
  if (v >= 1e9) return (v / 1e9).toFixed(1).replace(/\.0$/, '') + 'B'
  if (v >= 1e6) return (v / 1e6).toFixed(1).replace(/\.0$/, '') + 'M'
  if (v >= 1e3) return (v / 1e3).toFixed(1).replace(/\.0$/, '') + 'k'
  return String(v)
}

export function relTime(iso?: string): string {
  if (!iso) return ''
  const d = new Date(iso).getTime()
  if (Number.isNaN(d)) return ''
  const s = Math.max(0, Math.round((Date.now() - d) / 1000))
  if (s < 60) return `${s}s`
  const m = Math.round(s / 60)
  if (m < 60) return `${m}m`
  const h = Math.round(m / 60)
  if (h < 24) return `${h}h`
  return `${Math.round(h / 24)}d`
}

/* ─── Sparkline ────────────────────────────────────────────────────────── */

export function Sparkline({ points, color = 'currentColor', className }: { points: number[]; color?: string; className?: string }) {
  if (points.length < 2) return <svg className={className} />
  const w = 100
  const h = 30
  const max = Math.max(...points, 1)
  const xs = points.map((_, i) => (i * w) / (points.length - 1))
  const ys = points.map((v) => h - (v / max) * (h - 3) - 1.5)
  const line = xs.map((x, i) => `${i ? 'L' : 'M'} ${x.toFixed(1)} ${ys[i].toFixed(1)}`).join(' ')
  const area = `${line} L ${w} ${h} L 0 ${h} Z`
  const gid = useMemo(() => `spk-${Math.round(xs[1] * 1000) + points.length}`, [points.length, xs])
  return (
    <svg viewBox={`0 0 ${w} ${h}`} preserveAspectRatio="none" className={className}>
      <defs>
        <linearGradient id={gid} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={color} stopOpacity="0.25" />
          <stop offset="100%" stopColor={color} stopOpacity="0" />
        </linearGradient>
      </defs>
      <path d={area} fill={`url(#${gid})`} />
      <path d={line} fill="none" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

/* ─── Multi-series area chart with hover tooltip + draw-in ─────────────── */

export interface ChartSeries {
  key: string
  label: string
  color: string
  points: { t: string; v: number }[]
}

export function AreaChart({ series, height = 200, format = compact }: { series: ChartSeries[]; height?: number; format?: (n: number) => string }) {
  const mounted = useMounted()
  const [hover, setHover] = useState<number | null>(null)
  const wrapRef = useRef<HTMLDivElement>(null)

  const len = Math.max(...series.map((s) => s.points.length), 0)
  const W = 1000
  const H = 260
  const padTop = 10
  const padBottom = 22
  const max = Math.max(1, ...series.flatMap((s) => s.points.map((p) => p.v))) * 1.15

  const x = (i: number) => (len <= 1 ? 0 : (i * W) / (len - 1))
  const y = (v: number) => padTop + (1 - v / max) * (H - padTop - padBottom)

  const paths = series.map((s) => {
    const pts = s.points
    const line = pts.reduce((acc, p, i) => {
      const xi = x(i)
      const yi = y(p.v)
      if (i === 0) return `M ${xi} ${yi}`
      const px = x(i - 1)
      const cx1 = px + (xi - px) / 2
      const cx2 = xi - (xi - px) / 2
      return `${acc} C ${cx1} ${y(pts[i - 1].v)}, ${cx2} ${yi}, ${xi} ${yi}`
    }, '')
    const area = `${line} L ${x(pts.length - 1)} ${H - padBottom} L ${x(0)} ${H - padBottom} Z`
    return { ...s, line, area }
  })

  // x-axis time ticks (~5)
  const base = series[0]?.points ?? []
  const ticks = base.length > 1 ? [0, 0.25, 0.5, 0.75, 1].map((f) => Math.round(f * (base.length - 1))) : []
  const fmtTick = (iso: string) => {
    const d = new Date(iso)
    return Number.isNaN(d.getTime()) ? '' : d.toLocaleString(undefined, { month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit' })
  }

  const onMove = (e: React.MouseEvent) => {
    const rect = wrapRef.current?.getBoundingClientRect()
    if (!rect || len < 2) return
    const rel = (e.clientX - rect.left) / rect.width
    setHover(Math.max(0, Math.min(len - 1, Math.round(rel * (len - 1)))))
  }

  return (
    <div ref={wrapRef} className="relative w-full select-none" style={{ height }} onMouseMove={onMove} onMouseLeave={() => setHover(null)}>
      <svg viewBox={`0 0 ${W} ${H}`} preserveAspectRatio="none" className="h-full w-full overflow-visible">
        {/* grid lines */}
        {[0.25, 0.5, 0.75].map((f) => (
          <line key={f} x1="0" x2={W} y1={padTop + f * (H - padTop - padBottom)} y2={padTop + f * (H - padTop - padBottom)} stroke="currentColor" strokeOpacity="0.06" strokeWidth="1" />
        ))}
        {paths.map((p) => (
          <g key={p.key}>
            <defs>
              <linearGradient id={`ag-${p.key}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={p.color} stopOpacity="0.22" />
                <stop offset="100%" stopColor={p.color} stopOpacity="0" />
              </linearGradient>
            </defs>
            <path d={p.area} fill={`url(#ag-${p.key})`} style={{ opacity: mounted ? 1 : 0, transition: 'opacity 700ms ease' }} />
            <path
              d={p.line}
              fill="none"
              stroke={p.color}
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              pathLength={1}
              style={{ strokeDasharray: 1, strokeDashoffset: mounted ? 0 : 1, transition: 'stroke-dashoffset 900ms ease' }}
            />
          </g>
        ))}
        {hover != null && (
          <line x1={x(hover)} x2={x(hover)} y1={padTop} y2={H - padBottom} stroke="currentColor" strokeOpacity="0.25" strokeWidth="1" />
        )}
        {hover != null &&
          paths.map((p) =>
            p.points[hover] ? <circle key={p.key} cx={x(hover)} cy={y(p.points[hover].v)} r="3.5" fill="var(--background)" stroke={p.color} strokeWidth="2" /> : null,
          )}
      </svg>

      {/* x-axis labels */}
      <div className="pointer-events-none absolute inset-x-0 bottom-0 flex justify-between px-1 text-[10px] text-muted-foreground">
        {ticks.map((ti, i) => (
          <span key={i} className={cn(i === 0 && 'translate-x-0', i === ticks.length - 1 && '-translate-x-2')}>
            {fmtTick(base[ti]?.t ?? '')}
          </span>
        ))}
      </div>

      {/* tooltip */}
      {hover != null && base[hover] && (
        <div
          className="pointer-events-none absolute top-1 z-10 rounded-lg border border-border bg-popover px-3 py-2 text-xs shadow-lg"
          style={{ left: `min(max(${(hover / Math.max(1, len - 1)) * 100}%, 70px), calc(100% - 70px))`, transform: 'translateX(-50%)' }}
        >
          <div className="mb-1 font-mono text-[10px] text-muted-foreground">{fmtTick(base[hover].t)}</div>
          {series.map((s) => (
            <div key={s.key} className="flex items-center gap-2">
              <span className="h-2 w-2 rounded-full" style={{ backgroundColor: s.color }} />
              <span className="text-muted-foreground">{s.label}</span>
              <span className="ml-auto font-mono font-medium tabular-nums">{format(s.points[hover]?.v ?? 0)}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

/* ─── Donut ────────────────────────────────────────────────────────────── */

export function Donut({ segments, size = 150, thickness = 16 }: { segments: { value: number; color: string; label: string }[]; size?: number; thickness?: number }) {
  const mounted = useMounted()
  const total = segments.reduce((a, s) => a + s.value, 0) || 1
  const r = (size - thickness) / 2
  const c = 2 * Math.PI * r
  let acc = 0
  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} className="-rotate-90">
      <circle cx={size / 2} cy={size / 2} r={r} fill="none" stroke="currentColor" strokeOpacity="0.08" strokeWidth={thickness} />
      {segments.map((s) => {
        const frac = s.value / total
        const dash = mounted ? frac * c : 0
        const off = -acc * c
        acc += frac
        return (
          <circle
            key={s.label}
            cx={size / 2}
            cy={size / 2}
            r={r}
            fill="none"
            stroke={s.color}
            strokeWidth={thickness}
            strokeLinecap="round"
            strokeDasharray={`${dash} ${c - dash}`}
            strokeDashoffset={off}
            style={{ transition: 'stroke-dasharray 800ms ease' }}
          />
        )
      })}
    </svg>
  )
}

/* ─── Treemap (slice-and-dice) ─────────────────────────────────────────── */

export interface TreemapItem {
  key: string
  value: number
  color: string
}

/** Squarified-ish treemap via recursive slice-and-dice; good enough for a top-N view. */
export function Treemap({ items, height = 260 }: { items: TreemapItem[]; height?: number }) {
  const mounted = useMounted()
  const total = items.reduce((a, i) => a + i.value, 0) || 1
  const rects = layout(items, 0, 0, 100, 100, total)
  return (
    <div className="relative w-full overflow-hidden rounded-lg" style={{ height }}>
      {rects.map((r) => (
        <div
          key={r.key}
          title={`${r.key}: ${compact(r.value)}`}
          className="absolute overflow-hidden border border-background p-1.5"
          style={{
            left: `${r.x}%`,
            top: `${r.y}%`,
            width: `${r.w}%`,
            height: `${r.h}%`,
            backgroundColor: r.color,
            opacity: mounted ? 1 : 0,
            transform: mounted ? 'scale(1)' : 'scale(0.85)',
            transition: 'opacity 500ms ease, transform 500ms ease',
          }}
        >
          {r.w > 9 && r.h > 9 && (
            <div className="leading-tight text-white drop-shadow">
              <div className="truncate text-[10px] font-medium">{r.key || '—'}</div>
              <div className="font-mono text-[10px] opacity-90">{compact(r.value)}</div>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

function layout(items: TreemapItem[], x: number, y: number, w: number, h: number, total: number): (TreemapItem & { x: number; y: number; w: number; h: number })[] {
  if (items.length === 0) return []
  if (items.length === 1) return [{ ...items[0], x, y, w, h }]
  // split items into two groups of ~equal value
  let half = 0
  let idx = 0
  const target = total / 2
  for (let i = 0; i < items.length; i++) {
    if (half + items[i].value > target && idx > 0) break
    half += items[i].value
    idx = i + 1
  }
  const a = items.slice(0, idx)
  const b = items.slice(idx)
  const aTotal = a.reduce((s, i) => s + i.value, 0)
  const bTotal = total - aTotal
  if (w >= h) {
    const aw = (aTotal / total) * w
    return [...layout(a, x, y, aw, h, aTotal), ...layout(b, x + aw, y, w - aw, h, bTotal)]
  }
  const ah = (aTotal / total) * h
  return [...layout(a, x, y, w, ah, aTotal), ...layout(b, x, y + ah, w, h - ah, bTotal)]
}
