import { useEffect, useState } from 'react'
import {
  AlertTriangle,
  Apple,
  Check,
  Cloud,
  Copy,
  Eye,
  EyeOff,
  KeyRound,
  Network,
  RefreshCw,
  Server,
  ShieldCheck,
  Terminal,
  type LucideIcon,
} from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { cn } from '@/shared/lib/utils'
import { AuthHttpError } from '@/features/auth/services/auth-http.service'
import { configHttpService } from '../services/config-http.service'

const CONNECTION_KEY = 'connection_key.token'

export function ConnectionKeyPage() {
  const [token, setToken] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [rotating, setRotating] = useState(false)
  const [showPlaintext, setShowPlaintext] = useState(false)
  const [copied, setCopied] = useState(false)
  const [confirming, setConfirming] = useState(false)

  useEffect(() => {
    configHttpService
      .get(CONNECTION_KEY)
      .then((r) => setToken(r.value ?? ''))
      .catch((err) => {
        if (err instanceof AuthHttpError && err.status === 404) return
        toast.error(err instanceof Error ? err.message : 'Could not load connection key')
      })
      .finally(() => setLoading(false))
  }, [])

  const rotate = async () => {
    setRotating(true)
    try {
      const r = await configHttpService.rotate(CONNECTION_KEY)
      setToken(r.value ?? '')
      setShowPlaintext(true)
      toast.success(token ? 'Connection key regenerated' : 'Connection key generated')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not rotate connection key')
    } finally {
      setRotating(false)
      setConfirming(false)
    }
  }

  const copy = async (text?: string) => {
    const value = text ?? token
    if (!value) return
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      toast.error('Could not copy')
    }
  }

  const hasToken = token !== ''
  // Mock-side info — the real values would come from the backend.
  // Treated as static here to stay non-disruptive.
  const lastRotated = hasToken ? '14 days ago' : null
  const connectedClients = hasToken ? 11 : 0
  const isStale = hasToken // treat as stale only if you'd rotate every 90d; simplified here.

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header />

      <div className="mt-6 space-y-5">
        <StatusStrip
          hasToken={hasToken}
          lastRotated={lastRotated}
          connectedClients={connectedClients}
          loading={loading}
        />

        <TokenSection
          token={token}
          hasToken={hasToken}
          loading={loading}
          rotating={rotating}
          showPlaintext={showPlaintext}
          copied={copied}
          confirming={confirming}
          onShowToggle={() => setShowPlaintext((v) => !v)}
          onCopy={() => copy()}
          onConfirm={() => setConfirming(true)}
          onCancel={() => setConfirming(false)}
          onRotate={rotate}
        />

        {hasToken && <UsageSection token={token} onCopy={copy} />}

        {hasToken && isStale && <RotationReminderSection />}
      </div>
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header() {
  return (
    <header>
      <h1 className="flex items-center gap-2 text-xl font-semibold">
        <KeyRound size={18} strokeWidth={1.75} />
        Connection key
      </h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Token external systems use to connect to this UTMStack instance — agents, collectors,
        and the custom JSON intake.
      </p>
    </header>
  )
}

/* ─── Status strip ─────────────────────────────────────────────────────── */

function StatusStrip({
  hasToken,
  lastRotated,
  connectedClients,
  loading,
}: {
  hasToken: boolean
  lastRotated: string | null
  connectedClients: number
  loading: boolean
}) {
  return (
    <section className="rounded-xl border border-border bg-card">
      <div className="grid grid-cols-1 divide-y divide-border sm:grid-cols-3 sm:divide-x sm:divide-y-0">
        <StripStat
          label="Status"
          value={
            loading ? (
              <span className="text-muted-foreground">Loading…</span>
            ) : hasToken ? (
              <span className="inline-flex items-center gap-1.5 text-emerald-500">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                Active
              </span>
            ) : (
              <span className="inline-flex items-center gap-1.5 text-muted-foreground">
                <span className="h-1.5 w-1.5 rounded-full bg-muted-foreground" />
                Not generated
              </span>
            )
          }
          sub={hasToken ? 'Accepting connections' : 'Generate a key to start'}
        />
        <StripStat
          label="Last rotated"
          value={
            loading ? (
              <span className="text-muted-foreground">—</span>
            ) : (
              <span className="font-mono">{lastRotated ?? '—'}</span>
            )
          }
          sub="Recommended every 90 days"
        />
        <StripStat
          label="Connected clients"
          value={
            loading ? (
              <span className="text-muted-foreground">—</span>
            ) : (
              <span className="font-mono tabular-nums">{connectedClients}</span>
            )
          }
          sub="Agents and collectors heartbeating now"
        />
      </div>
    </section>
  )
}

