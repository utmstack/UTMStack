import { useEffect, useRef, useState } from 'react'
import { Check, ChevronDown, Download, Lock, Pencil, Plus, RefreshCw, Search, Tag as TagIcon, Trash2, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { TimeRangePicker, type TimeRange } from '@/shared/components/ui/time-range-picker'
import { SELECT_CLS, TAG_COLORS } from '../lib/alert-meta'
import type { AlertTag, SeverityKey } from '../types/alert.types'

export function AlertsToolbar({
  search,
  onSearch,
  severity,
  onSeverity,
  range,
  onRange,
  tagCatalog,
  tagFilter,
  onTagFilter,
  onCreateTag,
  onUpdateTag,
  onDeleteTag,
  onRefresh,
  onExport,
  loading,
  showSeverity = true,
}: {
  search: string
  onSearch: (s: string) => void
  severity: SeverityKey | 'all'
  onSeverity: (s: SeverityKey | 'all') => void
  range: TimeRange
  onRange: (r: TimeRange) => void
  tagCatalog: AlertTag[]
  tagFilter: string[]
  onTagFilter: (tags: string[]) => void
  onCreateTag: (tagName: string, tagColor: string) => void
  onUpdateTag: (id: number, tagName: string, tagColor: string) => void
  onDeleteTag: (id: number, tagName: string) => void
  onRefresh: () => void
  onExport: () => void
  loading: boolean
  showSeverity?: boolean
}) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-wrap items-center gap-2">
      <div className="relative min-w-[240px] flex-1">
        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder={t('alerts.toolbar.searchPlaceholder')}
          value={search}
          onChange={(e) => onSearch(e.target.value)}
          className="h-9 pl-9"
        />
      </div>
      {showSeverity && (
        <select
          value={severity}
          onChange={(e) => onSeverity(e.target.value as SeverityKey | 'all')}
          className={SELECT_CLS}
        >
          <option value="all">{t('alerts.toolbar.allSeverities')}</option>
          <option value="high">{t('alerts.severity.high')}</option>
          <option value="medium">{t('alerts.severity.medium')}</option>
          <option value="low">{t('alerts.severity.low')}</option>
        </select>
      )}
      <TagFilter
        catalog={tagCatalog}
        selected={tagFilter}
        onSelected={onTagFilter}
        onCreateTag={onCreateTag}
        onUpdateTag={onUpdateTag}
        onDeleteTag={onDeleteTag}
      />
      <TimeRangePicker value={range} onChange={onRange} allowAllTime align="right" />
      <Button variant="outline" size="sm" onClick={onExport} title={t('alerts.toolbar.export')}>
        <Download size={14} />
      </Button>
      <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading} title={t('alerts.toolbar.refresh')}>
        <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
      </Button>
    </div>
  )
}

