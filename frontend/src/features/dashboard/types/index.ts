export interface Dashboard {
  id: number
  name: string
  description?: string
  config?: string
  systemOwner?: boolean
  createdDate?: string
  modifiedDate?: string
}

export interface Visualization {
  id: number
  name: string
  description?: string
  sqlQuery: string
  config: string
  systemOwner?: boolean
  createdDate?: string
  modifiedDate?: string
}

export interface DashboardVisualization {
  id: number
  idDashboard: number
  idVisualization: number
  layout: string
}

export interface DashboardCreateInput {
  name: string
  description?: string
  config?: string
}

export interface DashboardUpdateInput {
  id: number
  name: string
  description?: string
  config?: string
}

export interface VisualizationCreateInput {
  name: string
  description?: string
  sqlQuery: string
  config: string
}

export interface VisualizationUpdateInput {
  id: number
  name: string
  description?: string
  sqlQuery: string
  config: string
}

export interface DashboardLayoutCreateInput {
  idDashboard: number
  idVisualization: number
  layout: string
}

export interface DashboardLayoutUpdateInput {
  id: number
  idDashboard: number
  idVisualization: number
  layout: string
}

export interface DashboardListParams {
  name?: string
  page?: number
  size?: number
}

export interface VisualizationListParams {
  name?: string
  page?: number
  size?: number
}

export interface DashboardLayoutListParams {
  idDashboard?: number
  idVisualization?: number
  page?: number
  size?: number
}

export interface GridLayoutItem {
  i: string
  x: number
  y: number
  w: number
  h: number
}

export interface WidgetLayout {
  x: number
  y: number
  w: number
  h: number
  order?: number
}
