import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { useQuery } from '@tanstack/react-query'
import { datasourcesHttpService } from '@/features/datasources/services/datasources-http.service'
import { alertsHttpService } from '@/features/alerts/services/alerts-http.service'
import { incidentsHttpService } from '@/features/incidents/services/incidents-http.service'
import { soarFlowsService } from '@/features/soar/services/soar-flows.service'
import { complianceService } from '@/features/compliance/services/compliance-http.service'
import { aboutHttpService } from '@/features/settings/services/about-http.service'
import type { FilterType } from '@/features/alerts/types/alert.types'

/**
 * Data layer for the Overview (/home) page. Every card has its own query so a
 * single failing endpoint degrades one tile instead of blanking the page — the
 * same partial-failure posture the Alerts/Data Sources pages use. The overview
 * is a glanceable summary, so a 1-minute stale window is plenty.
 */
import { SEVERITY_VALUE } from '@/features/alerts/types/alert.types'
import { resolveRange } from '@/shared/components/ui/time-range-picker'

const STALE = 60_000

// A last-N-window scope for the alert endpoints. `parentId IS ''` rolls
// deduplicated child echoes up under their parent (matches the Alerts page
// default). The window is resolved to instants here: the store takes datetimes,
// not the relative tokens the pickers carry around.
const NO_PARENT: FilterType = { field: 'parentId', operator: 'IS', value: '' }
const lastWindow = (fromToken: string): FilterType[] => {
  const { from, to } = resolveRange({ from: fromToken, to: 'now', interval: 'hour' })
  return [NO_PARENT, { field: '@timestamp', operator: 'IS_BETWEEN', value: [from, to] }]
}

/* ─── Datasources: count, active sources, events + ingestion timeline ───────── */

export interface DatasourcesOverview {
  total: number
  activeSources: number
  events: number
  from?: string
  to?: string
  points: { t: string; count: number }[]
  isLoading: boolean
}

export function useDatasourcesOverview(): DatasourcesOverview {
  const count = useQuery({
    queryKey: ['overview', 'ds', 'count'],
    queryFn: () => datasourcesHttpService.count(),
    staleTime: STALE,
  })
  const totals = useQuery({
    queryKey: ['overview', 'ds', 'totals'],
    queryFn: () => datasourcesHttpService.ingestionTotals(),
    staleTime: STALE,
  })
  const timeline = useQuery({
    queryKey: ['overview', 'ds', 'timeline'],
    queryFn: () => datasourcesHttpService.ingestionTimeline(),
    staleTime: STALE,
  })

  return useMemo<DatasourcesOverview>(
    () => ({
      total: count.data?.count ?? 0,
      // Sources that actually shipped events in the stats window (one bucket each).
      activeSources: totals.data?.buckets?.length ?? 0,
      events: totals.data?.total ?? 0,
      from: totals.data?.from ?? timeline.data?.from,
      to: totals.data?.to ?? timeline.data?.to,
      points: (timeline.data?.points ?? []).map((p) => ({ t: p.timestamp, count: p.count })),
      isLoading: count.isLoading || totals.isLoading || timeline.isLoading,
    }),
    [count.data, count.isLoading, totals.data, totals.isLoading, timeline.data, timeline.isLoading],
  )
}

/* ─── Alert KPIs: total + high severity in the last 24h, hourly sparkline ────── */

export interface AlertKpis {
  alerts24h: number
  high24h: number
  sparkline: number[]
  isLoading: boolean
}

export function useAlertKpis(): AlertKpis {
  const sev = useQuery({
    queryKey: ['overview', 'alerts', 'sev24h'],
    queryFn: () => alertsHttpService.counts('severity', lastWindow('now-24h')),
    staleTime: STALE,
  })
  const tl = useQuery({
    queryKey: ['overview', 'alerts', 'timeline24h'],
    queryFn: () => alertsHttpService.timeline(lastWindow('now-24h'), '1h'),
    staleTime: STALE,
  })

  return useMemo<AlertKpis>(() => {
    const buckets = sev.data?.top ?? []
    // Severity is present on every alert and has only 3 values, so summing the
    // buckets is the total alert count for the window regardless of how the
    // endpoint defines its own `total`.
    const alerts24h = buckets.reduce((acc, b) => acc + b.count, 0)
    const high24h = buckets.find((b) => b.value === SEVERITY_VALUE.high)?.count ?? 0
    return {
      alerts24h,
      high24h,
      sparkline: tl.data?.values ?? [],
      isLoading: sev.isLoading || tl.isLoading,
    }
  }, [sev.data, sev.isLoading, tl.data, tl.isLoading])
}

/* ─── Open incidents (count) ────────────────────────────────────────────────── */

export function useOpenIncidents(): { count: number; isLoading: boolean } {
  const q = useQuery({
    queryKey: ['overview', 'incidents', 'open'],
    queryFn: () => incidentsHttpService.list({ incidentStatus: 'Open', page: 1, size: 1 }),
    staleTime: STALE,
  })
  return { count: q.data?.total ?? 0, isLoading: q.isLoading }
}

