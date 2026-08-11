import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown, Clock } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from './button'

/**
 * Time window for the pages that read the event store. `from`/`to` are values the backend's
 * IS_BETWEEN accepts: date-math (`now-7d`, `now`) or absolute ISO. `interval` is
 * the histogram bucket token. Ported from the legacy elastic-filter-time: quick
 * presets + a relative "last N <unit>" builder + an absolute from/to range.
 */
export interface TimeRange {
  from: string | null
  to: string
  interval: string
}

type Unit = 'minutes' | 'hours' | 'days' | 'weeks' | 'months'
const UNIT_TOKEN: Record<Unit, string> = { minutes: 'm', hours: 'h', days: 'd', weeks: 'w', months: 'M' }
const UNIT_MS: Record<Unit, number> = {
  minutes: 60_000,
  hours: 3_600_000,
  days: 86_400_000,
  weeks: 604_800_000,
  months: 2_592_000_000,
}

const PRESETS: { id: string; token: string; interval: string }[] = [
  { id: '15m', token: 'now-15m', interval: 'minute' },
  { id: '1h', token: 'now-1h', interval: 'minute' },
  { id: '6h', token: 'now-6h', interval: 'hour' },
  { id: '24h', token: 'now-24h', interval: 'hour' },
  { id: '7d', token: 'now-7d', interval: 'day' },
  { id: '30d', token: 'now-30d', interval: 'day' },
]

function intervalForSpan(ms: number): string {
  if (ms <= 2 * 3_600_000) return 'minute' // ≤ 2h
  if (ms <= 2 * 86_400_000) return 'hour' // ≤ 2d
  if (ms <= 60 * 86_400_000) return 'day' // ≤ 60d
  return 'week'
}

/** Build a TimeRange from a quick-preset id. */
export function presetRange(id: string): TimeRange {
  const p = PRESETS.find((x) => x.id === id) ?? PRESETS[3]
  return { from: p.token, to: 'now', interval: p.interval }
}

const REL_RE = /^now-(\d+)([mhdwM])$/

export const ALL_TIME: TimeRange = { from: null, to: 'now', interval: 'day' }

/**
 * Turns a range into the absolute instants a query needs.
 *
 * The picker keeps relative tokens — "now-7d" — because that is what the label
 * and the saved view are about: a window that follows the clock. The event
 * store takes datetimes, not expressions, so the token is resolved here, once,
 * rather than being sent and rejected.
 */
export function resolveRange(r: TimeRange): { from: string | null; to: string } {
  return { from: resolveInstant(r.from), to: resolveInstant(r.to) ?? new Date().toISOString() }
}

function resolveInstant(v: string | null): string | null {
  if (!v) return null
  if (v === 'now') return new Date().toISOString()

  const m = v.match(REL_RE)
  if (!m) return v // already absolute

  const n = Number(m[1])
  const d = new Date()
  switch (m[2]) {
    case 'm': d.setMinutes(d.getMinutes() - n); break
    case 'h': d.setHours(d.getHours() - n); break
    case 'd': d.setDate(d.getDate() - n); break
    case 'w': d.setDate(d.getDate() - n * 7); break
    case 'M': d.setMonth(d.getMonth() - n); break
  }
  return d.toISOString()
}

const TOKEN_UNIT: Record<string, Unit> = { m: 'minutes', h: 'hours', d: 'days', w: 'weeks', M: 'months' }

