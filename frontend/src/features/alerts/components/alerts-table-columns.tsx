import { Tag } from 'lucide-react'
import type { TFunction } from 'i18next'
import type { ColumnDef } from '@tanstack/react-table'
import { cn } from '@/shared/lib/utils'
import { TS, absTime, relativeTime, statusKey } from '../lib/alert-meta'
import type { Alert, AlertTag } from '../types/alert.types'
import { AlertSummaryCell, SeverityBadge, SeverityBar } from './alert-cells'
import { AlertIncidentTarget } from './alert-incident-target'
import { EchoesChip } from './echoes-chip'
import { EndpointMini } from './endpoint-mini'
import { StatusChangeMenu } from './status-change-menu'

const TH = 'whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-2.5 align-middle'
const TD_ICON = 'whitespace-nowrap px-0 py-2.5 align-middle text-center'
const stopRowClick = (e: React.MouseEvent) => e.stopPropagation()

export const ALERTS_FLEX_COLUMN = 'alert'

export interface AlertColumnDeps {
  t: TFunction
  tagCatalog: AlertTag[]
  selected: Set<string>
  expandedEchoes: Set<string>
  allChecked: boolean
  onTogglePage: () => void
  onToggle: (id: string) => void
  onCreateRule: (alert: Alert) => void
  onIncident: (alert: Alert) => void
  onToggleEchoes: (id: string) => void
  onStatus: (alert: Alert, status: string, observation: string, fp: boolean) => void
}

export function buildAlertColumns({
  t,
  tagCatalog,
  selected,
  expandedEchoes,
  allChecked,
  onTogglePage,
  onToggle,
  onCreateRule,
  onIncident,
  onToggleEchoes,
  onStatus,
}: AlertColumnDeps): ColumnDef<Alert>[] {
  return [
    {
      id: 'severityBar',
      header: () => null,
      size: 6,
      minSize: 6,
      enableResizing: false,
      meta: { headerClassName: `${TH} p-0`, cellClassName: 'relative p-0' },
      cell: ({ row }) => <SeverityBar alert={row.original} />,
    },
    {
      id: 'select',
      header: () => (
        <button onClick={onTogglePage} className="mx-auto flex h-4 w-4 items-center justify-center rounded border border-input">
          {allChecked && <span className="h-2 w-2 rounded-sm bg-primary" />}
        </button>
      ),
      size: 36,
      minSize: 36,
      enableResizing: false,
      meta: {
        headerClassName: cn(TH, 'px-0'),
        cellClassName: cn(TD_ICON, 'cursor-pointer'),
        cellProps: (a) => ({
          onClick: (e) => {
            e.stopPropagation()
            onToggle(a.id)
          },
        }),
      },
      cell: ({ row }) => (
        <span
          className={cn(
            'mx-auto flex h-4 w-4 items-center justify-center rounded border',
            selected.has(row.original.id) ? 'border-primary bg-primary' : 'border-input',
          )}
        >
          {selected.has(row.original.id) && <span className="h-2 w-2 rounded-sm bg-primary-foreground" />}
        </span>
      ),
    },
    {
      id: 'createRule',
      header: t('alerts.table.actions'),
      size: 38,
      minSize: 38,
      enableResizing: false,
      // The label is wider than this icon column and reads across the next
      // (header-less) one, so it must be allowed to spill.
      meta: { headerClassName: cn(TH, 'overflow-visible'), cellClassName: TD_ICON },
      cell: ({ row }) => (
        <button
          onClick={(e) => {
            e.stopPropagation()
            onCreateRule(row.original)
          }}
          title={t('alerts.row.createRuleFromAlert')}
          aria-label={t('alerts.row.createRuleFromAlert')}
          className="mx-auto flex h-7 w-7 items-center justify-center rounded text-muted-foreground/60 transition hover:bg-background hover:text-primary"
        >
          <Tag size={13} />
        </button>
      ),
    },
    {
      id: 'incident',
      header: () => null,
      size: 38,
      minSize: 38,
      enableResizing: false,
      meta: { headerClassName: TH, cellClassName: cn(TD_ICON, 'overflow-visible') },
      cell: ({ row }) => (
        <div className="flex justify-center">
          <AlertIncidentTarget alert={row.original} onIncident={onIncident} />
        </div>
      ),
    },
    {
      id: ALERTS_FLEX_COLUMN,
      header: t('alerts.table.alert'),
      size: 360,
      minSize: 120,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => <AlertSummaryCell alert={row.original} tagCatalog={tagCatalog} />,
    },
    {
      id: 'status',
      header: t('alerts.table.status'),
      size: 130,
      minSize: 100,
      meta: { headerClassName: TH, cellClassName: cn(TD, 'overflow-visible'), cellProps: () => ({ onClick: stopRowClick }) },
      cell: ({ row }) => (
        <StatusChangeMenu
          status={statusKey(row.original)}
          variant="pill"
          tagCatalog={tagCatalog}
          onStatus={(status, observation, fp) => onStatus(row.original, status, observation, fp)}
        />
      ),
    },
    {
      id: 'technique',
      header: t('alerts.table.technique'),
      size: 180,
      minSize: 60,
      meta: {
        headerClassName: TH,
        cellClassName: cn(TD, 'font-mono text-[11px] text-muted-foreground'),
        cellProps: (a) => ({ title: a.technique }),
      },
      cell: ({ row }) => <span className="block truncate">{row.original.technique || '—'}</span>,
    },
    {
      id: 'source',
      header: t('alerts.table.source'),
      size: 160,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => <EndpointMini ep={row.original.target} />,
    },
    {
      id: 'adversary',
      header: t('alerts.table.adversary'),
      size: 160,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => (
        <div className="block truncate">
          <EndpointMini ep={row.original.adversary} accent />
        </div>
      ),
    },
    {
      id: 'severity',
      header: t('alerts.table.severity'),
      size: 90,
      minSize: 70,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => <SeverityBadge alert={row.original} />,
    },
    {
      id: 'echoes',
      header: t('alerts.table.echoes'),
      size: 90,
      minSize: 60,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => (
        <EchoesChip
          count={row.original.echoes ?? 0}
          expanded={expandedEchoes.has(row.original.id)}
          onClick={() => onToggleEchoes(row.original.id)}
        />
      ),
    },
    {
      id: 'time',
      header: t('alerts.table.time'),
      size: 160,
      minSize: 100,
      meta: {
        headerClassName: TH,
        cellClassName: cn(TD, 'font-mono text-[11px] text-muted-foreground'),
        cellProps: (a) => ({ title: relativeTime(a[TS]) }),
      },
      cell: ({ row }) => <span className="block truncate">{absTime(row.original[TS])}</span>,
    },
  ]
}
