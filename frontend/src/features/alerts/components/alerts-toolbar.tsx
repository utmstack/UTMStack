import { Download, RefreshCw, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { TimeRangePicker, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { SELECT_CLS } from '../lib/alert-meta'
import type { AlertTag, SeverityKey } from '../types/alert.types'
import { TagFilter } from './tag-filter'

export function AlertsToolbar({
  search,
  onSearch,
  severity,
  onSeverity,
  range,
  onRange,
  tagCatalog,
  tagFilter,
  onTagFilter,
  onCreateTag,
  onUpdateTag,
  onDeleteTag,
  onCreateRule,
  onRefresh,
  onExport,
  loading,
  showSeverity = true,
}: {
  search: string
  onSearch: (s: string) => void
  severity: SeverityKey | 'all'
  onSeverity: (s: SeverityKey | 'all') => void
  range: TimeRange
  onRange: (r: TimeRange) => void
  tagCatalog: AlertTag[]
  tagFilter: string[]
  onTagFilter: (tags: string[]) => void
  onCreateTag: (tagName: string, tagColor: string) => void
  onUpdateTag: (id: number, tagName: string, tagColor: string) => void
  onDeleteTag: (id: number, tagName: string) => void
  onCreateRule: (tg: AlertTag) => void
  onRefresh: () => void
  onExport: () => void
  loading: boolean
  showSeverity?: boolean
}) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-wrap items-center gap-2">
      <div className="relative min-w-[240px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder={t('alerts.toolbar.searchPlaceholder')}
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9"
        />
      </div>
      {showSeverity && (
        <select
          value={severity}
          onChange={(e) => onSeverity(e.target.value as SeverityKey | 'all')}
          className={SELECT_CLS}
        >
          <option value="all">{t('alerts.toolbar.allSeverities')}</option>
          <option value="high">{t('alerts.severity.high')}</option>
          <option value="medium">{t('alerts.severity.medium')}</option>
          <option value="low">{t('alerts.severity.low')}</option>
        </select>
      )}
      <TagFilter
        catalog={tagCatalog}
        selected={tagFilter}
        onSelected={onTagFilter}
        onCreateTag={onCreateTag}
        onUpdateTag={onUpdateTag}
        onDeleteTag={onDeleteTag}
        onCreateRule={onCreateRule}
      />
      <TimeRangePicker value={range} onChange={onRange} allowAllTime align="right" />
      <Button variant="outline" size="sm" onClick={onExport} title={t('alerts.toolbar.export')}>
        <Download size={14} />
      </Button>
      <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading} title={t('alerts.toolbar.refresh')}>
        <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
      </Button>
    </div>
  )
}
