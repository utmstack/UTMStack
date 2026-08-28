import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Plus, X } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { ALERT_FIELDS } from '../lib/alert-fields'
import { enrichmentAncestors } from '../lib/ancestors'
import {
  SOAR_MULTI_VALUE_OPERATORS,
  SOAR_NO_VALUE_OPERATORS,
  SOAR_OPERATORS,
  type FlowCondition,
  type FlowNode,
  type SoarOperator,
} from '../types/soar.types'

const SELECT =
  'h-8 rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

interface Props {
  nodeId: string
  nodes: Record<string, FlowNode>
  params: unknown
  readOnly?: boolean
  onChange: (params: { conditions: FlowCondition[] }) => void
}

// ponytail: reuses SoarOperator + native <datalist>; success/fail routing rides
// on the DAG's existing green/red handles — no bespoke branch state.
export function ConditionalParamsEditor({ nodeId, nodes, params, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const conditions = normalize(params)
  const listId = `soar-cond-fields-${nodeId}`
  const suggestions = useMemo(() => buildSuggestions(nodes, nodeId), [nodes, nodeId])

  const setAt = (i: number, patch: Partial<FlowCondition>) =>
    onChange({ conditions: conditions.map((c, k) => (k === i ? { ...c, ...patch } : c)) })

  const valueStr = (c: FlowCondition) =>
    Array.isArray(c.value) ? c.value.join(', ') : c.value == null ? '' : String(c.value)

  return (
    <div className="space-y-1.5">
      <p className="text-[11px] text-muted-foreground">{t('soar.editor.canvas.conditionalHint')}</p>
      <datalist id={listId}>
        {suggestions.map((s) => (
          <option key={s} value={s} />
        ))}
      </datalist>
      <div className="space-y-1.5">
        {conditions.map((c, i) => (
          <div key={i} className="flex flex-wrap items-center gap-1.5">
            <Input
              list={listId}
              value={c.field}
              readOnly={readOnly}
              onChange={(e) => setAt(i, { field: e.target.value })}
              placeholder="alert.severity"
              className="h-8 min-w-[160px] flex-1 font-mono text-[11px]"
            />
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
                className="h-8 min-w-[140px] flex-1 font-mono text-[11px]"
              />
            )}
            {!readOnly && (
              <button
                type="button"
                onClick={() => onChange({ conditions: conditions.filter((_, k) => k !== i) })}
                className="rounded p-1 text-muted-foreground hover:text-red-500"
                title={t('soar.editor.canvas.deleteNode')}
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
          onClick={() =>
            onChange({ conditions: [...conditions, { operator: 'IS', field: '', value: '' }] })
          }
        >
          <Plus size={12} className="mr-1" /> {t('soar.editor.addCondition')}
        </Button>
      )}
    </div>
  )
}

function normalize(params: unknown): FlowCondition[] {
  if (!params || typeof params !== 'object') return []
  const list = (params as { conditions?: unknown }).conditions
  return Array.isArray(list) ? (list as FlowCondition[]) : []
}

// Suggestion list for the field <datalist>: alert.* plus every reachable
// enrichment ancestor's declared fields. Paths match the runtime context bag.
function buildSuggestions(nodes: Record<string, FlowNode>, currentNodeId: string): string[] {
  const out: string[] = ALERT_FIELDS.map((af) => `alert.${af.field}`)
  for (const a of enrichmentAncestors(nodes, currentNodeId)) {
    if (a.fields.length) out.push(...a.fields.map((f) => `${a.nodeId}.${f}`))
    else out.push(`${a.nodeId}.`)
  }
  return out
}
