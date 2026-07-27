import { useState } from 'react'
import { Bookmark, Save, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'
import type { FilterType } from '../types/log-explorer.types'

// ── Saved searches ────────────────────────────────────────────────────────
// Reusable query snapshots persisted in localStorage so an analyst's daily
// queries survive reloads. Backend-free by design (per-browser, like the tabs).

export interface SavedSearchState {
  patternStr: string | null
  range: TimeRange
  filters: FilterType[]
  searchInput: string
  appliedQuery: string
}

interface SavedSearch extends SavedSearchState {
  name: string
}

const SAVED_SEARCHES_KEY = 'utmstack-logexplorer-saved-searches'

function loadSavedSearches(): SavedSearch[] {
  if (typeof window === 'undefined') return []
  try {
    const raw = window.localStorage.getItem(SAVED_SEARCHES_KEY)
    const arr = raw ? JSON.parse(raw) : []
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}

function persistSavedSearches(list: SavedSearch[]) {
  try {
    window.localStorage.setItem(SAVED_SEARCHES_KEY, JSON.stringify(list))
  } catch {
    /* ignore quota/availability errors */
  }
}

export function SavedSearches({
  snapshot,
  onLoad,
}: {
  snapshot: () => SavedSearchState
  onLoad: (s: SavedSearchState) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [list, setList] = useState<SavedSearch[]>(() => loadSavedSearches())
  const [saving, setSaving] = useState(false)
  const [name, setName] = useState('')

  const commit = (next: SavedSearch[]) => {
    setList(next)
    persistSavedSearches(next)
  }

  const save = () => {
    const n = name.trim()
    if (!n) return
    const next = [...list.filter((s) => s.name !== n), { name: n, ...snapshot() }]
    commit(next)
    setName('')
    setSaving(false)
    toast.success(t('logExplorer.saved.saved', { name: n }))
  }

  return (
    <div className="relative">
      <button
        onClick={() => setOpen((o) => !o)}
        className="inline-flex h-8 items-center gap-1.5 rounded-md border border-border bg-card px-2.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
      >
        <Bookmark size={13} /> {t('logExplorer.saved.title')}
        {list.length > 0 && <span className="font-mono text-[10px]">({list.length})</span>}
      </button>
      {open && (
        <>
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} />
          <div className="absolute left-0 z-20 mt-1 w-72 overflow-hidden rounded-md border border-border bg-card shadow-lg">
            <div className="max-h-64 overflow-y-auto">
              {list.length === 0 ? (
                <div className="px-3 py-3 text-xs text-muted-foreground">{t('logExplorer.saved.empty')}</div>
              ) : (
                list.map((s) => (
                  <div key={s.name} className="group flex items-center gap-2 px-3 py-1.5 text-sm hover:bg-muted/40">
                    <button
                      onClick={() => {
                        onLoad(s)
                        setOpen(false)
                      }}
                      className="min-w-0 flex-1 truncate text-left"
                      title={s.name}
                    >
                      {s.name}
                    </button>
                    <button
                      onClick={() => commit(list.filter((x) => x.name !== s.name))}
                      title={t('logExplorer.saved.delete')}
                      className="shrink-0 text-muted-foreground opacity-0 transition-opacity hover:text-red-500 group-hover:opacity-100"
                    >
                      <Trash2 size={12} />
                    </button>
                  </div>
                ))
              )}
            </div>
            <div className="border-t border-border p-2">
              {saving ? (
                <div className="flex items-center gap-1.5">
                  <input
                    autoFocus
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') save()
                      if (e.key === 'Escape') setSaving(false)
                    }}
                    placeholder={t('logExplorer.saved.namePlaceholder')}
                    className="h-7 min-w-0 flex-1 rounded border border-input bg-background px-2 text-xs outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  />
                  <button onClick={save} className="rounded bg-primary px-2 py-1 text-[11px] font-medium text-primary-foreground hover:opacity-90">
                    {t('logExplorer.saved.save')}
                  </button>
                </div>
              ) : (
                <button
                  onClick={() => setSaving(true)}
                  className="flex w-full items-center gap-1.5 rounded px-2 py-1.5 text-xs text-muted-foreground hover:bg-muted/40 hover:text-foreground"
                >
                  <Save size={12} /> {t('logExplorer.saved.saveCurrent')}
                </button>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
