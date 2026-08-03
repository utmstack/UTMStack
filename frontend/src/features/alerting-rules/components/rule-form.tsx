import type { Dispatch, SetStateAction } from 'react'
import { useEffect, useRef, useState } from 'react'
import type { TFunction } from 'i18next'
import { AlertTriangle, Check, ChevronDown, Plus, Trash2, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/components/ui/input'
import { ChipInput } from '@/shared/components/ui/chip-input'
import type { CorrelationRule, DataTypeOption, DataTypeRef, SaveRuleInput } from '../services/alerting-rules-http.service'

/* ─── Structured form model ────────────────────────────────────────────── */

export const OPERATORS = ['filter_term', 'must_not_term', 'filter_match', 'must_not_match'] as const

export interface Condition {
  field: string
  operator: string
  value: string
}
export interface AfterStep {
  indexPattern: string
  within: string
  count: number
  with: Condition[]
  or: { with: Condition[] }[]
}

export interface RuleFormState {
  name: string
  adversary: string
  confidentiality: number
  integrity: number
  availability: number
  category: string
  technique: string
  description: string
  ruleActive: boolean
  dataTypes: string[]
  definition: string
  correlation: AfterStep[]
  groupBy: string[]
  deduplicateBy: string[]
  references: string[]
}

const SELECT_CLS = 'h-9 rounded-md border border-border bg-background px-2 text-sm'

/* ─── Parse (rule → form) / serialize (form → request) ─────────────────── */

function valueToStr(v: unknown): string {
  if (v == null) return ''
  return typeof v === 'string' ? v : JSON.stringify(v)
}
function strArray(v: unknown): string[] {
  return Array.isArray(v) ? v.map((x) => String(x)) : []
}
function parseConditions(w: unknown): Condition[] {
  if (!Array.isArray(w)) return []
  return w.map((c) => {
    const o = (c ?? {}) as Record<string, unknown>
    return { field: String(o.field ?? ''), operator: String(o.operator ?? 'filter_term'), value: valueToStr(o.value) }
  })
}
function parseSteps(raw: unknown): AfterStep[] {
  if (!Array.isArray(raw)) return []
  return raw.map((s) => {
    const o = (s ?? {}) as Record<string, unknown>
    return {
      indexPattern: String(o.indexPattern ?? ''),
      within: String(o.within ?? ''),
      count: Number(o.count ?? 0),
      with: parseConditions(o.with),
      or: Array.isArray(o.or) ? (o.or as unknown[]).map((g) => ({ with: parseConditions((g as Record<string, unknown>)?.with) })) : [],
    }
  })
}

export function ruleToForm(r?: CorrelationRule): RuleFormState {
  return {
    name: r?.name ?? '',
    adversary: r?.adversary ?? 'origin',
    confidentiality: r?.confidentiality ?? 0,
    integrity: r?.integrity ?? 0,
    availability: r?.availability ?? 0,
    category: r?.category ?? '',
    technique: r?.technique ?? '',
    description: r?.description ?? '',
    ruleActive: r?.ruleActive ?? true,
    dataTypes: (r?.dataTypes ?? []).filter((d) => d.included).map((d) => d.dataType),
    definition: r?.definition ?? '',
    correlation: parseSteps(r?.correlation ?? r?.afterEvents),
    groupBy: strArray(r?.groupBy),
    deduplicateBy: strArray(r?.deduplicateBy),
    references: strArray(r?.references),
  }
}

function condToJson(cs: Condition[]): Record<string, unknown>[] {
  return cs.filter((c) => c.field.trim()).map((c) => ({ field: c.field.trim(), operator: c.operator, value: c.value }))
}
function stepsToJson(steps: AfterStep[]): Record<string, unknown>[] {
  return steps
    .filter((s) => s.indexPattern.trim())
    .map((s) => {
      const count = Number(s.count) || 0
      const base: Record<string, unknown> = { indexPattern: s.indexPattern.trim(), within: s.within.trim(), count, with: condToJson(s.with) }
      const ors = s.or.filter((g) => g.with.some((c) => c.field.trim()))
      if (ors.length) base.or = ors.map((g) => ({ indexPattern: s.indexPattern.trim(), within: s.within.trim(), count, with: condToJson(g.with) }))
      return base
    })
}

export function formToInput(f: RuleFormState, relPath?: string): SaveRuleInput {
  const dataTypes: DataTypeRef[] = f.dataTypes.map((dt) => ({ dataType: dt, included: true }))
  return {
    relPath,
    name: f.name.trim(),
    adversary: f.adversary,
    confidentiality: f.confidentiality,
    integrity: f.integrity,
    availability: f.availability,
    category: f.category.trim(),
    technique: f.technique.trim(),
    description: f.description.trim(),
    references: f.references.filter(Boolean),
    definition: f.definition,
    correlation: stepsToJson(f.correlation),
    groupBy: f.groupBy.filter(Boolean),
    deduplicateBy: f.deduplicateBy.filter(Boolean),
    ruleActive: f.ruleActive,
    dataTypes,
  }
}

/* ─── Form ─────────────────────────────────────────────────────────────── */

export function RuleForm({ form, setForm, dataTypeOptions, t }: { form: RuleFormState; setForm: Dispatch<SetStateAction<RuleFormState>>; dataTypeOptions: DataTypeOption[]; t: TFunction }) {
  const set = <K extends keyof RuleFormState>(k: K, v: RuleFormState[K]) => setForm((f) => ({ ...f, [k]: v }))
  const setSteps = (steps: AfterStep[]) => set('correlation', steps)

  // `where` condition: visual builder ⇄ raw CEL. Visual is the source when it
  // parses; otherwise we fall back to code for hand-written/advanced CEL.
  const [mode, setMode] = useState<'visual' | 'code'>(() => (whereToModel(form.definition) ? 'visual' : 'code'))
  const [whereModel, setWhereModel] = useState<WhereModel>(() => whereToModel(form.definition) ?? { match: 'all', conditions: [] })
  const [parseWarn, setParseWarn] = useState(false)
  const updateWhere = (m: WhereModel) => { setWhereModel(m); set('definition', modelToCel(m)) }
  const switchMode = (next: 'visual' | 'code') => {
    if (next === 'visual') {
      const m = whereToModel(form.definition)
      if (m) { setWhereModel(m); setParseWarn(false) } else setParseWarn(true)
    }
    setMode(next)
  }

  return (
    <div className="space-y-4">
      <Section title={t('alertingRules.editor.metadata')}>
        <div className="space-y-3">
          <Field label={t('alertingRules.editor.name')}><Input value={form.name} onChange={(e) => set('name', e.target.value)} className="h-8" /></Field>
          <div className="grid grid-cols-2 gap-3">
            <Field label={t('alertingRules.table.category')}><Input value={form.category} onChange={(e) => set('category', e.target.value)} className="h-8" /></Field>
            <Field label={t('alertingRules.table.technique')}><Input value={form.technique} onChange={(e) => set('technique', e.target.value)} className="h-8" /></Field>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <Field label={t('alertingRules.table.adversary')}>
              <select value={form.adversary} onChange={(e) => set('adversary', e.target.value)} className={cn(SELECT_CLS, 'h-8 w-full')}>
                <option value="origin">{t('alertingRules.adversary.origin')}</option>
                <option value="target">{t('alertingRules.adversary.target')}</option>
              </select>
            </Field>
            <Field label={t('alertingRules.editor.dataTypes')}>
              <DataTypeSelect values={form.dataTypes} options={dataTypeOptions} onChange={(v) => set('dataTypes', v)} t={t} />
            </Field>
          </div>
          <div className="grid grid-cols-3 gap-3">
            {(['confidentiality', 'integrity', 'availability'] as const).map((k) => (
              <Field key={k} label={t(`alertingRules.editor.${k}`)}>
                <select value={form[k]} onChange={(e) => set(k, Number(e.target.value))} className={cn(SELECT_CLS, 'h-8 w-full')}>
                  {[0, 1, 2, 3].map((n) => <option key={n} value={n}>{n}</option>)}
                </select>
              </Field>
            ))}
          </div>
          <Field label={t('alertingRules.view.description')}>
            <textarea value={form.description} onChange={(e) => set('description', e.target.value)} rows={2} className="w-full rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" />
          </Field>
          <label className="flex items-center gap-2 text-xs">
            <input type="checkbox" checked={form.ruleActive} onChange={(e) => set('ruleActive', e.target.checked)} /> {t('alertingRules.editor.activeOnSave')}
          </label>
        </div>
      </Section>

      <div className="rounded-lg border border-border bg-card p-4">
        <div className="mb-2 flex items-center justify-between">
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{t('alertingRules.editor.where')}</div>
          <div className="inline-flex rounded-md border border-border p-0.5">
            {(['visual', 'code'] as const).map((m) => (
              <button key={m} onClick={() => switchMode(m)} className={cn('rounded px-2.5 py-0.5 text-[11px] transition-colors', mode === m ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:text-foreground')}>
                {t(`alertingRules.editor.${m}`)}
              </button>
            ))}
          </div>
        </div>
        <p className="mb-2 text-[11px] text-muted-foreground">{t('alertingRules.editor.whereHint')}</p>
        {mode === 'visual' ? (
          <>
            {parseWarn && (
              <div className="mb-2 flex items-center gap-1.5 rounded-md border border-amber-500/30 bg-amber-500/10 px-2 py-1.5 text-[11px] text-amber-600 dark:text-amber-400">
                <AlertTriangle size={12} /> {t('alertingRules.where.advanced')}
              </div>
            )}
            <WhereBuilder model={whereModel} onChange={updateWhere} t={t} />
          </>
        ) : (
          <textarea
            value={form.definition}
            onChange={(e) => set('definition', e.target.value)}
            rows={4}
            spellCheck={false}
            placeholder={'equals("action", "denied") && exists("origin.ip")'}
            className="w-full rounded-md border border-input bg-background p-2.5 font-mono text-[11px] leading-relaxed focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          />
        )}
      </div>

      <Section title={t('alertingRules.editor.correlationSteps')}>
        <p className="mb-3 text-[11px] text-muted-foreground">{t('alertingRules.editor.correlationStepsHint')}</p>
        <div className="space-y-3">
          {form.correlation.map((step, i) => (
            <StepCard
              key={i}
              step={step}
              index={i}
              t={t}
              onChange={(s) => setSteps(form.correlation.map((x, idx) => (idx === i ? s : x)))}
              onRemove={() => setSteps(form.correlation.filter((_, idx) => idx !== i))}
            />
          ))}
          <button
            onClick={() => setSteps([...form.correlation, { indexPattern: 'v11-log-*', within: '2m', count: 1, with: [], or: [] }])}
            className="flex w-full items-center justify-center gap-1.5 rounded-md border border-dashed border-border py-2 text-xs text-muted-foreground hover:border-primary/40 hover:text-foreground"
          >
            <Plus size={13} /> {t('alertingRules.editor.addStep')}
          </button>
        </div>
      </Section>

      <Section title={t('alertingRules.view.correlation')}>
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
          <Field label={t('alertingRules.view.groupBy')} hint={t('alertingRules.editor.fieldsHint')}><ChipInput values={form.groupBy} onChange={(v) => set('groupBy', v)} placeholder="adversary.ip" mono /></Field>
          <Field label={t('alertingRules.view.deduplicateBy')} hint={t('alertingRules.editor.fieldsHint')}><ChipInput values={form.deduplicateBy} onChange={(v) => set('deduplicateBy', v)} placeholder="target.ip" mono /></Field>
        </div>
        <div className="mt-3">
          <Field label={t('alertingRules.view.references')} hint={t('alertingRules.editor.referencesHint')}><ChipInput values={form.references} onChange={(v) => set('references', v)} placeholder="https://attack.mitre.org/…" /></Field>
        </div>
      </Section>
    </div>
  )
}

/* ─── correlation steps builder pieces ────────────────────────────────── */

function StepCard({ step, index, onChange, onRemove, t }: { step: AfterStep; index: number; onChange: (s: AfterStep) => void; onRemove: () => void; t: TFunction }) {
  const patch = (p: Partial<AfterStep>) => onChange({ ...step, ...p })
  return (
    <div className="rounded-lg border border-border bg-background/40 p-3">
      <div className="mb-2 flex items-center justify-between">
        <span className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">{t('alertingRules.editor.step', { n: index + 1 })}</span>
        <button onClick={onRemove} title={t('alertingRules.editor.removeStep')} className="rounded p-1 text-muted-foreground hover:bg-red-500/10 hover:text-red-500"><Trash2 size={13} /></button>
      </div>
      <div className="grid grid-cols-[1fr_110px_80px] gap-2">
        <Field label={t('alertingRules.editor.indexPattern')}><Input value={step.indexPattern} onChange={(e) => patch({ indexPattern: e.target.value })} className="h-8 font-mono text-xs" /></Field>
        <Field label={t('alertingRules.editor.within')}><Input value={step.within} onChange={(e) => patch({ within: e.target.value })} className="h-8 font-mono text-xs" placeholder="2m" /></Field>
        <Field label={t('alertingRules.editor.count')}><Input type="number" value={step.count} onChange={(e) => patch({ count: Number(e.target.value) })} className="h-8 font-mono text-xs" /></Field>
      </div>

      <ConditionList label={t('alertingRules.editor.conditions')} conditions={step.with} onChange={(w) => patch({ with: w })} t={t} />

      {step.or.map((g, gi) => (
        <ConditionList
          key={gi}
          label={t('alertingRules.editor.orGroup', { n: gi + 1 })}
          conditions={g.with}
          onChange={(w) => patch({ or: step.or.map((x, idx) => (idx === gi ? { with: w } : x)) })}
          onRemoveGroup={() => patch({ or: step.or.filter((_, idx) => idx !== gi) })}
          t={t}
        />
      ))}
      <button onClick={() => patch({ or: [...step.or, { with: [] }] })} className="mt-2 text-[11px] text-primary hover:underline">+ {t('alertingRules.editor.addOr')}</button>
    </div>
  )
}

function ConditionList({ label, conditions, onChange, onRemoveGroup, t }: { label: string; conditions: Condition[]; onChange: (c: Condition[]) => void; onRemoveGroup?: () => void; t: TFunction }) {
  return (
    <div className="mt-3 rounded-md border border-border/60 p-2">
      <div className="mb-1.5 flex items-center justify-between">
        <span className="text-[10px] font-medium uppercase tracking-wider text-muted-foreground">{label}</span>
        {onRemoveGroup && <button onClick={onRemoveGroup} className="text-[10px] text-muted-foreground hover:text-red-500">{t('alertingRules.editor.removeOr')}</button>}
      </div>
      <div className="space-y-1.5">
        {conditions.map((c, i) => (
          <div key={i} className="flex items-center gap-1.5">
            <Input value={c.field} onChange={(e) => onChange(conditions.map((x, idx) => (idx === i ? { ...x, field: e.target.value } : x)))} placeholder={t('alertingRules.editor.field')} className="h-7 flex-1 font-mono text-xs" />
            <select value={c.operator} onChange={(e) => onChange(conditions.map((x, idx) => (idx === i ? { ...x, operator: e.target.value } : x)))} className="h-7 rounded-md border border-border bg-background px-1 text-[11px]">
              {OPERATORS.map((op) => <option key={op} value={op}>{t(`alertingRules.operator.${op}`)}</option>)}
            </select>
            <Input value={c.value} onChange={(e) => onChange(conditions.map((x, idx) => (idx === i ? { ...x, value: e.target.value } : x)))} placeholder={t('alertingRules.editor.value')} className="h-7 flex-1 font-mono text-xs" />
            <button onClick={() => onChange(conditions.filter((_, idx) => idx !== i))} className="shrink-0 rounded p-1 text-muted-foreground hover:text-red-500"><X size={12} /></button>
          </div>
        ))}
      </div>
      <button onClick={() => onChange([...conditions, { field: '', operator: 'filter_term', value: '' }])} className="mt-1.5 flex items-center gap-1 text-[11px] text-primary hover:underline">
        <Plus size={11} /> {t('alertingRules.editor.addCondition')}
      </button>
    </div>
  )
}

/* ─── Data type select (colored chips from the catalog) ────────────────── */

const DT_PALETTE = [
  'rgb(14 165 233)', 'rgb(99 102 241)', 'rgb(168 85 247)', 'rgb(236 72 153)', 'rgb(244 63 94)',
  'rgb(249 115 22)', 'rgb(234 179 8)', 'rgb(16 185 129)', 'rgb(20 184 166)', 'rgb(6 182 212)',
]
/** Stable color per data type (hash → palette) so the same type always looks the same. */
export function dtColor(dt: string): string {
  let h = 0
  for (let i = 0; i < dt.length; i++) h = (h * 31 + dt.charCodeAt(i)) >>> 0
  return DT_PALETTE[h % DT_PALETTE.length]
}

export function DataTypeChip({ dataType, onRemove }: { dataType: string; onRemove?: () => void }) {
  const c = dtColor(dataType)
  return (
    <span className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-[11px] font-medium" style={{ backgroundColor: `${c}22`, color: c, border: `1px solid ${c}55` }}>
      <span className="h-1.5 w-1.5 rounded-full" style={{ backgroundColor: c }} /> {dataType}
      {onRemove && <button onClick={onRemove} className="opacity-70 hover:opacity-100"><X size={10} /></button>}
    </span>
  )
}

function DataTypeSelect({ values, options, onChange, t }: { values: string[]; options: DataTypeOption[]; onChange: (v: string[]) => void; t: TFunction }) {
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const ref = useRef<HTMLDivElement>(null)
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  const toggle = (dt: string) => onChange(values.includes(dt) ? values.filter((x) => x !== dt) : [...values, dt])
  const filtered = options.filter((o) => (q ? (o.dataType + ' ' + o.name).toLowerCase().includes(q.toLowerCase()) : true))

  return (
    <div ref={ref} className="relative">
      <div className="flex min-h-8 flex-wrap items-center gap-1 rounded-md border border-input bg-background p-1">
        {values.map((v) => <DataTypeChip key={v} dataType={v} onRemove={() => toggle(v)} />)}
        <button onClick={() => setOpen((o) => !o)} className="inline-flex items-center gap-1 px-1 text-[11px] text-muted-foreground hover:text-foreground">
          <Plus size={11} /> {values.length ? '' : t('alertingRules.editor.dataTypesPick')} <ChevronDown size={10} className="opacity-60" />
        </button>
      </div>
      {open && (
        <div className="absolute left-0 top-full z-30 mt-1 w-64 rounded-md border border-border bg-popover py-1 shadow-lg">
          <div className="px-2 pb-1.5 pt-1">
            <input value={q} onChange={(e) => setQ(e.target.value)} autoFocus placeholder={t('alertingRules.editor.dataTypesSearch')} className="h-7 w-full rounded-md border border-input bg-background px-2 text-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" />
          </div>
          <div className="max-h-56 overflow-y-auto">
            {filtered.length === 0 && <div className="px-3 py-1.5 text-xs text-muted-foreground">{t('alertingRules.editor.dataTypesNone')}</div>}
            {filtered.map((o) => {
              const on = values.includes(o.dataType)
              return (
                <button key={o.dataType} onClick={() => toggle(o.dataType)} className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted">
                  <span className="h-2 w-2 shrink-0 rounded-full" style={{ backgroundColor: dtColor(o.dataType) }} />
                  <span className="min-w-0 flex-1 truncate font-mono">{o.dataType}</span>
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

/* ─── CEL `where` visual builder ───────────────────────────────────────── */

// value kinds: none = no argument, auto = infer (number/bool/string),
// text = quoted string, number = numeric, list = comma → CEL list, two = two times.
type ValueKind = 'none' | 'auto' | 'text' | 'number' | 'list' | 'two'
export const WHERE_OPS: { fn: string; value: ValueKind }[] = [
  { fn: 'exists', value: 'none' },
  { fn: 'equals', value: 'auto' },
  { fn: 'equalsIgnoreCase', value: 'text' },
  { fn: 'contains', value: 'text' },
  { fn: 'oneOf', value: 'list' },
  { fn: 'containsAll', value: 'list' },
  { fn: 'startsWith', value: 'text' },
  { fn: 'endsWith', value: 'text' },
  { fn: 'regexMatch', value: 'text' },
  { fn: 'lessThan', value: 'number' },
  { fn: 'lessOrEqual', value: 'number' },
  { fn: 'greaterThan', value: 'number' },
  { fn: 'greaterOrEqual', value: 'number' },
  { fn: 'inCIDR', value: 'text' },
  { fn: 'isHour', value: 'number' },
  { fn: 'isMinute', value: 'number' },
  { fn: 'isDayOfWeek', value: 'number' },
  { fn: 'isWeekend', value: 'none' },
  { fn: 'isWorkDay', value: 'none' },
  { fn: 'isBetweenTime', value: 'two' },
]
const OP_KIND = (fn: string): ValueKind => WHERE_OPS.find((o) => o.fn === fn)?.value ?? 'text'

export interface WhereCond { field: string; fn: string; value: string; value2: string; negate: boolean }
export interface WhereModel { match: 'all' | 'any'; conditions: WhereCond[] }

const q = (s: string) => `"${s.replace(/\\/g, '\\\\').replace(/"/g, '\\"')}"`
function genArg(value: string, kind: ValueKind): string {
  if (kind === 'number') return /^-?\d+(\.\d+)?$/.test(value.trim()) ? value.trim() : q(value)
  if (kind === 'auto') {
    const v = value.trim()
    if (/^-?\d+(\.\d+)?$/.test(v)) return v
    if (v === 'true' || v === 'false') return v
    return q(value)
  }
  return q(value)
}
function genCond(c: WhereCond): string {
  const kind = OP_KIND(c.fn)
  let args = q(c.field)
  if (kind === 'two') args += `, ${q(c.value)}, ${q(c.value2)}`
  else if (kind === 'list') args += `, [${c.value.split(',').map((s) => q(s.trim())).filter((s) => s !== '""').join(', ')}]`
  else if (kind !== 'none') args += `, ${genArg(c.value, kind)}`
  const expr = `${c.fn}(${args})`
  return c.negate ? `!${expr}` : expr
}
export function modelToCel(m: WhereModel): string {
  const valid = m.conditions.filter((c) => c.field.trim() && c.fn)
  if (!valid.length) return ''
  return valid.map(genCond).join(m.match === 'all' ? ' &&\n' : ' ||\n')
}

const TERM_RE = /^(!)?\s*([a-zA-Z]+)\(\s*"((?:[^"\\]|\\.)*)"\s*(?:,\s*(.+))?\)$/
function unq(s: string): string {
  const m = s.trim().match(/^"((?:[^"\\]|\\.)*)"$/)
  return m ? m[1].replace(/\\"/g, '"').replace(/\\\\/g, '\\') : s.trim()
}
/** Best-effort CEL → model. Returns null when the expression isn't the simple
 *  function-call grammar the builder produces (parens, mixed &&/||, etc.). */
export function whereToModel(cel: string): WhereModel | null {
  const src = cel.trim()
  if (!src) return { match: 'all', conditions: [] }
  if (src.includes('(') === false) return null
  // No support for grouping parens around sub-expressions.
  const hasAnd = /\)\s*&&/.test(src)
  const hasOr = /\)\s*\|\|/.test(src)
  if (hasAnd && hasOr) return null
  const match: 'all' | 'any' = hasOr ? 'any' : 'all'
  const parts = splitTop(src, match === 'any' ? '||' : '&&')
  const conditions: WhereCond[] = []
  for (const raw of parts) {
    const m = raw.trim().match(TERM_RE)
    if (!m) return null
    const [, neg, fn, field, rest] = m
    const op = WHERE_OPS.find((o) => o.fn === fn)
    if (!op) return null
    const cond: WhereCond = { field, fn, value: '', value2: '', negate: !!neg }
    if (op.value === 'two') {
      const args = splitTop(rest ?? '', ',')
      cond.value = unq(args[0] ?? '')
      cond.value2 = unq(args[1] ?? '')
    } else if (op.value === 'list') {
      const inner = (rest ?? '').trim().replace(/^\[/, '').replace(/\]$/, '')
      cond.value = splitTop(inner, ',').map((x) => unq(x)).join(', ')
    } else if (op.value !== 'none') {
      cond.value = unq(rest ?? '')
    }
    conditions.push(cond)
  }
  return { match, conditions }
}
// split on a top-level operator (ignores commas/operators inside quotes/brackets).
function splitTop(s: string, op: string): string[] {
  const out: string[] = []
  let depth = 0
  let inStr = false
  let cur = ''
  for (let i = 0; i < s.length; i++) {
    const ch = s[i]
    if (ch === '"' && s[i - 1] !== '\\') inStr = !inStr
    if (!inStr) {
      if (ch === '[' || ch === '(') depth++
      else if (ch === ']' || ch === ')') depth--
      else if (depth === 0 && s.startsWith(op, i)) { out.push(cur); cur = ''; i += op.length - 1; continue }
    }
    cur += ch
  }
  out.push(cur)
  return out.map((x) => x.trim()).filter(Boolean)
}

export function WhereBuilder({ model, onChange, t }: { model: WhereModel; onChange: (m: WhereModel) => void; t: TFunction }) {
  const setCond = (i: number, c: WhereCond) => onChange({ ...model, conditions: model.conditions.map((x, idx) => (idx === i ? c : x)) })
  return (
    <div>
      <div className="mb-2 flex items-center gap-2 text-xs text-muted-foreground">
        {t('alertingRules.where.matchPrefix')}
        <select value={model.match} onChange={(e) => onChange({ ...model, match: e.target.value as 'all' | 'any' })} className="h-7 rounded-md border border-border bg-background px-1.5 text-xs">
          <option value="all">{t('alertingRules.where.all')}</option>
          <option value="any">{t('alertingRules.where.any')}</option>
        </select>
        {t('alertingRules.where.matchSuffix')}
      </div>
      <div className="space-y-1.5">
        {model.conditions.map((c, i) => {
          const kind = OP_KIND(c.fn)
          return (
            <div key={i} className="flex flex-wrap items-center gap-1.5">
              <Input value={c.field} onChange={(e) => setCond(i, { ...c, field: e.target.value })} placeholder={t('alertingRules.editor.field')} className="h-7 min-w-[130px] flex-1 font-mono text-xs" />
              <button onClick={() => setCond(i, { ...c, negate: !c.negate })} title={t('alertingRules.where.negate')} className={cn('h-7 rounded-md border px-2 text-xs', c.negate ? 'border-red-500/40 bg-red-500/10 text-red-500' : 'border-border text-muted-foreground')}>not</button>
              <select value={c.fn} onChange={(e) => setCond(i, { ...c, fn: e.target.value })} className="h-7 rounded-md border border-border bg-background px-1 text-[11px]">
                {WHERE_OPS.map((o) => <option key={o.fn} value={o.fn}>{t(`alertingRules.celOp.${o.fn}`)}</option>)}
              </select>
              {kind !== 'none' && (
                <Input value={c.value} onChange={(e) => setCond(i, { ...c, value: e.target.value })} placeholder={kind === 'list' ? t('alertingRules.where.listHint') : t('alertingRules.editor.value')} className="h-7 min-w-[110px] flex-1 font-mono text-xs" />
              )}
              {kind === 'two' && (
                <Input value={c.value2} onChange={(e) => setCond(i, { ...c, value2: e.target.value })} placeholder={t('alertingRules.editor.value')} className="h-7 min-w-[90px] font-mono text-xs" />
              )}
              <button onClick={() => onChange({ ...model, conditions: model.conditions.filter((_, idx) => idx !== i) })} className="shrink-0 rounded p-1 text-muted-foreground hover:text-red-500"><X size={12} /></button>
            </div>
          )
        })}
      </div>
      <button onClick={() => onChange({ ...model, conditions: [...model.conditions, { field: '', fn: 'equals', value: '', value2: '', negate: false }] })} className="mt-1.5 flex items-center gap-1 text-[11px] text-primary hover:underline">
        <Plus size={11} /> {t('alertingRules.editor.addCondition')}
      </button>
    </div>
  )
}

/* ─── Layout helpers ───────────────────────────────────────────────────── */

function Field({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <div className="mb-1 flex items-baseline justify-between gap-2">
        <span className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">{label}</span>
        {hint && <span className="text-[10px] text-muted-foreground/70">{hint}</span>}
      </div>
      {children}
    </label>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">{title}</div>
      {children}
    </div>
  )
}
