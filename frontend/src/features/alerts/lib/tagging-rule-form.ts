import { operatorById, RULE_FIELDS } from './tagging-rule-meta'
import type { AlertTag, FilterType, TaggingRule } from '../types/tagging-rule.types'

export interface FormState {
  name: string
  description: string
  conditions: FilterType[]
  tags: AlertTag[]
}

export function ruleToForm(
  rule?: TaggingRule | null,
  initialTags?: AlertTag[],
  initialConditions?: FilterType[]
): FormState {
  const conds = rule?.conditions?.length
    ? rule.conditions.map((c) => ({ ...c }))
    : initialConditions?.length
      ? initialConditions.map((c) => ({ ...c }))
      : [{ field: RULE_FIELDS[0].field, operator: 'IS', value: '' }]
  return {
    name: rule?.name ?? '',
    description: rule?.description ?? '',
    conditions: conds,
    tags: rule?.tags ? [...rule.tags] : initialTags ? [...initialTags] : [],
  }
}

/** Conditions are stored as FilterType. For multi-value operators (IS_ONE_OF
 * etc.) the backend accepts an array; we expose a comma-separated input and
 * convert at the boundary. */
export function serializeConditions(conds: FilterType[]): FilterType[] {
  return conds.map((c) => {
    const op = operatorById(c.operator)
    if (!op) return c
    if (op.needs === 'none') return { field: c.field, operator: c.operator }
    if (op.needs === 'list') {
      const raw = typeof c.value === 'string' ? c.value : Array.isArray(c.value) ? c.value.join(',') : ''
      const arr = raw
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean)
      return { field: c.field, operator: c.operator, value: arr }
    }
    return { field: c.field, operator: c.operator, value: c.value ?? '' }
  })
}

export function deserializeConditions(conds: FilterType[]): FilterType[] {
  return conds.map((c) => {
    const op = operatorById(c.operator)
    if (op?.needs === 'list' && Array.isArray(c.value)) {
      return { ...c, value: (c.value as unknown[]).join(', ') }
    }
    return { ...c }
  })
}
