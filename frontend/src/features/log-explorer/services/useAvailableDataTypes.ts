import { useQuery } from '@tanstack/react-query'
import { DEFAULT_DATASET, logExplorerHttpService } from './log-explorer-http.service'

const key = ['log-analyzer', 'data-types', DEFAULT_DATASET] as const

export function useAvailableDataTypes() {
  const query = useQuery<string[]>({
    queryKey: key,
    queryFn: () => logExplorerHttpService.dataTypes(DEFAULT_DATASET),
    staleTime: 5 * 60 * 1000,
  })
  return {
    dataTypes: query.data ?? [],
    isReady: query.data != null,
  }
}
