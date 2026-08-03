import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { CustomFilterBar } from '@/shared/components/filters/CustomFilterBar'
import type { CustomFilter, FilterFieldDef, FilterOpDef } from '@/shared/components/filters/custom-filter.types'
import { alertsHttpService } from '../services/alerts-http.service'
import { FILTER_FIELDS, FILTER_OPS, fieldKey } from '../lib/alert-meta'

export function AlertsFilterBar({
  filters,
  onAdd,
  onUpdate,
  onRemove,
  onClear,
}: {
  filters: CustomFilter[]
  onAdd: (f: CustomFilter) => void
  onUpdate?: (i: number, f: CustomFilter) => void
  onRemove: (i: number) => void
  onClear: () => void
}) {
  const { t } = useTranslation()

  const fields = useMemo<FilterFieldDef[]>(
    () => FILTER_FIELDS.map((f) => ({ field: f.field, label: t(`alerts.fields.${fieldKey(f.field)}`) })),
    [t],
  )
  const operators = useMemo<FilterOpDef[]>(
    () => FILTER_OPS.map((o) => ({ id: o.id, label: t(`alerts.ops.${o.id}`), needsValue: o.needsValue })),
    [t],
  )
  const labels = useMemo(
    () => ({
      add: t('alerts.filters.add'),
      clearAll: t('alerts.filters.clearAll'),
      filterValues: t('alerts.filters.filterValues'),
      loadingValues: t('alerts.filters.loadingValues'),
      noValues: t('alerts.filters.noValues'),
      pickValue: t('alerts.filters.pickValue'),
      empty: t('alerts.filters.empty'),
      cancel: t('alerts.filters.cancel'),
      addBtn: t('alerts.filters.addBtn'),
    }),
    [t],
  )

  return (
    <CustomFilterBar
      filters={filters}
      onAdd={onAdd}
      onUpdate={onUpdate}
      onRemove={onRemove}
      onClear={onClear}
      fields={fields}
      operators={operators}
      fetchValues={(field) => alertsHttpService.fieldValues(field)}
      labels={labels}
    />
  )
}
