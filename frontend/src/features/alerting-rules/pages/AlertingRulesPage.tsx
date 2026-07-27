import { useCallback, useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import {
  AlertTriangle,
  CheckCircle2,
  Code2,
  Crosshair,
  Download,
  FlaskConical,
  LayoutList,
  Loader2,
  Lock,
  Pencil,
  Plus,
  RefreshCw,
  Search,
  ShieldAlert,
  Trash2,
  Upload,
  X,
  XCircle,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'
import { ConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import { YamlCodeEditor } from '@/shared/components/YamlCodeEditor'
import { useDateFormat } from '@/shared/lib/datetime'
import { TestPlaygroundModal } from '@/features/playground/components/TestPlaygroundModal'
import {
  alertingRulesHttpService as svc,
  AlertingRulesHttpError,
  type CorrelationRule,
  type DataTypeOption,
  type ImportRulesResponse,
} from '../services/alerting-rules-http.service'
import { RuleForm, ruleToForm, formToInput, DataTypeChip, type RuleFormState } from '../components/rule-form'
import { ruleFormToYaml, yamlToRuleForm } from '../lib/rule-yaml'
import { parseCelTree, type CelNode } from '../lib/cel-tree'

const SELECT_CLS = 'h-9 rounded-md border border-border bg-background px-2 text-sm'
const COLS = '1.4fr 1fr 110px 100px 80px 48px 50px'

function maxImpact(r: { confidentiality: number; integrity: number; availability: number }): number {
  return Math.max(r.confidentiality, r.integrity, r.availability)
}
function impactKey(n: number): 'high' | 'medium' | 'low' | 'none' {
  if (n >= 3) return 'high'
  if (n === 2) return 'medium'
  if (n === 1) return 'low'
  return 'none'
}
const IMPACT_TONE: Record<string, string> = { high: 'text-red-500', medium: 'text-amber-500', low: 'text-sky-500', none: 'text-muted-foreground' }

/* ─── Page ─────────────────────────────────────────────────────────────── */

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
        <div className="min-h-0 flex-1 overflow-y-auto">
          <Table rules={rules} onOpen={setOpen} onToggle={toggleActive} t={t} />
          <InfiniteScrollSentinel
            onReach={() => setPage((p) => p + 1)}
            hasMore={rules.length < total}
            loading={loading}
            endLabel={t('common.allLoaded', { count: total })}
          />
        </div>
      )}

      {open && <RuleDrawer rule={open} dataTypeOptions={dataTypeOptions} onClose={() => setOpen(null)} onToggle={toggleActive} onDelete={remove} onSaved={() => { setOpen(null); refresh() }} t={t} />}
      {creating && <RuleDrawer create dataTypeOptions={dataTypeOptions} onClose={() => setCreating(false)} onSaved={() => { setCreating(false); refresh() }} t={t} />}
      {importResults && <ImportResultsDialog res={importResults} onClose={() => setImportResults(null)} t={t} />}
      {showTestModal && <TestPlaygroundModal mode="rule" dataTypeOptions={testDataTypes} onClose={() => setShowTestModal(false)} />}
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

/* ─── Import results dialog ─────────────────────────────────────────────── */

function ImportResultsDialog({ res, onClose, t }: { res: ImportRulesResponse; onClose: () => void; t: TFunction }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div className="flex max-h-[80vh] w-full max-w-[560px] flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="flex items-center justify-between border-b border-border px-5 py-3.5">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <Upload size={16} /> {t('alertingRules.import.resultTitle')}
          </h2>
          <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
            <X size={16} />
          </button>
        </header>

        <div className="flex shrink-0 items-center gap-4 border-b border-border bg-muted/20 px-5 py-2.5 text-xs">
          <span className="inline-flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400">
            <CheckCircle2 size={14} /> {t('alertingRules.import.approved')}: <b>{res.approved}</b>
          </span>
          <span className="inline-flex items-center gap-1.5 text-red-600 dark:text-red-400">
            <XCircle size={14} /> {t('alertingRules.import.rejected')}: <b>{res.rejected}</b>
          </span>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto p-3">
          <div className="space-y-1.5">
            {res.results.map((r, i) => (
              <div key={i} className="flex items-start gap-2 rounded-md border border-border px-3 py-2 text-xs">
                {r.approved ? (
                  <CheckCircle2 size={15} className="mt-0.5 shrink-0 text-emerald-500" />
                ) : (
                  <XCircle size={15} className="mt-0.5 shrink-0 text-red-500" />
                )}
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-x-2">
                    <span className="font-mono text-[11px] text-muted-foreground">{r.filename}</span>
                    {r.name && <span className="font-medium">{r.name}</span>}
                  </div>
                  {!r.approved && r.error && <p className="mt-0.5 text-[11px] text-red-600 dark:text-red-400">{r.error}</p>}
                </div>
              </div>
            ))}
          </div>
        </div>

        <footer className="flex justify-end border-t border-border px-5 py-3">
          <Button size="sm" onClick={onClose}>{t('alertingRules.import.done')}</Button>
        </footer>
      </div>
    </div>
  )
}

