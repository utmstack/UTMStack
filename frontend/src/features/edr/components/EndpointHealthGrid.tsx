import { useMemo, useState } from 'react'
import { Check, ChevronDown, ChevronUp, Filter, Search, ShieldAlert, ShieldCheck, Square, X } from 'lucide-react'
import type { TFunction } from 'i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { ColumnResizeHandle } from '@/shared/components/ui/column-resize-handle'
import { useResizableColumns } from '@/shared/hooks/useResizableColumns'
import type { EdrBulkAction, EdrEndpoint, EdrHealth } from '../types/edr.types'

interface EndpointHealthGridProps {
  endpoints: EdrEndpoint[]
  selectedIds: string[]
  onSelectionChange: (ids: string[]) => void
  onOpen: (endpoint: EdrEndpoint) => void
  onBulkAction: (action: EdrBulkAction) => void
  t: TFunction
}

type SortKey = keyof Pick<EdrEndpoint, 'hostName' | 'operatingSystem' | 'moduleVersion' | 'serviceState' | 'engineHealth' | 'signatureAge' | 'assignedPolicy' | 'policyDrift' | 'detectionsLastWeek' | 'quarantinedFiles' | 'isolationState' | 'lastReportTime'>

const columns: Array<{ key: SortKey; label: string }> = [
  { key: 'hostName', label: 'hostName' },
  { key: 'operatingSystem', label: 'operatingSystem' },
  { key: 'moduleVersion', label: 'moduleVersion' },
  { key: 'serviceState', label: 'serviceState' },
  { key: 'engineHealth', label: 'engineHealth' },
  { key: 'signatureAge', label: 'signatureAge' },
  { key: 'assignedPolicy', label: 'assignedPolicy' },
  { key: 'policyDrift', label: 'policyDrift' },
  { key: 'detectionsLastWeek', label: 'detectionsLastWeek' },
  { key: 'quarantinedFiles', label: 'quarantinedFiles' },
  { key: 'isolationState', label: 'isolationState' },
  { key: 'lastReportTime', label: 'lastReportTime' },
]

const endpointColumnWidths = [40, 180, 150, 120, 120, 130, 120, 170, 120, 120, 125, 120, 170, 230]
const endpointColumnMins = [40, 130, 120, 100, 100, 110, 100, 130, 100, 110, 100, 110, 140, 170]

const healthTone: Record<EdrHealth, string> = {
  healthy: 'text-emerald-600 dark:text-emerald-400',
  unhealthy: 'text-red-600 dark:text-red-400',
  stale: 'text-amber-600 dark:text-amber-400',
  degraded: 'text-orange-600 dark:text-orange-400',
  off: 'text-slate-500',
  unknown: 'text-muted-foreground',
}

const displayValue = (endpoint: EdrEndpoint, key: SortKey, t: TFunction): string => {
  const value = endpoint[key]
  if (key === 'signatureAge') return value == null ? '—' : `${value} ${t('edr.units.days')}`
  if (key === 'policyDrift') return value ? t('edr.values.drifted') : t('edr.values.inSync')
  if (key === 'lastReportTime') return value ? new Date(String(value)).toLocaleString() : '—'
  return String(value ?? '—')
}

