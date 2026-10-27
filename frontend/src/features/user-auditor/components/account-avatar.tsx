import { cn } from '@/shared/lib/utils'
import { UserCog, UserX } from 'lucide-react'
import type { ADUser } from '../types/ad-user.types'
import { isService, STATUS_DOT, statusOf } from '../helpers/user-auditor-display'

export function AccountAvatar({ user, large }: { user: ADUser; large?: boolean }) {
  const status = statusOf(user)
  const Icon = status === 'deleted' ? UserX : UserCog
  return (
    <div className="relative">
      <div
        className={cn(
          'flex shrink-0 items-center justify-center rounded-lg bg-muted text-foreground/70 ring-2 ring-border',
          large ? 'h-11 w-11' : 'h-7 w-7',
        )}
      >
        <Icon size={large ? 18 : 13} className={isService(user) ? 'text-violet-500' : 'text-sky-500'} />
      </div>
      <span className={cn('absolute -right-0.5 -top-0.5 h-2.5 w-2.5 rounded-full ring-2 ring-card', STATUS_DOT[status])} />
    </div>
  )
}
