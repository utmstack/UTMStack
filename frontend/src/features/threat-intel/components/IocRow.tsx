import { useTranslation } from 'react-i18next'
import { MoreHorizontal } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { searchItemValue, type EntitySummary } from '../domain/threat-intel.types'
import { REPUTATION_STYLE, reputationLabelKey, reputationTone, typeMeta } from './utils/severity-style'
import { relativeTime } from './utils/time-format'

interface IocRowProps {
  ioc: EntitySummary
  onOpen: (id: string) => void
}

const IOC_COLS = '4px 90px 1fr 130px 1fr 110px 36px'

export function IocRow({ ioc, onOpen }: IocRowProps) {
  const { t } = useTranslation()
  const tone = reputationTone(ioc.reputation)
  const rep = REPUTATION_STYLE[tone]
  const typeMeta_ = typeMeta(ioc.type)
  const TIcon = typeMeta_.icon

  return (
    <div
      onClick={() => onOpen(ioc.id)}
      className="group grid cursor-pointer items-center gap-3 border-b border-border/60 px-4 py-2.5 text-xs hover:bg-muted/40 last:border-b-0"
      style={{ gridTemplateColumns: IOC_COLS }}
    >
      <span className={cn('h-3 w-1 rounded-full', rep.bar)} />
      <div className="flex items-center gap-1.5">
        <TIcon size={11} className={typeMeta_.tone} />
        <span className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
          {t(typeMeta_.labelKey)}
        </span>
      </div>
      <div className="min-w-0 truncate font-mono">{searchItemValue(ioc)}</div>
      <div className="text-right">
        <span className={cn('text-[11px] font-medium', rep.tone)}>
          {t(reputationLabelKey(ioc.reputation))}
        </span>
      </div>
      <div className="flex flex-wrap gap-1">
        {ioc.tags?.slice(0, 3).map((tag) => (
          <span key={tag} className="rounded bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground">
            {tag}
          </span>
        ))}
        {ioc.tags && ioc.tags.length > 3 && (
          <span className="text-[10px] text-muted-foreground">+{ioc.tags.length - 3}</span>
        )}
      </div>
      <div className="font-mono text-[11px] text-muted-foreground">
        {ioc.lastSeen ? relativeTime(ioc.lastSeen) : '—'}
      </div>
      <div className="flex justify-end opacity-0 group-hover:opacity-100">
        <button
          onClick={(e) => e.stopPropagation()}
          className="flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-background hover:text-foreground"
        >
          <MoreHorizontal size={14} />
        </button>
      </div>
    </div>
  )
}
