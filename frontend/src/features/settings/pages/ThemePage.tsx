import { useState } from 'react'
import {
  Check,
  Eye,
  Monitor,
  Moon,
  Palette,
  Sun,
  Type,
  Zap,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useThemeContext } from '@/app/providers'

type ThemeMode = 'light' | 'dark' | 'system'

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function ThemePage() {
  const { mode, setMode, theme } = useThemeContext()
  const [reduceMotion, setReduceMotion] = useState(false)
  const [highContrast, setHighContrast] = useState(false)
  const [fontScale, setFontScale] = useState<'sm' | 'md' | 'lg'>('md')
  const [accent, setAccent] = useState<AccentKey>('default')

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header />

      <div className="mt-6 grid grid-cols-1 gap-5 lg:grid-cols-[1fr_400px]">
        <div className="space-y-5">
          <ModeSection mode={mode} effective={theme} onChange={setMode} />
          <AccentSection accent={accent} onChange={setAccent} />
          <DensitySection fontScale={fontScale} onChange={setFontScale} />
          <AccessibilitySection
            reduceMotion={reduceMotion}
            onReduceMotion={setReduceMotion}
            highContrast={highContrast}
            onHighContrast={setHighContrast}
          />
        </div>

        <aside className="lg:sticky lg:top-4 lg:h-fit">
          <PreviewSection effective={theme} accent={accent} />
        </aside>
      </div>
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header() {
  return (
    <header>
      <h1 className="flex items-center gap-2 text-xl font-semibold">
        <Palette size={18} strokeWidth={1.75} />
        Theme
      </h1>
      <p className="mt-1 text-sm text-muted-foreground">
        How UTMStack looks on your screen. These preferences are stored locally — they don't apply
        to other members of your organization.
      </p>
    </header>
  )
}

/* ─── Mode picker ──────────────────────────────────────────────────────── */

const MODES: { id: ThemeMode; label: string; icon: LucideIcon; description: string }[] = [
  { id: 'light', label: 'Light', icon: Sun, description: 'High contrast, daylight-friendly.' },
  { id: 'dark', label: 'Dark', icon: Moon, description: 'Low-glare, easy on the eyes.' },
  { id: 'system', label: 'System', icon: Monitor, description: 'Follow your operating system.' },
]

function ModeSection({
  mode,
  effective,
  onChange,
}: {
  mode: ThemeMode
  effective: 'light' | 'dark'
  onChange: (m: ThemeMode) => void
}) {
  return (
    <Section title="Appearance" subtitle="Choose how UTMStack looks. The change applies immediately.">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        {MODES.map((m) => (
          <ModeCard
            key={m.id}
            id={m.id}
            label={m.label}
            description={m.description}
            icon={m.icon}
            active={mode === m.id}
            onClick={() => onChange(m.id)}
          />
        ))}
      </div>

      {mode === 'system' && (
        <div className="mt-3 inline-flex items-center gap-2 rounded-md border border-border bg-muted/40 px-3 py-1.5 text-[11px] text-muted-foreground">
          <span className="inline-block h-1.5 w-1.5 rounded-full bg-emerald-500" />
          Currently following your OS preference: <span className="font-medium text-foreground">{effective}</span>
        </div>
      )}
    </Section>
  )
}

function ModeCard({
  id,
  label,
  description,
  icon: Icon,
  active,
  onClick,
}: {
  id: ThemeMode
  label: string
  description: string
  icon: LucideIcon
  active: boolean
  onClick: () => void
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'group flex flex-col overflow-hidden rounded-xl border text-left transition-all',
        active ? 'border-primary/50 ring-1 ring-primary/20' : 'border-border hover:border-foreground/20'
      )}
    >
      {/* Mini preview swatch */}
      <ModeSwatch id={id} />

      {/* Label */}
      <div className="flex items-start gap-3 border-t border-border bg-card p-3">
        <Icon size={16} className={active ? 'text-primary' : 'text-muted-foreground'} />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5 text-sm font-semibold">
            {label}
            {active && (
              <Check size={12} strokeWidth={3} className="text-primary" />
            )}
          </div>
          <p className="mt-0.5 text-[11px] text-muted-foreground">{description}</p>
        </div>
      </div>
    </button>
  )
}

