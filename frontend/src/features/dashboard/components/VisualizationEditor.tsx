import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { ChartTypePicker } from '@/features/dashboard/components/editor/ChartTypePicker'
import { IndexPatternSelect } from '@/features/dashboard/components/editor/IndexPatternSelect'
import { FilterBuilder } from '@/features/dashboard/components/editor/FilterBuilder'
import { MetricPicker } from '@/features/dashboard/components/editor/MetricPicker'
import { DimensionPicker } from '@/features/dashboard/components/editor/DimensionPicker'
import { SqlPreview } from '@/features/dashboard/components/editor/SqlPreview'
import { ChartPreviewPanel } from '@/features/dashboard/components/editor/ChartPreviewPanel'
import { useIndexPatternFields, useIndexPatterns } from '@/features/dashboard/hooks/useIndexPatterns'
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
}

export function VisualizationEditor({ initial }: VisualizationEditorProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const isEdit = initial != null

  const [name, setName] = useState(initial?.name ?? '')
  const [description, setDescription] = useState(initial?.description ?? '')

  const initialParsed = useMemo(
    () => parseBuilderConfig(initial?.config ?? null),
    [initial?.config]
  )

  const [option, setOption] = useState<Record<string, unknown>>(() => {
    if (initial && initialParsed.builder && Object.keys(initialParsed.option).length > 0) {
      return initialParsed.option
    }
    const startBuilder = initialParsed.builder ?? makeInitialBuilder()
    return getChartTypeMeta(startBuilder.chartType).defaultConfig
  })

  const [builder, setBuilder] = useState<BuilderState>(() => {
    if (initialParsed.builder) return initialParsed.builder
    if (initial) {
      // Legacy visualization without __builder — open in raw mode preserving SQL.
      return {
        ...makeInitialBuilder(),
        rawMode: true,
        rawSql: initial.sqlQuery ?? null,
      }
    }
    return makeInitialBuilder()
  })

  const patternsQuery = useIndexPatterns()
  const patternId = useMemo(() => {
    const list = patternsQuery.data?.data ?? []
    return list.find((p) => p.pattern === builder.indexPattern)?.id ?? null
  }, [patternsQuery.data?.data, builder.indexPattern])
  const fieldsQuery = useIndexPatternFields(patternId)
  const fields = fieldsQuery.data?.fields ?? []

  const composedSql = useMemo(() => composeSql(builder), [builder])

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

  const valid =
    name.trim().length > 0 &&
    builder.indexPattern.trim().length > 0 &&
    sqlForSave.length > 0

  const save = (e: React.FormEvent) => {
    e.preventDefault()
    if (!valid || busy) return

    const configJson = serializeBuilderConfig(option, builder)

    if (isEdit && initial) {
      updateVisualization.mutate(
        {
          id: initial.id,
          name: name.trim(),
          description: description.trim() || undefined,
          sqlQuery: sqlForSave,
          config: configJson,
        },
        {
          onSuccess: () => {
            toast.success(t('dashboards.toast.visualizationUpdated'))
            navigate('/dashboards/visualizations')
          },
          onError: (err) =>
            toast.error(err.message ?? t('dashboards.toast.visualizationUpdateFailed')),
        }
      )
    } else {
      createVisualization.mutate(
        {
          name: name.trim(),
          description: description.trim() || undefined,
          sqlQuery: sqlForSave,
          config: configJson,
        },
        {
          onSuccess: () => {
            toast.success(t('dashboards.toast.visualizationCreated'))
            navigate('/dashboards/visualizations')
          },
          onError: (err) =>
            toast.error(err.message ?? t('dashboards.toast.visualizationCreateFailed')),
        }
      )
    }
  }

  return (
    <form
      onSubmit={save}
      className="mx-auto flex h-full w-full max-w-[1400px] flex-col gap-4 px-6 py-6"
    >
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
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
          <Button type="submit" size="sm" disabled={!valid || busy}>
            {busy && <Loader2 size={14} className="mr-1 animate-spin" />}
            {isEdit ? t('dashboards.form.save') : t('dashboards.form.create')}
          </Button>
        </div>
      </header>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <div className="flex flex-col gap-4">
          <section className="rounded-lg border border-border bg-card p-4">
            <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
              <div>
                <label className="mb-1.5 block text-xs font-medium text-foreground/80">
                  {t('dashboards.form.name')}
                </label>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder={t('dashboards.form.namePlaceholder') ?? ''}
                  autoFocus
                />
              </div>
              <div>
                <label className="mb-1.5 block text-xs font-medium text-foreground/80">
                  {t('dashboards.form.description')}
                </label>
                <Input
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder={t('dashboards.form.descriptionPlaceholder') ?? ''}
                />
              </div>
            </div>
          </section>

          <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
            <SectionTitle>{t('dashboards.editor.chartType.title')}</SectionTitle>
            <ChartTypePicker value={builder.chartType} onChange={setChartType} />
          </section>

          <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
            <SectionTitle>{t('dashboards.editor.dataSource.title')}</SectionTitle>
            <div>
              <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                {t('dashboards.editor.dataSource.indexPattern')}
              </label>
              <IndexPatternSelect
                value={builder.indexPattern}
                onChange={(pattern) =>
                  setBuilder((b) => ({ ...b, indexPattern: pattern }))
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
              fields={fields}
              onChange={(next) => setBuilder((b) => ({ ...b, filters: next }))}
            />
          </section>

          {(showMetric || showDimension) && (
            <section className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
              {showMetric && (
                <>
                  <SectionTitle>{t('dashboards.editor.metric.title')}</SectionTitle>
                  <MetricPicker
                    value={builder.metric}
                    fields={fields}
                    onChange={(next) => setBuilder((b) => ({ ...b, metric: next }))}
                  />
                </>
              )}
              {showDimension && (
                <>
                  <SectionTitle>{t('dashboards.editor.dimension.title')}</SectionTitle>
                  <DimensionPicker
                    value={builder.dimension}
                    fields={fields}
                    onChange={(next) => setBuilder((b) => ({ ...b, dimension: next }))}
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
                  setBuilder((b) => ({
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
            <SqlPreview
              sql={composedSql}
              rawMode={builder.rawMode}
              onToggleRaw={(next) =>
                setBuilder((b) => ({
                  ...b,
                  rawMode: next,
                  rawSql: next && !b.rawSql ? composedSql : b.rawSql,
                }))
              }
              rawSql={builder.rawSql ?? ''}
              onChangeRawSql={(next) =>
                setBuilder((b) => ({ ...b, rawSql: next }))
              }
            />
          </section>
        </div>

        <div className="lg:sticky lg:top-4 lg:self-start">
          <section className="rounded-lg border border-border bg-card p-4">
            <ChartPreviewPanel
              sql={sqlForSave}
              option={option}
              renderer={meta.renderer}
              label={name || undefined}
            />
          </section>
        </div>
      </div>
    </form>
  )
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return <h2 className="text-sm font-semibold text-foreground/90">{children}</h2>
}

export default VisualizationEditor
