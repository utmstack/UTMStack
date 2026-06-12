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
import { SqlPreview } from '@/features/dashboard/components/editor/SqlPreview'
import { ChartPreviewPanel } from '@/features/dashboard/components/editor/ChartPreviewPanel'
import { ChartTypeModal } from '@/features/dashboard/components/editor/ChartTypeModal'
import { SaveVisualizationDialog } from '@/features/dashboard/components/editor/SaveVisualizationDialog'
import { useIndexProperties } from '@/features/dashboard/hooks/useIndexProperties'
import { useVisualizationMutations } from '@/features/dashboard/hooks/useVisualizations'
import { getChartTypeMeta } from '@/features/dashboard/constants'
import { composeSql } from '@/features/dashboard/utils/sql-builder'
import {
  makeInitialBuilder,
  parseBuilderConfig,
  serializeBuilderConfig,
} from '@/features/dashboard/utils/builder-config'
import type {
  BuilderState,
  ChartTypeId,
  Visualization,
} from '@/features/dashboard/types'

export interface VisualizationEditorProps {
  initial: Visualization | null
  initialChartType?: ChartTypeId
}

type EditorTab = 'visual' | 'sql'

export function VisualizationEditor({ initial, initialChartType }: VisualizationEditorProps) {
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
    if (initialParsed.builder) return initialParsed.builder
    if (initial) {
      return {
        ...makeInitialBuilder(),
        chartType: initialChartType ?? 'bar',
        rawMode: true,
        rawSql: initial.sqlQuery ?? null,
      }
    }
    return { ...makeInitialBuilder(), chartType: initialChartType ?? 'bar' }
  })

  const [tab, setTab] = useState<EditorTab>(() => (builder.rawMode ? 'sql' : 'visual'))
  const [chartTypeOpen, setChartTypeOpen] = useState(false)
  const [saveOpen, setSaveOpen] = useState(false)

  const fieldsQuery = useIndexProperties(builder.indexPattern)
  const fields = fieldsQuery.data ?? []
  const fieldsLoading = fieldsQuery.isFetching && !!builder.indexPattern

  const composedSql = useMemo(() => composeSql(builder), [builder])

  const switchTab = (next: EditorTab) => {
    setTab(next)
    if (next === 'sql') {
      setBuilder((b) => ({
        ...b,
        rawMode: true,
        rawSql: b.rawSql && b.rawSql.trim() ? b.rawSql : composedSql,
      }))
    } else {
      setBuilder((b) => ({ ...b, rawMode: false }))
    }
  }

  const setChartType = (chartType: ChartTypeId) => {
    setBuilder((b) => ({ ...b, chartType }))
    if (!builder.configTouched) {
      setOption(getChartTypeMeta(chartType).defaultConfig)
    }
  }

  const meta = getChartTypeMeta(builder.chartType)
  const showDimension = builder.chartType !== 'metric' && builder.chartType !== 'table'
  const showMetric = builder.chartType !== 'table'

  const { createVisualization, updateVisualization } = useVisualizationMutations()
  const busy = createVisualization.isPending || updateVisualization.isPending

  const sqlForSave = builder.rawMode ? (builder.rawSql ?? '').trim() : composedSql.trim()

  const ready =
    builder.indexPattern.trim().length > 0 && sqlForSave.length > 0

  const openSaveDialog = () => {
    if (!ready) {
      toast.error(t('dashboards.editor.toast.notReady'))
      return
    }
    setSaveOpen(true)
  }

  const handleSubmit = ({ name, description }: { name: string; description?: string }) => {
    const configJson = serializeBuilderConfig(option, builder)

    if (isEdit && initial) {
      updateVisualization.mutate(
        {
          id: initial.id,
          name,
          description,
          sqlQuery: sqlForSave,
          config: configJson,
        },
        {
          onSuccess: () => {
            toast.success(t('dashboards.toast.visualizationUpdated'))
            setSaveOpen(false)
            navigate('/dashboards/visualizations')
          },
          onError: (err) =>
            toast.error(err.message ?? t('dashboards.toast.visualizationUpdateFailed')),
        }
      )
    } else {
      createVisualization.mutate(
        {
          name,
          description,
          sqlQuery: sqlForSave,
          config: configJson,
        },
        {
          onSuccess: () => {
            toast.success(t('dashboards.toast.visualizationCreated'))
            setSaveOpen(false)
            navigate('/dashboards/visualizations')
          },
          onError: (err) =>
            toast.error(err.message ?? t('dashboards.toast.visualizationCreateFailed')),
        }
      )
    }
  }

  return (
    <div className="mx-auto flex h-full w-full max-w-[1400px] flex-col gap-4 px-6 py-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex flex-col gap-1.5">
          <h1 className="text-xl font-semibold">
            {isEdit
              ? t('dashboards.editor.editTitle')
              : t('dashboards.editor.newTitle')}
          </h1>
          <p className="text-sm text-muted-foreground">
            {isEdit
              ? t('dashboards.editor.editSubtitle')
              : t('dashboards.editor.newSubtitle')}
          </p>
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
            onClick={() => navigate('/dashboards/visualizations')}
            disabled={busy}
          >
            {t('dashboards.form.cancel')}
          </Button>
          <Button type="button" size="sm" onClick={openSaveDialog} disabled={!ready || busy}>
            {busy && <Loader2 size={14} className="mr-1 animate-spin" />}
            {t('dashboards.editor.saveVisualization')}
          </Button>
        </div>
      </header>

      <ModeTabs tab={tab} onChange={switchTab} />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <div className="flex flex-col gap-4">
          {tab === 'visual' ? (
            <VisualTab
              builder={builder}
              fields={fields}
              fieldsLoading={fieldsLoading}
              showMetric={showMetric}
              showDimension={showDimension}
              composedSql={composedSql}
              onBuilderChange={setBuilder}
            />
          ) : (
            <SqlTab
              rawSql={builder.rawSql ?? ''}
              onChange={(next) => setBuilder((b) => ({ ...b, rawSql: next }))}
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

      <SaveVisualizationDialog
        open={saveOpen}
        mode={isEdit ? 'update' : 'create'}
        initialName={initial?.name}
        initialDescription={initial?.description}
        busy={busy}
        onClose={() => (busy ? null : setSaveOpen(false))}
        onSubmit={handleSubmit}
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
  fieldsLoading,
  showMetric,
  showDimension,
  composedSql,
  onBuilderChange,
}: {
  builder: BuilderState
  fields: ReturnType<typeof useIndexProperties>['data'] extends infer T ? T extends undefined ? never : T : never
  fieldsLoading: boolean
  showMetric: boolean
  showDimension: boolean
  composedSql: string
  onBuilderChange: React.Dispatch<React.SetStateAction<BuilderState>>
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
                fields={fields ?? []}
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
                fields={fields ?? []}
                loading={fieldsLoading}
                onChange={(next) => onBuilderChange((b) => ({ ...b, dimension: next }))}
              />
            </>
          )}
        </section>
      )}

      {builder.chartType === 'table' && (
        <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
          <SectionTitle>{t('dashboards.editor.table.title')}</SectionTitle>
          <textarea
            value={builder.advancedSelect ?? ''}
            onChange={(e) =>
              onBuilderChange((b) => ({
                ...b,
                advancedSelect: e.target.value || null,
              }))
            }
            rows={4}
            placeholder={t('dashboards.editor.table.placeholder') ?? ''}
            spellCheck={false}
            className="w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-xs leading-relaxed shadow-sm focus:border-ring focus:outline-none focus:ring-1 focus:ring-ring"
          />
        </section>
      )}

      <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
        <SqlPreview sql={composedSql} />
      </section>
    </>
  )
}

function SqlTab({
  rawSql,
  onChange,
}: {
  rawSql: string
  onChange: (next: string) => void
}) {
  const { t } = useTranslation()
  return (
    <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
      <SectionTitle>{t('dashboards.editor.tabs.sql')}</SectionTitle>
      <textarea
        value={rawSql}
        onChange={(e) => onChange(e.target.value)}
        spellCheck={false}
        rows={14}
        placeholder={t('dashboards.editor.sqlPreview.rawPlaceholder') ?? ''}
        className="w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-xs leading-relaxed shadow-sm focus:border-ring focus:outline-none focus:ring-1 focus:ring-ring"
      />
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
