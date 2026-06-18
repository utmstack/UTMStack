import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { federationInstancesService } from '../services/federation-instances.service'
import type { FederationInstanceInput } from '../types'

const KEY = ['fs', 'instances'] as const

export function useInstances() {
  return useQuery({
    queryKey: KEY,
    queryFn: () => federationInstancesService.list(),
    staleTime: 60_000,
  })
}

export function useInstanceMutations() {
  const qc = useQueryClient()
  const invalidate = () => qc.invalidateQueries({ queryKey: KEY })

  const create = useMutation({
    mutationFn: (input: FederationInstanceInput) => federationInstancesService.create(input),
    onSuccess: invalidate,
  })
  const update = useMutation({
    mutationFn: ({ id, input }: { id: number; input: FederationInstanceInput }) =>
      federationInstancesService.update(id, input),
    onSuccess: invalidate,
  })
  const remove = useMutation({
    mutationFn: (id: number) => federationInstancesService.remove(id),
    onSuccess: invalidate,
  })

  return { create, update, remove }
}
