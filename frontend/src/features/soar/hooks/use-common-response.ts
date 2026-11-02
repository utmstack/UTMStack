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
export function useCommonResponse(alert: Alert | null | undefined): UseCommonResponseResult {
  const [flows, setFlows] = useState<Flow[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!alert) return
    let alive = true
    setLoading(true)
    soarFlowsService
      .list({ page: 0, size: 20, ...flowNameFilter(alert) })
      .then(({ data }) => {
        if (alive) setFlows(data)
      })
      .catch(() => {
        if (alive) setFlows([])
      })
      .finally(() => {
        if (alive) setLoading(false)
      })
    return () => {
      alive = false
    }
  }, [alert?.id])

  const relPath = alert && !loading ? resolveCommonResponse(alert, flows) : null

  console.log(relPath)

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
