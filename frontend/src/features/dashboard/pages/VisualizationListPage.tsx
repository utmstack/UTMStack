import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Loader2, Plus, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { ConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import {
  useVisualizationMutations,
  useVisualizations,
} from '@/features/dashboard/hooks/useVisualizations'
import { CHART_TYPES } from '@/features/dashboard/constants'
import { parseBuilderConfig } from '@/features/dashboard/utils/builder-config'
import type { ChartTypeId, Visualization } from '@/features/dashboard/types'

const LIST_SIZE = 500

interface FilterState {
  chartType: ChartTypeId | ''
  source: string
  createdFrom: string
  createdTo: string
}

const INITIAL_FILTERS: FilterState = {
  chartType: '',
  source: '',
  createdFrom: '',
  createdTo: '',
}

export function VisualizationListPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [filters, setFilters] = useState<FilterState>(INITIAL_FILTERS)
  const [pendingDelete, setPendingDelete] = useState<Visualization | null>(null)

  const query = useVisualizations({
    name: search || undefined,
    page: 0,
    size: LIST_SIZE,
  })
  const rawItems = query.data?.data ?? []
  const total = query.data?.total ?? rawItems.length

  const enriched = useMemo(
    () =>
      rawItems.map((viz) => {
        const builder = parseBuilderConfig(viz.config ?? null).builder
        return {
          viz,
          chartType: builder?.chartType ?? null,
          source: builder?.indexPattern?.trim() || null,
        }
      }),
    [rawItems]
  )

  const sourceOptions = useMemo(() => {
    const set = new Set<string>()
    for (const e of enriched) {
      if (e.source) set.add(e.source)
    }
    return Array.from(set).sort()
  }, [enriched])

  const createdFromMs = filters.createdFrom ? Date.parse(filters.createdFrom) : null
  const createdToMs = filters.createdTo
    ? Date.parse(filters.createdTo) + 24 * 60 * 60 * 1000 - 1
    : null

  const filtered = useMemo(
    () =>
      enriched.filter(({ viz, chartType, source }) => {
        if (filters.chartType && chartType !== filters.chartType) return false
        if (filters.source && source !== filters.source) return false
        if (createdFromMs != null) {
          const t0 = viz.createdDate ? Date.parse(viz.createdDate) : NaN
          if (!Number.isFinite(t0) || t0 < createdFromMs) return false
        }
        if (createdToMs != null) {
          const t0 = viz.createdDate ? Date.parse(viz.createdDate) : NaN
          if (!Number.isFinite(t0) || t0 > createdToMs) return false
        }
        return true
      }),
    [enriched, filters, createdFromMs, createdToMs]
  )

  const filtersActive =
    filters.chartType !== '' ||
    filters.source !== '' ||
    filters.createdFrom !== '' ||
    filters.createdTo !== ''

  const { deleteVisualization } = useVisualizationMutations()

  const confirmDelete = () => {
    if (!pendingDelete) return
    const target = pendingDelete
    deleteVisualization.mutate(target.id, {
      onSuccess: () => {
        toast.success(t('dashboards.toast.visualizationDeleted'))
        setPendingDelete(null)
      },
      onError: (err) =>
        toast.error(err.message ?? t('dashboards.toast.visualizationDeleteFailed')),
    })
  }

  return (
    <div className="mx-auto flex h-full w-full max-w-[1400px] flex-col gap-4 px-6 py-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-xl font-semibold">{t('dashboards.visualizationList.title')}</h1>
          <p className="text-sm text-muted-foreground">
            {filtersActive
              ? t('dashboards.visualizationList.subtitleFiltered', {
                  shown: filtered.length,
                  total,
                })
              : t('dashboards.visualizationList.subtitle', { count: total })}
          </p>
        </div>
        <Button asChild size="sm">
          <Link to="/dashboards/visualizations/new">
            <Plus size={14} className="mr-1" />
            {t('dashboards.visualizationList.create')}
          </Link>
        </Button>
      </div>

      <div className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
        <div className="flex flex-wrap items-end gap-2">
          <div className="min-w-[200px] flex-1">
            <Label>{t('dashboards.visualizationList.filters.search')}</Label>
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t('dashboards.visualizationList.searchPlaceholder') ?? ''}
              className="h-9"
            />
          </div>

          <div className="w-44">
            <Label>{t('dashboards.visualizationList.filters.chartType')}</Label>
            <Select
              value={filters.chartType}
              onChange={(v) =>
                setFilters((f) => ({ ...f, chartType: v as ChartTypeId | '' }))
              }
            >
              <option value="">
                {t('dashboards.visualizationList.filters.allChartTypes')}
              </option>
              {CHART_TYPES.map((c) => (
                <option key={c.id} value={c.id}>
                  {t(`dashboards.editor.chartTypes.${c.id}.label`)}
                </option>
              ))}
            </Select>
          </div>

          <div className="w-44">
            <Label>{t('dashboards.visualizationList.filters.source')}</Label>
            <Select
              value={filters.source}
              onChange={(v) => setFilters((f) => ({ ...f, source: v }))}
            >
              <option value="">
                {t('dashboards.visualizationList.filters.allSources')}
              </option>
              {sourceOptions.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </Select>
          </div>

          <div className="w-40">
            <Label>{t('dashboards.visualizationList.filters.createdFrom')}</Label>
            <Input
              type="date"
              value={filters.createdFrom}
              onChange={(e) =>
                setFilters((f) => ({ ...f, createdFrom: e.target.value }))
              }
              className="h-9"
            />
          </div>

          <div className="w-40">
            <Label>{t('dashboards.visualizationList.filters.createdTo')}</Label>
            <Input
              type="date"
              value={filters.createdTo}
              onChange={(e) =>
                setFilters((f) => ({ ...f, createdTo: e.target.value }))
              }
              className="h-9"
            />
          </div>

          {filtersActive && (
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setFilters(INITIAL_FILTERS)}
            >
              <X size={14} className="mr-1" />
              {t('dashboards.visualizationList.filters.clear')}
            </Button>
          )}
        </div>

        <div className="min-h-0 flex-1 overflow-x-auto">
          {query.isLoading && (
            <div className="flex items-center justify-center gap-2 py-10 text-xs text-muted-foreground">
              <Loader2 size={14} className="animate-spin" />
              {t('dashboards.visualizationList.loading')}
            </div>
          )}
          {!query.isLoading && filtered.length === 0 && (
            <div className="py-10 text-center text-xs text-muted-foreground">
              {filtersActive
                ? t('dashboards.visualizationList.emptyFiltered')
                : t('dashboards.visualizationList.empty')}
            </div>
          )}
          {!query.isLoading && filtered.length > 0 && (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left text-xs uppercase tracking-wide text-muted-foreground">
                  <th className="px-3 py-2 font-medium">
                    {t('dashboards.visualizationList.col.name')}
                  </th>
                  <th className="px-3 py-2 font-medium">
                    {t('dashboards.visualizationList.col.chartType')}
                  </th>
                  <th className="px-3 py-2 font-medium">
                    {t('dashboards.visualizationList.col.source')}
                  </th>
                  <th className="px-3 py-2 font-medium">
                    {t('dashboards.visualizationList.col.description')}
                  </th>
                  <th className="px-3 py-2 font-medium">
                    {t('dashboards.visualizationList.col.created')}
                  </th>
                  <th className="px-3 py-2 font-medium">
                    {t('dashboards.visualizationList.col.modified')}
                  </th>
                  <th className="px-3 py-2 font-medium" />
                </tr>
              </thead>
              <tbody>
                {filtered.map(({ viz: v, chartType, source }) => (
                  <tr
                    key={v.id}
                    className="border-b border-border/60 last:border-0 hover:bg-muted/40"
                  >
                    <td className="px-3 py-2 font-medium">
                      <Link
                        to={`/dashboards/visualizations/${v.id}`}
                        className="text-foreground hover:text-primary hover:underline"
                      >
                        {v.name}
                      </Link>
                    </td>
                    <td className="px-3 py-2 text-xs text-muted-foreground">
                      {chartType ? t(`dashboards.editor.chartTypes.${chartType}.label`) : '—'}
                    </td>
                    <td className="px-3 py-2 text-xs text-muted-foreground">{source ?? '—'}</td>
                    <td className="px-3 py-2 text-muted-foreground">{v.description || '—'}</td>
                    <td className="px-3 py-2 text-xs text-muted-foreground">
                      {v.createdDate ? new Date(v.createdDate).toLocaleString() : '—'}
                    </td>
                    <td className="px-3 py-2 text-xs text-muted-foreground">
                      {v.modifiedDate ? new Date(v.modifiedDate).toLocaleString() : '—'}
                    </td>
                    <td className="px-3 py-2 text-right">
                      {!v.systemOwner && (
                        <button
                          type="button"
                          onClick={() => setPendingDelete(v)}
                          className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive disabled:opacity-50"
                          aria-label={t('dashboards.visualizationList.delete') ?? 'Delete'}
                        >
                          <Trash2 size={14} />
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>

      <ConfirmDialog
        open={pendingDelete != null}
        title={t('dashboards.visualizationList.deleteTitle')}
        body={t('dashboards.visualizationList.confirmDelete', {
          name: pendingDelete?.name ?? '',
        })}
        confirmLabel={t('dashboards.visualizationList.delete') ?? undefined}
        danger
        busy={deleteVisualization.isPending}
        onClose={() => setPendingDelete(null)}
        onConfirm={confirmDelete}
      />
    </div>
  )
}

function Label({ children }: { children: React.ReactNode }) {
  return (
    <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
      {children}
    </label>
  )
}

function Select({
  value,
  onChange,
  children,
}: {
  value: string
  onChange: (next: string) => void
  children: React.ReactNode
}) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
    >
      {children}
    </select>
  )
}

export default VisualizationListPage
