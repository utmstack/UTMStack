import type { ReactNode } from 'react'
import type { TFunction } from 'i18next'
import type { ColumnDef } from '@tanstack/react-table'
import { Crosshair, Lock } from 'lucide-react'
import { ResizableDataTable } from '@/shared/components/ui/resizable-data-table'
import { cn } from '@/shared/lib/utils'
import type { CorrelationRule } from '../services/alerting-rules-http.service'
import { impactKey } from '../lib/impact-key'
import { maxImpact } from '../lib/max-impact'
import { DataTypeChip } from './rule-form'
import { SelectAllCheckbox } from './select-all-checkbox'
import { Toggle } from './toggle'

const TH = 'whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-2.5 align-middle'
const IMPACT_TONE: Record<string, string> = { high: 'text-red-500', medium: 'text-amber-500', low: 'text-sky-500', none: 'text-muted-foreground' }

const keepRowClosed = { onClick: (e: { stopPropagation: () => void }) => e.stopPropagation() }

interface ColumnOptions {
  t: TFunction
  selected: Set<string>
  allChecked: boolean
  someChecked: boolean
  onToggleSelected: (relPath: string) => void
  onSelectAll: (checked: boolean) => void
  onToggle: (r: CorrelationRule, next: boolean) => void
}

function buildRuleColumns({ t, selected, allChecked, someChecked, onToggleSelected, onSelectAll, onToggle }: ColumnOptions): ColumnDef<CorrelationRule>[] {
  const muted = `${TD} text-xs text-muted-foreground`
  return [
    {
      id: 'select',
      header: () => (
        <div className="flex justify-center">
          <SelectAllCheckbox checked={allChecked} indeterminate={someChecked} onChange={onSelectAll} label={t('alertingRules.table.selectAll')} />
        </div>
      ),
      size: 48,
      minSize: 48,
      enableResizing: false,
      meta: { headerClassName: `${TH} px-0`, cellClassName: `${TD} px-0`, cellProps: () => keepRowClosed },
      cell: ({ row }) => (
        <div className="flex justify-center">
          <input
            type="checkbox"
            checked={selected.has(row.original.relPath)}
            onChange={() => onToggleSelected(row.original.relPath)}
            aria-label={t('alertingRules.table.selectRow', { name: row.original.name })}
            className="h-4 w-4 cursor-pointer accent-primary"
          />
        </div>
      ),
    },
    {
      id: 'name',
      header: t('alertingRules.table.name'),
      size: 360,
      minSize: 140,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const r = row.original
        return (
          <>
            <div className="flex items-center gap-1.5">
              <span className="truncate font-medium">{r.name}</span>
              {r.systemOwner && <Lock size={11} className="shrink-0 text-muted-foreground/50" />}
            </div>
            {r.description && <div className="truncate text-muted-foreground">{r.description}</div>}
          </>
        )
      },
    },
    {
      id: 'dataTypes',
      header: t('alertingRules.table.dataTypes'),
      size: 220,
      minSize: 100,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const dts = (row.original.dataTypes ?? []).filter((d) => d.included).map((d) => d.dataType)
        return (
          <div className="flex items-center gap-1">
            {dts.slice(0, 2).map((dt) => <DataTypeChip key={dt} dataType={dt} />)}
            {dts.length > 2 && <span className="text-[10px] text-muted-foreground">+{dts.length - 2}</span>}
            {dts.length === 0 && <span className="text-xs text-muted-foreground/60">—</span>}
          </div>
        )
      },
    },
    {
      id: 'category',
      header: t('alertingRules.table.category'),
      size: 160,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: muted, cellProps: (r) => ({ title: r.category }) },
      cell: ({ row }) => <span className="block truncate">{row.original.category || '—'}</span>,
    },
    {
      id: 'technique',
      header: t('alertingRules.table.technique'),
      size: 180,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: `${muted} font-mono`, cellProps: (r) => ({ title: r.technique }) },
      cell: ({ row }) => <span className="block truncate">{row.original.technique || '—'}</span>,
    },
    {
      id: 'adversary',
      header: t('alertingRules.table.adversary'),
      size: 160,
      minSize: 90,
      meta: { headerClassName: TH, cellClassName: muted },
      cell: ({ row }) => (
        <div className="flex items-center gap-1">
          <Crosshair size={11} className="shrink-0" />
          <span className="truncate">{row.original.adversary ? t(`alertingRules.adversary.${row.original.adversary}`) : '—'}</span>
        </div>
      ),
    },
    {
      id: 'impact',
      header: t('alertingRules.table.impact'),
      size: 90,
      minSize: 70,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => (
        <span className={cn('font-mono text-xs font-semibold', IMPACT_TONE[impactKey(maxImpact(row.original))])}>{maxImpact(row.original)}</span>
      ),
    },
    {
      id: 'active',
      header: t('alertingRules.table.active'),
      size: 90,
      minSize: 70,
      meta: { headerClassName: TH, cellClassName: TD, cellProps: () => keepRowClosed },
      cell: ({ row }) => <Toggle on={row.original.ruleActive} onChange={(v) => onToggle(row.original, v)} />,
    },
  ]
}

export function Table({ rules, selected, onToggleSelected, onSelectAll, onOpen, onToggle, t, footer }: { rules: CorrelationRule[]; selected: Set<string>; onToggleSelected: (relPath: string) => void; onSelectAll: (checked: boolean) => void; onOpen: (r: CorrelationRule) => void; onToggle: (r: CorrelationRule, next: boolean) => void; t: TFunction; footer?: ReactNode }) {
  const allChecked = rules.length > 0 && rules.every((r) => selected.has(r.relPath))
  const someChecked = !allChecked && rules.some((r) => selected.has(r.relPath))
  return (
    <div className="mt-4 min-h-0 flex-1 overflow-auto rounded-xl border border-border">
      <ResizableDataTable
        columns={buildRuleColumns({ t, selected, allChecked, someChecked, onToggleSelected, onSelectAll, onToggle })}
        data={rules}
        flexColumnId="name"
        storageKey="alerting-rules-table-sizing"
        getRowId={(r) => r.relPath}
        onRowClick={onOpen}
        rowClassName={() => 'last:border-b-0'}
      />
      {footer}
    </div>
  )
}
