import { Trash2 } from 'lucide-react'
import type { TFunction } from 'i18next'
import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/components/ui/input'
import {
  RULE_FIELDS,
  RULE_OPERATORS,
  SELECT_CLS,
  operatorById,
  type TaggingOperator,
} from '../lib/tagging-rule-meta'
import type { FilterType } from '../types/tagging-rule.types'

export function TaggingRuleConditionRow({
  cond,
  onChange,
  onRemove,
  t,
}: {
  cond: FilterType
  onChange: (patch: Partial<FilterType>) => void
  onRemove: () => void
  t: TFunction
}) {
  const op: TaggingOperator = operatorById(cond.operator) ?? RULE_OPERATORS[0]
  const value = typeof cond.value === 'string' ? cond.value : Array.isArray(cond.value) ? (cond.value as unknown[]).join(', ') : ''
  return (
    <div className="flex flex-wrap items-center gap-1.5 rounded-md border border-input bg-background p-2">
      <select
        value={cond.field}
        onChange={(e) => onChange({ field: e.target.value })}
        className={cn(SELECT_CLS, 'min-w-[170px] flex-1')}
      >
        {RULE_FIELDS.map((f) => (
          <option key={f.field} value={f.field}>
            {f.label}
          </option>
        ))}
      </select>
      <select
        value={cond.operator}
        onChange={(e) => onChange({ operator: e.target.value, value: '' })}
        className={cn(SELECT_CLS, 'min-w-[150px]')}
      >
        {RULE_OPERATORS.map((o) => (
          <option key={o.id} value={o.id}>
            {o.label}
          </option>
        ))}
      </select>
      {op.needs !== 'none' && (
        <Input
          value={value}
          onChange={(e) => onChange({ value: e.target.value })}
          placeholder={
            op.needs === 'list' ? t('taggingRules.form.valueListPlaceholder') : t('taggingRules.form.valuePlaceholder')
          }
          className="h-9 min-w-[160px] flex-1"
        />
      )}
      <button
        type="button"
        onClick={onRemove}
        title={t('taggingRules.form.removeCondition')}
        className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
      >
        <Trash2 size={14} />
      </button>
    </div>
  )
}
