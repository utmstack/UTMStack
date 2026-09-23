import { useCallback, useEffect, useMemo, useState } from 'react'
import { Activity, AlertTriangle, Download, HeartPulse, Loader2, Network, RefreshCw, ShieldAlert, ShieldCheck, Siren, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { cn } from '@/shared/lib/utils'
import { edrHttpService } from '../services/edr-http.service'
import type { EdrBulkAction, EdrEndpoint, EdrOverview } from '../types/edr.types'
import { EndpointHealthGrid } from '../components/EndpointHealthGrid'

const value = (input: unknown, fallback = 0) => typeof input === 'number' && Number.isFinite(input) ? input : fallback

const DEMO_ENDPOINTS: EdrEndpoint[] = [
  { id: 'demo-01', hostName: 'FIN-WS-042', operatingSystem: 'Windows 11', moduleVersion: '7.4.2', serviceState: 'running', engineHealth: 'healthy', signatureAge: 1, sensors: ['File', 'Process', 'Network'], assignedPolicy: 'Corporate Standard', policyDrift: false, detectionsLastWeek: 0, quarantinedFiles: 0, isolationState: 'normal', lastReportTime: '2026-09-23T10:42:00Z' },
  { id: 'demo-02', hostName: 'ENG-LT-017', operatingSystem: 'Windows 10', moduleVersion: '7.4.1', serviceState: 'running', engineHealth: 'stale', signatureAge: 12, sensors: ['File', 'Process'], assignedPolicy: 'Engineering', policyDrift: true, detectionsLastWeek: 4, quarantinedFiles: 2, isolationState: 'alert', lastReportTime: '2026-09-23T09:18:00Z' },
  { id: 'demo-03', hostName: 'HR-WS-008', operatingSystem: 'Windows 11', moduleVersion: '7.4.2', serviceState: 'running', engineHealth: 'unhealthy', signatureAge: 3, sensors: ['File', 'Network'], assignedPolicy: 'Restricted Users', policyDrift: false, detectionsLastWeek: 7, quarantinedFiles: 5, isolationState: 'responding', lastReportTime: '2026-09-23T10:01:00Z' },
  { id: 'demo-04', hostName: 'SRV-FILE-02', operatingSystem: 'Windows Server 2022', moduleVersion: '7.3.8', serviceState: 'degraded', engineHealth: 'degraded', signatureAge: 2, sensors: ['File', 'Network'], assignedPolicy: 'Server Baseline', policyDrift: true, detectionsLastWeek: 2, quarantinedFiles: 1, isolationState: 'isolated', lastReportTime: '2026-09-23T08:44:00Z' },
  { id: 'demo-05', hostName: 'MKT-MAC-003', operatingSystem: 'macOS 14', moduleVersion: '6.9.0', serviceState: 'running', engineHealth: 'off', signatureAge: 28, sensors: [], assignedPolicy: 'Marketing', policyDrift: true, detectionsLastWeek: 1, quarantinedFiles: 0, isolationState: 'normal', lastReportTime: '2026-09-22T17:33:00Z' },
  { id: 'demo-06', hostName: 'OPS-LNX-011', operatingSystem: 'Ubuntu 24.04', moduleVersion: '7.4.2', serviceState: 'running', engineHealth: 'healthy', signatureAge: 0, sensors: ['Process', 'Network'], assignedPolicy: 'Linux Servers', policyDrift: false, detectionsLastWeek: 3, quarantinedFiles: 0, isolationState: 'normal', lastReportTime: '2026-09-23T10:39:00Z' },
]

const DEMO_OVERVIEW: EdrOverview = {
  coverage: { protected: 42, unprotected: 6, total: 48 },
  health: { healthy: 35, unhealthyEngine: 2, staleSignatures: 4, degradedNetwork: 3, sensorsOff: 2, policyDrift: 5 },
  protection: { alertMode: 8, responding: 3, isolated: 2 },
  activity: {
    detectionsOverTime: [3, 1, 5, 2, 8, 4, 6, 2, 9, 5, 3, 7].map((count, index) => ({ timestamp: `2026-09-${12 + index}`, source: index % 2 ? 'Network' : 'File', count })),
    ransomwareContainments: 2,
    blockedConnections: 147,
    topDetections: [{ name: 'Suspicious PowerShell', count: 18 }, { name: 'Credential dumping', count: 11 }, { name: 'Malicious URL', count: 7 }],
    affectedEndpoints: [{ hostName: 'HR-WS-008', count: 7 }, { hostName: 'ENG-LT-017', count: 4 }, { hostName: 'SRV-FILE-02', count: 2 }],
  },
}

function StatCard({ label, number, detail, tone = 'text-foreground', icon: Icon }: { label: string; number: number; detail: string; tone?: string; icon: typeof ShieldCheck }) {
  return <div className="rounded-xl border border-border bg-card p-4"><div className="flex items-start justify-between gap-3"><div><p className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</p><p className={cn('mt-2 text-3xl font-semibold tabular-nums', tone)}>{number.toLocaleString()}</p></div><Icon size={18} className={tone} /></div><p className="mt-2 text-xs text-muted-foreground">{detail}</p></div>
}

function normalizeOverview(payload: EdrOverview): EdrOverview {
  return payload ?? {}
}

export function EdrDashboardPage() {
  const { t } = useTranslation()
  const [overview, setOverview] = useState<EdrOverview>({})
  const [endpoints, setEndpoints] = useState<EdrEndpoint[]>([])
  const [selectedIds, setSelectedIds] = useState<string[]>([])
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
      setSelectedIds((current) => current.filter((id) => nextEndpoints.some((endpoint) => endpoint.id === id)))
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : t('edr.errors.load'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => { void load() }, [load])

  const displayedOverview = showDemo ? DEMO_OVERVIEW : overview
  const displayedEndpoints = showDemo ? DEMO_ENDPOINTS : endpoints
  const displayedCoverage = displayedOverview.coverage ?? {}
  const health = displayedOverview.health ?? {}
  const protection = displayedOverview.protection ?? {}
  const activity = displayedOverview.activity ?? {}
  const topDetections = useMemo(() => activity.topDetections ?? [], [activity.topDetections])
  const affectedEndpoints = useMemo(() => activity.affectedEndpoints ?? [], [activity.affectedEndpoints])
  const runBulkAction = async (action: EdrBulkAction) => {
    if (showDemo) {
      toast.success(t(`edr.actions.success.${action}`))
      setSelectedIds([])
      return
    }
    try {
      await edrHttpService.bulkAction(action, selectedIds)
      toast.success(t(`edr.actions.success.${action}`))
      setSelectedIds([])
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
          <Button variant={showDemo ? 'default' : 'outline'} size="sm" onClick={() => { setShowDemo((current) => !current); setSelectedIds([]); setSelectedEndpoint(null) }}>
            {showDemo ? t('edr.demo.hide', { defaultValue: 'Hide demo data' }) : t('edr.demo.show', { defaultValue: 'Show demo data' })}
          </Button>
          <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}><RefreshCw size={14} className={cn('mr-1.5', loading && 'animate-spin')} />{t('edr.refresh')}</Button>
        </div>
      </header>

      {showDemo && <div className="mb-4 rounded-lg border border-sky-500/30 bg-sky-500/5 px-4 py-3 text-xs text-sky-700 dark:text-sky-300">{t('edr.demo.notice', { defaultValue: 'Demo data is local and does not affect the EDR backend.' })}</div>}

      {error && <div className="mb-4 flex items-center justify-between gap-3 rounded-lg border border-red-500/30 bg-red-500/5 px-4 py-3 text-sm text-red-600 dark:text-red-300"><span className="flex items-center gap-2"><AlertTriangle size={16} />{error}</span><Button size="sm" variant="outline" onClick={() => void load()}>{t('edr.retry')}</Button></div>}

      <section className="mb-5 grid gap-3 lg:grid-cols-3">
        <div className="rounded-xl border border-border bg-card p-4 lg:col-span-1"><div className="flex items-center justify-between"><p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">{t('edr.coverage.title')}</p><ShieldCheck size={18} className="text-primary" /></div><div className="mt-4 grid grid-cols-2 gap-4"><div><p className="text-4xl font-semibold tabular-nums text-red-600 dark:text-red-400">{value(displayedCoverage.unprotected).toLocaleString()}</p><p className="mt-1 text-xs font-medium">{t('edr.coverage.unprotected')}</p></div><div><p className="text-3xl font-semibold tabular-nums text-emerald-600 dark:text-emerald-400">{value(displayedCoverage.protected).toLocaleString()}</p><p className="mt-1 text-xs text-muted-foreground">{t('edr.coverage.protected')}</p></div></div><div className="mt-4 h-2 overflow-hidden rounded-full bg-muted"><div className="h-full rounded-full bg-emerald-500" style={{ width: `${Math.min(100, (value(displayedCoverage.protected) / Math.max(1, value(displayedCoverage.total, displayedEndpoints.length))) * 100)}%` }} /></div><p className="mt-2 text-[11px] text-muted-foreground">{t('edr.coverage.total', { count: value(displayedCoverage.total, displayedEndpoints.length) })}</p></div>
        <div className="rounded-xl border border-border bg-card p-4 lg:col-span-1"><div className="flex items-center justify-between"><p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">{t('edr.health.title')}</p><HeartPulse size={18} className="text-emerald-500" /></div><div className="mt-3 grid grid-cols-2 gap-y-2 text-xs"><span className="text-muted-foreground">{t('edr.health.healthy')}</span><strong className="text-right text-emerald-600 dark:text-emerald-400">{value(health.healthy)}</strong><span className="text-muted-foreground">{t('edr.health.unhealthy')}</span><strong className="text-right text-red-600 dark:text-red-400">{value(health.unhealthyEngine)}</strong><span className="text-muted-foreground">{t('edr.health.stale')}</span><strong className="text-right text-amber-600 dark:text-amber-400">{value(health.staleSignatures)}</strong><span className="text-muted-foreground">{t('edr.health.degraded')}</span><strong className="text-right text-orange-600 dark:text-orange-400">{value(health.degradedNetwork)}</strong><span className="text-muted-foreground">{t('edr.health.off')}</span><strong className="text-right text-slate-500">{value(health.sensorsOff)}</strong><span className="text-muted-foreground">{t('edr.health.drift')}</span><strong className="text-right text-amber-600 dark:text-amber-400">{value(health.policyDrift)}</strong></div></div>
        <div className="rounded-xl border border-border bg-card p-4 lg:col-span-1"><div className="flex items-center justify-between"><p className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">{t('edr.protection.title')}</p><ShieldAlert size={18} className="text-orange-500" /></div><div className="mt-4 grid grid-cols-3 gap-2"><StatCard label={t('edr.protection.alert')} number={value(protection.alertMode)} detail={t('edr.protection.mode')} tone="text-amber-600 dark:text-amber-400" icon={Siren} /><StatCard label={t('edr.protection.responding')} number={value(protection.responding)} detail={t('edr.protection.mode')} tone="text-sky-600 dark:text-sky-400" icon={Activity} /><StatCard label={t('edr.protection.isolated')} number={value(protection.isolated)} detail={t('edr.protection.now')} tone="text-red-600 dark:text-red-400" icon={Network} /></div></div>
      </section>

      <section className="mb-5 grid gap-3 xl:grid-cols-[1.4fr_1fr_1fr]"><div className="rounded-xl border border-border bg-card p-4"><div className="flex items-center gap-2 text-sm font-medium"><Activity size={16} className="text-primary" />{t('edr.activity.detections')}</div><div className="mt-4 flex h-28 items-end gap-1 border-b border-border/60">{(activity.detectionsOverTime ?? []).slice(-24).map((point, index) => <div key={`${point.timestamp}-${index}`} className="min-w-1.5 flex-1 rounded-t bg-primary/70" style={{ height: `${Math.max(6, Math.min(100, value(point.count) * 10))}%` }} title={`${point.source ?? ''}: ${point.count ?? 0}`} />)}</div><div className="mt-3 flex flex-wrap gap-2 text-[11px] text-muted-foreground">{Array.from(new Set((activity.detectionsOverTime ?? []).map((point) => point.source).filter(Boolean))).map((source) => <span key={source} className="rounded bg-muted px-2 py-1">{source}</span>)}</div></div><div className="rounded-xl border border-border bg-card p-4"><div className="flex items-center gap-2 text-sm font-medium"><Download size={16} className="text-amber-500" />{t('edr.activity.ransomware')}</div><p className="mt-5 text-4xl font-semibold text-amber-600 dark:text-amber-400">{value(activity.ransomwareContainments)}</p><p className="mt-1 text-xs text-muted-foreground">{t('edr.activity.containments')}</p><p className="mt-4 text-xs text-muted-foreground">{t('edr.activity.blockedConnections')}: <strong className="text-foreground">{value(activity.blockedConnections)}</strong></p></div><div className="rounded-xl border border-border bg-card p-4"><div className="flex items-center gap-2 text-sm font-medium"><Siren size={16} className="text-red-500" />{t('edr.activity.topDetections')}</div><div className="mt-3 space-y-2">{topDetections.slice(0, 5).map((item) => <div key={item.name} className="flex items-center justify-between gap-3 text-xs"><span className="truncate">{item.name ?? '—'}</span><strong>{value(item.count)}</strong></div>)}{topDetections.length === 0 && <span className="text-xs text-muted-foreground">{t('edr.activity.empty')}</span>}</div><div className="mt-4 border-t border-border/60 pt-3 text-[11px] text-muted-foreground">{t('edr.activity.affectedEndpoints')}: {affectedEndpoints.slice(0, 3).map((item) => `${item.hostName ?? '—'} (${value(item.count)})`).join(', ') || '—'}</div></div></section>

      {!showDemo && loading && endpoints.length === 0 ? <div className="flex items-center justify-center gap-2 rounded-xl border border-border bg-card py-20 text-sm text-muted-foreground"><Loader2 size={16} className="animate-spin" />{t('edr.loading')}</div> : <EndpointHealthGrid endpoints={displayedEndpoints} selectedIds={selectedIds} onSelectionChange={setSelectedIds} onOpen={setSelectedEndpoint} onBulkAction={(action) => void runBulkAction(action)} t={t} />}

      {selectedEndpoint && <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40" onClick={() => setSelectedEndpoint(null)}><aside className="w-full max-w-[520px] overflow-y-auto border-l border-border bg-card p-6 shadow-xl" onClick={(event) => event.stopPropagation()}><div className="flex items-start justify-between gap-3"><div><p className="text-xs uppercase tracking-wider text-primary">{t('edr.endpoint.title')}</p><h2 className="mt-1 text-xl font-semibold">{selectedEndpoint.hostName}</h2></div><button onClick={() => setSelectedEndpoint(null)} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted"><X size={16} /></button></div><div className="mt-6 grid grid-cols-2 gap-3">{[['operatingSystem', selectedEndpoint.operatingSystem], ['moduleVersion', selectedEndpoint.moduleVersion], ['serviceState', selectedEndpoint.serviceState], ['engineHealth', selectedEndpoint.engineHealth], ['assignedPolicy', selectedEndpoint.assignedPolicy], ['lastReportTime', selectedEndpoint.lastReportTime || '—']].map(([key, value]) => <div key={key} className="rounded-lg border border-border p-3"><p className="text-[10px] uppercase tracking-wider text-muted-foreground">{t(`edr.columns.${key}`)}</p><p className="mt-1 break-words text-sm">{value}</p></div>)}</div><div className="mt-4 rounded-lg border border-border p-4"><p className="text-xs font-medium">{t('edr.columns.sensors')}</p><p className="mt-2 text-sm text-muted-foreground">{selectedEndpoint.sensors.join(', ') || '—'}</p></div></aside></div>}
    </div>
  )
}
