import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { threatIntelHttpService } from '../services/threat-intel-http.service'
import { describeError, isNotConfigured, isNotFound } from '../services/ti-errors'
import type { EntityLookupRequest } from '../domain/threat-intel.types'

export function useTiEntityLookup() {
  return useMutation({
    mutationFn: (req: EntityLookupRequest) => threatIntelHttpService.entityLookup(req),
    onError: (e) => { if (!isNotConfigured(e) && !isNotFound(e)) toast.error(describeError(e)) },
  })
}