/* ─── Table ────────────────────────────────────────────────────────────── */

function Table({ rules, onOpen, onToggle, t }: { rules: CorrelationRule[]; onOpen: (r: CorrelationRule) => void; onToggle: (r: CorrelationRule, next: boolean) => void; t: TFunction }) {
  return (
    <div className="mt-4 min-h-0 flex-1 overflow-y-auto rounded-xl border border-border">
      <div className="grid items-center gap-3 border-b border-border bg-muted/30 px-4 py-2.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground" style={{ gridTemplateColumns: COLS }}>
        <div>{t('alertingRules.table.name')}</div>
        <div>{t('alertingRules.table.dataTypes')}</div>
        <div>{t('alertingRules.table.category')}</div>
        <div>{t('alertingRules.table.technique')}</div>
        <div>{t('alertingRules.table.adversary')}</div>
        <div className="text-center">{t('alertingRules.table.impact')}</div>
        <div className="text-center">{t('alertingRules.table.active')}</div>
      </div>
      {rules.map((r) => {
        const dts = (r.dataTypes ?? []).filter((d) => d.included).map((d) => d.dataType)
        return (
        <div key={r.relPath} className="grid items-center gap-3 border-b border-border/60 px-4 py-3 text-sm last:border-b-0 hover:bg-muted/30" style={{ gridTemplateColumns: COLS }}>
          <button onClick={() => onOpen(r)} className="min-w-0 text-left">
            <div className="flex items-center gap-1.5">
              <span className="truncate font-medium">{r.name}</span>
              {r.systemOwner && <Lock size={11} className="shrink-0 text-muted-foreground/50" />}
            </div>
            {r.description && <div className="truncate text-xs text-muted-foreground">{r.description}</div>}
          </button>
          <button onClick={() => onOpen(r)} className="flex min-w-0 flex-wrap items-center gap-1 text-left">
            {dts.slice(0, 2).map((dt) => <DataTypeChip key={dt} dataType={dt} />)}
            {dts.length > 2 && <span className="text-[10px] text-muted-foreground">+{dts.length - 2}</span>}
            {dts.length === 0 && <span className="text-xs text-muted-foreground/60">—</span>}
          </button>
          <button onClick={() => onOpen(r)} className="truncate text-left text-xs text-muted-foreground">{r.category || '—'}</button>
          <button onClick={() => onOpen(r)} className="truncate text-left font-mono text-xs text-muted-foreground" title={r.technique}>{r.technique || '—'}</button>
          <div className="flex items-center gap-1 text-xs text-muted-foreground"><Crosshair size={11} /> {r.adversary ? t(`alertingRules.adversary.${r.adversary}`) : '—'}</div>
          <div className="text-center"><span className={cn('font-mono text-xs font-semibold', IMPACT_TONE[impactKey(maxImpact(r))])}>{maxImpact(r)}</span></div>
          <div className="flex justify-center"><Toggle on={r.ruleActive} onChange={(v) => onToggle(r, v)} /></div>
        </div>
        )
      })}
    </div>
  )
}

