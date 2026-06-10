import { useState } from 'react'
import {
  Activity,
  AlertTriangle,
  Boxes,
  Building2,
  CheckCircle2,
  Copy,
  Cpu,
  Database,
  Download,
  ExternalLink,
  FileText,
  Heart,
  HelpCircle,
  Info,
  KeyRound,
  Network,
  Package,
  RefreshCw,
  Server,
  ShieldCheck,
  Sparkles,
  Wifi,
  Workflow,
  type LucideIcon,
} from 'lucide-react'
import { Link } from 'react-router-dom'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { APP_INFO } from '@/shared/config/app'

/* ─── Types ────────────────────────────────────────────────────────────── */

type ServiceStatus = 'up' | 'degraded' | 'down'

interface Service {
  name: string
  description: string
  status: ServiceStatus
  version: string
  icon: LucideIcon
}

interface ComponentVer {
  name: string
  version: string
  link?: string
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const SERVICES: Service[] = [
  { name: 'Panel API', description: 'REST + GraphQL gateway and auth.', status: 'up', version: '11.2.7', icon: Workflow },
  { name: 'Agent manager', description: 'Manages connected agents, commands, and heartbeats.', status: 'up', version: '11.2.7', icon: Server },
  { name: 'Event processor', description: 'Parses, normalizes, and indexes incoming events.', status: 'up', version: '11.2.7', icon: Activity },
  { name: 'Correlation engine', description: 'Runs detection rules in real time against the event stream.', status: 'up', version: '11.2.7', icon: ShieldCheck },
  { name: 'OpenSearch', description: 'Log and event storage backend.', status: 'up', version: '2.15.0', icon: Database },
  { name: 'PostgreSQL', description: 'Configuration and metadata database.', status: 'up', version: '16.4', icon: Database },
  { name: 'Kafka', description: 'Event ingestion bus.', status: 'up', version: '3.7.1', icon: Wifi },
  { name: 'SOC AI service', description: 'AI summarization and triage worker.', status: 'degraded', version: '0.9.4', icon: Sparkles },
  { name: 'Frontend', description: 'This panel.', status: 'up', version: '11.2.7', icon: Package },
]

const COMPONENTS: ComponentVer[] = [
  { name: 'utmstack-backend', version: '11.2.7', link: 'https://github.com/utmstack/UTMStack' },
  { name: 'utmstack-agent', version: '11.2.7' },
  { name: 'utmstack-collector', version: '11.2.7' },
  { name: 'utmstack-correlation', version: '11.2.7' },
  { name: 'utmstack-event-processor', version: '11.2.7' },
  { name: 'utmstack-frontend', version: '11.2.7' },
  { name: 'opensearch', version: '2.15.0' },
  { name: 'postgres', version: '16.4' },
  { name: 'kafka', version: '3.7.1' },
  { name: 'go runtime', version: '1.23.4' },
  { name: 'node runtime', version: '20.11.1' },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const STATUS_META: Record<ServiceStatus, { label: string; tone: string; dot: string }> = {
  up: { label: 'Healthy', tone: 'text-emerald-500', dot: 'bg-emerald-500' },
  degraded: { label: 'Degraded', tone: 'text-amber-500', dot: 'bg-amber-500' },
  down: { label: 'Down', tone: 'text-red-500', dot: 'bg-red-500 animate-pulse' },
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function AboutPage() {
  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header />

      <div className="mt-6">
        <InstanceCard />
      </div>

      <div className="mt-5 grid grid-cols-1 gap-5 lg:grid-cols-[1fr_360px]">
        <div className="space-y-5">
          <SystemHealthSection />
          <ComponentVersionsSection />
          <UpdatesSection />
        </div>

        <aside className="space-y-5">
          <SupportSection />
          <LegalSection />
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
        <Info size={18} strokeWidth={1.75} />
        About this instance
      </h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Version, license, system health, and useful links to support and documentation.
      </p>
    </header>
  )
}

/* ─── Instance card (hero) ─────────────────────────────────────────────── */

function InstanceCard() {
  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card">
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_320px]">
        {/* Left — version + edition */}
        <div className="border-b border-border p-6 lg:border-b-0 lg:border-r">
          <div className="flex items-start gap-3">
            <img src="/logo.svg" alt="UTMStack" className="h-12 w-12" />
            <div className="min-w-0 flex-1">
              <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
                {APP_INFO.edition === 'enterprise' ? `Enterprise${APP_INFO.tier ? ` · ${APP_INFO.tier}` : ''}` : 'Community'}
              </div>
              <div className="mt-0.5 flex items-baseline gap-2">
                <span className="text-3xl font-semibold tabular-nums">UTMStack</span>
                <span className="font-mono text-lg text-muted-foreground">v{APP_INFO.version}</span>
              </div>
              <div className="mt-1 text-xs text-muted-foreground">
                Released April 22, 2026 · Build <span className="font-mono">d4f8a02</span> · Channel{' '}
                <span className="font-mono">stable</span>
              </div>
            </div>
          </div>

          <dl className="mt-6 grid grid-cols-1 gap-y-2.5 text-xs sm:grid-cols-[160px_1fr]">
            <Row k="Instance URL"><CopyableMono value="https://utmstack.acme.com" /></Row>
            <Row k="Deployment ID"><CopyableMono value="dep_4f12c08e-902a-4ab1-9b0a-2f58cd0e88e1" /></Row>
            <Row k="Region"><span>us-east-1</span></Row>
            <Row k="Hostname"><span className="font-mono">panel-prod-01</span></Row>
            <Row k="Deployed"><span className="font-mono">Apr 22, 2026 · 14:18 UTC</span> <span className="text-muted-foreground">· 10 days ago</span></Row>
            <Row k="Uptime"><span className="font-mono text-emerald-500">9d 22h 14m</span></Row>
          </dl>
        </div>

        {/* Right — license (managed on the License page) */}
        <div className="flex flex-col justify-between bg-muted/20 p-6">
          <div>
            <div className="text-[11px] uppercase tracking-wider text-muted-foreground">License</div>
            <div className="mt-3 flex items-center gap-2 text-sm font-medium">
              {APP_INFO.edition === 'enterprise' ? (
                <>
                  <ShieldCheck size={16} className="text-primary" />
                  {APP_INFO.tier ? `Enterprise · ${APP_INFO.tier}` : 'Enterprise'}
                </>
              ) : (
                <>
                  <Building2 size={16} className="text-muted-foreground" />
                  Community
                </>
              )}
            </div>
            <p className="mt-1.5 text-xs text-muted-foreground">
              Edition, datasource limits, expiry, and license upload are managed on the License
              page.
            </p>
          </div>
          <Button asChild size="sm" variant="outline" className="mt-4 w-full">
            <Link to="/settings/license">
              <KeyRound size={13} className="mr-1.5" />
              View license
            </Link>
          </Button>
        </div>
      </div>
    </section>
  )
}

