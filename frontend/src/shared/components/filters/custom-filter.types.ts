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
}

export interface FilterValue {
  value: string
  count: number
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
