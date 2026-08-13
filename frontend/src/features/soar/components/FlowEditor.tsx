import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ArrowDown, Check, ChevronDown, Code2, LayoutList, Loader2, Lock, Play, Plus, Server, Trash2, X, Zap } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { YamlCodeEditor } from '@/shared/components/YamlCodeEditor'
import { PlatformBroadcastButton, broadcast, BULK_PATHS } from '@/features/platform-broadcast'
import { datasourcesHttpService } from '@/features/datasources/services/datasources-http.service'
import { VariablesManager } from '@/features/datasources/components/VariablesManager'
import { soarFlowsService, SoarHttpError } from '../services/soar-flows.service'
import { flowToForm, formToInput, flowFormToYaml, yamlToFlowForm, type FlowFormState } from '../lib/flow-yaml'
import { ALERT_FIELDS, COMMON_PLATFORMS, defaultShellForPlatform, shellsForPlatform } from '../lib/alert-fields'
import { COMMAND_TEMPLATES, shellKindFor } from '../lib/command-templates'
import { SOAR_MULTI_VALUE_OPERATORS, SOAR_NO_VALUE_OPERATORS, SOAR_OPERATORS, type Flow, type FlowCommand, type FlowCondition, type SoarCondition, type SoarOperator } from '../types/soar.types'

const SELECT = 'h-8 rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

