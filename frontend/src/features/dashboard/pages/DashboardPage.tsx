import { useEffect, useMemo, useRef, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { LayoutDashboard, Loader2, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { ConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import { presetRange, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { useDashboard, useDashboards } from '@/features/dashboard/hooks/useDashboards'
import { useVisualizations, useVisualizationMutations } from '@/features/dashboard/hooks/useVisualizations'
import { useDashboardEditor } from '@/features/dashboard/hooks/useDashboardEditor'
import { DashboardGrid } from '@/features/dashboard/components/DashboardGrid'
import { DashboardEditorBar } from '@/features/dashboard/components/DashboardEditorBar'
import { DashboardFormDialog } from '@/features/dashboard/components/DashboardFormDialog'
import { DashboardTimeRange } from '@/features/dashboard/components/DashboardTimeRange'
import { DashboardRefreshInterval } from '@/features/dashboard/components/DashboardRefreshInterval'
import {
  DashboardFilterBar,
  type ChipValueMap,
} from '@/features/dashboard/components/DashboardFilterBar'
import { DashboardGallery } from '@/features/dashboard/components/DashboardGallery'
import { DashboardPreviewHeader } from '@/features/dashboard/components/DashboardPreviewHeader'
import { DEFAULT_PAGE_SIZE, DEFAULT_WIDGET_LAYOUT } from '@/features/dashboard/constants'
import { nextRow, serializeLayout, toGridItems } from '@/features/dashboard/utils/layout'
import type { Dashboard, DashboardFilterChip, FilterType, GridLayoutItem } from '@/features/dashboard/types'

export function DashboardPage() {
  const { t } = useTranslation()
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [time, setTime] = useState<TimeRange>(() => presetRange('24h'))
  const [refreshMs, setRefreshMs] = useState<number | null>(null)
  const [formOpen, setFormOpen] = useState<null | { mode: 'create' | 'rename'; target: Dashboard | null }>(
    null
  )
  const [pendingDelete, setPendingDelete] = useState<Dashboard | null>(null)
  // Removing a widget now deletes its visualization for good (no more "unlink,
  // keep it around for another dashboard") — confirm before committing a save
  // that includes pending removals.
  const [confirmRemovals, setConfirmRemovals] = useState(false)
  // Chip *values* are session-only (v11 parity): they clear when the user
  // switches dashboards. The *config* lives on dashboard.filters and persists.
  const [chipValues, setChipValues] = useState<ChipValueMap>({})
  // When entering the preview from the table's edit action we want the layout
  // editor to open automatically once the dashboard data is available.
  const pendingEditRef = useRef(false)

  const dashboards = useDashboards({
    page: 0,
    size: DEFAULT_PAGE_SIZE,
    name: search || undefined,
  })
  const dashboardItems = dashboards.list.data?.data ?? []
  const noDashboards = dashboardItems.length === 0 && !search

  const selectedDashboard = useDashboard(selectedId)
  const navigate = useNavigate()
  const location = useLocation()

  // Coming back from the visualization editor (create/edit/cancel) via its
  // "back to dashboard" navigation — re-select the dashboard it belongs to
  // and resume editing (the only way to reach that editor is from edit mode,
  // so returning should never silently drop back to view mode).
  useEffect(() => {
    const state = location.state as { selectDashboardId?: string } | null
    if (state?.selectDashboardId != null) {
      setSelectedId(state.selectDashboardId)
      pendingEditRef.current = true
      navigate(location.pathname, { replace: true, state: null })
    }
    // Only react to navigation-state changes, not every render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.state])

  const visualizations = useVisualizations(
    selectedId != null ? { dashboardId: selectedId, page: 0, size: 500 } : { page: 0, size: 0 }
  )
  const vizItems = visualizations.data?.data ?? []
  const vizMutations = useVisualizationMutations()

  const vizById = useMemo(() => {
    const m = new Map<string, (typeof vizItems)[number]>()
    for (const v of vizItems) m.set(v.id, v)
    return m
  }, [vizItems])

  const initialItems = useMemo(() => toGridItems(vizItems), [vizItems])
  const editor = useDashboardEditor(initialItems)

  const chips = useMemo<DashboardFilterChip[]>(
    () => parseChipConfig(selectedDashboard.data?.filters),
    [selectedDashboard.data?.filters]
  )

  // Reset chip *values* whenever the dashboard changes; chip *config* persists.
  useEffect(() => {
    setChipValues({})
  }, [selectedId])

  const activeFilters = useMemo<FilterType[]>(
    () => chipsToFilters(chips, chipValues),
    [chips, chipValues]
  )

  // Once the selected dashboard resolves after a table "Edit" click, drop into edit mode.
  useEffect(() => {
    if (!pendingEditRef.current) return
    if (!selectedDashboard.data) return
    if (selectedDashboard.data.systemOwner) {
      pendingEditRef.current = false
      return
    }
    pendingEditRef.current = false
    editor.enter()
  }, [selectedDashboard.data, editor])

  const handleCreateDashboard = (data: { name: string; description?: string }) => {
    dashboards.createDashboard.mutate(
      { name: data.name, description: data.description, config: '' },
      {
        onSuccess: (created) => {
          toast.success(t('dashboards.toast.created'))
          // Freshly created dashboards drop straight into edit mode.
          pendingEditRef.current = true
          setSelectedId(created.id)
          setFormOpen(null)
        },
        onError: (err) => toast.error(err.message ?? t('dashboards.toast.createFailed')),
      }
    )
  }

  const handleRenameDashboard = (data: { name: string; description?: string }) => {
    if (!formOpen?.target) return
    const target = formOpen.target
    dashboards.updateDashboard.mutate(
      {
        id: target.id,
        name: data.name,
        description: data.description,
        config: target.config,
      },
      {
        onSuccess: () => {
          toast.success(t('dashboards.toast.updated'))
          setFormOpen(null)
        },
        onError: (err) => toast.error(err.message ?? t('dashboards.toast.updateFailed')),
      }
    )
  }

  const confirmDeleteDashboard = () => {
    if (!pendingDelete) return
    const target = pendingDelete
    dashboards.deleteDashboard.mutate(target.id, {
      onSuccess: () => {
        toast.success(t('dashboards.toast.deleted'))
        if (selectedId === target.id) setSelectedId(null)
        setPendingDelete(null)
      },
      onError: (err) => toast.error(err.message ?? t('dashboards.toast.deleteFailed')),
    })
  }

  // No more picking from an existing library — a widget can only ever belong to
  // this dashboard, so "Add widget" goes straight to creating a new one, seeded
  // with the next free grid row.
  const handleAddWidget = () => {
    if (selectedId == null) return
    const layout = serializeLayout({
      x: 0,
      y: nextRow(initialItems),
      w: DEFAULT_WIDGET_LAYOUT.w,
      h: DEFAULT_WIDGET_LAYOUT.h,
    })
    navigate(`/dashboards/${selectedId}/visualizations/new`, { state: { layout } })
  }

  const handleEditWidget = (id: string) => {
    if (selectedId == null) return
    navigate(`/dashboards/${selectedId}/visualizations/${id}`)
  }

  const doSave = async () => {
    if (selectedId == null) return
    try {
      // Persist x/y/w/h for every widget whose position or size changed.
      const baseById = new Map(editor.baseline.map((it) => [it.i, it]))
      for (const item of editor.working) {
        const base = baseById.get(item.i)
        const changed =
          !base ||
          base.x !== item.x ||
          base.y !== item.y ||
          base.w !== item.w ||
          base.h !== item.h
        if (!changed) continue
        const viz = vizById.get(item.i)
        if (!viz) continue
        await vizMutations.updateVisualization.mutateAsync({
          id: viz.id,
          dashboardId: viz.dashboardId,
          spec: viz.spec,
          config: viz.config,
          layout: serializeLayout({ x: item.x, y: item.y, w: item.w, h: item.h }),
        })
      }

      // A removed widget's visualization is gone for good — there's nothing to
      // "unlink" anymore. `confirmRemovals` (below) gates this before it runs.
      for (const removeId of editor.pendingRemovals) {
        await vizMutations.deleteVisualization.mutateAsync(removeId)
      }

      editor.commit()
      setConfirmRemovals(false)
      toast.success(t('dashboards.toast.layoutSaved'))
    } catch (err) {
      toast.error((err as Error).message ?? t('dashboards.toast.layoutSaveFailed'))
    }
  }

  const handleSave = () => {
    if (!editor.dirty || selectedId == null) return
    if (editor.pendingRemovals.length > 0) {
      setConfirmRemovals(true)
      return
    }
    void doSave()
  }

  const openFromTable = (id: string, options?: { edit?: boolean }) => {
    pendingEditRef.current = !!options?.edit
    setSelectedId(id)
  }

  const handleSaveFilters = (next: DashboardFilterChip[]) => {
    const target = selectedDashboard.data
    if (!target) return
    dashboards.updateDashboard.mutate(
      {
        id: target.id,
        name: target.name,
        description: target.description,
        config: target.config,
        filters: JSON.stringify(next),
      },
      {
        onSuccess: () => {
          toast.success(t('dashboards.toast.filtersSaved'))
          // Drop any values whose chip was removed/renamed.
          const validIds = new Set(next.map((c) => c.id))
          setChipValues((prev) => {
            const out: ChipValueMap = {}
            for (const k of Object.keys(prev)) if (validIds.has(k)) out[k] = prev[k]
            return out
          })
        },
        onError: (err) => toast.error(err.message ?? t('dashboards.toast.filtersSaveFailed')),
      }
    )
  }

  const backToList = () => {
    if (editor.editing) editor.discard()
    setSelectedId(null)
  }

  const gridItems: GridLayoutItem[] = editor.editing ? editor.working : initialItems
  const saving = vizMutations.updateVisualization.isPending || vizMutations.deleteVisualization.isPending

  const inPreview = selectedId != null
  const previewDashboard = selectedDashboard.data ?? null
  const canEditPreview =
    previewDashboard != null && !previewDashboard.systemOwner && !editor.editing

  return (
    <div className="flex h-full w-full flex-col gap-4 px-6 pb-6 pt-3">
      {inPreview && previewDashboard && (
        <DashboardPreviewHeader
          dashboard={previewDashboard}
          onBack={backToList}
          onEdit={canEditPreview ? editor.enter : undefined}
          onDelete={(d) => setPendingDelete(d)}
          right={
            <div className="flex flex-wrap items-center gap-2">
              <DashboardTimeRange value={time} onChange={setTime} />
              <DashboardRefreshInterval value={refreshMs} onChange={setRefreshMs} />
              {editor.editing && (
                <DashboardEditorBar
                  dirty={editor.dirty}
                  saving={saving}
                  onSave={handleSave}
                  onDiscard={editor.discard}
                  onAddWidget={handleAddWidget}
                />
              )}
            </div>
          }
        />
      )}

      {!inPreview && (
        <>
          {noDashboards && !dashboards.list.isLoading ? (
            <div className="flex h-full min-h-[320px] flex-col items-center justify-center gap-3 rounded-lg border border-border bg-background/30 p-6 text-center">
              <LayoutDashboard size={32} className="text-muted-foreground/40" />
              <div>
                <p className="text-sm font-medium">{t('dashboards.empty.title')}</p>
                <p className="mt-1 text-xs text-muted-foreground">{t('dashboards.empty.body')}</p>
              </div>
              <Button size="sm" onClick={() => setFormOpen({ mode: 'create', target: null })}>
                <Plus size={14} className="mr-1" /> {t('dashboards.empty.cta')}
              </Button>
            </div>
          ) : (
            <DashboardGallery
              dashboards={dashboardItems}
              loading={dashboards.list.isLoading}
              search={search}
              onSearchChange={setSearch}
              onSelect={(id) => openFromTable(id)}
              onCreate={() => setFormOpen({ mode: 'create', target: null })}
            />
          )}
        </>
      )}

      {inPreview && previewDashboard && (chips.length > 0 || editor.editing) && (
        <DashboardFilterBar
          chips={chips}
          values={chipValues}
          onChange={setChipValues}
          editable={editor.editing}
          savingChips={dashboards.updateDashboard.isPending}
          onSaveChips={handleSaveFilters}
        />
      )}

      {inPreview && (
        <div className="min-h-0 flex-1 overflow-y-auto rounded-lg border border-border bg-background/30 p-2">
          {visualizations.isLoading ? (
            <div className="flex h-full min-h-[300px] items-center justify-center gap-2 text-sm text-muted-foreground">
              <Loader2 size={16} className="animate-spin" />
              {t('dashboards.page.loading')}
            </div>
          ) : (
            <DashboardGrid
              items={gridItems}
              visualizations={vizItems}
              time={time}
              filters={activeFilters}
              refreshMs={refreshMs}
              editing={editor.editing}
              onLayoutChange={editor.replace}
              onEditItem={handleEditWidget}
              onRemoveItem={editor.remove}
            />
          )}
        </div>
      )}

      <DashboardFormDialog
        open={formOpen?.mode === 'create'}
        mode="create"
        initial={null}
        busy={dashboards.createDashboard.isPending}
        onClose={() => setFormOpen(null)}
        onSubmit={handleCreateDashboard}
      />

      <DashboardFormDialog
        open={formOpen?.mode === 'rename'}
        mode="rename"
        initial={formOpen?.target ?? null}
        busy={dashboards.updateDashboard.isPending}
        onClose={() => setFormOpen(null)}
        onSubmit={handleRenameDashboard}
      />

      <ConfirmDialog
        open={pendingDelete != null}
        title={t('dashboards.confirm.deleteTitle')}
        body={t('dashboards.confirm.delete', { name: pendingDelete?.name ?? '' })}
        confirmLabel={t('dashboards.list.delete') ?? undefined}
        danger
        busy={dashboards.deleteDashboard.isPending}
        onClose={() => setPendingDelete(null)}
        onConfirm={confirmDeleteDashboard}
      />

      <ConfirmDialog
        open={confirmRemovals}
        title={t('dashboards.confirm.deleteWidgetsTitle')}
        body={t('dashboards.confirm.deleteWidgets', { count: editor.pendingRemovals.length })}
        confirmLabel={t('dashboards.list.delete') ?? undefined}
        danger
        busy={saving}
        onClose={() => setConfirmRemovals(false)}
        onConfirm={() => void doSave()}
      />
    </div>
  )
}

export default DashboardPage

function parseChipConfig(json: string | undefined | null): DashboardFilterChip[] {
  if (!json) return []
  try {
    const parsed = JSON.parse(json)
    return Array.isArray(parsed) ? (parsed as DashboardFilterChip[]) : []
  } catch {
    return []
  }
}

function chipsToFilters(
  chips: DashboardFilterChip[],
  values: ChipValueMap
): FilterType[] {
  const out: FilterType[] = []
  for (const chip of chips) {
    const v = values[chip.id]
    if (v == null) continue
    if (Array.isArray(v)) {
      if (v.length === 0) continue
      out.push({ field: chip.field, operator: 'IS_ONE_OF_TERMS', value: v })
    } else if (typeof v === 'string' && v !== '') {
      out.push({ field: chip.field, operator: 'IS', value: v })
    }
  }
  return out
}
