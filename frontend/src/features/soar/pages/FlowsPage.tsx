import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, History, Loader2, Lock, Plus, RefreshCw, Search, Terminal, Workflow } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Pagination } from '@/shared/components/ui/pagination'
import { soarFlowsService } from '../services/soar-flows.service'
import { FlowEditor } from '../components/FlowEditor'
import { ExecutionsView } from '../components/ExecutionsView'
import type { Flow } from '../types/soar.types'

type PageTab = 'flows' | 'executions'

export function FlowsPage() {
  const { t } = useTranslation()
  const [tab, setTab] = useState<PageTab>('flows')

  return (
    <div className="mx-auto flex h-full min-h-0 w-full max-w-[1100px] flex-col px-6 py-6">
      <header className="shrink-0">
        <h1 className="flex items-center gap-2 text-xl font-semibold">
          <Workflow size={18} strokeWidth={1.75} /> {t('soar.title')}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">{t('soar.subtitle')}</p>
        <div className="mt-4 inline-flex rounded-md border border-border p-0.5">
          <TabButton active={tab === 'flows'} onClick={() => setTab('flows')} icon={Workflow} label={t('soar.tabs.flows')} />
          <TabButton active={tab === 'executions'} onClick={() => setTab('executions')} icon={History} label={t('soar.tabs.executions')} />
        </div>
      </header>

      <div className="mt-4 flex min-h-0 flex-1 flex-col">
        {tab === 'flows' ? <FlowsTab /> : <ExecutionsView />}
      </div>
    </div>
  )
}

function TabButton({ active, onClick, icon: Icon, label }: { active: boolean; onClick: () => void; icon: typeof Workflow; label: string }) {
  return (
    <button
      onClick={onClick}
      className={cn('inline-flex items-center gap-1.5 rounded px-3 py-1.5 text-xs transition-colors', active ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}
    >
      <Icon size={13} /> {label}
    </button>
  )
}

type ListTab = 'all' | 'active' | 'inactive' | 'system' | 'user'
const LIST_TABS: ListTab[] = ['all', 'active', 'inactive', 'system', 'user']
const COLS = 'minmax(200px,1fr) 120px 88px 90px 80px 56px'

function flowName(relPath: string): string {
  return (relPath.split('/').pop() ?? relPath).replace(/\.ya?ml$/i, '')
}

function FlowsTab() {
  const { t } = useTranslation()
  const [listTab, setListTab] = useState<ListTab>('all')
  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [items, setItems] = useState<Flow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(20)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [editing, setEditing] = useState<{ flow?: Flow; creating: boolean } | null>(null)

  useEffect(() => {
    const h = setTimeout(() => {
      setDebounced(search.trim())
      setPage(0)
    }, 300)
    return () => clearTimeout(h)
  }, [search])

  const query = useMemo(() => {
    const q: { name?: string; active?: boolean; systemOwner?: boolean; page: number; size: number } = {
      name: debounced || undefined,
      page,
      size: pageSize,
    }
    if (listTab === 'active') q.active = true
    else if (listTab === 'inactive') q.active = false
    else if (listTab === 'system') q.systemOwner = true
    else if (listTab === 'user') q.systemOwner = false
    return q
  }, [debounced, listTab, page, pageSize])

  const load = useCallback(() => {
    setLoading(true)
    setError(false)
    soarFlowsService
      .list(query)
      .then((r) => {
        setItems(r.data ?? [])
        setTotal(r.total ?? 0)
      })
      .catch(() => setError(true))
      .finally(() => setLoading(false))
  }, [query])
  useEffect(() => {
    load()
  }, [load])

  const toggleActive = async (f: Flow) => {
    const next = !f.active
    setItems((list) => list.map((x) => (x.relPath === f.relPath ? { ...x, active: next } : x)))
    try {
      await soarFlowsService.setEnabled(f.relPath, next)
    } catch {
      setItems((list) => list.map((x) => (x.relPath === f.relPath ? { ...x, active: f.active } : x)))
      toast.error(t('soar.toast.activateError'))
    }
  }

  return (
    <>
      <div className="flex shrink-0 flex-wrap items-center gap-2">
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder={t('soar.search')} className="w-[260px] pl-8" />
        </div>
        <div className="inline-flex rounded-md border border-border p-0.5">
          {LIST_TABS.map((tb) => (
            <button
              key={tb}
              onClick={() => {
                setListTab(tb)
                setPage(0)
              }}
              className={cn('rounded px-2.5 py-1 text-xs transition-colors', listTab === tb ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}
            >
              {t(`soar.tabs.${tb}`)}
            </button>
          ))}
        </div>
        <Button variant="outline" size="sm" onClick={load} disabled={loading} title={t('soar.refresh')}>
          <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
        </Button>
        <div className="ml-auto">
          <Button size="sm" onClick={() => setEditing({ creating: true })}>
            <Plus size={14} className="mr-1.5" /> {t('soar.new')}
          </Button>
        </div>
      </div>

      <div className="mt-3 flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
        <div className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground" style={{ gridTemplateColumns: COLS }}>
          <div>{t('soar.cols.flow')}</div>
          <div>{t('soar.cols.platform')}</div>
          <div className="text-center">{t('soar.cols.conditions')}</div>
          <div className="text-center">{t('soar.cols.commands')}</div>
          <div className="text-center">{t('soar.cols.active')}</div>
          <div />
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading && items.length === 0 ? (
            <Center><Loader2 className="h-4 w-4 animate-spin" /> {t('soar.loading')}</Center>
          ) : error ? (
            <Center>
              <AlertTriangle size={16} className="text-amber-500" /> {t('soar.loadError')}
              <Button variant="outline" size="sm" className="ml-2" onClick={load}>{t('soar.retry')}</Button>
            </Center>
          ) : items.length === 0 ? (
            <div className="px-6 py-16 text-center text-sm text-muted-foreground">{t('soar.empty')}</div>
          ) : (
            items.map((f) => (
              <FlowRow key={f.relPath} f={f} onOpen={() => setEditing({ flow: f, creating: false })} onToggle={() => toggleActive(f)} t={t} />
            ))
          )}
        </div>
        {total > 0 && (
          <div className="shrink-0 border-t border-border px-3 py-2">
            <Pagination page={page} pageSize={pageSize} total={total} onPageChange={setPage} onPageSizeChange={(s) => { setPageSize(s); setPage(0) }} />
          </div>
        )}
      </div>

      {editing && (
        <FlowEditor
          flow={editing.flow}
          creating={editing.creating}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null)
            load()
          }}
        />
      )}
    </>
  )
}

