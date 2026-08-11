import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Loader2, Pencil } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { DatasetSelect } from '@/features/dashboard/components/editor/DatasetSelect'
import { FilterBuilder } from '@/features/dashboard/components/editor/FilterBuilder'
import { DimensionPicker } from '@/features/dashboard/components/editor/DimensionPicker'
import { ColumnsPicker } from '@/features/dashboard/components/editor/ColumnsPicker'
import { ChartPreviewPanel } from '@/features/dashboard/components/editor/ChartPreviewPanel'
import { ChartTypeModal } from '@/features/dashboard/components/editor/ChartTypeModal'
import { useAggregatableFields } from '@/features/dashboard/hooks/useAggregatableFields'
import { useVisualizationMutations } from '@/features/dashboard/hooks/useVisualizations'
import { DEFAULT_WIDGET_LAYOUT, INTERVALS, getChartTypeMeta } from '@/features/dashboard/constants'
import { serializeLayout } from '@/features/dashboard/utils/layout'
import { builderToSpec, parseSpec, specChartFor, specToBuilder } from '@/features/dashboard/utils/spec'
import {
  makeInitialBuilder,
  parseBuilderConfig,
  serializeBuilderConfig,
} from '@/features/dashboard/utils/builder-config'
import type {
  BuilderState,
  ChartTypeId,
  IndexProperty,
  IntervalId,
  Visualization,
} from '@/features/dashboard/types'

export interface VisualizationEditorProps {
  initial: Visualization | null
  initialChartType?: ChartTypeId
  // The dashboard this visualization belongs to. For edits this always matches
  // `initial.dashboardId`; for a brand-new visualization it comes from the route.
  dashboardId: string
  // Grid position/size for a new widget (from DashboardPage's "Add widget"
  // seeding). Ignored on edit — the existing layout is preserved untouched;
  // repositioning happens by dragging on the dashboard grid, not here.
  initialLayout?: string
}

