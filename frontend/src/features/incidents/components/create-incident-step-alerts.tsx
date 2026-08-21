import { useEffect, useMemo, useState } from 'react'
import { AlertTriangle, Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { AlertsFilterBar } from '@/features/alerts/components/alerts-filter-bar'
import { useAlertsList } from '@/features/alerts/hooks/use-alerts-list'
import { useAlertTagCatalog } from '@/features/alerts/hooks/use-alert-tag-catalog'
import { FILTER_OPS } from '@/features/alerts/lib/alert-meta'
import type { Alert, CustomFilter, FilterType } from '@/features/alerts/types/alert.types'
import { IncidentAlertRow } from './incident-alert-row'
import { IncidentAlertsPickerHeader } from './incident-alerts-picker-header'

const COLS = 5

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
          <table className="min-w-full border-collapse">
            <IncidentAlertsPickerHeader allChecked={allChecked} onTogglePage={() => onToggleAll(alerts)} />
            <tbody>
              {loading && alerts.length === 0 ? (
                <tr>
                  <td colSpan={COLS} className="px-6 py-16 text-center text-sm text-muted-foreground">
                    <Loader2 className="mx-auto h-4 w-4 animate-spin" /> {t('alerts.list.loading')}
                  </td>
                </tr>
              ) : error ? (
                <tr>
                  <td colSpan={COLS} className="px-6 py-16 text-center text-sm">
                    <AlertTriangle size={16} className="mr-1 inline text-amber-500" />
                    {t('alerts.list.loadError')}
                    <button onClick={refresh} className="ml-2 text-primary hover:underline">
                      {t('alerts.list.retry')}
                    </button>
                  </td>
                </tr>
              ) : alerts.length === 0 ? (
                <tr>
                  <td colSpan={COLS} className="px-6 py-16 text-center text-sm text-muted-foreground">
                    {t('alerts.list.empty')}
                  </td>
                </tr>
              ) : (
                alerts.map((a) => (
                  <IncidentAlertRow
                    key={a.id}
                    alert={a}
                    tagCatalog={tagCatalog}
                    checked={selected.has(a.id)}
                    onToggle={() => onToggle(a.id)}
                  />
                ))
              )}
            </tbody>
          </table>
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
