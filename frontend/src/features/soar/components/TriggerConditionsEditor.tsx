import { useTranslation } from 'react-i18next'
import { Plus, X } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { cn } from '@/shared/lib/utils'
import { ALERT_FIELDS } from '../lib/alert-fields'
import {
  SOAR_MULTI_VALUE_OPERATORS,
  SOAR_NO_VALUE_OPERATORS,
  SOAR_OPERATORS,
  type FlowCondition,
  type SoarOperator,
} from '../types/soar.types'

const SELECT =
  'h-8 rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

interface Props {
  conditions: FlowCondition[]
  readOnly?: boolean
  onChange: (c: FlowCondition[]) => void
}

// Trigger-condition list: same field/op/value pattern reused across the app,
// but here it's the top-level alert-match, not a node-scoped predicate.
// ponytail: extracted from FlowEditor so the trigger inspector can host it.
export function TriggerConditionsEditor({ conditions, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const setAt = (i: number, patch: Partial<FlowCondition>) =>
    onChange(conditions.map((c, k) => (k === i ? { ...c, ...patch } : c)))
  const valueStr = (c: FlowCondition) =>
    Array.isArray(c.value) ? c.value.join(', ') : c.value == null ? '' : String(c.value)

  return (
    <div className="space-y-1.5">
      <p className="text-[11px] text-muted-foreground">{t('soar.editor.conditionsHint')}</p>
      <div className="space-y-1.5">
        {conditions.map((c, i) => (
          <div key={i} className="flex flex-wrap items-center gap-1.5">
            <select
              value={c.field}
              disabled={readOnly}
              onChange={(e) => setAt(i, { field: e.target.value })}
              className={cn(SELECT, 'min-w-[150px] flex-1')}
            >
              <option value="">{t('soar.editor.selectField')}</option>
              {ALERT_FIELDS.map((af) => (
                <option key={af.field} value={af.field}>
                  {af.label}
                </option>
              ))}
              {c.field && !ALERT_FIELDS.some((af) => af.field === c.field) && (
                <option value={c.field}>{c.field}</option>
              )}
            </select>
            <select
              value={c.operator}
              disabled={readOnly}
              onChange={(e) => setAt(i, { operator: e.target.value as SoarOperator })}
              className={SELECT}
            >
              {SOAR_OPERATORS.map((op) => (
                <option key={op} value={op}>
                  {t(`soar.operator.${op}`)}
                </option>
              ))}
            </select>
            {!SOAR_NO_VALUE_OPERATORS.includes(c.operator) && (
              <Input
                value={valueStr(c)}
                readOnly={readOnly}
                onChange={(e) => setAt(i, { value: e.target.value })}
                placeholder={
                  SOAR_MULTI_VALUE_OPERATORS.includes(c.operator)
                    ? t('soar.editor.valueList')
                    : t('soar.editor.value')
                }
                className="h-8 min-w-[140px] flex-1 font-mono text-xs"
              />
            )}
            {!readOnly && (
              <button
                type="button"
                onClick={() => onChange(conditions.filter((_, k) => k !== i))}
                className="rounded p-1 text-muted-foreground hover:text-red-500"
              >
                <X size={13} />
              </button>
            )}
          </div>
        ))}
      </div>
      {!readOnly && (
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="h-7"
          onClick={() => onChange([...conditions, { operator: 'IS', field: '', value: '' }])}
        >
          <Plus size={12} className="mr-1" /> {t('soar.editor.addCondition')}
        </Button>
      )}
    </div>
  )
}
