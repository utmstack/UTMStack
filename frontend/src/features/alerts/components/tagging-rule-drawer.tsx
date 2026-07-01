import { useMemo, useState } from 'react'
import { Loader2, Pencil, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { useDateFormat } from '@/shared/lib/datetime'
import { operatorById } from '../lib/tagging-rule-meta'
import {
  deserializeConditions,
  ruleToForm,
  serializeConditions,
  type FormState,
} from '../lib/tagging-rule-form'
import type { AlertTag, FilterType, TaggingRule } from '../types/tagging-rule.types'
import { TaggingRuleView } from './tagging-rule-view'
import { TaggingRuleForm } from './tagging-rule-form'

export function TaggingRuleDrawer({
  rule,
  create,
  initialTags,
  initialConditions,
  startInEdit,
  tagCatalog,
  onClose,
  onSubmit,
  onDelete,
  onCreateTag,
}: {
  /** Existing rule when editing/viewing. */
  rule?: TaggingRule | null
  /** Create mode (mutually exclusive with `rule`). */
  create?: boolean
  /** Pre-select these tags when creating a fresh rule (e.g. arriving from the
   *  tag editor's "create rule with this tag" deep-link). Ignored when editing. */
  initialTags?: AlertTag[]
  /** Pre-fill conditions when creating a fresh rule (e.g. arriving from the
   *  alerts list's "create rule from alert" deep-link). Ignored when editing. */
  initialConditions?: FilterType[]
  /** Open an existing rule directly in edit mode (no view-then-edit step). */
  startInEdit?: boolean
  tagCatalog: AlertTag[]
  onClose: () => void
  onSubmit: (input: FormState, id?: number) => Promise<unknown>
  onDelete?: (rule: TaggingRule) => Promise<unknown>
  onCreateTag: (tagName: string, tagColor: string) => Promise<AlertTag | null>
}) {
  const { t } = useTranslation()
  const df = useDateFormat()
  const [editing, setEditing] = useState(!!create || !!startInEdit)
  const [form, setForm] = useState<FormState>(() => ({
    ...ruleToForm(rule, initialTags, initialConditions),
    conditions: deserializeConditions(ruleToForm(rule, initialTags, initialConditions).conditions),
  }))
  const [busy, setBusy] = useState(false)

  const valid = useMemo(() => {
    if (!form.name.trim()) return false
    if (!form.description.trim()) return false
    if (form.tags.length === 0) return false
    if (form.conditions.length === 0) return false
    return form.conditions.every((c) => {
      const op = operatorById(c.operator)
      if (!op) return false
      if (op.needs === 'none') return !!c.field
      const v = typeof c.value === 'string' ? c.value.trim() : ''
      return !!c.field && v.length > 0
    })
  }, [form])

  const save = async () => {
    if (!valid || busy) {
      if (!valid) toast.error(t('taggingRules.form.incomplete'))
      return
    }
    setBusy(true)
    try {
      await onSubmit(
        { ...form, conditions: serializeConditions(form.conditions) },
        create ? undefined : rule?.id
      )
    } finally {
      setBusy(false)
    }
  }

  const cancelEdit = () => {
    setEditing(false)
    const fresh = ruleToForm(rule, initialTags, initialConditions)
    setForm({ ...fresh, conditions: deserializeConditions(fresh.conditions) })
  }

  const showForm = editing || !!create

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[720px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div className="min-w-0 flex-1">
            <div className="text-[11px] text-muted-foreground">{t('taggingRules.title')}</div>
            <h2 className="mt-1 truncate text-xl font-semibold">
              {create ? t('taggingRules.drawer.createTitle') : rule?.name}
            </h2>
          </div>
          <div className="flex shrink-0 items-center gap-2">
            {rule && !showForm && (
              <Button size="sm" variant="outline" onClick={() => setEditing(true)}>
                <Pencil size={13} className="mr-1.5" />
                {t('taggingRules.drawer.edit')}
              </Button>
            )}
            {rule && onDelete && !showForm && (
              <button
                onClick={() => void onDelete(rule)}
                title={t('taggingRules.drawer.delete')}
                className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
              >
                <Trash2 size={15} />
              </button>
            )}
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>
        </header>

        <div className="min-h-0 flex-1 overflow-y-auto bg-muted/10 p-6">
          {showForm ? (
            <TaggingRuleForm
              form={form}
              setForm={setForm}
              tagCatalog={tagCatalog}
              onCreateTag={onCreateTag}
              t={t}
            />
          ) : rule ? (
            <TaggingRuleView rule={rule} df={df} t={t} />
          ) : null}
        </div>

        {showForm && (
          <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
            {!create && (
              <Button size="sm" variant="outline" onClick={cancelEdit} disabled={busy}>
                {t('taggingRules.drawer.cancel')}
              </Button>
            )}
            <Button size="sm" onClick={() => void save()} disabled={busy || !valid}>
              {busy ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : null}
              {create ? t('taggingRules.drawer.create') : t('taggingRules.drawer.save')}
            </Button>
          </footer>
        )}
      </div>
    </div>
  )
}
