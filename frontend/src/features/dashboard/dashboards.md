# Legacy dashboards — reference for the React port

This document describes how the dashboards feature works in the **Angular app at `frontend-legacy/`**. It exists to give engineers porting the feature to the new React frontend at `frontend/` a single source of truth for legacy behavior — what to replicate, where it lives, and which endpoints/models it relies on.

The new Go backend at `backend/modules/dashboards/` exposes a **narrower** surface than the legacy backend, so not every behavior described here can be re-implemented immediately. The "Endpoint mapping" and "Behaviors that need backend work" sections call out the gaps explicitly.

---

## 1. Routing & entry points

The legacy app exposes **two route trees** for dashboards:

| Route | Component | Purpose |
|---|---|---|
| `/dashboard/overview` | `DashboardOverviewComponent` | System "home" dashboard, hardcoded charts |
| `/dashboard/view/:name` | `DashboardViewComponent` | Legacy iframe wrapper for menu integration |
| `/dashboard/render/:id/:dashboard` | `DashboardRenderComponent` | Main read-only viewer (gridster grid) |
| `/dashboard/export/:id/:dashboard` | `DashboardExportPdfComponent` | PDF export |
| `/dashboard/customize-export/:id/:dashboard` | `DashboardExportCustomComponent` | Per-widget export selection |
| `/dashboard/log-sources` | `DashboardLogSourcesComponent` | Log-sources status mini-dashboard |
| `/creator/dashboard/list` | `DashboardListComponent` | Dashboard management list (edit/delete/export) |
| `/creator/dashboard/builder` | `DashboardCreateComponent` | Drag-drop dashboard editor |

Route registration:
- `frontend-legacy/src/app/dashboard/dashboard-routing.module.ts`
- `frontend-legacy/src/app/graphic-builder/dashboard-builder/dashboard-builder-routing.module.ts`

Guards: `UserRouteAccessService` enforces `USER_ROLE` / `ADMIN_ROLE`.

The route resolver `DashboardResolverService` (`src/app/dashboard/shared/services/dashboard-resolver.service.ts`) preloads the dashboard's visualizations before `DashboardRenderComponent` mounts and populates `RenderLayoutService`.

---

## 2. Page components

| Component | Path | Role |
|---|---|---|
| `DashboardCreateComponent` | `graphic-builder/dashboard-builder/dashboard-create/` | Drag-drop editor: gridster + add/remove/edit widgets |
| `DashboardListComponent` | `graphic-builder/dashboard-builder/dashboard-list/` | Dashboards table with edit/delete/export actions |
| `DashboardRenderComponent` | `dashboard/dashboard-render/` | Read-only gridster view with filters/time |
| `DashboardOverviewComponent` | `dashboard/dashboard-overview/` | System overview (hardcoded widgets, 30s refresh) |
| `DashboardSaveComponent` | `graphic-builder/dashboard-builder/dashboard-save/` | Save modal: name, refresh interval, menu integration |
| `DashboardImportComponent` | `graphic-builder/dashboard-builder/dashboard-import/` | Multi-step JSON import wizard |
| `DashboardExportPdfComponent` | `dashboard/dashboard-export-pdf/` | Renders a dashboard for PDF print |
| `DashboardFilterCreateComponent` | (filter modal) | Add a dashboard-level filter definition |

---

## 3. Data flow & state

**HTTP services** (Angular `HttpClient`):

| Service | Endpoint root |
|---|---|
| `UtmDashboardService` | `api/utm-dashboards` |
| `UtmDashboardVisualizationService` | `api/utm-dashboard-visualizations` |
| `VisualizationService` | `api/utm-visualizations` |
| `UtmRenderVisualization` | `api/utm-dashboard-visualizations` (query with filters) |

**Reactive state** (RxJS `BehaviorSubject`s in `src/app/shared/behaviors/`):

| Subject | Holds |
|---|---|
| `DashboardBehavior.$dashboard` | Current menu/dashboard context |
| `DashboardBehavior.$filterDashboard` | Active filter values `{ indexPattern, filters[] }` |
| `TimeFilterBehavior.$time` | Global time range `{ from, to, update }` |

