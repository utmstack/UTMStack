import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import type { ColumnDef } from '@tanstack/react-table'
import { Tag as TagIcon } from 'lucide-react'
import { ResizableDataTable } from '@/shared/components/ui/resizable-data-table'
import type { TaggingRule } from '../types/tagging-rule.types'

const TH = 'whitespace-nowrap px-3 py-2.5 text-left align-middle font-medium'
const TD = 'whitespace-nowrap px-3 py-3 align-middle'

export function TaggingRulesTable({
  rules,
  onOpen,
}: {
  rules: TaggingRule[]
  onOpen: (rule: TaggingRule) => void
}) {
  const { t } = useTranslation()
  const columns = useMemo(() => buildColumns(t), [t])
  return (
    <div className="mt-4 min-h-0 flex-1 overflow-x-auto overflow-y-auto rounded-xl border border-border">
      <ResizableDataTable
        columns={columns}
        data={rules}
        flexColumnId="rule"
        storageKey="tagging-rules-table-sizing"
        getRowId={(rule) => rule.id}
        onRowClick={onOpen}
        rowClassName={() => 'border-border/60 last:border-b-0 hover:bg-muted/30'}
      />
    </div>
  )
}

function buildColumns(t: TFunction): ColumnDef<TaggingRule>[] {
  return [
    {
      id: 'rule',
      header: t('taggingRules.table.rule'),
      size: 320,
      minSize: 120,
      meta: { headerClassName: `${TH} pl-4`, cellClassName: `${TD} pl-4` },
      cell: ({ row }) => (
        <>
          <div className="truncate font-medium">{row.original.name}</div>
          {row.original.description && (
            <div className="truncate text-xs text-muted-foreground">{row.original.description}</div>
          )}
        </>
      ),
    },
    {
      id: 'tags',
      header: t('taggingRules.table.tags'),
      size: 240,
      minSize: 80,
      meta: { headerClassName: TH, cellClassName: TD },
      cell: ({ row }) => {
        const tags = row.original.tags ?? []
        return (
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
        )
      },
    },
    {
      id: 'conditions',
      header: t('taggingRules.table.conditions'),
      size: 100,
      minSize: 60,
      meta: { headerClassName: TH, cellClassName: `${TD} font-mono text-xs text-muted-foreground` },
      cell: ({ row }) => row.original.conditions?.length ?? 0,
    },
  ]
}
