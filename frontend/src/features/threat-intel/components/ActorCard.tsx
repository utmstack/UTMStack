import { useTranslation } from 'react-i18next'
import { UserX } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { searchItemValue, type EntitySummary } from '../domain/threat-intel.types'
import { REPUTATION_STYLE, reputationLabelKey, reputationTone } from './utils/severity-style'
import { Stat } from './Stat'
import { relativeTime } from './utils/time-format'

interface ActorCardProps {
  actor: EntitySummary
  onOpen: (id: string) => void
}

export function ActorCard({ actor, onOpen }: ActorCardProps) {
  const { t } = useTranslation()
  const tone = reputationTone(actor.reputation)
  const rep = REPUTATION_STYLE[tone]
  const dangerous = tone === 'danger'

  return (
    <button
      onClick={() => onOpen(actor.id)}
      className={cn(
        'group flex flex-col gap-3 rounded-xl border bg-card p-4 text-left transition-all hover:shadow-md',
        dangerous ? 'border-red-500/30' : 'border-border'
      )}
    >
      <div className="flex items-start gap-3">
        <div
          className={cn(
            'flex h-10 w-10 shrink-0 items-center justify-center rounded-lg ring-2',
            dangerous
              ? 'bg-red-500/15 text-red-500 ring-red-500/40'
              : 'bg-muted text-foreground/80 ring-border'
          )}
        >
          <UserX size={18} />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="truncate text-sm font-semibold">{searchItemValue(actor)}</span>
            <span className={cn('rounded px-1.5 py-0.5 text-[9px] font-medium uppercase ring-1', rep.tone)}>
              {t(reputationLabelKey(actor.reputation))}
            </span>
          </div>
          {actor.tags.length > 0 && (
            <div className="mt-0.5 truncate text-[11px] text-muted-foreground">
              {actor.tags.join(', ')}
            </div>
          )}
        </div>
      </div>

      <div className="grid grid-cols-3 gap-2 text-[11px]">
        <Stat label="Reputation" value={actor.reputation.toString()} />
        <Stat label="Accuracy" value={actor.accuracy.toString()} />
        <Stat label="Tags" value={actor.tags.length.toString()} />
      </div>

      <div className="border-t border-border pt-3 text-[10px] text-muted-foreground">
        Last seen <span className="font-mono">{actor.lastSeen ? relativeTime(actor.lastSeen) : '—'}</span>
      </div>
    </button>
  )
}
