import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Loader2, Pencil } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { cn } from '@/shared/lib/utils'
import { IndexPatternSelect } from '@/features/dashboard/components/editor/IndexPatternSelect'
import { FilterBuilder } from '@/features/dashboard/components/editor/FilterBuilder'
import { MetricPicker } from '@/features/dashboard/components/editor/MetricPicker'
import { DimensionPicker } from '@/features/dashboard/components/editor/DimensionPicker'
import { ColumnsPicker } from '@/features/dashboard/components/editor/ColumnsPicker'
import { ChartPreviewPanel } from '@/features/dashboard/components/editor/ChartPreviewPanel'
import { ChartTypeModal } from '@/features/dashboard/components/editor/ChartTypeModal'
import { useAggregatableFields } from '@/features/dashboard/hooks/useAggregatableFields'
import { useIndexPatterns } from '@/features/dashboard/hooks/useIndexPatterns'
import { useVisualizationMutations } from '@/features/dashboard/hooks/useVisualizations'
import { SqlQueryEditor } from '@/shared/components/sql-editor'
import { DEFAULT_WIDGET_LAYOUT, getChartTypeMeta } from '@/features/dashboard/constants'
import { serializeLayout } from '@/features/dashboard/utils/layout'
import { composeSql, parseSqlToBuilder } from '@/features/dashboard/utils/sql-builder'
import {
  makeInitialBuilder,
  parseBuilderConfig,
  serializeBuilderConfig,
} from '@/features/dashboard/utils/builder-config'
import type {
  BuilderState,
  ChartTypeId,
  IndexProperty,
  Visualization,
} from '@/features/dashboard/types'

export interface VisualizationEditorProps {
  initial: Visualization | null
  initialChartType?: ChartTypeId
  // The dashboard this visualization belongs to. For edits this always matches
  // `initial.dashboardId`; for a brand-new visualization it comes from the route.
  dashboardId: number
  // Grid position/size for a new widget (from DashboardPage's "Add widget"
  // seeding). Ignored on edit — the existing layout is preserved untouched;
  // repositioning happens by dragging on the dashboard grid, not here.
  initialLayout?: string
}

