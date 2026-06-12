import { useEffect, useRef, useState } from 'react'
import { Check, ChevronDown, Lock, Pencil, Plus, Tag as TagIcon, Trash2, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { TAG_COLORS } from '../lib/alert-meta'
import type { AlertTag } from '../types/alert.types'

/**
 * Tag picker for the drawer: toggle existing catalog tags, or create a brand-new
 * tag (name + color) that gets registered and applied. Unlike the toolbar's tag
 * filter, inner clicks don't dismiss the popover — only an outside click does.
 */
export function AlertTagEditor({
  tags,
  catalog,
  onTags,
  onCreateTag,
  onUpdateTag,
  onDeleteTag,
}: {
  tags: string[]
  catalog: AlertTag[]
  onTags: (tags: string[]) => void
  onCreateTag: (tagName: string, tagColor: string) => void
  onUpdateTag: (id: number, tagName: string, tagColor: string) => void
  onDeleteTag: (id: number, tagName: string) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [color, setColor] = useState(TAG_COLORS[5])
  // Inline edit of an existing catalog tag.
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editName, setEditName] = useState('')
  const [editColor, setEditColor] = useState(TAG_COLORS[5])
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const trimmed = name.trim()
  const exists = catalog.some((tagItem) => tagItem.tagName.toLowerCase() === trimmed.toLowerCase())
  const canCreate = trimmed.length > 0 && !exists
  const create = () => {
    if (!canCreate) return
    onCreateTag(trimmed, color)
    setName('')
  }
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
        className="inline-flex items-center gap-1 rounded-md border border-border bg-card px-2 py-1 text-sm hover:bg-muted"
      >
        <TagIcon size={12} /> {t('alerts.tagEditor.tag')} <ChevronDown size={12} />
      </button>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-64 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="max-h-48 overflow-y-auto">
            {catalog.length === 0 && (
              <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('alerts.tagEditor.noTagsYet')}</div>
            )}
            {catalog.map((tg) => {
              const has = tags.includes(tg.tagName)
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
                  <button
                    onClick={() => onTags(has ? tags.filter((x) => x !== tg.tagName) : [...tags, tg.tagName])}
                    className="flex min-w-0 flex-1 items-center gap-2 text-left"
                  >
                    <span
                      className="h-2.5 w-2.5 shrink-0 rounded-full"
                      style={{ backgroundColor: tg.tagColor || '#64748b' }}
                    />
                    <span className="truncate">{tg.tagName}</span>
                    {has && <Check size={14} className="ml-auto shrink-0 text-primary" />}
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
                title={t('alerts.tagEditor.createApply')}
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
