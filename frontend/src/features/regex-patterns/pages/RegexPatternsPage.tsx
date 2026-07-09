import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Loader2, Lock, Plus, Regex, RefreshCw, Search, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import {
  RegexPatternsHttpError,
  regexPatternsHttpService,
  type RegexPattern,
} from '../services/regex-patterns-http.service'

type Tab = 'all' | 'system' | 'user'
const TABS: Tab[] = ['all', 'system', 'user']

const COLS = 'minmax(150px,1fr) minmax(200px,1.6fr) 100px 60px'

export function RegexPatternsPage() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<Tab>('all')
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [items, setItems] = useState<RegexPattern[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0) // 0-based
  const [pageSize] = useState(50)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [editing, setEditing] = useState<{ pattern: RegexPattern; creating: boolean } | null>(null)

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  const query = useMemo(() => {
    const q: { search?: string; system?: boolean; page: number; size: number } = {
      search: debounced || undefined,
      page,
      size: pageSize,
    }
    if (tab === 'system') q.system = true
    else if (tab === 'user') q.system = false
    return q
  }, [debounced, tab, page, pageSize])

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    regexPatternsHttpService
      .list(query)
      .then((r) => {
        setItems((prev) => (page === 0 ? (r.data ?? []) : [...prev, ...(r.data ?? [])]))
        setTotal(r.total ?? 0)
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [query])
  useEffect(() => {
    load()
  }, [load])

  return (
    <div className="flex h-full min-h-0 w-full flex-col px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <Regex size={14} strokeWidth={1.75} />
          <span><span className="font-medium text-foreground">{total}</span> {t('regexPatterns.title').toLowerCase()}</span>
        </div>
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            onClick={() =>
              setEditing({
                pattern: { patternId: '', patternDescription: '', patternDefinition: '', systemOwner: false },
                creating: true,
              })
            }
          >
            <Plus size={14} className="mr-1.5" /> {t('regexPatterns.new')}
          </Button>
        </div>
      </header>

      <div className="mt-4 flex shrink-0 flex-wrap items-center gap-2">
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('regexPatterns.search')}
            className="w-[260px] pl-8"
          />
        </div>
        <div className="inline-flex rounded-md border border-border p-0.5">
          {TABS.map((tb) => (
            <button
              key={tb}
              onClick={() => {
                setTab(tb)
                setPage(0)
              }}
              className={cn(
                'rounded px-2.5 py-1 text-xs transition-colors',
                tab === tb ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
              )}
            >
              {t(`regexPatterns.tabs.${tb}`)}
            </button>
          ))}
        </div>
        <Button variant="outline" size="sm" onClick={load} disabled={loading} title={t('regexPatterns.refresh')}>
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </Button>
      </div>

      <div className="mt-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: COLS }}
        >
          <div>{t('regexPatterns.cols.pattern')}</div>
          <div>{t('regexPatterns.cols.definition')}</div>
          <div>{t('regexPatterns.cols.type')}</div>
          <div />
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading && items.length === 0 ? (
            <Center>
              <Loader2 className="h-4 w-4 animate-spin" /> {t('regexPatterns.loading')}
            </Center>
          ) : error ? (
            <Center>
              <AlertTriangle size={16} className="text-amber-500" /> {t('regexPatterns.loadError')}
              <Button variant="outline" size="sm" className="ml-2" onClick={load}>
                {t('regexPatterns.retry')}
              </Button>
            </Center>
          ) : items.length === 0 ? (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('regexPatterns.empty')}</div>
          ) : (
            <>
              {items.map((p) => (
                <Row key={p.patternId} p={p} onOpen={() => setEditing({ pattern: p, creating: false })} />
              ))}
              <InfiniteScrollSentinel
                onReach={() => setPage((p) => p + 1)}
                hasMore={items.length < total}
                loading={loading}
                endLabel={t('common.allLoaded', { count: total })}
              />
            </>
          )}
        </div>
      </div>

      {editing && (
        <PatternEditor
          pattern={editing.pattern}
          creating={editing.creating}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null)
            load()
          }}
        />
      )}
    </div>
  )
}

function Row({ p, onOpen }: { p: RegexPattern; onOpen: () => void }) {
  const { t } = useTranslation()
  return (
    <div
      className="grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-sm transition-colors last:border-0 hover:bg-muted/40"
      style={{ gridTemplateColumns: COLS }}
      onClick={onOpen}
    >
      <div className="flex min-w-0 items-center gap-2" title={p.patternId}>
        <Regex size={14} className="shrink-0 text-muted-foreground" />
        <span className="truncate font-mono text-[13px]">{p.patternId}</span>
      </div>
      <div className="min-w-0 truncate font-mono text-[11px] text-foreground/70" title={p.patternDefinition}>
        {p.patternDefinition || '—'}
      </div>
      <div>
        <span
          className={cn(
            'inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-medium',
            p.systemOwner ? 'bg-violet-500/15 text-violet-500' : 'bg-sky-500/15 text-sky-500',
          )}
        >
          {p.systemOwner && <Lock size={9} />}
          {t(p.systemOwner ? 'regexPatterns.system' : 'regexPatterns.user')}
        </span>
      </div>
      <div className="text-right text-[11px] text-muted-foreground">{t('regexPatterns.view')}</div>
    </div>
  )
}

