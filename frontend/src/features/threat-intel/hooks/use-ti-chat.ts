import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { threatIntelHttpService } from '../services/threat-intel-http.service'
import { describeError, isNotConfigured } from '../services/ti-errors'
import type { ChatRequest } from '../domain/threat-intel.types'

export function useTiChat() {
  return useMutation({
    mutationFn: (req: ChatRequest) => threatIntelHttpService.chat(req),
    onError: (e) => { if (!isNotConfigured(e)) toast.error(describeError(e)) },
  })
}
