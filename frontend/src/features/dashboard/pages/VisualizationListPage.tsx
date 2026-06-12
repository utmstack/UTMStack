import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Calendar, Loader2, Plus, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { ConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import {
  DateRangePickerDialog,
  EMPTY_RANGE,
  formatRange,
  type DateRangeValue,
} from '@/shared/components/ui/date-range-picker'
import {
  useVisualizationMutations,
  useVisualizations,
} from '@/features/dashboard/hooks/useVisualizations'
import { CHART_TYPES } from '@/features/dashboard/constants'
import { parseBuilderConfig } from '@/features/dashboard/utils/builder-config'
import type { ChartTypeId, Visualization } from '@/features/dashboard/types'

const LIST_SIZE = 500

type DatePresetId = 'all' | 'today' | '7d' | '30d' | '1y' | 'custom'

interface DateFilter {
  preset: DatePresetId
  customRange: DateRangeValue
}

const INITIAL_DATE: DateFilter = { preset: 'all', customRange: EMPTY_RANGE }

interface FilterState {
  chartType: ChartTypeId | ''
  source: string
  created: DateFilter
  modified: DateFilter
}

const INITIAL_FILTERS: FilterState = {
  chartType: '',
  source: '',
  created: INITIAL_DATE,
  modified: INITIAL_DATE,
}

function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate(), 0, 0, 0, 0)
}
function endOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate(), 23, 59, 59, 999)
}

function resolveDateFilter(df: DateFilter): { fromMs: number | null; toMs: number | null } {
  const now = new Date()
  switch (df.preset) {
    case 'today':
      return { fromMs: startOfDay(now).getTime(), toMs: endOfDay(now).getTime() }
    case '7d': {
      const from = new Date(now)
      from.setDate(from.getDate() - 6)
      return { fromMs: startOfDay(from).getTime(), toMs: endOfDay(now).getTime() }
    }
    case '30d': {
      const from = new Date(now)
      from.setDate(from.getDate() - 29)
      return { fromMs: startOfDay(from).getTime(), toMs: endOfDay(now).getTime() }
    }
    case '1y': {
      const from = new Date(now)
      from.setFullYear(from.getFullYear() - 1)
      from.setDate(from.getDate() + 1)
      return { fromMs: startOfDay(from).getTime(), toMs: endOfDay(now).getTime() }
    }
    case 'custom':
      return {
        fromMs: df.customRange.from ? startOfDay(df.customRange.from).getTime() : null,
        toMs: df.customRange.to ? endOfDay(df.customRange.to).getTime() : null,
      }
    case 'all':
    default:
      return { fromMs: null, toMs: null }
  }
}

function inRange(iso: string | undefined, fromMs: number | null, toMs: number | null): boolean {
  if (fromMs == null && toMs == null) return true
  if (!iso) return false
  const t = Date.parse(iso)
  if (!Number.isFinite(t)) return false
  if (fromMs != null && t < fromMs) return false
  if (toMs != null && t > toMs) return false
  return true
}

function dateFilterActive(df: DateFilter): boolean {
  if (df.preset === 'all') return false
  if (df.preset === 'custom') return !!(df.customRange.from || df.customRange.to)
  return true
}

type ModalTarget = 'created' | 'modified' | null

