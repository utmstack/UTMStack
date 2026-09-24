import { Download, RefreshCw, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { TimeRangePicker, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { SELECT_CLS } from '../lib/incident-meta'
import type { IncidentSeverity } from '../types/incident.types'

const SEVERITIES: IncidentSeverity[] = ['high', 'medium', 'low']

export function IncidentsToolbar({
  search,
  onSearch,
  severity,
  onSeverity,
  range,
  onRange,
  onExport,
  onRefresh,
  loading,
}: {
  search: string
  onSearch: (s: string) => void
  severity: IncidentSeverity | 'all'
  onSeverity: (s: IncidentSeverity | 'all') => void
  range: TimeRange
  onRange: (r: TimeRange) => void
  onExport: () => void
  onRefresh: () => void
  loading: boolean
}) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-wrap items-center gap-2">
      <div className="relative min-w-[240px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder={t('incidents.toolbar.search')}
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9"
        />
      </div>
      <select
        value={severity}
        onChange={(e) => onSeverity(e.target.value as IncidentSeverity | 'all')}
        className={SELECT_CLS}
      >
        <option value="all">{t('incidents.toolbar.allSeverities')}</option>
        {SEVERITIES.map((s) => (
          <option key={s} value={s}>
            {t(`incidents.sev.${s}`)}
          </option>
        ))}
      </select>
      <TimeRangePicker value={range} onChange={onRange} allowAllTime align="right" />
      <Button variant="outline" size="sm" onClick={onExport} title={t('incidents.toolbar.export')}>
        <Download size={14} />
      </Button>
      <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading} title={t('incidents.refresh')}>
        <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
      </Button>
    </div>
  )
}
