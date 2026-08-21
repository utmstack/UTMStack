import { Plus, Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { SELECT_CLS, STATUSES , statusKey} from '../lib/incident-meta'
import type { IncidentStatus } from '../types/incident.types'

export interface IncidentsToolbarProps {
  search: string
  onSearch: (v: string) => void
  statusFilter: IncidentStatus | 'all'
  onStatusFilter: (v: IncidentStatus | 'all') => void
  assignee: string
  onAssignee: (v: string) => void
  assigneeOptions: string[]
  dateFrom: string
  onDateFrom: (v: string) => void
  dateTo: string
  onDateTo: (v: string) => void
  hasFilters: boolean
  onClear: () => void
  onCreate: () => void
}

export function IncidentsToolbar({
  search,
  onSearch,
  statusFilter,
  onStatusFilter,
  assignee,
  onAssignee,
  assigneeOptions,
  dateFrom,
  onDateFrom,
  dateTo,
  onDateTo,
  hasFilters,
  onClear,
  onCreate,
}: IncidentsToolbarProps) {
  const { t } = useTranslation()
  return (
    <div className="mt-4 flex flex-wrap items-center gap-2">
      <div className="relative min-w-[220px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder={t('incidents.toolbar.search')}
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9"
        />
      </div>
      <select
        value={statusFilter}
        onChange={(e) => onStatusFilter(e.target.value as IncidentStatus | 'all')}
        className={SELECT_CLS}
      >
        <option value="all">{t('incidents.toolbar.allStatuses')}</option>
        {STATUSES.map((s) => (
          <option key={s} value={s}>
            {t(`incidents.status.${statusKey(s)}`)}
          </option>
        ))}
      </select>
      <select value={assignee} onChange={(e) => onAssignee(e.target.value)} className={SELECT_CLS}>
        <option value="all">{t('incidents.toolbar.allAssignees')}</option>
        {assigneeOptions.map((a) => (
          <option key={a} value={a}>
            {a}
          </option>
        ))}
      </select>
      <input
        type="date"
        value={dateFrom}
        onChange={(e) => onDateFrom(e.target.value)}
        title={t('incidents.toolbar.from')}
        className={SELECT_CLS}
      />
      <input
        type="date"
        value={dateTo}
        onChange={(e) => onDateTo(e.target.value)}
        title={t('incidents.toolbar.to')}
        className={SELECT_CLS}
      />
      {hasFilters && (
        <button onClick={onClear} className="text-xs text-muted-foreground hover:text-foreground hover:underline">
          {t('incidents.toolbar.clear')}
        </button>
      )}
      <Button size="sm" onClick={onCreate} className="ml-auto">
        <Plus size={14} className="mr-1.5" />
        {t('incidents.toolbar.create')}
      </Button>
    </div>
  )
}