function ModeSwatch({ id }: { id: ThemeMode }) {
  // Static visual for each mode — bg, foreground, accent dots
  if (id === 'light') {
    return (
      <div className="flex h-24 items-stretch bg-zinc-50">
        <div className="w-12 border-r border-zinc-200 bg-zinc-100">
          <div className="m-2 h-1.5 w-6 rounded-full bg-zinc-300" />
          <div className="mx-2 mt-1 h-1.5 w-6 rounded-full bg-zinc-300" />
        </div>
        <div className="flex-1 p-2">
          <div className="h-1.5 w-12 rounded-full bg-zinc-300" />
          <div className="mt-1 h-1.5 w-16 rounded-full bg-zinc-200" />
          <div className="mt-2 flex gap-1">
            <span className="h-3 w-6 rounded-sm bg-rose-400" />
            <span className="h-3 w-6 rounded-sm bg-amber-400" />
            <span className="h-3 w-6 rounded-sm bg-sky-400" />
          </div>
        </div>
      </div>
    )
  }
  if (id === 'dark') {
    return (
      <div className="flex h-24 items-stretch bg-zinc-950">
        <div className="w-12 border-r border-zinc-800 bg-zinc-900">
          <div className="m-2 h-1.5 w-6 rounded-full bg-zinc-700" />
          <div className="mx-2 mt-1 h-1.5 w-6 rounded-full bg-zinc-700" />
        </div>
        <div className="flex-1 p-2">
          <div className="h-1.5 w-12 rounded-full bg-zinc-600" />
          <div className="mt-1 h-1.5 w-16 rounded-full bg-zinc-700" />
          <div className="mt-2 flex gap-1">
            <span className="h-3 w-6 rounded-sm bg-rose-500" />
            <span className="h-3 w-6 rounded-sm bg-amber-500" />
            <span className="h-3 w-6 rounded-sm bg-sky-500" />
          </div>
        </div>
      </div>
    )
  }
  // System — split light/dark diagonally
  return (
    <div className="relative flex h-24 items-stretch overflow-hidden">
      <div className="flex flex-1 items-stretch bg-zinc-50">
        <div className="w-6 bg-zinc-100" />
        <div className="flex-1 p-2">
          <div className="h-1.5 w-10 rounded-full bg-zinc-300" />
          <div className="mt-2 flex gap-1">
            <span className="h-2.5 w-4 rounded-sm bg-rose-400" />
            <span className="h-2.5 w-4 rounded-sm bg-sky-400" />
          </div>
        </div>
      </div>
      <div className="flex flex-1 items-stretch bg-zinc-950">
        <div className="w-6 bg-zinc-900" />
        <div className="flex-1 p-2">
          <div className="h-1.5 w-10 rounded-full bg-zinc-600" />
          <div className="mt-2 flex gap-1">
            <span className="h-2.5 w-4 rounded-sm bg-rose-500" />
            <span className="h-2.5 w-4 rounded-sm bg-sky-500" />
          </div>
        </div>
      </div>
    </div>
  )
}

/* ─── Accent picker ────────────────────────────────────────────────────── */

type AccentKey = 'default' | 'fuchsia' | 'sky' | 'emerald' | 'amber' | 'rose'

const ACCENTS: { id: AccentKey; label: string; hex: string }[] = [
  { id: 'default', label: 'Default', hex: '#8b5cf6' }, // violet
  { id: 'fuchsia', label: 'Fuchsia', hex: '#d946ef' },
  { id: 'sky', label: 'Sky', hex: '#0ea5e9' },
  { id: 'emerald', label: 'Emerald', hex: '#10b981' },
  { id: 'amber', label: 'Amber', hex: '#f59e0b' },
  { id: 'rose', label: 'Rose', hex: '#f43f5e' },
]

function AccentSection({
  accent,
  onChange,
}: {
  accent: AccentKey
  onChange: (k: AccentKey) => void
}) {
  return (
    <Section
      title="Accent color"
      subtitle="Used for primary buttons, links, and active states. Default is the UTMStack brand."
    >
      <div className="flex flex-wrap items-center gap-2.5">
        {ACCENTS.map((a) => {
          const active = accent === a.id
          return (
            <button
              key={a.id}
              onClick={() => onChange(a.id)}
              className={cn(
                'group flex items-center gap-2 rounded-md border px-2.5 py-1.5 text-xs transition-colors',
                active ? 'border-foreground/30 bg-muted/40' : 'border-border hover:bg-muted/40'
              )}
            >
              <span
                className="h-4 w-4 shrink-0 rounded-full ring-2 ring-offset-1 ring-offset-card"
                style={{ background: a.hex, boxShadow: `0 0 0 2px ${active ? a.hex : 'transparent'}` }}
              />
              <span className={cn(active && 'font-medium')}>{a.label}</span>
              {active && <Check size={11} strokeWidth={3} className="text-foreground" />}
            </button>
          )
        })}
      </div>
      <p className="mt-3 text-[11px] text-muted-foreground">
        Stored locally for your account. Doesn't change branding for other members of the org.
      </p>
    </Section>
  )
}