function isoToLocalInput(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  // datetime-local wants local time without timezone/seconds.
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function fmtAbs(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString(undefined, { month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit' })
}

export function TimeRangePicker({
  value,
  onChange,
  allowAllTime = false,
  align = 'left',
}: {
  value: TimeRange
  onChange: (r: TimeRange) => void
  allowAllTime?: boolean
  align?: 'left' | 'right'
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  // Relative builder local state.
  const [relN, setRelN] = useState(3)
  const [relUnit, setRelUnit] = useState<Unit>('days')
  // Absolute builder local state.
  const [absFrom, setAbsFrom] = useState('')
  const [absTo, setAbsTo] = useState('')

  useEffect(() => {
    const h = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    window.addEventListener('mousedown', h)
    return () => window.removeEventListener('mousedown', h)
  }, [])

  // Seed the absolute inputs from the current value when opening, if it's absolute.
  useEffect(() => {
    if (!open) return
    if (value.from && !REL_RE.test(value.from) && value.from !== 'now') {
      setAbsFrom(isoToLocalInput(value.from))
      setAbsTo(value.to !== 'now' ? isoToLocalInput(value.to) : '')
    }
  }, [open, value.from, value.to])

  const label = deriveLabel(value, t)

  const applyPreset = (id: string) => {
    onChange(presetRange(id))
    setOpen(false)
  }
  const applyRelative = () => {
    const n = Math.max(1, Math.floor(relN || 1))
    onChange({ from: `now-${n}${UNIT_TOKEN[relUnit]}`, to: 'now', interval: intervalForSpan(n * UNIT_MS[relUnit]) })
    setOpen(false)
  }
  const applyAbsolute = () => {
    if (!absFrom) return
    const fromIso = new Date(absFrom).toISOString()
    const toIso = absTo ? new Date(absTo).toISOString() : 'now'
    const span = (absTo ? new Date(absTo).getTime() : Date.now()) - new Date(absFrom).getTime()
    onChange({ from: fromIso, to: toIso, interval: intervalForSpan(Math.max(span, 60_000)) })
    setOpen(false)
  }

  return (
    <div ref={ref} className="relative">
      <Button variant="outline" size="sm" onClick={() => setOpen((v) => !v)} className="gap-1.5">
        <Clock size={14} />
        <span className="max-w-[200px] truncate">{label}</span>
        <ChevronDown size={12} className="opacity-60" />
      </Button>

      {open && (
        <div
          className={cn(
            'absolute z-50 mt-1 w-[320px] rounded-lg border border-border bg-popover p-3 text-popover-foreground shadow-lg',
            align === 'right' ? 'right-0' : 'left-0',
          )}
        >
          {/* Quick presets */}
          <div className="mb-1 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
            {t('timeRange.quick')}
          </div>
          <div className="grid grid-cols-3 gap-1">
            {PRESETS.map((p) => (
              <button
                key={p.id}
                onClick={() => applyPreset(p.id)}
                className={cn(
                  'rounded-md border px-2 py-1 text-xs transition-colors',
                  value.from === p.token ? 'border-primary bg-primary/5 text-foreground' : 'border-border hover:bg-muted',
                )}
              >
                {t(`timeRange.presets.${p.id}`)}
              </button>
            ))}
            {allowAllTime && (
              <button
                onClick={() => {
                  onChange(ALL_TIME)
                  setOpen(false)
                }}
                className={cn(
                  'col-span-3 rounded-md border px-2 py-1 text-xs transition-colors',
                  value.from === null ? 'border-primary bg-primary/5 text-foreground' : 'border-border hover:bg-muted',
                )}
              >
                {t('timeRange.allTime')}
              </button>
            )}
          </div>

          {/* Relative */}
          <div className="mb-1 mt-3 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
            {t('timeRange.relative')}
          </div>
          <div className="flex items-center gap-1.5">
            <span className="text-xs text-muted-foreground">{t('timeRange.last')}</span>
            <input
              type="number"
              min={1}
              value={relN}
              onChange={(e) => setRelN(Number(e.target.value))}
              className="h-8 w-16 rounded-md border border-border bg-background px-2 text-xs"
            />
            <select
              value={relUnit}
              onChange={(e) => setRelUnit(e.target.value as Unit)}
              className="h-8 flex-1 rounded-md border border-border bg-background px-2 text-xs"
            >
              {(Object.keys(UNIT_TOKEN) as Unit[]).map((u) => (
                <option key={u} value={u}>
                  {t(`timeRange.units.${u}`)}
                </option>
              ))}
            </select>
            <Button size="sm" onClick={applyRelative}>
              {t('timeRange.apply')}
            </Button>
          </div>

          {/* Absolute */}
          <div className="mb-1 mt-3 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
            {t('timeRange.absolute')}
          </div>
          <div className="space-y-1.5">
            <label className="flex items-center gap-2 text-xs">
              <span className="w-10 text-muted-foreground">{t('timeRange.from')}</span>
              <input
                type="datetime-local"
                value={absFrom}
                onChange={(e) => setAbsFrom(e.target.value)}
                className="h-8 flex-1 rounded-md border border-border bg-background px-2 text-xs"
              />
            </label>
            <label className="flex items-center gap-2 text-xs">
              <span className="w-10 text-muted-foreground">{t('timeRange.to')}</span>
              <input
                type="datetime-local"
                value={absTo}
                onChange={(e) => setAbsTo(e.target.value)}
                className="h-8 flex-1 rounded-md border border-border bg-background px-2 text-xs"
              />
            </label>
            <div className="flex justify-end">
              <Button size="sm" disabled={!absFrom} onClick={applyAbsolute}>
                {t('timeRange.apply')}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function deriveLabel(r: TimeRange, t: ReturnType<typeof useTranslation>['t']): string {
  if (r.from === null) return t('timeRange.allTime')
  if (r.to === 'now' && r.from) {
    const preset = PRESETS.find((p) => p.token === r.from)
    if (preset) return t(`timeRange.presets.${preset.id}`)
    const m = r.from.match(REL_RE)
    if (m) {
      const n = Number(m[1])
      const unit = TOKEN_UNIT[m[2]]
      return `${t('timeRange.last')} ${n} ${t(`timeRange.units.${unit}`)}`
    }
  }
  // Absolute
  if (r.from) return `${fmtAbs(r.from)} → ${r.to === 'now' ? t('timeRange.now') : fmtAbs(r.to)}`
  return t('timeRange.allTime')
}
