import { useCallback, useEffect, useState } from 'react'
import { incidentsHttpService as svc } from '../services/incidents-http.service'
import { STATUSES } from '../lib/incident-meta'
import type { Incident, IncidentStatus } from '../types/incident.types'

export interface IncidentsListFilters {
  search: string
  statusFilter: IncidentStatus | 'all'
  assignee: string
  dateFrom: string
  dateTo: string
  page: number // 0-based
  pageSize: number
  isBoard: boolean
}

export interface UseIncidentsListResult {
  incidents: Incident[]
  total: number
  loading: boolean
  error: boolean
  counts: Record<IncidentStatus, number>
  assigneeOptions: string[]
  refresh: () => void
}

/**
 * Server-side incident list with the board/table toggle (board fetches a wider
 * window, no pager). Status counts and assignee options are reloaded on every
 * refresh tick so mutations propagate.
 */
export function useIncidentsList(filters: IncidentsListFilters): UseIncidentsListResult {
  const { search, statusFilter, assignee, dateFrom, dateTo, page, pageSize, isBoard } = filters
  const [incidents, setIncidents] = useState<Incident[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [counts, setCounts] = useState<Record<IncidentStatus, number>>({ Open: 0, 'In review': 0, Completed: 0, Merged: 0 })
  const [assigneeOptions, setAssigneeOptions] = useState<string[]>([])
  const [nonce, setNonce] = useState(0)

  const reqSize = isBoard ? 200 : pageSize
  const reqPage = isBoard ? 1 : page + 1

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(false)
    svc
      .list({
        incidentName: search || undefined,
        incidentStatus: statusFilter === 'all' ? undefined : statusFilter,
        incidentAssignedTo: assignee === 'all' ? undefined : assignee,
        createdDateStart: dateFrom ? `${dateFrom}T00:00:00Z` : undefined,
        createdDateEnd: dateTo ? `${dateTo}T23:59:59Z` : undefined,
        page: reqPage,
        size: reqSize,
      })
      .then(({ data, total }) => {
        if (cancelled) return
        // page is 0-based; the first page replaces, later pages append (board
        // mode always fetches page 0, so it replaces).
        setIncidents((prev) => (page === 0 ? (data ?? []) : [...prev, ...(data ?? [])]))
        setTotal(total)
      })
      .catch(() => !cancelled && setError(true))
      .finally(() => !cancelled && setLoading(false))
    return () => {
      cancelled = true
    }
  }, [search, statusFilter, assignee, dateFrom, dateTo, reqPage, reqSize, nonce])

  useEffect(() => {
    let cancelled = false
    Promise.all(
      STATUSES.map((s) =>
        svc
          .list({ incidentStatus: s, size: 1 })
          .then((r) => [s, r.total] as const)
          .catch(() => [s, 0] as const)
      )
    ).then((pairs) => !cancelled && setCounts(Object.fromEntries(pairs) as Record<IncidentStatus, number>))
    // The assignees are free text, so the filter offers the ones actually in
    // use rather than the platform's user list.
    svc
      .assignees()
      .then((names) => !cancelled && setAssigneeOptions(names))
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [nonce])

  const refresh = useCallback(() => setNonce((n) => n + 1), [])

  return { incidents, total, loading, error, counts, assigneeOptions, refresh }
}