/* ─── Active playbooks (enabled SOAR flows, count) ──────────────────────────── */

export function useActivePlaybooks(): { count: number; isLoading: boolean } {
  const q = useQuery({
    queryKey: ['overview', 'soar', 'active'],
    queryFn: () => soarFlowsService.list({ active: true, page: 0, size: 1 }),
    staleTime: STALE,
  })
  return { count: q.data?.total ?? 0, isLoading: q.isLoading }
}

/* ─── Top MITRE ATT&CK techniques (by alert volume) ─────────────────────────── */

export interface RankedValue {
  value: string
  count: number
}

export function useTopTechniques(): { items: RankedValue[]; isLoading: boolean } {
  const q = useQuery({
    queryKey: ['overview', 'alerts', 'techniques'],
    queryFn: () => alertsHttpService.fieldValues('technique', 6),
    staleTime: STALE,
  })
  return { items: q.data ?? [], isLoading: q.isLoading }
}

/* ─── Top targeted assets (by alert volume on target.host) ──────────────────── */

export function useTopAssets(): { items: RankedValue[]; isLoading: boolean } {
  const q = useQuery({
    queryKey: ['overview', 'alerts', 'assets'],
    queryFn: () => alertsHttpService.fieldValues('target.host', 6),
    staleTime: STALE,
  })
  return { items: q.data ?? [], isLoading: q.isLoading }
}

/* ─── Compliance scores (latest snapshot per framework) ─────────────────────── */

export interface ComplianceScore {
  key: string
  name: string
  score: number
}

export function useComplianceScores(): { items: ComplianceScore[]; isLoading: boolean } {
  const frameworks = useQuery({
    queryKey: ['overview', 'compliance', 'frameworks'],
    queryFn: () => complianceService.listFrameworks(),
    staleTime: 5 * 60 * 1000,
  })
  const reports = useQuery({
    queryKey: ['overview', 'compliance', 'reports'],
    queryFn: () => complianceService.listReports(),
    staleTime: STALE,
  })

  const items = useMemo<ComplianceScore[]>(() => {
    const names = new Map((frameworks.data ?? []).map((f) => [f.key, f.name]))
    // One report per framework, so the score is simply the row's.
    const latest = new Map<string, number>()
    for (const r of reports.data ?? []) latest.set(r.frameworkKey, r.score)
    return Array.from(latest.entries())
      .map(([key, score]) => ({ key, name: names.get(key) ?? key, score: Math.round(score) }))
      .sort((a, b) => b.score - a.score)
      .slice(0, 6)
  }, [frameworks.data, reports.data])

  return { items, isLoading: frameworks.isLoading || reports.isLoading }
}

/* ─── System health (backend / event store / SOC-AI status) ─────────────────── */

// Mirrors the Settings → About health model: a neutral `unknown` distinct from a
// real problem, so we never raise a false alarm.
export type HealthStatus = 'up' | 'degraded' | 'down' | 'unknown'

export interface HealthService {
  name: string
  status: HealthStatus
  detail: string
}

export function useSystemHealth(): { services: HealthService[]; isLoading: boolean } {
  const { t } = useTranslation()
  const version = useQuery({
    queryKey: ['overview', 'health', 'version'],
    queryFn: () => aboutHttpService.version(),
    staleTime: STALE,
    retry: false,
  })
  const mcp = useQuery({
    queryKey: ['overview', 'health', 'mcp'],
    queryFn: () => aboutHttpService.mcpHealth(),
    staleTime: STALE,
    retry: false,
  })

  const services = useMemo<HealthService[]>(() => {
    // Backend is reachable if any authenticated call resolved (same signal the
    // About page uses); only "everything failed" is a real Down.
    const anyOk = version.isSuccess || mcp.isSuccess
    const allError = version.isError && mcp.isError
    const backend: HealthService = {
      name: t('home.health.services.backend'),
      status: anyOk ? 'up' : allError ? 'down' : 'unknown',
      detail: version.data?.version
        ? version.data.version
        : allError
          ? t('home.health.detail.unreachable')
          : t('home.health.detail.pending'),
    }

    // SOC-AI is optional infra, not core uptime: enabled = Up, while disabled or
    // unreachable read neutral (Unknown) rather than alarming.
    const socai: HealthService = {
      name: t('home.health.services.socai'),
      status: mcp.data?.enabled ? 'up' : 'unknown',
      detail: mcp.data?.enabled
        ? t('home.health.detail.tools', { n: mcp.data.tools_registered })
        : mcp.isError
          ? t('home.health.detail.unreachable')
          : t('home.health.detail.disabled'),
    }

    return [backend, socai]
  }, [t, version.data, version.isSuccess, version.isError, mcp.data, mcp.isSuccess, mcp.isError])

  return {
    services,
    isLoading: version.isLoading || mcp.isLoading,
  }
}
