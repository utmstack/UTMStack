import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useDateFormat } from '@/shared/lib/datetime'
import { complianceService } from '../services/compliance-http.service'
import type { ScorePoint } from '../types/compliance.types'

/**
 * Score over time for one framework.
 *
 * Two series on one axis, both percentages, because the score on its own can
 * mislead: it rises when requirements leave the denominator exactly as it does
 * when they start passing. Plotting coverage beside it makes the difference
 * visible — a score climbing while coverage falls is a source going quiet, not
 * a control being fixed.
 *
 * Both are shares of the same whole, so they share one scale. Two y-axes would
 * let any pair of lines be made to cross wherever the reader's eye wanted.
 */

// Validated with the dataviz palette checker against both surfaces: lightness
// band, chroma floor, CVD separation (ΔE 31 protan / 25 tritan) and contrast.
// score    light #0075f0  dark #2f90ee
// coverage light #a05a00  dark #c07f1f

const W = 720
const H = 190
const PAD = { l: 34, r: 14, t: 12, b: 26 }
const innerW = W - PAD.l - PAD.r
const innerH = H - PAD.t - PAD.b

export function ScoreHistory({ frameworkKey }: { frameworkKey: string }) {
  const { t } = useTranslation()
  const df = useDateFormat()
  const [points, setPoints] = useState<ScorePoint[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [hover, setHover] = useState<number | null>(null)
  const svgRef = useRef<SVGSVGElement>(null)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(false)
    complianceService
      .history(frameworkKey)
      .then((p) => !cancelled && setPoints(p ?? []))
      .catch(() => !cancelled && setError(true))
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [frameworkKey])

  const series = useMemo(() => {
    return points.map((p, i) => ({
      ...p,
      // Coverage is what share of the framework could be judged at all.
      coverage: p.total > 0 ? Math.round((p.evaluated / p.total) * 100) : 0,
      x: points.length === 1 ? PAD.l + innerW / 2 : PAD.l + (i * innerW) / (points.length - 1),
    }))
  }, [points])

  if (loading) {
    return <div className="h-[190px] animate-pulse rounded-xl bg-muted" />
  }
  if (error) {
    return (
      <div className="flex h-[190px] items-center justify-center gap-2 text-xs text-muted-foreground">
        <AlertTriangle size={14} className="text-amber-500" />
        {t('compliance.history.error', { defaultValue: 'Could not load the score history' })}
      </div>
    )
  }
  // One point is not a trend, and an empty state explaining that is a box of
  // text where a chart should be. With nothing to plot the section is simply
  // absent, and the report starts at its findings.
  if (series.length < 2) {
    return null
  }

  // The axis is 0–100 for both series: they are shares of the same whole, and a
  // fitted scale would exaggerate a two-point wobble into a cliff.
  const y = (v: number) => PAD.t + innerH - (v / 100) * innerH
  const path = (pick: (p: (typeof series)[number]) => number) =>
    series.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${y(pick(p))}`).join(' ')

  const onMove = (e: React.MouseEvent<SVGSVGElement>) => {
    const rect = svgRef.current?.getBoundingClientRect()
    if (!rect) return
    const px = ((e.clientX - rect.left) / rect.width) * W
    let best = 0
    for (let i = 1; i < series.length; i++) {
      if (Math.abs(series[i].x - px) < Math.abs(series[best].x - px)) best = i
    }
    setHover(best)
  }

  const at = hover != null ? series[hover] : null
  const labelEvery = Math.max(1, Math.ceil(series.length / 6))

  return (
    <div className="rounded-xl border border-border bg-card p-4">
      <div className="flex flex-wrap items-baseline justify-between gap-3">
        <h3 className="text-sm font-medium">{t('compliance.history.title', { defaultValue: 'Score over time' })}</h3>
        {/* Two series, so a legend is always present — identity is never left
            to colour alone. */}
        <div className="flex items-center gap-4 text-[11px] text-muted-foreground">
          <Key label={t('compliance.history.score', { defaultValue: 'Compliance score' })} />
          <Key label={t('compliance.history.coverage', { defaultValue: 'Evaluated' })} dashed />
        </div>
      </div>

      <div className="relative mt-2">
        <svg
          ref={svgRef}
          viewBox={`0 0 ${W} ${H}`}
          className="h-[190px] w-full"
          onMouseMove={onMove}
          onMouseLeave={() => setHover(null)}
        >
          <g className="fill-muted-foreground" fontSize="9">
            {[0, 50, 100].map((v) => (
              <g key={v}>
                <line
                  x1={PAD.l}
                  x2={W - PAD.r}
                  y1={y(v)}
                  y2={y(v)}
                  className="stroke-border"
                  strokeWidth="1"
                  strokeDasharray={v === 0 ? undefined : '3 3'}
                />
                <text x={PAD.l - 6} y={y(v) + 3} textAnchor="end">
                  {v}%
                </text>
              </g>
            ))}
          </g>

          <path d={path((p) => p.coverage)} fill="none" strokeWidth="2" strokeDasharray="5 3"
            className="stroke-[var(--cov)] [--cov:#a05a00] dark:[--cov:#c07f1f]" strokeLinecap="round" />
          <path d={path((p) => p.score)} fill="none" strokeWidth="2"
            className="stroke-[var(--sc)] [--sc:#0075f0] dark:[--sc:#2f90ee]" strokeLinecap="round" />

          {at && (
            <g>
              <line x1={at.x} x2={at.x} y1={PAD.t} y2={PAD.t + innerH} className="stroke-border" strokeWidth="1" />
              {/* A 2px surface ring keeps the marker readable wherever it lands
                  on the line beneath it. */}
              <circle cx={at.x} cy={y(at.coverage)} r="4.5" className="fill-card stroke-[var(--cov)] [--cov:#a05a00] dark:[--cov:#c07f1f]" strokeWidth="2" />
              <circle cx={at.x} cy={y(at.score)} r="4.5" className="fill-card stroke-[var(--sc)] [--sc:#0075f0] dark:[--sc:#2f90ee]" strokeWidth="2" />
            </g>
          )}

          <g className="fill-muted-foreground" fontSize="9">
            {series.map((p, i) =>
              // The last date is always labelled, so a periodic one landing
              // beside it is dropped rather than printed on top of it.
              (i % labelEvery === 0 && i < series.length - 1 - labelEvery / 2) || i === series.length - 1 ? (
                <text key={p.day} x={p.x} y={H - 8} textAnchor={i === 0 ? 'start' : i === series.length - 1 ? 'end' : 'middle'}>
                  {df.formatDate(p.day)}
                </text>
              ) : null,
            )}
          </g>
        </svg>

        {at && (
          <div
            className="pointer-events-none absolute top-2 rounded-md border border-border bg-popover px-2.5 py-1.5 text-[11px] shadow-md"
            style={{ left: `${Math.min(Math.max((at.x / W) * 100, 8), 76)}%` }}
          >
            <div className="font-medium">{df.formatDate(at.day)}</div>
            <div className="mt-1 flex items-center gap-1.5">
              <Dot score />
              <span className="text-muted-foreground">{t('compliance.history.score', { defaultValue: 'Compliance score' })}</span>
              <span className="ml-auto font-semibold tabular-nums">{at.score}%</span>
            </div>
            <div className="flex items-center gap-1.5">
              <Dot />
              <span className="text-muted-foreground">{t('compliance.history.coverage', { defaultValue: 'Evaluated' })}</span>
              <span className="ml-auto font-semibold tabular-nums">
                {at.evaluated}/{at.total}
              </span>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function Key({ label, dashed }: { label: string; dashed?: boolean }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <svg width="16" height="2" aria-hidden>
        <line
          x1="0"
          y1="1"
          x2="16"
          y2="1"
          strokeWidth="2"
          strokeDasharray={dashed ? '4 2' : undefined}
          className={cn('stroke-[var(--k)]', dashed ? '[--k:#a05a00] dark:[--k:#c07f1f]' : '[--k:#0075f0] dark:[--k:#2f90ee]')}
        />
      </svg>
      {label}
    </span>
  )
}

function Dot({ score }: { score?: boolean }) {
  return (
    <span
      className={cn(
        'h-2 w-2 shrink-0 rounded-full bg-[var(--d)]',
        score ? '[--d:#0075f0] dark:[--d:#2f90ee]' : '[--d:#a05a00] dark:[--d:#c07f1f]',
      )}
    />
  )
}