type EditorTab = 'visual' | 'sql'

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

  const [builder, setBuilder] = useState<BuilderState>(() => {
    // Seed rawSql so the SQL tab and the visual tab start out synchronised —
    // legacy widgets stored just `sqlQuery`, newer ones round-trip via composeSql.
    const seedRawSql = (b: BuilderState) => ({
      ...b,
      rawSql: b.rawSql && b.rawSql.trim() ? b.rawSql : composeSql({ ...b, rawMode: false }),
    })
    if (initialParsed.builder) return seedRawSql(initialParsed.builder)
    if (initial) {
      return seedRawSql({
        ...makeInitialBuilder(),
        chartType: initialChartType ?? 'bar',
        rawSql: initial.sqlQuery ?? null,
      })
    }
    return seedRawSql({ ...makeInitialBuilder(), chartType: initialChartType ?? 'bar' })
  })

  const [name, setName] = useState(initial?.name ?? '')
  const [description, setDescription] = useState(initial?.description ?? '')
  // Open on whichever tab the widget was last saved from: legacy widgets
  // (no builder config) land on SQL because we only have raw SQL for them.
  const [tab, setTab] = useState<EditorTab>(() => {
    if (initialParsed.builder?.rawMode) return 'sql'
    if (initial && !initialParsed.builder) return 'sql'
    return 'visual'
  })
  const [chartTypeOpen, setChartTypeOpen] = useState(false)

  const {
    fields,
    groupableFields: groupable,
    isLoading: fieldsLoading,
  } = useAggregatableFields(builder.indexPattern)
  const indexPatterns = useIndexPatterns()
  const patternList = useMemo(
    () => indexPatterns.data?.data ?? [],
    [indexPatterns.data?.data]
  )

  // Visual edits regenerate rawSql so the SQL tab reflects the change on switch,
  // and SQL edits reverse-parse into the builder so the visual tab reflects it
  // on switch. See parseSqlToBuilder for the exact fields that round-trip
  // (filters and chartType are intentionally NOT parsed back from SQL).
  const applyVisualEdit = (updater: (b: BuilderState) => BuilderState) => {
    setBuilder((b) => {
      const next = updater(b)
      return {
        ...next,
        rawMode: false,
        rawSql: composeSql({ ...next, rawMode: false }),
      }
    })
  }

  const applySqlEdit = (nextSql: string) => {
    setBuilder((b) => {
      const patch = parseSqlToBuilder(nextSql) ?? {}
      return { ...b, ...patch, rawSql: nextSql, rawMode: true }
    })
  }

  const setChartType = (chartType: ChartTypeId) => {
    applyVisualEdit((b) => ({ ...b, chartType }))
    if (!builder.configTouched) {
      setOption(getChartTypeMeta(chartType).defaultConfig)
    }
  }

  const meta = getChartTypeMeta(builder.chartType)
  const showDimension = builder.chartType !== 'metric' && builder.chartType !== 'table'
  const showMetric = builder.chartType !== 'table'

  const { createVisualization, updateVisualization } = useVisualizationMutations()
  const busy = createVisualization.isPending || updateVisualization.isPending

  const sqlForSave = (builder.rawSql ?? '').trim()

  const ready =
    name.trim().length > 0 &&
    builder.indexPattern.trim().length > 0 &&
    sqlForSave.length > 0

  // Save directly — name and description are captured inline, so no extra dialog.
  const handleSave = () => {
    if (!ready) {
      toast.error(t('dashboards.editor.toast.notReady'))
      return
    }
    if (busy) return

    const configJson = serializeBuilderConfig(option, builder)
    const backToDashboard = () => navigate('/dashboards/list', { state: { selectDashboardId: dashboardId } })

    if (isEdit && initial) {
      updateVisualization.mutate(
        {
          id: initial.id,
          dashboardId: initial.dashboardId,
          name: name.trim(),
          description: description.trim() || undefined,
          sqlQuery: sqlForSave,
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
          name: name.trim(),
          description: description.trim() || undefined,
          sqlQuery: sqlForSave,
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

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <div>
          <label
            htmlFor="viz-name"
            className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
          >
            {t('dashboards.editor.nameLabel')}
          </label>
          <input
            id="viz-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t('dashboards.editor.namePlaceholder') ?? ''}
            className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          />
        </div>
        <div>
          <label
            htmlFor="viz-description"
            className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
          >
            {t('dashboards.editor.descriptionLabel')}
          </label>
          <input
            id="viz-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={t('dashboards.form.descriptionPlaceholder') ?? ''}
            className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          />
        </div>
      </div>

      <ModeTabs tab={tab} onChange={setTab} />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <div className="flex flex-col gap-4">
          {tab === 'visual' ? (
            <VisualTab
              builder={builder}
              fields={fields}
              groupableFields={groupable}
              fieldsLoading={fieldsLoading}
              showMetric={showMetric}
              showDimension={showDimension}
              onBuilderChange={applyVisualEdit}
            />
          ) : (
            <SqlTab
              rawSql={builder.rawSql ?? ''}
              onChange={applySqlEdit}
              fields={fields}
              patterns={patternList}
            />
          )}
        </div>

        <div className="lg:sticky lg:top-4 lg:self-start">
          <section className="rounded-lg border border-border bg-card p-4">
            <ChartPreviewPanel
              sql={sqlForSave}
              option={option}
              renderer={meta.renderer}
              label={initial?.name}
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

function ModeTabs({
  tab,
  onChange,
}: {
  tab: EditorTab
  onChange: (next: EditorTab) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="flex items-center gap-1 border-b border-border">
      <TabButton active={tab === 'visual'} onClick={() => onChange('visual')}>
        {t('dashboards.editor.tabs.visual')}
      </TabButton>
      <TabButton active={tab === 'sql'} onClick={() => onChange('sql')}>
        {t('dashboards.editor.tabs.sql')}
      </TabButton>
    </div>
  )
}

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean
  onClick: () => void
  children: React.ReactNode
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        '-mb-px border-b-2 px-3 py-2 text-sm font-medium transition-colors',
        active
          ? 'border-primary text-foreground'
          : 'border-transparent text-muted-foreground hover:text-foreground'
      )}
    >
      {children}
    </button>
  )
}