**Layout state**:
- `LayoutService` (editor) — array of `{ grid: GridsterItem; visualization: VisualizationType }` with `addItem`, `setItem`, `deleteItem`, `dropItem`
- `RenderLayoutService` (viewer) — populated by the route resolver

Widgets react to filter and time changes via `rebuildVisualizationFilterTime()` (`src/app/shared/util/`).

---

## 4. Domain types (TypeScript DTOs in legacy)

```ts
UtmDashboardType {
  id, name, description,
  createdDate, modifiedDate, userCreated, userModified,
  refreshTime?: number,        // ms — auto-refresh interval
  filters?: string,            // JSON-stringified DashboardFilterType[]
  systemOwner?: boolean,
}

UtmDashboardVisualizationType {       // join table
  id, idDashboard, idVisualization,
  visualization, dashboard,           // populated by API
  left, top, width, height, order,    // grid coordinates
  gridInfo,                           // raw gridster JSON
  defaultTimeRange?, showTimeFilter?,
}

VisualizationType {
  id, name, description,
  chartType: ChartTypeEnum,           // LINE | AREA | PIE | BAR | TABLE | GAUGE | METRIC | HEATMAP | COORDINATE_MAP | LIST | TEXT | TAG_CLOUD | GOAL
  chartConfig: string,                // JSON — ECharts option-ish
  chartAction: string,                // JSON — click handler config
  pattern, filterType[],
  queryType, sqlQuery, queryLanguage, eventType,
  showTime, systemOwner,
}

DashboardFilterType {
  id, indexPattern, filterLabel, field,
  multiple, searchable, clearable,
  placeholder, maxSelectedItems,
}
```

---

## 5. Visualization mechanics

### Charting library — **ngx-echarts v4.1.1** (ECharts)
- Chart factory: `src/app/shared/chart/factories/echart-factory/chart-factory.ts`
- Option builder: `src/app/shared/chart/factories/echart-factory/chart-option.ts`
- Click handlers: `src/app/shared/chart/factories/echart-click-factory/`
- Supported chart types are enumerated in `ChartTypeEnum`.

### Grid system — **angular-gridster2 v8.2.0**
- 30 columns
- Row height: 430px (fixed)
- Column width: 500px (fixed)
- Draggable + resizable in `DashboardCreateComponent`, locked in `DashboardRenderComponent`
- Position serialized as `gridInfo` JSON on `UtmDashboardVisualizationType`

### Widget config storage
- `chartConfig` and `chartAction` are JSON strings on `VisualizationType`, parsed when rendered.
- A widget on a dashboard is a row in `UtmDashboardVisualizationType` linking a `Dashboard` and a `Visualization` plus its layout/timing.

---

## 6. Filters, time range, refresh

### Dashboard-level filters
- Defined per dashboard, stored as JSON in `UtmDashboardType.filters`.
- Parsed to `DashboardFilterType[]` — each is a searchable dropdown bound to an index pattern + field.
- Created/edited via `DashboardFilterCreateComponent` (modal).
- Selection updates `DashboardBehavior.$filterDashboard` → all widgets re-query.

### Time range
- Global `TimeFilterBehavior.$time { from, to }`.
- Default 24-hour window (`TIME_DASHBOARD_REFRESH` constant).
- A widget can override via `UtmDashboardVisualizationType.defaultTimeRange` + `showTimeFilter`.

### Auto-refresh
- `UtmDashboardType.refreshTime` (ms, nullable).
- Configured in `DashboardSaveComponent`.
- Hardcoded 30s in `DashboardOverviewComponent`.

---

## 7. Import / export

### Import
- `DashboardImportComponent` is a wizard: upload JSON file → preview detected dashboards → choose overrides → submit.
- Calls `POST /api/utm-dashboards/import` with `{ dashboards: UtmDashboardVisualizationType[], override: boolean }`.

