import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { threatIntelHttpService } from '../services/threat-intel-http.service'
import { describeError, isNotConfigured, isNotFound } from '../services/ti-errors'
import type { AdvancedSearchRequest } from '../domain/threat-intel.types'

export function useTiSearchAdvanced() {
  return useMutation({
    mutationFn: (input: { body: AdvancedSearchRequest; limit?: number; page?: number }) =>
      threatIntelHttpService.searchAdvanced(input.body, { limit: input.limit, page: input.page }),
    onError: (e) => { if (!isNotConfigured(e) && !isNotFound(e)) toast.error(describeError(e)) },
  })
}
