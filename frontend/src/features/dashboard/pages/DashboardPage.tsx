import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { LayoutDashboard, Loader2, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { ConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import { presetRange, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { useDashboard, useDashboards } from '@/features/dashboard/hooks/useDashboards'
import { useDashboardLayouts } from '@/features/dashboard/hooks/useDashboardLayouts'
import { createVisualizationsService } from '@/features/dashboard/service/visualizations.service'
import { useDashboardEditor } from '@/features/dashboard/hooks/useDashboardEditor'
import { DashboardGrid } from '@/features/dashboard/components/DashboardGrid'
import { DashboardEditorBar } from '@/features/dashboard/components/DashboardEditorBar'
import { DashboardFormDialog } from '@/features/dashboard/components/DashboardFormDialog'
import { DashboardTimeRange } from '@/features/dashboard/components/DashboardTimeRange'
import { AddVisualizationDrawer } from '@/features/dashboard/components/AddVisualizationDrawer'
import { DashboardTabsBar } from '@/features/dashboard/components/DashboardTabsBar'
import { DEFAULT_PAGE_SIZE, DEFAULT_WIDGET_LAYOUT } from '@/features/dashboard/constants'
import { nextRow, serializeLayout, toGridItems } from '@/features/dashboard/utils/layout'
import type { Dashboard, GridLayoutItem, Visualization } from '@/features/dashboard/types'
import { useQueries } from '@tanstack/react-query'
import { VISUALIZATIONS_QUERY_KEYS } from '@/features/dashboard/hooks/useVisualizations'
const DEFAULT_DASHBOARD_KEY = 'utmstack-default-dashboard'

export function DashboardPage() {
  const { t } = useTranslation()
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [time, setTime] = useState<TimeRange>(() => presetRange('24h'))
  const [formOpen, setFormOpen] = useState<null | { mode: 'create' | 'rename'; target: Dashboard | null }>(
    null
  )
  const [addOpen, setAddOpen] = useState(false)
  const [pendingDelete, setPendingDelete] = useState<Dashboard | null>(null)

  const dashboards = useDashboards({ page: 0, size: DEFAULT_PAGE_SIZE })
  const dashboardItems = dashboards.list.data?.data ?? []
  // No dashboards AND no active search → genuinely empty (show the create CTA).
  const noDashboards = dashboardItems.length === 0

  // "Default" dashboard (the one shown on entry), persisted per browser.
  const [defaultId, setDefaultId] = useState<number | null>(() => {
    if (typeof window === 'undefined') return null
    const raw = window.localStorage.getItem(DEFAULT_DASHBOARD_KEY)
    const n = raw == null ? NaN : Number(raw)
    return Number.isFinite(n) ? n : null
  })
  const setDefaultDashboard = (id: number | null) => {
    setDefaultId(id)
    try {
      if (id == null) window.localStorage.removeItem(DEFAULT_DASHBOARD_KEY)
      else window.localStorage.setItem(DEFAULT_DASHBOARD_KEY, String(id))
    } catch {
      /* ignore */
    }
  }

  // On entry land on the default dashboard → else the first one. Keep the
  // selection valid as the list changes.
  useEffect(() => {
    if (dashboardItems.length === 0) return
    const valid = (id: number | null) => id != null && dashboardItems.some((d) => d.id === id)
    if (valid(selectedId)) return
    setSelectedId(valid(defaultId) ? defaultId : dashboardItems[0].id)
  }, [dashboardItems, selectedId, defaultId])

  const selectedDashboard = useDashboard(selectedId)

  const layouts = useDashboardLayouts(
    selectedId != null ? { idDashboard: selectedId, page: 0, size: 500 } : { page: 0, size: 0 }
  )
  const layoutRows = layouts.list.data?.data ?? []

  const visualizationIds = useMemo(
    () => Array.from(new Set(layoutRows.map((r) => r.idVisualization))),
    [layoutRows]
  )

  const vizQueries = useQueries({
    queries: visualizationIds.map((id) => ({
      queryKey: VISUALIZATIONS_QUERY_KEYS.one(id),
      queryFn: () => createVisualizationsService().getVisualization(id),
      enabled: id > 0,
    })),
  })

  const visualizationsById = useMemo(() => {
    const map = new Map<number, Visualization>()
    for (const q of vizQueries) {
      if (q.data) map.set(q.data.id, q.data)
    }
    return map
  }, [vizQueries])

  const initialItems = useMemo(() => toGridItems(layoutRows), [layoutRows])
  const editor = useDashboardEditor(initialItems)

  const layoutsById = useMemo(() => {
    const m = new Map<string, number>()
    for (const dv of layoutRows) m.set(String(dv.id), dv.id)
    return m
  }, [layoutRows])

  const excludedVizIds = useMemo(() => {
    const s = new Set<number>()
    for (const dv of layoutRows) s.add(dv.idVisualization)
    return s
  }, [layoutRows])

  const handleCreateDashboard = (data: { name: string; description?: string }) => {
    dashboards.createDashboard.mutate(
      { name: data.name, description: data.description, config: '' },
      {
        onSuccess: (created) => {
          toast.success(t('dashboards.toast.created'))
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

  const handleAddVisualizations = async (vizs: Visualization[]) => {
    if (selectedId == null || vizs.length === 0) return
    try {
      // Drop new widgets below the current layout, one per row.
      let y = nextRow(initialItems)
      for (let i = 0; i < vizs.length; i++) {
        await layouts.createLayout.mutateAsync({
          idDashboard: selectedId,
          idVisualization: vizs[i].id,
          layout: serializeLayout({
            x: 0,
            y,
            w: DEFAULT_WIDGET_LAYOUT.w,
            h: DEFAULT_WIDGET_LAYOUT.h,
          }),
        })
        y += DEFAULT_WIDGET_LAYOUT.h
      }
      toast.success(
        vizs.length === 1
          ? t('dashboards.toast.widgetAdded')
          : t('dashboards.toast.widgetsAdded', { count: vizs.length })
      )
      setAddOpen(false)
    } catch (err) {
      toast.error((err as Error)?.message ?? t('dashboards.toast.widgetAddFailed'))
    }
  }

  const handleSave = async () => {
    if (!editor.dirty || selectedId == null) return
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
        const layoutId = layoutsById.get(item.i)
        if (layoutId == null) continue
        const dv = layoutRows.find((r) => r.id === layoutId)
        if (!dv) continue
        await layouts.updateLayout.mutateAsync({
          id: dv.id,
          idDashboard: dv.idDashboard,
          idVisualization: dv.idVisualization,
          layout: serializeLayout({ x: item.x, y: item.y, w: item.w, h: item.h }),
        })
      }

      for (const removeId of editor.pendingRemovals) {
        await layouts.deleteLayout.mutateAsync(removeId)
      }

      editor.commit()
      toast.success(t('dashboards.toast.layoutSaved'))
    } catch (err) {
      toast.error((err as Error).message ?? t('dashboards.toast.layoutSaveFailed'))
    }
  }

  const gridItems: GridLayoutItem[] = editor.editing ? editor.working : initialItems
  const saving =
    layouts.updateLayout.isPending ||
    layouts.deleteLayout.isPending ||
    layouts.createLayout.isPending

  return (
    <div className="mx-auto flex h-full w-full max-w-[1800px] flex-col gap-4 px-6 pb-6 pt-3">
      {!noDashboards && (
        <DashboardTabsBar
          dashboards={dashboardItems}
          selectedId={selectedId}
          defaultId={defaultId}
          onSelect={setSelectedId}
          onSetDefault={setDefaultDashboard}
          onCreate={() => setFormOpen({ mode: 'create', target: null })}
          onRename={(d) => setFormOpen({ mode: 'rename', target: d })}
          onDelete={(d) => setPendingDelete(d)}
          onEdit={
            selectedDashboard.data && !editor.editing && !selectedDashboard.data.systemOwner
              ? editor.enter
              : undefined
          }
          right={
            editor.editing ? (
              <DashboardEditorBar
                dirty={editor.dirty}
                saving={saving}
                onAddWidget={() => setAddOpen(true)}
                onSave={handleSave}
                onDiscard={editor.discard}
              />
            ) : selectedId != null ? (
              <DashboardTimeRange value={time} onChange={setTime} />
            ) : null
          }
        />
      )}

      <div className="min-h-0 flex-1 overflow-y-auto rounded-lg border border-border bg-background/30 p-2">
        {noDashboards && !dashboards.list.isLoading ? (
          <div className="flex h-full min-h-[320px] flex-col items-center justify-center gap-3 text-center">
            <LayoutDashboard size={32} className="text-muted-foreground/40" />
            <div>
              <p className="text-sm font-medium">{t('dashboards.empty.title')}</p>
              <p className="mt-1 text-xs text-muted-foreground">{t('dashboards.empty.body')}</p>
            </div>
            <Button size="sm" onClick={() => setFormOpen({ mode: 'create', target: null })}>
              <Plus size={14} className="mr-1" /> {t('dashboards.empty.cta')}
            </Button>
          </div>
        ) : selectedId != null && layouts.list.isLoading ? (
          <div className="flex h-full min-h-[300px] items-center justify-center gap-2 text-sm text-muted-foreground">
            <Loader2 size={16} className="animate-spin" />
            {t('dashboards.page.loading')}
          </div>
        ) : selectedId != null && !layouts.list.isLoading ? (
          <DashboardGrid
            items={gridItems}
            layouts={layoutRows}
            visualizationsById={visualizationsById}
            time={time}
            editing={editor.editing}
            onLayoutChange={editor.replace}
            onRemoveItem={editor.remove}
          />
        ) : null}
      </div>

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

      <AddVisualizationDrawer
        open={addOpen}
        excludedIds={excludedVizIds}
        busy={layouts.createLayout.isPending}
        onClose={() => setAddOpen(false)}
        onAdd={handleAddVisualizations}
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
    </div>
  )
}

export default DashboardPage