function CopyableMono({ value }: { value: string }) {
  const [copied, setCopied] = useState(false)
  return (
    <span className="inline-flex items-center gap-1.5">
      <code className="font-mono text-[11px]">{value}</code>
      <button
        onClick={() => {
          navigator.clipboard.writeText(value)
          setCopied(true)
          setTimeout(() => setCopied(false), 1500)
        }}
        className="rounded p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground"
      >
        {copied ? <CheckCircle2 size={11} className="text-emerald-500" /> : <Copy size={11} />}
      </button>
    </span>
  )
}

function Row({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  )
}

/* ─── System health ────────────────────────────────────────────────────── */

function SystemHealthSection() {
  const issues = SERVICES.filter((s) => s.status !== 'up').length
  return (
    <Section
      title="System health"
      subtitle={
        issues === 0
          ? 'All services healthy. Refresh to recheck.'
          : `${issues} service${issues === 1 ? '' : 's'} need attention.`
      }
      action={
        <Button variant="outline" size="sm">
          <RefreshCw size={13} className="mr-1.5" />
          Refresh
        </Button>
      }
    >
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
        {SERVICES.map((s) => (
          <ServiceRow key={s.name} service={s} />
        ))}
      </div>
    </Section>
  )
}

function ServiceRow({ service: s }: { service: Service }) {
  const meta = STATUS_META[s.status]
  const Icon = s.icon
  return (
    <div className="flex items-center gap-3 rounded-md border border-border bg-card px-3 py-2.5">
      <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-muted text-muted-foreground">
        <Icon size={13} />
      </span>
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-medium">{s.name}</span>
          <span className="font-mono text-[10px] text-muted-foreground">v{s.version}</span>
        </div>
        <div className="truncate text-[11px] text-muted-foreground">{s.description}</div>
      </div>
      <span className="inline-flex items-center gap-1.5 text-[11px]">
        <span className={cn('h-1.5 w-1.5 rounded-full', meta.dot)} />
        <span className={meta.tone}>{meta.label}</span>
      </span>
    </div>
  )
}

/* ─── Component versions ───────────────────────────────────────────────── */

function ComponentVersionsSection() {
  const [copied, setCopied] = useState(false)
  const copyAll = () => {
    const text = COMPONENTS.map((c) => `${c.name} ${c.version}`).join('\n')
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }
  return (
    <Section
      title="Component versions"
      subtitle="Useful when filing a bug report — we'll need these."
      action={
        <Button variant="outline" size="sm" onClick={copyAll}>
          {copied ? (
            <>
              <CheckCircle2 size={13} className="mr-1.5 text-emerald-500" />
              Copied
            </>
          ) : (
            <>
              <Copy size={13} className="mr-1.5" />
              Copy all
            </>
          )}
        </Button>
      }
    >
      <div className="overflow-hidden rounded-md border border-border bg-card">
        {COMPONENTS.map((c, i) => (
          <div
            key={c.name}
            className={cn(
              'grid grid-cols-[1fr_140px_36px] items-center gap-3 px-3 py-2 text-xs',
              i < COMPONENTS.length - 1 && 'border-b border-border/60'
            )}
          >
            <span className="font-mono">{c.name}</span>
            <span className="font-mono text-muted-foreground">v{c.version}</span>
            <div className="flex justify-end">
              {c.link && (
                <a
                  href={c.link}
                  target="_blank"
                  rel="noopener"
                  className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
                >
                  <ExternalLink size={11} />
                </a>
              )}
            </div>
          </div>
        ))}
      </div>
    </Section>
  )
}

