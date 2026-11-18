import type { FilterType } from '../types/log-explorer.types'

export function resolveDashboardDataType(
  filters: FilterType[],
  preferredDataType: string | undefined,
  availableDataTypes: string[],
): { dataType: string | null; filters: FilterType[] } {
  const dataTypeFilter = filters.find((filter) => filter.field === 'dataType')
  const filterValues = Array.isArray(dataTypeFilter?.value)
    ? dataTypeFilter.value
    : dataTypeFilter?.value == null
      ? []
      : [dataTypeFilter.value]
  const filterDataType = filterValues.length === 1 ? String(filterValues[0]) : undefined
  const requestedDataType = preferredDataType ?? filterDataType

  if (!requestedDataType || !availableDataTypes.includes(requestedDataType)) {
    return { dataType: null, filters }
  }

  return {
    dataType: requestedDataType,
    filters: filters.filter((filter) => filter.field !== 'dataType'),
  }
}