export function EndpointHealthGrid({ endpoints, selectedIds, onSelectionChange, onOpen, onBulkAction, t }: EndpointHealthGridProps) {
  const [query, setQuery] = useState('')
  const [healthFilter, setHealthFilter] = useState<EdrHealth | 'all'>('all')
  const [sort, setSort] = useState<{ key: SortKey; direction: 'asc' | 'desc' }>({ key: 'hostName', direction: 'asc' })
  const { widths, startDrag } = useResizableColumns(endpointColumnWidths, {
    min: endpointColumnMins,
    storageKey: 'edr-endpoint-health-grid-columns',
  })

  const visible = useMemo(() => {
    const needle = query.trim().toLowerCase()
    return endpoints
      .filter((endpoint) => healthFilter === 'all' || endpoint.engineHealth === healthFilter)
      .filter((endpoint) => {
        if (!needle) return true
        return [endpoint.hostName, endpoint.operatingSystem, endpoint.moduleVersion, endpoint.serviceState, endpoint.engineHealth, endpoint.assignedPolicy, endpoint.sensors.join(' '), endpoint.isolationState]
          .join(' ').toLowerCase().includes(needle)
      })
      .sort((a, b) => {
        const left = displayValue(a, sort.key, t)
        const right = displayValue(b, sort.key, t)
        const result = left.localeCompare(right, undefined, { numeric: true, sensitivity: 'base' })
        return sort.direction === 'asc' ? result : -result
      })
  }, [endpoints, healthFilter, query, sort, t])

  const visibleIds = visible.map((endpoint) => endpoint.id)
  const allVisibleSelected = visibleIds.length > 0 && visibleIds.every((id) => selectedIds.includes(id))
  const toggleSort = (key: SortKey) => setSort((current) => current.key === key ? { key, direction: current.direction === 'asc' ? 'desc' : 'asc' } : { key, direction: 'asc' })
  const toggleSelection = (id: string) => onSelectionChange(selectedIds.includes(id) ? selectedIds.filter((value) => value !== id) : [...selectedIds, id])
  const toggleAll = () => onSelectionChange(allVisibleSelected ? selectedIds.filter((id) => !visibleIds.includes(id)) : [...new Set([...selectedIds, ...visibleIds])])

  return (
    <section className="min-w-0 overflow-hidden rounded-xl border border-border bg-card">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border px-4 py-3">
        <div className="flex items-center gap-2 text-sm font-medium">
          <ShieldCheck size={16} className="text-emerald-500" />
          {t('edr.grid.title')}
          <span className="text-xs font-normal text-muted-foreground">{t('edr.grid.count', { count: visible.length })}</span>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {selectedIds.length > 0 && (
            <>
              <Button size="sm" variant="outline" onClick={() => onBulkAction('apply-policy')}><ShieldCheck size={13} className="mr-1.5" />{t('edr.actions.applyPolicy')}</Button>
              <Button size="sm" variant="outline" onClick={() => onBulkAction('scan')}><Search size={13} className="mr-1.5" />{t('edr.actions.scan')}</Button>
              <Button size="sm" variant="outline" onClick={() => onBulkAction('isolate')}><ShieldAlert size={13} className="mr-1.5 text-red-500" />{t('edr.actions.isolate')}</Button>
            </>
          )}
          <div className="relative"><Search size={14} className="absolute left-2.5 top-2.5 text-muted-foreground" /><Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder={t('edr.grid.search')} className="h-8 w-52 pl-8 text-xs" /></div>
          <div className="relative"><Filter size={14} className="absolute left-2.5 top-2.5 text-muted-foreground" /><select value={healthFilter} onChange={(event) => setHealthFilter(event.target.value as EdrHealth | 'all')} className="h-8 appearance-none rounded-md border border-input bg-background py-1 pl-8 pr-7 text-xs"><option value="all">{t('edr.filters.allHealth')}</option><option value="healthy">{t('edr.health.healthy')}</option><option value="unhealthy">{t('edr.health.unhealthy')}</option><option value="stale">{t('edr.health.stale')}</option><option value="degraded">{t('edr.health.degraded')}</option><option value="off">{t('edr.health.off')}</option></select></div>
          {(query || healthFilter !== 'all') && <button onClick={() => { setQuery(''); setHealthFilter('all') }} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted" title={t('edr.filters.clear')}><X size={14} /></button>}
        </div>
      </div>
      <div className="overflow-x-auto">
        <table className="w-max min-w-[1800px] table-fixed border-collapse text-xs">
          <colgroup>{widths.map((width, index) => <col key={index} style={{ width: `${width}px` }} />)}</colgroup>
          <thead className="bg-muted/40 text-[10px] uppercase tracking-wider text-muted-foreground">
            <tr>
              <th data-resizable-col className="relative px-3 py-2 text-left"><button onClick={toggleAll} aria-label={t('edr.grid.selectAll')} className="text-muted-foreground">{allVisibleSelected ? <Check size={15} /> : <Square size={15} />}</button></th>
              {columns.map((column, index) => <th key={column.key} data-resizable-col className="relative whitespace-nowrap px-3 py-2 text-left"><button onClick={() => toggleSort(column.key)} className="inline-flex items-center gap-1 hover:text-foreground">{t(`edr.columns.${column.label}`)}{sort.key === column.key && (sort.direction === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}</button><ColumnResizeHandle onMouseDown={startDrag(index + 1)} /></th>)}
              <th data-resizable-col className="relative px-3 py-2 text-left">{t('edr.columns.sensors')}</th>
            </tr>
          </thead>
          <tbody>
            {visible.map((endpoint) => (
              <tr key={endpoint.id} onClick={() => onOpen(endpoint)} className="cursor-pointer border-t border-border/60 hover:bg-muted/30">
                <td className="px-3 py-2" onClick={(event) => event.stopPropagation()}><button onClick={() => toggleSelection(endpoint.id)} aria-label={t('edr.grid.selectEndpoint')} className="text-muted-foreground hover:text-foreground">{selectedIds.includes(endpoint.id) ? <Check size={15} className="text-primary" /> : <Square size={15} />}</button></td>
                {columns.map((column) => <td key={column.key} className={cn('whitespace-nowrap px-3 py-2', column.key === 'engineHealth' && healthTone[endpoint.engineHealth], column.key === 'policyDrift' && (endpoint.policyDrift ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'), column.key === 'isolationState' && endpoint.isolationState === 'isolated' && 'font-semibold text-red-600 dark:text-red-400')}>{displayValue(endpoint, column.key, t)}</td>)}
                <td className="max-w-[220px] px-3 py-2"><div className="flex flex-wrap gap-1">{endpoint.sensors.length > 0 ? endpoint.sensors.map((sensor) => <span key={sensor} className="rounded bg-muted px-1.5 py-0.5 text-[10px]">{sensor}</span>) : <span className="text-muted-foreground">—</span>}</div></td>
              </tr>
            ))}
          </tbody>
        </table>
        {visible.length === 0 && <div className="px-6 py-14 text-center text-sm text-muted-foreground">{t('edr.grid.empty')}</div>}
      </div>
    </section>
  )
}