/* ─── Updates ──────────────────────────────────────────────────────────── */

function UpdatesSection() {
  const upToDate = false
  const latest = '11.2.8'
  return (
    <Section title="Updates" subtitle="Releases follow a stable / candidate / nightly channel.">
      {upToDate ? (
        <div className="flex items-center gap-3 rounded-md border border-emerald-500/30 bg-emerald-500/5 p-3">
          <CheckCircle2 size={18} className="shrink-0 text-emerald-500" />
          <div className="flex-1 text-xs">
            <div className="font-medium">You're on the latest version</div>
            <div className="text-muted-foreground">Last checked just now.</div>
          </div>
          <Button variant="outline" size="sm">
            <RefreshCw size={13} className="mr-1.5" />
            Check again
          </Button>
        </div>
      ) : (
        <div className="flex items-center gap-3 rounded-md border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-3">
          <Sparkles size={18} className="shrink-0 text-fuchsia-500" />
          <div className="flex-1 text-xs">
            <div className="font-medium">
              <span className="font-mono">v{latest}</span> is available
            </div>
            <div className="text-muted-foreground">
              Released 3 days ago · Bug fixes for Kerberos parser, Suricata EVE updates.
            </div>
          </div>
          <Button size="sm">
            <Download size={13} className="mr-1.5" />
            Install update
          </Button>
        </div>
      )}

      <div className="mt-3 flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 p-3 text-xs">
        <AlertTriangle size={13} className="mt-0.5 shrink-0 text-amber-500" />
        <div className="text-muted-foreground">
          <span className="font-medium text-foreground">Updates trigger a brief restart</span> of
          the panel and event processor. Ingestion is buffered during the upgrade window
          (~3 minutes typical).
        </div>
      </div>

      <div className="mt-3 flex items-center gap-3 text-[11px]">
        <a className="inline-flex items-center gap-1 text-primary hover:underline" href="#">
          <FileText size={11} />
          View changelog
        </a>
        <span className="text-muted-foreground">·</span>
        <a className="inline-flex items-center gap-1 text-primary hover:underline" href="#">
          <Boxes size={11} />
          Release notes
        </a>
        <span className="text-muted-foreground">·</span>
        <a className="inline-flex items-center gap-1 text-primary hover:underline" href="#">
          <Cpu size={11} />
          Roadmap
        </a>
      </div>
    </Section>
  )
}

/* ─── Support ──────────────────────────────────────────────────────────── */

function SupportSection() {
  return (
    <Section title="Support">
      <ul className="space-y-1.5 text-sm">
        <LinkRow icon={FileText} label="Documentation" href="#" />
        <LinkRow icon={HelpCircle} label="Knowledge base" href="#" />
        <LinkRow icon={Heart} label="Contact support" href="#" sub="Avg response 14 min" />
        <LinkRow icon={Network} label="Status page" href="#" sub="status.utmstack.com" />
        <LinkRow icon={Activity} label="Community Slack" href="#" sub="2,400+ members" />
      </ul>
    </Section>
  )
}

/* ─── Legal ────────────────────────────────────────────────────────────── */

function LegalSection() {
  return (
    <Section title="Legal & licenses">
      <ul className="space-y-1.5 text-sm">
        <LinkRow icon={FileText} label="Terms of service" href="#" />
        <LinkRow icon={ShieldCheck} label="Privacy policy" href="#" />
        <LinkRow icon={Package} label="Open-source licenses" href="#" sub="427 third-party packages" />
        <LinkRow icon={KeyRound} label="Security disclosure" href="#" />
      </ul>
      <div className="mt-4 text-[10px] leading-relaxed text-muted-foreground">
        © {new Date().getFullYear()} UTMStack. All rights reserved. UTMStack is a registered trademark.
      </div>
    </Section>
  )
}

function LinkRow({
  icon: Icon,
  label,
  href,
  sub,
}: {
  icon: LucideIcon
  label: string
  href: string
  sub?: string
}) {
  return (
    <li>
      <a
        href={href}
        target="_blank"
        rel="noopener"
        className="group flex items-start gap-3 rounded-md border border-transparent px-2 py-1.5 transition-colors hover:border-border hover:bg-muted/40"
      >
        <Icon size={14} className="mt-0.5 shrink-0 text-muted-foreground group-hover:text-foreground" />
        <div className="min-w-0 flex-1">
          <div className="text-xs font-medium">{label}</div>
          {sub && <div className="text-[10px] text-muted-foreground">{sub}</div>}
        </div>
        <ExternalLink size={11} className="mt-1 shrink-0 text-muted-foreground opacity-0 group-hover:opacity-100" />
      </a>
    </li>
  )
}

/* ─── Section wrapper ──────────────────────────────────────────────────── */

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