/* ─── Density picker ───────────────────────────────────────────────────── */

function DensitySection({
  fontScale,
  onChange,
}: {
  fontScale: 'sm' | 'md' | 'lg'
  onChange: (s: 'sm' | 'md' | 'lg') => void
}) {
  const opts = [
    { id: 'sm' as const, label: 'Compact', sub: 'More on screen', sample: 'text-xs' },
    { id: 'md' as const, label: 'Default', sub: 'Recommended', sample: 'text-sm' },
    { id: 'lg' as const, label: 'Comfortable', sub: 'Easier to read', sample: 'text-base' },
  ]
  return (
    <Section
      title="Density & text size"
      subtitle="Tighter rows show more rows per screen; looser is easier on long sessions."
    >
      <div className="grid grid-cols-3 gap-2">
        {opts.map((o) => {
          const active = fontScale === o.id
          return (
            <button
              key={o.id}
              onClick={() => onChange(o.id)}
              className={cn(
                'flex flex-col gap-1 rounded-lg border p-3 text-left transition-colors',
                active ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/40'
              )}
            >
              <div className="flex items-center gap-1.5">
                <Type size={12} className={active ? 'text-primary' : 'text-muted-foreground'} />
                <span className={cn('font-semibold', active ? 'text-primary' : 'text-foreground')}>
                  {o.label}
                </span>
                {active && <Check size={11} strokeWidth={3} className="ml-auto text-primary" />}
              </div>
              <span className="text-[10px] text-muted-foreground">{o.sub}</span>
              <span className={cn('mt-1 font-mono', o.sample)}>Aa</span>
            </button>
          )
        })}
      </div>
    </Section>
  )
}

/* ─── Accessibility ────────────────────────────────────────────────────── */

function AccessibilitySection({
  reduceMotion,
  onReduceMotion,
  highContrast,
  onHighContrast,
}: {
  reduceMotion: boolean
  onReduceMotion: (v: boolean) => void
  highContrast: boolean
  onHighContrast: (v: boolean) => void
}) {
  return (
    <Section title="Accessibility" subtitle="Reduce visual noise where it gets in the way.">
      <ToggleRow
        icon={Zap}
        label="Reduce motion"
        description="Skip non-essential animations and transitions."
        enabled={reduceMotion}
        onChange={onReduceMotion}
      />
      <ToggleRow
        icon={Eye}
        label="High contrast"
        description="Stronger borders and text contrast for low-light or low-vision use."
        enabled={highContrast}
        onChange={onHighContrast}
      />
    </Section>
  )
}

