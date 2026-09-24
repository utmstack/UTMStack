import { useEffect, useMemo, useState } from 'react'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import type { ColumnDef } from '@tanstack/react-table'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { ResizableDataTable } from '@/shared/components/ui/resizable-data-table'
import { cn } from '@/shared/lib/utils'
import { AlertSummaryCell, SeverityBadge, SeverityBar } from '@/features/alerts/components/alert-cells'
import { AlertsFilterBar } from '@/features/alerts/components/alerts-filter-bar'
import { useAlertsList } from '@/features/alerts/hooks/use-alerts-list'
import { useAlertTagCatalog } from '@/features/alerts/hooks/use-alert-tag-catalog'
import { FILTER_OPS, TS, absTime, relativeTime } from '@/features/alerts/lib/alert-meta'
import type { Alert, AlertTag, CustomFilter, FilterType } from '@/features/alerts/types/alert.types'

const TH = 'whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-2.5 align-middle'

export function CreateIncidentStepAlerts({
  selected,
  onToggle,
  onToggleAll,
  onAlertsChange,
}: {
  selected: Set<string>
  onToggle: (id: string) => void
  onToggleAll: (page: Alert[]) => void
  onAlertsChange: (alerts: Alert[]) => void
}) {
  const { t } = useTranslation()
  const [customFilters, setCustomFilters] = useState<CustomFilter[]>([])
  const [page, setPage] = useState(0)
  const pageSize = 50

  const filters = useMemo<FilterType[]>(() => {
    const f: FilterType[] = [{ field: 'parentId', operator: 'IS', value: '' }]
    for (const cf of customFilters) {
      const needsValue = FILTER_OPS.find((o) => o.id === cf.operator)?.needsValue ?? true
      f.push({ field: cf.field, operator: cf.operator, value: needsValue ? cf.value : undefined })
    }
    return f
  }, [customFilters])

  const { alerts, total, hasMore, loading, error, refresh } = useAlertsList(page, pageSize, filters)
  const { tagCatalog } = useAlertTagCatalog(() => {})

  useEffect(() => { onAlertsChange(alerts) }, [alerts, onAlertsChange])

  const allChecked = alerts.length > 0 && alerts.every((a) => selected.has(a.id))
  const columns = buildColumns({ t, tagCatalog, selected, allChecked, onTogglePage: () => onToggleAll(alerts) })

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="shrink-0">
        <AlertsFilterBar
          filters={customFilters}
          onAdd={(cf) => { setCustomFilters((c) => [...c, cf]); setPage(0) }}
          onUpdate={(i, cf) => { setCustomFilters((c) => c.map((f, idx) => (idx === i ? cf : f))); setPage(0) }}
          onRemove={(i) => { setCustomFilters((c) => c.filter((_, idx) => idx !== i)); setPage(0) }}
          onClear={() => { setCustomFilters([]); setPage(0) }}
        />
      </div>

      <div className="mt-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        <div className="min-h-0 flex-1 overflow-auto">
          <ResizableDataTable
            columns={columns}
            data={alerts}
            flexColumnId="alert"
            storageKey="incident-alerts-picker-sizing"
            getRowId={(a) => a.id}
            onRowClick={(a) => onToggle(a.id)}
            rowClassName={() => 'border-border/50 text-[13px] last:border-b-0 hover:bg-muted/20'}
            loading={loading && alerts.length === 0}
            loadingContent={
              <div className="px-6 py-16 text-center text-sm text-muted-foreground">
                <Loader2 className="mx-auto h-4 w-4 animate-spin" /> {t('alerts.list.loading')}
              </div>
            }
            error={error}
            errorContent={
              <div className="px-6 py-16 text-center text-sm">
                <AlertTriangle size={16} className="mr-1 inline text-amber-500" />
                {t('alerts.list.loadError')}
                <button onClick={refresh} className="ml-2 text-primary hover:underline">
                  {t('alerts.list.retry')}
                </button>
              </div>
            }
            emptyContent={
              <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('alerts.list.empty')}</div>
            }
          />
          {alerts.length > 0 && (
            <InfiniteScrollSentinel
              onReach={() => setPage((p) => p + 1)}
              hasMore={hasMore}
              loading={loading}
              endLabel={t('common.allLoaded', { count: total })}
            />
          )}
        </div>
      </div>
    </div>
  )
}

function buildColumns({
  t,
  tagCatalog,
  selected,
  allChecked,
  onTogglePage,
}: {
  t: TFunction
  tagCatalog: AlertTag[]
  selected: Set<string>
  allChecked: boolean
  onTogglePage: () => void
}): ColumnDef<Alert>[] {
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
      meta: { headerClassName: `${TH} px-0`, cellClassName: 'whitespace-nowrap px-0 py-2.5 align-middle' },
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
      id: 'alert',
      header: t('alerts.table.alert'),
      size: 360,
      minSize: 120,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => <AlertSummaryCell alert={row.original} tagCatalog={tagCatalog} />,
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
      id: 'time',
      header: t('alerts.table.time'),
      size: 160,
      minSize: 100,
      meta: {
        headerClassName: TH,
        cellClassName: `${TD} font-mono text-[11px] text-muted-foreground`,
        cellProps: (a) => ({ title: absTime(a[TS]) }),
      },
      cell: ({ row }) => <span className="block truncate">{relativeTime(row.original[TS])}</span>,
    },
  ]
}
