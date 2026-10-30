import { Link } from 'react-router-dom'
import type { TFunction } from 'i18next'
import { Copy, ExternalLink, Server, UserCircle2, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import type { ADUserWithIdentity } from '../types/ad-user.types'
import {
  accountIdentity,
  accountName,
  accountScope,
  isCloudSource,
  isService,
  isStale,
  logExplorerHref,
  STATUS_DOT,
  STATUS_TEXT,
  statusOf,
} from '../helpers/user-auditor-display'
import { AccountAvatar } from './account-avatar'
import { IdentityGroupPicker } from './identity-group-picker'
import { Row, Section } from './section-row'
import { Timeline } from './timeline'

export function UserDrawer({
  user,
  canManage,
  onClose,
  onUserChanged,
  t,
}: {
  user: ADUserWithIdentity
  canManage: boolean
  onClose: () => void
  onUserChanged: (u: ADUserWithIdentity) => void
  t: TFunction
}) {
  const status = statusOf(user)
  const identity = accountIdentity(user)
  const isCloud = isCloudSource(user.source)
  const isLinux = user.source === 'linux'
  const logsHref = logExplorerHref(user)

  const copyId = () => {
    void navigator.clipboard?.writeText(identity)
    toast.success(t('userAuditor.drawer.copied'))
  }

  const onAliasChanged = (name: string | null) => {
    onUserChanged({ ...user, identity: name })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[680px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <AccountAvatar user={user} large />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <Server size={11} />
                  <span className="font-mono">{accountScope(user)}</span>
                  <span>·</span>
                  <span>{t('userAuditor.drawer.tenantInline', { id: user.tenantId })}</span>
                </div>
                <h2 className="mt-0.5 truncate text-xl font-semibold">{accountName(user)}</h2>
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                  <span className="rounded-md bg-muted px-1.5 py-0.5 font-medium uppercase text-muted-foreground ring-1 ring-border">
                    {user.source}
                  </span>
                  <span className={cn('inline-flex items-center gap-1.5', STATUS_TEXT[status])}>
                    <span className={cn('h-1.5 w-1.5 rounded-full', STATUS_DOT[status])} />
                    {t(`userAuditor.status.${status}`)}
                  </span>
                  {isService(user) && (
                    <span className="rounded-md bg-violet-500/15 px-1.5 py-0.5 font-medium text-violet-600 ring-1 ring-violet-500/30 dark:text-violet-300">
                      {t('userAuditor.drawer.serviceAccount')}
                    </span>
                  )}
                  {isStale(user) && (
                    <span className="rounded-md bg-amber-500/15 px-1.5 py-0.5 font-medium text-amber-600 ring-1 ring-amber-500/30 dark:text-amber-300">
                      {t('userAuditor.drawer.stale')}
                    </span>
                  )}
                </div>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          <div className="mt-4 flex flex-wrap items-center gap-2">
            {logsHref && (
              <Button size="sm" variant="outline" asChild>
                <Link to={logsHref}>
                  <ExternalLink size={14} className="mr-1.5" />
                  {t('userAuditor.drawer.viewInLogs')}
                </Link>
              </Button>
            )}
            <Button size="sm" variant="outline" onClick={copyId}>
              <Copy size={14} className="mr-1.5" />
              {isLinux || isCloud ? t('userAuditor.drawer.copyUid') : t('userAuditor.drawer.copySid')}
            </Button>
            {canManage}
            {canManage && (
              <IdentityGroupPicker userId={user.id} current={user.identity ?? null} onChanged={onAliasChanged} />
            )}
          </div>
        </header>

        <div className="flex-1 space-y-4 overflow-y-auto bg-muted/20 p-6">
          <Section title={t('userAuditor.drawer.identity')}>
            <dl className="grid grid-cols-[150px_1fr] gap-y-2 text-xs">
              <Row k={t('userAuditor.drawer.accountName')}>
                <span className="font-mono">{accountName(user)}</span>
              </Row>
              {isLinux && (
                <>
                  <Row k={t('userAuditor.drawer.hostname')}>
                    <span className="font-mono">{accountScope(user)}</span>
                  </Row>
                  <Row k={t('userAuditor.drawer.uidNumber')}>
                    <span className="font-mono">{user.uidNumber ?? '—'}</span>
                  </Row>
                  <Row k={t('userAuditor.drawer.machineId')}>
                    <span className="break-all font-mono">{user.machineId ?? '—'}</span>
                  </Row>
                </>
              )}
              {user.source === 'windows' && (
                <>
                  <Row k={t('userAuditor.drawer.domain')}>
                    <span className="font-mono">{accountScope(user)}</span>
                  </Row>
                  <Row k={t('userAuditor.drawer.sid')}>
                    <span className="break-all font-mono">{user.sid ?? '—'}</span>
                  </Row>
                </>
              )}
              <Row k={t('userAuditor.drawer.tenant')}>
                <span className="font-mono">{user.tenantId}</span>
              </Row>
              <Row k={t('userAuditor.drawer.adStatus')}>
                {user.active ? t('userAuditor.drawer.enabled') : t('userAuditor.drawer.disabled')}
              </Row>
              <Row k={t('userAuditor.drawer.aliasGroup')}>
                {user.identity ? (
                  <span className="inline-flex items-center gap-1.5 text-xs">
                    <UserCircle2 size={13} className="text-sky-600 dark:text-sky-400" />
                    {user.identity}
                  </span>
                ) : (
                  <span className="text-muted-foreground">{t('userAuditor.drawer.alias.unassigned')}</span>
                )}
              </Row>
            </dl>
          </Section>

          <Section title={t('userAuditor.drawer.lifecycle')}>
            <Timeline user={user} t={t} />
          </Section>
        </div>
      </div>
    </div>
  )
}
