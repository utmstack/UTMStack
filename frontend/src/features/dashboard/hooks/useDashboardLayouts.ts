import { useMemo } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createDashboardLayoutsService } from '@/features/dashboard/service/dashboard-layouts.service'
import type {
  DashboardLayoutCreateInput,
  DashboardLayoutListParams,
  DashboardLayoutUpdateInput,
  DashboardVisualization,
} from '@/features/dashboard/types'
import type { Paged } from '@/shared/lib/api-client'

export const DASHBOARD_LAYOUTS_QUERY_KEYS = {
  all: ['dashboard-layouts'] as const,
  list: (params: DashboardLayoutListParams) =>
    [...DASHBOARD_LAYOUTS_QUERY_KEYS.all, 'list', params] as const,
  one: (id: number) => [...DASHBOARD_LAYOUTS_QUERY_KEYS.all, 'one', id] as const,
}

export function useDashboardLayouts(params: DashboardLayoutListParams) {
  const queryClient = useQueryClient()
  const service = useMemo(() => createDashboardLayoutsService(), [])
  const enabled = params.idDashboard != null && params.idDashboard > 0

  const list = useQuery<Paged<DashboardVisualization[]>>({
    queryKey: DASHBOARD_LAYOUTS_QUERY_KEYS.list(params),
    queryFn: () => service.listLayouts({ size: 500, ...params }),
    enabled,
  })

  const createLayout = useMutation({
    mutationFn: (data: DashboardLayoutCreateInput) => service.createLayout(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DASHBOARD_LAYOUTS_QUERY_KEYS.all })
    },
  })

  const updateLayout = useMutation({
    mutationFn: (data: DashboardLayoutUpdateInput) => service.updateLayout(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DASHBOARD_LAYOUTS_QUERY_KEYS.all })
    },
  })

  const deleteLayout = useMutation({
    mutationFn: (id: number) => service.deleteLayout(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DASHBOARD_LAYOUTS_QUERY_KEYS.all })
    },
  })

  return { list, createLayout, updateLayout, deleteLayout }
}
