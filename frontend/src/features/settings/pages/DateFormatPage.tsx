import { useMemo, useState } from 'react'
import {
  Calendar,
  CalendarClock,
  Check,
  Clock,
  Globe2,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type DateStyle = 'short' | 'medium' | 'long'
type TimeStyle = '24h' | '12h'
type WeekStart = 'mon' | 'sun'

/* ─── Timezone catalog (shortened — typical popular options) ──────────── */

interface TimezoneOption {
  id: string
  label: string
  region: string
  offset: string
}

const TIMEZONES: TimezoneOption[] = [
  { id: 'system', label: 'Browser default', region: 'Auto-detect', offset: 'auto' },
  { id: 'UTC', label: 'UTC', region: 'Coordinated Universal Time', offset: '+00:00' },
  { id: 'America/Los_Angeles', label: 'Pacific (Los Angeles)', region: 'Americas', offset: '-08:00' },
  { id: 'America/Denver', label: 'Mountain (Denver)', region: 'Americas', offset: '-07:00' },
  { id: 'America/Chicago', label: 'Central (Chicago)', region: 'Americas', offset: '-06:00' },
  { id: 'America/New_York', label: 'Eastern (New York)', region: 'Americas', offset: '-05:00' },
  { id: 'America/Sao_Paulo', label: 'São Paulo', region: 'Americas', offset: '-03:00' },
  { id: 'Europe/London', label: 'London', region: 'Europe', offset: '+00:00' },
  { id: 'Europe/Madrid', label: 'Madrid', region: 'Europe', offset: '+01:00' },
  { id: 'Europe/Berlin', label: 'Berlin', region: 'Europe', offset: '+01:00' },
  { id: 'Europe/Athens', label: 'Athens', region: 'Europe', offset: '+02:00' },
  { id: 'Africa/Cairo', label: 'Cairo', region: 'Africa', offset: '+02:00' },
  { id: 'Asia/Dubai', label: 'Dubai', region: 'Asia', offset: '+04:00' },
  { id: 'Asia/Karachi', label: 'Karachi', region: 'Asia', offset: '+05:00' },
  { id: 'Asia/Kolkata', label: 'Kolkata', region: 'Asia', offset: '+05:30' },
  { id: 'Asia/Singapore', label: 'Singapore', region: 'Asia', offset: '+08:00' },
  { id: 'Asia/Shanghai', label: 'Shanghai', region: 'Asia', offset: '+08:00' },
  { id: 'Asia/Tokyo', label: 'Tokyo', region: 'Asia', offset: '+09:00' },
  { id: 'Australia/Sydney', label: 'Sydney', region: 'Oceania', offset: '+11:00' },
]

/* ─── Helpers (mock formatting based on choices) ───────────────────────── */

const SAMPLE_DATE = new Date('2026-05-02T14:32:18Z')

function formatPreview({
  dateStyle,
  timeStyle,
  weekStart,
  timezone,
}: {
  dateStyle: DateStyle
  timeStyle: TimeStyle
  weekStart: WeekStart
  timezone: string
}) {
  const tz = timezone === 'system' ? undefined : timezone
  const time = new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: timeStyle === '12h',
    timeZone: tz,
    timeZoneName: 'short',
  }).format(SAMPLE_DATE)
  const date =
    dateStyle === 'short'
      ? new Intl.DateTimeFormat(undefined, { dateStyle: 'short', timeZone: tz }).format(SAMPLE_DATE)
      : dateStyle === 'medium'
        ? new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeZone: tz }).format(SAMPLE_DATE)
        : new Intl.DateTimeFormat(undefined, { dateStyle: 'full', timeZone: tz }).format(SAMPLE_DATE)
  return { date, time, weekStartLabel: weekStart === 'mon' ? 'Monday' : 'Sunday' }
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function DateFormatPage() {
  const [timezone, setTimezone] = useState('system')
  const [dateStyle, setDateStyle] = useState<DateStyle>('medium')
  const [timeStyle, setTimeStyle] = useState<TimeStyle>('24h')
  const [weekStart, setWeekStart] = useState<WeekStart>('mon')
  const [relativeWhenRecent, setRelativeWhenRecent] = useState(true)
  const [dirty, setDirty] = useState(false)

  const update = <T,>(setter: (v: T) => void) => (v: T) => {
    setter(v)
    setDirty(true)
  }

  const preview = useMemo(
    () => formatPreview({ dateStyle, timeStyle, weekStart, timezone }),
    [dateStyle, timeStyle, weekStart, timezone]
  )

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header dirty={dirty} onSave={() => setDirty(false)} onDiscard={() => setDirty(false)} />

      <div className="mt-6 grid grid-cols-1 gap-5 lg:grid-cols-[1fr_400px]">
        <div className="space-y-5">
          <TimezoneSection timezone={timezone} onChange={update(setTimezone)} />
          <FormatSection
            dateStyle={dateStyle}
            timeStyle={timeStyle}
            weekStart={weekStart}
            onDateStyle={update(setDateStyle)}
            onTimeStyle={update(setTimeStyle)}
            onWeekStart={update(setWeekStart)}
          />
          <RelativeSection enabled={relativeWhenRecent} onChange={update(setRelativeWhenRecent)} />
        </div>

        <aside className="lg:sticky lg:top-4 lg:h-fit">
          <PreviewSection
            preview={preview}
            timezone={timezone}
            dateStyle={dateStyle}
            timeStyle={timeStyle}
            relative={relativeWhenRecent}
          />
        </aside>
      </div>
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ dirty, onSave, onDiscard }: { dirty: boolean; onSave: () => void; onDiscard: () => void }) {
  return (
    <header className="flex items-end justify-between gap-3">
      <div>
        <h1 className="flex items-center gap-2 text-xl font-semibold">
          <CalendarClock size={18} strokeWidth={1.75} />
          Date format
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          How dates and times are rendered for your account. Stored data is always in UTC; this only
          affects display.
        </p>
      </div>
      {dirty && (
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={onDiscard}>
            Discard
          </Button>
          <Button size="sm" onClick={onSave}>
            Save changes
          </Button>
        </div>
      )}
    </header>
  )
}

