import { useRef, useState } from 'react'
import { CalendarDays, ChevronDown, ChevronUp, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { FloatingPanel } from '@/shared/components/ui/floating-panel'

export type TimePreset = 'all' | '15m' | '30m' | '1h' | '24h' | '7d' | '30d' | '90d' | '1y'
export type TimeFilter = { kind: 'preset'; preset: TimePreset } | { kind: 'custom'; from: string; to: string }

const PRESET_MINUTES: Record<Exclude<TimePreset, 'all'>, number> = {
  '15m': 15,
  '30m': 30,
  '1h': 60,
  '24h': 1440,
  '7d': 10080,
  '30d': 43200,
  '90d': 129600,
  '1y': 525600,
}
const PRESET_KEYS: Record<TimePreset, string> = {
  all: 'all',
  '15m': 'last15m',
  '1h': 'last1h',
  '7d': 'last7d',
  '90d': 'last90d',
  '30m': 'last30m',
  '24h': 'last24h',
  '30d': 'last30d',
  '1y': 'last1y',
}
const PRESET_ORDER: Exclude<TimePreset, 'all'>[] = ['15m', '30m', '1h', '24h', '7d', '30d', '90d', '1y']

export function resolveTimeBounds(time: TimeFilter): { from: number; to: number } {
  if (time.kind === 'custom') {
    const from = Date.parse(time.from)
    const to = Date.parse(time.to)
    return { from: Number.isNaN(from) ? 0 : from, to: Number.isNaN(to) ? Infinity : to }
  }
  if (time.preset === 'all') return { from: 0, to: Infinity }
  const now = Date.now()
  return { from: now - PRESET_MINUTES[time.preset] * 60_000, to: Infinity }
}

interface PanelTime {
  date: string
  h: number
  m: number
  s: number
}

function todayString(): string {
  return new Date().toISOString().slice(0, 10)
}

function parsePanelTime(iso: string): PanelTime {
  return {
    date: iso.slice(0, 10),
    h: Number(iso.slice(11, 13)) || 0,
    m: Number(iso.slice(14, 16)) || 0,
    s: Number(iso.slice(17, 19)) || 0,
  }
}

function panelToIso(panel: PanelTime): string {
  const pad = (value: number) => String(value).padStart(2, '0')
  return `${panel.date}T${pad(panel.h)}:${pad(panel.m)}:${pad(panel.s)}`
}

export function TimeRangeFilter({
  value,
  onChange,
}: {
  value: TimeFilter
  onChange: (time: TimeFilter) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [view, setView] = useState<'preset' | 'custom'>('preset')
  const [from, setFrom] = useState<PanelTime>({ date: todayString(), h: 0, m: 0, s: 0 })
  const [to, setTo] = useState<PanelTime>({ date: todayString(), h: 23, m: 59, s: 59 })
  const triggerRef = useRef<HTMLButtonElement>(null)
  const fromInputRef = useRef<HTMLInputElement>(null)
  const toInputRef = useRef<HTMLInputElement>(null)

  const label = value.kind === 'custom'
    ? t('darkWebMonitoring.time.custom')
    : t(`darkWebMonitoring.time.${PRESET_KEYS[value.preset]}`)

  const toggle = () => setOpen((current) => !current)

  const openCustom = () => {
    if (value.kind === 'custom') {
      setFrom(parsePanelTime(value.from))
      setTo(parsePanelTime(value.to))
    } else {
      setFrom({ date: todayString(), h: 0, m: 0, s: 0 })
      setTo({ date: todayString(), h: 23, m: 59, s: 59 })
    }
    setView('custom')
  }

  const fromIso = panelToIso(from)
  const toIso = panelToIso(to)
  const invalid = Date.parse(fromIso) > Date.parse(toIso)

  const pickPreset = (preset: TimePreset) => {
    onChange({ kind: 'preset', preset })
    setOpen(false)
    setView('preset')
  }

  const openDatePicker = (input: HTMLInputElement | null) => {
    if (!input) return
    if (typeof input.showPicker === 'function') {
      try {
        input.showPicker()
      } catch {
        input.focus()
      }
    } else {
      input.focus()
    }
  }

  return (
    <>
      <button
        ref={triggerRef}
        type="button"
        aria-haspopup="menu"
        aria-expanded={open}
        onClick={toggle}
        className={`flex h-9 items-center gap-2 rounded-lg border border-border px-2.5 text-xs transition-colors hover:bg-muted/40 ${open ? 'bg-background text-foreground' : 'text-muted-foreground'}`}
      >
        <CalendarDays size={13} />
        <span className="whitespace-nowrap">{label}</span>
        <ChevronDown size={12} />
      </button>
      <FloatingPanel
        open={open}
        onClose={() => {
          setOpen(false)
          setView('preset')
        }}
        triggerRef={triggerRef}
        className="z-[95] w-80 max-h-[calc(100vh-16px)] overflow-y-auto rounded-xl border border-border bg-popover p-3 shadow-2xl"
      >
          {view === 'preset' ? (
            <div>
              <div className="mb-2 flex items-center gap-1.5 text-xs font-semibold text-foreground">
                <ChevronDown size={12} className="text-muted-foreground" />
                {t('darkWebMonitoring.time.commonlyUsed')}
              </div>
              <div className="grid grid-cols-2 gap-x-6 gap-y-1">
                {PRESET_ORDER.map((preset) => (
                  <button
                    key={preset}
                    type="button"
                    onClick={() => pickPreset(preset)}
                    className={`rounded-md px-2 py-1.5 text-left font-mono text-xs transition-colors hover:bg-muted ${value.kind === 'preset' && value.preset === preset ? 'text-foreground' : 'text-muted-foreground'}`}
                  >
                    {t(`darkWebMonitoring.time.${PRESET_KEYS[preset]}`)}
                  </button>
                ))}
                <button
                  type="button"
                  onClick={() => pickPreset('all')}
                  className={`rounded-md px-2 py-1.5 text-left font-mono text-xs transition-colors hover:bg-muted ${value.kind === 'preset' && value.preset === 'all' ? 'text-foreground' : 'text-muted-foreground'}`}
                >
                  {t('darkWebMonitoring.time.all')}
                </button>
              </div>
              <div className="my-2 border-t border-border" />
              <button
                type="button"
                onClick={openCustom}
                className="flex items-center gap-1.5 text-xs font-semibold text-foreground transition-colors hover:text-primary"
              >
                <ChevronDown size={12} className="text-muted-foreground" />
                {t('darkWebMonitoring.time.custom')}
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              <PanelTime
                label={t('darkWebMonitoring.time.timeFrom')}
                panel={from}
                onChange={setFrom}
                inputRef={fromInputRef}
                onPickDate={() => openDatePicker(fromInputRef.current)}
              />
              <PanelTime
                label={t('darkWebMonitoring.time.timeTo')}
                panel={to}
                onChange={setTo}
                inputRef={toInputRef}
                onPickDate={() => openDatePicker(toInputRef.current)}
              />
              {invalid && (
                <div className="text-xs font-semibold text-destructive">{t('darkWebMonitoring.time.invalidRange')}</div>
              )}
              <div className="flex gap-2 pt-1">
                <button
                  type="button"
                  onClick={() => setView('preset')}
                  className="flex flex-1 items-center justify-center gap-1.5 rounded-lg border border-border px-3 py-2 text-xs font-semibold text-foreground transition-colors hover:bg-muted"
                >
                  <X size={13} />
                  {t('darkWebMonitoring.time.cancel')}
                </button>
                <button
                  type="button"
                  disabled={invalid}
                  onClick={() => {
                    onChange({ kind: 'custom', from: fromIso, to: toIso })
                    setOpen(false)
                    setView('preset')
                  }}
                  className="flex flex-1 items-center justify-center gap-1.5 rounded-lg bg-primary px-3 py-2 text-xs font-semibold text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-40"
                >
                  <CalendarDays size={13} />
                  {t('darkWebMonitoring.time.select')}
                </button>
              </div>
            </div>
          )}
      </FloatingPanel>
    </>
  )
}

function PanelTime({
  label,
  panel,
  onChange,
  inputRef,
  onPickDate,
}: {
  label: string
  panel: PanelTime
  onChange: (panel: PanelTime) => void
  inputRef: React.RefObject<HTMLInputElement | null>
  onPickDate: () => void
}) {
  const set = (patch: Partial<PanelTime>) => onChange({ ...panel, ...patch })
  return (
    <div>
      <div className="mb-1.5 text-xs font-bold text-foreground">{label}</div>
      <div className="flex">
        <input
          ref={inputRef}
          type="date"
          value={panel.date}
          onChange={(event) => set({ date: event.target.value })}
          className="dwm-date-input h-9 min-w-0 flex-1 rounded-l-lg border border-r-0 border-border bg-background px-2.5 font-mono text-xs text-foreground outline-none"
        />
        <button
          type="button"
          tabIndex={-1}
          onClick={onPickDate}
          className="flex h-9 w-10 shrink-0 items-center justify-center rounded-r-lg bg-primary/15 text-primary transition-colors hover:bg-primary/25"
        >
          <CalendarDays size={14} />
        </button>
      </div>
      <div className="mt-2 flex items-start justify-center gap-3">
        <TimeSpinner value={panel.h} max={24} onChange={(v) => set({ h: v })} />
        <span className="pt-3.5 text-xs text-muted-foreground">:</span>
        <TimeSpinner value={panel.m} max={60} onChange={(v) => set({ m: v })} />
        <span className="pt-3.5 text-xs text-muted-foreground">:</span>
        <TimeSpinner value={panel.s} max={60} onChange={(v) => set({ s: v })} />
      </div>
    </div>
  )
}

function TimeSpinner({ value, max, onChange }: { value: number; max: number; onChange: (value: number) => void }) {
  const step = (delta: number) => onChange((((value + delta) % max) + max) % max)
  return (
    <div className="flex flex-col items-center gap-1">
      <button type="button" aria-label="↑" onClick={() => step(1)} className="rounded p-0.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">
        <ChevronUp size={13} />
      </button>
      <div className="w-11 rounded-lg bg-background py-1.5 text-center font-mono text-sm font-semibold text-foreground ring-1 ring-border">
        {String(value).padStart(2, '0')}
      </div>
      <button type="button" aria-label="↓" onClick={() => step(-1)} className="rounded p-0.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">
        <ChevronDown size={13} />
      </button>
    </div>
  )
}
