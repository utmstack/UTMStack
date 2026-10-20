import type { TFunction } from 'i18next'
import type { LucideIcon } from 'lucide-react'
import { Clock, Fingerprint, UserPlus, UserX } from 'lucide-react'
import type { ADUser } from '../types/ad-user.types'
import { absTimestamp, relativeTime } from '../helpers/user-auditor-display'

export function Timeline({ user, t }: { user: ADUser; t: TFunction }) {
  const events: { icon: LucideIcon; label: string; ts?: string; tone: string }[] = [
    { icon: UserPlus, label: t('userAuditor.drawer.created'), ts: user.accountCreatedAt, tone: 'text-sky-500' },
    { icon: Fingerprint, label: t('userAuditor.drawer.lastLogon'), ts: user.lastLogon, tone: 'text-emerald-500' },
    { icon: Clock, label: t('userAuditor.drawer.lastSeen'), ts: user.lastSeen, tone: 'text-muted-foreground' },
    { icon: UserX, label: t('userAuditor.drawer.deleted'), ts: user.accountDeletedAt, tone: 'text-red-500' },
  ].filter((e) => e.ts)

  if (events.length === 0) {
    return <div className="text-xs text-muted-foreground">{t('userAuditor.drawer.noLifecycle')}</div>
  }

  return (
    <ol className="relative space-y-4 pl-6">
      <span className="absolute bottom-1 left-[9px] top-1 w-px bg-border" />
      {events.map((e) => (
        <li key={e.label} className="relative">
          <span className="absolute -left-6 flex h-[18px] w-[18px] items-center justify-center rounded-full bg-card ring-1 ring-border">
            <e.icon size={11} className={e.tone} />
          </span>
          <div className="flex items-baseline justify-between gap-3">
            <span className="text-xs font-medium">{e.label}</span>
            <span className="font-mono text-[11px] text-muted-foreground">{relativeTime(e.ts, t)}</span>
          </div>
          <div className="font-mono text-[11px] text-muted-foreground">{absTimestamp(e.ts!)}</div>
        </li>
      ))}
    </ol>
  )
}
