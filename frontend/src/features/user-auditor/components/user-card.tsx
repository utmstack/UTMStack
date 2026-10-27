import type { TFunction } from 'i18next'
import { cn } from '@/shared/lib/utils'
import type { ADUser } from '../types/ad-user.types'
import {
  accountIdentity,
  accountName,
  accountScope,
  relativeTime,
  STATUS_DOT,
  STATUS_TEXT,
  statusOf,
} from '../helpers/user-auditor-display'
import { AccountAvatar } from './account-avatar'

export function UserCard({ user, onOpen, t }: { user: ADUser; onOpen: () => void; t: TFunction }) {
  const status = statusOf(user)
  const identity = accountIdentity(user)
  return (
    <button
      onClick={onOpen}
      className="group flex flex-col gap-3 rounded-xl border border-border bg-card p-4 text-left transition-all hover:shadow-md"
    >
      <div className="flex items-start gap-3">
        <AccountAvatar user={user} large />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="truncate text-sm font-semibold">{accountName(user)}</span>
            <span className="rounded bg-muted px-1.5 py-0.5 text-[9px] font-medium uppercase text-muted-foreground ring-1 ring-border">
              {user.source}
            </span>
          </div>
          <div className="truncate font-mono text-[11px] text-muted-foreground">{accountScope(user)}</div>
          <div className="mt-1">
            <span className={cn('inline-flex items-center gap-1.5 text-[11px]', STATUS_TEXT[status])}>
              <span className={cn('h-1.5 w-1.5 rounded-full', STATUS_DOT[status])} />
              {t(`userAuditor.status.${status}`)}
            </span>
          </div>
        </div>
      </div>
      <div className="truncate border-t border-border pt-2 font-mono text-[10px] text-muted-foreground" title={identity}>
        {identity}
      </div>
      <div className="flex items-center justify-between text-[11px] text-muted-foreground">
        <span>
          {t('userAuditor.card.lastLogon')}{' '}
          <span className="font-mono text-foreground">{relativeTime(user.lastLogon, t)}</span>
        </span>
        <span className="font-mono">{user.tenantId}</span>
      </div>
    </button>
  )
}
