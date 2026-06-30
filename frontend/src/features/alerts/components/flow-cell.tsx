import { ArrowRight } from 'lucide-react'
import type { Side } from '../types/alert.types'
import { EndpointMini } from './endpoint-mini'

export function FlowCell({ source, adversary }: { source?: Side; adversary?: Side }) {
  return (
    <div className="flex items-center justify-between gap-1.5 whitespace-nowrap text-[11px]">
      <EndpointMini ep={source}  />
      <ArrowRight size={11} className="shrink-0 text-muted-foreground/60" />
      <EndpointMini ep={adversary} accent />
    </div>
  )
}
