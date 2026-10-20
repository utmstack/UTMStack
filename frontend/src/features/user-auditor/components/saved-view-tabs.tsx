import type { TFunction } from 'i18next'
import { cn } from '@/shared/lib/utils'
import { VIEW_IDS, type ViewId } from '../helpers/user-auditor-display'

export function SavedViewTabs({
  current,
  counts,
  onChange,
  t,
}: {
  current: ViewId
  counts: Record<ViewId, number>
  onChange: (id: ViewId) => void
  t: TFunction
}) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {VIEW_IDS.map((id) => {
        const active = current === id
        return (
          <button
            key={id}
            onClick={() => onChange(id)}
            className={cn(
              'group relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground',
            )}
          >
            <span>{t(`userAuditor.views.${id}`)}</span>
            <span
              className={cn(
                'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground',
              )}
            >
              {counts[id] ?? 0}
            </span>
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}
