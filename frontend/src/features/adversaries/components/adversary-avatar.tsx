import { cn } from '@/shared/lib/utils'
import { KIND_ICON, THREAT } from '../lib/adversary-meta'
import type { Adversary } from '../types/adversary.types'

export function AdversaryAvatar({ a, large }: { a: Adversary; large?: boolean }) {
  const Icon = KIND_ICON[a.kind]
  return (
    <div
      className={cn(
        'flex shrink-0 items-center justify-center rounded-lg bg-muted ring-2',
        large ? 'h-11 w-11' : 'h-7 w-7',
        THREAT[a.threatLevel].ring
      )}
    >
      <Icon size={large ? 18 : 13} className={THREAT[a.threatLevel].text} />
    </div>
  )
}