function TagFilter({
  catalog,
  selected,
  onSelected,
  onCreateTag,
  onUpdateTag,
  onDeleteTag,
}: {
  catalog: AlertTag[]
  selected: string[]
  onSelected: (tags: string[]) => void
  onCreateTag: (tagName: string, tagColor: string) => void
  onUpdateTag: (id: number, tagName: string, tagColor: string) => void
  onDeleteTag: (id: number, tagName: string) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editName, setEditName] = useState('')
  const [editColor, setEditColor] = useState(TAG_COLORS[5])
  const [name, setName] = useState('')
  const [color, setColor] = useState(TAG_COLORS[5])
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const trimmed = name.trim()
  const exists = catalog.some((tg) => tg.tagName.toLowerCase() === trimmed.toLowerCase())
  const canCreate = trimmed.length > 0 && !exists
  const create = () => {
    if (!canCreate) return
    onCreateTag(trimmed, color)
    setName('')
  }

  const toggle = (name: string) =>
    onSelected(selected.includes(name) ? selected.filter((t) => t !== name) : [...selected, name])
  const startEdit = (tg: AlertTag) => {
    setEditingId(tg.id)
    setEditName(tg.tagName)
    setEditColor(tg.tagColor || TAG_COLORS[5])
  }
  const saveEdit = () => {
    const n = editName.trim()
    if (!n || editingId == null) return
    onUpdateTag(editingId, n, editColor)
    setEditingId(null)
  }

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen((v) => !v)}
        className={cn(
          'inline-flex h-9 items-center gap-1.5 rounded-md border px-3 text-sm transition-colors',
          selected.length ? 'border-primary bg-primary/5 text-foreground' : 'border-input bg-background hover:bg-muted'
        )}
      >
        <TagIcon size={14} className="text-muted-foreground" />
        {selected.length ? t('alerts.tagFilter.tag', { count: selected.length }) : t('alerts.tagFilter.tags')}
        <ChevronDown size={12} className="opacity-60" />
      </button>
      {open && (
        <div className="absolute right-0 top-full z-30 mt-1 w-64 rounded-md border border-border bg-popover py-1 shadow-lg">
          {selected.length > 0 && (
            <button
              onClick={() => onSelected([])}
              className="mb-1 flex w-full items-center gap-2 border-b border-border px-3 py-1.5 text-left text-xs text-muted-foreground hover:bg-muted"
            >
              <X size={12} /> {t('alerts.tagFilter.clearSelection')}
            </button>
          )}
          <div className="max-h-56 overflow-y-auto">
            {catalog.length === 0 && (
              <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('alerts.tagFilter.noTags')}</div>
            )}
            {catalog.map((tg) => {
              const on = selected.includes(tg.tagName)
              if (editingId === tg.id) {
                return (
                  <div key={tg.id} className="px-3 py-1.5">
                    <div className="flex items-center gap-1.5">
                      <input
                        value={editName}
                        onChange={(e) => setEditName(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') saveEdit()
                          if (e.key === 'Escape') setEditingId(null)
                        }}
                        autoFocus
                        className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                      />
                      <button
                        onClick={saveEdit}
                        disabled={!editName.trim()}
                        title={t('alerts.tagEditor.save')}
                        className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground disabled:opacity-40"
                      >
                        <Check size={13} />
                      </button>
                      <button
                        onClick={() => setEditingId(null)}
                        title={t('alerts.tagEditor.cancel')}
                        className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md border border-input hover:bg-muted"
                      >
                        <X size={13} />
                      </button>
                    </div>
                    <div className="mt-1.5 flex flex-wrap gap-1">
                      {TAG_COLORS.map((c) => (
                        <button
                          key={c}
                          onClick={() => setEditColor(c)}
                          className={cn(
                            'h-4 w-4 rounded-full ring-offset-1 ring-offset-popover transition',
                            editColor === c && 'ring-2 ring-ring'
                          )}
                          style={{ backgroundColor: c }}
                          aria-label={t('alerts.tagEditor.color', { color: c })}
                        />
                      ))}
                    </div>
                  </div>
                )
              }
              return (
                <div key={tg.id} className="group flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-muted">
                  <button onClick={() => toggle(tg.tagName)} className="flex min-w-0 flex-1 items-center gap-2 text-left">
                    <span
                      className="h-2.5 w-2.5 shrink-0 rounded-full"
                      style={{ backgroundColor: tg.tagColor || '#64748b' }}
                    />
                    <span className="truncate">{tg.tagName}</span>
                    {on && <Check size={14} className="ml-auto shrink-0 text-primary" />}
                  </button>
                  {tg.systemOwner ? (
                    <Lock size={11} className="shrink-0 text-muted-foreground/50" aria-label={t('alerts.tagEditor.systemLocked')} />
                  ) : (
                    <span className="flex shrink-0 items-center gap-0.5 opacity-0 transition group-hover:opacity-100">
                      <button
                        onClick={() => startEdit(tg)}
                        title={t('alerts.tagEditor.edit')}
                        className="rounded p-1 text-muted-foreground hover:bg-background hover:text-foreground"
                      >
                        <Pencil size={12} />
                      </button>
                      <button
                        onClick={() => onDeleteTag(tg.id, tg.tagName)}
                        title={t('alerts.tagEditor.delete')}
                        className="rounded p-1 text-muted-foreground hover:bg-background hover:text-red-500"
                      >
                        <Trash2 size={12} />
                      </button>
                    </span>
                  )}
                </div>
              )
            })}
          </div>
          <div className="mt-1 border-t border-border px-3 pb-2 pt-2">
            <div className="mb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              {t('alerts.tagEditor.createTag')}
            </div>
            <div className="flex items-center gap-1.5">
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && create()}
                placeholder={t('alerts.tagEditor.newTagName')}
                className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
              <button
                onClick={create}
                disabled={!canCreate}
                className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground disabled:opacity-40"
                title={t('alerts.tagEditor.createTag')}
              >
                <Plus size={14} />
              </button>
            </div>
            <div className="mt-2 flex flex-wrap gap-1.5">
              {TAG_COLORS.map((c) => (
                <button
                  key={c}
                  onClick={() => setColor(c)}
                  className={cn(
                    'h-5 w-5 rounded-full ring-offset-1 ring-offset-popover transition',
                    color === c && 'ring-2 ring-ring'
                  )}
                  style={{ backgroundColor: c }}
                  aria-label={t('alerts.tagEditor.color', { color: c })}
                />
              ))}
            </div>
            {exists && trimmed.length > 0 && (
              <div className="mt-1.5 text-[10px] text-amber-500">{t('alerts.tagEditor.alreadyExists')}</div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