interface AgentOption {
  name: string
  platform: string
}

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

  const [agents, setAgents] = useState<AgentOption[]>([])
  const scrollRef = useRef<HTMLDivElement>(null)

  const set = <K extends keyof FlowFormState>(k: K, v: FlowFormState[K]) => setForm((f) => ({ ...f, [k]: v }))

  // Load agent datasources to populate the platform + agent selects.
  useEffect(() => {
    datasourcesHttpService
      .list({ page: 1, size: 1000, kind: 'agent' })
      .then((r) =>
        setAgents(
          (r.items ?? []).map((d) => ({
            name: d.name,
            platform: typeof d.metadata?.osPlatform === 'string' ? d.metadata.osPlatform : '',
          })),
        ),
      )
      .catch(() => setAgents([]))
  }, [])

  // Platform options: common defaults + platforms seen on agents + current value.
  const platforms = useMemo(() => {
    const seen = new Set<string>()
    const out: string[] = []
    for (const p of [...COMMON_PLATFORMS, ...agents.map((a) => a.platform), form.agentPlatform]) {
      const v = p.trim()
      if (v && !seen.has(v.toLowerCase())) {
        seen.add(v.toLowerCase())
        out.push(v)
      }
    }
    return out
  }, [agents, form.agentPlatform])

  // Agents matching the selected platform (fall back to all when none match).
  const platformAgents = useMemo(() => {
    const p = form.agentPlatform.trim().toLowerCase()
    if (!p) return agents
    const m = agents.filter((a) => a.platform && (a.platform.toLowerCase().includes(p) || p.includes(a.platform.toLowerCase())))
    return m.length ? m : agents
  }, [agents, form.agentPlatform])

  const onPlatformChange = (v: string) =>
    setForm((f) => {
      const shells = shellsForPlatform(v)
      return { ...f, agentPlatform: v, shell: shells.includes(f.shell) ? f.shell : defaultShellForPlatform(v) }
    })

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
    setForm({ ...r.form, active: form.active }) // active isn't in YAML
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
    if (input.commands.length === 0) {
      toast.error(t('soar.editor.commandsRequired'))
      return
    }
    if (!input.agentPlatform) {
      toast.error(t('soar.editor.platformRequired'))
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
          <div ref={scrollRef} className="min-h-0 px-8 flex-1 overflow-y-auto bg-muted/10 px-8 py-6">
            {/* Flow identity */}
            <div className="mb-2 space-y-3 rounded-xl border border-border bg-card p-4">
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

            {/* Workflow: WHEN → ON → THEN RUN, top to bottom */}
            <div>
              <Connector />
              <WorkflowNode icon={Zap} tone="amber" badge={t('soar.editor.when')} title={t('soar.editor.whenTitle')}>
                <ConditionsEditor conditions={form.conditions} readOnly={readOnly} onChange={(c) => set('conditions', c)} t={t} />
              </WorkflowNode>

              <Connector />
              <WorkflowNode icon={Server} tone="sky" badge={t('soar.editor.on')} title={t('soar.editor.onTitle')}>
                <div className="grid grid-cols-2 gap-4">
                  <Field label={t('soar.editor.agentPlatform')} required>
                    <select value={form.agentPlatform} disabled={readOnly} onChange={(e) => onPlatformChange(e.target.value)} className={cn(SELECT, 'h-9 w-full')}>
                      <option value="">{t('soar.editor.selectPlatform')}</option>
                      {platforms.map((p) => (
                        <option key={p} value={p}>{p}</option>
                      ))}
                    </select>
                  </Field>
                  <Field label={t('soar.editor.shell')} required>
                    <select value={form.shell} disabled={readOnly || !form.agentPlatform} onChange={(e) => set('shell', e.target.value)} className={cn(SELECT, 'h-9 w-full')}>
                      <option value="">{t('soar.editor.shellAuto')}</option>
                      {shellsForPlatform(form.agentPlatform).map((s) => (
                        <option key={s} value={s}>{s}</option>
                      ))}
                    </select>
                  </Field>
                  <div className="col-span-2">
                    <RunScopeField
                      defaultAgent={form.defaultAgent}
                      excludedAgents={form.excludedAgents}
                      agents={platformAgents}
                      readOnly={readOnly}
                      onChange={(defaultAgent, excludedAgents) => setForm((f) => ({ ...f, defaultAgent, excludedAgents }))}
                      t={t}
                    />
                  </div>
                </div>
              </WorkflowNode>

              <Connector />
              <WorkflowNode icon={Play} tone="emerald" badge={t('soar.editor.thenRun')} title={t('soar.editor.thenTitle')}>
                <CommandsEditor commands={form.commands} agentPlatform={form.agentPlatform} shell={form.shell} readOnly={readOnly} onChange={(c) => set('commands', c)} scrollRef={scrollRef} t={t} />
              </WorkflowNode>
            </div>
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

function Field({ label, required, children }: { label: string; required?: boolean; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">
        {label}
        {required && <span className="ml-0.5 text-red-500">*</span>}
      </label>
      {children}
    </div>
  )
}

const NODE_TONES: Record<string, { chip: string; ring: string }> = {
  amber: { chip: 'bg-amber-500/15 text-amber-500', ring: 'border-amber-500/25' },
  sky: { chip: 'bg-sky-500/15 text-sky-500', ring: 'border-sky-500/25' },
  emerald: { chip: 'bg-emerald-500/15 text-emerald-500', ring: 'border-emerald-500/25' },
}

/** A step card in the vertical flow narrative (When → On → Then run). */
function WorkflowNode({
  icon: Icon,
  tone,
  badge,
  title,
  children,
}: {
  icon: typeof Zap
  tone: keyof typeof NODE_TONES
  badge: string
  title: string
  children: React.ReactNode
}) {
  const tn = NODE_TONES[tone]
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

/** Downward arrow connecting two workflow nodes. */
function Connector() {
  return (
    <div className="flex justify-center" aria-hidden>
      <div className="flex h-7 flex-col items-center justify-center">
        <div className="h-3 w-px bg-border" />
        <ArrowDown size={14} className="-mt-0.5 text-muted-foreground/50" />
      </div>
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

function CommandsEditor({
  commands,
  agentPlatform,
  shell,
  readOnly,
  onChange,
  scrollRef,
  t,
}: {
  commands: FlowCommand[]
  agentPlatform: string
  shell: string
  readOnly?: boolean
  onChange: (c: FlowCommand[]) => void
  scrollRef?: React.RefObject<HTMLDivElement | null>
  t: ReturnType<typeof useTranslation>['t']
}) {
  const refs = useRef<(HTMLInputElement | null)[]>([])
  const [active, setActive] = useState(0)
  const [fieldOpen, setFieldOpen] = useState(false)
  const [templatesOpen, setTemplatesOpen] = useState(false)
  const [varsOpen, setVarsOpen] = useState(false)
  const fieldRef = useRef<HTMLDivElement>(null)
  const templatesRef = useRef<HTMLDivElement>(null)
  const shellKind = shellKindFor(agentPlatform, shell)
  const prevLen = useRef(commands.length)

  useEffect(() => {
    if (commands.length > prevLen.current) {
      const el = scrollRef?.current
      if (el) requestAnimationFrame(() => el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' }))
    }
    prevLen.current = commands.length
  }, [commands.length, scrollRef])

  useEffect(() => {
    if (!fieldOpen) return
    const onDoc = (e: MouseEvent) => {
      if (fieldRef.current && !fieldRef.current.contains(e.target as Node)) setFieldOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [fieldOpen])

  useEffect(() => {
    if (!templatesOpen) return
    const onDoc = (e: MouseEvent) => {
      if (templatesRef.current && !templatesRef.current.contains(e.target as Node)) setTemplatesOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [templatesOpen])

  // Templates add a new row (they're a starting command, not a token to drop
  // into whatever's already being edited) and focus it, like "Add command".
  const insertTemplate = (command: string) => {
    const next = [...commands, { command }]
    onChange(next)
    const idx = next.length - 1
    setActive(idx)
    requestAnimationFrame(() => refs.current[idx]?.focus())
  }

  // Insert a token into the active command at the caret, like the agent console.
  const insert = (token: string) => {
    const idx = Math.min(active, Math.max(0, commands.length - 1))
    const el = refs.current[idx]
    const cur = commands[idx]?.command ?? ''
    const start = el?.selectionStart ?? cur.length
    const end = el?.selectionEnd ?? cur.length
    const next = cur.slice(0, start) + token + cur.slice(end)
    onChange(commands.map((c, k) => (k === idx ? { ...c, command: next } : c)))
    requestAnimationFrame(() => {
      const e2 = refs.current[idx]
      if (e2) {
        e2.focus()
        const pos = start + token.length
        e2.setSelectionRange(pos, pos)
      }
    })
  }

  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-foreground/80">{t('soar.editor.commands')}</label>
      <p className="text-[11px] text-muted-foreground">{t('soar.editor.commandsHint')}</p>

      {!readOnly && (
        <div className="flex flex-wrap items-center gap-1.5">
          <span className="text-[11px] text-muted-foreground">{t('soar.editor.insertInto', { n: Math.min(active, Math.max(0, commands.length - 1)) + 1 })}</span>
          <div ref={fieldRef} className="relative">
            <button type="button" onClick={() => setFieldOpen((o) => !o)} className="inline-flex items-center gap-1 rounded-md border border-border bg-card px-2 py-1 text-[11px] hover:bg-muted">
              {t('soar.editor.insertField')} <ChevronDown size={11} className="opacity-60" />
            </button>
            {fieldOpen && (
              <div className="absolute left-0 top-full z-30 mt-1 max-h-60 w-56 overflow-y-auto rounded-md border border-border bg-popover py-1 shadow-lg">
                {ALERT_FIELDS.map((af) => (
                  <button
                    key={af.field}
                    type="button"
                    onClick={() => {
                      insert(`$(${af.field})`)
                      setFieldOpen(false)
                    }}
                    className="flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted"
                  >
                    <span className="truncate">{af.label}</span>
                    <span className="shrink-0 font-mono text-[10px] text-muted-foreground">{af.field}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
          <button type="button" onClick={() => setVarsOpen(true)} className="inline-flex items-center gap-1 rounded-md border border-border bg-card px-2 py-1 text-[11px] hover:bg-muted">
            {t('soar.editor.variables')}
          </button>
        </div>
      )}

      <div className="space-y-1.5">
        {commands.map((cmd, i) => (
          <div key={i} className={i === commands.length - 1 ? 'mb-4' : undefined}>
            {i > 0 && (
              <div className="relative flex h-10 items-center justify-center">
                <div className="soar-flow-line absolute inset-y-0 left-1/2 w-[2px] -translate-x-1/2 rounded-full" />
                <select
                  value={cmd.condition ?? 'Always'}
                  disabled={readOnly}
                  onChange={(e) => onChange(commands.map((x, k) => (k === i ? { ...x, condition: e.target.value as SoarCondition } : x)))}
                  className={cn(SELECT, 'relative z-10 h-6 py-0 shadow-sm')}
                >
                  <option value="Always">{t('soar.editor.connector.always')}</option>
                  <option value="OnSuccess">{t('soar.editor.connector.success')}</option>
                  <option value="OnFailure">{t('soar.editor.connector.failure')}</option>
                </select>
              </div>
            )}
            <div className="flex items-center gap-1.5">
              <span className="shrink-0 font-mono text-[11px] text-muted-foreground">{i + 1}</span>
              <Input
                ref={(el) => {
                  refs.current[i] = el
                }}
                value={cmd.command}
                readOnly={readOnly}
                onFocus={() => setActive(i)}
                onClick={() => setActive(i)}
                onChange={(e) => onChange(commands.map((x, k) => (k === i ? { ...x, command: e.target.value } : x)))}
                placeholder='net user "$(target.user)" /active:no'
                className="h-8 flex-1 font-mono text-xs"
              />
              {!readOnly && (
                <button type="button" onClick={() => onChange(commands.filter((_, k) => k !== i))} className="rounded p-1 text-muted-foreground hover:text-red-500">
                  <X size={13} />
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      {!readOnly && (
        <div className="flex flex-wrap items-center gap-1.5">
          <Button type="button" variant="outline" size="sm" className="h-7" onClick={() => onChange([...commands, { command: '' }])}>
            <Plus size={12} className="mr-1" /> {t('soar.editor.addCommand')}
          </Button>
          <div ref={templatesRef} className="relative">
            <Button type="button" variant="outline" size="sm" className="h-7" onClick={() => setTemplatesOpen((o) => !o)}>
              <Zap size={12} className="mr-1" /> {t('soar.editor.templatesButton')} <ChevronDown size={11} className="ml-0.5 opacity-60" />
            </Button>
            {templatesOpen && (
              <div className="absolute left-0 top-full z-30 mt-1 w-72 overflow-hidden rounded-md border border-border bg-popover py-1 shadow-lg">
                {COMMAND_TEMPLATES.map((tpl) => (
                  <button
                    key={tpl.id}
                    type="button"
                    onClick={() => {
                      insertTemplate(tpl.command(shellKind))
                      setTemplatesOpen(false)
                    }}
                    className="flex w-full items-center px-3 py-1.5 text-left text-xs hover:bg-muted"
                  >
                    {t(`soar.editor.templates.${tpl.id}`)}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      {varsOpen && (
        <VariablesManager
          onClose={() => setVarsOpen(false)}
          onInsert={(name) => {
            insert(`$[variables.${name}]`)
            setVarsOpen(false)
          }}
        />
      )}
    </div>
  )
}

type RunScope = 'all' | 'specific' | 'except'

function scopeFromState(defaultAgent: string, excludedAgents: string[]): RunScope {
  if (defaultAgent) return 'specific'
  if (excludedAgents.length > 0) return 'except'
  return 'all'
}

/** Where a flow's commands run: every agent matching the platform, exactly one
 * agent, or every matching agent except a chosen few. `defaultAgent` and
 * `excludedAgents` are mutually exclusive at the data level (a single target
 * agent has nothing to exclude from), so picking a scope resets the other. */
function RunScopeField({
  defaultAgent,
  excludedAgents,
  agents,
  readOnly,
  onChange,
  t,
}: {
  defaultAgent: string
  excludedAgents: string[]
  agents: AgentOption[]
  readOnly?: boolean
  onChange: (defaultAgent: string, excludedAgents: string[]) => void
  t: ReturnType<typeof useTranslation>['t']
}) {
  // The active tab is its own state, not purely derived from defaultAgent/
  // excludedAgents: switching to "One agent" before actually picking one
  // leaves defaultAgent empty, which would otherwise be indistinguishable
  // from "all" and snap the tab back.
  const [scope, setScopeState] = useState<RunScope>(() => scopeFromState(defaultAgent, excludedAgents))
  const options = agents.map((a) => a.name)

  const setScope = (next: RunScope) => {
    if (next === scope) return
    setScopeState(next)
    if (next === 'all') onChange('', [])
    else if (next === 'specific') onChange('', [])
    else onChange('', excludedAgents)
  }

  return (
    <div className="space-y-2">
      <label className="block text-xs font-medium text-foreground/80">{t('soar.editor.whereToRun')}</label>
      <div className="inline-flex rounded-md border border-border p-0.5">
        {(['all', 'specific', 'except'] as const).map((s) => (
          <button
            key={s}
            type="button"
            disabled={readOnly}
            onClick={() => setScope(s)}
            className={cn(
              'rounded px-2.5 py-1 text-xs transition-colors disabled:cursor-not-allowed',
              scope === s ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            {t(`soar.editor.run.${s}`)}
          </button>
        ))}
      </div>

      {scope === 'all' && <p className="text-xs text-muted-foreground">{t('soar.editor.run.allHint')}</p>}

      {scope === 'specific' && (
        <select
          value={defaultAgent}
          disabled={readOnly}
          onChange={(e) => onChange(e.target.value, [])}
          className={cn(SELECT, 'h-9 w-full max-w-xs')}
        >
          <option value="">{t('soar.editor.selectAgentPlaceholder')}</option>
          {options.map((name) => (
            <option key={name} value={name}>{name}</option>
          ))}
          {defaultAgent && !options.includes(defaultAgent) && <option value={defaultAgent}>{defaultAgent}</option>}
        </select>
      )}

      {scope === 'except' && (
        <AgentMultiSelect options={options} values={excludedAgents} readOnly={readOnly} onChange={(v) => onChange('', v)} t={t} />
      )}
    </div>
  )
}

/** Searchable multi-select of agent hostnames (for excluded agents). */
function AgentMultiSelect({
  options,
  values,
  readOnly,
  onChange,
  t,
}: {
  options: string[]
  values: string[]
  readOnly?: boolean
  onChange: (v: string[]) => void
  t: ReturnType<typeof useTranslation>['t']
}) {
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const toggle = (name: string) => onChange(values.includes(name) ? values.filter((x) => x !== name) : [...values, name])
  const filtered = options.filter((o) => (q ? o.toLowerCase().includes(q.toLowerCase()) : true))

  return (
    <div ref={ref} className="relative">
      <div className="flex min-h-9 flex-wrap items-center gap-1 rounded-md border border-input bg-background/40 p-1.5">
        {values.map((v) => (
          <span key={v} className="inline-flex items-center gap-1 rounded bg-muted px-1.5 py-0.5 font-mono text-[11px]">
            {v}
            {!readOnly && (
              <button type="button" onClick={() => toggle(v)} className="text-muted-foreground hover:text-red-500">
                <X size={11} />
              </button>
            )}
          </span>
        ))}
        {!readOnly && (
          <button type="button" onClick={() => setOpen((o) => !o)} className="inline-flex items-center gap-1 px-1 text-[11px] text-muted-foreground hover:text-foreground">
            <Plus size={11} /> {values.length ? '' : t('soar.editor.excludedAgentsHint')}
            <ChevronDown size={10} className="opacity-60" />
          </button>
        )}
      </div>
      {open && !readOnly && (
        <div className="absolute left-0 top-full z-30 mt-1 w-64 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="px-2 pb-1.5 pt-1">
            <input value={q} onChange={(e) => setQ(e.target.value)} autoFocus placeholder={t('soar.editor.searchAgents')} className="h-7 w-full rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" />
          </div>
          <div className="max-h-52 overflow-y-auto">
            {filtered.length === 0 && <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('soar.editor.noAgents')}</div>}
            {filtered.map((o) => {
              const on = values.includes(o)
              return (
                <button key={o} type="button" onClick={() => toggle(o)} className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted">
                  <span className="min-w-0 flex-1 truncate font-mono">{o}</span>
                  {on && <Check size={13} className="shrink-0 text-primary" />}
                </button>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}