export function VisualizationListPage() {
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [filters, setFilters] = useState<FilterState>(INITIAL_FILTERS)
  const [modalFor, setModalFor] = useState<ModalTarget>(null)
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

  const createdResolved = useMemo(() => resolveDateFilter(filters.created), [filters.created])
  const modifiedResolved = useMemo(() => resolveDateFilter(filters.modified), [filters.modified])

  const filtered = useMemo(
    () =>
      enriched.filter(({ viz, chartType, source }) => {
        if (filters.chartType && chartType !== filters.chartType) return false
        if (filters.source && source !== filters.source) return false
        if (createdResolved.fromMs != null || createdResolved.toMs != null) {
          if (!inRange(viz.createdDate, createdResolved.fromMs, createdResolved.toMs)) return false
        }
        if (modifiedResolved.fromMs != null || modifiedResolved.toMs != null) {
          if (!inRange(viz.modifiedDate, modifiedResolved.fromMs, modifiedResolved.toMs))
            return false
        }
        return true
      }),
    [enriched, filters.chartType, filters.source, createdResolved, modifiedResolved]
  )

  const filtersActive =
    filters.chartType !== '' ||
    filters.source !== '' ||
    dateFilterActive(filters.created) ||
    dateFilterActive(filters.modified)

  const setDateField = (target: 'created' | 'modified', next: DateFilter) => {
    setFilters((f) => ({ ...f, [target]: next }))
  }

  const onPresetChange = (target: 'created' | 'modified', next: DatePresetId) => {
    setDateField(target, {
      preset: next,
      customRange: next === 'custom' ? filters[target].customRange : EMPTY_RANGE,
    })
    if (next === 'custom') setModalFor(target)
  }

  const onRangeConfirm = (range: DateRangeValue) => {
    if (!modalFor) return
    setDateField(modalFor, { preset: 'custom', customRange: range })
    setModalFor(null)
  }

  const onModalClose = () => {
    if (!modalFor) return
    const current = filters[modalFor]
    if (
      current.preset === 'custom' &&
      !current.customRange.from &&
      !current.customRange.to
    ) {
      setDateField(modalFor, INITIAL_DATE)
    }
    setModalFor(null)
  }

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

          <DateFilterControl
            label={t('dashboards.visualizationList.filters.createdAt')}
            value={filters.created}
            onPresetChange={(p) => onPresetChange('created', p)}
            onOpenModal={() => setModalFor('created')}
          />

          <DateFilterControl
            label={t('dashboards.visualizationList.filters.modifiedAt')}
            value={filters.modified}
            onPresetChange={(p) => onPresetChange('modified', p)}
            onOpenModal={() => setModalFor('modified')}
          />

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

      <DateRangePickerDialog
        open={modalFor !== null}
        value={modalFor ? filters[modalFor].customRange : EMPTY_RANGE}
        title={
          modalFor === 'created'
            ? t('dashboards.visualizationList.filters.createdRangeTitle')
            : modalFor === 'modified'
              ? t('dashboards.visualizationList.filters.modifiedRangeTitle')
              : undefined
        }
        onClose={onModalClose}
        onConfirm={onRangeConfirm}
      />

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

function DateFilterControl({
  label,
  value,
  onPresetChange,
  onOpenModal,
}: {
  label: string
  value: DateFilter
  onPresetChange: (next: DatePresetId) => void
  onOpenModal: () => void
}) {
  const { t } = useTranslation()
  return (
    <div className="w-52">
      <Label>{label}</Label>
      <Select
        value={value.preset}
        onChange={(v) => onPresetChange(v as DatePresetId)}
      >
        <option value="all">
          {t('dashboards.visualizationList.filters.datePresets.all')}
        </option>
        <option value="today">
          {t('dashboards.visualizationList.filters.datePresets.today')}
        </option>
        <option value="7d">
          {t('dashboards.visualizationList.filters.datePresets.last7')}
        </option>
        <option value="30d">
          {t('dashboards.visualizationList.filters.datePresets.last30')}
        </option>
        <option value="1y">
          {t('dashboards.visualizationList.filters.datePresets.lastYear')}
        </option>
        <option value="custom">
          {t('dashboards.visualizationList.filters.datePresets.custom')}
        </option>
      </Select>
      {value.preset === 'custom' && (
        <button
          type="button"
          onClick={onOpenModal}
          className="mt-1 inline-flex h-7 w-full items-center gap-1.5 rounded-md border border-border bg-background px-2 text-xs text-foreground/90 hover:bg-muted/60"
        >
          <Calendar size={12} className="text-muted-foreground" />
          <span className="truncate">
            {formatRange(value.customRange) ||
              t('dashboards.visualizationList.filters.pickRange')}
          </span>
        </button>
      )}
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
