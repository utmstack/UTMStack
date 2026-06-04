import { useMemo, useState } from 'react'
import {
  AlertTriangle,
  Archive,
  Camera,
  ChevronDown,
  Database,
  Flame,
  HardDriveDownload,
  Snowflake,
  Sparkles,
  Thermometer,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type Tier = 'hot' | 'warm' | 'cold' | 'frozen'

interface RetentionPolicy {
  id: string
  category: string
  description: string
  // days the data is searchable in each tier; null = skip that tier
  hot: number
  warm: number | null
  cold: number | null
  frozen: number | null
  // total bytes currently stored across tiers
  bytes: number
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

// Disk usage (in GB) — these would come from the backend in production.
const DISK = {
  total: 4_000, // 4 TB total disk
  used: 2_212,
  perTier: {
    hot: 612,    // GB
    warm: 884,   // GB
    cold: 540,   // GB
    frozen: 176, // GB
  } as Record<Tier, number>,
}

const POLICIES: RetentionPolicy[] = [
  {
    id: 'logs',
    category: 'Logs / events',
    description: 'Raw and parsed events from agents, collectors, and cloud integrations.',
    hot: 30,
    warm: 90,
    cold: 365,
    frozen: null,
    bytes: 1_240 * 1024 * 1024 * 1024,
  },
  {
    id: 'alerts',
    category: 'Alerts & incidents',
    description: 'Security alerts, incidents, and their full case history.',
    hot: 90,
    warm: 365,
    cold: 1825,
    frozen: null,
    bytes: 18 * 1024 * 1024 * 1024,
  },
  {
    id: 'audit',
    category: 'Application audit log',
    description: 'Tamper-evident record of who did what inside the panel. Required for compliance.',
    hot: 90,
    warm: 365,
    cold: 1825,
    frozen: 2555,
    bytes: 4 * 1024 * 1024 * 1024,
  },
  {
    id: 'threat-intel',
    category: 'Threat intel matches',
    description: 'IOC match events from any active feed.',
    hot: 30,
    warm: 90,
    cold: 365,
    frozen: null,
    bytes: 12 * 1024 * 1024 * 1024,
  },
  {
    id: 'soar',
    category: 'SOAR / playbook runs',
    description: 'Execution history of automation flows and interactive console sessions.',
    hot: 30,
    warm: 90,
    cold: 365,
    frozen: null,
    bytes: 6 * 1024 * 1024 * 1024,
  },
  {
    id: 'compliance',
    category: 'Compliance reports',
    description: 'Generated reports for HIPAA, PCI, SOC 2, ISO 27001, etc.',
    hot: 365,
    warm: 1825,
    cold: 2555,
    frozen: null,
    bytes: 600 * 1024 * 1024,
  },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const TIER_META: Record<Tier, { label: string; icon: LucideIcon; tone: string; bg: string; description: string }> = {
  hot: {
    label: 'Hot',
    icon: Flame,
    tone: 'text-rose-500',
    bg: 'bg-rose-500',
    description: 'Local NVMe — full-text search, sub-second query, dashboards, real-time alerting.',
  },
  warm: {
    label: 'Warm',
    icon: Thermometer,
    tone: 'text-amber-500',
    bg: 'bg-amber-500',
    description: 'Local SSD — searchable but slower queries; used for investigation and historical drilldown.',
  },
  cold: {
    label: 'Cold',
    icon: HardDriveDownload,
    tone: 'text-sky-500',
    bg: 'bg-sky-500',
    description: 'Object storage (S3 / GCS) — searchable with delay (~minutes); cost-optimized for compliance.',
  },
  frozen: {
    label: 'Frozen',
    icon: Snowflake,
    tone: 'text-violet-500',
    bg: 'bg-violet-500',
    description: 'Long-term archive — restore-only. Required for regulated retention (7+ years).',
  },
}

const COMPLIANCE_PRESETS = [
  { id: 'pci', label: 'PCI DSS (1y online · 1y archive)' },
  { id: 'hipaa', label: 'HIPAA (6 years for PHI access)' },
  { id: 'sox', label: 'SOX (7 years financial systems)' },
  { id: 'iso27001', label: 'ISO 27001 (3 years minimum)' },
  { id: 'gdpr', label: 'GDPR (purpose-limited, no default)' },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function DataRetentionPage() {
  const [policies, setPolicies] = useState(POLICIES)
  const [snapshotEnabled, setSnapshotEnabled] = useState(true)
  const [snapshotFrequency, setSnapshotFrequency] = useState('daily')
  const [snapshotDest, setSnapshotDest] = useState('s3://utmstack-snapshots/acme')
  const [autoDelete, setAutoDelete] = useState(true)
  const [autoDeleteThreshold, setAutoDeleteThreshold] = useState(85)
  const [dirty, setDirty] = useState(false)

  const updatePolicy = <K extends keyof RetentionPolicy>(
    id: string,
    field: K,
    value: RetentionPolicy[K]
  ) => {
    setPolicies((prev) => prev.map((p) => (p.id === id ? { ...p, [field]: value } : p)))
    setDirty(true)
  }

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header dirty={dirty} onSave={() => setDirty(false)} onDiscard={() => { setPolicies(POLICIES); setDirty(false) }} />

      <div className="mt-6 space-y-5">
        <StorageUsageSection />
        <RetentionPoliciesSection policies={policies} updatePolicy={updatePolicy} setDirty={setDirty} />
        <SnapshotSection
          enabled={snapshotEnabled}
          onEnabledChange={(v) => { setSnapshotEnabled(v); setDirty(true) }}
          frequency={snapshotFrequency}
          onFrequencyChange={(v) => { setSnapshotFrequency(v); setDirty(true) }}
          dest={snapshotDest}
          onDestChange={(v) => { setSnapshotDest(v); setDirty(true) }}
        />
        <AutoCleanupSection
          enabled={autoDelete}
          onEnabledChange={(v) => { setAutoDelete(v); setDirty(true) }}
          threshold={autoDeleteThreshold}
          onThresholdChange={(v) => { setAutoDeleteThreshold(v); setDirty(true) }}
        />
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
          <HardDriveDownload size={18} strokeWidth={1.75} />
          Data retention
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          How long different kinds of data stay searchable, and where it goes when it ages.
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

/* ─── Storage usage ────────────────────────────────────────────────────── */

function StorageUsageSection() {
  const usedPct = (DISK.used / DISK.total) * 100
  const tiers: Tier[] = ['hot', 'warm', 'cold', 'frozen']

  return (
    <Section
      title="Storage usage"
      subtitle="Current footprint across all tiers. Updated every minute."
    >
      <div className="flex flex-wrap items-baseline gap-3">
        <span className="text-2xl font-semibold tabular-nums">
          {DISK.used.toLocaleString()} <span className="text-sm font-normal text-muted-foreground">GB used</span>
        </span>
        <span className="text-sm text-muted-foreground">
          of <span className="text-foreground">{DISK.total.toLocaleString()} GB</span> ·{' '}
          <span className={cn(usedPct > 80 ? 'text-amber-500' : 'text-foreground')}>
            {usedPct.toFixed(1)}%
          </span>
        </span>
      </div>

      {/* Stacked bar */}
      <div className="mt-4 flex h-3 overflow-hidden rounded-full bg-muted">
        {tiers.map((t) => {
          const pct = (DISK.perTier[t] / DISK.total) * 100
          if (pct < 0.5) return null
          return (
            <div
              key={t}
              className={TIER_META[t].bg}
              style={{ width: `${pct}%` }}
              title={`${TIER_META[t].label}: ${DISK.perTier[t]} GB`}
            />
          )
        })}
      </div>

      {/* Per-tier breakdown */}
      <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
        {tiers.map((t) => {
          const meta = TIER_META[t]
          const Icon = meta.icon
          const gb = DISK.perTier[t]
          const pct = (gb / DISK.used) * 100
          return (
            <div key={t} className="rounded-md border border-border bg-background/40 p-3">
              <div className="flex items-center gap-1.5 text-[11px]">
                <Icon size={12} className={meta.tone} />
                <span className="font-medium">{meta.label}</span>
              </div>
              <div className="mt-1 font-mono text-base font-semibold tabular-nums">
                {gb} <span className="text-xs font-normal text-muted-foreground">GB</span>
              </div>
              <div className="text-[10px] text-muted-foreground">{pct.toFixed(1)}% of used</div>
            </div>
          )
        })}
      </div>

      {usedPct > 80 && (
        <div className="mt-4 flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 p-3 text-xs">
          <AlertTriangle size={14} className="mt-0.5 shrink-0 text-amber-500" />
          <div>
            <span className="font-medium text-foreground">Approaching capacity.</span> Auto-cleanup
            will start dropping the oldest hot-tier data when usage reaches{' '}
            <span className="font-mono">85%</span>.
          </div>
        </div>
      )}
    </Section>
  )
}

/* ─── Retention policies ───────────────────────────────────────────────── */

function RetentionPoliciesSection({
  policies,
  updatePolicy,
  setDirty,
}: {
  policies: RetentionPolicy[]
  updatePolicy: <K extends keyof RetentionPolicy>(id: string, field: K, value: RetentionPolicy[K]) => void
  setDirty: (v: boolean) => void
}) {
  return (
    <Section
      title="Retention policies"
      subtitle="Number of days each data category stays in each tier before moving down (or being deleted)."
      action={<CompliancePresetButton onApply={() => setDirty(true)} />}
    >
      {/* Tier legend */}
      <div className="mb-3 flex flex-wrap items-center gap-3 text-[11px]">
        {(['hot', 'warm', 'cold', 'frozen'] as Tier[]).map((t) => {
          const meta = TIER_META[t]
          const Icon = meta.icon
          return (
            <span key={t} className="inline-flex items-center gap-1.5 text-muted-foreground">
              <Icon size={11} className={meta.tone} />
              <span className={meta.tone}>{meta.label}</span>
            </span>
          )
        })}
      </div>

      <div className="overflow-hidden rounded-md border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: '1fr 80px 80px 80px 80px 80px' }}
        >
          <div>Category</div>
          <div className="text-right">Hot</div>
          <div className="text-right">Warm</div>
          <div className="text-right">Cold</div>
          <div className="text-right">Frozen</div>
          <div className="text-right">Size</div>
        </div>
        {policies.map((p) => (
          <PolicyRow key={p.id} policy={p} updatePolicy={updatePolicy} />
        ))}
      </div>

      <p className="mt-3 text-[11px] text-muted-foreground">
        Total online storage equals the sum of Hot + Warm + Cold per category. Set a tier to{' '}
        <span className="font-mono">0</span> to skip it. Frozen data is not searchable until restored.
      </p>
    </Section>
  )
}

function PolicyRow({
  policy: p,
  updatePolicy,
}: {
  policy: RetentionPolicy
  updatePolicy: <K extends keyof RetentionPolicy>(id: string, field: K, value: RetentionPolicy[K]) => void
}) {
  return (
    <div
      className="grid items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs last:border-b-0 hover:bg-muted/30"
      style={{ gridTemplateColumns: '1fr 80px 80px 80px 80px 80px' }}
    >
      <div className="min-w-0">
        <div className="font-medium">{p.category}</div>
        <div className="truncate text-[11px] text-muted-foreground">{p.description}</div>
      </div>
      <DaysInput value={p.hot} onChange={(v) => updatePolicy(p.id, 'hot', v ?? 0)} />
      <DaysInput value={p.warm} onChange={(v) => updatePolicy(p.id, 'warm', v)} nullable />
      <DaysInput value={p.cold} onChange={(v) => updatePolicy(p.id, 'cold', v)} nullable />
      <DaysInput value={p.frozen} onChange={(v) => updatePolicy(p.id, 'frozen', v)} nullable />
      <div className="text-right font-mono text-[11px] tabular-nums text-muted-foreground">
        {formatBytes(p.bytes)}
      </div>
    </div>
  )
}

function DaysInput({
  value,
  onChange,
  nullable,
}: {
  value: number | null
  onChange: (v: number | null) => void
  nullable?: boolean
}) {
  return (
    <div className="relative">
      <input
        type="number"
        min={0}
        value={value ?? ''}
        placeholder={nullable ? '—' : '0'}
        onChange={(e) => {
          const n = e.target.value === '' ? (nullable ? null : 0) : Number(e.target.value)
          onChange(n)
        }}
        className="h-8 w-full rounded-md border border-border bg-background/40 px-2 text-right font-mono text-[11px] tabular-nums focus:bg-card focus:outline-none focus:ring-1 focus:ring-ring"
      />
    </div>
  )
}

function CompliancePresetButton({ onApply }: { onApply: () => void }) {
  const [open, setOpen] = useState(false)
  return (
    <div className="relative">
      <Button variant="outline" size="sm" onClick={() => setOpen((v) => !v)}>
        <Sparkles size={13} className="mr-1.5 text-fuchsia-500" />
        Apply preset
        <ChevronDown size={11} className="ml-1.5 opacity-60" />
      </Button>
      {open && (
        <div className="absolute right-0 top-full z-10 mt-1 w-72 overflow-hidden rounded-md border border-border bg-popover shadow-lg">
          <div className="border-b border-border px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
            Compliance presets
          </div>
          {COMPLIANCE_PRESETS.map((p) => (
            <button
              key={p.id}
              onClick={() => {
                setOpen(false)
                onApply()
              }}
              className="flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-xs hover:bg-muted"
            >
              <span>{p.label}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

/* ─── Snapshots ────────────────────────────────────────────────────────── */

function SnapshotSection({
  enabled,
  onEnabledChange,
  frequency,
  onFrequencyChange,
  dest,
  onDestChange,
}: {
  enabled: boolean
  onEnabledChange: (v: boolean) => void
  frequency: string
  onFrequencyChange: (v: string) => void
  dest: string
  onDestChange: (v: string) => void
}) {
  return (
    <Section
      title="Snapshots"
      subtitle="Periodic backups of your indices to object storage. Snapshots run on a schedule and don't affect query performance."
      action={
        <Toggle enabled={enabled} onChange={() => onEnabledChange(!enabled)} />
      }
    >
      <div className={cn('space-y-4', !enabled && 'pointer-events-none opacity-50')}>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <Field label="Frequency">
            <select
              value={frequency}
              onChange={(e) => onFrequencyChange(e.target.value)}
              className="h-9 w-full rounded-md border border-border bg-background/40 px-2 text-sm focus:bg-card focus:outline-none focus:ring-1 focus:ring-ring"
            >
              <option value="hourly">Every hour</option>
              <option value="6hours">Every 6 hours</option>
              <option value="daily">Daily · 02:00 UTC</option>
              <option value="weekly">Weekly · Sunday 02:00 UTC</option>
            </select>
          </Field>

          <Field label="Retention of snapshots">
            <select className="h-9 w-full rounded-md border border-border bg-background/40 px-2 text-sm focus:bg-card focus:outline-none focus:ring-1 focus:ring-ring">
              <option value="14">Keep last 14 snapshots</option>
              <option value="30">Keep last 30 snapshots</option>
              <option value="90">Keep last 90 snapshots</option>
              <option value="forever">Keep forever</option>
            </select>
          </Field>
        </div>

        <Field
          label="Destination"
          hint="An S3-compatible object storage path (S3, GCS, MinIO, R2). Credentials are configured under Identity providers."
        >
          <Input
            value={dest}
            onChange={(e) => onDestChange(e.target.value)}
            placeholder="s3://my-bucket/utmstack-snapshots"
            className="h-9 font-mono text-xs"
          />
        </Field>

        <div className="flex items-center gap-3 rounded-md border border-border bg-background/40 px-3 py-2.5 text-xs">
          <Camera size={14} className="text-emerald-500" />
          <div className="min-w-0 flex-1">
            <div className="font-medium">Last snapshot</div>
            <div className="truncate text-[11px] text-muted-foreground">
              <span className="font-mono">snap-2026-05-02-0200</span> · 2.1 TB · 4h 12m ago ·{' '}
              <span className="text-emerald-500">success</span>
            </div>
          </div>
          <Button variant="outline" size="sm">
            Run now
          </Button>
        </div>
      </div>
    </Section>
  )
}

/* ─── Auto-cleanup ─────────────────────────────────────────────────────── */

function AutoCleanupSection({
  enabled,
  onEnabledChange,
  threshold,
  onThresholdChange,
}: {
  enabled: boolean
  onEnabledChange: (v: boolean) => void
  threshold: number
  onThresholdChange: (v: number) => void
}) {
  return (
    <Section
      title="Auto-cleanup"
      subtitle="What UTMStack does when disk usage reaches the threshold. Prevents writes from failing in case of a sudden surge."
      action={<Toggle enabled={enabled} onChange={() => onEnabledChange(!enabled)} />}
    >
      <div className={cn('space-y-4', !enabled && 'pointer-events-none opacity-50')}>
        <Field
          label="Trigger when disk usage reaches"
          hint="When this threshold is crossed, the oldest data in the hot tier is moved down (or deleted, if no warm tier is configured)."
        >
          <div className="flex items-center gap-3">
            <input
              type="range"
              min={50}
              max={95}
              step={1}
              value={threshold}
              onChange={(e) => onThresholdChange(Number(e.target.value))}
              className="flex-1 accent-primary"
            />
            <span className="w-14 text-right font-mono text-sm tabular-nums">{threshold}%</span>
          </div>
        </Field>

        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <ActionPick
            id="downgrade"
            label="Downgrade tier"
            description="Move oldest hot data to warm, oldest warm to cold, etc."
            active={true}
          />
          <ActionPick
            id="delete"
            label="Delete oldest"
            description="Skip downgrade, delete oldest data outright."
            active={false}
          />
          <ActionPick
            id="readonly"
            label="Read-only mode"
            description="Stop ingestion until an admin acts."
            active={false}
          />
        </div>

        <div className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 p-3 text-xs">
          <AlertTriangle size={14} className="mt-0.5 shrink-0 text-amber-500" />
          <div className="text-muted-foreground">
            <span className="font-medium text-foreground">Safety:</span> the application audit log
            and compliance reports are <span className="font-medium">never</span> auto-deleted —
            only their tier may be downgraded. To remove them you must do so manually.
          </div>
        </div>
      </div>
    </Section>
  )
}

function ActionPick({
  label,
  description,
  active,
}: {
  id: string
  label: string
  description: string
  active: boolean
}) {
  return (
    <button
      className={cn(
        'flex flex-col gap-1 rounded-lg border p-3 text-left text-xs transition-colors',
        active ? 'border-primary/40 bg-primary/5' : 'border-border bg-card hover:bg-muted/40'
      )}
    >
      <div className="flex items-center gap-2">
        <Database size={12} className={active ? 'text-primary' : 'text-muted-foreground'} />
        <span className={cn('font-medium', active ? 'text-primary' : 'text-foreground')}>
          {label}
        </span>
      </div>
      <p className="text-[11px] text-muted-foreground">{description}</p>
    </button>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({
  title,
  subtitle,
  action,
  children,
}: {
  title: string
  subtitle?: string
  action?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <section className="rounded-xl border border-border bg-card p-6">
      <header className="mb-4 flex items-start justify-between gap-3">
        <div>
          <h2 className="text-sm font-semibold">{title}</h2>
          {subtitle && <p className="mt-0.5 text-xs text-muted-foreground">{subtitle}</p>}
        </div>
        {action}
      </header>
      {children}
    </section>
  )
}

function Field({
  label,
  hint,
  children,
}: {
  label: string
  hint?: string
  children: React.ReactNode
}) {
  return (
    <div>
      <label className="mb-1 block text-[11px] uppercase tracking-wider text-muted-foreground">
        {label}
      </label>
      {children}
      {hint && <p className="mt-1 text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}

function Toggle({
  enabled,
  onChange,
}: {
  enabled: boolean
  onChange: () => void
}) {
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

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function formatBytes(bytes: number) {
  if (bytes >= 1024 ** 4) return `${(bytes / 1024 ** 4).toFixed(1)} TB`
  if (bytes >= 1024 ** 3) return `${(bytes / 1024 ** 3).toFixed(1)} GB`
  if (bytes >= 1024 ** 2) return `${(bytes / 1024 ** 2).toFixed(1)} MB`
  if (bytes >= 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${bytes} B`
}

// Unused-import suppression
void Archive
void useMemo