function Toggle({ on, onChange }: { on: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      onClick={(e) => { e.stopPropagation(); onChange(!on) }}
      className={cn('relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors', on ? 'bg-emerald-500' : 'bg-muted-foreground/30')}
      role="switch"
      aria-checked={on}
    >
      <span className={cn('inline-block h-4 w-4 transform rounded-full bg-white transition-transform', on ? 'translate-x-4' : 'translate-x-0.5')} />
    </button>
  )
}

/* ─── Drawer (view / edit / create) ────────────────────────────────────── */

function RuleDrawer({
  rule,
  create,
  dataTypeOptions,
  onClose,
  onToggle,
  onDelete,
  onSaved,
  t,
}: {
  rule?: CorrelationRule
  create?: boolean
  dataTypeOptions: DataTypeOption[]
  onClose: () => void
  onToggle?: (r: CorrelationRule, next: boolean) => void
  onDelete?: (r: CorrelationRule) => void
  onSaved: () => void
  t: TFunction
}) {
  const df = useDateFormat()
  const readOnly = !!rule?.systemOwner
  const [editing, setEditing] = useState(!!create)
  const [form, setForm] = useState<RuleFormState>(() => ruleToForm(rule))
  const [busy, setBusy] = useState(false)
  const [showTestModal, setShowTestModal] = useState(false)

  // Visual ↔ Code. The structured form stays canonical; Code shows/edits the
  // whole rule as YAML and syncs back into the form on toggle / save.
  const [mode, setMode] = useState<'visual' | 'code'>('visual')
  const [yaml, setYaml] = useState('')

  const toCode = () => {
    setYaml(ruleFormToYaml(form))
    setMode('code')
  }
  const toVisual = () => {
    // Only sync YAML → form when actually editing; in view mode (incl. system
    // rules) the YAML is read-only, so there's nothing to apply back.
    if (showForm) {
      const r = yamlToRuleForm(yaml)
      if (!r.ok) {
        toast.error(t('alertingRules.editor.yamlError', { error: r.error }))
        return
      }
      // active isn't part of the YAML — keep the current value.
      setForm({ ...r.form, ruleActive: form.ruleActive })
    }
    setMode('visual')
  }

  const cancelEdit = () => {
    setEditing(false)
    setForm(ruleToForm(rule))
    setMode('visual')
  }

  const save = async () => {
    if (busy) return
    // In code mode the YAML is the source of truth — parse it first.
    let f = form
    if (mode === 'code') {
      const r = yamlToRuleForm(yaml)
      if (!r.ok) { toast.error(t('alertingRules.editor.yamlError', { error: r.error })); return }
      f = { ...r.form, ruleActive: form.ruleActive } // active isn't in the YAML
      setForm(f)
    }
    if (!f.name.trim()) { toast.error(t('alertingRules.editor.nameRequired')); return }
    if (!f.definition.trim()) { toast.error(t('alertingRules.editor.definitionRequired')); return }
    const input = formToInput(f, create ? undefined : rule?.relPath)
    setBusy(true)
    try {
      if (create) await svc.create(input)
      else await svc.update(input)
      toast.success(create ? t('alertingRules.toast.created') : t('alertingRules.toast.saved'))
      onSaved()
    } catch (e) {
      toast.error(e instanceof AlertingRulesHttpError ? e.message : t('alertingRules.toast.saveError'))
    } finally {
      setBusy(false)
    }
  }

  const showForm = editing || !!create

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-[780px] flex-col overflow-hidden border-l border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
              {create ? t('alertingRules.editor.createTitle') : rule?.category || t('alertingRules.title')}
              {readOnly && <span className="inline-flex items-center gap-1"><Lock size={11} /> {t('alertingRules.owner.system')}</span>}
            </div>
            <h2 className="mt-1 truncate text-xl font-semibold">{create ? t('alertingRules.editor.createTitle') : rule?.name}</h2>
          </div>
          <div className="flex shrink-0 items-center gap-2">
            {rule && !readOnly && !showForm && <Button size="sm" variant="outline" onClick={() => setEditing(true)}><Pencil size={13} className="mr-1.5" /> {t('alertingRules.editor.edit')}</Button>}
            {rule && <Button size="sm" variant="outline" onClick={() => downloadRuleYaml(rule)}><Download size={13} className="mr-1.5" /> {t('alertingRules.editor.export')}</Button>}
            {rule && !readOnly && onDelete && <button onClick={() => onDelete(rule)} title={t('alertingRules.editor.delete')} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-red-500/10 hover:text-red-500"><Trash2 size={15} /></button>}
            {rule && onToggle && <Toggle on={rule.ruleActive} onChange={(v) => onToggle(rule, v)} />}
            <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"><X size={16} /></button>
          </div>
        </header>

        <div className="flex min-h-0 flex-1 flex-col bg-muted/10">
          {(showForm || rule) && (
            <div className="flex shrink-0 items-center justify-end border-b border-border px-6 py-2">
              <div className="inline-flex rounded-md border border-border p-0.5">
                <button
                  type="button"
                  onClick={() => mode !== 'visual' && toVisual()}
                  className={cn(
                    'inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors',
                    mode === 'visual' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  <LayoutList size={13} /> {t('alertingRules.editor.visualTab')}
                </button>
                <button
                  type="button"
                  onClick={() => mode !== 'code' && toCode()}
                  className={cn(
                    'inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors',
                    mode === 'code' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  <Code2 size={13} /> {t('alertingRules.editor.codeTab')}
                </button>
              </div>
            </div>
          )}
          {mode === 'code' ? (
            // Editable only while editing/creating; system rules (and plain view)
            // are read-only.
            <div className="flex min-h-0 flex-1 flex-col p-6">
              <YamlCodeEditor value={yaml} onChange={setYaml} readOnly={!showForm} />
            </div>
          ) : (
            <div className="min-h-0 flex-1 overflow-y-auto p-6">
              {showForm ? <RuleForm form={form} setForm={setForm} dataTypeOptions={dataTypeOptions} t={t} /> : rule ? <RuleView rule={rule} df={df} t={t} /> : null}
            </div>
          )}
        </div>

        {(showForm || rule) && (
          <footer className="flex items-center justify-between gap-2 border-t border-border px-6 py-3">
            <Button size="sm" variant="outline" onClick={() => setShowTestModal(true)}>
              <FlaskConical size={13} className="mr-1.5" /> {t('alertingRules.editor.test')}
            </Button>
            {showForm && (
              <div className="flex items-center gap-2">
                {!create && <Button size="sm" variant="outline" onClick={cancelEdit} disabled={busy}>{t('alertingRules.editor.cancel')}</Button>}
                <Button size="sm" onClick={() => void save()} disabled={busy}>
                  {busy ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : null} {t('alertingRules.editor.save')}
                </Button>
              </div>
            )}
          </footer>
        )}
      </div>

      {showTestModal && (
        <TestPlaygroundModal
          mode="rule"
          dataTypeOptions={form.dataTypes}
          draftContent={ruleFormToYaml(form)}
          onClose={() => setShowTestModal(false)}
        />
      )}
    </div>
  )
}

function RuleView({ rule, df, t }: { rule: CorrelationRule; df: ReturnType<typeof useDateFormat>; t: TFunction }) {
  const steps = ruleToForm(rule).correlation
  return (
    <div className="space-y-4">
      {rule.description && <Section title={t('alertingRules.view.description')}><RuleDescription text={rule.description} /></Section>}
      <Section title={t('alertingRules.view.details')}>
        <dl className="grid grid-cols-[150px_1fr] gap-y-2 text-xs">
          <Row k={t('alertingRules.table.category')}>{rule.category || '—'}</Row>
          <Row k={t('alertingRules.table.technique')}><span className="font-mono">{rule.technique || '—'}</span></Row>
          <Row k={t('alertingRules.table.adversary')}>{rule.adversary ? t(`alertingRules.adversary.${rule.adversary}`) : '—'}</Row>
          <Row k={t('alertingRules.view.impact')}><span className="font-mono">C{rule.confidentiality} · I{rule.integrity} · A{rule.availability}</span></Row>
          <Row k={t('alertingRules.view.dataTypes')}><span className="font-mono">{(rule.dataTypes ?? []).filter((d) => d.included).map((d) => d.dataType).join(', ') || '—'}</span></Row>
          {rule.ruleLastUpdate && <Row k={t('alertingRules.view.updated')}>{df.formatDateTime(rule.ruleLastUpdate)}</Row>}
        </dl>
      </Section>
      <Section title={t('alertingRules.view.definition')}>
        <DefinitionView definition={rule.definition} t={t} />
      </Section>
      {steps.length > 0 && (
        <Section title={t('alertingRules.view.correlationSteps')}>
          <AfterEventsView steps={steps} t={t} />
        </Section>
      )}
      {(hasItems(rule.groupBy) || hasItems(rule.deduplicateBy)) && (
        <Section title={t('alertingRules.view.correlation')}>
          <dl className="grid grid-cols-[150px_1fr] gap-y-2 text-xs">
            <Row k={t('alertingRules.view.groupBy')}><span className="font-mono">{asList(rule.groupBy)}</span></Row>
            <Row k={t('alertingRules.view.deduplicateBy')}><span className="font-mono">{asList(rule.deduplicateBy)}</span></Row>
          </dl>
        </Section>
      )}
    </div>
  )
}

/* ─── Read-only visual renderers (CEL condition, correlation steps, markdown) ── */

/** Friendly read-only view of the CEL `where` condition; falls back to raw CEL. */
function DefinitionView({ definition, t }: { definition: string; t: TFunction }) {
  const tree = parseCelTree(definition)
  if (!tree) {
    return <pre className="overflow-x-auto rounded-md border border-border bg-card p-3 font-mono text-[11px] leading-relaxed">{definition || '—'}</pre>
  }
  return <CelNodeView node={tree} t={t} depth={0} />
}

/** Recursively renders a CEL boolean tree as nested AND/OR groups. */
function CelNodeView({ node, t, depth }: { node: CelNode; t: TFunction; depth: number }) {
  if (node.type === 'cond') {
    return (
      <div className="flex flex-wrap items-center gap-1.5 text-xs">
        {node.negate && <span className="rounded bg-red-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-red-500">not</span>}
        <span className="rounded bg-background px-1.5 py-0.5 font-mono text-[11px]">{node.field}</span>
        <span className="text-[11px] text-muted-foreground">{t(`alertingRules.celOp.${node.fn}`)}</span>
        {node.values.filter(Boolean).map((v, i) => (
          <span key={i} className="rounded bg-primary/10 px-1.5 py-0.5 font-mono text-[11px] text-primary">{v}</span>
        ))}
      </div>
    )
  }
  const connector = node.type === 'and' ? t('alertingRules.where.and') : t('alertingRules.where.or')
  return (
    <div className={cn('space-y-1.5', depth > 0 && 'rounded-md border border-border/60 bg-background/30 p-2')}>
      {node.negate && (
        <span className="inline-block rounded bg-red-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-red-500">not</span>
      )}
      {node.children.map((child, i) => (
        <div key={i}>
          {i > 0 && (
            <div className="mb-1.5 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground/70">{connector}</div>
          )}
          <CelNodeView node={child} t={t} depth={depth + 1} />
        </div>
      ))}
    </div>
  )
}

/** Read-only cards for the correlation (after-events) steps. */
function AfterEventsView({ steps, t }: { steps: ReturnType<typeof ruleToForm>['correlation']; t: TFunction }) {
  return (
    <div className="space-y-2">
      {steps.map((s, i) => (
        <div key={i} className="rounded-md border border-border bg-card p-3">
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-[11px] text-muted-foreground">
            <span>{t('alertingRules.editor.indexPattern')}: <span className="font-mono text-foreground">{s.indexPattern || '—'}</span></span>
            <span>{t('alertingRules.editor.within')}: <span className="font-mono text-foreground">{s.within || '—'}</span></span>
            <span>{t('alertingRules.editor.count')} ≥ <span className="font-mono text-foreground">{s.count}</span></span>
          </div>
          {s.with.length > 0 && <ConditionList conds={s.with} t={t} />}
          {s.or.map((g, gi) => (
            <div key={gi} className="mt-1.5 border-t border-border/60 pt-1.5">
              <span className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">{t('alertingRules.where.or')}</span>
              <ConditionList conds={g.with} t={t} />
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}

function ConditionList({ conds, t }: { conds: { field: string; operator: string; value: string }[]; t: TFunction }) {
  return (
    <div className="mt-1.5 space-y-1">
      {conds.map((c, i) => (
        <div key={i} className="flex flex-wrap items-center gap-1.5 text-xs">
          <span className="rounded bg-background px-1.5 py-0.5 font-mono text-[11px]">{c.field}</span>
          <span className="text-[11px] text-muted-foreground">{t(`alertingRules.operator.${c.operator}`)}</span>
          {c.value && <span className="rounded bg-primary/10 px-1.5 py-0.5 font-mono text-[11px] text-primary">{c.value}</span>}
        </div>
      ))}
    </div>
  )
}

/** Lightweight markdown for rule descriptions: **bold**, `code`, lists, paragraphs. */
function RuleDescription({ text }: { text: string }) {
  const lines = text.split('\n')
  const blocks: React.ReactNode[] = []
  let i = 0
  let key = 0
  const isOl = (l: string) => /^\d+[.)]\s+/.test(l.trim())
  const isUl = (l: string) => /^[-*]\s+/.test(l.trim())

  while (i < lines.length) {
    if (lines[i].trim() === '') {
      i++
      continue
    }
    if (isOl(lines[i])) {
      const items: string[] = []
      while (i < lines.length && isOl(lines[i])) items.push(lines[i++].trim().replace(/^\d+[.)]\s+/, ''))
      blocks.push(<ol key={key++} className="ml-5 list-decimal space-y-1">{items.map((it, j) => <li key={j}>{inlineFmt(it)}</li>)}</ol>)
      continue
    }
    if (isUl(lines[i])) {
      const items: string[] = []
      while (i < lines.length && isUl(lines[i])) items.push(lines[i++].trim().replace(/^[-*]\s+/, ''))
      blocks.push(<ul key={key++} className="ml-5 list-disc space-y-1">{items.map((it, j) => <li key={j}>{inlineFmt(it)}</li>)}</ul>)
      continue
    }
    const para: string[] = []
    while (i < lines.length && lines[i].trim() !== '' && !isOl(lines[i]) && !isUl(lines[i])) para.push(lines[i++].trim())
    blocks.push(<p key={key++}>{inlineFmt(para.join(' '))}</p>)
  }
  return <div className="space-y-2 text-xs leading-relaxed text-muted-foreground">{blocks}</div>
}

/** Inline markdown: **bold** and `code`. */
function inlineFmt(text: string): React.ReactNode[] {
  const out: React.ReactNode[] = []
  const re = /\*\*([^*]+)\*\*|`([^`]+)`/g
  let last = 0
  let key = 0
  let m: RegExpExecArray | null
  while ((m = re.exec(text))) {
    if (m.index > last) out.push(text.slice(last, m.index))
    if (m[1] != null) out.push(<strong key={key++} className="font-semibold text-foreground">{m[1]}</strong>)
    else if (m[2] != null) out.push(<code key={key++} className="rounded bg-muted px-1 py-0.5 font-mono text-[11px]">{m[2]}</code>)
    last = m.index + m[0].length
  }
  if (last < text.length) out.push(text.slice(last))
  return out
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">{title}</div>
      {children}
    </div>
  )
}

function Row({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="min-w-0">{children}</dd>
    </>
  )
}

function Center({ children }: { children: React.ReactNode }) {
  return <div className="mt-4 flex flex-1 items-center justify-center gap-2 rounded-xl border border-border bg-card text-sm text-muted-foreground">{children}</div>
}

function downloadRuleYaml(rule: CorrelationRule): void {
  const yaml = ruleFormToYaml(ruleToForm(rule))
  const base = rule.relPath.split('/').pop() || `${rule.name || 'rule'}.yaml`
  const name = /\.ya?ml$/i.test(base) ? base : `${base}.yaml`
  const url = URL.createObjectURL(new Blob([yaml], { type: 'text/yaml;charset=utf-8' }))
  const a = document.createElement('a')
  a.href = url
  a.download = name
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

function hasItems(v: unknown): boolean {
  return Array.isArray(v) ? v.length > 0 : v != null && typeof v === 'object' && Object.keys(v).length > 0
}
function asList(v: unknown): string {
  return Array.isArray(v) ? v.join(', ') || '—' : '—'
}