function ToggleRow({
  icon: Icon,
  label,
  description,
  enabled,
  onChange,
}: {
  icon: LucideIcon
  label: string
  description: string
  enabled: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <div className="flex items-start gap-3 border-b border-border/60 py-3 last:border-b-0">
      <Icon size={14} className="mt-0.5 shrink-0 text-muted-foreground" />
      <div className="min-w-0 flex-1">
        <div className="text-sm font-medium">{label}</div>
        <div className="text-xs text-muted-foreground">{description}</div>
      </div>
      <Toggle enabled={enabled} onChange={() => onChange(!enabled)} />
    </div>
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

/* ─── Preview ──────────────────────────────────────────────────────────── */

function PreviewSection({
  effective,
  accent,
}: {
  effective: 'light' | 'dark'
  accent: AccentKey
}) {
  const accentHex = ACCENTS.find((a) => a.id === accent)?.hex ?? '#8b5cf6'
  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card">
      <div className="border-b border-border px-5 py-3">
        <div className="text-[11px] uppercase tracking-wider text-muted-foreground">Preview</div>
        <div className="mt-0.5 text-sm font-semibold">Live · {effective}</div>
      </div>

      <div className={cn('p-4', effective === 'dark' ? 'bg-zinc-950' : 'bg-zinc-50')}>
        {/* Faux app shell */}
        <div
          className={cn(
            'overflow-hidden rounded-lg border',
            effective === 'dark' ? 'border-zinc-800 bg-zinc-900' : 'border-zinc-200 bg-white'
          )}
        >
          {/* Topbar */}
          <div
            className={cn(
              'flex items-center gap-2 border-b px-3 py-2',
              effective === 'dark' ? 'border-zinc-800' : 'border-zinc-200'
            )}
          >
            <span className="h-5 w-5 rounded" style={{ background: accentHex }} />
            <span className={cn('text-xs font-semibold', effective === 'dark' ? 'text-zinc-100' : 'text-zinc-900')}>
              UTMStack
            </span>
            <span className={cn('text-[10px]', effective === 'dark' ? 'text-zinc-500' : 'text-zinc-400')}>
              / Acme Corp
            </span>
            <div className="ml-auto flex gap-1">
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
              <span className={cn('h-1.5 w-1.5 rounded-full', effective === 'dark' ? 'bg-zinc-700' : 'bg-zinc-300')} />
            </div>
          </div>

          {/* Body */}
          <div className="flex">
            {/* Sidebar */}
            <div
              className={cn(
                'w-16 border-r p-2',
                effective === 'dark' ? 'border-zinc-800 bg-zinc-900' : 'border-zinc-200 bg-zinc-100'
              )}
            >
              <div
                className="mb-1 h-2 w-full rounded"
                style={{ background: accentHex, opacity: 0.85 }}
              />
              <div
                className={cn(
                  'mb-1 h-2 w-full rounded',
                  effective === 'dark' ? 'bg-zinc-700' : 'bg-zinc-300'
                )}
              />
              <div
                className={cn(
                  'mb-1 h-2 w-full rounded',
                  effective === 'dark' ? 'bg-zinc-700' : 'bg-zinc-300'
                )}
              />
              <div
                className={cn(
                  'h-2 w-full rounded',
                  effective === 'dark' ? 'bg-zinc-700' : 'bg-zinc-300'
                )}
              />
            </div>

            {/* Main content — sample alert rows */}
            <div className="flex-1 p-3">
              <div
                className={cn(
                  'mb-2 text-xs font-semibold',
                  effective === 'dark' ? 'text-zinc-100' : 'text-zinc-900'
                )}
              >
                Alerts
              </div>
              <div className="space-y-1.5">
                <SampleRow effective={effective} severity="critical" text="Suspected Kerberoasting" />
                <SampleRow effective={effective} severity="high" text="Brute force attempts" />
                <SampleRow effective={effective} severity="medium" text="SMB enumeration" />
                <SampleRow effective={effective} severity="low" text="Signed binary proxy" />
              </div>

              {/* Sample button */}
              <button
                className="mt-3 rounded-md px-2.5 py-1 text-[10px] font-medium text-white"
                style={{ background: accentHex }}
              >
                Investigate
              </button>
            </div>
          </div>
        </div>

        <p
          className={cn(
            'mt-3 text-[11px]',
            effective === 'dark' ? 'text-zinc-500' : 'text-zinc-500'
          )}
        >
          This preview updates as you change settings.
        </p>
      </div>
    </section>
  )
}

function SampleRow({
  effective,
  severity,
  text,
}: {
  effective: 'light' | 'dark'
  severity: 'critical' | 'high' | 'medium' | 'low'
  text: string
}) {
  const sevColor =
    severity === 'critical'
      ? 'bg-red-500'
      : severity === 'high'
        ? 'bg-orange-500'
        : severity === 'medium'
          ? 'bg-amber-500'
          : 'bg-sky-500'
  return (
    <div
      className={cn(
        'flex items-center gap-2 rounded px-2 py-1.5 text-[10px]',
        effective === 'dark' ? 'bg-zinc-800/60 text-zinc-200' : 'bg-zinc-100 text-zinc-800'
      )}
    >
      <span className={cn('h-3 w-1 rounded-full', sevColor)} />
      <span className="truncate">{text}</span>
      <span className={cn('ml-auto', effective === 'dark' ? 'text-zinc-500' : 'text-zinc-400')}>
        2m
      </span>
    </div>
  )
}

/* ─── Section wrapper ──────────────────────────────────────────────────── */

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
