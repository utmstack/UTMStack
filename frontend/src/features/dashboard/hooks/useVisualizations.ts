import { useMemo } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { createVisualizationsService } from '@/features/dashboard/service/visualizations.service'
import type {
  Visualization,
  VisualizationCreateInput,
  VisualizationListParams,
  VisualizationUpdateInput,
} from '@/features/dashboard/types'
import type { Paged } from '@/shared/lib/api-client'

export const VISUALIZATIONS_QUERY_KEYS = {
  all: ['visualizations'] as const,
  list: (params: VisualizationListParams) =>
    [...VISUALIZATIONS_QUERY_KEYS.all, 'list', params] as const,
  one: (id: string) => [...VISUALIZATIONS_QUERY_KEYS.all, 'one', id] as const,
}

export function useVisualizations(params: VisualizationListParams = {}) {
  const service = useMemo(() => createVisualizationsService(), [])
  return useQuery<Paged<Visualization[]>>({
    queryKey: VISUALIZATIONS_QUERY_KEYS.list(params),
    queryFn: () => service.listVisualizations(params),
    // When scoping by dashboard, don't fire until a dashboard is actually selected.
    enabled: params.dashboardId == null || !!params.dashboardId,
  })
}

export function useVisualization(id: string | null) {
  const service = useMemo(() => createVisualizationsService(), [])
  return useQuery<Visualization>({
    queryKey: VISUALIZATIONS_QUERY_KEYS.one(id ?? ''),
    queryFn: () => service.getVisualization(id as string),
    enabled: !!id,
  })
}

export function useVisualizationMutations() {
  const queryClient = useQueryClient()
  const service = useMemo(() => createVisualizationsService(), [])

  const createVisualization = useMutation<Visualization, Error, VisualizationCreateInput>({
    mutationFn: (data) => service.createVisualization(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: VISUALIZATIONS_QUERY_KEYS.all })
    },
  })

  const updateVisualization = useMutation<Visualization, Error, VisualizationUpdateInput>({
    mutationFn: (data) => service.updateVisualization(data),
    onSuccess: (_, vars) => {
      queryClient.invalidateQueries({ queryKey: VISUALIZATIONS_QUERY_KEYS.one(vars.id) })
      queryClient.invalidateQueries({ queryKey: VISUALIZATIONS_QUERY_KEYS.all })
    },
  })

  const deleteVisualization = useMutation<void, Error, string>({
    mutationFn: (id) => service.deleteVisualization(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: VISUALIZATIONS_QUERY_KEYS.all })
    },
  })

  return { createVisualization, updateVisualization, deleteVisualization }
}
