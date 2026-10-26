import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { Download, Plus, Trash2, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import type { BreachStatus, DarkWebRecord, OrganizationElementType } from '../types'
import type { ExportAction } from './export-types'

type ExportFilterField = 'organizationElementType' | 'breachStatus' | 'employee' | 'domain' | 'source'
type ExportFilterOperator = 'equals' | 'contains' | 'notContains'

interface ExportFilter {
  field: ExportFilterField
  operator: ExportFilterOperator
  value: string
}

interface AdvancedExportOptionsDialogProps {
  records: DarkWebRecord[]
  onClose: () => void
  onExport: (records: DarkWebRecord[], action: ExportAction, filename: string, maskSensitive: boolean) => boolean
}

const FILTER_FIELDS: ExportFilterField[] = ['organizationElementType', 'breachStatus', 'employee', 'domain', 'source']
const FILTER_OPERATORS: ExportFilterOperator[] = ['equals', 'contains', 'notContains']
const ELEMENT_TYPES: OrganizationElementType[] = ['DOMAIN', 'EMAIL', 'IP', 'USER', 'PHONE']
const BREACH_STATUSES: BreachStatus[] = ['OPEN', 'IN_MITIGATION', 'COMPLETED']
const STATUS_KEYS: Record<BreachStatus, string> = {
  OPEN: 'open',
  IN_MITIGATION: 'mitigation',
  COMPLETED: 'completed',
}

export function AdvancedExportOptionsDialog({ records, onClose, onExport }: AdvancedExportOptionsDialogProps) {
  const { t } = useTranslation()
  const [employeeOnly, setEmployeeOnly] = useState(false)
  const [maskSensitive, setMaskSensitive] = useState(true)
  const [action, setAction] = useState<ExportAction>('CSV')
  const [field, setField] = useState<ExportFilterField>('organizationElementType')
  const [operator, setOperator] = useState<ExportFilterOperator>('contains')
  const [value, setValue] = useState('')
  const [filters, setFilters] = useState<ExportFilter[]>([])

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [onClose])

  const getFilterValue = (record: DarkWebRecord, filterField: ExportFilterField): string => {
    switch (filterField) {
      case 'organizationElementType':
        return record.organizationElementType
      case 'breachStatus':
        return record.breachStatus
      case 'employee':
        return record.employee ? 'true' : 'false'
      case 'domain':
        return record.domain ?? ''
      case 'source':
        return record.source ?? ''
    }
  }

  const addFilter = () => {
    if (!value.trim()) return
    setFilters((current) => [...current, { field, operator, value: value.trim() }])
    setValue('')
  }

  const handleExport = () => {
    const filteredRecords = records.filter((record) => {
      if (employeeOnly && !record.employee) return false
      return filters.every((filter) => {
        const recordValue = getFilterValue(record, filter.field).toLowerCase()
        const filterValue = filter.value.toLowerCase()
        if (filter.operator === 'equals') return recordValue === filterValue
        if (filter.operator === 'notContains') return !recordValue.includes(filterValue)
        return recordValue.includes(filterValue)
      })
    })
    if (onExport(filteredRecords, action, 'dark-web-monitoring-advanced-export', maskSensitive)) onClose()
  }

  return createPortal(
    <div
      className="fixed inset-0 z-[80] flex items-center justify-center bg-black/65 p-4"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) onClose()
      }}
    >
      <section
        role="dialog"
        aria-modal="true"
        aria-labelledby="advanced-export-title"
        className="w-full max-w-lg overflow-hidden rounded-xl border border-border bg-card shadow-2xl"
      >
        <header className="flex items-center justify-between border-b border-border px-5 py-4">
          <h2 id="advanced-export-title" className="font-mono text-sm font-semibold">{t('darkWebMonitoring.exportDialogs.advanced.title')}</h2>
          <button type="button" aria-label={t('darkWebMonitoring.exportDialogs.close')} onClick={onClose} className="text-muted-foreground hover:text-foreground">
            <X size={16} />
          </button>
        </header>
        <div className="space-y-5 px-6 py-5 font-mono text-xs">
          <div className="space-y-3">
            <label className="flex cursor-pointer items-center gap-2 text-muted-foreground">
              <input type="checkbox" checked={employeeOnly} onChange={(event) => setEmployeeOnly(event.target.checked)} className="h-4 w-4 accent-primary" />
              {t('darkWebMonitoring.exportDialogs.advanced.employeesOnly')}
            </label>
            <label className="flex cursor-pointer items-center gap-2 text-muted-foreground">
              <input type="checkbox" checked={maskSensitive} onChange={(event) => setMaskSensitive(event.target.checked)} className="h-4 w-4 accent-primary" />
              {t('darkWebMonitoring.exportDialogs.advanced.maskSensitive')}
            </label>
          </div>
          <label className="block space-y-1.5">
            <span className="font-semibold">{t('darkWebMonitoring.exportDialogs.action')}</span>
            <select value={action} onChange={(event) => setAction(event.target.value as ExportAction)} className="h-9 w-full rounded-md border border-input bg-background px-2 text-xs">
              <option value="CSV">CSV</option>
              <option value="PDF">PDF</option>
            </select>
          </label>
          <div className="space-y-2">
            <span className="font-semibold">{t('darkWebMonitoring.exportDialogs.advanced.additionalFilters')}</span>
            <div className="grid gap-2 sm:grid-cols-[1fr_1fr_1fr_auto]">
              <select aria-label={t('darkWebMonitoring.exportDialogs.advanced.field')} value={field} onChange={(event) => setField(event.target.value as ExportFilterField)} className="h-9 min-w-0 rounded-md border border-input bg-background px-2 text-xs">
                {FILTER_FIELDS.map((filterField) => <option key={filterField} value={filterField}>{t(`darkWebMonitoring.exportDialogs.fields.${filterField}`)}</option>)}
              </select>
              <select aria-label={t('darkWebMonitoring.exportDialogs.advanced.operator')} value={operator} onChange={(event) => setOperator(event.target.value as ExportFilterOperator)} className="h-9 min-w-0 rounded-md border border-input bg-background px-2 text-xs">
                {FILTER_OPERATORS.map((filterOperator) => <option key={filterOperator} value={filterOperator}>{t(`darkWebMonitoring.exportDialogs.operators.${filterOperator}`)}</option>)}
              </select>
              {field === 'organizationElementType' ? (
                <select aria-label={t('darkWebMonitoring.exportDialogs.advanced.value')} value={value} onChange={(event) => setValue(event.target.value)} className="h-9 min-w-0 rounded-md border border-input bg-background px-2 text-xs">
                  <option value="">{t('darkWebMonitoring.exportDialogs.advanced.chooseValue')}</option>
                  {ELEMENT_TYPES.map((elementType) => <option key={elementType} value={elementType}>{t(`darkWebMonitoring.exportDialogs.elementTypes.${elementType}`)}</option>)}
                </select>
              ) : field === 'breachStatus' ? (
                <select aria-label={t('darkWebMonitoring.exportDialogs.advanced.value')} value={value} onChange={(event) => setValue(event.target.value)} className="h-9 min-w-0 rounded-md border border-input bg-background px-2 text-xs">
                  <option value="">{t('darkWebMonitoring.exportDialogs.advanced.chooseValue')}</option>
                  {BREACH_STATUSES.map((breachStatus) => <option key={breachStatus} value={breachStatus}>{t(`darkWebMonitoring.status.${STATUS_KEYS[breachStatus]}`)}</option>)}
                </select>
              ) : field === 'employee' ? (
                <select aria-label={t('darkWebMonitoring.exportDialogs.advanced.value')} value={value} onChange={(event) => setValue(event.target.value)} className="h-9 min-w-0 rounded-md border border-input bg-background px-2 text-xs">
                  <option value="">{t('darkWebMonitoring.exportDialogs.advanced.chooseValue')}</option>
                  <option value="true">{t('darkWebMonitoring.employee')}</option>
                  <option value="false">{t('darkWebMonitoring.other')}</option>
                </select>
              ) : (
                <input aria-label={t('darkWebMonitoring.exportDialogs.advanced.value')} value={value} onChange={(event) => setValue(event.target.value)} className="h-9 min-w-0 rounded-md border border-input bg-background px-2 text-xs" />
              )}
              <Button type="button" size="icon" aria-label={t('darkWebMonitoring.exportDialogs.advanced.addFilter')} onClick={addFilter} disabled={!value.trim()}>
                <Plus size={14} />
              </Button>
            </div>
          </div>
          <div className="border-t border-border pt-3">
            <h3 className="font-semibold">{t('darkWebMonitoring.exportDialogs.advanced.addedFilters')}</h3>
            {filters.length === 0 ? (
              <p className="mt-2 text-[10px] text-muted-foreground">{t('darkWebMonitoring.exportDialogs.advanced.noFilters')}</p>
            ) : (
              <ul className="mt-2 space-y-2">
                {filters.map((filter, index) => (
                  <li key={`${filter.field}-${filter.operator}-${filter.value}-${index}`} className="flex items-center justify-between gap-2 rounded-md border border-border px-2 py-1.5">
                    <span className="min-w-0 truncate text-muted-foreground">
                      {t(`darkWebMonitoring.exportDialogs.fields.${filter.field}`)} {t(`darkWebMonitoring.exportDialogs.operators.${filter.operator}`)} {filter.value}
                    </span>
                    <button type="button" aria-label={t('darkWebMonitoring.exportDialogs.advanced.removeFilter')} onClick={() => setFilters((current) => current.filter((_, itemIndex) => itemIndex !== index))} className="rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground">
                      <Trash2 size={14} />
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
        <footer className="flex justify-end gap-3 px-6 pb-5">
          <Button variant="secondary" size="sm" onClick={onClose}>{t('darkWebMonitoring.exportDialogs.cancel')}</Button>
          <Button size="sm" onClick={handleExport}><Download size={13} className="mr-1.5" />{t('darkWebMonitoring.exportDialogs.advanced.export')}</Button>
        </footer>
      </section>
    </div>,
    document.body,
  )
}
