import { useCallback, useEffect, useState } from 'react'
import {
  Activity,
  Building2,
  Database,
  ExternalLink,
  FileText,
  Github,
  Globe,
  Info,
  KeyRound,
  Loader2,
  RefreshCw,
  Server,
  ShieldCheck,
  type LucideIcon,
} from 'lucide-react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { aboutHttpService } from '../services/about-http.service'
import type {
  ClusterHealth,
  DatasourceUsage,
  LicenseInfo,
  VersionInfo,
} from '../types/about.types'

/* ─── Data model ───────────────────────────────────────────────────────── */

interface AboutData {
  version: VersionInfo | null
  license: LicenseInfo | null
  usage: DatasourceUsage | null
  cluster: ClusterHealth | null
}

const EMPTY: AboutData = { version: null, license: null, usage: null, cluster: null }

function settled<T>(r: PromiseSettledResult<T>): T | null {
  return r.status === 'fulfilled' ? r.value : null
}

/* ─── Status meta ──────────────────────────────────────────────────────── */

type Status = 'up' | 'degraded' | 'down' | 'unknown'

const STATUS_META: Record<Status, { tone: string; dot: string }> = {
  up: { tone: 'text-emerald-500', dot: 'bg-emerald-500' },
  degraded: { tone: 'text-amber-500', dot: 'bg-amber-500' },
  down: { tone: 'text-red-500', dot: 'bg-red-500' },
  unknown: { tone: 'text-muted-foreground', dot: 'bg-muted-foreground/50' },
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function AboutPage() {
  const { t } = useTranslation()
  const [data, setData] = useState<AboutData>(EMPTY)
  const [loading, setLoading] = useState(true)
  const [backendUp, setBackendUp] = useState(false)
  const [refreshing, setRefreshing] = useState(false)

  const loadAll = useCallback(async () => {
    setLoading(true)
    const [version, license, usage, cluster] = await Promise.allSettled([
      aboutHttpService.version(),
      aboutHttpService.license(),
      aboutHttpService.datasourceUsage(),
      aboutHttpService.clusterHealth(),
    ])
    // The backend is reachable if any authenticated call resolved.
    setBackendUp([version, license, usage, cluster].some((r) => r.status === 'fulfilled'))
    setData({
      version: settled(version),
      license: settled(license),
      usage: settled(usage),
      cluster: settled(cluster),
    })
    setLoading(false)
  }, [])

  // Re-check only the live bit (OpenSearch cluster) without a full reload.
  const refreshHealth = useCallback(async () => {
    setRefreshing(true)
    const [cluster] = await Promise.allSettled([aboutHttpService.clusterHealth()])
    setBackendUp(true)
    setData((d) => ({ ...d, cluster: settled(cluster) }))
    setRefreshing(false)
  }, [])

  useEffect(() => {
    void loadAll()
  }, [loadAll])

  const edition = (data.license?.edition || data.version?.edition || 'community').toLowerCase()
  const isEnterprise = edition === 'enterprise'

  return (
    <div className="w-full px-6 pb-6 pt-3">
      <header>
        <h1 className="flex items-center gap-2 text-base font-semibold">
          <Info size={16} strokeWidth={1.75} />
          {t('about.title')}
        </h1>
      </header>

      {loading ? (
        <div className="mt-6 flex items-center justify-center gap-2 rounded-xl border border-border bg-card px-6 py-16 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          {t('about.loading')}
        </div>
      ) : (
        <>
          <div className="mt-6">
            <InstanceCard data={data} isEnterprise={isEnterprise} />
          </div>

          <div className="mt-5">
            <SystemHealthSection
              data={data}
              backendUp={backendUp}
              refreshing={refreshing}
              onRefresh={() => void refreshHealth()}
            />
          </div>

          <div className="mt-5 grid grid-cols-1 gap-5 md:grid-cols-3">
            <DatasourceUsageSection usage={data.usage} />
            <SupportSection />
            <LegalSection />
          </div>
        </>
      )}
    </div>
  )
}

/* ─── Instance card (hero) ─────────────────────────────────────────────── */

function InstanceCard({ data, isEnterprise }: { data: AboutData; isEnterprise: boolean }) {
  const { t } = useTranslation()
  const version = data.version?.version || '—'
  const hostname = data.version?.server
  const instanceId = data.version?.instanceId
  const license = data.license

  // Community never expires, and the backend serializes the Go zero time
  // ("0001-01-01T00:00:00Z") instead of omitting it — so only show a real,
  // enterprise expiry date.
  const expiry = isEnterprise && license?.expiresAt ? new Date(license.expiresAt) : null
  const hasExpiry = !!expiry && !Number.isNaN(expiry.getTime()) && expiry.getFullYear() > 2000

  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card">
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_320px]">
        {/* Left — version + edition */}
        <div className="flex flex-col justify-center border-b border-border p-6 lg:border-b-0 lg:border-r">
          <div className="flex items-start gap-3">
            <img src="/logo.svg" alt="UTMStack" className="h-12 w-12" />
            <div className="min-w-0 flex-1">
              <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
                {isEnterprise ? t('about.edition.enterprise') : t('about.edition.community')}
              </div>
              <div className="mt-0.5 flex items-baseline gap-2">
                <span className="text-3xl font-semibold">UTMStack</span>
                <span className="font-mono text-lg text-muted-foreground">{version}</span>
              </div>
            </div>
          </div>

          <dl className="mt-6 grid grid-cols-1 gap-y-2.5 text-xs sm:grid-cols-[160px_1fr]">
            {hostname && (
              <Row k={t('about.fields.hostname')}>
                <span className="font-mono">{hostname}</span>
              </Row>
            )}
            {instanceId && (
              <Row k={t('about.fields.instanceId')}>
                <CopyableMono value={instanceId} />
              </Row>
            )}
            {!hostname && !instanceId && (
              <p className="text-muted-foreground sm:col-span-2">{t('about.fields.noInstanceMeta')}</p>
            )}
          </dl>
        </div>

        {/* Right — license (managed on the License page) */}
        <div className="flex flex-col justify-between bg-muted/20 p-6">
          <div>
            <div className="text-[11px] uppercase tracking-wider text-muted-foreground">
              {t('about.license.title')}
            </div>
            <div className="mt-3 flex items-center gap-2 text-sm font-medium">
              {isEnterprise ? (
                <>
                  <ShieldCheck size={16} className="text-primary" />
                  {license?.mssp ? t('about.license.enterpriseMssp') : t('about.edition.enterprise')}
                </>
              ) : (
                <>
                  <Building2 size={16} className="text-muted-foreground" />
                  {t('about.edition.community')}
                </>
              )}
            </div>
            <dl className="mt-3 space-y-1.5 text-xs">
              <LicRow k={t('about.license.datasources')}>
                {license && !isEnterprise
                  ? t('about.license.communityLimited')
                  : !license || license.datasources === 0
                    ? t('about.license.unlimited')
                    : String(license.datasources)}
              </LicRow>
              {hasExpiry && expiry && (
                <LicRow k={t('about.license.expires')}>
                  <span className="font-mono">{expiry.toLocaleDateString()}</span>
                </LicRow>
              )}
            </dl>
            <p className="mt-3 text-xs text-muted-foreground">{t('about.license.managedNote')}</p>
          </div>
          <Button asChild size="sm" variant="outline" className="mt-4 w-full">
            <Link to="/settings/license">
              <KeyRound size={13} className="mr-1.5" />
              {t('about.license.view')}
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
          void navigator.clipboard.writeText(value)
          setCopied(true)
          setTimeout(() => setCopied(false), 1500)
        }}
        className="rounded p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground"
      >
        {copied ? <Activity size={11} className="text-emerald-500" /> : <FileText size={11} />}
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

