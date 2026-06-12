import { useEffect, useMemo, useState } from 'react'
import { ChevronDown, Loader2, Pencil, Plus, Tag as TagIcon, Trash2, X } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useDateFormat } from '@/shared/lib/datetime'
import { TAG_COLORS } from '../lib/alert-meta'
import {
  RULE_FIELDS,
  RULE_OPERATORS,
  SELECT_CLS,
  operatorById,
  type TaggingOperator,
} from '../lib/tagging-rule-meta'
import type { AlertTag, FilterType, TaggingRule } from '../types/tagging-rule.types'

interface FormState {
  name: string
  description: string
  conditions: FilterType[]
  tags: AlertTag[]
}

function ruleToForm(
  rule?: TaggingRule | null,
  initialTags?: AlertTag[],
  initialConditions?: FilterType[]
): FormState {
  const conds = rule?.conditions?.length
    ? rule.conditions.map((c) => ({ ...c }))
    : initialConditions?.length
      ? initialConditions.map((c) => ({ ...c }))
      : [{ field: RULE_FIELDS[0].field, operator: 'IS', value: '' }]
  return {
    name: rule?.name ?? '',
    description: rule?.description ?? '',
    conditions: conds,
    tags: rule?.tags ? [...rule.tags] : initialTags ? [...initialTags] : [],
  }
}

/** Conditions are stored as FilterType. For multi-value operators (IS_ONE_OF
 * etc.) the backend accepts an array; we expose a comma-separated input and
 * convert at the boundary. */
function serializeConditions(conds: FilterType[]): FilterType[] {
  return conds.map((c) => {
    const op = operatorById(c.operator)
    if (!op) return c
    if (op.needs === 'none') return { field: c.field, operator: c.operator }
    if (op.needs === 'list') {
      const raw = typeof c.value === 'string' ? c.value : Array.isArray(c.value) ? c.value.join(',') : ''
      const arr = raw
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean)
      return { field: c.field, operator: c.operator, value: arr }
    }
    return { field: c.field, operator: c.operator, value: c.value ?? '' }
  })
}

function deserializeConditions(conds: FilterType[]): FilterType[] {
  return conds.map((c) => {
    const op = operatorById(c.operator)
    if (op?.needs === 'list' && Array.isArray(c.value)) {
      return { ...c, value: (c.value as unknown[]).join(', ') }
    }
    return { ...c }
  })
}

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
   *  alerts list's "create rule from this alert" deep-link). Ignored when editing. */
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
            <RuleForm
              form={form}
              setForm={setForm}
              tagCatalog={tagCatalog}
              onCreateTag={onCreateTag}
              t={t}
            />
          ) : rule ? (
            <RuleView rule={rule} df={df} t={t} />
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

/* ─── Read-only view ─────────────────────────────────────────────────────── */

function RuleView({
  rule,
  df,
  t,
}: {
  rule: TaggingRule
  df: ReturnType<typeof useDateFormat>
  t: TFunction
}) {
  return (
    <div className="space-y-4">
      {rule.description && (
        <Section title={t('taggingRules.view.description')}>
          <p className="text-xs leading-relaxed text-muted-foreground">{rule.description}</p>
        </Section>
      )}

      <Section title={t('taggingRules.view.tags')}>
        <div className="flex flex-wrap gap-1.5">
          {rule.tags.length === 0 ? (
            <span className="text-xs text-muted-foreground">—</span>
          ) : (
            rule.tags.map((tg) => (
              <span
                key={tg.id}
                className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px]"
                style={{
                  backgroundColor: (tg.tagColor || '#64748b') + '22',
                  color: tg.tagColor || '#64748b',
                }}
              >
                <TagIcon size={10} /> {tg.tagName}
              </span>
            ))
          )}
        </div>
      </Section>

      <Section title={t('taggingRules.view.conditions')}>
        {rule.conditions.length === 0 ? (
          <span className="text-xs text-muted-foreground">—</span>
        ) : (
          <ul className="space-y-1.5">
            {rule.conditions.map((c, i) => (
              <li key={i} className="flex flex-wrap items-center gap-1.5 text-xs">
                <span className="rounded bg-background px-1.5 py-0.5 font-mono text-[11px]">{c.field}</span>
                <span className="text-[11px] text-muted-foreground">
                  {operatorById(c.operator)?.label || c.operator}
                </span>
                {c.value !== undefined && c.value !== '' && (
                  <span className="rounded bg-primary/10 px-1.5 py-0.5 font-mono text-[11px] text-primary">
                    {Array.isArray(c.value) ? (c.value as unknown[]).join(', ') : String(c.value)}
                  </span>
                )}
              </li>
            ))}
          </ul>
        )}
      </Section>

      <Section title={t('taggingRules.view.audit')}>
        <dl className="grid grid-cols-[150px_1fr] gap-y-2 text-xs">
          <Row k={t('taggingRules.view.createdBy')}>{rule.createdBy || '—'}</Row>
          <Row k={t('taggingRules.view.createdDate')}>
            {rule.createdDate ? df.formatDateTime(rule.createdDate) : '—'}
          </Row>
          {rule.lastModifiedBy && <Row k={t('taggingRules.view.modifiedBy')}>{rule.lastModifiedBy}</Row>}
          {rule.lastModifiedDate && (
            <Row k={t('taggingRules.view.modifiedDate')}>{df.formatDateTime(rule.lastModifiedDate)}</Row>
          )}
        </dl>
      </Section>
    </div>
  )
}

