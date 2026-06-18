import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { THREAT, flagEmoji, relativeTime } from '../lib/adversary-meta'
import type { Adversary } from '../types/adversary.types'
import { AdversaryAvatar } from './adversary-avatar'

export function AdversariesGrid({ adversaries, onOpen }: { adversaries: Adversary[]; onOpen: (a: Adversary) => void }) {
  return (
    <div className="mt-3 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
      {adversaries.map((a) => (
        <GridCard key={a.id} a={a} onOpen={() => onOpen(a)} />
      ))}
    </div>
  )
}

function GridCard({ a, onOpen }: { a: Adversary; onOpen: () => void }) {
  const { t } = useTranslation()
  return (
    <button
      onClick={onOpen}
      className="flex flex-col gap-3 rounded-xl border border-border bg-card p-4 text-left transition-all hover:shadow-md"
    >
      <div className="flex items-start gap-3">
        <AdversaryAvatar a={a} large />
        <div className="min-w-0 flex-1">
          <div className="truncate font-mono text-sm font-semibold">{a.identifier}</div>
          <div className="truncate text-[11px] text-muted-foreground">
            {a.geo?.country ? `${flagEmoji(a.geo.countryCode)} ${a.geo.country}` : t(`adversaries.kind.${a.kind}`)}
          </div>
          <div className="mt-1">
            <span className={cn('inline-flex items-center gap-1.5 text-[11px]', THREAT[a.threatLevel].text)}>
              <span className={cn('h-1.5 w-1.5 rounded-full', THREAT[a.threatLevel].dot)} />
              {t(`adversaries.threat.${a.threatLevel}`)}
            </span>
          </div>
        </div>
      </div>
      <div className="grid grid-cols-3 gap-2 border-t border-border pt-3 text-[11px]">
        <Stat label={t('adversaries.card.alerts')} value={a.alertsCount} />
        <Stat label={t('adversaries.card.techniques')} value={a.techniques.length} />
        <Stat label={t('adversaries.card.targets')} value={a.targets.length} />
      </div>
      <div className="text-[11px] text-muted-foreground">
        {t('adversaries.col.lastSeen')} <span className="font-mono text-foreground">{relativeTime(a.lastSeen, t)}</span>
      </div>
    </button>
  )
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-md border border-border bg-background/40 px-2 py-1.5">
      <div className="text-[9px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className="mt-0.5 font-semibold tabular-nums">{value}</div>
    </div>
  )
}
