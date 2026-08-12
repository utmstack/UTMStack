import { useEffect, useState } from 'react'
import type { TFunction } from 'i18next'
import { Code2, Download, FlaskConical, LayoutList, Loader2, Lock, Pencil, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { YamlCodeEditor } from '@/shared/components/YamlCodeEditor'
import { TestPlaygroundModal } from '@/features/playground/components/TestPlaygroundModal'
import {
  alertingRulesHttpService as svc,
  AlertingRulesHttpError,
  type CorrelationRule,
  type DataTypeOption,
} from '../services/alerting-rules-http.service'
import { downloadRuleYaml } from '../lib/download-rule-yaml'
import { ruleFormToYaml, yamlToRuleForm } from '../lib/rule-yaml'
import { RuleForm, ruleToForm, formToInput, type RuleFormState } from './rule-form'
import { RuleView } from './rule-view'
import { Toggle } from './toggle'
import {
  PlatformBroadcastButton,
  broadcast,
  BULK_PATHS,
  type BulkSelector,
} from '@/features/platform-broadcast'

export function RuleDrawer({
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
  const readOnly = !!rule?.systemOwner
  const [editing, setEditing] = useState(!!create)
  const [form, setForm] = useState<RuleFormState>(() => ruleToForm(rule))
  const [busy, setBusy] = useState(false)
  const [showTestModal, setShowTestModal] = useState(false)

  // Visual ↔ Code sync live: form is canonical while visual is active, YAML is
  // canonical while code is active. Whichever pane is not being edited derives
  // from the other on every keystroke, so switching is just `setMode`.
  const [mode, setMode] = useState<'visual' | 'code'>('visual')
  const [yaml, setYaml] = useState(() => ruleFormToYaml(form))

  useEffect(() => {
    if (mode === 'visual') setYaml(ruleFormToYaml(form))
  }, [form, mode])

  const onYamlChange = (next: string) => {
    setYaml(next)
    if (!showForm) return // read-only path
    const r = yamlToRuleForm(next)
    if (r.ok) setForm({ ...r.form, ruleActive: form.ruleActive })
    // Invalid YAML → keep the last-valid form; save-time will re-toast.
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

  const buildBroadcastInput = () => {
    let f = form
    if (mode === 'code') {
      const r = yamlToRuleForm(yaml)
      if (!r.ok) throw new Error(t('alertingRules.editor.yamlError', { error: r.error }))
      f = { ...r.form, ruleActive: form.ruleActive }
    }
    if (!f.name.trim()) throw new Error(t('alertingRules.editor.nameRequired'))
    if (!f.definition.trim()) throw new Error(t('alertingRules.editor.definitionRequired'))
    return formToInput(f, create ? undefined : rule?.relPath)
  }

  const onBroadcastCreate = async (selector: BulkSelector) => {
    return broadcast(BULK_PATHS.correlationRules.create, selector, buildBroadcastInput())
  }
  const onBroadcastUpdate = async (selector: BulkSelector) => {
    return broadcast(BULK_PATHS.correlationRules.update, selector, buildBroadcastInput())
  }
  const onBroadcastDelete = async (selector: BulkSelector) => {
    if (!rule) throw new Error('No rule to delete')
    return broadcast(BULK_PATHS.correlationRules.delete, selector, { relPath: rule.relPath })
  }
  const onBroadcastActivate = async (selector: BulkSelector) => {
    if (!rule) throw new Error('No rule to activate')
    return broadcast(BULK_PATHS.correlationRules.activate, selector, {
      relPath: rule.relPath,
      active: !rule.ruleActive,
    })
  }

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
            {rule && !readOnly && (
              <PlatformBroadcastButton
                label={t('platformBroadcast.button')}
                title={t('platformBroadcast.action.delete', { resource: t('platformBroadcast.resource.correlationRule') })}
                onBroadcast={onBroadcastDelete}
                variant="outline"
                size="sm"
              />
            )}
            {rule && onToggle && <Toggle on={rule.ruleActive} onChange={(v) => onToggle(rule, v)} />}
            {rule && (
              <PlatformBroadcastButton
                label={t('platformBroadcast.button')}
                title={t('platformBroadcast.action.activate', { resource: t('platformBroadcast.resource.correlationRule') })}
                onBroadcast={onBroadcastActivate}
                variant="outline"
                size="sm"
              />
            )}
            <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"><X size={16} /></button>
          </div>
        </header>

        <div className="flex min-h-0 flex-1 flex-col bg-muted/10">
          {(showForm || rule) && (
            <div className="flex shrink-0 items-center justify-end border-b border-border px-6 py-2">
              <div className="inline-flex rounded-md border border-border p-0.5">
                <button
                  type="button"
                  onClick={() => setMode('visual')}
                  className={cn(
                    'inline-flex items-center gap-1.5 rounded px-2.5 py-1 text-xs transition-colors',
                    mode === 'visual' ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground',
                  )}
                >
                  <LayoutList size={13} /> {t('alertingRules.editor.visualTab')}
                </button>
                <button
                  type="button"
                  onClick={() => setMode('code')}
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
              <YamlCodeEditor value={yaml} onChange={onYamlChange} readOnly={!showForm} />
            </div>
          ) : (
            <div className="min-h-0 flex-1 overflow-y-auto p-6">
              {showForm ? <RuleForm form={form} setForm={setForm} dataTypeOptions={dataTypeOptions} t={t} /> : rule ? <RuleView rule={rule} t={t} /> : null}
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
                {create ? (
                  <PlatformBroadcastButton
                    label={t('platformBroadcast.button')}
                    title={t('platformBroadcast.action.create', { resource: t('platformBroadcast.resource.correlationRule') })}
                    onBroadcast={onBroadcastCreate}
                    disabled={busy}
                    size="sm"
                  />
                ) : (
                  !readOnly && (
                    <PlatformBroadcastButton
                      label={t('platformBroadcast.button')}
                      title={t('platformBroadcast.action.update', { resource: t('platformBroadcast.resource.correlationRule') })}
                      onBroadcast={onBroadcastUpdate}
                      disabled={busy}
                      size="sm"
                    />
                  )
                )}
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
          titleKey="playground.titleRule"
          dataTypeOptions={form.dataTypes}
          draftContent={ruleFormToYaml(form)}
          onClose={() => setShowTestModal(false)}
        />
      )}
    </div>
  )
}
