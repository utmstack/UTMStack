import { useCallback, useEffect, useState } from 'react'
import { Activity, AlertTriangle, Download, HeartPulse, Loader2, Network, RefreshCw, ShieldCheck, Siren, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { cn } from '@/shared/lib/utils'
import { edrHttpService } from '../services/edr-http.service'
import type { EdrBulkAction, EdrEndpoint, EdrOverview } from '../types/edr.types'
import { EndpointHealthGrid } from '../components/EndpointHealthGrid'

const DEMO_ENDPOINTS: EdrEndpoint[] = [
  { agentId: 1, hostname: 'FIN-WS-042', ip: '10.0.4.21', os: 'Windows', platform: 'windows', version: '7.4.2', status: 'online', lastSeen: '2026-09-23T10:42:00Z', hasModule: true, moduleHealthy: true, sigdbVersion: '20260923', staleSignature: false, policyVersion: 'Corporate Standard', drift: false, lastReport: '2026-09-23T10:42:00Z' },
  { agentId: 2, hostname: 'ENG-LT-017', ip: '10.0.7.102', os: 'Windows', platform: 'windows', version: '7.4.1', status: 'online', lastSeen: '2026-09-23T09:18:00Z', hasModule: true, moduleHealthy: true, sigdbVersion: '20260911', staleSignature: true, policyVersion: 'Engineering-2', drift: true, lastReport: '2026-09-23T09:18:00Z' },
  { agentId: 3, hostname: 'HR-WS-008', ip: '10.0.9.33', os: 'Windows', platform: 'windows', version: '7.4.2', status: 'online', lastSeen: '2026-09-23T10:01:00Z', hasModule: true, moduleHealthy: false, sigdbVersion: '20260923', staleSignature: false, policyVersion: 'Restricted Users', drift: false, lastReport: '2026-09-23T10:01:00Z' },
  { agentId: 4, hostname: 'SRV-FILE-02', ip: '10.0.1.40', os: 'Windows', platform: 'windows', version: '7.3.8', status: 'offline', lastSeen: '2026-09-22T08:44:00Z', hasModule: true, moduleHealthy: false, sigdbVersion: '20260921', staleSignature: false, policyVersion: 'Server Baseline', drift: true, lastReport: '2026-09-23T08:44:00Z' },
  { agentId: 5, hostname: 'MKT-MAC-003', ip: '10.0.12.7', os: 'macOS', platform: 'darwin', version: '6.9.0', status: 'online', lastSeen: '2026-09-22T17:33:00Z', hasModule: true, moduleHealthy: false, sigdbVersion: '20260826', staleSignature: true, policyVersion: 'Marketing', drift: true, lastReport: '2026-09-22T17:33:00Z' },
  { agentId: 6, hostname: 'OPS-LNX-011', ip: '10.0.3.55', os: 'Linux', platform: 'linux', version: '7.4.2', status: 'online', lastSeen: '2026-09-23T10:39:00Z', hasModule: false, moduleHealthy: false, staleSignature: false, drift: false, lastReport: null },
]

const DEMO_OVERVIEW: EdrOverview = {
  endpoints: 48,
  withModule: 42,
  healthy: 35,
  unhealthy: 2,
  staleSignatures: 4,
  drifted: 5,
  detectionsToday: 9,
}

function StatCard({ label, number, detail, tone = 'text-foreground', icon: Icon }: { label: string; number: number; detail: string; tone?: string; icon: typeof ShieldCheck }) {
  return <div className="rounded-xl border border-border bg-card p-4"><div className="flex items-start justify-between gap-3"><div><p className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</p><p className={cn('mt-2 text-3xl font-semibold tabular-nums', tone)}>{number.toLocaleString()}</p></div><Icon size={18} className={tone} /></div><p className="mt-2 text-xs text-muted-foreground">{detail}</p></div>
}

const normalizeOverview = (payload: EdrOverview): EdrOverview => payload

export function EdrDashboardPage() {
  const { t } = useTranslation()
  const EMPTY_OVERVIEW: EdrOverview = { endpoints: 0, withModule: 0, healthy: 0, unhealthy: 0, staleSignatures: 0, drifted: 0, detectionsToday: 0 }
  const [overview, setOverview] = useState<EdrOverview>(EMPTY_OVERVIEW)
  const [endpoints, setEndpoints] = useState<EdrEndpoint[]>([])
  const [selectedAgentIds, setSelectedAgentIds] = useState<number[]>([])
  const selectedIds = selectedAgentIds.map(String)
  const [selectedEndpoint, setSelectedEndpoint] = useState<EdrEndpoint | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showDemo, setShowDemo] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [nextOverview, nextEndpoints] = await Promise.all([edrHttpService.overview(), edrHttpService.endpoints()])
      setOverview(normalizeOverview(nextOverview))
      setEndpoints(nextEndpoints)
      setSelectedAgentIds((current) => current.filter((id) => nextEndpoints.some((endpoint) => endpoint.agentId === id)))
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : t('edr.errors.load'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => { void load() }, [load])

  const displayedOverview = showDemo ? DEMO_OVERVIEW : overview
  const displayedEndpoints = showDemo ? DEMO_ENDPOINTS : endpoints
  const coverage = { total: displayedOverview.endpoints, protected: displayedOverview.withModule, unprotected: displayedOverview.endpoints - displayedOverview.withModule }
  const runBulkAction = async (action: EdrBulkAction) => {
    if (showDemo) {
      toast.success(t(`edr.actions.success.${action}`))
      setSelectedAgentIds([])
      return
    }
    try {
      await edrHttpService.bulkAction(action, selectedAgentIds)
      toast.success(t(`edr.actions.success.${action}`))
      setSelectedAgentIds([])
      void load()
    } catch {
      toast.error(t('edr.errors.action'))
    }
  }

  return (
    <div className="flex min-h-full flex-col px-6 pb-6 pt-4">
      <header className="mb-5 flex flex-wrap items-start justify-between gap-4">
        <div><div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wider text-primary"><ShieldCheck size={15} />{t('edr.eyebrow')}</div><h1 className="mt-1 text-2xl font-semibold tracking-tight">{t('edr.title')}</h1><p className="mt-1 max-w-2xl text-sm text-muted-foreground">{t('edr.subtitle')}</p></div>
        <div className="flex flex-wrap items-center gap-2">
          <Button variant={showDemo ? 'default' : 'outline'} size="sm" onClick={() => { setShowDemo((current) => !current); setSelectedAgentIds([]); setSelectedEndpoint(null) }}>
            {showDemo ? t('edr.demo.hide', { defaultValue: 'Hide demo data' }) : t('edr.demo.show', { defaultValue: 'Show demo data' })}
          </Button>
          <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw size={14} className={cn('mr-1.5', loading && 'animate-spin')} />{t('edr.refresh')}</Button>
        </div>
      </header>

      {showDemo && <div className="mb-4 rounded-lg border border-sky-500/30 bg-sky-500/5 px-4 py-3 text-xs text-sky-700 dark:text-sky-300">{t('edr.demo.notice', { defaultValue: 'Demo data is local and does not affect the EDR backend.' })}</div>}

      {error && <div className="mb-4 flex items-center justify-between gap-3 rounded-lg border border-red-500/30 bg-red-500/5 px-4 py-3 text-sm text-red-600 dark:text-red-300"><span className="flex items-center gap-2"><AlertTriangle size={16} />{error}</span><Button size="sm" variant="outline" onClick={() => void load()}>{t('edr.retry')}</Button></div>}

      <section className="mb-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-6">
        <StatCard label={t('edr.coverage.title')} number={coverage.protected} detail={`${coverage.unprotected} ${t('edr.coverage.unprotected')}`} tone="text-emerald-600 dark:text-emerald-400" icon={ShieldCheck} />
        <StatCard label={t('edr.health.healthy')} number={displayedOverview.healthy} detail={t('edr.health.title')} tone="text-emerald-600 dark:text-emerald-400" icon={HeartPulse} />
        <StatCard label={t('edr.health.unhealthy')} number={displayedOverview.unhealthy} detail={t('edr.health.title')} tone="text-red-600 dark:text-red-400" icon={Siren} />
        <StatCard label={t('edr.health.stale')} number={displayedOverview.staleSignatures} detail={t('edr.health.title')} tone="text-amber-600 dark:text-amber-400" icon={Download} />
        <StatCard label={t('edr.health.drift')} number={displayedOverview.drifted} detail={t('edr.health.title')} tone="text-orange-600 dark:text-orange-400" icon={Network} />
        <StatCard label={t('edr.activity.detections')} number={displayedOverview.detectionsToday} detail="24h" tone="text-primary" icon={Activity} />
      </section>

      {!showDemo && loading && endpoints.length === 0 ? <div className="flex items-center justify-center gap-2 rounded-xl border border-border bg-card py-20 text-sm text-muted-foreground"><Loader2 size={16} className="animate-spin" />{t('edr.loading')}</div> : <EndpointHealthGrid endpoints={displayedEndpoints} selectedIds={selectedIds} onSelectionChange={(ids) => setSelectedAgentIds(ids.map(Number))} onOpen={setSelectedEndpoint} onBulkAction={(action) => void runBulkAction(action)} t={t} />}

      {selectedEndpoint && <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40" onClick={() => setSelectedEndpoint(null)}><aside className="w-full max-w-[520px] overflow-y-auto border-l border-border bg-card p-6 shadow-xl" onClick={(event) => event.stopPropagation()}><div className="flex items-start justify-between gap-3"><div><p className="text-xs uppercase tracking-wider text-primary">{t('edr.endpoint.title')}</p><h2 className="mt-1 text-xl font-semibold">{selectedEndpoint.hostname}</h2><p className="text-xs text-muted-foreground">{selectedEndpoint.ip}</p></div><button onClick={() => setSelectedEndpoint(null)} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted"><X size={16} /></button></div><div className="mt-6 grid grid-cols-2 gap-3">{[['os', selectedEndpoint.os], ['platform', selectedEndpoint.platform], ['version', selectedEndpoint.version], ['status', selectedEndpoint.status], ['moduleHealthy', selectedEndpoint.moduleHealthy ? t('edr.health.healthy') : t('edr.health.unhealthy')], ['sigdbVersion', selectedEndpoint.sigdbVersion ?? '—'], ['staleSignature', selectedEndpoint.staleSignature ? t('edr.health.stale') : t('edr.values.inSync')], ['policyVersion', selectedEndpoint.policyVersion ?? '—'], ['drift', selectedEndpoint.drift ? t('edr.values.drifted') : t('edr.values.inSync')], ['lastSeen', selectedEndpoint.lastSeen], ['lastReport', selectedEndpoint.lastReport || '—']].map(([key, cellValue]) => <div key={key} className="rounded-lg border border-border p-3"><p className="text-[10px] uppercase tracking-wider text-muted-foreground">{t(`edr.columns.${key}`)}</p><p className="mt-1 break-words text-sm">{cellValue}</p></div>)}</div></aside></div>}
    </div>
  )
}
