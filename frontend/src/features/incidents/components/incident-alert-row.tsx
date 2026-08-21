import { Sparkles, UserCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { SEV_BADGE, SEV_META, TS, absTime, relativeTime, sevKey } from '../../alerts/lib/alert-meta'
import { isAiNote } from '../../alerts/lib/ai-note'
import type { Alert, AlertTag } from '../../alerts/types/alert.types'
import { TagChip } from '../../alerts/components/tag-chip'

const TD = 'whitespace-nowrap px-3 py-2.5 align-middle'

export function IncidentAlertRow({
  alert: a,
  tagCatalog,
  checked,
  onToggle,
}: {
  alert: Alert
  tagCatalog: AlertTag[]
  checked: boolean
  onToggle: () => void
}) {
  const { t } = useTranslation()
  const sk = sevKey(a)
  const sev = SEV_META[sk]
  return (
    <tr
      className="group cursor-pointer border-b border-border/50 text-[13px] last:border-b-0 hover:bg-muted/20"
      onClick={onToggle}
    >
      <td className={`${TD} relative w-[6px] p-0`}>
        <span className={cn('absolute inset-y-0 left-0 w-[3px]', sev.bar)} title={t(`alerts.severity.${sk}`)} aria-hidden />
      </td>
      <td className={cn(TD, 'cursor-pointer relative')}>
        <span
          className={cn(
            'flex h-4 w-4 items-center justify-center rounded border',
            checked ? 'border-primary bg-primary' : 'border-input'
          )}
        >
          {checked && <span className="h-2 w-2 rounded-sm bg-primary-foreground" />}
        </span>
      </td>
      <td className={`${TD} max-w-[480px]`}>
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
      </td>
      <td className={`${TD} text-center`}>
        <span
          className={cn(
            'inline-flex items-center justify-center rounded-md px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide ring-1 ring-inset',
            SEV_BADGE[sk],
          )}
        >
          {t(`alerts.severity.${sk}`)}
        </span>
      </td>
      <td className={`${TD} text-center font-mono text-[11px] text-muted-foreground`} title={absTime(a[TS])}>
        {relativeTime(a[TS])}
      </td>
    </tr>
  )
}
