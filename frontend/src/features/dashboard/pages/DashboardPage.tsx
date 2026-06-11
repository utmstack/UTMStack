import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, Pencil } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { presetRange, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { useDashboard, useDashboards } from '@/features/dashboard/hooks/useDashboards'
import { useDashboardLayouts } from '@/features/dashboard/hooks/useDashboardLayouts'
import { createVisualizationsService } from '@/features/dashboard/service/visualizations.service'
import { useDashboardEditor } from '@/features/dashboard/hooks/useDashboardEditor'
import { DashboardList } from '@/features/dashboard/components/DashboardList'
import { DashboardGrid } from '@/features/dashboard/components/DashboardGrid'
import { DashboardEditorBar } from '@/features/dashboard/components/DashboardEditorBar'
import { DashboardFormDialog } from '@/features/dashboard/components/DashboardFormDialog'
import { DashboardTimeRange } from '@/features/dashboard/components/DashboardTimeRange'
import { AddVisualizationDrawer } from '@/features/dashboard/components/AddVisualizationDrawer'
import {
  DEFAULT_PAGE_SIZE,
  DEFAULT_WIDGET_LAYOUT,
} from '@/features/dashboard/constants'
import { fromGridItem, serializeLayout, toGridItems } from '@/features/dashboard/utils/layout'
import type { Dashboard, GridLayoutItem, Visualization } from '@/features/dashboard/types'
import { useQueries } from '@tanstack/react-query'
import { VISUALIZATIONS_QUERY_KEYS } from '@/features/dashboard/hooks/useVisualizations'

export function DashboardPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [time, setTime] = useState<TimeRange>(() => presetRange('24h'))
  const [formOpen, setFormOpen] = useState<null | { mode: 'create' | 'rename'; target: Dashboard | null }>(
    null
  )
  const [addOpen, setAddOpen] = useState(false)

  const dashboards = useDashboards({ name: search || undefined, page: 0, size: DEFAULT_PAGE_SIZE })
  const dashboardItems = dashboards.list.data?.data ?? []

  useEffect(() => {
    if (selectedId == null && dashboardItems.length > 0) {
      setSelectedId(dashboardItems[0].id)
    }
    if (selectedId != null && !dashboardItems.some((d) => d.id === selectedId)) {
      setSelectedId(dashboardItems[0]?.id ?? null)
    }
  }, [dashboardItems, selectedId])

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

  const handleDeleteDashboard = (d: Dashboard) => {
    if (!window.confirm(t('dashboards.confirm.delete', { name: d.name }))) return
    dashboards.deleteDashboard.mutate(d.id, {
      onSuccess: () => {
        toast.success(t('dashboards.toast.deleted'))
        if (selectedId === d.id) setSelectedId(null)
      },
      onError: (err) => toast.error(err.message ?? t('dashboards.toast.deleteFailed')),
    })
  }

  const handleAddVisualization = (viz: Visualization) => {
    if (selectedId == null) return
    layouts.createLayout.mutate(
      {
        idDashboard: selectedId,
        idVisualization: viz.id,
        layout: serializeLayout({ ...DEFAULT_WIDGET_LAYOUT }),
      },
      {
        onSuccess: () => {
          toast.success(t('dashboards.toast.widgetAdded'))
          setAddOpen(false)
        },
        onError: (err) => toast.error(err.message ?? t('dashboards.toast.widgetAddFailed')),
      }
    )
  }

  const handleSave = async () => {
    if (!editor.dirty || selectedId == null) return
    try {
      const baselineMap = new Map(editor.baseline.map((it) => [it.i, it]))
      const movedItems = editor.working.filter((it) => {
        const base = baselineMap.get(it.i)
        return !base || base.x !== it.x || base.y !== it.y || base.w !== it.w || base.h !== it.h
      })

      for (let i = 0; i < movedItems.length; i++) {
        const item = movedItems[i]
        const layoutId = layoutsById.get(item.i)
        if (layoutId == null) continue
        const dv = layoutRows.find((r) => r.id === layoutId)
        if (!dv) continue
        await layouts.updateLayout.mutateAsync({
          id: dv.id,
          idDashboard: dv.idDashboard,
          idVisualization: dv.idVisualization,
          layout: serializeLayout(fromGridItem(item, i)),
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
    <div className="mx-auto flex h-full w-full max-w-[1800px] flex-col gap-4 px-6 py-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold">{t('dashboards.title')}</h1>
          <p className="text-sm text-muted-foreground">{t('dashboards.subtitle')}</p>
        </div>
        <div className="flex items-center gap-2">
          <DashboardTimeRange value={time} onChange={setTime} />
          {selectedDashboard.data && !editor.editing && !selectedDashboard.data.systemOwner && (
            <Button variant="outline" size="sm" onClick={editor.enter}>
              <Pencil size={14} className="mr-1" />
              {t('dashboards.actions.edit')}
            </Button>
          )}
        </div>
      </div>

      <div className="flex min-h-0 flex-1 gap-4">
        <DashboardList
          dashboards={dashboardItems}
          selectedId={selectedId}
          search={search}
          loading={dashboards.list.isLoading}
          onSearchChange={setSearch}
          onSelect={setSelectedId}
          onCreate={() => setFormOpen({ mode: 'create', target: null })}
          onRename={(d) => setFormOpen({ mode: 'rename', target: d })}
          onDelete={handleDeleteDashboard}
        />

        <div className="flex min-w-0 flex-1 flex-col gap-3">
          {editor.editing && (
            <DashboardEditorBar
              dirty={editor.dirty}
              saving={saving}
              onAddWidget={() => setAddOpen(true)}
              onSave={handleSave}
              onDiscard={editor.discard}
            />
          )}

          <div className="min-h-0 flex-1 overflow-y-auto rounded-lg border border-border bg-background/30 p-2">
            {selectedId == null && !dashboards.list.isLoading && (
              <div className="flex h-full min-h-[300px] items-center justify-center text-sm text-muted-foreground">
                {t('dashboards.page.noSelection')}
              </div>
            )}
            {selectedId != null && layouts.list.isLoading && (
              <div className="flex h-full min-h-[300px] items-center justify-center gap-2 text-sm text-muted-foreground">
                <Loader2 size={16} className="animate-spin" />
                {t('dashboards.page.loading')}
              </div>
            )}
            {selectedId != null && !layouts.list.isLoading && (
              <DashboardGrid
                items={gridItems}
                layouts={layoutRows}
                visualizationsById={visualizationsById}
                time={time}
                editing={editor.editing}
                onLayoutChange={editor.replace}
                onRemoveItem={editor.remove}
              />
            )}
          </div>
        </div>
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
        onClose={() => setAddOpen(false)}
        onPick={handleAddVisualization}
      />
    </div>
  )
}

export default DashboardPage
