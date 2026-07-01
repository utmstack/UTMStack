import { useEffect, useState } from 'react'
import {
  AlertTriangle,
  Apple,
  Check,
  Copy,
  Eye,
  EyeOff,
  KeyRound,
  RefreshCw,
  Server,
  ShieldCheck,
  Terminal,
  type LucideIcon,
} from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { cn } from '@/shared/lib/utils'
import { connectionKeyHttpService } from '../services/connection-key-http.service'

export function ConnectionKeyPage() {
  const { t } = useTranslation()
  const [token, setToken] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [rotating, setRotating] = useState(false)
  const [showPlaintext, setShowPlaintext] = useState(false)
  const [copied, setCopied] = useState(false)
  const [confirming, setConfirming] = useState(false)

  const load = () => {
    setLoading(true)
    setError(false)
    connectionKeyHttpService
      .get()
      .then((r) => setToken(r.connectionKey ?? ''))
      .catch(() => {
        setError(true)
        // Don't surface the raw backend/gRPC error — the status banner already
        // explains the agent-manager is unreachable.
        toast.error(t('connectionKey.toast.loadFailed'))
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    load()
  }, [])

  const rotate = async () => {
    setRotating(true)
    try {
      const r = await connectionKeyHttpService.rotate()
      setToken(r.connectionKey ?? '')
      setShowPlaintext(true)
      toast.success(t('connectionKey.toast.rotated'))
    } catch {
      toast.error(t('connectionKey.toast.rotateFailed'))
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
      toast.error(t('connectionKey.toast.copyFailed'))
    }
  }

  const hasToken = token !== ''

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <Header />

      <div className="mt-6 space-y-5">
        <StatusStrip loading={loading} hasToken={hasToken} error={error} />

        <TokenSection
          token={token}
          hasToken={hasToken}
          loading={loading}
          error={error}
          rotating={rotating}
          showPlaintext={showPlaintext}
          copied={copied}
          confirming={confirming}
          onRetry={load}
          onShowToggle={() => setShowPlaintext((v) => !v)}
          onCopy={() => copy()}
          onConfirm={() => setConfirming(true)}
          onCancel={() => setConfirming(false)}
          onRotate={rotate}
        />

        {hasToken && <UsageSection token={token} onCopy={copy} />}

        {hasToken && <RotationReminderSection />}
      </div>
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header() {
  const { t } = useTranslation()
  return (
    <header>
      <h1 className="flex items-center gap-2 text-base font-semibold">
        <KeyRound size={16} strokeWidth={1.75} />
        {t('connectionKey.title')}
      </h1>
    </header>
  )
}

/* ─── Status strip ─────────────────────────────────────────────────────── */

function StatusStrip({
  loading,
  hasToken,
  error,
}: {
  loading: boolean
  hasToken: boolean
  error: boolean
}) {
  const { t } = useTranslation()
  const unavailable = error && !hasToken
  return (
    <section className="rounded-xl border border-border bg-card px-5 py-4">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
        {t('connectionKey.status.label')}
      </div>
      <div className="mt-1 text-xl font-semibold">
        {loading ? (
          <span className="text-muted-foreground">{t('connectionKey.status.loading')}</span>
        ) : unavailable ? (
          <span className="inline-flex items-center gap-1.5 text-amber-500">
            <span className="h-1.5 w-1.5 rounded-full bg-amber-500" />
            {t('connectionKey.status.unavailable')}
          </span>
        ) : (
          <span className="inline-flex items-center gap-1.5 text-emerald-500">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
            {t('connectionKey.status.active')}
          </span>
        )}
      </div>
      <div className="mt-0.5 text-[11px] text-muted-foreground">
        {unavailable
          ? t('connectionKey.status.unavailableHint')
          : t('connectionKey.status.activeHint')}
      </div>
    </section>
  )
}

/* ─── Token section ────────────────────────────────────────────────────── */

function TokenSection({
  token,
  hasToken,
  loading,
  error,
  rotating,
  showPlaintext,
  copied,
  confirming,
  onRetry,
  onShowToggle,
  onCopy,
  onConfirm,
  onCancel,
  onRotate,
}: {
  token: string
  hasToken: boolean
  loading: boolean
  error: boolean
  rotating: boolean
  showPlaintext: boolean
  copied: boolean
  confirming: boolean
  onRetry: () => void
  onShowToggle: () => void
  onCopy: () => void
  onConfirm: () => void
  onCancel: () => void
  onRotate: () => void
}) {
  const { t } = useTranslation()
  return (
    <Section title={t('connectionKey.token.title')} subtitle={t('connectionKey.token.subtitle')}>
      {loading ? (
        <div className="text-sm text-muted-foreground">{t('connectionKey.status.loading')}</div>
      ) : !hasToken ? (
        <div className="rounded-md border border-dashed border-border bg-muted/30 p-6 text-center">
          <AlertTriangle size={24} strokeWidth={1.5} className="mx-auto text-amber-500" />
          <div className="mt-2 text-sm font-medium">
            {error
              ? t('connectionKey.token.loadFailedTitle')
              : t('connectionKey.token.unavailableTitle')}
          </div>
          <p className="mt-1 text-xs text-muted-foreground">
            {t('connectionKey.token.loadFailedHint')}
          </p>
          <Button variant="outline" onClick={onRetry} disabled={loading} className="mt-4">
            <RefreshCw size={14} className={cn('mr-1.5', loading && 'animate-spin')} />
            {t('connectionKey.token.retry')}
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
                aria-label={showPlaintext ? t('connectionKey.token.hide') : t('connectionKey.token.show')}
              >
                {showPlaintext ? <EyeOff size={14} /> : <Eye size={14} />}
              </button>
            </div>
            <Button variant="outline" size="sm" onClick={onCopy}>
              {copied ? <Check size={14} className="mr-1.5 text-emerald-500" /> : <Copy size={14} className="mr-1.5" />}
              {copied ? t('connectionKey.token.copied') : t('connectionKey.token.copy')}
            </Button>
          </div>

          <div className="mt-5 border-t border-border pt-4">
            {!confirming ? (
              <Button variant="outline" onClick={onConfirm} disabled={loading || rotating}>
                <RefreshCw size={14} className={cn('mr-1.5', rotating && 'animate-spin')} />
                {t('connectionKey.token.regenerate')}
              </Button>
            ) : (
              <div className="rounded-md border border-amber-500/30 bg-amber-500/5 p-4">
                <div className="flex items-center gap-2 text-sm font-medium text-amber-600 dark:text-amber-300">
                  <AlertTriangle size={14} />
                  {t('connectionKey.token.confirmTitle')}
                </div>
                <p className="mt-1 text-xs text-muted-foreground">
                  {t('connectionKey.token.confirmBody')}
                </p>
                <div className="mt-3 flex gap-2">
                  <Button size="sm" onClick={onRotate} disabled={rotating}>
                    {rotating ? t('connectionKey.token.regenerating') : t('connectionKey.token.confirmYes')}
                  </Button>
                  <Button variant="ghost" size="sm" onClick={onCancel} disabled={rotating}>
                    {t('connectionKey.token.cancel')}
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
  const { t } = useTranslation()
  const [pick, setPick] = useState<string>('linux')
  const tk = token || 'YOUR_TOKEN'
  const usages: Usage[] = [
    {
      id: 'linux',
      label: t('connectionKey.install.linux'),
      icon: Terminal,
      iconTone: 'text-amber-500',
      command: `curl -fsSL https://install.utmstack.com/agent.sh | sudo bash -s -- \\
  --token="${tk}"`,
    },
    {
      id: 'windows',
      label: t('connectionKey.install.windows'),
      icon: Server,
      iconTone: 'text-sky-500',
      command: `# Run in PowerShell as Administrator
iwr https://install.utmstack.com/agent.ps1 -OutFile install.ps1
./install.ps1 -Token "${tk}"`,
    },
    {
      id: 'macos',
      label: t('connectionKey.install.macos'),
      icon: Apple,
      iconTone: 'text-violet-500',
      command: `curl -fsSL https://install.utmstack.com/agent-macos.sh | sudo bash -s -- \\
  --token="${tk}"`,
    },
  ]
  const current = usages.find((u) => u.id === pick) ?? usages[0]
  const CurrentIcon = current.icon

  return (
    <Section
      title={t('connectionKey.install.title')}
      subtitle={t('connectionKey.install.subtitle')}
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
            {t('connectionKey.install.copyCommand')}
          </button>
        </div>
        <pre className="overflow-x-auto p-3 font-mono text-[11px] leading-relaxed">{current.command}</pre>
      </div>
    </Section>
  )
}

/* ─── Rotation reminder ────────────────────────────────────────────────── */

function RotationReminderSection() {
  const { t } = useTranslation()
  return (
    <Section title={t('connectionKey.security.title')} subtitle={t('connectionKey.security.subtitle')}>
      <ul className="space-y-2.5 text-sm">
        <li className="flex items-start gap-3">
          <ShieldCheck size={14} className="mt-0.5 shrink-0 text-emerald-500" />
          <div>
            <div className="font-medium">{t('connectionKey.security.rotateTitle')}</div>
            <div className="text-xs text-muted-foreground">
              {t('connectionKey.security.rotateBody')}{' '}
              <button className="text-primary hover:underline">
                {t('connectionKey.security.rotateAction')}
              </button>
            </div>
          </div>
        </li>
        <li className="flex items-start gap-3">
          <ShieldCheck size={14} className="mt-0.5 shrink-0 text-emerald-500" />
          <div>
            <div className="font-medium">{t('connectionKey.security.secretTitle')}</div>
            <div className="text-xs text-muted-foreground">{t('connectionKey.security.secretBody')}</div>
          </div>
        </li>
        <li className="flex items-start gap-3">
          <ShieldCheck size={14} className="mt-0.5 shrink-0 text-emerald-500" />
          <div>
            <div className="font-medium">{t('connectionKey.security.auditTitle')}</div>
            <div className="text-xs text-muted-foreground">
              {t('connectionKey.security.auditBody')}{' '}
              <a href="/settings/audit-logs" className="text-primary hover:underline">
                {t('connectionKey.security.auditAction')}
              </a>
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