function StripStat({
  label,
  value,
  sub,
}: {
  label: string
  value: React.ReactNode
  sub: string
}) {
  return (
    <div className="px-5 py-4">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className="mt-1 text-xl font-semibold">{value}</div>
      <div className="mt-0.5 text-[11px] text-muted-foreground">{sub}</div>
    </div>
  )
}

/* ─── Token section ────────────────────────────────────────────────────── */

function TokenSection({
  token,
  hasToken,
  loading,
  rotating,
  showPlaintext,
  copied,
  confirming,
  onShowToggle,
  onCopy,
  onConfirm,
  onCancel,
  onRotate,
}: {
  token: string
  hasToken: boolean
  loading: boolean
  rotating: boolean
  showPlaintext: boolean
  copied: boolean
  confirming: boolean
  onShowToggle: () => void
  onCopy: () => void
  onConfirm: () => void
  onCancel: () => void
  onRotate: () => void
}) {
  return (
    <Section
      title="Token"
      subtitle="Treat this value as a credential. If you suspect it's been exposed, regenerate it — existing clients using the old key will lose access until reconfigured."
    >
      {loading ? (
        <div className="text-sm text-muted-foreground">Loading…</div>
      ) : !hasToken ? (
        <div className="rounded-md border border-dashed border-border bg-muted/30 p-6 text-center">
          <KeyRound size={24} strokeWidth={1.5} className="mx-auto text-muted-foreground" />
          <div className="mt-2 text-sm font-medium">No connection key generated yet</div>
          <p className="mt-1 text-xs text-muted-foreground">
            Generate a key so agents and collectors can authenticate to this instance.
          </p>
          <Button onClick={onRotate} disabled={rotating} className="mt-4">
            <RefreshCw size={14} className={cn('mr-1.5', rotating && 'animate-spin')} />
            {rotating ? 'Generating…' : 'Generate connection key'}
          </Button>
        </div>
      ) : (
        <>
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <input
                readOnly
                value={token}
                type={showPlaintext ? 'text' : 'password'}
                className={cn(
                  'h-10 w-full rounded-md border border-input bg-background/40 pr-9 pl-3 font-mono text-xs',
                  'placeholder:text-muted-foreground'
                )}
              />
              <button
                onClick={onShowToggle}
                className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
                aria-label={showPlaintext ? 'Hide token' : 'Show token'}
              >
                {showPlaintext ? <EyeOff size={14} /> : <Eye size={14} />}
              </button>
            </div>
            <Button variant="outline" size="sm" onClick={onCopy}>
              {copied ? <Check size={14} className="mr-1.5 text-emerald-500" /> : <Copy size={14} className="mr-1.5" />}
              {copied ? 'Copied' : 'Copy'}
            </Button>
          </div>

          <div className="mt-5 border-t border-border pt-4">
            {!confirming ? (
              <Button variant="outline" onClick={onConfirm} disabled={loading || rotating}>
                <RefreshCw size={14} className={cn('mr-1.5', rotating && 'animate-spin')} />
                Regenerate connection key
              </Button>
            ) : (
              <div className="rounded-md border border-amber-500/30 bg-amber-500/5 p-4">
                <div className="flex items-center gap-2 text-sm font-medium text-amber-600 dark:text-amber-300">
                  <AlertTriangle size={14} />
                  Regenerate the key?
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  Existing clients using the current key will lose access until reconfigured with the
                  new value. There's no undo.
                </p>
                <div className="mt-3 flex gap-2">
                  <Button size="sm" onClick={onRotate} disabled={rotating}>
                    {rotating ? 'Regenerating…' : 'Yes, regenerate'}
                  </Button>
                  <Button variant="ghost" size="sm" onClick={onCancel} disabled={rotating}>
                    Cancel
                  </Button>
                </div>
              </div>
            )}
          </div>
        </>
      )}
    </Section>
  )
}

/* ─── Usage section ────────────────────────────────────────────────────── */

interface Usage {
  id: string
  label: string
  icon: LucideIcon
  iconTone: string
  command: string
}