/* ─── Timezone section ─────────────────────────────────────────────────── */

function TimezoneSection({
  timezone,
  onChange,
}: {
  timezone: string
  onChange: (tz: string) => void
}) {
  const [search, setSearch] = useState('')
  const filtered = TIMEZONES.filter((t) =>
    search ? (t.label + t.region + t.offset + t.id).toLowerCase().includes(search.toLowerCase()) : true
  )
  const grouped = filtered.reduce<Record<string, TimezoneOption[]>>((acc, t) => {
    acc[t.region] = acc[t.region] ?? []
    acc[t.region].push(t)
    return acc
  }, {})

  return (
    <Section title="Timezone" subtitle="Used to render every timestamp in the panel.">
      <Input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Search city, region, offset…"
        className="h-9"
      />
      <div className="mt-3 max-h-72 space-y-3 overflow-y-auto rounded-md border border-border bg-background/40 p-2">
        {Object.entries(grouped).map(([region, items]) => (
          <div key={region}>
            <div className="px-2 py-1 text-[10px] uppercase tracking-wider text-muted-foreground">
              {region}
            </div>
            <div className="space-y-0.5">
              {items.map((t) => {
                const active = timezone === t.id
                return (
                  <button
                    key={t.id}
                    onClick={() => onChange(t.id)}
                    className={cn(
                      'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-xs transition-colors',
                      active ? 'bg-primary/10 text-primary' : 'hover:bg-muted/60'
                    )}
                  >
                    <Globe2 size={12} className={active ? 'text-primary' : 'text-muted-foreground'} />
                    <span className="flex-1 truncate">{t.label}</span>
                    <span className="font-mono text-[10px] text-muted-foreground">{t.offset}</span>
                    {active && <Check size={11} strokeWidth={3} className="text-primary" />}
                  </button>
                )
              })}
            </div>
          </div>
        ))}
        {filtered.length === 0 && (
          <div className="px-3 py-6 text-center text-[11px] italic text-muted-foreground">
            No timezone matches that search.
          </div>
        )}
      </div>
    </Section>
  )
}

/* ─── Format section ───────────────────────────────────────────────────── */