function PatternEditor({
  pattern,
  creating,
  onClose,
  onSaved,
}: {
  pattern: RegexPattern
  creating: boolean
  onClose: () => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const [patternId, setPatternId] = useState(pattern.patternId)
  const [description, setDescription] = useState(pattern.patternDescription)
  const [definition, setDefinition] = useState(pattern.patternDefinition)
  const [saving, setSaving] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const readOnly = pattern.systemOwner

  const dirty = creating
    ? patternId.trim() !== '' || definition.trim() !== ''
    : description !== pattern.patternDescription || definition !== pattern.patternDefinition

  const save = async () => {
    if (creating && !patternId.trim()) {
      toast.error(t('regexPatterns.toast.patternIdRequired'))
      return
    }
    if (!definition.trim()) {
      toast.error(t('regexPatterns.toast.definitionRequired'))
      return
    }
    setSaving(true)
    try {
      const body = {
        patternId: creating ? patternId.trim() : pattern.patternId,
        patternDescription: description,
        patternDefinition: definition,
      }
      if (creating) await regexPatternsHttpService.create(body)
      else await regexPatternsHttpService.update(body)
      toast.success(t('regexPatterns.toast.saved'))
      onSaved()
    } catch (err) {
      toast.error(err instanceof RegexPatternsHttpError ? err.message : t('regexPatterns.toast.saveError'))
    } finally {
      setSaving(false)
    }
  }

  const remove = async () => {
    setSaving(true)
    try {
      await regexPatternsHttpService.remove(pattern.patternId)
      toast.success(t('regexPatterns.toast.deleted'))
      onSaved()
    } catch {
      toast.error(t('regexPatterns.toast.deleteError'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[760px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div className="min-w-0">
            <h2 className="flex items-center gap-2 truncate text-lg font-semibold">
              {creating ? t('regexPatterns.editor.titleNew') : pattern.patternId}
              {readOnly && (
                <span className="inline-flex items-center gap-1 rounded bg-violet-500/15 px-1.5 py-0.5 text-[10px] font-medium text-violet-500">
                  <Lock size={9} /> {t('regexPatterns.system')}
                </span>
              )}
            </h2>
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-6">
          {readOnly && (
            <div className="flex shrink-0 items-center gap-2 rounded-md bg-violet-500/10 px-3 py-2 text-xs text-violet-600 dark:text-violet-300">
              <Lock size={13} /> {t('regexPatterns.editor.readOnly')}
            </div>
          )}
          {creating && (
            <div className="shrink-0 space-y-1.5">
              <label className="block text-xs font-medium text-foreground/80">{t('regexPatterns.editor.patternId')}</label>
              <Input
                value={patternId}
                onChange={(e) => setPatternId(e.target.value)}
                placeholder="my_custom_pattern"
                className="font-mono"
              />
              <p className="text-[11px] text-muted-foreground">{t('regexPatterns.editor.patternIdHint')}</p>
            </div>
          )}
          <div className="shrink-0 space-y-1.5">
            <label className="block text-xs font-medium text-foreground/80">{t('regexPatterns.editor.description')}</label>
            <Input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              readOnly={readOnly}
              placeholder={t('regexPatterns.editor.descriptionPlaceholder')}
              className={cn(readOnly && 'opacity-80')}
            />
          </div>
          <div className="flex min-h-0 flex-1 flex-col gap-1.5">
            <label className="block shrink-0 text-xs font-medium text-foreground/80">
              {t('regexPatterns.editor.definition')}
            </label>
            <textarea
              value={definition}
              onChange={(e) => setDefinition(e.target.value)}
              readOnly={readOnly}
              spellCheck={false}
              placeholder="(?P<name>\\d+)"
              className="min-h-[160px] flex-1 resize-none rounded-md border border-input bg-background/40 p-3 font-mono text-[12px] leading-relaxed text-foreground caret-foreground outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring read-only:opacity-80"
            />
            <p className="shrink-0 text-[11px] text-muted-foreground">{t('regexPatterns.editor.definitionHint')}</p>
          </div>
        </div>

        <footer className="flex items-center justify-between gap-2 border-t border-border px-6 py-3">
          <div>
            {!creating && !readOnly &&
              (confirmDelete ? (
                <div className="flex items-center gap-2">
                  <Button size="sm" variant="destructive" onClick={() => void remove()} disabled={saving}>
                    <Trash2 size={13} className="mr-1.5" /> {t('regexPatterns.editor.confirmDelete')}
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => setConfirmDelete(false)} disabled={saving}>
                    {t('regexPatterns.editor.cancel')}
                  </Button>
                </div>
              ) : (
                <button
                  onClick={() => setConfirmDelete(true)}
                  className="rounded-md border border-red-500/30 bg-red-500/5 px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-500/10 dark:text-red-300"
                >
                  {t('regexPatterns.editor.delete')}
                </button>
              ))}
          </div>
          {!readOnly && (
            <Button size="sm" disabled={!dirty || saving} onClick={() => void save()}>
              {saving ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : null}
              {saving ? t('regexPatterns.editor.saving') : t('regexPatterns.editor.save')}
            </Button>
          )}
        </footer>
      </div>
    </div>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