### Export
- PDF via `DashboardExportPdfComponent` — re-renders the gridster grid into a print stylesheet, then `window.print()`.
- Customizable export via `DashboardExportCustomComponent` — user selects which widgets to include.
- All data serialization happens client-side.

### Templates / system dashboards
- `systemOwner: boolean` marks a dashboard (or visualization) as system-owned and read-only for end users.
- `DashboardOverviewComponent` is a fully hardcoded system dashboard, not configurable.

---

## 8. Legacy endpoint catalogue

| Method | Path | Purpose |
|---|---|---|
| POST | `/api/utm-dashboards` | Create dashboard |
| PUT | `/api/utm-dashboards` | Update dashboard |
| GET | `/api/utm-dashboards` | List (paginated) |
| GET | `/api/utm-dashboards/{id}` | Fetch one |
| DELETE | `/api/utm-dashboards/{id}` | Delete |
| POST | `/api/utm-dashboards/import` | Bulk import w/ override flag |
| POST | `/api/utm-dashboard-visualizations` | Place a visualization on a dashboard |
| PUT | `/api/utm-dashboard-visualizations` | Update placement |
| GET | `/api/utm-dashboard-visualizations` | List (filtered by `idDashboard.equals`) |
| DELETE | `/api/utm-dashboard-visualizations/{id}` | Remove placement |
| POST | `/api/utm-visualizations` | Create visualization |
| PUT | `/api/utm-visualizations` | Update visualization |
| GET | `/api/utm-visualizations` | List |
| GET | `/api/utm-visualizations/{id}` | Fetch one |
| DELETE | `/api/utm-visualizations/{id}` | Delete |
| DELETE | `/api/utm-visualizations/bulk-delete/{ids}` | Bulk delete |
| POST | `/api/utm-visualizations/run` | Execute query with params |
| POST | `/api/utm-visualizations/batch` | Bulk import visualizations |

---

## 9. Key files & their roles

| # | Path | Role |
|---|---|---|
| 1 | `src/app/dashboard/dashboard-routing.module.ts` | Viewer routes |
| 2 | `src/app/graphic-builder/dashboard-builder/dashboard-builder-routing.module.ts` | Editor routes |
| 3 | `src/app/graphic-builder/dashboard-builder/dashboard-create/dashboard-create.component.ts` | Drag-drop editor |
| 4 | `src/app/graphic-builder/dashboard-builder/dashboard-list/dashboard-list.component.ts` | List/management UI |
| 5 | `src/app/dashboard/dashboard-render/dashboard-render.component.ts` | Viewer |
| 6 | `src/app/graphic-builder/dashboard-builder/shared/services/utm-dashboard.service.ts` | Dashboards HTTP service |
| 7 | `src/app/graphic-builder/dashboard-builder/shared/services/utm-dashboard-visualization.service.ts` | Placement HTTP service |
| 8 | `src/app/graphic-builder/visualization/shared/services/visualization.service.ts` | Visualizations HTTP service |
| 9 | `src/app/dashboard/shared/services/dashboard-resolver.service.ts` | Route resolver |
| 10 | `src/app/dashboard/shared/services/render-layout.service.ts` | Viewer layout state |
| 11 | `src/app/graphic-builder/dashboard-builder/shared/services/layout.service.ts` | Editor layout state |
| 12 | `src/app/shared/chart/types/dashboard/utm-dashboard.type.ts` | Dashboard DTO |
| 13 | `src/app/shared/chart/types/dashboard/utm-dashboard-visualization.type.ts` | Placement DTO |
| 14 | `src/app/shared/chart/types/visualization.type.ts` | Visualization DTO |
| 15 | `src/app/shared/types/filter/dashboard-filter.type.ts` | Filter DTO |
| 16 | `src/app/shared/behaviors/dashboard.behavior.ts` | Filter/context RxJS state |
| 17 | `src/app/shared/behaviors/time-filter.behavior.ts` | Time-range RxJS state |
| 18 | `src/app/shared/chart/factories/echart-factory/chart-factory.ts` | ECharts factory |
| 19 | `src/app/graphic-builder/dashboard-builder/dashboard-save/dashboard-save.component.ts` | Save modal |
| 20 | `src/app/graphic-builder/dashboard-builder/dashboard-import/dashboard-import.component.ts` | Import wizard |