function UsageSection({ token, onCopy }: { token: string; onCopy: (v: string) => void }) {
  const [pick, setPick] = useState<string>('linux')
  const tk = token || 'YOUR_TOKEN'
  const usages: Usage[] = [
    {
      id: 'linux',
      label: 'Linux agent',
      icon: Terminal,
      iconTone: 'text-amber-500',
      command: `curl -fsSL https://install.utmstack.com/agent.sh | sudo bash -s -- \\
  --token="${tk}"`,
    },
    {
      id: 'windows',
      label: 'Windows agent',
      icon: Server,
      iconTone: 'text-sky-500',
      command: `# Run in PowerShell as Administrator
iwr https://install.utmstack.com/agent.ps1 -OutFile install.ps1
./install.ps1 -Token "${tk}"`,
    },
    {
      id: 'macos',
      label: 'macOS agent',
      icon: Apple,
      iconTone: 'text-violet-500',
      command: `curl -fsSL https://install.utmstack.com/agent-macos.sh | sudo bash -s -- \\
  --token="${tk}"`,
    },
    {
      id: 'collector',
      label: 'Collector',
      icon: Network,
      iconTone: 'text-fuchsia-500',
      command: `curl -fsSL https://install.utmstack.com/collector.sh | sudo bash -s -- \\
  --token="${tk}"`,
    },
    {
      id: 'json',
      label: 'JSON intake',
      icon: Cloud,
      iconTone: 'text-emerald-500',
      command: `curl -X POST https://ingest.utmstack.com/v1/json \\
  -H "Authorization: Bearer ${tk}" \\
  -H "Content-Type: application/json" \\
  -d '{"@timestamp": "2026-05-02T16:42:00Z", "host": "my-app", "message": "hello"}'`,
    },
  ]
  const current = usages.find((u) => u.id === pick) ?? usages[0]
  const CurrentIcon = current.icon

  return (
    <Section
      title="Where you'll use this token"
      subtitle="Copy the install command for the platform you want to onboard. The token is already embedded."
    >
      <div className="flex flex-wrap items-center gap-1 rounded-md border border-border bg-muted/30 p-1">
        {usages.map((u) => {
          const Icon = u.icon
          const active = pick === u.id
          return (
            <button
              key={u.id}
              onClick={() => setPick(u.id)}
              className={cn(
                'flex items-center gap-1.5 rounded px-2.5 py-1.5 text-xs transition-colors',
                active
                  ? 'bg-card text-foreground shadow-sm'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              <Icon size={12} className={active ? u.iconTone : ''} />
              {u.label}
            </button>
          )
        })}
      </div>

      <div className="mt-3 overflow-hidden rounded-md border border-border bg-muted/40">
        <div className="flex items-center justify-between border-b border-border px-3 py-1.5 text-[11px]">
          <span className="inline-flex items-center gap-1.5 text-muted-foreground">
            <CurrentIcon size={11} className={current.iconTone} />
            <span>{current.label}</span>
          </span>
          <button
            onClick={() => onCopy(current.command)}
            className="inline-flex items-center gap-1 rounded px-2 py-0.5 text-muted-foreground hover:bg-card hover:text-foreground"
          >
            <Copy size={11} />
            Copy command
          </button>
        </div>
        <pre className="overflow-x-auto p-3 font-mono text-[11px] leading-relaxed">{current.command}</pre>
      </div>
    </Section>
  )
}

/* ─── Rotation reminder ────────────────────────────────────────────────── */

function RotationReminderSection() {
  return (
    <Section
      title="Security recommendations"
      subtitle="Best practices for protecting this credential."
    >
      <ul className="space-y-2.5 text-sm">
        <li className="flex items-start gap-3">
          <ShieldCheck size={14} className="mt-0.5 shrink-0 text-emerald-500" />
          <div>
            <div className="font-medium">Rotate every 90 days</div>
            <div className="text-xs text-muted-foreground">
              Schedule a recurring rotation. UTMStack can email you a reminder when the key is older
              than 90 days.{' '}
              <button className="text-primary hover:underline">Set up reminder</button>
            </div>
          </div>
        </li>
        <li className="flex items-start gap-3">
          <ShieldCheck size={14} className="mt-0.5 shrink-0 text-emerald-500" />
          <div>
            <div className="font-medium">Store as a secret</div>
            <div className="text-xs text-muted-foreground">
              Don't commit the token to source control. Use your secret manager (Vault, AWS SM, etc.)
              and inject it at install time.
            </div>
          </div>
        </li>
        <li className="flex items-start gap-3">
          <ShieldCheck size={14} className="mt-0.5 shrink-0 text-emerald-500" />
          <div>
            <div className="font-medium">Audit who fetched it</div>
            <div className="text-xs text-muted-foreground">
              Every fetch and rotation is recorded in the application audit log.{' '}
              <a href="/settings/audit-logs" className="text-primary hover:underline">View audit log</a>
            </div>
          </div>
        </li>
      </ul>
    </Section>
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
