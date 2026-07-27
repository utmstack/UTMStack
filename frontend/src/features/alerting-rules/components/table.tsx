import type { TFunction } from 'i18next'
import { Crosshair, Lock } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import type { CorrelationRule } from '../services/alerting-rules-http.service'
import { impactKey } from '../lib/impact-key'
import { maxImpact } from '../lib/max-impact'
import { DataTypeChip } from './rule-form'
import { SelectAllCheckbox } from './select-all-checkbox'
import { Toggle } from './toggle'

const COLS = '32px 1.4fr 1fr 110px 100px 80px 48px 50px'
const IMPACT_TONE: Record<string, string> = { high: 'text-red-500', medium: 'text-amber-500', low: 'text-sky-500', none: 'text-muted-foreground' }

export function Table({ rules, selected, onToggleSelected, onSelectAll, onOpen, onToggle, t }: { rules: CorrelationRule[]; selected: Set<string>; onToggleSelected: (relPath: string) => void; onSelectAll: (checked: boolean) => void; onOpen: (r: CorrelationRule) => void; onToggle: (r: CorrelationRule, next: boolean) => void; t: TFunction }) {
  const allChecked = rules.length > 0 && rules.every((r) => selected.has(r.relPath))
  const someChecked = !allChecked && rules.some((r) => selected.has(r.relPath))
  return (
    <div className="mt-4 min-h-0 flex-1 overflow-y-auto rounded-xl border border-border">
      <div className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground" style={{ gridTemplateColumns: COLS }}>
        <div className="flex justify-center">
          <SelectAllCheckbox checked={allChecked} indeterminate={someChecked} onChange={onSelectAll} label={t('alertingRules.table.selectAll')} />
        </div>
        <div>{t('alertingRules.table.name')}</div>
        <div>{t('alertingRules.table.dataTypes')}</div>
        <div>{t('alertingRules.table.category')}</div>
        <div>{t('alertingRules.table.technique')}</div>
        <div>{t('alertingRules.table.adversary')}</div>
        <div className="text-center">{t('alertingRules.table.impact')}</div>
        <div className="text-center">{t('alertingRules.table.active')}</div>
      </div>
      {rules.map((r) => {
        const dts = (r.dataTypes ?? []).filter((d) => d.included).map((d) => d.dataType)
        return (
        <div key={r.relPath} className="grid items-center gap-3 border-b border-border/60 px-4 py-3 text-sm last:border-b-0 hover:bg-muted/30" style={{ gridTemplateColumns: COLS }}>
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
          <button onClick={() => onOpen(r)} className="min-w-0 text-left">
            <div className="flex items-center gap-1.5">
              <span className="truncate font-medium">{r.name}</span>
              {r.systemOwner && <Lock size={11} className="shrink-0 text-muted-foreground/50" />}
            </div>
            {r.description && <div className="truncate text-xs text-muted-foreground">{r.description}</div>}
          </button>
          <button onClick={() => onOpen(r)} className="flex min-w-0 flex-wrap items-center gap-1 text-left">
            {dts.slice(0, 2).map((dt) => <DataTypeChip key={dt} dataType={dt} />)}
            {dts.length > 2 && <span className="text-[10px] text-muted-foreground">+{dts.length - 2}</span>}
            {dts.length === 0 && <span className="text-xs text-muted-foreground/60">—</span>}
          </button>
          <button onClick={() => onOpen(r)} className="truncate text-left text-xs text-muted-foreground">{r.category || '—'}</button>
          <button onClick={() => onOpen(r)} className="truncate text-left font-mono text-xs text-muted-foreground" title={r.technique}>{r.technique || '—'}</button>
          <div className="flex items-center gap-1 text-xs text-muted-foreground"><Crosshair size={11} /> {r.adversary ? t(`alertingRules.adversary.${r.adversary}`) : '—'}</div>
          <div className="text-center"><span className={cn('font-mono text-xs font-semibold', IMPACT_TONE[impactKey(maxImpact(r))])}>{maxImpact(r)}</span></div>
          <div className="flex justify-center"><Toggle on={r.ruleActive} onChange={(v) => onToggle(r, v)} /></div>
        </div>
        )
      })}
    </div>
  )
}