/* ─── Form ───────────────────────────────────────────────────────────────── */

function RuleForm({
  form,
  setForm,
  tagCatalog,
  onCreateTag,
  t,
}: {
  form: FormState
  setForm: (f: FormState) => void
  tagCatalog: AlertTag[]
  onCreateTag: (tagName: string, tagColor: string) => Promise<AlertTag | null>
  t: TFunction
}) {
  const addCondition = () =>
    setForm({
      ...form,
      conditions: [...form.conditions, { field: RULE_FIELDS[0].field, operator: 'IS', value: '' }],
    })

  const updateCondition = (i: number, patch: Partial<FilterType>) => {
    const next = form.conditions.map((c, idx) => {
      if (idx !== i) return c
      const merged = { ...c, ...patch }
      // Re-default value when switching to a value-less op so we don't send stale text.
      if (patch.operator) {
        const nextOp = operatorById(patch.operator)
        if (nextOp?.needs === 'none') return { field: merged.field, operator: merged.operator }
        if (!('value' in patch)) return { ...merged, value: merged.value ?? '' }
      }
      return merged
    })
    setForm({ ...form, conditions: next })
  }

  const removeCondition = (i: number) =>
    setForm({ ...form, conditions: form.conditions.filter((_, idx) => idx !== i) })

  return (
    <div className="space-y-4">
      <Section title={t('taggingRules.form.basics')}>
        <div className="space-y-2">
          <label className="block text-xs">
            <span className="mb-1 block text-muted-foreground">{t('taggingRules.form.name')}</span>
            <Input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder={t('taggingRules.form.namePlaceholder')}
              className="h-9"
            />
          </label>
          <label className="block text-xs">
            <span className="mb-1 block text-muted-foreground">{t('taggingRules.form.description')}</span>
            <textarea
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              placeholder={t('taggingRules.form.descriptionPlaceholder')}
              maxLength={512}
              rows={3}
              className="w-full resize-none rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </label>
        </div>
      </Section>

      <Section title={t('taggingRules.form.tags')}>
        <TagPicker
          selected={form.tags}
          catalog={tagCatalog}
          onChange={(tags) => setForm({ ...form, tags })}
          onCreateTag={onCreateTag}
          t={t}
        />
      </Section>

      <Section
        title={t('taggingRules.form.conditions')}
        right={
          <button
            type="button"
            onClick={addCondition}
            className="inline-flex items-center gap-1 rounded-md border border-input bg-background px-2 py-1 text-xs hover:bg-muted"
          >
            <Plus size={12} /> {t('taggingRules.form.addCondition')}
          </button>
        }
      >
        <div className="space-y-2">
          {form.conditions.map((c, i) => (
            <ConditionRow
              key={i}
              cond={c}
              onChange={(patch) => updateCondition(i, patch)}
              onRemove={() => removeCondition(i)}
              t={t}
            />
          ))}
          {form.conditions.length === 0 && (
            <div className="text-xs text-muted-foreground">{t('taggingRules.form.noConditions')}</div>
          )}
        </div>
      </Section>
    </div>
  )
}

function ConditionRow({
  cond,
  onChange,
  onRemove,
  t,
}: {
  cond: FilterType
  onChange: (patch: Partial<FilterType>) => void
  onRemove: () => void
  t: TFunction
}) {
  const op: TaggingOperator = operatorById(cond.operator) ?? RULE_OPERATORS[0]
  const value = typeof cond.value === 'string' ? cond.value : Array.isArray(cond.value) ? (cond.value as unknown[]).join(', ') : ''
  return (
    <div className="flex flex-wrap items-center gap-1.5 rounded-md border border-input bg-background p-2">
      <select
        value={cond.field}
        onChange={(e) => onChange({ field: e.target.value })}
        className={cn(SELECT_CLS, 'min-w-[170px] flex-1')}
      >
        {RULE_FIELDS.map((f) => (
          <option key={f.field} value={f.field}>
            {f.label}
          </option>
        ))}
      </select>
      <select
        value={cond.operator}
        onChange={(e) => onChange({ operator: e.target.value, value: '' })}
        className={cn(SELECT_CLS, 'min-w-[150px]')}
      >
        {RULE_OPERATORS.map((o) => (
          <option key={o.id} value={o.id}>
            {o.label}
          </option>
        ))}
      </select>
      {op.needs !== 'none' && (
        <Input
          value={value}
          onChange={(e) => onChange({ value: e.target.value })}
          placeholder={
            op.needs === 'list' ? t('taggingRules.form.valueListPlaceholder') : t('taggingRules.form.valuePlaceholder')
          }
          className="h-9 min-w-[160px] flex-1"
        />
      )}
      <button
        type="button"
        onClick={onRemove}
        title={t('taggingRules.form.removeCondition')}
        className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
      >
        <Trash2 size={14} />
      </button>
    </div>
  )
}

