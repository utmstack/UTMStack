import { memo, useCallback, useMemo, useState } from 'react'
import { Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Input } from '@/shared/components/ui/input'
import type { FilterType, IndexField } from '../types/log-explorer.types'
import { FieldItem } from './FieldItem'
import { SidebarSectionLabel } from './SidebarSectionLabel'

// Memoized: rendered even in SQL mode, but its props are reference-stable so
// SQL-typing keystrokes don't re-run FieldItem() N times.
function FieldSidebarImpl({
  fields,
  dataset,

  pattern,
  filters,
  columns,
  onAdd,
  onToggleColumn,
}: {
  fields: IndexField[]
  dataset: string
  pattern: string | null
  filters: FilterType[]
  columns: string[]
  onAdd: (f: FilterType) => void
  onToggleColumn: (name: string) => void
}) {
  const { t } = useTranslation()
  const [q, setQ] = useState('')
  const [openField, setOpenField] = useState<string | null>(null)
  const handleToggleOpen = useCallback(
    (name: string) => setOpenField((prev) => (prev === name ? null : name)),
    [],
  )

  // Hide raw .keyword variants — the base field covers them.
  const visible = useMemo(
    () =>
      fields
        .filter((f) => !f.name.endsWith('.keyword'))
        .filter((f) => (q ? f.name.toLowerCase().includes(q.toLowerCase()) : true))
        .sort((a, b) => a.name.localeCompare(b.name)),
    [fields, q]
  )

  // Selected fields (in column order) on top, the rest below — like the legacy.
  const selected = columns
    .map((name) => visible.find((f) => f.name === name))
    .filter((f): f is IndexField => !!f)
  const available = visible.filter((f) => !columns.includes(f.name))

  const renderItem = (f: IndexField) => (
    <FieldItem
      key={f.name}
      field={f}
      dataset={dataset}
      pattern={pattern}
      filters={filters}
      isColumn={columns.includes(f.name)}
      open={openField === f.name}
      onToggle={handleToggleOpen}
      onAdd={onAdd}
      onToggleColumn={onToggleColumn}
    />
  )

  return (
    <aside className="hidden w-64 shrink-0 flex-col bg-muted/10 lg:flex">
      <div className="flex items-center justify-between border-b border-border/70 px-3 pt-2.5 pb-1.5">
        <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
          {t('logExplorer.fields.title')}
        </span>
        <span className="font-mono text-[11px] text-muted-foreground/60">{visible.length}</span>
      </div>
      <div className="border-b border-border/70 p-2">
        <div className="relative">
          <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            placeholder={t('logExplorer.fields.filterPlaceholder')}
            className="h-8 border-0 bg-transparent pl-7 text-xs shadow-none focus-visible:bg-card focus-visible:ring-1 focus-visible:ring-border"
          />
        </div>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto p-1.5">
        {visible.length === 0 && (
          <div className="px-2 py-6 text-center text-xs text-muted-foreground">{t('logExplorer.fields.none')}</div>
        )}

        {selected.length > 0 && (
          <>
            <SidebarSectionLabel>{t('logExplorer.fields.selected')}</SidebarSectionLabel>
            {selected.map(renderItem)}
            <SidebarSectionLabel className="mt-3">{t('logExplorer.fields.available')}</SidebarSectionLabel>
          </>
        )}
        {available.map(renderItem)}
      </div>
    </aside>
  )
}

export const FieldSidebar = memo(FieldSidebarImpl)
