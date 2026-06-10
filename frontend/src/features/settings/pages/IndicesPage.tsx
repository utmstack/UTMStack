import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  AlertTriangle,
  Boxes,
  Check,
  Database,
  Layers,
  Loader2,
  Pencil,
  Plus,
  RefreshCw,
  Search,
  Trash2,
  X,
  type LucideIcon,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Pagination } from '@/shared/components/ui/pagination'
import {
  dataViewsHttpService,
  IndexHttpError,
  indicesHttpService,
} from '../services/index-management-http.service'
import type { IndexInfo, IndexPattern, IndexPatternUpsert } from '../types/index-management.types'

type Tab = 'dataViews' | 'indices'

export function IndicesPage() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<Tab>('dataViews')
  return (
    <div className="mx-auto w-full max-w-[1100px] px-6 py-6">
      <header>
        <h1 className="flex items-center gap-2 text-xl font-semibold">
          <Boxes size={18} strokeWidth={1.75} />
          {t('indices.title')}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">{t('indices.subtitle')}</p>
        <div className="mt-4 flex items-center gap-1 border-b border-border">
          <TabBtn active={tab === 'dataViews'} onClick={() => setTab('dataViews')} icon={Layers}>
            {t('indices.tabs.dataViews')}
          </TabBtn>
          <TabBtn active={tab === 'indices'} onClick={() => setTab('indices')} icon={Database}>
            {t('indices.tabs.indices')}
          </TabBtn>
        </div>
      </header>

      {tab === 'dataViews' ? <DataViewsTab /> : <IndicesTab />}
    </div>
  )
}

function TabBtn({
  active,
  onClick,
  icon: Icon,
  children,
}: {
  active: boolean
  onClick: () => void
  icon: LucideIcon
  children: React.ReactNode
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'relative flex items-center gap-1.5 px-3 py-2 text-sm transition-colors',
        active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground',
      )}
    >
      <Icon size={14} strokeWidth={1.75} />
      {children}
      {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
    </button>
  )
}

/* ─────────────────────────────────────────────────────────────────────────
 * Data views (index patterns)
 * ───────────────────────────────────────────────────────────────────────── */

