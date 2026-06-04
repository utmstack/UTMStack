import { useState } from 'react'
import {
  AlertTriangle,
  Check,
  ExternalLink,
  Eye,
  EyeOff,
  Lock,
  Mail,
  RefreshCw,
  Send,
  Shield,
  Sparkles,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type Encryption = 'none' | 'starttls' | 'tls'
type Provider = 'custom' | 'sendgrid' | 'aws-ses' | 'mailgun' | 'postmark' | 'resend' | 'ms365' | 'gmail'

interface EmailTemplate {
  id: string
  name: string
  description: string
  subject: string
  systemOwner: boolean
  enabled: boolean
}

interface EmailLog {
  ts: string
  to: string
  subject: string
  status: 'delivered' | 'queued' | 'failed' | 'bounced'
  reason?: string
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-02T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()

const TEMPLATES: EmailTemplate[] = [
  { id: 't-alert', name: 'Critical alert notification', description: 'Sent to recipients when a critical or high-severity alert fires.', subject: '[UTMStack] {{severity}} alert · {{alert.name}}', systemOwner: true, enabled: true },
  { id: 't-incident', name: 'Incident assigned', description: 'Sent when an incident is assigned to a member.', subject: '[UTMStack] Incident #{{id}} assigned to you', systemOwner: true, enabled: true },
  { id: 't-incident-status', name: 'Incident status changed', description: 'Sent when status moves to In review / Completed / Merged.', subject: '[UTMStack] Incident #{{id}} · status: {{status}}', systemOwner: true, enabled: true },
  { id: 't-compliance', name: 'Compliance report ready', description: 'Sent when a scheduled compliance report finishes generating.', subject: '[UTMStack] {{framework}} report ready', systemOwner: true, enabled: true },
  { id: 't-invite', name: 'User invitation', description: 'Sent when an admin invites a new member.', subject: 'You\'ve been invited to UTMStack · {{org}}', systemOwner: true, enabled: true },
  { id: 't-passwordreset', name: 'Password reset', description: 'Sent for local accounts when a password reset is requested.', subject: 'Reset your UTMStack password', systemOwner: true, enabled: true },
  { id: 't-keyrotation', name: 'Connection key rotated', description: 'Heads-up when the connection key is rotated by an admin.', subject: '[UTMStack] Connection key rotated by {{actor}}', systemOwner: true, enabled: false },
]

const RECENT: EmailLog[] = [
  { ts: min(2), to: 'jsmith@utmstack.com', subject: '[UTMStack] CRITICAL alert · Suspected Kerberoasting', status: 'delivered' },
  { ts: min(8), to: 'soc@utmstack.com', subject: '[UTMStack] Incident #1142 · status: In review', status: 'delivered' },
  { ts: min(45), to: 'compliance@utmstack.com', subject: '[UTMStack] PCI DSS report ready', status: 'delivered' },
  { ts: hr(2), to: 'ext@unreachable.test', subject: '[UTMStack] CRITICAL alert · Outbound to C2', status: 'bounced', reason: 'Mailbox unavailable (550 5.1.1)' },
  { ts: hr(4), to: 'mthomas@utmstack.com', subject: 'You\'ve been invited to UTMStack · Acme Corp', status: 'delivered' },
  { ts: hr(8), to: 'queue@utmstack.com', subject: '[UTMStack] HIGH alert · PowerShell encoded', status: 'queued' },
]

const PROVIDERS: { id: Provider; label: string; host: string; port: number; encryption: Encryption; tone: string }[] = [
  { id: 'custom', label: 'Custom SMTP', host: 'smtp.example.com', port: 587, encryption: 'starttls', tone: 'text-zinc-500' },
  { id: 'sendgrid', label: 'SendGrid', host: 'smtp.sendgrid.net', port: 587, encryption: 'starttls', tone: 'text-sky-500' },
  { id: 'aws-ses', label: 'AWS SES', host: 'email-smtp.us-east-1.amazonaws.com', port: 587, encryption: 'starttls', tone: 'text-orange-500' },
  { id: 'mailgun', label: 'Mailgun', host: 'smtp.mailgun.org', port: 587, encryption: 'starttls', tone: 'text-rose-500' },
  { id: 'postmark', label: 'Postmark', host: 'smtp.postmarkapp.com', port: 587, encryption: 'starttls', tone: 'text-amber-500' },
  { id: 'resend', label: 'Resend', host: 'smtp.resend.com', port: 587, encryption: 'starttls', tone: 'text-emerald-500' },
  { id: 'ms365', label: 'Microsoft 365', host: 'smtp.office365.com', port: 587, encryption: 'starttls', tone: 'text-fuchsia-500' },
  { id: 'gmail', label: 'Google Workspace', host: 'smtp.gmail.com', port: 587, encryption: 'starttls', tone: 'text-amber-500' },
]

const STATUS_META: Record<EmailLog['status'], { label: string; tone: string; dot: string }> = {
  delivered: { label: 'Delivered', tone: 'text-emerald-500', dot: 'bg-emerald-500' },
  queued: { label: 'Queued', tone: 'text-sky-500', dot: 'bg-sky-500 animate-pulse' },
  failed: { label: 'Failed', tone: 'text-red-500', dot: 'bg-red-500' },
  bounced: { label: 'Bounced', tone: 'text-amber-500', dot: 'bg-amber-500' },
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function EmailConfigurationPage() {
  const [provider, setProvider] = useState<Provider>('aws-ses')
  const [host, setHost] = useState('email-smtp.us-east-1.amazonaws.com')
  const [port, setPort] = useState(587)
  const [encryption, setEncryption] = useState<Encryption>('starttls')
  const [authUser, setAuthUser] = useState('AKIA...')
  const [authPass, setAuthPass] = useState('rxYjk...')
  const [showPass, setShowPass] = useState(false)
  const [fromName, setFromName] = useState('UTMStack')
  const [fromEmail, setFromEmail] = useState('alerts@acme.utmstack.com')
  const [replyTo, setReplyTo] = useState('soc@acme.utmstack.com')
  const [testStatus, setTestStatus] = useState<'idle' | 'sending' | 'success' | 'fail'>('idle')
  const [dirty, setDirty] = useState(false)

  const onProviderPick = (p: Provider) => {
    const meta = PROVIDERS.find((x) => x.id === p)!
    setProvider(p)
    setHost(meta.host)
    setPort(meta.port)
    setEncryption(meta.encryption)
    setDirty(true)
  }

  const sendTest = () => {
    setTestStatus('sending')
    setTimeout(() => setTestStatus('success'), 900)
  }

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header dirty={dirty} onSave={() => setDirty(false)} onDiscard={() => setDirty(false)} />

      <div className="mt-6">
        <StatsStrip />
      </div>

      <div className="mt-6 space-y-5">
        <ProviderSection provider={provider} onPick={onProviderPick} />

        <ConnectionSection
          host={host}
          port={port}
          encryption={encryption}
          authUser={authUser}
          authPass={authPass}
          showPass={showPass}
          onHost={(v) => { setHost(v); setDirty(true) }}
          onPort={(v) => { setPort(v); setDirty(true) }}
          onEncryption={(v) => { setEncryption(v); setDirty(true) }}
          onAuthUser={(v) => { setAuthUser(v); setDirty(true) }}
          onAuthPass={(v) => { setAuthPass(v); setDirty(true) }}
          onShowPass={() => setShowPass((v) => !v)}
        />

        <SenderSection
          fromName={fromName}
          fromEmail={fromEmail}
          replyTo={replyTo}
          onFromName={(v) => { setFromName(v); setDirty(true) }}
          onFromEmail={(v) => { setFromEmail(v); setDirty(true) }}
          onReplyTo={(v) => { setReplyTo(v); setDirty(true) }}
        />

        <TestConnectionSection status={testStatus} onSend={sendTest} />

        <TemplatesSection />

        <RecentSection />
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
          <Mail size={18} strokeWidth={1.75} />
          Email configuration
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          SMTP relay used to deliver alerts, incident notifications, invitations, and reports.
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

/* ─── Stats strip ──────────────────────────────────────────────────────── */

function StatsStrip() {
  return (
    <section className="rounded-xl border border-border bg-card">
      <div className="grid grid-cols-1 divide-y divide-border sm:grid-cols-4 sm:divide-x sm:divide-y-0">
        <StripStat label="Status" value={
          <span className="inline-flex items-center gap-1.5 text-emerald-500">
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
            Connected
          </span>
        } sub="Last test 8m ago" />
        <StripStat label="Sent · 24h" value={<span className="font-mono">312</span>} sub="0.3% bounce rate" />
        <StripStat label="Sent · 30d" value={<span className="font-mono">8,412</span>} sub="98.7% delivered" />
        <StripStat label="Quota" value={<span className="font-mono">312 / 50,000</span>} sub="Daily SES limit" />
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

/* ─── Provider section ─────────────────────────────────────────────────── */

function ProviderSection({ provider, onPick }: { provider: Provider; onPick: (p: Provider) => void }) {
  return (
    <Section title="Provider" subtitle="Pick a known SMTP provider for sensible defaults, or use Custom SMTP for anything else.">
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
        {PROVIDERS.map((p) => {
          const active = provider === p.id
          return (
            <button
              key={p.id}
              onClick={() => onPick(p.id)}
              className={cn(
                'flex items-center gap-2 rounded-lg border px-3 py-2.5 text-left text-xs transition-colors',
                active ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/40'
              )}
            >
              <span className={cn('h-2 w-2 rounded-full', active ? 'bg-primary' : 'bg-muted-foreground/40')} />
              <span className={cn('font-medium', active && 'text-primary')}>{p.label}</span>
              {active && <Check size={11} strokeWidth={3} className="ml-auto text-primary" />}
            </button>
          )
        })}
      </div>
    </Section>
  )
}

/* ─── Connection section ───────────────────────────────────────────────── */

function ConnectionSection({
  host,
  port,
  encryption,
  authUser,
  authPass,
  showPass,
  onHost,
  onPort,
  onEncryption,
  onAuthUser,
  onAuthPass,
  onShowPass,
}: {
  host: string
  port: number
  encryption: Encryption
  authUser: string
  authPass: string
  showPass: boolean
  onHost: (v: string) => void
  onPort: (v: number) => void
  onEncryption: (v: Encryption) => void
  onAuthUser: (v: string) => void
  onAuthPass: (v: string) => void
  onShowPass: () => void
}) {
  return (
    <Section title="Connection" subtitle="Where UTMStack will deliver outgoing email.">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-[1fr_120px_180px]">
        <Field label="SMTP host">
          <Input
            value={host}
            onChange={(e) => onHost(e.target.value)}
            placeholder="smtp.example.com"
            className="h-9 font-mono text-xs"
          />
        </Field>
        <Field label="Port">
          <Input
            type="number"
            value={port}
            onChange={(e) => onPort(Number(e.target.value) || 0)}
            className="h-9 font-mono text-xs"
          />
        </Field>
        <Field label="Encryption">
          <select
            value={encryption}
            onChange={(e) => onEncryption(e.target.value as Encryption)}
            className="h-9 w-full rounded-md border border-border bg-background/40 px-2 text-sm focus:bg-card focus:outline-none focus:ring-1 focus:ring-ring"
          >
            <option value="none">None</option>
            <option value="starttls">STARTTLS</option>
            <option value="tls">TLS / SSL</option>
          </select>
        </Field>
      </div>

      <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Field label="Username">
          <Input
            value={authUser}
            onChange={(e) => onAuthUser(e.target.value)}
            placeholder="api-user"
            className="h-9 font-mono text-xs"
          />
        </Field>
        <Field label="Password">
          <div className="relative">
            <Input
              type={showPass ? 'text' : 'password'}
              value={authPass}
              onChange={(e) => onAuthPass(e.target.value)}
              className="h-9 pr-9 font-mono text-xs"
            />
            <button
              onClick={onShowPass}
              type="button"
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              {showPass ? <EyeOff size={13} /> : <Eye size={13} />}
            </button>
          </div>
        </Field>
      </div>

      {encryption === 'none' && (
        <div className="mt-3 flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 p-3 text-xs">
          <AlertTriangle size={13} className="mt-0.5 shrink-0 text-amber-500" />
          <div>
            <span className="font-medium text-foreground">Unencrypted SMTP.</span> Credentials and
            email contents will travel in plaintext. Use STARTTLS or TLS unless you're testing
            against a local relay.
          </div>
        </div>
      )}
    </Section>
  )
}

/* ─── Sender section ───────────────────────────────────────────────────── */

function SenderSection({
  fromName,
  fromEmail,
  replyTo,
  onFromName,
  onFromEmail,
  onReplyTo,
}: {
  fromName: string
  fromEmail: string
  replyTo: string
  onFromName: (v: string) => void
  onFromEmail: (v: string) => void
  onReplyTo: (v: string) => void
}) {
  return (
    <Section title="Sender identity" subtitle="What recipients see in their inbox.">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <Field label="From name">
          <Input value={fromName} onChange={(e) => onFromName(e.target.value)} className="h-9" />
        </Field>
        <Field label="From email">
          <Input value={fromEmail} onChange={(e) => onFromEmail(e.target.value)} className="h-9 font-mono text-xs" />
        </Field>
        <Field label="Reply-to (optional)" hint="Replies from analysts go here. Leave blank to use From email.">
          <Input value={replyTo} onChange={(e) => onReplyTo(e.target.value)} className="h-9 font-mono text-xs" />
        </Field>
      </div>

      <div className="mt-4 rounded-md border border-border bg-background/40 p-3 text-xs">
        <div className="text-[10px] uppercase tracking-wider text-muted-foreground">Preview</div>
        <div className="mt-1.5">
          <div className="flex items-baseline gap-2">
            <span className="text-sm font-medium">{fromName}</span>
            <span className="font-mono text-[11px] text-muted-foreground">&lt;{fromEmail}&gt;</span>
          </div>
          {replyTo && (
            <div className="mt-0.5 text-[11px] text-muted-foreground">
              Reply-to: <span className="font-mono">{replyTo}</span>
            </div>
          )}
        </div>
      </div>

      <div className="mt-3 flex items-start gap-2 rounded-md border border-emerald-500/30 bg-emerald-500/5 p-3 text-xs">
        <Shield size={13} className="mt-0.5 shrink-0 text-emerald-500" />
        <div className="text-muted-foreground">
          <span className="font-medium text-foreground">SPF, DKIM, and DMARC verified</span> for{' '}
          <span className="font-mono text-foreground">{fromEmail.split('@')[1] || 'your domain'}</span>.
          Email is more likely to land in the inbox.{' '}
          <button className="text-primary hover:underline">Re-check</button>
        </div>
      </div>
    </Section>
  )
}

/* ─── Test connection ──────────────────────────────────────────────────── */

function TestConnectionSection({
  status,
  onSend,
}: {
  status: 'idle' | 'sending' | 'success' | 'fail'
  onSend: () => void
}) {
  const [to, setTo] = useState('me@utmstack.com')
  return (
    <Section title="Test connection" subtitle="Send yourself a test email to verify the configuration works end-to-end.">
      <div className="flex flex-wrap items-end gap-3">
        <div className="flex-1 min-w-[260px]">
          <label className="mb-1.5 block text-[11px] uppercase tracking-wider text-muted-foreground">
            Send to
          </label>
          <Input value={to} onChange={(e) => setTo(e.target.value)} className="h-9 font-mono text-xs" />
        </div>
        <Button size="sm" onClick={onSend} disabled={status === 'sending'}>
          {status === 'sending' ? (
            <RefreshCw size={13} className="mr-1.5 animate-spin" />
          ) : (
            <Send size={13} className="mr-1.5" />
          )}
          {status === 'sending' ? 'Sending…' : 'Send test'}
        </Button>
      </div>

      {status === 'success' && (
        <div className="mt-3 flex items-start gap-2 rounded-md border border-emerald-500/30 bg-emerald-500/5 p-3 text-xs">
          <Check size={13} className="mt-0.5 shrink-0 text-emerald-500" strokeWidth={2.5} />
          <div>
            <span className="font-medium text-foreground">Test email delivered</span>{' '}
            in <span className="font-mono">412ms</span>. Server accepted with code 250.
          </div>
        </div>
      )}
      {status === 'fail' && (
        <div className="mt-3 flex items-start gap-2 rounded-md border border-red-500/30 bg-red-500/5 p-3 text-xs">
          <AlertTriangle size={13} className="mt-0.5 shrink-0 text-red-500" />
          <div>
            <span className="font-medium text-foreground">Test failed.</span> 535 5.7.8
            Authentication credentials invalid. Check the username and password.
          </div>
        </div>
      )}
    </Section>
  )
}

/* ─── Templates section ────────────────────────────────────────────────── */

function TemplatesSection() {
  return (
    <Section title="Email templates" subtitle="What UTMStack sends and when. Edit a template to change subject and body content.">
      <div className="overflow-hidden rounded-md border border-border bg-card">
        <div className="grid grid-cols-[40px_1fr_1fr_80px] gap-3 border-b border-border bg-muted/40 px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
          <div />
          <div>Template</div>
          <div>Subject</div>
          <div className="text-right">Active</div>
        </div>
        {TEMPLATES.map((t) => (
          <div
            key={t.id}
            className="group grid cursor-pointer grid-cols-[40px_1fr_1fr_80px] items-center gap-3 border-b border-border/60 px-3 py-2.5 text-xs last:border-b-0 hover:bg-muted/40"
          >
            <span className="flex h-7 w-7 items-center justify-center rounded-md bg-muted text-muted-foreground">
              <Mail size={13} />
            </span>
            <div className="min-w-0">
              <div className="flex items-center gap-1.5">
                <span className="truncate text-sm font-medium">{t.name}</span>
                {t.systemOwner && (
                  <span className="inline-flex items-center gap-0.5 rounded bg-muted px-1.5 py-0.5 text-[9px] uppercase text-muted-foreground">
                    <Lock size={9} />
                    system
                  </span>
                )}
              </div>
              <div className="truncate text-[11px] text-muted-foreground">{t.description}</div>
            </div>
            <div className="truncate font-mono text-[11px] text-muted-foreground">{t.subject}</div>
            <div className="flex justify-end">
              <Toggle enabled={t.enabled} onChange={() => {}} />
            </div>
          </div>
        ))}
      </div>
    </Section>
  )
}

/* ─── Recent activity ──────────────────────────────────────────────────── */

function RecentSection() {
  return (
    <Section title="Recent deliveries" subtitle="The last few emails UTMStack sent. The full history lives in the audit log.">
      <div className="overflow-hidden rounded-md border border-border bg-card">
        <div className="grid grid-cols-[100px_220px_1fr_110px_36px] gap-3 border-b border-border bg-muted/40 px-3 py-2 text-[10px] uppercase tracking-wider text-muted-foreground">
          <div>When</div>
          <div>To</div>
          <div>Subject</div>
          <div>Status</div>
          <div />
        </div>
        {RECENT.map((r, i) => {
          const st = STATUS_META[r.status]
          return (
            <div
              key={i}
              className="group grid grid-cols-[100px_220px_1fr_110px_36px] items-center gap-3 border-b border-border/60 px-3 py-2 text-xs last:border-b-0 hover:bg-muted/30"
            >
              <div className="font-mono text-[11px] text-muted-foreground">{relativeTime(r.ts)}</div>
              <div className="truncate font-mono text-[11px]">{r.to}</div>
              <div className="min-w-0">
                <div className="truncate font-mono text-[11px]">{r.subject}</div>
                {r.reason && <div className="truncate text-[10px] text-red-500">{r.reason}</div>}
              </div>
              <div className="inline-flex items-center gap-1.5 text-[11px]">
                <span className={cn('h-1.5 w-1.5 rounded-full', st.dot)} />
                <span className={st.tone}>{st.label}</span>
              </div>
              <div className="flex justify-end opacity-0 group-hover:opacity-100">
                <ExternalLink size={11} className="text-muted-foreground" />
              </div>
            </div>
          )
        })}
      </div>

      <div className="mt-3 inline-flex items-center gap-1.5 rounded-md bg-fuchsia-500/10 px-3 py-2 text-[11px] text-fuchsia-600 dark:text-fuchsia-300">
        <Sparkles size={12} className="text-fuchsia-500" />
        Bounce rate <span className="font-mono">0.3%</span> · well below the recommended 2% threshold.
      </div>
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
      <label className="mb-1.5 block text-[11px] uppercase tracking-wider text-muted-foreground">
        {label}
      </label>
      {children}
      {hint && <p className="mt-1 text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}

function Toggle({ enabled, onChange }: { enabled: boolean; onChange: () => void }) {
  return (
    <button
      onClick={onChange}
      className={cn(
        'relative h-4 w-7 shrink-0 rounded-full transition-colors',
        enabled ? 'bg-primary' : 'bg-muted'
      )}
    >
      <span
        className={cn(
          'absolute top-0.5 h-3 w-3 rounded-full bg-white shadow transition-all',
          enabled ? 'left-3.5' : 'left-0.5'
        )}
      />
    </button>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function relativeTime(iso: string) {
  if (!iso) return '—'
  const diff = NOW - new Date(iso).getTime()
  const m = Math.round(diff / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.round(m / 60)
  if (h < 24) return `${h}h ago`
  return `${Math.round(h / 24)}d ago`
}
