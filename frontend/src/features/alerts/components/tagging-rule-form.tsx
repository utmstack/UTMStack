import { Plus } from 'lucide-react'
import type { TFunction } from 'i18next'
import { Input } from '@/shared/components/ui/input'
import { RULE_FIELDS, operatorById } from '../lib/tagging-rule-meta'
import type { FormState } from '../lib/tagging-rule-form'
import type { AlertTag, FilterType } from '../types/tagging-rule.types'
import { TaggingRuleSection } from './tagging-rule-section'
import { TaggingRuleConditionRow } from './tagging-rule-condition-row'
import { TaggingRuleTagPicker } from './tagging-rule-tag-picker'

export function TaggingRuleForm({
  form,
  setForm,
  tagCatalog,
  onCreateTag,
  t,
}: {
  form: FormState
  setForm: (f: FormState) => void
  tagCatalog: AlertTag[]
  onCreateTag: (tagName: string, tagColor: string) => Promise<AlertTag | null>
  t: TFunction
}) {
  const addCondition = () =>
    setForm({
      ...form,
      conditions: [...form.conditions, { field: RULE_FIELDS[0].field, operator: 'IS', value: '' }],
    })

  const updateCondition = (i: number, patch: Partial<FilterType>) => {
    const next = form.conditions.map((c, idx) => {
      if (idx !== i) return c
      const merged = { ...c, ...patch }
      // Re-default value when switching to a value-less op so we don't send stale text.
      if (patch.operator) {
        const nextOp = operatorById(patch.operator)
        if (nextOp?.needs === 'none') return { field: merged.field, operator: merged.operator }
        if (!('value' in patch)) return { ...merged, value: merged.value ?? '' }
      }
      return merged
    })
    setForm({ ...form, conditions: next })
  }

  const removeCondition = (i: number) =>
    setForm({ ...form, conditions: form.conditions.filter((_, idx) => idx !== i) })

  return (
    <div className="space-y-4">
      <TaggingRuleSection title={t('taggingRules.form.basics')}>
        <div className="space-y-2">
          <label className="block text-xs">
            <span className="mb-1 block text-muted-foreground">{t('taggingRules.form.name')}</span>
            <Input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder={t('taggingRules.form.namePlaceholder')}
              className="h-9"
            />
          </label>
          <label className="block text-xs">
            <span className="mb-1 block text-muted-foreground">{t('taggingRules.form.description')}</span>
            <textarea
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              placeholder={t('taggingRules.form.descriptionPlaceholder')}
              maxLength={512}
              rows={3}
              className="w-full resize-none rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </label>
        </div>
      </TaggingRuleSection>

      <TaggingRuleSection title={t('taggingRules.form.tags')}>
        <TaggingRuleTagPicker
          selected={form.tags}
          catalog={tagCatalog}
          onChange={(tags) => setForm({ ...form, tags })}
          onCreateTag={onCreateTag}
          t={t}
        />
      </TaggingRuleSection>

      <TaggingRuleSection
        title={t('taggingRules.form.conditions')}
        right={
          <button
            type="button"
            onClick={addCondition}
            className="inline-flex items-center gap-1 rounded-md border border-input bg-background px-2 py-1 text-xs hover:bg-muted"
          >
            <Plus size={12} /> {t('taggingRules.form.addCondition')}
          </button>
        }
      >
        <div className="space-y-2">
          {form.conditions.map((c, i) => (
            <TaggingRuleConditionRow
              key={i}
              cond={c}
              onChange={(patch) => updateCondition(i, patch)}
              onRemove={() => removeCondition(i)}
              t={t}
            />
          ))}
          {form.conditions.length === 0 && (
            <div className="text-xs text-muted-foreground">{t('taggingRules.form.noConditions')}</div>
          )}
        </div>
      </TaggingRuleSection>
    </div>
  )
}
