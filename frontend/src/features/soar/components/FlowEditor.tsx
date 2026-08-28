import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Code2, LayoutList, Loader2, Lock, Play, Plus, Trash2, X, Zap } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { YamlCodeEditor } from '@/shared/components/YamlCodeEditor'
import { PlatformBroadcastButton, broadcast, BULK_PATHS } from '@/features/platform-broadcast'
import { soarFlowsService, SoarHttpError } from '../services/soar-flows.service'
import { flowToForm, formToInput, flowFormToYaml, yamlToFlowForm, type FlowFormState } from '../lib/flow-yaml'
import { clearHttpBodyErrors, firstHttpBodyError, isValidHttpUrl } from '../lib/http-node-validity'
import { ALERT_FIELDS } from '../lib/alert-fields'
import {
  SOAR_MULTI_VALUE_OPERATORS,
  SOAR_NO_VALUE_OPERATORS,
  SOAR_OPERATORS,
  type Flow,
  type FlowCondition,
  type SoarOperator,
} from '../types/soar.types'
import { FlowCanvas } from './FlowCanvas'

const SELECT = 'h-8 rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

export function FlowEditor({
  flow,
  creating,
  onClose,
  onSaved,
}: {
  flow?: Flow
  creating: boolean
  onClose: () => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const readOnly = !!flow?.systemOwner
  const [form, setForm] = useState<FlowFormState>(() => flowToForm(flow))
  const [mode, setMode] = useState<'visual' | 'code'>('visual')
  const [yaml, setYaml] = useState('')
  const [busy, setBusy] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const set = <K extends keyof FlowFormState>(k: K, v: FlowFormState[K]) => setForm((f) => ({ ...f, [k]: v }))

  useEffect(() => {
    clearHttpBodyErrors()
    return () => clearHttpBodyErrors()
  }, [flow?.relPath])

  const toCode = () => {
    setYaml(flowFormToYaml(form))
    setMode('code')
  }
  const toVisual = () => {
    const r = yamlToFlowForm(yaml)
    if (!r.ok) {
      toast.error(t('soar.editor.yamlError', { error: r.error }))
      return
    }
    setForm({ ...r.form, active: form.active })
    setMode('visual')
  }

  const save = async () => {
    if (busy) return
    let f = form
    if (mode === 'code') {
      const r = yamlToFlowForm(yaml)
      if (!r.ok) {
        toast.error(t('soar.editor.yamlError', { error: r.error }))
        return
      }
      f = { ...r.form, active: form.active }
      setForm(f)
    }
    const input = formToInput(f)
    if (!input.name) {
      toast.error(t('soar.editor.nameRequired'))
      return
    }
    if (input.conditions.length === 0) {
      toast.error(t('soar.editor.conditionsRequired'))
      return
    }
    if (input.roots.length === 0 || Object.keys(input.nodes).length === 0) {
      toast.error(t('soar.editor.nodesRequired', 'Add at least one node and connect it to the trigger.'))
      return
    }
    for (const [id, n] of Object.entries(input.nodes)) {
      if (n.executor === 'http') {
        const url = (n.params as { url?: string } | undefined)?.url ?? ''
        if (!isValidHttpUrl(url)) {
          toast.error(t('soar.editor.httpUrlInvalid', { id }))
          return
        }
      }
      if (n.executor === 'incident') {
        const name = (n.params as { name?: string } | undefined)?.name?.trim() ?? ''
        if (!name) {
          toast.error(t('soar.editor.incidentNameRequired', { id }))
          return
        }
      }
      if (n.executor === 'mail') {
        const mp = (n.params as { to?: string; subject?: string } | undefined) ?? {}
        const to = (mp.to ?? '').split(',').map((s) => s.trim()).filter(Boolean)
        if (to.length === 0) {
          toast.error(t('soar.editor.mailToRequired', { id }))
          return
        }
        if (!(mp.subject ?? '').trim()) {
          toast.error(t('soar.editor.mailSubjectRequired', { id }))
          return
        }
      }
    }
    const bodyErr = firstHttpBodyError()
    if (bodyErr) {
      toast.error(t('soar.editor.httpBodyInvalid', { id: bodyErr.nodeId, error: bodyErr.err }))
      return
    }
    setBusy(true)
    try {
      if (creating) await soarFlowsService.create(input)
      else await soarFlowsService.update(flow!.relPath, input)
      toast.success(creating ? t('soar.toast.created') : t('soar.toast.saved'))
      onSaved()
    } catch (e) {
      toast.error(e instanceof SoarHttpError ? e.message : t('soar.toast.saveError'))
    } finally {
      setBusy(false)
    }
  }

  const remove = async () => {
    if (!flow) return
    setBusy(true)
    try {
      await soarFlowsService.remove(flow.relPath)
      toast.success(t('soar.toast.deleted'))
      onSaved()
    } catch {
      toast.error(t('soar.toast.deleteError'))
    } finally {
      setBusy(false)
    }
  }

  const broadcastCreate = async (selector: { tenantIds: string[]; allTenants: boolean }) => {
    let f = form
    if (mode === 'code') {
      const r = yamlToFlowForm(yaml)
      if (!r.ok) throw new Error(r.error)
      f = { ...r.form, active: form.active }
    }
    const input = formToInput(f)
    return broadcast(BULK_PATHS.soarRules.create, selector, { rule: input })
  }

  const broadcastUpdate = async (selector: { tenantIds: string[]; allTenants: boolean }) => {
    if (!flow) throw new Error('No flow to update')
    let f = form
    if (mode === 'code') {
      const r = yamlToFlowForm(yaml)
      if (!r.ok) throw new Error(r.error)
      f = { ...r.form, active: form.active }
    }
    const input = formToInput(f)
    return broadcast(BULK_PATHS.soarRules.update, selector, { relPath: flow.relPath, rule: input })
  }

  const broadcastDelete = async (selector: { tenantIds: string[]; allTenants: boolean }) => {
    if (!flow) throw new Error('No flow to delete')
    return broadcast(BULK_PATHS.soarRules.delete, selector, { relPath: flow.relPath })
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-card">
      <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
        <div className="min-w-0">
          <h2 className="flex items-center gap-2 truncate text-lg font-semibold">
            {creating ? t('soar.editor.createTitle') : flow?.name}
            {readOnly && (
              <span className="inline-flex items-center gap-1 rounded bg-violet-500/15 px-1.5 py-0.5 text-[10px] font-medium text-violet-500">
                <Lock size={9} /> {t('soar.system')}
              </span>
            )}
          </h2>
          {!creating && <p className="mt-0.5 truncate font-mono text-[11px] text-muted-foreground">{flow?.relPath}</p>}
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <div className="inline-flex rounded-md border border-border p-0.5">
            <button
              type="button"
              onClick={() => mode !== 'visual' && toVisual()}
              className={cn('inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors', mode === 'visual' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}
            >
              <LayoutList size={13} /> {t('soar.editor.visualTab')}
            </button>
            <button
              type="button"
              onClick={() => mode !== 'code' && toCode()}
              className={cn('inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors', mode === 'code' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}
            >
              <Code2 size={13} /> {t('soar.editor.codeTab')}
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
        <div className="min-h-0 flex-1 overflow-y-auto bg-muted/10 px-6 py-4 space-y-4">
          {/* Identity + flow-level knobs */}
          <div className="space-y-3 rounded-xl border border-border bg-card p-4">
            <div className="grid grid-cols-[1fr_120px] gap-3">
              <div className="space-y-1">
                <label className="block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                  {t('soar.editor.nameLabel')}
                  {!readOnly && <span className="ml-0.5 text-red-500">*</span>}
                </label>
                <Input
                  value={form.name}
                  readOnly={readOnly}
                  onChange={(e) => set('name', e.target.value)}
                  placeholder={t('soar.editor.namePlaceholder')}
                  className="text-base font-semibold"
                />
              </div>
              <div className="space-y-1">
                <label className="block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                  {t('soar.editor.maxDepth', 'Max depth')}
                </label>
                <Input
                  type="number"
                  min={1}
                  max={1000}
                  value={form.maxDepth}
                  readOnly={readOnly}
                  onChange={(e) => set('maxDepth', Number(e.target.value) || 50)}
                  className="text-sm"
                />
              </div>
            </div>
            <div className="space-y-1">
              <label className="block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                {t('soar.editor.descriptionLabel')}
              </label>
              <textarea
                value={form.description}
                readOnly={readOnly}
                onChange={(e) => set('description', e.target.value)}
                rows={2}
                placeholder={t('soar.editor.descriptionHint')}
                className="w-full resize-none rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
            </div>
          </div>

          {/* WHEN — matching conditions */}
          <SectionCard icon={Zap} tone="amber" badge={t('soar.editor.when')} title={t('soar.editor.whenTitle')}>
            <ConditionsEditor conditions={form.conditions} readOnly={readOnly} onChange={(c) => set('conditions', c)} t={t} />
          </SectionCard>

          {/* THEN — the DAG canvas */}
          <SectionCard icon={Play} tone="emerald" badge={t('soar.editor.thenRun')} title={t('soar.editor.thenTitle')}>
            <FlowCanvas
              roots={form.roots}
              nodes={form.nodes}
              readOnly={readOnly}
              onChange={(patch) => setForm((f) => ({ ...f, roots: patch.roots, nodes: patch.nodes }))}
            />
          </SectionCard>
        </div>
      )}

      <footer className="flex items-center justify-between gap-2 border-t border-border px-6 py-3">
        <div>
          {!creating && !readOnly &&
            (confirmDelete ? (
              <div className="flex items-center gap-2">
                <Button size="sm" variant="destructive" onClick={() => void remove()} disabled={busy}>
                  <Trash2 size={13} className="mr-1.5" /> {t('soar.editor.confirmDelete')}
                </Button>
                <PlatformBroadcastButton
                  label={t('platformBroadcast.button')}
                  title={t('platformBroadcast.action.delete', { resource: t('platformBroadcast.resource.soarFlow') })}
                  onBroadcast={broadcastDelete}
                  disabled={busy}
                  size="sm"
                  variant="destructive"
                />
                <Button size="sm" variant="outline" onClick={() => setConfirmDelete(false)} disabled={busy}>{t('soar.editor.cancel')}</Button>
              </div>
            ) : (
              <button onClick={() => setConfirmDelete(true)} className="rounded-md border border-red-500/30 bg-red-500/5 px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-500/10 dark:text-red-300">
                {t('soar.editor.delete')}
              </button>
            ))}
        </div>
        {!readOnly && (
          <div className="flex items-center gap-2">
            {!creating && (
              <PlatformBroadcastButton
                label={t('platformBroadcast.button')}
                title={t('platformBroadcast.action.update', { resource: t('platformBroadcast.resource.soarFlow') })}
                onBroadcast={broadcastUpdate}
                disabled={busy}
                size="sm"
              />
            )}
            {creating && (
              <PlatformBroadcastButton
                label={t('platformBroadcast.button')}
                title={t('platformBroadcast.action.create', { resource: t('platformBroadcast.resource.soarFlow') })}
                onBroadcast={broadcastCreate}
                disabled={busy}
                size="sm"
              />
            )}
            <Button size="sm" disabled={busy} onClick={() => void save()}>
              {busy ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : null} {t('soar.editor.save')}
            </Button>
          </div>
        )}
      </footer>
    </div>
  )
}

const SECTION_TONES: Record<string, { chip: string; ring: string }> = {
  amber: { chip: 'bg-amber-500/15 text-amber-500', ring: 'border-amber-500/25' },
  sky: { chip: 'bg-sky-500/15 text-sky-500', ring: 'border-sky-500/25' },
  emerald: { chip: 'bg-emerald-500/15 text-emerald-500', ring: 'border-emerald-500/25' },
}

function SectionCard({
  icon: Icon,
  tone,
  badge,
  title,
  children,
}: {
  icon: typeof Zap
  tone: keyof typeof SECTION_TONES
  badge: string
  title: string
  children: React.ReactNode
}) {
  const tn = SECTION_TONES[tone]
  return (
    <div className={cn('rounded-xl border bg-card shadow-sm', tn.ring)}>
      <div className="flex items-center gap-2.5 border-b border-border px-4 py-2.5">
        <span className={cn('flex h-7 w-7 shrink-0 items-center justify-center rounded-lg', tn.chip)}>
          <Icon size={15} />
        </span>
        <div className="min-w-0">
          <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">{badge}</div>
          <div className="truncate text-sm font-medium">{title}</div>
        </div>
      </div>
      <div className="p-4">{children}</div>
    </div>
  )
}

function ConditionsEditor({
  conditions,
  readOnly,
  onChange,
  t,
}: {
  conditions: FlowCondition[]
  readOnly?: boolean
  onChange: (c: FlowCondition[]) => void
  t: ReturnType<typeof useTranslation>['t']
}) {
  const setAt = (i: number, patch: Partial<FlowCondition>) => onChange(conditions.map((c, k) => (k === i ? { ...c, ...patch } : c)))
  const valueStr = (c: FlowCondition) => (Array.isArray(c.value) ? c.value.join(', ') : c.value == null ? '' : String(c.value))
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">{t('soar.editor.conditions')}</label>
      <p className="text-[11px] text-muted-foreground">{t('soar.editor.conditionsHint')}</p>
      <div className="space-y-1.5">
        {conditions.map((c, i) => (
          <div key={i} className="flex flex-wrap items-center gap-1.5">
            <select value={c.field} disabled={readOnly} onChange={(e) => setAt(i, { field: e.target.value })} className={cn(SELECT, 'h-8 min-w-[150px] flex-1')}>
              <option value="">{t('soar.editor.selectField')}</option>
              {ALERT_FIELDS.map((af) => (
                <option key={af.field} value={af.field}>{af.label}</option>
              ))}
              {c.field && !ALERT_FIELDS.some((af) => af.field === c.field) && <option value={c.field}>{c.field}</option>}
            </select>
            <select value={c.operator} disabled={readOnly} onChange={(e) => setAt(i, { operator: e.target.value as SoarOperator })} className={SELECT}>
              {SOAR_OPERATORS.map((op) => (
                <option key={op} value={op}>{t(`soar.operator.${op}`)}</option>
              ))}
            </select>
            {!SOAR_NO_VALUE_OPERATORS.includes(c.operator) && (
              <Input value={valueStr(c)} readOnly={readOnly} onChange={(e) => setAt(i, { value: e.target.value })} placeholder={SOAR_MULTI_VALUE_OPERATORS.includes(c.operator) ? t('soar.editor.valueList') : t('soar.editor.value')} className="h-8 min-w-[140px] flex-1 font-mono text-xs" />
            )}
            {!readOnly && (
              <button type="button" onClick={() => onChange(conditions.filter((_, k) => k !== i))} className="rounded p-1 text-muted-foreground hover:text-red-500">
                <X size={13} />
              </button>
            )}
          </div>
        ))}
      </div>
      {!readOnly && (
        <Button type="button" variant="outline" size="sm" className="h-7" onClick={() => onChange([...conditions, { operator: 'IS', field: '', value: '' }])}>
          <Plus size={12} className="mr-1" /> {t('soar.editor.addCondition')}
        </Button>
      )}
    </div>
  )
}
