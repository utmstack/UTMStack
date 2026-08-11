import { Code2, Download, Play, RefreshCw, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { TimeRangePicker, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { DatasetSelector, type DatasetTypes } from './DatasetSelector'
import { SqlQueryEditor } from '@/shared/components/sql-editor'
import type { IndexField } from '../types/log-explorer.types'

export function QueryBar({
  sources,
  dataset,
  dataType,
  onSelect,
  textSearchable,
  searchInput,
  onSearchInput,
  sqlMode,
  onSqlMode,
  sqlInput,
  onSqlInput,
  fields,
  range,
  onRange,
  onRun,
  onRefresh,
  loading,
  onExport,
}: {
  sources: DatasetTypes[]
  dataset: string
  dataType: string | null
  onSelect: (dataset: string, dataType: string | null) => void
  textSearchable: boolean
  searchInput: string
  onSearchInput: (q: string) => void
  sqlMode: boolean
  onSqlMode: (b: boolean) => void
  sqlInput: string
  onSqlInput: (q: string) => void
  fields: IndexField[]
  range: TimeRange
  onRange: (r: TimeRange) => void
  onRun: () => void
  onRefresh: () => void
  loading: boolean
  onExport: () => void
}) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-wrap items-center gap-2 rounded-xl border border-border bg-card p-2">
      <DatasetSelector sources={sources} dataset={dataset} dataType={dataType} onSelect={onSelect} />

      <div className="h-5 w-px bg-border" />

      {/* Search input — free text or SQL */}
      <div className="relative min-w-[300px] flex-1">
        {sqlMode ? (
          <SqlQueryEditor
            value={sqlInput}
            onChange={onSqlInput}
            onRun={onRun}
            fields={fields}
            tables={sources.map((s) => s.dataset)}
            placeholder={'SELECT * FROM logs ORDER BY `@timestamp` DESC   —   Enter runs, Shift+Enter for a new line'}
          />
        ) : (
          <>
            <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={searchInput}
              onChange={(e) => onSearchInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && onRun()}
              disabled={!textSearchable}
              placeholder={t(
                textSearchable ? 'logExplorer.query.searchPlaceholder' : 'logExplorer.query.noTextSearch',
              )}
              className="h-9 border-0 bg-transparent pl-8 font-mono text-xs shadow-none focus-visible:ring-0"
            />
          </>
        )}
      </div>

      <div className="h-5 w-px bg-border" />

      {/* Time range */}
      <TimeRangePicker value={range} onChange={onRange} align="right" />

      <button
        onClick={() => onSqlMode(!sqlMode)}
        className={cn(
          'flex h-9 items-center gap-1.5 rounded-md px-2.5 text-xs transition-colors',
          sqlMode
            ? 'bg-violet-500/15 text-violet-600 dark:text-violet-300'
            : 'text-muted-foreground hover:bg-muted hover:text-foreground'
        )}
        title="Toggle SQL mode"
      >
        <Code2 size={13} />
        SQL
      </button>

      <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading} title={t('logExplorer.query.refresh')}>
        <RefreshCw size={13} className={cn(loading && 'animate-spin')} />
      </Button>

      <Button variant="outline" size="sm" onClick={onExport} title={t('logExplorer.query.exportCsv')}>
        <Download size={13} />
      </Button>

      {/* The verb follows the box: one searches, the other runs a statement. */}
      <Button size="sm" onClick={onRun}>
        {sqlMode ? <Play size={12} className="mr-1.5" /> : <Search size={12} className="mr-1.5" />}
        {t(sqlMode ? 'logExplorer.query.run' : 'logExplorer.query.search')}
      </Button>
    </div>
  )
}
