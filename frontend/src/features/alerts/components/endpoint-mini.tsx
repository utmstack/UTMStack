import { cn } from '@/shared/lib/utils'
import { flagEmoji } from '../lib/alert-meta'
import type { Side } from '../types/alert.types'

export function EndpointMini({ ep, accent }: { ep?: Side; accent?: boolean }) {
  if (!ep || (!ep.host && !ep.ip && !ep.user)) return <span className="text-muted-foreground/50">—</span>
  const cc = ep.geolocation?.countryCode
  const flag = flagEmoji(cc)
  return (
    <span className={cn('inline-flex  items-center  w-auto gap-1', accent ? 'text-foreground/90' : 'text-muted-foreground')}>
      {flag && <span title={ep.geolocation?.country || cc}>{flag}</span>}
      <span className="min-w-0 whitespace-nowrap font-mono">{ep.host || ep.user || ep.ip}</span>
    </span>
  )
}