---

## 10. Legacy ↔ new backend mapping

The new Go module `backend/modules/dashboards/` does NOT 1:1 mirror the legacy backend. Mapping:

| Legacy resource | Legacy URL | New URL | Notes |
|---|---|---|---|
| `UtmDashboard` | `/api/utm-dashboards` | `/dashboards` | Loses `filters`, `refreshTime` (no fields in new domain). `systemOwner` survives. |
| `UtmDashboardVisualization` | `/api/utm-dashboard-visualizations` | — | **Removed.** The new backend is 1 dashboard : many visualizations, not N:M — there's no join table. `gridInfo`/`left`/`top`/etc. collapsed into a single `Layout` JSON string that now lives directly on `Visualization`. `defaultTimeRange`/`showTimeFilter` not in new domain. |
| `UtmVisualization` | `/api/utm-visualizations` | `/visualizations` | Many fields gone: `chartType` (separate column → moved into `Config`), `pattern`, `filterType[]`, `queryType`, `queryLanguage`, `eventType`, `chartAction`, `showTime`, plus `Name`/`Description` (each widget belongs to exactly one dashboard now, so a separate label serves no purpose). Gained `DashboardID` (owning dashboard, required) and `Layout` (grid position, from the former join table). Only `Spec`, `Config`, `SystemOwner`, `DashboardID`, `Layout` remain — and `SQLQuery` became `Spec`: a widget states the question (dataset, chart, breakdown, filters) instead of carrying SQL. |
| Query execution | `POST /api/utm-visualizations/run` | `POST /visualizations/query` | Takes a spec and answers it against the event store. There is no SQL: the tenant comes from the session and cannot be named in the request. |
| Bulk import/export | `POST /api/utm-dashboards/import`, `POST /api/utm-visualizations/batch` | — | **Not exposed.** Will require new backend endpoints. |
| Filters | stored on `UtmDashboardType.filters` | — | **No backend field.** Would have to live inside `Dashboard.Config` JSON. |
| Refresh | `UtmDashboardType.refreshTime` | — | Same — push into `Dashboard.Config` JSON or add a column. |

---

## 11. Behaviors to preserve in the React port

Targeted for the initial port (Backend-MVP + client-only time range):

1. **Gridster-like drag/resize editor** — implemented with `react-grid-layout` (12-col responsive grid).
2. **1:N dashboard ↔ visualization model** — a visualization belongs to exactly one dashboard and is not reusable across dashboards; this is a deliberate departure from the legacy N:M model, decided during the React port. There is no join table — `DashboardID` + `Layout` live directly on `Visualization`.
3. **JSON-config rendering** — `Visualization.Config` is treated as ECharts option JSON merged with a runtime time range.
4. **System-owned read-only badge** — surface `systemOwner` in the UI; disable destructive actions.
5. **Client-side time range** — reuse `@/shared/components/ui/TimeRangePicker`; pass to widget queries (not persisted yet).

## 12. Behaviors that need backend work before they can be ported

| Behavior | Blocker |
|---|---|
| Global dashboard filters | No `filters` field on new `Dashboard` (would need `Config` schema + frontend chip UI) |
| Auto-refresh interval | No `refreshTime` field |
| JSON import / export | No backend endpoints |
| PDF export | Could be client-only (print stylesheet), but needs design |
| Per-widget time override | No `defaultTimeRange`/`showTimeFilter` fields on new `Visualization` |
| Visualization editor / chart builder | New `Visualization` flattens many legacy fields; the editor's UX needs a new shape |
| Query execution against datasources | Lives outside `dashboards` module; frontend needs to discover where |

When picking up these items, refer back to the legacy components named in §9.