function FlowRow({ f, onOpen, onToggle, t }: { f: Flow; onOpen: () => void; onToggle: () => void; t: ReturnType<typeof useTranslation>['t'] }) {
  return (
    <div className="grid cursor-pointer items-center gap-3 border-b border-border px-4 py-2.5 text-sm transition-colors last:border-0 hover:bg-muted/40" style={{ gridTemplateColumns: COLS }} onClick={onOpen}>
      <div className="flex min-w-0 items-center gap-2" title={f.relPath}>
        <Terminal size={14} className="shrink-0 text-muted-foreground" />
        <div className="min-w-0">
          <div className="truncate text-[13px]">{f.name || flowName(f.relPath)}</div>
          {f.description && <div className="truncate text-[11px] text-muted-foreground">{f.description}</div>}
        </div>
      </div>
      <div>
        {f.agentPlatform ? (
          <span className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px]">{f.agentPlatform}</span>
        ) : (
          <span className="text-[11px] text-muted-foreground">—</span>
        )}
      </div>
      <div className="text-center font-mono text-[11px] text-muted-foreground">{f.conditions?.length ?? 0}</div>
      <div className="text-center font-mono text-[11px] text-muted-foreground">{f.commands?.length ?? 0}</div>
      <div className="flex justify-center" onClick={(e) => e.stopPropagation()}>
        <Toggle checked={f.active} onChange={onToggle} />
      </div>
      <div className="flex items-center justify-end gap-1.5">
        {f.systemOwner && <Lock size={11} className="text-muted-foreground/60" />}
        <span className="text-[11px] text-muted-foreground">{t('soar.view')}</span>
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
      className={cn('relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors', checked ? 'bg-primary' : 'bg-muted-foreground/30')}
    >
      <span className={cn('inline-block h-4 w-4 transform rounded-full bg-white shadow transition-transform', checked ? 'translate-x-4' : 'translate-x-0.5')} />
    </button>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="flex items-center justify-center gap-2 px-6 py-16 text-sm text-muted-foreground">{children}</div>
}
