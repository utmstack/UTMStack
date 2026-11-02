import { useCallback, useEffect, useState } from 'react'
import { soarFlowsService } from '../services/soar-flows.service'
import { resolveCommonResponse } from '../lib/common-response'
import type { Flow } from '../types/soar.types'
import type { Alert } from '@/features/alerts/types/alert.types'

export interface UseCommonResponseResult {
  /** The system flow relpath that answers this alert, null when none covers it. */
  flow: string | null
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
      .list({ page: 0, size: 1000, systemOwner: true })
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

  const flow = alert && !loading ? resolveCommonResponse(alert, flows) : null

  const trigger = useCallback(async (relPath: string, a: Alert) => {
    const agent = a.target?.host ?? a.dataSource ?? ''
    await soarFlowsService.trigger(relPath, a, agent)
  }, [])

  return { flow, loading, trigger }
}
