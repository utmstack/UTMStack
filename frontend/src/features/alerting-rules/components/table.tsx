import type { ReactNode } from 'react'
import type { TFunction } from 'i18next'
import { Crosshair, Lock } from 'lucide-react'
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

export function Table({ rules, selected, onToggleSelected, onSelectAll, onOpen, onToggle, t, footer }: { rules: CorrelationRule[]; selected: Set<string>; onToggleSelected: (relPath: string) => void; onSelectAll: (checked: boolean) => void; onOpen: (r: CorrelationRule) => void; onToggle: (r: CorrelationRule, next: boolean) => void; t: TFunction; footer?: ReactNode }) {
  const allChecked = rules.length > 0 && rules.every((r) => selected.has(r.relPath))
  const someChecked = !allChecked && rules.some((r) => selected.has(r.relPath))
  return (
    <div className="mt-4 min-h-0 flex-1 overflow-auto rounded-xl border border-border">
      <table className="min-w-full border-collapse">
        <thead className="sticky top-0 z-10 bg-muted/90 text-[10px] uppercase tracking-wider text-muted-foreground">
          <tr className="border-b border-border">
            <th className={`${TH} text-center`}>
              <div className="flex justify-center">
                <SelectAllCheckbox checked={allChecked} indeterminate={someChecked} onChange={onSelectAll} label={t('alertingRules.table.selectAll')} />
              </div>
            </th>
            <th className={TH}>{t('alertingRules.table.name')}</th>
            <th className={TH}>{t('alertingRules.table.dataTypes')}</th>
            <th className={TH}>{t('alertingRules.table.category')}</th>
            <th className={TH}>{t('alertingRules.table.technique')}</th>
            <th className={TH}>{t('alertingRules.table.adversary')}</th>
            <th className={`${TH} text-center`}>{t('alertingRules.table.impact')}</th>
            <th className={`${TH} text-center`}>{t('alertingRules.table.active')}</th>
          </tr>
        </thead>
        <tbody>
          {rules.map((r) => {
            const dts = (r.dataTypes ?? []).filter((d) => d.included).map((d) => d.dataType)
            return (
              <tr key={r.relPath} className="border-b border-border/60 text-sm last:border-b-0 hover:bg-muted/30">
                <td className={`${TD} text-center`}>
                  <div className="flex justify-center">
                    <input
                      type="checkbox"
                      checked={selected.has(r.relPath)}
                      onChange={() => onToggleSelected(r.relPath)}
                      onClick={(e) => e.stopPropagation()}
                      aria-label={t('alertingRules.table.selectRow', { name: r.name })}
                      className="h-4 w-4 cursor-pointer accent-primary"
                    />
                  </div>
                </td>
                <td className={`${TD} max-w-[360px]`}>
                  <button onClick={() => onOpen(r)} className="min-w-0 text-left">
                    <div className="flex items-center gap-1.5">
                      <span className="truncate font-medium">{r.name}</span>
                      {r.systemOwner && <Lock size={11} className="shrink-0 text-muted-foreground/50" />}
                    </div>
                    {r.description && <div className="truncate max-w-[18rem] text-muted-foreground">{r.description}</div>}
                  </button>
                </td>
                <td className={TD}>
                  <button onClick={() => onOpen(r)} className="flex min-w-0 flex-wrap items-center gap-1 text-left">
                    {dts.slice(0, 2).map((dt) => <DataTypeChip key={dt} dataType={dt} />)}
                    {dts.length > 2 && <span className="text-[10px] text-muted-foreground">+{dts.length - 2}</span>}
                    {dts.length === 0 && <span className="text-xs text-muted-foreground/60">—</span>}
                  </button>
                </td>
                <td className={`${TD} max-w-[200px]`}>
                  <button onClick={() => onOpen(r)} className="block w-full truncate text-left text-xs text-muted-foreground">{r.category || '—'}</button>
                </td>
                <td className={`${TD} max-w-[200px]`}>
                  <button onClick={() => onOpen(r)} className="block w-full truncate text-left font-mono text-xs text-muted-foreground" title={r.technique}>{r.technique || '—'}</button>
                </td>
                <td className={TD}>
                  <div className="flex items-center gap-1 text-xs text-muted-foreground"><Crosshair size={11} /> {r.adversary ? t(`alertingRules.adversary.${r.adversary}`) : '—'}</div>
                </td>
                <td className={`${TD} text-center`}>
                  <span className={cn('font-mono text-xs font-semibold', IMPACT_TONE[impactKey(maxImpact(r))])}>{maxImpact(r)}</span>
                </td>
                <td className={`${TD} text-center`}>
                  <div className="flex justify-center"><Toggle on={r.ruleActive} onChange={(v) => onToggle(r, v)} /></div>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
      {footer}
    </div>
  )
}
