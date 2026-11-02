import { useCallback, useEffect, useState } from 'react'
import { soarFlowsService } from '../services/soar-flows.service'
import { resolveCommonResponse, flowNameFilter } from '../lib/common-response'
import type { Flow } from '../types/soar.types'
import type { Alert } from '@/features/alerts/types/alert.types'

export interface CommonFlow {
  /** Flow display name (from the flow definition). */
  name: string
  /** Relpath without the `.yaml` suffix — what the trigger endpoint expects. */
  relPath: string
}

export interface UseCommonResponseResult {
  /** The system flow that answers this alert, null when none covers it. */
  flow: CommonFlow | null
  loading: boolean
  /**
   * Triggers a flow for the alert. The endpoint agent is extracted from the
   * alert (the host the alert landed on, falling back to its data source —
   * which is the agent's name, as executions store it).
   */
  trigger: (relPath: string, alert: Alert) => Promise<void>
}

/**
 * Shared "run the common response" behavior for alert views: given the alert
 * data, finds the matching system SOAR flow and triggers it on the alert's
 * agent.
 */
// Module cache: an alerts table renders one hook per row; they all resolve the
// same template+platform filter, so fetch each filter key once per session.
const flowsCache = new Map<string, Promise<Flow[]>>()

function fetchFlows(filter: { relPath: string; systemOwner: true }): Promise<Flow[]> {
  let p = flowsCache.get(filter.relPath)
  if (!p) {
    p = soarFlowsService
      .list({ page: 0, size: 20, ...filter })
      .then(({ data }) => data)
      .catch(() => [] as Flow[])
    flowsCache.set(filter.relPath, p)
  }
  return p
}

export function useCommonResponse(alert: Alert | null | undefined): UseCommonResponseResult {
  const [flows, setFlows] = useState<Flow[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!alert) return
    let alive = true
    setLoading(true)
    // relPath filter pre-narrows server-side: only this alert's template+platform flows come back
    fetchFlows(flowNameFilter(alert))
      .then((data) => {
        if (alive) setFlows(data)
      })
      .finally(() => {
        if (alive) setLoading(false)
      })
    return () => {
      alive = false
    }
  }, [alert?.id])

  const relPath = alert && !loading ? resolveCommonResponse(alert, flows) : null

  const flow = relPath ? (() => {
    const f = flows.find((x) => x.relPath === relPath)
    return { name: f?.name ?? relPath, relPath }
  })() : null

  const trigger = useCallback(async (relPath: string, a: Alert) => {
    const agent = a.target?.host ?? a.dataSource ?? ''
    await soarFlowsService.trigger(relPath, a, agent)
  }, [])

  return { flow, loading, trigger }
}
