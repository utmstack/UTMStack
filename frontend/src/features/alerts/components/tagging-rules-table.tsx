import { useTranslation } from 'react-i18next'
import type { TaggingRule } from '../types/tagging-rule.types'
import { TaggingRulesTableRow } from './tagging-rules-table-row'

export const TAGGING_RULES_TABLE_COLS = '1.6fr 0.9fr 90px'

export function TaggingRulesTable({
  rules,
  onOpen,
}: {
  rules: TaggingRule[]
  onOpen: (rule: TaggingRule) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="mt-4 min-h-0 flex-1 overflow-y-auto rounded-xl border border-border">
      <div
        className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
        style={{ gridTemplateColumns: TAGGING_RULES_TABLE_COLS }}
      >
        <div>{t('taggingRules.table.rule')}</div>
        <div>{t('taggingRules.table.tags')}</div>
        <div className="text-center">{t('taggingRules.table.conditions')}</div>
      </div>
      {rules.map((rule) => (
        <TaggingRulesTableRow key={rule.id} rule={rule} onOpen={onOpen} />
      ))}
    </div>
  )
}
