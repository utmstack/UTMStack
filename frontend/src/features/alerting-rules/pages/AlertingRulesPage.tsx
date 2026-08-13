import { useCallback, useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import {
  AlertTriangle,
  Download,
  FlaskConical,
  Loader2,
  Plus,
  RefreshCw,
  Search,
  ShieldAlert,
  Upload,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { ConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import { TestPlaygroundModal } from '@/features/playground/components/TestPlaygroundModal'
import {
  alertingRulesHttpService as svc,
  AlertingRulesHttpError,
  type CorrelationRule,
  type DataTypeOption,
  type ImportRulesResponse,
} from '../services/alerting-rules-http.service'
import { Center } from '../components/center'
import { ImportResultsDialog } from '../components/import-results-dialog'
import { RuleDrawer } from '../components/rule-drawer'
import { Table } from '../components/table'

const SELECT_CLS = 'h-9 rounded-md border border-border bg-background px-2 text-sm'

export function AlertingRulesPage() {
  const { t } = useTranslation()
  const [rules, setRules] = useState<CorrelationRule[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const [search, setSearch] = useState('')
  const [debounced, setDebounced] = useState('')
  const [category, setCategory] = useState('all')
  const [adversary, setAdversary] = useState('all')
  const [active, setActive] = useState<'all' | 'active' | 'inactive'>('all')
  const [ownership, setOwnership] = useState<'all' | 'system' | 'custom'>('all')
  // Seed from ?dataType on first render so the initial fetch is already filtered
  // (avoids a race where the unfiltered request resolves last and wins).
  const [dataType, setDataType] = useState(() => new URLSearchParams(window.location.search).get('dataType') ?? 'all')

  const [page, setPage] = useState(0)
  const [pageSize] = useState(50)
  const [nonce, setNonce] = useState(0)

  const [categories, setCategories] = useState<string[]>([])
  const [dataTypeOptions, setDataTypeOptions] = useState<DataTypeOption[]>([])
  const [open, setOpen] = useState<CorrelationRule | null>(null)
  const [creating, setCreating] = useState(false)
  const [pendingDelete, setPendingDelete] = useState<CorrelationRule | null>(null)
  const [deleting, setDeleting] = useState(false)

  const fileInputRef = useRef<HTMLInputElement>(null)
  const [importBusy, setImportBusy] = useState(false)
  const [importResults, setImportResults] = useState<ImportRulesResponse | null>(null)
  const [showTestModal, setShowTestModal] = useState(false)
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [exportBusy, setExportBusy] = useState(false)

  const toggleSelected = useCallback((relPath: string) => {
    setSelected((cur) => {
      const next = new Set(cur)
      if (next.has(relPath)) next.delete(relPath)
      else next.add(relPath)
      return next
    })
  }, [])

  const exportSelected = async () => {
    if (exportBusy) return
    setExportBusy(true)
    try {
      const blob = await svc.exportRules([...selected])
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `alerting-rules-${new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-')}.zip`
      document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
    } catch (e) {
      toast.error(e instanceof AlertingRulesHttpError ? e.message : t('alertingRules.export.error'))
    } finally {
      setExportBusy(false)
    }
  }

  const onImportFiles = async (fileList: FileList | null) => {
    if (!fileList || fileList.length === 0) return
    setImportBusy(true)
    try {
      const files = await Promise.all(
        Array.from(fileList).map(async (f) => ({ filename: f.name, content: await f.text() })),
      )
      const res = await svc.importRules(files)
      setImportResults(res)
      if (res.approved > 0) refresh()
    } catch (e) {
      toast.error(e instanceof AlertingRulesHttpError ? e.message : t('alertingRules.import.error'))
    } finally {
      setImportBusy(false)
      if (fileInputRef.current) fileInputRef.current.value = '' // re-allow same selection
    }
  }

  // Deep-link: ?dataType=<value> pre-filters the list to that data type
  // (e.g. opened from an integration's "Rules" button).
  const [searchParams] = useSearchParams()
  useEffect(() => {
    const dt = searchParams.get('dataType')
    if (dt) {
      setDataType(dt)
      setPage(0)
    }
  }, [searchParams])

  useEffect(() => {
    const h = setTimeout(() => { setDebounced(search.trim()); setPage(0) }, 300)
    return () => clearTimeout(h)
  }, [search])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(false)
    svc
      .list({
        ruleName: debounced || undefined,
        ruleCategory: category === 'all' ? undefined : [category],
        ruleAdversary: adversary === 'all' ? undefined : [adversary],
        ruleActive: active === 'all' ? undefined : active === 'active',
        systemOwner: ownership === 'all' ? undefined : ownership === 'system',
        dataTypes: dataType === 'all' ? undefined : [dataType],
        page,
        size: pageSize,
      })
      .then(({ data, total }) => {
        if (cancelled) return
        setRules((prev) => (page === 0 ? (data ?? []) : [...prev, ...(data ?? [])]))
        setTotal(total)
      })
      .catch(() => !cancelled && setError(true))
      .finally(() => !cancelled && setLoading(false))
    return () => { cancelled = true }
  }, [debounced, category, adversary, active, ownership, dataType, page, pageSize, nonce])

  // Filter dropdown sources: distinct categories + the data-type catalog.
  useEffect(() => {
    svc.propertyValues('rule_category').then(setCategories).catch(() => {})
    svc.dataTypes().then((d) => setDataTypeOptions(d ?? [])).catch(() => {})
  }, [nonce])

  const refresh = useCallback(() => setNonce((n) => n + 1), [])

  const toggleActive = async (r: CorrelationRule, next: boolean) => {
    try {
      await svc.setActive(r.relPath, next)
      setRules((cur) => cur.map((x) => (x.relPath === r.relPath ? { ...x, ruleActive: next } : x)))
      setOpen((prev) => (prev && prev.relPath === r.relPath ? { ...prev, ruleActive: next } : prev))
      toast.success(next ? t('alertingRules.toast.activated') : t('alertingRules.toast.deactivated'))
    } catch (e) {
      toast.error(e instanceof AlertingRulesHttpError ? e.message : t('alertingRules.toast.toggleError'))
    }
  }

  // Test entry point (Entry A) scope: dataTypes with ≥1 active rule among the
  const testDataTypes = [
    ...new Set(
      rules
        .filter((r) => r.ruleActive)
        .flatMap((r) => (r.dataTypes ?? []).filter((d) => d.included).map((d) => d.dataType)),
    ),
  ]

  const remove = (r: CorrelationRule) => setPendingDelete(r)

  const confirmDelete = async () => {
    if (!pendingDelete) return
    setDeleting(true)
    try {
      await svc.remove(pendingDelete.relPath)
      toast.success(t('alertingRules.toast.deleted'))
      setOpen(null)
      refresh()
    } catch (e) {
      toast.error(e instanceof AlertingRulesHttpError ? e.message : t('alertingRules.toast.deleteError'))
    } finally {
      setPendingDelete(null)
      setDeleting(false)
    }
  }

  return (
    <div className="flex h-full min-h-0 w-full flex-col px-6 pb-6 pt-3">
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <ShieldAlert size={14} strokeWidth={1.75} />
          <span><span className="font-medium text-foreground">{total}</span> {t('alertingRules.title').toLowerCase()}</span>
        </div>
        <div className="flex items-center gap-2">
          <input
            ref={fileInputRef}
            type="file"
            accept=".yaml,.yml"
            multiple
            className="hidden"
            onChange={(e) => void onImportFiles(e.target.files)}
          />
          <Button size="sm" variant="outline" disabled={importBusy} onClick={() => fileInputRef.current?.click()}>
            {importBusy ? <Loader2 size={14} className="mr-1.5 animate-spin" /> : <Upload size={14} className="mr-1.5" />}
            {t('alertingRules.import.button')}
          </Button>
          <Button size="sm" variant="outline" disabled={exportBusy} onClick={() => void exportSelected()}>
            {exportBusy ? <Loader2 size={14} className="mr-1.5 animate-spin" /> : <Download size={14} className="mr-1.5" />}
            {selected.size > 0 ? t('alertingRules.export.selected', { count: selected.size }) : t('alertingRules.export.all')}
          </Button>
          <Button size="sm" variant="outline" onClick={() => setShowTestModal(true)}>
            <FlaskConical size={14} className="mr-1.5" />
            {t('alertingRules.test')}
          </Button>
          <Button size="sm" onClick={() => setCreating(true)}>
            <Plus size={14} className="mr-1.5" /> {t('alertingRules.new')}
          </Button>
          <button onClick={refresh} title={t('alertingRules.refresh')} className="flex h-9 w-9 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground">
            <RefreshCw size={14} className={cn(loading && 'animate-spin')} />
          </button>
        </div>
      </header>

      {/* Toolbar */}
      <div className="mt-4 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[220px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input placeholder={t('alertingRules.toolbar.search')} value={search} onChange={(e) => setSearch(e.target.value)} className="h-9 pl-9" />
        </div>
        <select value={dataType} onChange={(e) => { setDataType(e.target.value); setPage(0) }} className={SELECT_CLS}>
          <option value="all">{t('alertingRules.toolbar.allDataTypes')}</option>
          {dataTypeOptions.map((d) => <option key={d.dataType} value={d.dataType}>{d.dataType}</option>)}
        </select>
        <select value={category} onChange={(e) => { setCategory(e.target.value); setPage(0) }} className={SELECT_CLS}>
          <option value="all">{t('alertingRules.toolbar.allCategories')}</option>
          {categories.map((c) => <option key={c} value={c}>{c}</option>)}
        </select>
        <select value={adversary} onChange={(e) => { setAdversary(e.target.value); setPage(0) }} className={SELECT_CLS}>
          <option value="all">{t('alertingRules.toolbar.allAdversaries')}</option>
          <option value="origin">{t('alertingRules.adversary.origin')}</option>
          <option value="target">{t('alertingRules.adversary.target')}</option>
        </select>
        <select value={active} onChange={(e) => { setActive(e.target.value as 'all' | 'active' | 'inactive'); setPage(0) }} className={SELECT_CLS}>
          <option value="all">{t('alertingRules.toolbar.allStates')}</option>
          <option value="active">{t('alertingRules.state.active')}</option>
          <option value="inactive">{t('alertingRules.state.inactive')}</option>
        </select>
        <select value={ownership} onChange={(e) => { setOwnership(e.target.value as 'all' | 'system' | 'custom'); setPage(0) }} className={SELECT_CLS}>
          <option value="all">{t('alertingRules.toolbar.allOwners')}</option>
          <option value="system">{t('alertingRules.owner.system')}</option>
          <option value="custom">{t('alertingRules.owner.custom')}</option>
        </select>
      </div>

      {/* Content */}
      {error ? (
        <Center>
          <AlertTriangle size={16} className="text-amber-500" /> {t('alertingRules.loadError')}
          <button onClick={refresh} className="ml-2 text-primary hover:underline">{t('alertingRules.retry')}</button>
        </Center>
      ) : loading && rules.length === 0 ? (
        <Center><Loader2 className="h-5 w-5 animate-spin text-muted-foreground" /></Center>
      ) : rules.length === 0 ? (
        <Center>{t('alertingRules.empty')}</Center>
      ) : (
        <Table
          rules={rules}
          selected={selected}
          onToggleSelected={toggleSelected}
          onSelectAll={(v) => setSelected(v ? new Set(rules.map((r) => r.relPath)) : new Set())}
          onOpen={setOpen}
          onToggle={toggleActive}
          t={t}
          footer={
            <InfiniteScrollSentinel
              onReach={() => setPage((p) => p + 1)}
              hasMore={rules.length < total}
              loading={loading}
              endLabel={t('common.allLoaded', { count: total })}
            />
          }
        />
      )}

      {open && <RuleDrawer rule={open} dataTypeOptions={dataTypeOptions} onClose={() => setOpen(null)} onToggle={toggleActive} onDelete={remove} onSaved={() => { setOpen(null); refresh() }} t={t} />}
      {creating && <RuleDrawer create dataTypeOptions={dataTypeOptions} onClose={() => setCreating(false)} onSaved={() => { setCreating(false); refresh() }} t={t} />}
      {importResults && <ImportResultsDialog res={importResults} onClose={() => setImportResults(null)} t={t} />}
      {showTestModal && <TestPlaygroundModal mode="rule" titleKey="playground.titleRule" dataTypeOptions={testDataTypes} onClose={() => setShowTestModal(false)} />}
      <ConfirmDialog
        open={pendingDelete != null}
        title={t('alertingRules.editor.delete') ?? 'Delete'}
        body={pendingDelete ? t('alertingRules.deleteConfirm', { name: pendingDelete.name }) : ''}
        confirmLabel={t('alertingRules.editor.delete') ?? undefined}
        danger
        busy={deleting}
        onClose={() => !deleting && setPendingDelete(null)}
        onConfirm={confirmDelete}
      />
    </div>
  )
}
