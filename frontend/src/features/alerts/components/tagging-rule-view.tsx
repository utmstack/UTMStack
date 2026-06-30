import { Tag as TagIcon } from 'lucide-react'
import type { TFunction } from 'i18next'
import type { useDateFormat } from '@/shared/lib/datetime'
import { operatorById } from '../lib/tagging-rule-meta'
import type { TaggingRule } from '../types/tagging-rule.types'
import { TaggingRuleSection } from './tagging-rule-section'
import { TaggingRuleRow } from './tagging-rule-row'

export function TaggingRuleView({
  rule,
  df,
  t,
}: {
  rule: TaggingRule
  df: ReturnType<typeof useDateFormat>
  t: TFunction
}) {
  return (
    <div className="space-y-4">
      {rule.description && (
        <TaggingRuleSection title={t('taggingRules.view.description')}>
          <p className="text-xs leading-relaxed text-muted-foreground">{rule.description}</p>
        </TaggingRuleSection>
      )}

      <TaggingRuleSection title={t('taggingRules.view.tags')}>
        <div className="flex flex-wrap gap-1.5">
          {rule.tags.length === 0 ? (
            <span className="text-xs text-muted-foreground">—</span>
          ) : (
            rule.tags.map((tg) => (
              <span
                key={tg.id}
                className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px]"
                style={{
                  backgroundColor: (tg.tagColor || '#64748b') + '22',
                  color: tg.tagColor || '#64748b',
                }}
              >
                <TagIcon size={10} /> {tg.tagName}
              </span>
            ))
          )}
        </div>
      </TaggingRuleSection>

      <TaggingRuleSection title={t('taggingRules.view.conditions')}>
        {rule.conditions.length === 0 ? (
          <span className="text-xs text-muted-foreground">—</span>
        ) : (
          <ul className="space-y-1.5">
            {rule.conditions.map((c, i) => (
              <li key={i} className="flex flex-wrap items-center gap-1.5 text-xs">
                <span className="rounded bg-background px-1.5 py-0.5 font-mono text-[11px]">{c.field}</span>
                <span className="text-[11px] text-muted-foreground">
                  {operatorById(c.operator)?.label || c.operator}
                </span>
                {c.value !== undefined && c.value !== '' && (
                  <span className="rounded bg-primary/10 px-1.5 py-0.5 font-mono text-[11px] text-primary">
                    {Array.isArray(c.value) ? (c.value as unknown[]).join(', ') : String(c.value)}
                  </span>
                )}
              </li>
            ))}
          </ul>
        )}
      </TaggingRuleSection>

      <TaggingRuleSection title={t('taggingRules.view.audit')}>
        <dl className="grid grid-cols-[150px_1fr] gap-y-2 text-xs">
          <TaggingRuleRow k={t('taggingRules.view.createdBy')}>{rule.createdBy || '—'}</TaggingRuleRow>
          <TaggingRuleRow k={t('taggingRules.view.createdDate')}>
            {rule.createdDate ? df.formatDateTime(rule.createdDate) : '—'}
          </TaggingRuleRow>
          {rule.lastModifiedBy && <TaggingRuleRow k={t('taggingRules.view.modifiedBy')}>{rule.lastModifiedBy}</TaggingRuleRow>}
          {rule.lastModifiedDate && (
            <TaggingRuleRow k={t('taggingRules.view.modifiedDate')}>{df.formatDateTime(rule.lastModifiedDate)}</TaggingRuleRow>
          )}
        </dl>
      </TaggingRuleSection>
    </div>
  )
}