function DataViewsTab() {
  const { t } = useTranslation()
  const [items, setItems] = useState<IndexPattern[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [total, setTotal] = useState(0)
  const [dialog, setDialog] = useState<{ mode: 'create' } | { mode: 'edit'; dv: IndexPattern } | null>(null)
  const [confirmDelete, setConfirmDelete] = useState<IndexPattern | null>(null)

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const { data, total } = await dataViewsHttpService.listPaged(page, pageSize, debounced)
      setItems(data)
      setTotal(total)
    } catch {
      setError(true)
      setItems([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, debounced])

  useEffect(() => {
    void load()
  }, [load])

  const list = items ?? []

  return (
    <>
      <div className="mt-4 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[260px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('indices.dataViews.search')}
            className="h-9 pl-9"
          />
        </div>
        <Button size="sm" onClick={() => setDialog({ mode: 'create' })}>
          <Plus size={14} className="mr-1.5" />
          {t('indices.dataViews.add')}
        </Button>
      </div>

      <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: DV_COLS }}
        >
          <div>{t('indices.dataViews.colPattern')}</div>
          <div>{t('indices.dataViews.colModule')}</div>
          <div className="text-right">{t('indices.dataViews.colActions')}</div>
        </div>

        {loading && (
          <RowMessage>
            <Loader2 className="h-4 w-4 animate-spin" /> {t('indices.dataViews.loading')}
          </RowMessage>
        )}
        {!loading && error && (
          <RowMessage>
            <AlertTriangle size={16} className="text-amber-500" /> {t('indices.dataViews.loadFailed')}
            <Button variant="outline" size="sm" className="ml-2" onClick={() => void load()}>
              {t('indices.retry')}
            </Button>
          </RowMessage>
        )}
        {!loading && !error && list.length === 0 && (
          <div className="px-6 py-14 text-center text-sm text-muted-foreground">
            {debounced ? t('indices.dataViews.noSearchMatch') : t('indices.dataViews.none')}
          </div>
        )}

        {!loading &&
          !error &&
          list.map((dv) => (
            <div
              key={dv.id}
              className="grid items-center gap-3 border-b border-border px-4 py-3 text-xs last:border-b-0 hover:bg-muted/30"
              style={{ gridTemplateColumns: DV_COLS }}
            >
              <div className="flex min-w-0 items-center gap-2">
                <Layers size={13} className="shrink-0 text-muted-foreground" />
                <span className="truncate font-mono">{dv.pattern}</span>
                {dv.patternSystem && (
                  <span className="rounded bg-muted px-1.5 py-0.5 text-[9px] uppercase text-muted-foreground">
                    {t('indices.dataViews.system')}
                  </span>
                )}
              </div>
              <div className="truncate text-muted-foreground">{dv.patternModule || '—'}</div>
              <div className="flex items-center justify-end gap-1">
                <IconAction
                  label={dv.patternSystem ? t('indices.dataViews.systemLocked') : t('indices.dataViews.edit')}
                  disabled={!!dv.patternSystem}
                  onClick={() => setDialog({ mode: 'edit', dv })}
                >
                  <Pencil size={13} />
                </IconAction>
                <IconAction
                  label={dv.patternSystem ? t('indices.dataViews.systemLocked') : t('indices.dataViews.delete')}
                  danger
                  disabled={!!dv.patternSystem}
                  onClick={() => setConfirmDelete(dv)}
                >
                  <Trash2 size={13} />
                </IconAction>
              </div>
            </div>
          ))}
      </div>

      {!loading && !error && (
        <Pagination
          page={page}
          pageSize={pageSize}
          total={total}
          loading={loading}
          onPageChange={setPage}
          onPageSizeChange={(s) => {
            setPageSize(s)
            setPage(0)
          }}
        />
      )}

      {dialog && (
        <DataViewDialog
          state={dialog}
          onClose={() => setDialog(null)}
          onSaved={() => {
            setDialog(null)
            void load()
          }}
        />
      )}
      {confirmDelete && (
        <ConfirmDialog
          title={t('indices.dataViews.deleteTitle')}
          body={t('indices.dataViews.deleteBody', { pattern: confirmDelete.pattern })}
          confirmLabel={t('indices.dataViews.deleteConfirm')}
          onClose={() => setConfirmDelete(null)}
          onConfirm={async () => {
            await dataViewsHttpService.remove(confirmDelete.id)
            toast.success(t('indices.dataViews.toast.deleted'))
            setConfirmDelete(null)
            // Stepping back if we just emptied a non-first page; else reload in place.
            if (list.length === 1 && page > 0) setPage((p) => p - 1)
            else void load()
          }}
        />
      )}
    </>
  )
}

const DV_COLS = '1.8fr 1fr 90px'

