import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { threatIntelHttpService } from '../services/threat-intel-http.service'
import { describeError, isNotConfigured, isNotFound } from '../services/ti-errors'
import type { AdvancedSearchRequest } from '../domain/threat-intel.types'

export function useTiSearchAdvanced() {
  return useMutation({
    mutationFn: async (input: { body: AdvancedSearchRequest; limit?: number; page?: number }) =>{
      try{
        return await threatIntelHttpService.searchAdvanced(input.body, { limit: input.limit, page: input.page })
      }catch(e){
        if(isNotFound(e)){
          const resp = await threatIntelHttpService.searchAdvanced({}, { limit: input.limit, page: input.page })
          resp.kind="empty"
          return resp
        }
        throw e
      }
    },
    onError: (e) => { if (!isNotConfigured(e) && !isNotFound(e)) toast.error(describeError(e)) },
  })
}
