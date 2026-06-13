import { useTranslation } from 'react-i18next'
import { Tag as TagIcon } from 'lucide-react'
import { useDateFormat } from '@/shared/lib/datetime'
import type { TaggingRule } from '../types/tagging-rule.types'

const COLS = '1.6fr 0.9fr 90px 140px 140px'

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
        style={{ gridTemplateColumns: COLS }}
      >
        <div>{t('taggingRules.table.rule')}</div>
        <div>{t('taggingRules.table.tags')}</div>
        <div className="text-center">{t('taggingRules.table.conditions')}</div>
        <div>{t('taggingRules.table.modifiedBy')}</div>
        <div>{t('taggingRules.table.modifiedDate')}</div>
      </div>
      {rules.map((rule) => (
        <RuleRow key={rule.id} rule={rule} onOpen={onOpen} />
      ))}
    </div>
  )
}

function RuleRow({
  rule,
  onOpen,
}: {
  rule: TaggingRule
  onOpen: (rule: TaggingRule) => void
}) {
  const df = useDateFormat()
  const tags = rule.tags ?? []
  const date = rule.lastModifiedDate ?? rule.createdDate
  const by = rule.lastModifiedBy || rule.createdBy
  return (
    <div
      onClick={() => onOpen(rule)}
      className="grid cursor-pointer items-center gap-3 border-b border-border/60 px-4 py-3 text-sm last:border-b-0 hover:bg-muted/30"
      style={{ gridTemplateColumns: COLS }}
    >
      <div className="min-w-0">
        <div className="truncate font-medium">{rule.name}</div>
        {rule.description && <div className="truncate text-xs text-muted-foreground">{rule.description}</div>}
      </div>
      <div className="flex min-w-0 flex-wrap items-center gap-1">
        {tags.slice(0, 3).map((tg) => (
          <span
            key={tg.id}
            className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px]"
            style={{ backgroundColor: (tg.tagColor || '#64748b') + '22', color: tg.tagColor || '#64748b' }}
          >
            <TagIcon size={10} /> {tg.tagName}
          </span>
        ))}
        {tags.length > 3 && <span className="text-[10px] text-muted-foreground">+{tags.length - 3}</span>}
        {tags.length === 0 && <span className="text-xs text-muted-foreground/60">—</span>}
      </div>
      <div className="text-center font-mono text-xs text-muted-foreground">{rule.conditions?.length ?? 0}</div>
      <div className="truncate text-xs text-muted-foreground">{by || '—'}</div>
      <div className="text-xs text-muted-foreground">{date ? df.formatDateTime(date) : '—'}</div>
    </div>
  )
}