export function VisualizationEditor({
  initial,
  initialChartType,
  dashboardId,
  initialLayout,
}: VisualizationEditorProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const isEdit = initial != null

  const initialParsed = useMemo(
    () => parseBuilderConfig(initial?.config ?? null),
    [initial?.config]
  )

  const [option, setOption] = useState<Record<string, unknown>>(() => {
    if (initial && initialParsed.builder && Object.keys(initialParsed.option).length > 0) {
      return initialParsed.option
    }
    const startChartType =
      initialParsed.builder?.chartType ?? initialChartType ?? 'bar'
    return getChartTypeMeta(startChartType).defaultConfig
  })

  // The widget's saved spec is the question; the builder is how it is edited.
  // The chart type lives in the config next to it, because which chart draws an
  // answer is not part of the question.
  const [builder, setBuilder] = useState<BuilderState>(() => {
    const chartType = initialParsed.builder?.chartType ?? initialChartType ?? 'bar'
    const base = { ...makeInitialBuilder(), ...(initialParsed.builder ?? {}), chartType }
    const spec = parseSpec(initial?.spec)
    return spec ? { ...base, ...specToBuilder(spec) } : base
  })

  const [chartTypeOpen, setChartTypeOpen] = useState(false)

  const {
    fields,
    groupableFields: groupable,
    isLoading: fieldsLoading,
  } = useAggregatableFields(builder.dataset)

  const update = (updater: (b: BuilderState) => BuilderState) => setBuilder(updater)

  const setChartType = (chartType: ChartTypeId) => {
    update((b) => ({ ...b, chartType }))
    if (!builder.configTouched) {
      setOption(getChartTypeMeta(chartType).defaultConfig)
    }
  }

  const meta = getChartTypeMeta(builder.chartType)
  const chart = specChartFor(builder.chartType, builder.breakdown)
  const spec = useMemo(() => builderToSpec(builder), [builder])

  const { createVisualization, updateVisualization } = useVisualizationMutations()
  const busy = createVisualization.isPending || updateVisualization.isPending

  // What the backend will refuse: a widget with no dataset, or a breakdown
  // chart with nothing to break down by.
  const ready =
    builder.dataset.trim().length > 0 && (chart !== 'category' || !!builder.dimension)

  const handleSave = () => {
    if (!ready) {
      toast.error(t('dashboards.editor.toast.notReady'))
      return
    }
    if (busy) return

    const specJson = JSON.stringify(spec)
    const configJson = serializeBuilderConfig(option, builder)
    const backToDashboard = () => navigate('/dashboards/list', { state: { selectDashboardId: dashboardId } })

    if (isEdit && initial) {
      updateVisualization.mutate(
        {
          id: initial.id,
          dashboardId: initial.dashboardId,
          spec: specJson,
          config: configJson,
          layout: initial.layout,
        },
        {
          onSuccess: () => {
            toast.success(t('dashboards.toast.visualizationUpdated'))
            backToDashboard()
          },
          onError: (err) =>
            toast.error(err.message ?? t('dashboards.toast.visualizationUpdateFailed')),
        }
      )
    } else {
      createVisualization.mutate(
        {
          dashboardId,
          spec: specJson,
          config: configJson,
          layout: initialLayout ?? serializeLayout({ x: 0, y: 0, w: DEFAULT_WIDGET_LAYOUT.w, h: DEFAULT_WIDGET_LAYOUT.h }),
        },
        {
          onSuccess: () => {
            toast.success(t('dashboards.toast.visualizationCreated'))
            backToDashboard()
          },
          onError: (err) =>
            toast.error(err.message ?? t('dashboards.toast.visualizationCreateFailed')),
        }
      )
    }
  }

  return (
    <div className="flex h-full w-full flex-col gap-4 px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <h1 className="text-base font-semibold">
            {isEdit
              ? t('dashboards.editor.editTitle')
              : t('dashboards.editor.newTitle')}
          </h1>
          <ChartTypeChip
            chartType={builder.chartType}
            onChange={() => setChartTypeOpen(true)}
          />
        </div>
        <div className="flex items-center gap-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => navigate('/dashboards/list', { state: { selectDashboardId: dashboardId } })}
            disabled={busy}
          >
            {t('dashboards.form.cancel')}
          </Button>
          <Button type="button" size="sm" onClick={handleSave} disabled={!ready || busy}>
            {busy && <Loader2 size={14} className="mr-1 animate-spin" />}
            {t('dashboards.editor.saveVisualization')}
          </Button>
        </div>
      </header>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <div className="flex flex-col gap-4">
          <QuestionPanel
            builder={builder}
            chart={chart}
            fields={fields}
            groupableFields={groupable}
            fieldsLoading={fieldsLoading}
            onChange={update}
          />
        </div>

        <div className="lg:sticky lg:top-4 lg:self-start">
          <section className="rounded-lg border border-border bg-card p-4">
            <ChartPreviewPanel
              spec={ready ? spec : null}
              option={option}
              renderer={meta.renderer}
            />
          </section>
        </div>
      </div>

      <ChartTypeModal
        open={chartTypeOpen}
        initial={builder.chartType}
        title={t('dashboards.editor.chartTypeModal.swapTitle') ?? undefined}
        confirmLabel={t('dashboards.editor.chartTypeModal.swapConfirm') ?? undefined}
        onConfirm={(next) => {
          setChartType(next)
          setChartTypeOpen(false)
        }}
        onClose={() => setChartTypeOpen(false)}
      />
    </div>
  )
}

function ChartTypeChip({
  chartType,
  onChange,
}: {
  chartType: ChartTypeId
  onChange: () => void
}) {
  const { t } = useTranslation()
  return (
    <button
      type="button"
      onClick={onChange}
      className="inline-flex items-center gap-1.5 self-start rounded-full border border-border bg-muted/50 px-2.5 py-1 text-xs text-foreground/80 hover:bg-muted"
    >
      <span className="text-muted-foreground">
        {t('dashboards.editor.chartTypeChip.label')}
      </span>
      <span className="font-medium">
        {t(`dashboards.editor.chartTypes.${chartType}.label`)}
      </span>
      <Pencil size={11} className="text-muted-foreground" />
    </button>
  )
}

/**
 * The question the widget asks: which records, narrowed how, counted over what.
 * There is no measure to pick — the store counts records and nothing else.
 */
