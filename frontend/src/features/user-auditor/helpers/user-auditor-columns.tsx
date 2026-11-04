import type { TFunction } from 'i18next'
import type { ColumnDef } from '@tanstack/react-table'
import { ExternalLink } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import type { ADUserWithIdentity } from '../types/ad-user.types'
import {
  accountIdentity,
  accountName,
  accountScope,
  isService,
  isStale,
  relativeTime,
  STATUS_DOT,
  STATUS_TEXT,
  statusOf,
} from './user-auditor-display'
import { AccountAvatar } from '../components/account-avatar'

export const TH = 'whitespace-nowrap px-3 py-2 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-2.5 align-middle'

export function buildUserColumns(t: TFunction): ColumnDef<ADUserWithIdentity>[] {
  const muted = `${TD} font-mono text-[11px] text-muted-foreground`
  return [
    {
      id: 'avatar',
      header: () => null,
      size: 52,
      minSize: 52,
      enableResizing: false,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => <AccountAvatar user={row.original} />,
    },
    {
      id: 'account',
      header: t('userAuditor.list.account'),
      size: 320,
      minSize: 120,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const user = row.original
        return (
          <>
            <div className="flex items-center gap-2">
              <span className="truncate text-sm font-medium">{accountName(user)}</span>
              <span className="rounded bg-muted px-1.5 py-0.5 text-[9px] font-medium uppercase text-muted-foreground ring-1 ring-border">
                {user.source}
              </span>
              {isService(user) && (
                <span className="rounded bg-violet-500/15 px-1.5 py-0.5 text-[9px] font-medium uppercase text-violet-600 ring-1 ring-violet-500/30 dark:text-violet-300">
                  {t('userAuditor.badge.service')}
                </span>
              )}
              {isStale(user) && (
                <span className="rounded bg-amber-500/15 px-1.5 py-0.5 text-[9px] font-medium uppercase text-amber-600 ring-1 ring-amber-500/30 dark:text-amber-300">
                  {t('userAuditor.badge.stale')}
                </span>
              )}
            </div>
            <div className="truncate font-mono text-[11px] text-muted-foreground">{accountScope(user)}</div>
          </>
        )
      },
    },
    {
      id: 'identity',
      header: t('userAuditor.list.accountId'),
      size: 260,
      minSize: 80,
      meta: {
        headerClassName: TH,
        cellClassName: muted,
        cellProps: (u) => ({ title: accountIdentity(u) }),
      },
      cell: ({ row }) => <span className="block truncate">{accountIdentity(row.original)}</span>,
    },
    {
      id: 'identityGroup',
      header: t('userAuditor.list.identityGroup'),
      size: 180,
      minSize: 100,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const group = (row.original as ADUserWithIdentity).identity
        if (!group) return <span className="text-[11px] text-muted-foreground">—</span>
        return (
          <span className="rounded-md bg-sky-500/15 px-1.5 py-0.5 text-[11px] font-medium text-sky-700 ring-1 ring-sky-500/30 dark:text-sky-300">
            {group}
          </span>
        )
      },
    },
    {
      id: 'status',
      header: t('userAuditor.list.status'),
      size: 120,
      minSize: 90,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const status = statusOf(row.original)
        return (
          <span className={cn('inline-flex items-center gap-1.5 text-[11px]', STATUS_TEXT[status])}>
            <span className={cn('h-1.5 w-1.5 rounded-full', STATUS_DOT[status])} />
            {t(`userAuditor.status.${status}`)}
          </span>
        )
      },
    },
    {
      id: 'lastLogon',
      header: t('userAuditor.list.lastLogon'),
      size: 120,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: muted },
      cell: ({ row }) => relativeTime(row.original.lastLogon, t),
    },
    {
      id: 'lastSeen',
      header: t('userAuditor.list.lastSeen'),
      size: 120,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: muted },
      cell: ({ row }) => relativeTime(row.original.lastSeen, t),
    },
    {
      id: 'tenant',
      header: t('userAuditor.list.tenant'),
      size: 160,
      minSize: 80,
      meta: {
        headerClassName: TH,
        cellClassName: `${TD} text-[11px] text-muted-foreground`,
        cellProps: (u) => ({ title: u.tenantId }),
      },
      cell: ({ row }) => <span className="block truncate">{row.original.tenantId}</span>,
    },
    {
      id: 'open',
      header: () => null,
      size: 44,
      minSize: 44,
      enableResizing: false,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: () => (
        <div className="flex justify-end opacity-0 group-hover:opacity-100">
          <ExternalLink size={13} className="text-muted-foreground" />
        </div>
      ),
    },
  ]
}