function FormatSection({
  dateStyle,
  timeStyle,
  weekStart,
  onDateStyle,
  onTimeStyle,
  onWeekStart,
}: {
  dateStyle: DateStyle
  timeStyle: TimeStyle
  weekStart: WeekStart
  onDateStyle: (s: DateStyle) => void
  onTimeStyle: (s: TimeStyle) => void
  onWeekStart: (s: WeekStart) => void
}) {
  return (
    <Section title="Format" subtitle="How dates and times look.">
      {/* Date style */}
      <div className="mb-5">
        <label className="mb-2 block text-[11px] uppercase tracking-wider text-muted-foreground">
          Date style
        </label>
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
          <FormatCard
            id="short"
            label="Short"
            sample={new Intl.DateTimeFormat(undefined, { dateStyle: 'short' }).format(SAMPLE_DATE)}
            icon={Calendar}
            active={dateStyle === 'short'}
            onClick={() => onDateStyle('short')}
          />
          <FormatCard
            id="medium"
            label="Medium"
            sample={new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(SAMPLE_DATE)}
            icon={Calendar}
            active={dateStyle === 'medium'}
            onClick={() => onDateStyle('medium')}
          />
          <FormatCard
            id="long"
            label="Long"
            sample={new Intl.DateTimeFormat(undefined, { dateStyle: 'full' }).format(SAMPLE_DATE)}
            icon={Calendar}
            active={dateStyle === 'long'}
            onClick={() => onDateStyle('long')}
          />
        </div>
      </div>

      {/* Time style */}
      <div className="mb-5">
        <label className="mb-2 block text-[11px] uppercase tracking-wider text-muted-foreground">
          Time style
        </label>
        <div className="grid grid-cols-2 gap-2">
          <FormatCard
            id="24h"
            label="24-hour"
            sample={new Intl.DateTimeFormat(undefined, {
              hour: '2-digit',
              minute: '2-digit',
              second: '2-digit',
              hour12: false,
            }).format(SAMPLE_DATE)}
            icon={Clock}
            active={timeStyle === '24h'}
            onClick={() => onTimeStyle('24h')}
          />
          <FormatCard
            id="12h"
            label="12-hour"
            sample={new Intl.DateTimeFormat(undefined, {
              hour: '2-digit',
              minute: '2-digit',
              second: '2-digit',
              hour12: true,
            }).format(SAMPLE_DATE)}
            icon={Clock}
            active={timeStyle === '12h'}
            onClick={() => onTimeStyle('12h')}
          />
        </div>
      </div>

      {/* Week start */}
      <div>
        <label className="mb-2 block text-[11px] uppercase tracking-wider text-muted-foreground">
          Week starts on
        </label>
        <div className="grid grid-cols-2 gap-2">
          <button
            onClick={() => onWeekStart('mon')}
            className={cn(
              'rounded-lg border px-3 py-2 text-xs transition-colors',
              weekStart === 'mon'
                ? 'border-primary/40 bg-primary/5 font-medium text-primary'
                : 'border-border hover:bg-muted/40'
            )}
          >
            Monday
          </button>
          <button
            onClick={() => onWeekStart('sun')}
            className={cn(
              'rounded-lg border px-3 py-2 text-xs transition-colors',
              weekStart === 'sun'
                ? 'border-primary/40 bg-primary/5 font-medium text-primary'
                : 'border-border hover:bg-muted/40'
            )}
          >
            Sunday
          </button>
        </div>
      </div>
    </Section>
  )
}

function FormatCard({
  label,
  sample,
  icon: Icon,
  active,
  onClick,
}: {
  id: string
  label: string
  sample: string
  icon: LucideIcon
  active: boolean
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'flex flex-col items-start gap-1 rounded-lg border p-3 text-left transition-colors',
        active ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/40'
      )}
    >
      <div className="flex w-full items-center gap-2">
        <Icon size={12} className={active ? 'text-primary' : 'text-muted-foreground'} />
        <span className={cn('text-xs font-semibold', active && 'text-primary')}>{label}</span>
        {active && <Check size={11} strokeWidth={3} className="ml-auto text-primary" />}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">{sample}</div>
    </button>
  )
}

/* ─── Relative time section ────────────────────────────────────────────── */

function RelativeSection({ enabled, onChange }: { enabled: boolean; onChange: (v: boolean) => void }) {
  return (
    <Section
      title="Relative timestamps"
      subtitle="When recent, show times like “2m ago” instead of an absolute date."
    >
      <div className="flex items-start gap-3 py-1">
        <CalendarClock size={14} className="mt-0.5 shrink-0 text-muted-foreground" />
        <div className="min-w-0 flex-1">
          <div className="text-sm font-medium">Use relative time within 24 hours</div>
          <div className="text-xs text-muted-foreground">
            Lists like alerts, audit log, and events show <span className="font-mono">2m</span>,{' '}
            <span className="font-mono">14h</span>, etc. Older than 24 hours always show a full
            timestamp.
          </div>
        </div>
        <Toggle enabled={enabled} onChange={() => onChange(!enabled)} />
      </div>
    </Section>
  )
}

/* ─── Preview ──────────────────────────────────────────────────────────── */

