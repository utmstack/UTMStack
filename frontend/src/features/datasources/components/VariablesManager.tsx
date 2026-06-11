import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { CornerDownLeft, KeyRound, Loader2, Pencil, Plus, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { soarVariablesService, type SoarVariable } from '@/features/soar/services/soar-variables.service'

interface FormState {
  id: number | null // null = creating
  variableName: string
  variableDescription: string
  variableValue: string
  isSecret: boolean
}

const EMPTY: FormState = { id: null, variableName: '', variableDescription: '', variableValue: '', isSecret: false }

/**
 * Manage the SOAR command variables an analyst can reference in console commands
 * as `$[variables.NAME]`. Launched from the agent console.
 */
export function VariablesManager({ onClose, onInsert }: { onClose: () => void; onInsert?: (variableName: string) => void }) {
  const { t } = useTranslation()
  const [items, setItems] = useState<SoarVariable[]>([])
  const [loading, setLoading] = useState(true)
  const [form, setForm] = useState<FormState | null>(null)
  const [saving, setSaving] = useState(false)

  const load = () => {
    setLoading(true)
    soarVariablesService
      .list()
      .then((v) => setItems(v ?? []))
      .catch(() => toast.error(t('datasources.console.vars.loadError')))
      .finally(() => setLoading(false))
  }
  useEffect(load, []) // eslint-disable-line react-hooks/exhaustive-deps

  const startEdit = (v: SoarVariable) =>
    setForm({
      id: v.id,
      variableName: v.variableName ?? '',
      variableDescription: v.variableDescription ?? '',
      variableValue: '', // never prefill (secrets are masked; mask-preserving on save)
      isSecret: v.isSecret,
    })

  const save = async () => {
    if (!form) return
    if (!form.variableName.trim()) {
      toast.error(t('datasources.console.vars.nameRequired'))
      return
    }
    setSaving(true)
    try {
      const payload = {
        variableName: form.variableName.trim(),
        variableDescription: form.variableDescription.trim() || undefined,
        variableValue: form.variableValue, // blank keeps stored secret (mask-preserving)
        isSecret: form.isSecret,
      }
      if (form.id == null) await soarVariablesService.create(payload)
      else await soarVariablesService.update({ id: form.id, ...payload })
      toast.success(t('datasources.console.vars.saved'))
      setForm(null)
      load()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('datasources.console.vars.saveError'))
    } finally {
      setSaving(false)
    }
  }

  const remove = async (v: SoarVariable) => {
    try {
      await soarVariablesService.remove(v.id)
      toast.success(t('datasources.console.vars.deleted'))
      load()
    } catch {
      toast.error(t('datasources.console.vars.deleteError'))
    }
  }

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/50 p-4" onClick={onClose}>
      <div
        className="flex max-h-[80vh] w-full max-w-[640px] flex-col overflow-hidden rounded-xl border border-border bg-card shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between border-b border-border px-5 py-3">
          <div>
            <h2 className="flex items-center gap-2 text-sm font-semibold">
              <KeyRound size={15} /> {t('datasources.console.vars.title')}
            </h2>
            <p className="mt-0.5 text-[11px] text-muted-foreground">
              {t('datasources.console.vars.subtitle')} <code className="rounded bg-muted px-1 font-mono">$[variables.NAME]</code>
            </p>
          </div>
          <button onClick={onClose} className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground">
            <X size={16} />
          </button>
        </header>

        <div className="flex-1 overflow-y-auto p-5">
          {form ? (
            <div className="space-y-3 rounded-lg border border-border bg-muted/20 p-4">
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Field label={t('datasources.console.vars.name')}>
                  <Input value={form.variableName} onChange={(e) => setForm({ ...form, variableName: e.target.value })} placeholder="my_token" />
                </Field>
                <Field label={t('datasources.console.vars.value')} hint={form.id != null && form.isSecret ? t('datasources.console.vars.secretHint') : undefined}>
                  <Input
                    type={form.isSecret ? 'password' : 'text'}
                    value={form.variableValue}
                    onChange={(e) => setForm({ ...form, variableValue: e.target.value })}
                    placeholder={form.id != null && form.isSecret ? '••••••••' : ''}
                    autoComplete="new-password"
                  />
                </Field>
              </div>
              <Field label={t('datasources.console.vars.description')}>
                <Input value={form.variableDescription} onChange={(e) => setForm({ ...form, variableDescription: e.target.value })} />
              </Field>
              <label className="flex items-center gap-2 text-xs">
                <input type="checkbox" checked={form.isSecret} onChange={(e) => setForm({ ...form, isSecret: e.target.checked })} className="h-3.5 w-3.5" />
                {t('datasources.console.vars.isSecret')}
              </label>
              <div className="flex justify-end gap-2 pt-1">
                <Button size="sm" variant="outline" onClick={() => setForm(null)} disabled={saving}>
                  {t('datasources.console.vars.cancel')}
                </Button>
                <Button size="sm" onClick={() => void save()} disabled={saving}>
                  {saving ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : null}
                  {t('datasources.console.vars.save')}
                </Button>
              </div>
            </div>
          ) : (
            <Button size="sm" variant="outline" onClick={() => setForm(EMPTY)}>
              <Plus size={13} className="mr-1.5" /> {t('datasources.console.vars.add')}
            </Button>
          )}

          <div className="mt-4">
            {loading ? (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 size={14} className="animate-spin" /> {t('datasources.console.vars.loading')}
              </div>
            ) : items.length === 0 ? (
              <p className="py-6 text-center text-sm text-muted-foreground">{t('datasources.console.vars.empty')}</p>
            ) : (
              <div className="divide-y divide-border rounded-lg border border-border">
                {items.map((v) => (
                  <div key={v.id} className="flex items-center gap-3 px-3 py-2.5">
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <span className="truncate font-mono text-xs font-medium">{v.variableName}</span>
                        {v.isSecret && (
                          <span className="rounded bg-amber-500/15 px-1 py-0.5 text-[9px] font-semibold uppercase text-amber-600 dark:text-amber-400">
                            {t('datasources.console.vars.secret')}
                          </span>
                        )}
                      </div>
                      <div className="mt-0.5 truncate text-[11px] text-muted-foreground">
                        {v.isSecret ? '••••••••' : v.variableValue || '—'}
                        {v.variableDescription && <span className="ml-2 opacity-70">· {v.variableDescription}</span>}
                      </div>
                    </div>
                    {onInsert && v.variableName && (
                      <button
                        onClick={() => onInsert(v.variableName!)}
                        title={t('datasources.console.vars.insert')}
                        className="flex h-7 items-center gap-1 rounded border border-primary/40 bg-primary/5 px-2 text-[11px] font-medium text-primary hover:bg-primary/10"
                      >
                        <CornerDownLeft size={12} /> {t('datasources.console.vars.insert')}
                      </button>
                    )}
                    <button onClick={() => startEdit(v)} title={t('datasources.console.vars.edit')} className="flex h-7 w-7 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground">
                      <Pencil size={13} />
                    </button>
                    <button onClick={() => void remove(v)} title={t('datasources.console.vars.delete')} className="flex h-7 w-7 items-center justify-center rounded text-muted-foreground hover:bg-red-500/10 hover:text-red-500">
                      <Trash2 size={13} />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1">
      <label className="block text-xs font-medium text-foreground/80">{label}</label>
      {children}
      {hint && <p className={cn('text-[10px] text-muted-foreground')}>{hint}</p>}
    </div>
  )
}
