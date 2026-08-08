import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Code2, LayoutList, Loader2, Lock, Plus, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { YamlCodeEditor } from '@/shared/components/YamlCodeEditor'
import { complianceService, ComplianceHttpError } from '../services/compliance-http.service'
import {
  CHECK_DATASETS,
  CHECK_OPERATORS,
  CHECK_RULES,
  LIST_OPERATORS,
  VALUELESS_OPERATORS,
  CONTROL_SCOPES,
  CONTROL_STRATEGIES,
  controlToForm,
  controlFormToYaml,
  formToControl,
  yamlToControlForm,
  type ControlFormState,
} from '../lib/control-yaml'
import type { Check, CheckFilter, CheckFilterOperator, CheckRule, Control, Dataset } from '../types/compliance.types'

const SELECT = 'h-9 w-full rounded-md border border-input bg-background px-2 text-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:opacity-60'
const TA = 'w-full resize-none rounded-md border border-input bg-background/40 p-2 text-sm focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring read-only:opacity-80'

export function ControlEditor({
  control,
  creating,
  prefillName,
  onClose,
  onSaved,
}: {
  control?: Control
  creating: boolean
  prefillName?: string
  onClose: () => void
  onSaved: (created?: Control) => void
}) {
  const { t } = useTranslation()
  // The shipped catalogue is not editable, and neither is anything the
  // licence withholds.
  const readOnly = !!control?.system || !!control?.locked
  const [form, setForm] = useState<ControlFormState>(() => {
    const f = controlToForm(control)
    if (!control && prefillName) f.name = prefillName
    return f
  })
  const [mode, setMode] = useState<'visual' | 'code'>('visual')
  const [yaml, setYaml] = useState('')
  const [busy, setBusy] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [dataTypes, setDataTypes] = useState<string[]>([])

  const set = <K extends keyof ControlFormState>(k: K, v: ControlFormState[K]) => setForm((f) => ({ ...f, [k]: v }))

  useEffect(() => {
    complianceService
      .listDataTypes('logs')
      .then((d) => setDataTypes(d ?? []))
      .catch(() => setDataTypes([]))
  }, [])

  const toCode = () => {
    setYaml(controlFormToYaml(form))
    setMode('code')
  }
  const toVisual = () => {
    const r = yamlToControlForm(yaml)
    if (!r.ok) {
      toast.error(t('compliance.controls.yamlError', { error: r.error }))
      return
    }
    setForm(r.form)
    setMode('visual')
  }

  const save = async () => {
    if (busy) return
    let f = form
    if (mode === 'code') {
      const r = yamlToControlForm(yaml)
      if (!r.ok) {
        toast.error(t('compliance.controls.yamlError', { error: r.error }))
        return
      }
      f = r.form
      setForm(f)
    }
    const payload = formToControl(f)
    if (!payload.id) {
      toast.error(t('compliance.controls.idRequired'))
      return
    }
    if (!payload.name) {
      toast.error(t('compliance.controls.nameRequired'))
      return
    }
    setBusy(true)
    try {
      if (creating) {
        const created = await complianceService.createControl(payload)
        toast.success(t('compliance.controls.toast.created'))
        onSaved(created ?? payload)
      } else {
        await complianceService.updateControl(control!.id, payload)
        toast.success(t('compliance.controls.toast.saved'))
        onSaved()
      }
    } catch (e) {
      toast.error(e instanceof ComplianceHttpError ? e.message : t('compliance.controls.toast.saveError'))
    } finally {
      setBusy(false)
    }
  }

  const remove = async () => {
    if (!control) return
    setBusy(true)
    try {
      await complianceService.deleteControl(control.id)
      toast.success(t('compliance.controls.toast.deleted'))
      onSaved()
    } catch {
      toast.error(t('compliance.controls.toast.deleteError'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div className="flex w-full max-w-[760px] flex-col overflow-hidden border-l border-border bg-card shadow-xl" onClick={(e) => e.stopPropagation()}>
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div className="min-w-0">
            <h2 className="flex items-center gap-2 truncate text-lg font-semibold">
              {creating ? t('compliance.controls.createTitle') : control?.id}
              {readOnly && (
                <span className="inline-flex items-center gap-1 rounded bg-violet-500/15 px-1.5 py-0.5 text-[10px] font-medium text-violet-500">
                  <Lock size={9} /> {t('compliance.system')}
                </span>
              )}
            </h2>
            {!creating && <p className="mt-0.5 truncate text-[11px] text-muted-foreground">{control?.name}</p>}
          </div>
          <div className="flex shrink-0 items-center gap-2">
            <div className="inline-flex rounded-md border border-border p-0.5">
              <button type="button" onClick={() => mode !== 'visual' && toVisual()} className={cn('inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors', mode === 'visual' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}>
                <LayoutList size={13} /> {t('compliance.controls.visualTab')}
              </button>
              <button type="button" onClick={() => mode !== 'code' && toCode()} className={cn('inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors', mode === 'code' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}>
                <Code2 size={13} /> {t('compliance.controls.codeTab')}
              </button>
            </div>
            <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
              <X size={16} />
            </button>
          </div>
        </header>

        {mode === 'code' ? (
          <div className="flex min-h-0 flex-1 flex-col p-6">
            <YamlCodeEditor value={yaml} onChange={setYaml} readOnly={readOnly} />
          </div>
        ) : (
          <div className="min-h-0 flex-1 space-y-4 overflow-y-auto p-6">
            <div className="grid grid-cols-2 gap-4">
              <Field label={t('compliance.controls.id')}>
                <Input value={form.id} readOnly={readOnly || !creating} onChange={(e) => set('id', e.target.value)} placeholder="ac-2" className="font-mono text-xs" />
              </Field>
              <Field label={t('compliance.controls.name')}>
                <Input value={form.name} readOnly={readOnly} onChange={(e) => set('name', e.target.value)} placeholder="Account Management" />
              </Field>
              <Field label={t('compliance.controls.family')}>
                <Input value={form.family} readOnly={readOnly} onChange={(e) => set('family', e.target.value)} placeholder="ac" className="font-mono text-xs" />
              </Field>
              <Field label={t('compliance.controls.familyName')}>
                <Input value={form.familyName} readOnly={readOnly} onChange={(e) => set('familyName', e.target.value)} placeholder="Access Control" />
              </Field>
              <Field label={t('compliance.controls.scope')}>
                <select value={form.scope} disabled={readOnly} onChange={(e) => set('scope', e.target.value)} className={SELECT}>
                  {CONTROL_SCOPES.map((s) => (
                    <option key={s} value={s}>{t(`compliance.controls.scopeOpt.${s}`)}</option>
                  ))}
                </select>
              </Field>
              <Field label={t('compliance.controls.strategy')}>
                <select value={form.strategy} disabled={readOnly} onChange={(e) => set('strategy', e.target.value)} className={SELECT}>
                  {CONTROL_STRATEGIES.map((s) => (
                    <option key={s} value={s}>{s}</option>
                  ))}
                </select>
              </Field>
            </div>
            <Field label={t('compliance.controls.statement')}>
              <textarea value={form.statement} readOnly={readOnly} onChange={(e) => set('statement', e.target.value)} rows={2} className={TA} />
            </Field>
            <Field label={t('compliance.controls.remediation')}>
              <textarea value={form.remediation} readOnly={readOnly} onChange={(e) => set('remediation', e.target.value)} rows={2} className={TA} />
            </Field>

            <ChecksEditor checks={form.checks} dataTypes={dataTypes} readOnly={readOnly} onChange={(c) => set('checks', c)} t={t} />
          </div>
        )}

        <footer className="flex items-center justify-between gap-2 border-t border-border px-6 py-3">
          <div>
            {!creating && !readOnly &&
              (confirmDelete ? (
                <div className="flex items-center gap-2">
                  <Button size="sm" variant="destructive" onClick={() => void remove()} disabled={busy}><Trash2 size={13} className="mr-1.5" /> {t('compliance.controls.confirmDelete')}</Button>
                  <Button size="sm" variant="outline" onClick={() => setConfirmDelete(false)} disabled={busy}>{t('compliance.controls.cancel')}</Button>
                </div>
              ) : (
                <button onClick={() => setConfirmDelete(true)} className="rounded-md border border-red-500/30 bg-red-500/5 px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-500/10 dark:text-red-300">{t('compliance.controls.delete')}</button>
              ))}
          </div>
          {!readOnly && (
            <Button size="sm" disabled={busy} onClick={() => void save()}>
              {busy ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : null} {t('compliance.controls.save')}
            </Button>
          )}
        </footer>
      </div>
    </div>
  )
}

function ChecksEditor({ checks, dataTypes, readOnly, onChange, t }: { checks: Check[]; dataTypes: string[]; readOnly?: boolean; onChange: (c: Check[]) => void; t: ReturnType<typeof useTranslation>['t'] }) {
  const setAt = (i: number, patch: Partial<Check>) => onChange(checks.map((c, k) => (k === i ? { ...c, ...patch } : c)))
  return (
    <div className="space-y-2">
      <label className="block text-xs font-medium text-foreground/80">{t('compliance.controls.checks')}</label>
      <p className="text-[11px] text-muted-foreground">{t('compliance.controls.checksHint')}</p>
      {checks.map((c, i) => (
        <div key={i} className="space-y-2 rounded-lg border border-border bg-card/60 p-3">
          <div className="flex items-center gap-2">
            <Input value={c.key} readOnly={readOnly} onChange={(e) => setAt(i, { key: e.target.value })} placeholder={t('compliance.controls.checkKey')} className="h-8 w-32 font-mono text-xs" />
            <Input value={c.name} readOnly={readOnly} onChange={(e) => setAt(i, { name: e.target.value })} placeholder={t('compliance.controls.checkName')} className="h-8 flex-1 text-xs" />
            {!readOnly && (
              <button type="button" onClick={() => onChange(checks.filter((_, k) => k !== i))} className="rounded p-1 text-muted-foreground hover:text-red-500"><X size={13} /></button>
            )}
          </div>
          {/* Which slice of the event store this measures. It is also the
              applicability declaration: a check against a data type the tenant
              never receives goes unevaluated rather than failing. */}
          <div className="flex flex-wrap items-center gap-2">
            <select value={c.dataset ?? 'logs'} disabled={readOnly} onChange={(e) => setAt(i, { dataset: e.target.value as Dataset, dataType: '' })} className={cn(SELECT, 'h-8 w-auto text-xs')}>
              {CHECK_DATASETS.map((d) => (
                <option key={d} value={d}>{d}</option>
              ))}
            </select>
            <select value={c.dataType ?? ''} disabled={readOnly} onChange={(e) => setAt(i, { dataType: e.target.value })} className={cn(SELECT, 'h-8 w-auto font-mono text-xs')}>
              <option value="">{t('compliance.controls.anyDataType', { defaultValue: 'Any data type' })}</option>
              {dataTypes.map((d) => (
                <option key={d} value={d}>{d}</option>
              ))}
              {c.dataType && !dataTypes.includes(c.dataType) && <option value={c.dataType}>{c.dataType}</option>}
            </select>
          </div>

          <FilterRows readOnly={!!readOnly} filters={c.filters ?? []} onChange={(filters) => setAt(i, { filters })} />

          <div className="flex flex-wrap items-center gap-2">
            <select value={c.rule ?? ''} disabled={readOnly} onChange={(e) => setAt(i, { rule: (e.target.value || undefined) as CheckRule | undefined })} className={cn(SELECT, 'h-8 w-auto text-xs')}>
              {CHECK_RULES.map((r) => (
                <option key={r} value={r}>{r ? t(`compliance.controls.rule.${r}`) : t('compliance.controls.ruleNone')}</option>
              ))}
            </select>
            {c.rule && (
              <Input type="number" value={c.ruleValue ?? ''} readOnly={readOnly} onChange={(e) => setAt(i, { ruleValue: e.target.value === '' ? undefined : Number(e.target.value) })} placeholder={t('compliance.controls.ruleValue')} className="h-8 w-28 text-xs" />
            )}
          </div>
        </div>
      ))}
      {!readOnly && (
        <Button type="button" variant="outline" size="sm" className="h-7" onClick={() => onChange([...checks, { key: '', name: '' }])}>
          <Plus size={12} className="mr-1" /> {t('compliance.controls.addCheck')}
        </Button>
      )}
    </div>
  )
}

/**
 * The check's filters. They select; the rule judges. Keeping the two apart is
 * why a check no longer carries a query dialect of its own.
 */
function FilterRows({
  filters,
  onChange,
  readOnly,
}: {
  filters: CheckFilter[]
  onChange: (f: CheckFilter[]) => void
  readOnly: boolean
}) {
  const { t } = useTranslation()
  const setAt = (i: number, patch: Partial<CheckFilter>) =>
    onChange(filters.map((f, k) => (k === i ? { ...f, ...patch } : f)))

  return (
    <div className="space-y-1.5">
      {filters.map((f, i) => {
        const isList = LIST_OPERATORS.has(f.operator)
        const valueless = VALUELESS_OPERATORS.has(f.operator)
        return (
          <div key={i} className="flex items-center gap-2">
            <Input
              value={f.field}
              readOnly={readOnly}
              onChange={(e) => setAt(i, { field: e.target.value })}
              placeholder={t('compliance.controls.filterField', { defaultValue: 'log.eventID' })}
              className="h-8 flex-1 font-mono text-xs"
            />
            <select
              value={f.operator}
              disabled={readOnly}
              onChange={(e) => setAt(i, { operator: e.target.value as CheckFilterOperator, value: undefined })}
              className={cn(SELECT, 'h-8 w-auto text-xs')}
            >
              {CHECK_OPERATORS.map((op) => (
                <option key={op} value={op}>{op}</option>
              ))}
            </select>
            {!valueless && (
              <Input
                value={Array.isArray(f.value) ? f.value.join(', ') : String(f.value ?? '')}
                readOnly={readOnly}
                onChange={(e) =>
                  setAt(i, {
                    value: isList
                      ? e.target.value.split(',').map((v) => v.trim()).filter(Boolean)
                      : e.target.value,
                  })
                }
                placeholder={isList ? t('compliance.controls.filterValues', { defaultValue: 'a, b, c' }) : t('compliance.controls.filterValue', { defaultValue: 'value' })}
                className="h-8 flex-1 font-mono text-xs"
              />
            )}
            {!readOnly && (
              <button type="button" onClick={() => onChange(filters.filter((_, k) => k !== i))} className="rounded p-1 text-muted-foreground hover:text-red-500">
                <X size={13} />
              </button>
            )}
          </div>
        )
      })}
      {!readOnly && (
        <button
          type="button"
          onClick={() => onChange([...filters, { field: '', operator: 'IS' }])}
          className="text-[11px] text-muted-foreground underline-offset-2 hover:underline"
        >
          + {t('compliance.controls.addFilter', { defaultValue: 'Add filter' })}
        </button>
      )}
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">{label}</label>
      {children}
    </div>
  )
}