function LicRow({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="font-medium">{children}</dd>
    </div>
  )
}

/* ─── System health ────────────────────────────────────────────────────── */

interface ServiceItem {
  key: string
  name: string
  description: string
  status: Status
  version?: string
  detail?: string
  icon: LucideIcon
}

function buildServices(t: ReturnType<typeof useTranslation>['t'], data: AboutData, backendUp: boolean): ServiceItem[] {
  const items: ServiceItem[] = []

  items.push({
    key: 'backend',
    name: t('about.services.backend.name'),
    description: t('about.services.backend.desc'),
    status: backendUp ? 'up' : 'down',
    version: data.version?.version,
    icon: Server,
  })

  // OpenSearch — red is always Down. Yellow means replicas are unassigned: on a
  // single-node cluster that's the expected normal state (no second node to host a
  // replica), so it reads healthy; on a multi-node cluster a yellow is a real
  // degradation. Green is always healthy.
  if (data.cluster) {
    const c = data.cluster
    const status: Status =
      c.status === 'red'
        ? 'down'
        : c.status === 'yellow' && c.number_of_nodes > 1
          ? 'degraded'
          : 'up'
    items.push({
      key: 'opensearch',
      name: t('about.services.opensearch.name'),
      description: t('about.services.opensearch.desc'),
      status,
      detail: t('about.services.opensearch.detail', {
        nodes: c.number_of_nodes,
        shards: c.active_shards,
      }),
      icon: Database,
    })
  } else {
    items.push({
      key: 'opensearch',
      name: t('about.services.opensearch.name'),
      description: t('about.services.opensearch.desc'),
      status: 'unknown',
      detail: t('about.services.unreachable'),
      icon: Database,
    })
  }

  return items
}

