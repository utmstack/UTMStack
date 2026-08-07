import { useCallback, useEffect, useState } from 'react'
import { Bookmark, Loader2, Save, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import type { TimeRange } from '@/shared/components/ui/time-range-picker'
import type { FilterType } from '../types/log-explorer.types'
import { savedQueriesHttpService, type SavedQuery } from '../services/saved-queries-http.service'

// ── Saved searches ────────────────────────────────────────────────────────
// Kept by the backend, tenant-scoped and owned by whoever saved them, so they
// follow the analyst to another browser instead of living in this one.

export interface SavedSearchState {
  dataset: string
  patternStr: string | null
  range: TimeRange
  filters: FilterType[]
  searchInput: string
  appliedQuery: string
}

const LEGACY_KEY = 'utmstack-logexplorer-saved-searches'

/** The stored row carries the dataset on its own column and the rest as JSON. */
function toState(q: SavedQuery): SavedSearchState | null {
  try {
    const rest = JSON.parse(q.filters || '{}') as Omit<SavedSearchState, 'patternStr'>
    return { ...rest, patternStr: q.dataset || null }
  } catch {
    return null
  }
}

function toInput(name: string, s: SavedSearchState) {
  const { patternStr, ...rest } = s
  return { name, dataset: patternStr ?? '', filters: JSON.stringify(rest) }
}

/**
 * Moves anything this browser saved before these lived on the server, once.
 * Without it an upgrade silently empties the list of someone who had built one.
 */
async function importLegacy(existing: SavedQuery[]): Promise<boolean> {
  let raw: string | null = null
  try {
    raw = window.localStorage.getItem(LEGACY_KEY)
  } catch {
    return false
  }
  if (!raw) return false

  let parsed: (SavedSearchState & { name: string })[] = []
  try {
    const arr = JSON.parse(raw)
    parsed = Array.isArray(arr) ? arr : []
  } catch {
    parsed = []
  }

  const taken = new Set(existing.map((q) => q.name))
  const pending = parsed.filter((s) => s.name && !taken.has(s.name))
  for (const s of pending) {
    try {
      await savedQueriesHttpService.create(toInput(s.name, s))
    } catch {
      // Leave the key in place so the next load can try again.
      return false
    }
  }
  try {
    window.localStorage.removeItem(LEGACY_KEY)
  } catch {
    /* ignore */
  }
  return pending.length > 0
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
  const [list, setList] = useState<SavedQuery[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [busy, setBusy] = useState(false)
  const [name, setName] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const items = await savedQueriesHttpService.list()
      if (await importLegacy(items)) {
        setList(await savedQueriesHttpService.list())
      } else {
        setList(items)
      }
    } catch {
      toast.error(t('logExplorer.saved.loadFailed'))
    } finally {
      setLoading(false)
    }
  }, [t])

  useEffect(() => {
    void load()
  }, [load])

  const save = async () => {
    const n = name.trim()
    if (!n || busy) return
    setBusy(true)
    try {
      // Saving the same name replaces it, which is what the local list did.
      const clash = list.find((s) => s.name === n)
      if (clash) await savedQueriesHttpService.remove(clash.id)
      await savedQueriesHttpService.create(toInput(n, snapshot()))
      setName('')
      setSaving(false)
      toast.success(t('logExplorer.saved.saved', { name: n }))
      await load()
    } catch {
      toast.error(t('logExplorer.saved.saveFailed'))
    } finally {
      setBusy(false)
    }
  }

  const remove = async (q: SavedQuery) => {
    if (busy) return
    setBusy(true)
    try {
      await savedQueriesHttpService.remove(q.id)
      setList((l) => l.filter((x) => x.id !== q.id))
    } catch {
      toast.error(t('logExplorer.saved.deleteFailed'))
    } finally {
      setBusy(false)
    }
  }

  const apply = (q: SavedQuery) => {
    const state = toState(q)
    if (!state) {
      toast.error(t('logExplorer.saved.unreadable'))
      return
    }
    onLoad(state)
    setOpen(false)
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
              {loading ? (
                <div className="flex items-center gap-2 px-3 py-3 text-xs text-muted-foreground">
                  <Loader2 className="h-3 w-3 animate-spin" />
                  {t('logExplorer.saved.loading')}
                </div>
              ) : list.length === 0 ? (
                <div className="px-3 py-3 text-xs text-muted-foreground">{t('logExplorer.saved.empty')}</div>
              ) : (
                list.map((s) => (
                  <div key={s.id} className="group flex items-center gap-2 px-3 py-1.5 text-sm hover:bg-muted/40">
                    <button
                      onClick={() => apply(s)}
                      className="min-w-0 flex-1 truncate text-left"
                      title={s.owner ? `${s.name} — ${s.owner}` : s.name}
                    >
                      {s.name}
                    </button>
                    <button
                      onClick={() => void remove(s)}
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
                      if (e.key === 'Enter') void save()
                      if (e.key === 'Escape') setSaving(false)
                    }}
                    placeholder={t('logExplorer.saved.namePlaceholder')}
                    className="h-7 min-w-0 flex-1 rounded border border-input bg-background px-2 text-xs outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  />
                  <button
                    onClick={() => void save()}
                    disabled={busy}
                    className="rounded bg-primary px-2 py-1 text-[11px] font-medium text-primary-foreground hover:opacity-90 disabled:opacity-50"
                  >
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