function QuestionPanel({
  builder,
  chart,
  fields,
  groupableFields,
  fieldsLoading,
  onChange,
}: {
  builder: BuilderState
  chart: ReturnType<typeof specChartFor>
  fields: IndexProperty[]
  groupableFields: IndexProperty[]
  fieldsLoading: boolean
  onChange: (updater: (b: BuilderState) => BuilderState) => void
}) {
  const { t } = useTranslation()
  const plots = chart === 'time' || chart === 'category'

  return (
    <>
      <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
        <SectionTitle>{t('dashboards.editor.dataSource.title')}</SectionTitle>
        <div>
          <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
            {t('dashboards.editor.dataSource.dataset')}
          </label>
          <DatasetSelect
            value={builder.dataset}
            onChange={(dataset) => onChange((b) => ({ ...b, dataset }))}
          />
          {!builder.dataset && (
            <p className="mt-1 text-[10px] text-muted-foreground">
              {t('dashboards.editor.dataSource.noDatasetHint')}
            </p>
          )}
        </div>
      </section>

      <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
        <SectionTitle>{t('dashboards.editor.filters.title')}</SectionTitle>
        <FilterBuilder
          filters={builder.filters}
          fields={fields ?? []}
          loading={fieldsLoading}
          onChange={(next) => onChange((b) => ({ ...b, filters: next }))}
        />
      </section>

      {plots && (
        <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
          <SectionTitle>{t('dashboards.editor.breakdown.title')}</SectionTitle>
          <BreakdownToggle
            value={builder.breakdown}
            onChange={(breakdown) => onChange((b) => ({ ...b, breakdown }))}
          />

          {chart === 'time' ? (
            <>
              <IntervalSelect
                value={builder.interval}
                onChange={(interval) => onChange((b) => ({ ...b, interval }))}
              />
              <DimensionPicker
                value={builder.dimension}
                fields={groupableFields ?? []}
                loading={fieldsLoading}
                onChange={(next) => onChange((b) => ({ ...b, dimension: next }))}
              />
              <p className="text-[11px] text-muted-foreground">
                {t('dashboards.editor.breakdown.splitHint')}
              </p>
            </>
          ) : (
            <DimensionPicker
              value={builder.dimension}
              fields={groupableFields ?? []}
              loading={fieldsLoading}
              onChange={(next) => onChange((b) => ({ ...b, dimension: next }))}
            />
          )}

          <LimitInput
            value={builder.limit}
            onChange={(limit) => onChange((b) => ({ ...b, limit }))}
          />
        </section>
      )}

      {chart === 'table' && (
        <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
          <SectionTitle>{t('dashboards.editor.table.columnsTitle')}</SectionTitle>
          <ColumnsPicker
            value={builder.columns}
            fields={fields ?? []}
            loading={fieldsLoading}
            onChange={(next) => onChange((b) => ({ ...b, columns: next }))}
          />
          <p className="text-[11px] text-muted-foreground">{t('dashboards.editor.table.hint')}</p>
          <LimitInput
            value={builder.limit}
            onChange={(limit) => onChange((b) => ({ ...b, limit }))}
          />
        </section>
      )}
    </>
  )
}

function BreakdownToggle({
  value,
  onChange,
}: {
  value: BuilderState['breakdown']
  onChange: (next: BuilderState['breakdown']) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="inline-flex w-full overflow-hidden rounded-md border border-input">
      {(['time', 'field'] as const).map((mode) => (
        <button
          key={mode}
          type="button"
          onClick={() => onChange(mode)}
          className={
            'flex-1 px-3 py-1.5 text-xs font-medium transition-colors ' +
            (value === mode
              ? 'bg-primary text-primary-foreground'
              : 'bg-background text-muted-foreground hover:bg-muted')
          }
        >
          {t(`dashboards.editor.breakdown.${mode}`)}
        </button>
      ))}
    </div>
  )
}

function IntervalSelect({
  value,
  onChange,
}: {
  value: IntervalId
  onChange: (next: IntervalId) => void
}) {
  const { t } = useTranslation()
  return (
    <div>
      <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
        {t('dashboards.editor.breakdown.interval')}
      </label>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value as IntervalId)}
        className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
      >
        {INTERVALS.map((id) => (
          <option key={id || 'auto'} value={id}>
            {id === '' ? t('dashboards.editor.breakdown.intervalAuto') : id}
          </option>
        ))}
      </select>
    </div>
  )
}

function LimitInput({
  value,
  onChange,
}: {
  value: number | null
  onChange: (next: number | null) => void
}) {
  const { t } = useTranslation()
  return (
    <div>
      <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
        {t('dashboards.editor.breakdown.limit')}
      </label>
      <input
        type="number"
        min={1}
        max={10000}
        value={value ?? ''}
        placeholder={t('dashboards.editor.breakdown.limitPlaceholder') ?? undefined}
        onChange={(e) => {
          const n = Number(e.target.value)
          onChange(e.target.value === '' || Number.isNaN(n) ? null : n)
        }}
        className="h-9 w-full rounded-md border border-input bg-background px-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
      />
    </div>
  )
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return <h2 className="text-sm font-semibold text-foreground/90">{children}</h2>
}

export default VisualizationEditor
