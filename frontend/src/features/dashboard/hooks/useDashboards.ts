import { useMemo } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createDashboardsService } from '@/features/dashboard/service/dashboards.service'
import { createVisualizationsService } from '@/features/dashboard/service/visualizations.service'
import { VISUALIZATIONS_QUERY_KEYS } from '@/features/dashboard/hooks/useVisualizations'
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
  one: (id: string) => [...DASHBOARDS_QUERY_KEYS.all, 'one', id] as const,
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
    onSuccess: (updated, vars) => {
      // Push the server response straight into the cache so consumers observing
      // the single-dashboard query see the change synchronously — no wait for
      // the invalidation-triggered refetch to complete.
      queryClient.setQueryData(DASHBOARDS_QUERY_KEYS.one(vars.id), updated)
      queryClient.invalidateQueries({ queryKey: DASHBOARDS_QUERY_KEYS.all })
    },
  })

  const deleteDashboard = useMutation({
    mutationFn: (id: string) => service.deleteDashboard(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DASHBOARDS_QUERY_KEYS.all })
    },
  })

  // Widgets are their own rows, not part of the dashboard document, so
  // "copy an existing dashboard" is a client-side orchestration: read the
  // source's dashboard + widgets, then recreate them under a new dashboard id
  // via the same create endpoints a manual create/add-widget already uses.
  // Nothing server-side needs to know a copy happened.
  const duplicateDashboard = useMutation({
    mutationFn: async ({ sourceId, name }: { sourceId: string; name: string }) => {
      const vizService = createVisualizationsService()
      const [source, sourceViz] = await Promise.all([
        service.getDashboard(sourceId),
        vizService.listVisualizations({ dashboardId: sourceId, page: 0, size: 500 }),
      ])
      const created = await service.createDashboard({
        name,
        description: source.description,
        config: source.config,
      })
      await Promise.all(
        (sourceViz.data ?? []).map((v) =>
          vizService.createVisualization({
            dashboardId: created.id,
            spec: v.spec,
            config: v.config,
            layout: v.layout,
          })
        )
      )
      return created
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DASHBOARDS_QUERY_KEYS.all })
      queryClient.invalidateQueries({ queryKey: VISUALIZATIONS_QUERY_KEYS.all })
    },
  })

  return { list, createDashboard, updateDashboard, deleteDashboard, duplicateDashboard }
}

export function useDashboard(id: string | null) {
  const service = useMemo(() => createDashboardsService(), [])
  return useQuery<Dashboard>({
    queryKey: DASHBOARDS_QUERY_KEYS.one(id ?? ''),
    queryFn: () => service.getDashboard(id as string),
    enabled: !!id,
  })
}