function DataViewDialog({
  state,
  onClose,
  onSaved,
}: {
  state: { mode: 'create' } | { mode: 'edit'; dv: IndexPattern }
  onClose: () => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const editing = state.mode === 'edit'
  const existing = editing ? state.dv : null
  const [pattern, setPattern] = useState(existing?.pattern ?? '')
  // Module/system/active are not user-editable here (matching the legacy UI, which
  // only exposes the glob). They're preserved as-is from the existing record:
  // pattern_module is a seed-only link, system is forced false on create, and
  // is_active is curation managed elsewhere.
  const [module] = useState(existing?.patternModule ?? null)
  const [system] = useState(!!existing?.patternSystem)
  const [active] = useState(existing?.isActive ?? true)
  const [busy, setBusy] = useState(false)

  // A pattern is only creatable if its glob matches at least one real index and
  // isn't a duplicate of an existing pattern — mirroring the legacy create flow.
  const trimmed = pattern.trim()
  const [checking, setChecking] = useState(false)
  const [matchCount, setMatchCount] = useState<number | null>(null)
  const [duplicate, setDuplicate] = useState(false)

  useEffect(() => {
    const glob = trimmed
    if (!glob) {
      setMatchCount(null)
      setDuplicate(false)
      setChecking(false)
      return
    }
    setChecking(true)
    const h = setTimeout(async () => {
      try {
        const [count, exists] = await Promise.all([
          indicesHttpService.countMatching(glob),
          dataViewsHttpService.existsByPattern(glob),
        ])
        setMatchCount(count)
        // Keeping its own glob in edit mode is not a duplicate.
        setDuplicate(exists && !(editing && existing?.pattern === glob))
      } catch {
        setMatchCount(null)
        setDuplicate(false)
      } finally {
        setChecking(false)
      }
    }, 350)
    return () => clearTimeout(h)
  }, [trimmed, editing, existing?.pattern])

  const canSubmit = !!trimmed && !checking && !duplicate && matchCount !== null && matchCount > 0

  const submit = async () => {
    if (!canSubmit || busy) return
    setBusy(true)
    const payload: IndexPatternUpsert = {
      id: existing?.id,
      pattern: pattern.trim(),
      patternModule: module,
      patternSystem: system,
      isActive: active,
    }
    try {
      if (editing) await dataViewsHttpService.update(payload)
      else await dataViewsHttpService.create(payload)
      toast.success(editing ? t('indices.dataViews.toast.updated') : t('indices.dataViews.toast.created'))
      onSaved()
    } catch (err) {
      toast.error(ipError(err, t))
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title={editing ? t('indices.dataViews.form.editTitle') : t('indices.dataViews.form.addTitle')}
      icon={Layers}
      onClose={onClose}
    >
      <div className="space-y-4 px-6 py-5">
        <Field label={t('indices.dataViews.form.pattern')} hint={t('indices.dataViews.form.patternHint')}>
          <Input value={pattern} onChange={(e) => setPattern(e.target.value)} placeholder="v11-log-wineventlog-*" className="font-mono text-xs" />
          {trimmed && (
            <div className="pt-0.5 text-[11px]">
              {checking ? (
                <span className="inline-flex items-center gap-1.5 text-muted-foreground">
                  <Loader2 size={12} className="animate-spin" /> {t('indices.dataViews.form.matchChecking')}
                </span>
              ) : duplicate ? (
                <span className="inline-flex items-center gap-1.5 text-red-500">
                  <AlertTriangle size={12} /> {t('indices.dataViews.form.duplicate')}
                </span>
              ) : matchCount !== null && matchCount > 0 ? (
                <span className="inline-flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400">
                  <Check size={12} /> {t('indices.dataViews.form.matchCount', { count: matchCount })}
                </span>
              ) : matchCount === 0 ? (
                <span className="inline-flex items-center gap-1.5 text-red-500">
                  <AlertTriangle size={12} /> {t('indices.dataViews.form.matchNone')}
                </span>
              ) : null}
            </div>
          )}
        </Field>
      </div>
      <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
        <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
          {t('indices.confirm.cancel')}
        </Button>
        <Button size="sm" disabled={!canSubmit || busy} onClick={() => void submit()}>
          {busy
            ? t('indices.dataViews.form.saving')
            : editing
              ? t('indices.dataViews.form.save')
              : t('indices.dataViews.form.create')}
        </Button>
      </footer>
    </Modal>
  )
}

/* ─────────────────────────────────────────────────────────────────────────
 * Indices (real OpenSearch indices)
 * ───────────────────────────────────────────────────────────────────────── */

function IndicesTab() {
  const { t } = useTranslation()
  // page is 0-based for the Pagination component; the endpoint is 1-based (page + 1).
  const [items, setItems] = useState<IndexInfo[] | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [total, setTotal] = useState(0)
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [includeSystem, setIncludeSystem] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState<IndexInfo | null>(null)

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const { data, total } = await indicesHttpService.listPaged(page + 1, pageSize, debounced, includeSystem)
      setItems(data)
      setTotal(total)
    } catch {
      setError(true)
      setItems([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, debounced, includeSystem])

  useEffect(() => {
    void load()
  }, [load])

  return (
    <>
      <div className="mt-4 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[260px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('indices.index.filter')}
            className="h-9 pl-9 font-mono text-xs"
          />
        </div>
        <label className="flex items-center gap-1.5 text-xs text-muted-foreground">
          <input
            type="checkbox"
            checked={includeSystem}
            onChange={(e) => {
              setIncludeSystem(e.target.checked)
              setPage(0)
            }}
            className="h-3.5 w-3.5 rounded border-input"
          />
          {t('indices.index.systemIndices')}
        </label>
        <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}>
          <RefreshCw size={14} className={cn('mr-1.5', loading && 'animate-spin')} />
          {t('indices.index.refresh')}
        </Button>
      </div>

      <div className="mt-3 overflow-hidden rounded-xl border border-border bg-card">
        <div
          className="grid items-center gap-3 border-b border-border bg-muted/40 px-4 py-2 text-[10px] uppercase tracking-wider text-muted-foreground"
          style={{ gridTemplateColumns: IDX_COLS }}
        >
          <div>{t('indices.index.colIndex')}</div>
          <div>{t('indices.index.colHealth')}</div>
          <div>{t('indices.index.colStatus')}</div>
          <div className="text-right">{t('indices.index.colDocs')}</div>
          <div className="text-right">{t('indices.index.colSize')}</div>
          <div className="text-right">{t('indices.index.colActions')}</div>
        </div>

        {loading && (
          <RowMessage>
            <Loader2 className="h-4 w-4 animate-spin" /> {t('indices.index.loading')}
          </RowMessage>
        )}
        {!loading && error && (
          <RowMessage>
            <AlertTriangle size={16} className="text-amber-500" /> {t('indices.index.loadFailed')}
            <Button variant="outline" size="sm" className="ml-2" onClick={() => void load()}>
              {t('indices.retry')}
            </Button>
          </RowMessage>
        )}
        {!loading && !error && items && items.length === 0 && (
          <div className="px-6 py-14 text-center text-sm text-muted-foreground">
            {debounced ? t('indices.index.noFilterMatch') : t('indices.index.none')}
          </div>
        )}

        {!loading &&
          !error &&
          items?.map((idx) => (
            <div
              key={idx.index}
              className="grid items-center gap-3 border-b border-border px-4 py-2.5 text-xs last:border-b-0 hover:bg-muted/30"
              style={{ gridTemplateColumns: IDX_COLS }}
            >
              <div className="truncate font-mono text-[11px]" title={idx.index}>
                {idx.index}
              </div>
              <div>
                <HealthDot health={idx.health} />
              </div>
              <div className="text-muted-foreground">{idx.status}</div>
              <div className="text-right font-mono text-[11px] tabular-nums text-muted-foreground">
                {formatNum(idx['docs.count'])}
              </div>
              <div className="text-right font-mono text-[11px] text-muted-foreground">
                {idx['store.size'] || '—'}
              </div>
              <div className="flex items-center justify-end">
                <IconAction label={t('indices.dataViews.delete')} danger onClick={() => setConfirmDelete(idx)}>
                  <Trash2 size={13} />
                </IconAction>
              </div>
            </div>
          ))}
      </div>

      {!loading && !error && (
        <Pagination
          page={page}
          pageSize={pageSize}
          total={total}
          loading={loading}
          onPageChange={setPage}
          onPageSizeChange={(s) => {
            setPageSize(s)
            setPage(0)
          }}
        />
      )}

      {confirmDelete && (
        <ConfirmDialog
          title={t('indices.index.deleteTitle')}
          danger
          body={t('indices.index.deleteBody', {
            name: confirmDelete.index,
            docs: formatNum(confirmDelete['docs.count']),
          })}
          confirmLabel={t('indices.index.deleteConfirm')}
          onClose={() => setConfirmDelete(null)}
          onConfirm={async () => {
            await indicesHttpService.remove([confirmDelete.index])
            toast.success(t('indices.index.toast.deleted'))
            setConfirmDelete(null)
            if ((items?.length ?? 0) === 1 && page > 0) setPage((p) => p - 1)
            else void load()
          }}
        />
      )}
    </>
  )
}

const IDX_COLS = '1.6fr 90px 90px 120px 110px 70px'

function HealthDot({ health }: { health: string }) {
  const color =
    health === 'green'
      ? 'bg-emerald-500'
      : health === 'yellow'
        ? 'bg-amber-500'
        : health === 'red'
          ? 'bg-red-500'
          : 'bg-muted-foreground'
  return (
    <span className="inline-flex items-center gap-1.5">
      <span className={cn('h-1.5 w-1.5 rounded-full', color)} />
      <span className="capitalize text-muted-foreground">{health || '—'}</span>
    </span>
  )
}

/* ─────────────────────────────────────────────────────────────────────────
 * Shared bits
 * ───────────────────────────────────────────────────────────────────────── */

function RowMessage({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-center gap-2 px-6 py-14 text-sm text-muted-foreground">
      {children}
    </div>
  )
}

function IconAction({
  label,
  onClick,
  danger,
  disabled,
  children,
}: {
  label: string
  onClick: () => void
  danger?: boolean
  disabled?: boolean
  children: React.ReactNode
}) {
  return (
    <button
      title={label}
      aria-label={label}
      onClick={onClick}
      disabled={disabled}
      className={cn(
        'flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground',
        danger && 'hover:bg-red-500/10 hover:text-red-500',
        disabled && 'cursor-not-allowed opacity-40 hover:bg-transparent hover:text-muted-foreground',
      )}
    >
      {children}
    </button>
  )
}

function ConfirmDialog({
  title,
  body,
  confirmLabel,
  danger,
  onClose,
  onConfirm,
}: {
  title: string
  body: React.ReactNode
  confirmLabel: string
  danger?: boolean
  onClose: () => void
  onConfirm: () => Promise<void>
}) {
  const { t } = useTranslation()
  const [busy, setBusy] = useState(false)
  const run = async () => {
    setBusy(true)
    try {
      await onConfirm()
    } catch (err) {
      toast.error(ipError(err, t))
    } finally {
      setBusy(false)
    }
  }
  return (
    <Modal title={title} icon={Trash2} onClose={onClose}>
      <div className="px-6 py-5 text-sm text-muted-foreground">{body}</div>
      <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
        <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
          {t('indices.confirm.cancel')}
        </Button>
        <Button size="sm" variant={danger ? 'destructive' : 'default'} disabled={busy} onClick={() => void run()}>
          {busy ? t('indices.confirm.working') : confirmLabel}
        </Button>
      </footer>
    </Modal>
  )
}

function Modal({
  title,
  icon: Icon,
  onClose,
  children,
}: {
  title: string
  icon: LucideIcon
  onClose: () => void
  children: React.ReactNode
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-lg flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between gap-4 border-b border-border px-6 py-4">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <Icon size={17} strokeWidth={1.75} />
            {title}
          </h2>
          <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
            <X size={16} />
          </button>
        </header>
        {children}
      </div>
    </div>
  )
}

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">{label}</label>
      {children}
      {hint && <p className="text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}

function formatNum(s: string): string {
  const n = Number(s)
  return Number.isFinite(n) ? n.toLocaleString() : s || '—'
}

function ipError(err: unknown, t: TFunction): string {
  if (err instanceof IndexHttpError) {
    if (err.status === 403) return t('indices.error.noPermission')
    if (err.status === 404) return t('indices.error.notFound')
    if (err.status === 400) return err.message || t('indices.error.invalidRequest')
  }
  return err instanceof Error ? err.message : t('indices.error.failed')
}