function VisualTab({
  builder,
  fields,
  groupableFields,
  fieldsLoading,
  showMetric,
  showDimension,
  onBuilderChange,
}: {
  builder: BuilderState
  fields: IndexProperty[]
  groupableFields: IndexProperty[]
  fieldsLoading: boolean
  showMetric: boolean
  showDimension: boolean
  onBuilderChange: (updater: (b: BuilderState) => BuilderState) => void
}) {
  const { t } = useTranslation()
  return (
    <>
      <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
        <SectionTitle>{t('dashboards.editor.dataSource.title')}</SectionTitle>
        <div>
          <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
            {t('dashboards.editor.dataSource.indexPattern')}
          </label>
          <IndexPatternSelect
            value={builder.indexPattern}
            onChange={(pattern) =>
              onBuilderChange((b) => ({ ...b, indexPattern: pattern }))
            }
          />
          {!builder.indexPattern && (
            <p className="mt-1 text-[10px] text-muted-foreground">
              {t('dashboards.editor.dataSource.noPatternHint')}
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
          onChange={(next) => onBuilderChange((b) => ({ ...b, filters: next }))}
        />
      </section>

      {(showMetric || showDimension) && (
        <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
          {showMetric && (
            <>
              <SectionTitle>{t('dashboards.editor.metric.title')}</SectionTitle>
              <MetricPicker
                value={builder.metric}
                fields={groupableFields ?? []}
                loading={fieldsLoading}
                onChange={(next) => onBuilderChange((b) => ({ ...b, metric: next }))}
              />
            </>
          )}
          {showDimension && (
            <>
              <SectionTitle>{t('dashboards.editor.dimension.title')}</SectionTitle>
              <DimensionPicker
                value={builder.dimension}
                fields={groupableFields ?? []}
                loading={fieldsLoading}
                onChange={(next) => onBuilderChange((b) => ({ ...b, dimension: next }))}
              />
            </>
          )}
        </section>
      )}

      {builder.chartType === 'table' && (
        <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
          <SectionTitle>{t('dashboards.editor.table.columnsTitle')}</SectionTitle>
          <ColumnsPicker
            value={builder.columns}
            fields={fields ?? []}
            loading={fieldsLoading}
            onChange={(next) => onBuilderChange((b) => ({ ...b, columns: next }))}
          />
          <p className="text-[11px] text-muted-foreground">{t('dashboards.editor.table.hint')}</p>
        </section>
      )}
    </>
  )
}

function SqlTab({
  rawSql,
  onChange,
  fields,
  patterns,
}: {
  rawSql: string
  onChange: (next: string) => void
  fields: IndexProperty[]
  patterns: { pattern: string }[]
}) {
  const { t } = useTranslation()
  return (
    <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
      <SectionTitle>{t('dashboards.editor.tabs.sql')}</SectionTitle>
      <div className="rounded-md border border-input bg-background shadow-sm">
        <SqlQueryEditor
          value={rawSql}
          onChange={onChange}
          fields={fields}
          patterns={patterns}
          placeholder={t('dashboards.editor.sqlPreview.rawPlaceholder') ?? undefined}
          minRows={10}
          maxRows={20}
        />
      </div>
      <p className="text-[10px] text-muted-foreground">
        {t('dashboards.editor.sqlPreview.hint')}
      </p>
    </section>
  )
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return <h2 className="text-sm font-semibold text-foreground/90">{children}</h2>
}

export default VisualizationEditor
