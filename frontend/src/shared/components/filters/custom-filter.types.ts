export interface CustomFilter {
  field: string
  label: string
  operator: string
  value: string
}

export interface FilterFieldDef {
  field: string
  label: string
}

export interface FilterOpDef {
  id: string
  label: string
  needsValue: boolean
  /** Text-based operators (e.g. CONTAINS) take free input even when fetchValues is available. */
  freeText?: boolean
}

export interface FilterValue {
  value: string
  /** Shown beside the value when the source can tell how many rows carry it. */
  count?: number
}

export interface FilterBarLabels {
  add: string
  clearAll: string
  filterValues: string
  loadingValues: string
  noValues: string
  pickValue: string
  empty: string
  cancel: string
  addBtn: string
}
