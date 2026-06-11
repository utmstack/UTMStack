import { useMemo } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createDashboardsService } from '@/features/dashboard/service/dashboards.service'
import type {
  Dashboard,
  DashboardCreateInput,
  DashboardListParams,
  DashboardUpdateInput,
} from '@/features/dashboard/types'
import type { Paged } from '@/shared/lib/api-client'

export const DASHBOARDS_QUERY_KEYS = {
  all: ['dashboards'] as const,
  list: (params: DashboardListParams) => [...DASHBOARDS_QUERY_KEYS.all, 'list', params] as const,
  one: (id: number) => [...DASHBOARDS_QUERY_KEYS.all, 'one', id] as const,
}

export function useDashboards(params: DashboardListParams = {}) {
  const queryClient = useQueryClient()
  const service = useMemo(() => createDashboardsService(), [])

  const list = useQuery<Paged<Dashboard[]>>({
    queryKey: DASHBOARDS_QUERY_KEYS.list(params),
    queryFn: () => service.listDashboards(params),
  })

  const createDashboard = useMutation({
    mutationFn: (data: DashboardCreateInput) => service.createDashboard(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DASHBOARDS_QUERY_KEYS.all })
    },
  })

  const updateDashboard = useMutation({
    mutationFn: (data: DashboardUpdateInput) => service.updateDashboard(data),
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: DASHBOARDS_QUERY_KEYS.all })
      queryClient.invalidateQueries({ queryKey: DASHBOARDS_QUERY_KEYS.one(vars.id) })
    },
  })

  const deleteDashboard = useMutation({
    mutationFn: (id: number) => service.deleteDashboard(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DASHBOARDS_QUERY_KEYS.all })
    },
  })

  return { list, createDashboard, updateDashboard, deleteDashboard }
}

export function useDashboard(id: number | null) {
  const service = useMemo(() => createDashboardsService(), [])
  return useQuery<Dashboard>({
    queryKey: DASHBOARDS_QUERY_KEYS.one(id ?? 0),
    queryFn: () => service.getDashboard(id as number),
    enabled: id != null && id > 0,
  })
}
