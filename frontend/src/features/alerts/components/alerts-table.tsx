import { ArrowRight, Sparkles, Tag, UserCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { SEV_META, ST_META, TABLE_COLS, TS, absTime, flagEmoji, relativeTime, riskOf, sevKey, statusKey } from '../lib/alert-meta'
import { isAiNote } from '../lib/ai-note'
import type { Alert, AlertTag, Side } from '../types/alert.types'
import { TagChip } from './ui-primitives'

export function AlertsTableHeader({ allChecked, onTogglePage }: { allChecked: boolean; onTogglePage: () => void }) {
  const { t } = useTranslation()
  return (
    <div
      className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
      style={{ gridTemplateColumns: TABLE_COLS }}
    >
      <button onClick={onTogglePage} className="flex h-4 w-4 items-center justify-center rounded border border-input">
        {allChecked && <span className="h-2 w-2 rounded-sm bg-primary" />}
      </button>
      <div />
      <div>{t('alerts.table.alert')}</div>
      <div>{t('alerts.table.status')}</div>
      <div>{t('alerts.table.technique')}</div>
      <div>{t('alerts.table.sourceAdversary')}</div>
      <div className="text-center" title={t('alerts.table.riskHint')}>{t('alerts.table.risk')}</div>
      <div className="text-center">{t('alerts.table.time')}</div>
      <div />
    </div>
  )
}

export function AlertRow({
  alert: a,
  tagCatalog,
  checked,
  onToggle,
  onOpen,
  onCreateRule,
}: {
  alert: Alert
  tagCatalog: AlertTag[]
  checked: boolean
  onToggle: () => void
  onOpen: () => void
  onCreateRule: (alert: Alert) => void
}) {
  const { t } = useTranslation()
  const stm = ST_META[statusKey(a)]
  const stmLabel = t(`alerts.status.${statusKey(a)}`)
  const sk = sevKey(a)
  const sev = SEV_META[sk]
  return (
    <div
      className="group relative grid cursor-pointer items-center gap-3 border-b border-border/50 px-4 py-2.5 text-[13px] last:border-b-0 hover:bg-muted/20"
      style={{ gridTemplateColumns: TABLE_COLS }}
      onClick={onOpen}
    >
      {/* Severity accent — colored left edge so the row's risk reads at a glance. */}
      <span
        className={cn('absolute inset-y-0 left-0 w-[3px]', sev.bar)}
        title={t(`alerts.severity.${sk}`)}
        aria-hidden
      />
      <button
        onClick={(e) => {
          e.stopPropagation()
          onToggle()
        }}
        className={cn(
          'flex h-4 w-4 items-center justify-center rounded border',
          checked ? 'border-primary bg-primary' : 'border-input'
        )}
      >
        {checked && <span className="h-2 w-2 rounded-sm bg-primary-foreground" />}
      </button>
      <button
        onClick={(e) => {
          e.stopPropagation()
          onCreateRule(a)
        }}
        title={t('alerts.row.createRuleFromAlert')}
        aria-label={t('alerts.row.createRuleFromAlert')}
        className="flex h-7 w-7 items-center justify-center rounded text-muted-foreground/60  transition group-hover:opacity-100 hover:bg-background hover:text-primary mr-4"
      >
        <Tag size={13} />
      </button>
      <div className="min-w-0">
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
      </div>
      <div>
        <span className={cn('inline-flex items-center rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', stm.pill)}>
          {stmLabel}
        </span>
      </div>
      <div className="truncate font-mono text-[11px] text-muted-foreground" title={a.technique}>
        {a.technique || '—'}
      </div>
      <FlowCell source={a.target} adversary={a.adversary} />
      <div className="flex justify-center">
        <span
          className={cn(
            'inline-flex min-w-[26px] items-center justify-center rounded-md px-1.5 py-0.5 font-mono text-[11px] font-semibold tabular-nums ring-1 ring-inset',
            sev.pill,
          )}
          title={t('alerts.table.riskHint')}
        >
          {riskOf(a)}
        </span>
      </div>
      <div className="text-center font-mono text-[11px] text-muted-foreground" title={absTime(a[TS])}>{relativeTime(a[TS])}</div>
    </div>
  )
}

function FlowCell({ source, adversary }: { source?: Side; adversary?: Side }) {
  return (
    <div className="flex min-w-0 items-center gap-1.5 text-[11px]">
      <EndpointMini ep={source} />
      <ArrowRight size={11} className="shrink-0 text-muted-foreground/60" />
      <EndpointMini ep={adversary} accent />
    </div>
  )
}

function EndpointMini({ ep, accent }: { ep?: Side; accent?: boolean }) {
  if (!ep || (!ep.host && !ep.ip && !ep.user)) return <span className="text-muted-foreground/50">—</span>
  const cc = ep.geolocation?.countryCode
  const flag = flagEmoji(cc)
  return (
    <span className={cn('inline-flex min-w-0 items-center gap-1', accent ? 'text-foreground/90' : 'text-muted-foreground')}>
      {flag && <span title={ep.geolocation?.country || cc}>{flag}</span>}
      <span className="min-w-0 truncate font-mono">{ep.host || ep.user || ep.ip}</span>
    </span>
  )
}
