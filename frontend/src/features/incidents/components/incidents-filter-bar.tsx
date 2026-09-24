import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { CustomFilterBar } from '@/shared/components/filters/CustomFilterBar'
import type { CustomFilter, FilterFieldDef, FilterOpDef } from '@/shared/components/filters/custom-filter.types'

export const ASSIGNEE_FIELD = 'incidentAssignedTo'

/** The filters the incident list can apply beyond the toolbar's own controls. */
export function IncidentsFilterBar({
  filters,
  assigneeOptions,
  onAdd,
  onUpdate,
  onRemove,
  onClear,
}: {
  filters: CustomFilter[]
  assigneeOptions: string[]
  onAdd: (f: CustomFilter) => void
  onUpdate: (i: number, f: CustomFilter) => void
  onRemove: (i: number) => void
  onClear: () => void
}) {
  const { t } = useTranslation()

  const fields = useMemo<FilterFieldDef[]>(() => [{ field: ASSIGNEE_FIELD, label: t('incidents.table.assignee') }], [t])
  const operators = useMemo<FilterOpDef[]>(() => [{ id: 'IS', label: t('incidents.ops.IS'), needsValue: true }], [t])
  const labels = useMemo(
    () => ({
      add: t('incidents.filters.add'),
      clearAll: t('incidents.filters.clearAll'),
      filterValues: t('incidents.filters.filterValues'),
      loadingValues: t('incidents.filters.loadingValues'),
      noValues: t('incidents.filters.noValues'),
      pickValue: t('incidents.filters.pickValue'),
      empty: t('incidents.filters.empty'),
      cancel: t('incidents.filters.cancel'),
      addBtn: t('incidents.filters.addBtn'),
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
      fetchValues={async () => assigneeOptions.map((value) => ({ value }))}
      labels={labels}
    />
  )
}
