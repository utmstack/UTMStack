import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Code2, LayoutList, Loader2, Lock, Pencil, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { YamlCodeEditor } from '@/shared/components/YamlCodeEditor'
import { PlatformBroadcastButton, broadcast, BULK_PATHS } from '@/features/platform-broadcast'
import { soarFlowsService, SoarHttpError } from '../services/soar-flows.service'
import { flowToForm, formToInput, flowFormToYaml, yamlToFlowForm, type FlowFormState } from '../lib/flow-yaml'
import { clearHttpBodyErrors, firstHttpBodyError, isValidHttpUrl } from '../lib/http-node-validity'
import { type Flow } from '../types/soar.types'
import { FlowCanvas } from './FlowCanvas'
import { FlowIdentityModal } from './FlowIdentityModal'

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
  const [identityOpen, setIdentityOpen] = useState(false)

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
            <span className="truncate">{form.name.trim() || (creating ? t('soar.editor.createTitle') : (flow?.name ?? ''))}</span>
            {!readOnly && (
              <button
                type="button"
                onClick={() => setIdentityOpen(true)}
                className="flex h-6 w-6 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
                title={t('soar.editor.identityModalTitle')}
              >
                <Pencil size={13} />
              </button>
            )}
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
        <div className="min-h-0 flex-1 overflow-hidden bg-muted/10 p-4">
          <FlowCanvas
            roots={form.roots}
            nodes={form.nodes}
            conditions={form.conditions}
            readOnly={readOnly}
            onChange={(patch) => setForm((f) => ({ ...f, roots: patch.roots, nodes: patch.nodes }))}
            onConditionsChange={(c) => set('conditions', c)}
          />
        </div>
      )}

      {identityOpen && (
        <FlowIdentityModal
          name={form.name}
          description={form.description}
          maxDepth={form.maxDepth}
          readOnly={readOnly}
          onChange={(patch) => setForm((f) => ({ ...f, ...patch }))}
          onClose={() => setIdentityOpen(false)}
        />
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

