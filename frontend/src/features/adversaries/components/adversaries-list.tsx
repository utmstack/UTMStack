import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { LIST_COLS, THREAT, flagEmoji, relativeTime } from '../lib/adversary-meta'
import type { Adversary } from '../types/adversary.types'
import { AdversaryAvatar } from './adversary-avatar'

export function AdversariesList({ adversaries, onOpen }: { adversaries: Adversary[]; onOpen: (a: Adversary) => void }) {
  return (
    <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
      <ListHeader />
      {adversaries.map((a) => (
        <ListRow key={a.id} a={a} onOpen={() => onOpen(a)} />
      ))}
    </div>
  )
}

function ListHeader() {
  const { t } = useTranslation()
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <div />
      <div>{t('adversaries.col.adversary')}</div>
      <div>{t('adversaries.col.threat')}</div>
      <div className="text-right">{t('adversaries.col.alerts')}</div>
      <div className="text-right">{t('adversaries.col.targets')}</div>
      <div>{t('adversaries.col.lastSeen')}</div>
    </div>
  )
}

function ListRow({ a, onOpen }: { a: Adversary; onOpen: () => void }) {
  const { t } = useTranslation()
  return (
    <div
      onClick={onOpen}
      className="grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-xs last:border-b-0 hover:bg-muted/40"
      style={{ gridTemplateColumns: LIST_COLS }}
    >
      <AdversaryAvatar a={a} />
      <div className="min-w-0">
        <div className="truncate font-mono text-[13px] font-medium">{a.identifier}</div>
        <div className="truncate text-[11px] text-muted-foreground">
          {a.geo?.country ? `${flagEmoji(a.geo.countryCode)} ${a.geo.country}` : t(`adversaries.kind.${a.kind}`)}
          {a.geo?.aso ? ` · ${a.geo.aso}` : ''}
        </div>
      </div>
      <div>
        <span className={cn('inline-flex items-center gap-1.5 text-[11px]', THREAT[a.threatLevel].text)}>
          <span className={cn('h-1.5 w-1.5 rounded-full', THREAT[a.threatLevel].dot)} />
          {t(`adversaries.threat.${a.threatLevel}`)}
        </span>
      </div>
      <div className="text-right font-mono tabular-nums text-muted-foreground">{a.alertsCount}</div>
      <div className="text-right font-mono tabular-nums text-muted-foreground">{a.targets.length}</div>
      <div className="font-mono text-[11px] text-muted-foreground">{relativeTime(a.lastSeen, t)}</div>
    </div>
  )
}