function SystemHealthSection({
  data,
  backendUp,
  refreshing,
  onRefresh,
}: {
  data: AboutData
  backendUp: boolean
  refreshing: boolean
  onRefresh: () => void
}) {
  const { t } = useTranslation()
  const services = buildServices(t, data, backendUp)
  const issues = services.filter((s) => s.status === 'down' || s.status === 'degraded').length

  return (
    <Section
      title={t('about.health.title')}
      subtitle={issues === 0 ? t('about.health.allHealthy') : t('about.health.issues', { count: issues })}
      action={
        <Button variant="outline" size="sm" onClick={onRefresh} disabled={refreshing}>
          <RefreshCw size={13} className={cn('mr-1.5', refreshing && 'animate-spin')} />
          {t('about.health.refresh')}
        </Button>
      }
    >
      <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
        {services.map((s) => (
          <ServiceRow key={s.key} service={s} />
        ))}
      </div>
    </Section>
  )
}

function ServiceRow({ service: s }: { service: ServiceItem }) {
  const { t } = useTranslation()
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
          {s.version && <span className="font-mono text-[10px] text-muted-foreground">{s.version}</span>}
        </div>
        <div className="truncate text-[11px] text-muted-foreground">{s.detail || s.description}</div>
      </div>
      <span className="inline-flex items-center gap-1.5 text-[11px]">
        <span className={cn('h-1.5 w-1.5 rounded-full', meta.dot)} />
        <span className={meta.tone}>{t(`about.status.${s.status}`)}</span>
      </span>
    </div>
  )
}

/* ─── Datasource usage ─────────────────────────────────────────────────── */

function DatasourceUsageSection({ usage }: { usage: DatasourceUsage | null }) {
  const { t } = useTranslation()
  if (!usage) return null
  const pct = usage.unlimited || usage.limit === 0 ? 0 : Math.min(100, Math.round((usage.count / usage.limit) * 100))

  return (
    <Section title={t('about.usage.title')} subtitle={t('about.usage.subtitle')}>
      <div className="flex items-baseline gap-2">
        <span className="text-2xl font-semibold tabular-nums">{usage.count}</span>
        <span className="text-sm text-muted-foreground">
          {usage.unlimited || usage.limit === 0
            ? t('about.usage.ofUnlimited')
            : t('about.usage.ofLimit', { limit: usage.limit })}
        </span>
      </div>
      {!usage.unlimited && usage.limit > 0 && (
        <div className="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-muted">
          <div
            className={cn('h-full rounded-full', usage.overLimit ? 'bg-red-500' : 'bg-primary')}
            style={{ width: `${Math.max(4, pct)}%` }}
          />
        </div>
      )}
      {usage.overLimit && (
        <p className="mt-2 text-xs text-red-500">{t('about.usage.overLimit')}</p>
      )}
    </Section>
  )
}

/* ─── Support / Legal (static, real links) ─────────────────────────────── */

function SupportSection() {
  const { t } = useTranslation()
  return (
    <Section title={t('about.support.title')}>
      <ul className="space-y-1.5 text-sm">
        <LinkRow icon={FileText} label={t('about.support.docs')} href="https://docs.utmstack.com" />
        <LinkRow icon={Github} label={t('about.support.repo')} href="https://github.com/utmstack/UTMStack" />
        <LinkRow
          icon={Activity}
          label={t('about.support.issues')}
          href="https://github.com/utmstack/UTMStack/issues"
        />
        <LinkRow icon={Globe} label={t('about.support.website')} href="https://utmstack.com" />
      </ul>
    </Section>
  )
}

function LegalSection() {
  const { t } = useTranslation()
  return (
    <Section title={t('about.legal.title')}>
      <ul className="space-y-1.5 text-sm">
        <LinkRow
          icon={FileText}
          label={t('about.legal.license')}
          href="https://github.com/utmstack/UTMStack/blob/main/LICENSE"
        />
        <LinkRow
          icon={ShieldCheck}
          label={t('about.legal.security')}
          href="https://github.com/utmstack/UTMStack/security"
        />
      </ul>
      <div className="mt-4 text-[10px] leading-relaxed text-muted-foreground">
        {t('about.legal.copyright', { year: new Date().getFullYear() })}
      </div>
    </Section>
  )
}

function LinkRow({ icon: Icon, label, href }: { icon: LucideIcon; label: string; href: string }) {
  return (
    <li>
      <a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        className="group flex items-center gap-3 rounded-md border border-transparent px-2 py-1.5 transition-colors hover:border-border hover:bg-muted/40"
      >
        <Icon size={14} className="shrink-0 text-muted-foreground group-hover:text-foreground" />
        <span className="min-w-0 flex-1 text-xs font-medium">{label}</span>
        <ExternalLink size={11} className="shrink-0 text-muted-foreground opacity-0 group-hover:opacity-100" />
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