/* ─── Tag picker (multi-select, inline create) ──────────────────────────── */

function TagPicker({
  selected,
  catalog,
  onChange,
  onCreateTag,
  t,
}: {
  selected: AlertTag[]
  catalog: AlertTag[]
  onChange: (tags: AlertTag[]) => void
  onCreateTag: (tagName: string, tagColor: string) => Promise<AlertTag | null>
  t: TFunction
}) {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [color, setColor] = useState(TAG_COLORS[5])
  const selectedIds = new Set(selected.map((s) => s.id))

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      const el = e.target as HTMLElement
      if (!el.closest('[data-tagging-rule-tag-picker]')) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const toggle = (tag: AlertTag) => {
    if (selectedIds.has(tag.id)) onChange(selected.filter((s) => s.id !== tag.id))
    else onChange([...selected, tag])
  }

  const create = async () => {
    const n = name.trim()
    if (!n) return
    const tag = await onCreateTag(n, color)
    if (tag) {
      setName('')
      onChange([...selected, tag])
    }
  }

  return (
    <div data-tagging-rule-tag-picker className="relative">
      <div className="flex flex-wrap items-center gap-1.5">
        {selected.length === 0 && <span className="text-xs text-muted-foreground">{t('taggingRules.form.noTags')}</span>}
        {selected.map((tg) => (
          <span
            key={tg.id}
            className="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px]"
            style={{
              backgroundColor: (tg.tagColor || '#64748b') + '22',
              color: tg.tagColor || '#64748b',
            }}
          >
            <TagIcon size={10} /> {tg.tagName}
            <button
              type="button"
              onClick={() => toggle(tg)}
              className="ml-0.5 rounded-full hover:bg-background/40"
              title={t('taggingRules.form.removeTag')}
            >
              <X size={10} />
            </button>
          </span>
        ))}
        <button
          type="button"
          onClick={() => setOpen((v) => !v)}
          className="inline-flex items-center gap-1 rounded-md border border-input bg-background px-2 py-1 text-xs hover:bg-muted"
        >
          <Plus size={12} /> {t('taggingRules.form.addTag')} <ChevronDown size={10} />
        </button>
      </div>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-72 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="max-h-48 overflow-y-auto">
            {catalog.length === 0 && (
              <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('taggingRules.form.noCatalog')}</div>
            )}
            {catalog.map((tg) => {
              const has = selectedIds.has(tg.id)
              return (
                <button
                  key={tg.id}
                  type="button"
                  onClick={() => toggle(tg)}
                  className={cn(
                    'flex w-full items-center gap-2 px-3 py-1.5 text-sm hover:bg-muted',
                    has && 'bg-muted/50'
                  )}
                >
                  <span
                    className="h-2.5 w-2.5 shrink-0 rounded-full"
                    style={{ backgroundColor: tg.tagColor || '#64748b' }}
                  />
                  <span className="truncate text-left">{tg.tagName}</span>
                  {has && <span className="ml-auto text-[10px] text-primary">✓</span>}
                </button>
              )
            })}
          </div>
          <div className="border-t border-border px-3 pb-2 pt-2">
            <div className="mb-1.5 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
              {t('taggingRules.form.createTag')}
            </div>
            <div className="flex items-center gap-1.5">
              <input
                value={name}
                onChange={(e) => setName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && void create()}
                placeholder={t('taggingRules.form.newTagName')}
                className="h-7 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
              <button
                type="button"
                onClick={() => void create()}
                disabled={!name.trim()}
                className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary text-primary-foreground disabled:opacity-40"
              >
                <Plus size={14} />
              </button>
            </div>
            <div className="mt-2 flex flex-wrap gap-1.5">
              {TAG_COLORS.map((c) => (
                <button
                  type="button"
                  key={c}
                  onClick={() => setColor(c)}
                  className={cn(
                    'h-5 w-5 rounded-full ring-offset-1 ring-offset-popover transition',
                    color === c && 'ring-2 ring-ring'
                  )}
                  style={{ backgroundColor: c }}
                />
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

/* ─── Tiny layout primitives ────────────────────────────────────────────── */

function Section({
  title,
  right,
  children,
}: {
  title: string
  right?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-2 flex items-center justify-between gap-2">
        <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{title}</div>
        {right}
      </div>
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
