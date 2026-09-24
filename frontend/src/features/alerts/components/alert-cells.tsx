import { Sparkles, UserCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { SEV_BADGE, SEV_META, sevKey } from '../lib/alert-meta'
import { isAiNote } from '../lib/ai-note'
import type { Alert, AlertTag } from '../types/alert.types'
import { TagChip } from './tag-chip'

export const SEVERITY_CHIP =
  'inline-flex items-center justify-center rounded-md px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide ring-1 ring-inset'

/** Colored left edge; the host cell must be `relative`. */
export function SeverityBar({ alert }: { alert: Alert }) {
  const { t } = useTranslation()
  const sk = sevKey(alert)
  return (
    <span
      className={cn('absolute inset-y-0 left-0 w-[3px]', SEV_META[sk].bar)}
      title={t(`alerts.severity.${sk}`)}
      aria-hidden
    />
  )
}

export function SeverityBadge({ alert }: { alert: Alert }) {
  const { t } = useTranslation()
  const sk = sevKey(alert)
  return (
    <span className={cn(SEVERITY_CHIP, SEV_BADGE[sk])}>
      {t(`alerts.severity.${sk}`)}
    </span>
  )
}

export function AlertSummaryCell({ alert: a, tagCatalog }: { alert: Alert; tagCatalog: AlertTag[] }) {
  const { t } = useTranslation()
  return (
    <>
      <div className="flex items-center gap-2">
        <span className="truncate font-medium">{a.name || '—'}</span>
        {(isAiNote(a.notes) || isAiNote(a.statusObservation)) && (
          <Sparkles size={11} className="shrink-0 text-fuchsia-500" aria-label={t('alerts.badge.aiAssessed')} />
        )}
        {a.isIncident && (
          <span className="shrink-0 rounded bg-red-500/15 px-1 py-0.5 text-[9px] font-semibold uppercase text-red-500">
            {t('alerts.badge.incident')}
          </span>
        )}
      </div>
      <div className="mt-0.5 flex items-center gap-1.5 overflow-hidden text-[11px] text-muted-foreground">
        <span className="truncate">
          {a.category}
          {a.dataSource && ` · ${a.dataSource}`}
        </span>
        {(a.tags ?? []).slice(0, 2).map((tag) => (
          <TagChip key={tag} name={tag} catalog={tagCatalog} size="xs" />
        ))}
        {(a.tags ?? []).length > 2 && (
          <span
            className="shrink-0 whitespace-nowrap rounded-md border border-border bg-muted px-1.5 py-0.5 text-[10px] font-medium leading-none text-muted-foreground"
            title={(a.tags ?? []).slice(2).join(', ')}
          >
            +{(a.tags ?? []).length - 2}
          </span>
        )}
        {a.assignee && (
          <span
            className="inline-flex shrink-0 items-center gap-1 whitespace-nowrap rounded-md bg-primary/10 px-1.5 py-0.5 text-[10px] font-medium leading-none text-primary"
            title={t('alerts.row.assignedTo', { user: a.assignee })}
          >
            <UserCheck size={10} />
            {a.assignee}
          </span>
        )}
      </div>
    </>
  )
}
