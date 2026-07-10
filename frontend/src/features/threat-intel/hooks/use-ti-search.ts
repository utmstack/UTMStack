import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { threatIntelHttpService } from '../services/threat-intel-http.service'
import { describeError, isNotConfigured } from '../services/ti-errors'
import type { EntitySearchRequest } from '../domain/threat-intel.types'

export function useTiSearch() {
  return useMutation({
    mutationFn: (req: EntitySearchRequest) => threatIntelHttpService.search(req),
    onError: (e) => { if (!isNotConfigured(e)) toast.error(describeError(e)) },
  })
}