function PreviewSection({
  preview,
  timezone,
  dateStyle,
  timeStyle,
  relative,
}: {
  preview: { date: string; time: string; weekStartLabel: string }
  timezone: string
  dateStyle: DateStyle
  timeStyle: TimeStyle
  relative: boolean
}) {
  const tz = TIMEZONES.find((t) => t.id === timezone)
  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card">
      <div className="border-b border-border px-5 py-3">
        <div className="text-[11px] uppercase tracking-wider text-muted-foreground">Preview</div>
        <div className="mt-0.5 text-sm font-semibold">Live</div>
      </div>

      <div className="space-y-4 p-5">
        {/* Big timestamp */}
        <div className="rounded-lg border border-border bg-background/40 p-4">
          <div className="text-[10px] uppercase tracking-wider text-muted-foreground">
            Sample timestamp
          </div>
          <div className="mt-2 font-mono text-2xl font-semibold tabular-nums">
            {preview.date}
          </div>
          <div className="mt-1 font-mono text-base tabular-nums text-muted-foreground">
            {preview.time}
          </div>
          <div className="mt-3 flex items-center gap-1.5 text-[11px] text-muted-foreground">
            <Globe2 size={11} />
            <span>{tz?.label ?? timezone}</span>
            {tz && <span className="font-mono">· UTC{tz.offset === 'auto' ? ' (auto)' : tz.offset}</span>}
          </div>
        </div>

        {/* In-context examples */}
        <div className="rounded-lg border border-border">
          <div className="border-b border-border px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
            How it appears in the panel
          </div>
          <div className="space-y-2.5 p-3">
            <div className="flex items-start gap-3 text-xs">
              <span className="h-3 w-1 shrink-0 rounded-full bg-red-500" />
              <div className="min-w-0 flex-1">
                <div className="truncate">Suspected Kerberoasting against domain controller</div>
                <div className="mt-0.5 text-[11px] text-muted-foreground">
                  {relative ? '2m ago' : preview.time} ·{' '}
                  <span className="font-mono">WIN-7F3K2L9Q8X</span>
                </div>
              </div>
              <span className="shrink-0 font-mono text-[11px] text-muted-foreground">
                {relative ? '2m' : '14:32'}
              </span>
            </div>
            <div className="flex items-start gap-3 text-xs">
              <span className="h-3 w-1 shrink-0 rounded-full bg-orange-500" />
              <div className="min-w-0 flex-1">
                <div className="truncate">Multiple failed logins followed by success</div>
                <div className="mt-0.5 text-[11px] text-muted-foreground">
                  {relative ? '14h ago' : preview.date} ·{' '}
                  <span className="font-mono">mail-srv-02</span>
                </div>
              </div>
              <span className="shrink-0 font-mono text-[11px] text-muted-foreground">
                {relative ? '14h' : '00:32'}
              </span>
            </div>
            <div className="flex items-start gap-3 text-xs">
              <span className="h-3 w-1 shrink-0 rounded-full bg-amber-500" />
              <div className="min-w-0 flex-1">
                <div className="truncate">GPO modification by non-administrator</div>
                <div className="mt-0.5 text-[11px] text-muted-foreground">
                  {preview.date}, {preview.time}
                </div>
              </div>
              <span className="shrink-0 font-mono text-[11px] text-muted-foreground">3d</span>
            </div>
          </div>
        </div>

        <div className="text-[11px] text-muted-foreground">
          <span className="font-medium text-foreground">{dateStyle}</span> date ·{' '}
          <span className="font-medium text-foreground">
            {timeStyle === '24h' ? '24-hour' : '12-hour'}
          </span>{' '}
          time · week starts <span className="font-medium text-foreground">{preview.weekStartLabel}</span>
        </div>
      </div>
    </section>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({
  title,
  subtitle,
  children,
}: {
  title: string
  subtitle?: string
  children: React.ReactNode
}) {
  return (
    <section className="rounded-xl border border-border bg-card p-6">
      <header className="mb-4">
        <h2 className="text-sm font-semibold">{title}</h2>
        {subtitle && <p className="mt-0.5 text-xs text-muted-foreground">{subtitle}</p>}
      </header>
      {children}
    </section>
  )
}

function Toggle({ enabled, onChange }: { enabled: boolean; onChange: () => void }) {
  return (
    <button
      onClick={onChange}
      className={cn(
        'relative h-5 w-9 shrink-0 rounded-full transition-colors',
        enabled ? 'bg-primary' : 'bg-muted'
      )}
    >
      <span
        className={cn(
          'absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-all',
          enabled ? 'left-4.5' : 'left-0.5'
        )}
      />
    </button>
  )
}
