import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, Code2, FileCode, LayoutList, Loader2, Lock, Plus, RefreshCw, Search, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import {
  DataProcessingHttpError,
  filtersHttpService,
} from '@/features/data-processing/services/data-processing-http.service'
import type { DataTypeOption, Filter } from '@/features/data-processing/types/data-processing.types'
import { parseFilter, serializeFilter, type FilterModel } from '../lib/filter-model'
import { VisualFilterEditor } from '../components/VisualFilterEditor'
import { PatternInsertButton } from '../components/PatternInsertButton'
import { YamlCodeEditor } from '@/shared/components/YamlCodeEditor'
import {
  regexPatternsHttpService,
  type RegexPattern,
} from '@/features/regex-patterns/services/regex-patterns-http.service'

type Tab = 'all' | 'active' | 'inactive' | 'system' | 'user'
const TABS: Tab[] = ['all', 'active', 'inactive', 'system', 'user']

/** Clean display name: drop the folder prefix and the .yaml/.yml extension. */
function displayName(relPath: string): string {
  return (relPath.split('/').pop() ?? relPath).replace(/\.ya?ml$/i, '')
}

const COLS = 'minmax(180px,1fr) minmax(160px,1.3fr) 100px 80px 60px'

export function ParsingFiltersPage() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<Tab>('all')
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  // Seed from ?dataType on first render so the initial fetch is already filtered
  // (avoids a race where the unfiltered request resolves last and wins).
  const [dataType, setDataType] = useState(() => new URLSearchParams(window.location.search).get('dataType') ?? '')
  const [dataTypeOptions, setDataTypeOptions] = useState<string[]>([])
  const [items, setItems] = useState<Filter[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0) // 0-based
  const [pageSize] = useState(50)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [editing, setEditing] = useState<{ filter: Filter; creating: boolean } | null>(null)

  // Deep-link: ?dataType=<value> pre-filters the list to that data type
  // (e.g. opened from an integration's "Filters" button).
  const [searchParams] = useSearchParams()
  useEffect(() => {
    const dt = searchParams.get('dataType')
    if (dt) {
      setDataType(dt)
      setPage(0)
    }
  }, [searchParams])

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  // Distinct dataTypes for the filter dropdown (loaded once).
  useEffect(() => {
    filtersHttpService
      .dataTypes()
      .then((d) => setDataTypeOptions(d ?? []))
      .catch(() => setDataTypeOptions([]))
  }, [])

  const query = useMemo(() => {
    const q: {
      relPathContains?: string
      isActive?: boolean
      system?: boolean
      dataType?: string
      page: number
      size: number
    } = {
      relPathContains: debounced || undefined,
      dataType: dataType || undefined,
      page: page + 1, // backend is 1-based
      size: pageSize,
    }
    if (tab === 'active') q.isActive = true
    else if (tab === 'inactive') q.isActive = false
    else if (tab === 'system') q.system = true
    else if (tab === 'user') q.system = false
    return q
  }, [debounced, tab, dataType, page, pageSize])

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    filtersHttpService
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

  const toggleActive = async (f: Filter) => {
    const next = !f.active
    setItems((list) => list.map((x) => (x.relPath === f.relPath ? { ...x, active: next } : x)))
    try {
      await filtersHttpService.activate(f.relPath, next)
    } catch {
      setItems((list) => list.map((x) => (x.relPath === f.relPath ? { ...x, active: f.active } : x)))
      toast.error(t('parsingFilters.toast.activateError'))
    }
  }

  return (
    <div className="mx-auto flex h-full min-h-0 w-full max-w-[1100px] flex-col px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <FileCode size={14} strokeWidth={1.75} />
          <span><span className="font-medium text-foreground">{total}</span> {t('parsingFilters.title').toLowerCase()}</span>
        </div>
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            onClick={() => setEditing({ filter: { relPath: '', content: '', system: false, active: true }, creating: true })}
          >
            <Plus size={14} className="mr-1.5" /> {t('parsingFilters.new')}
          </Button>
        </div>
      </header>

      <div className="mt-4 flex shrink-0 flex-wrap items-center gap-2">
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('parsingFilters.search')}
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
              {t(`parsingFilters.tabs.${tb}`)}
            </button>
          ))}
        </div>
        <select
          value={dataType}
          onChange={(e) => {
            setDataType(e.target.value)
            setPage(0)
          }}
          className="h-9 rounded-md border border-input bg-background px-2.5 text-xs text-foreground focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        >
          <option value="">{t('parsingFilters.allDataTypes')}</option>
          {dataTypeOptions.map((dt) => (
            <option key={dt} value={dt}>
              {dt}
            </option>
          ))}
        </select>
        <Button variant="outline" size="sm" onClick={load} disabled={loading} title={t('parsingFilters.refresh')}>
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </Button>
      </div>

      <div className="mt-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: COLS }}
        >
          <div>{t('parsingFilters.cols.filter')}</div>
          <div>{t('parsingFilters.cols.dataTypes')}</div>
          <div>{t('parsingFilters.cols.type')}</div>
          <div className="text-center">{t('parsingFilters.cols.active')}</div>
          <div />
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading && items.length === 0 ? (
            <Center>
              <Loader2 className="h-4 w-4 animate-spin" /> {t('parsingFilters.loading')}
            </Center>
          ) : error ? (
            <Center>
              <AlertTriangle size={16} className="text-amber-500" /> {t('parsingFilters.loadError')}
              <Button variant="outline" size="sm" className="ml-2" onClick={load}>
                {t('parsingFilters.retry')}
              </Button>
            </Center>
          ) : items.length === 0 ? (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('parsingFilters.empty')}</div>
          ) : (
            <>
              {items.map((f) => (
                <Row key={f.relPath} f={f} onOpen={() => setEditing({ filter: f, creating: false })} onToggle={() => toggleActive(f)} />
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
        <FilterEditor
          filter={editing.filter}
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

function DataTypeCells({ dataTypes }: { dataTypes?: string[] }) {
  const dts = dataTypes ?? []
  if (dts.length === 0) return <div className="text-[11px] text-muted-foreground">—</div>
  const shown = dts.slice(0, 3)
  const extra = dts.length - shown.length
  return (
    <div className="flex min-w-0 flex-wrap items-center gap-1" title={dts.join(', ')}>
      {shown.map((dt) => (
        <span key={dt} className="truncate rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] text-foreground/80">
          {dt}
        </span>
      ))}
      {extra > 0 && <span className="text-[10px] text-muted-foreground">+{extra}</span>}
    </div>
  )
}

function Row({ f, onOpen, onToggle }: { f: Filter; onOpen: () => void; onToggle: () => void }) {
  const { t } = useTranslation()
  return (
    <div
      className="grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-sm transition-colors last:border-0 hover:bg-muted/40"
      style={{ gridTemplateColumns: COLS }}
      onClick={onOpen}
    >
      <div className="flex min-w-0 items-center gap-2" title={f.relPath}>
        <FileCode size={14} className="shrink-0 text-muted-foreground" />
        <span className="truncate text-[13px]">{displayName(f.relPath)}</span>
      </div>
      <DataTypeCells dataTypes={f.dataTypes} />
      <div>
        <span
          className={cn(
            'inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] font-medium',
            f.system ? 'bg-violet-500/15 text-violet-500' : 'bg-sky-500/15 text-sky-500',
          )}
        >
          {f.system && <Lock size={9} />}
          {t(f.system ? 'parsingFilters.system' : 'parsingFilters.user')}
        </span>
      </div>
      <div className="flex justify-center" onClick={(e) => e.stopPropagation()}>
        <Toggle checked={f.active} onChange={onToggle} />
      </div>
      <div className="text-right text-[11px] text-muted-foreground">{t('parsingFilters.view')}</div>
    </div>
  )
}

function FilterEditor({
  filter,
  creating,
  onClose,
  onSaved,
}: {
  filter: Filter
  creating: boolean
  onClose: () => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const [relPath, setRelPath] = useState(filter.relPath)
  const [content, setContent] = useState(filter.content)
  const [saving, setSaving] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const readOnly = filter.system

  // Visual ↔ Code. Content (YAML string) stays canonical; the visual editor
  // serializes back into it on every change.
  const [mode, setMode] = useState<'visual' | 'code'>('code')
  const [model, setModel] = useState<FilterModel>({ pipeline: [] })
  const [dataTypeOptions, setDataTypeOptions] = useState<DataTypeOption[]>([])
  const [patternOptions, setPatternOptions] = useState<RegexPattern[]>([])
  const yamlTaRef = useRef<HTMLTextAreaElement | null>(null)

  // A pattern created on the fly is added to the catalog so it's immediately referenceable.
  const addPattern = (p: RegexPattern) =>
    setPatternOptions((prev) => (prev.some((x) => x.patternId === p.patternId) ? prev : [p, ...prev]))

  // Catalog of known dataTypes to pick from in the visual editor.
  useEffect(() => {
    filtersHttpService
      .dataTypeCatalog()
      .then((d) => setDataTypeOptions(d ?? []))
      .catch(() => setDataTypeOptions([]))
  }, [])

  // Catalog of regex patterns to reference from grok steps (`{{.name}}`).
  useEffect(() => {
    regexPatternsHttpService
      .list({ size: 200 })
      .then((r) => setPatternOptions(r.data ?? []))
      .catch(() => setPatternOptions([]))
  }, [])

  // Default to the visual editor when the content parses cleanly.
  useEffect(() => {
    const r = parseFilter(filter.content)
    if (r.ok) {
      setModel(r.model)
      setMode('visual')
    }
  }, [filter.content])

  const canVisual = useMemo(() => parseFilter(content).ok, [content])

  const toVisual = () => {
    const r = parseFilter(content)
    if (!r.ok) {
      toast.error(t('parsingFilters.visual.cantParse', { error: r.error }))
      return
    }
    setModel(r.model)
    setMode('visual')
  }

  const onModelChange = (m: FilterModel) => {
    setModel(m)
    setContent(serializeFilter(m))
  }

  const dirty = creating ? relPath.trim() !== '' || content.trim() !== '' : content !== filter.content

  const save = async () => {
    if (creating && !relPath.trim()) {
      toast.error(t('parsingFilters.toast.relPathRequired'))
      return
    }
    if (!content.trim()) {
      toast.error(t('parsingFilters.toast.contentRequired'))
      return
    }
    setSaving(true)
    try {
      if (creating) await filtersHttpService.create({ relPath: relPath.trim(), content })
      else await filtersHttpService.update({ relPath: filter.relPath, content })
      toast.success(t('parsingFilters.toast.saved'))
      onSaved()
    } catch (err) {
      toast.error(err instanceof DataProcessingHttpError ? err.message : t('parsingFilters.toast.saveError'))
    } finally {
      setSaving(false)
    }
  }

  const remove = async () => {
    setSaving(true)
    try {
      await filtersHttpService.remove(filter.relPath)
      toast.success(t('parsingFilters.toast.deleted'))
      onSaved()
    } catch {
      toast.error(t('parsingFilters.toast.deleteError'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-[760px] flex-col overflow-hidden border-l border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div className="min-w-0">
            <h2 className="flex items-center gap-2 truncate text-lg font-semibold">
              {creating ? t('parsingFilters.editor.titleNew') : displayName(filter.relPath)}
              {readOnly && (
                <span className="inline-flex items-center gap-1 rounded bg-violet-500/15 px-1.5 py-0.5 text-[10px] font-medium text-violet-500">
                  <Lock size={9} /> {t('parsingFilters.system')}
                </span>
              )}
            </h2>
            {!creating && <p className="mt-0.5 truncate font-mono text-[11px] text-muted-foreground">{filter.relPath}</p>}
          </div>
          <button onClick={onClose} className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
            <X size={16} />
          </button>
        </header>

        <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden p-6">
          {readOnly && (
            <div className="flex shrink-0 items-center gap-2 rounded-md bg-violet-500/10 px-3 py-2 text-xs text-violet-600 dark:text-violet-300">
              <Lock size={13} /> {t('parsingFilters.editor.readOnly')}
            </div>
          )}
          {creating && (
            <div className="shrink-0 space-y-1.5">
              <label className="block text-xs font-medium text-foreground/80">{t('parsingFilters.editor.relPath')}</label>
              <Input value={relPath} onChange={(e) => setRelPath(e.target.value)} placeholder="myorg/custom-firewall.yaml" className="font-mono" />
              <p className="text-[11px] text-muted-foreground">{t('parsingFilters.editor.relPathHint')}</p>
            </div>
          )}
          <div className="flex min-h-0 flex-1 flex-col gap-1.5">
            <div className="flex shrink-0 items-center justify-between">
              <label className="block text-xs font-medium text-foreground/80">{t('parsingFilters.editor.content')}</label>
              <div className="inline-flex rounded-md border border-border p-0.5">
                <button
                  type="button"
                  onClick={toVisual}
                  disabled={!canVisual && mode === 'code'}
                  title={!canVisual ? t('parsingFilters.visual.codeOnly') : undefined}
                  className={cn(
                    'inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors disabled:opacity-40',
                    mode === 'visual' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  <LayoutList size={13} /> {t('parsingFilters.visual.visualTab')}
                </button>
                <button
                  type="button"
                  onClick={() => setMode('code')}
                  className={cn(
                    'inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors',
                    mode === 'code' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  <Code2 size={13} /> {t('parsingFilters.visual.codeTab')}
                </button>
              </div>
            </div>
            {mode === 'visual' ? (
              <div className="min-h-0 flex-1 overflow-y-auto rounded-md border border-input bg-background/20 p-3">
                <VisualFilterEditor
                  value={model}
                  readOnly={readOnly}
                  dataTypeOptions={dataTypeOptions}
                  patternOptions={patternOptions}
                  onPatternCreated={addPattern}
                  onChange={onModelChange}
                />
              </div>
            ) : (
              <div className="relative flex min-h-0 flex-1 flex-col">
                {!readOnly && (
                  <div className="absolute right-2 top-2 z-10">
                    <PatternInsertButton
                      taRef={yamlTaRef}
                      value={content}
                      onChange={setContent}
                      patternOptions={patternOptions}
                      onPatternCreated={addPattern}
                    />
                  </div>
                )}
                <YamlCodeEditor
                  value={content}
                  onChange={setContent}
                  readOnly={readOnly}
                  textareaRef={yamlTaRef}
                  placeholder={'pipeline:\n  - dataTypes: [ ... ]\n    steps: [ ... ]'}
                />
              </div>
            )}
            <p className="shrink-0 text-[11px] text-muted-foreground">{t('parsingFilters.editor.contentHint')}</p>
          </div>
        </div>

        <footer className="flex items-center justify-between gap-2 border-t border-border px-6 py-3">
          <div>
            {!creating && !readOnly &&
              (confirmDelete ? (
                <div className="flex items-center gap-2">
                  <Button size="sm" variant="destructive" onClick={() => void remove()} disabled={saving}>
                    <Trash2 size={13} className="mr-1.5" /> {t('parsingFilters.editor.confirmDelete')}
                  </Button>
                  <Button size="sm" variant="outline" onClick={() => setConfirmDelete(false)} disabled={saving}>
                    {t('parsingFilters.editor.cancel')}
                  </Button>
                </div>
              ) : (
                <button
                  onClick={() => setConfirmDelete(true)}
                  className="rounded-md border border-red-500/30 bg-red-500/5 px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-500/10 dark:text-red-300"
                >
                  {t('parsingFilters.editor.delete')}
                </button>
              ))}
          </div>
          {!readOnly && (
            <Button size="sm" disabled={!dirty || saving} onClick={() => void save()}>
              {saving ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : null}
              {saving ? t('parsingFilters.editor.saving') : t('parsingFilters.editor.save')}
            </Button>
          )}
        </footer>
      </div>
    </div>
  )
}

function Toggle({ checked, onChange }: { checked: boolean; onChange: () => void }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={onChange}
      className={cn(
        'relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors',
        checked ? 'bg-primary' : 'bg-muted-foreground/30',
      )}
    >
      <span className={cn('inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform', checked ? 'translate-x-4' : 'translate-x-0.5')} />
    </button>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
