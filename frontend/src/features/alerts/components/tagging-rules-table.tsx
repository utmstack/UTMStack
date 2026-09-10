import { useTranslation } from 'react-i18next'
import { ResizableGridHeader } from '@/shared/components/ui/resizable-grid-header'
import { useResizableColumns } from '@/shared/hooks/useResizableColumns'
import type { TaggingRule } from '../types/tagging-rule.types'
import { TaggingRulesTableRow } from './tagging-rules-table-row'

export const TAGGING_RULES_TABLE_COLS = ['1.6fr', '0.9fr', 90]

export function TaggingRulesTable({
  rules,
  onOpen,
}: {
  rules: TaggingRule[]
  onOpen: (rule: TaggingRule) => void
}) {
  const { t } = useTranslation()
  const { template: tableCols, startDrag } = useResizableColumns(TAGGING_RULES_TABLE_COLS, {
    min: 60,
    storageKey: 'tagging-rules-table-columns',
  })
  return (
    <div className="mt-4 min-h-0 flex-1 overflow-x-auto overflow-y-auto rounded-xl border border-border">
      <ResizableGridHeader
        headers={[t('taggingRules.table.rule'), t('taggingRules.table.tags'), t('taggingRules.table.conditions')]}
        tableCols={tableCols}
        startDrag={startDrag}
        className="bg-muted/30 py-2.5 font-medium"
        cellClassName="last:text-center"
      />
      {rules.map((rule) => (
        <TaggingRulesTableRow key={rule.id} rule={rule} tableCols={tableCols} onOpen={onOpen} />
      ))}
    </div>
  )
}
