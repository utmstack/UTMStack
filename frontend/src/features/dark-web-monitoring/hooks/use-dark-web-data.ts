import { useCallback, useEffect, useState } from 'react'
import { darkWebMonitoringService } from '../services'
import type { BreachStatus, DarkWebRecord, OrganizationBreachStats } from '../types'

export interface DarkWebData {
  records: DarkWebRecord[]
  stats: OrganizationBreachStats | null
  loading: boolean
  error: boolean
  retry: () => void
  setRecordStatus: (id: string, status: BreachStatus) => void
}

/**
 * Loads the breach records and organization stats for the dark web page.
 * Data comes from `darkWebMonitoringService` — currently a mock — so swapping
 * in the real API later does not touch the page. Status changes are
 * optimistic: the local list updates immediately, and a failed API write
 * rolls the change back.
 */
export function useDarkWebData(): DarkWebData {
  const [records, setRecords] = useState<DarkWebRecord[]>([])
  const [stats, setStats] = useState<OrganizationBreachStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(false)
    Promise.all([darkWebMonitoringService.listRecords(), darkWebMonitoringService.getOrganizationStats()])
      .then(([recordList, organizationStats]) => {
        if (cancelled) return
        setRecords(recordList)
        setStats(organizationStats)
        setLoading(false)
      })
      .catch(() => {
        if (cancelled) return
        setError(true)
        setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [attempt])

  const setRecordStatus = useCallback((id: string, status: BreachStatus) => {
    // Capture the previous value from the current list, not from inside the
    // state updater: React may not have run the updater by the time the API
    // write rejects, and an optimistic write that loses its rollback would
    // leave the UI showing a status the backend never accepted.
    const previous: BreachStatus | undefined = records.find((record) => record.id === id)?.breachStatus
    setRecords((current) => current.map((record) => (record.id === id ? { ...record, breachStatus: status } : record)))
    void darkWebMonitoringService.updateRecordStatus(id, status).catch(() => {
      if (!previous) return
      setRecords((current) => current.map((record) => (record.id === id ? { ...record, breachStatus: previous } : record)))
    })
  }, [records])

  const retry = useCallback(() => setAttempt((current) => current + 1), [])

  return { records, stats, loading, error, retry, setRecordStatus }
}
