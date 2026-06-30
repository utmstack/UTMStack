import { useEffect, useState } from 'react'
import { ChevronDown, Plus, Tag as TagIcon, X } from 'lucide-react'
import type { TFunction } from 'i18next'
import { cn } from '@/shared/lib/utils'
import { TAG_COLORS } from '../lib/alert-meta'
import type { AlertTag } from '../types/tagging-rule.types'

export function TaggingRuleTagPicker({
  selected,
  catalog,
  onChange,
  onCreateTag,
  t,
}: {
  selected: AlertTag[]
  catalog: AlertTag[]
  onChange: (tags: AlertTag[]) => void
  onCreateTag: (tagName: string, tagColor: string) => Promise<AlertTag | null>
  t: TFunction
}) {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [color, setColor] = useState(TAG_COLORS[5])
  const selectedIds = new Set(selected.map((s) => s.id))

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      const el = e.target as HTMLElement
      if (!el.closest('[data-tagging-rule-tag-picker]')) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const toggle = (tag: AlertTag) => {
    if (selectedIds.has(tag.id)) onChange(selected.filter((s) => s.id !== tag.id))
    else onChange([...selected, tag])
  }

  const create = async () => {
    const n = name.trim()
    if (!n) return
    const tag = await onCreateTag(n, color)
    if (tag) {
      setName('')
      onChange([...selected, tag])
    }
  }

  return (
    <div data-tagging-rule-tag-picker className="relative">
      <div className="flex flex-wrap items-center gap-1.5">
        {selected.length === 0 && <span className="text-xs text-muted-foreground">{t('taggingRules.form.noTags')}</span>}
        {selected.map((tg) => (
          <span
            key={tg.id}
            className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px]"
            style={{
              backgroundColor: (tg.tagColor || '#64748b') + '22',
              color: tg.tagColor || '#64748b',
            }}
          >
            <TagIcon size={10} /> {tg.tagName}
            <button
              type="button"
              onClick={() => toggle(tg)}
              className="ml-0.5 rounded-full hover:bg-background/40"
              title={t('taggingRules.form.removeTag')}
            >
              <X size={10} />
            </button>
          </span>
        ))}
        <button
          type="button"
          onClick={() => setOpen((v) => !v)}
          className="inline-flex items-center gap-1 rounded-md border border-input bg-background px-2 py-1 text-xs hover:bg-muted"
        >
          <Plus size={12} /> {t('taggingRules.form.addTag')} <ChevronDown size={10} />
        </button>
      </div>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-72 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="max-h-48 overflow-y-auto">
            {catalog.length === 0 && (
              <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('taggingRules.form.noCatalog')}</div>
            )}
            {catalog.map((tg) => {
              const has = selectedIds.has(tg.id)
              return (
                <button
                  key={tg.id}
                  type="button"
                  onClick={() => toggle(tg)}
                  className={cn(
                    'flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-muted',
                    has && 'bg-muted/50'
                  )}
                >
                  <span
                    className="h-2.5 w-2.5 shrink-0 rounded-full"
                    style={{ backgroundColor: tg.tagColor || '#64748b' }}
                  />
                  <span className="truncate text-left">{tg.tagName}</span>
                  {has && <span className="ml-auto text-[10px] text-primary">✓</span>}
                </button>
              )
            })}
          </div>
          <div className="border-t border-border px-3 pb-2 pt-2">
            <div className="mb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              {t('taggingRules.form.createTag')}
            </div>
            <div className="flex items-center gap-1.5">
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && void create()}
                placeholder={t('taggingRules.form.newTagName')}
                className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
              <button
                type="button"
                onClick={() => void create()}
                disabled={!name.trim()}
                className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground disabled:opacity-40"
              >
                <Plus size={14} />
              </button>
            </div>
            <div className="mt-2 flex flex-wrap gap-1.5">
              {TAG_COLORS.map((c) => (
                <button
                  type="button"
                  key={c}
                  onClick={() => setColor(c)}
                  className={cn(
                    'h-5 w-5 rounded-full ring-offset-1 ring-offset-popover transition',
                    color === c && 'ring-2 ring-ring'
                  )}
                  style={{ backgroundColor: c }}
                />
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
