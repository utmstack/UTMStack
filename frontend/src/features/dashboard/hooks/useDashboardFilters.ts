import { useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createDashboardFiltersService } from '@/features/dashboard/service/dashboard-filters.service'
import type { DashboardFilterCreateInput, DashboardFilterUpdateInput } from '@/features/dashboard/types'

export const DASHBOARD_FILTERS_QUERY_KEYS = {
  all: ['dashboardFilters'] as const,
  list: (dashboardId: string) => [...DASHBOARD_FILTERS_QUERY_KEYS.all, 'list', dashboardId] as const,
}

export function useDashboardFilters(dashboardId: string | null) {
  const queryClient = useQueryClient()
  const service = useMemo(() => createDashboardFiltersService(), [])

  const list = useQuery({
    queryKey: DASHBOARD_FILTERS_QUERY_KEYS.list(dashboardId ?? ''),
    queryFn: () => service.listDashboardFilters({ dashboardId: dashboardId!, page: 0, size: 500 }),
    enabled: !!dashboardId,
  })

  const invalidate = () => {
    if (dashboardId) {
      queryClient.invalidateQueries({ queryKey: DASHBOARD_FILTERS_QUERY_KEYS.list(dashboardId) })
    }
  }

  const createFilter = useMutation({
    mutationFn: (data: DashboardFilterCreateInput) => service.createDashboardFilter(data),
    onSuccess: invalidate,
  })

  const updateFilter = useMutation({
    mutationFn: (data: DashboardFilterUpdateInput) => service.updateDashboardFilter(data),
    onSuccess: invalidate,
  })

  const deleteFilter = useMutation({
    mutationFn: (id: string) => service.deleteDashboardFilter(id),
    onSuccess: invalidate,
  })

  return { list, createFilter, updateFilter, deleteFilter }
}
