import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { dump, load } from 'js-yaml'
import { Braces, ChevronDown, ChevronUp, Plus, Trash2, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import type { DataTypeOption } from '@/features/data-processing/types/data-processing.types'
import type { RegexPattern } from '@/features/regex-patterns/services/regex-patterns-http.service'
import { DataTypeMultiSelect } from './DataTypeMultiSelect'
import { PatternMenu } from './PatternMenu'
import {
  COMMON_KINDS,
  emptyBlock,
  emptyStep,
  STEP_KINDS,
  type FilterModel,
  type PipelineBlock,
  type Step,
  type StepKind,
} from '../lib/filter-model'

interface Props {
  value: FilterModel
  readOnly?: boolean
  dataTypeOptions: DataTypeOption[]
  patternOptions?: RegexPattern[]
  onPatternCreated?: (p: RegexPattern) => void
  onChange: (model: FilterModel) => void
}

export function VisualFilterEditor({
  value,
  readOnly,
  dataTypeOptions,
  patternOptions = [],
  onPatternCreated,
  onChange,
}: Props) {
  const { t } = useTranslation()

  const setBlock = (i: number, block: PipelineBlock) => {
    const pipeline = value.pipeline.slice()
    pipeline[i] = block
    onChange({ pipeline })
  }
  const removeBlock = (i: number) => onChange({ pipeline: value.pipeline.filter((_, x) => x !== i) })
  const addBlock = () => onChange({ pipeline: [...value.pipeline, emptyBlock()] })

  return (
    <div className="space-y-4">
      {value.pipeline.length === 0 && (
        <p className="rounded-md border border-dashed border-border px-4 py-6 text-center text-sm text-muted-foreground">
          {t('parsingFilters.visual.emptyPipeline')}
        </p>
      )}
      {value.pipeline.map((block, i) => (
        <BlockCard
          key={i}
          index={i}
          block={block}
          readOnly={readOnly}
          dataTypeOptions={dataTypeOptions}
          patternOptions={patternOptions}
          onPatternCreated={onPatternCreated}
          onChange={(b) => setBlock(i, b)}
          onRemove={() => removeBlock(i)}
        />
      ))}
      {!readOnly && (
        <Button type="button" variant="outline" size="sm" onClick={addBlock}>
          <Plus size={14} className="mr-1.5" /> {t('parsingFilters.visual.addBlock')}
        </Button>
      )}
    </div>
  )
}

function BlockCard({
  index,
  block,
  readOnly,
  dataTypeOptions,
  patternOptions,
  onPatternCreated,
  onChange,
  onRemove,
}: {
  index: number
  block: PipelineBlock
  readOnly?: boolean
  dataTypeOptions: DataTypeOption[]
  patternOptions: RegexPattern[]
  onPatternCreated?: (p: RegexPattern) => void
  onChange: (b: PipelineBlock) => void
  onRemove: () => void
}) {
  const { t } = useTranslation()

  const setSteps = (steps: Step[]) => onChange({ ...block, steps })
  const setStep = (i: number, s: Step) => setSteps(block.steps.map((x, k) => (k === i ? s : x)))
  const removeStep = (i: number) => setSteps(block.steps.filter((_, k) => k !== i))
  const addStep = (kind: StepKind) => setSteps([...block.steps, emptyStep(kind)])
  const moveStep = (i: number, dir: -1 | 1) => {
    const j = i + dir
    if (j < 0 || j >= block.steps.length) return
    const steps = block.steps.slice()
    ;[steps[i], steps[j]] = [steps[j], steps[i]]
    setSteps(steps)
  }

  return (
    <div className="rounded-xl border border-border bg-background/40">
      <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-2.5">
        <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {t('parsingFilters.visual.block')} {index + 1}
        </span>
        {!readOnly && (
          <button
            type="button"
            onClick={onRemove}
            className="rounded p-1 text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
            title={t('parsingFilters.visual.removeBlock')}
          >
            <Trash2 size={14} />
          </button>
        )}
      </div>

      <div className="space-y-4 p-4">
        <div className="space-y-1.5">
          <label className="block text-xs font-medium text-foreground/80">
            {t('parsingFilters.visual.dataTypes')}
          </label>
          <DataTypeMultiSelect
            values={block.dataTypes}
            options={dataTypeOptions}
            readOnly={readOnly}
            onChange={(dataTypes) => onChange({ ...block, dataTypes })}
          />
        </div>

        <div className="space-y-2">
          <label className="block text-xs font-medium text-foreground/80">
            {t('parsingFilters.visual.steps')}
          </label>
          {block.steps.length === 0 && (
            <p className="text-[11px] text-muted-foreground">{t('parsingFilters.visual.noSteps')}</p>
          )}
          {block.steps.map((step, i) => (
            <StepCard
              key={i}
              step={step}
              readOnly={readOnly}
              patternOptions={patternOptions}
              onPatternCreated={onPatternCreated}
              canUp={i > 0}
              canDown={i < block.steps.length - 1}
              onChange={(s) => setStep(i, s)}
              onRemove={() => removeStep(i)}
              onMove={(dir) => moveStep(i, dir)}
            />
          ))}
          {!readOnly && <AddStep onAdd={addStep} />}
        </div>
      </div>
    </div>
  )
}

function AddStep({ onAdd }: { onAdd: (kind: StepKind) => void }) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-wrap items-center gap-1.5 pt-1">
      <span className="text-[11px] text-muted-foreground">{t('parsingFilters.visual.addStep')}</span>
      {STEP_KINDS.map((kind) => (
        <button
          key={kind}
          type="button"
          onClick={() => onAdd(kind)}
          className={cn(
            'rounded border px-2 py-0.5 text-[11px] font-medium transition-colors',
            COMMON_KINDS.includes(kind)
              ? 'border-primary/30 bg-primary/5 text-primary hover:bg-primary/10'
              : 'border-border text-muted-foreground hover:bg-muted',
          )}
        >
          {kind}
        </button>
      ))}
    </div>
  )
}

function StepCard({
  step,
  readOnly,
  patternOptions,
  onPatternCreated,
  canUp,
  canDown,
  onChange,
  onRemove,
  onMove,
}: {
  step: Step
  readOnly?: boolean
  patternOptions: RegexPattern[]
  onPatternCreated?: (p: RegexPattern) => void
  canUp: boolean
  canDown: boolean
  onChange: (s: Step) => void
  onRemove: () => void
  onMove: (dir: -1 | 1) => void
}) {
  const { t } = useTranslation()
  const setConfig = (config: Record<string, unknown>) => onChange({ ...step, config })

  return (
    <div className="rounded-lg border border-border bg-card/60">
      <div className="flex items-center justify-between gap-2 border-b border-border px-3 py-1.5">
        <span className="inline-flex items-center gap-1.5">
          <span className="rounded bg-primary/10 px-1.5 py-0.5 font-mono text-[11px] font-semibold text-primary">
            {step.kind}
          </span>
        </span>
        {!readOnly && (
          <div className="flex items-center gap-0.5">
            <button
              type="button"
              disabled={!canUp}
              onClick={() => onMove(-1)}
              className="rounded p-1 text-muted-foreground hover:bg-muted disabled:opacity-30"
              title={t('parsingFilters.visual.moveUp')}
            >
              <ChevronUp size={13} />
            </button>
            <button
              type="button"
              disabled={!canDown}
              onClick={() => onMove(1)}
              className="rounded p-1 text-muted-foreground hover:bg-muted disabled:opacity-30"
              title={t('parsingFilters.visual.moveDown')}
            >
              <ChevronDown size={13} />
            </button>
            <button
              type="button"
              onClick={onRemove}
              className="rounded p-1 text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
              title={t('parsingFilters.visual.removeStep')}
            >
              <Trash2 size={13} />
            </button>
          </div>
        )}
      </div>
      <div className="space-y-3 p-3">
        <StepFields
          kind={step.kind}
          config={step.config}
          readOnly={readOnly}
          patternOptions={patternOptions}
          onPatternCreated={onPatternCreated}
          onChange={setConfig}
        />
      </div>
    </div>
  )
}

/** Renders friendly fields for common processors, raw YAML for the rest. */
function StepFields({
  kind,
  config,
  readOnly,
  patternOptions,
  onPatternCreated,
  onChange,
}: {
  kind: StepKind
  config: Record<string, unknown>
  readOnly?: boolean
  patternOptions: RegexPattern[]
  onPatternCreated?: (p: RegexPattern) => void
  onChange: (config: Record<string, unknown>) => void
}) {
  const { t } = useTranslation()
  const set = (key: string, val: unknown) => onChange({ ...config, [key]: val })
  const str = (key: string) => (typeof config[key] === 'string' ? (config[key] as string) : '')
  const list = (key: string) =>
    Array.isArray(config[key]) ? (config[key] as unknown[]).map((x) => String(x)) : []
  const obj = (key: string): Record<string, unknown> =>
    config[key] && typeof config[key] === 'object' && !Array.isArray(config[key])
      ? (config[key] as Record<string, unknown>)
      : {}
  const objStr = (key: string, sub: string) => {
    const v = obj(key)[sub]
    return typeof v === 'string' ? v : ''
  }
  const setObj = (key: string, sub: string, val: string) => set(key, { ...obj(key), [sub]: val })

  switch (kind) {
    case 'grok':
      return (
        <>
          <TextField label="source" required value={str('source')} readOnly={readOnly} onChange={(v) => set('source', v)} />
          <PatternsField
            value={config.patterns}
            readOnly={readOnly}
            patternOptions={patternOptions}
            onPatternCreated={onPatternCreated}
            onChange={(v) => set('patterns', v)}
          />
          <WhereField value={str('where')} readOnly={readOnly} onChange={(v) => set('where', v)} />
        </>
      )
    case 'json':
      return (
        <>
          <TextField label="source" required value={str('source')} readOnly={readOnly} onChange={(v) => set('source', v)} />
          <WhereField value={str('where')} readOnly={readOnly} onChange={(v) => set('where', v)} />
        </>
      )
    case 'kv':
      return (
        <>
          <TextField label="source" required value={str('source')} readOnly={readOnly} onChange={(v) => set('source', v)} />
          <div className="grid grid-cols-2 gap-3">
            <TextField label="fieldSplit" value={str('fieldSplit')} readOnly={readOnly} onChange={(v) => set('fieldSplit', v)} />
            <TextField label="valueSplit" value={str('valueSplit')} readOnly={readOnly} onChange={(v) => set('valueSplit', v)} />
          </div>
          <WhereField value={str('where')} readOnly={readOnly} onChange={(v) => set('where', v)} />
        </>
      )
    case 'rename':
      return (
        <>
          <TextField label="to" required value={str('to')} readOnly={readOnly} onChange={(v) => set('to', v)} />
          <ListField label="from" required values={list('from')} readOnly={readOnly} onChange={(v) => set('from', v)} />
          <WhereField value={str('where')} readOnly={readOnly} onChange={(v) => set('where', v)} />
        </>
      )
    case 'add':
      return (
        <>
          <TextField label="function" required value={str('function')} readOnly={readOnly} onChange={(v) => set('function', v)} />
          <div className="space-y-1.5 rounded-md border border-border/60 bg-background/30 p-2">
            <div className="font-mono text-[11px] text-foreground/70">{t('parsingFilters.visual.parameters')}</div>
            <div className="grid grid-cols-2 gap-3">
              <TextField label={t('parsingFilters.visual.paramsKey')} required value={objStr('params', 'key')} readOnly={readOnly} onChange={(v) => setObj('params', 'key', v)} />
              <TextField label={t('parsingFilters.visual.paramsValue')} required value={objStr('params', 'value')} readOnly={readOnly} onChange={(v) => setObj('params', 'value', v)} />
            </div>
          </div>
          <WhereField value={str('where')} readOnly={readOnly} onChange={(v) => set('where', v)} />
        </>
      )
    case 'drop':
      return <WhereField value={str('where')} readOnly={readOnly} onChange={(v) => set('where', v)} />
    default:
      return <RawBody config={config} readOnly={readOnly} onChange={onChange} />
  }
}

function TextField({
  label,
  value,
  required,
  readOnly,
  onChange,
}: {
  label: string
  value: string
  required?: boolean
  readOnly?: boolean
  onChange: (v: string) => void
}) {
  return (
    <div className="space-y-1">
      <label className="block font-mono text-[11px] text-foreground/70">
        {label}
        {required && <span className="ml-0.5 text-red-500">*</span>}
      </label>
      <Input
        value={value}
        readOnly={readOnly}
        onChange={(e) => onChange(e.target.value)}
        className="h-8 font-mono text-[12px]"
      />
    </div>
  )
}

function WhereField({
  value,
  readOnly,
  onChange,
}: {
  value: string
  readOnly?: boolean
  onChange: (v: string) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="space-y-1">
      <label className="block font-mono text-[11px] text-foreground/70">where</label>
      <Input
        value={value}
        readOnly={readOnly}
        onChange={(e) => onChange(e.target.value)}
        placeholder={t('parsingFilters.visual.whereHint')}
        className="h-8 font-mono text-[12px]"
      />
    </div>
  )
}

/** Editable list of strings (rename.from). */
function ListField({
  label,
  values,
  required,
  readOnly,
  onChange,
}: {
  label: string
  values: string[]
  required?: boolean
  readOnly?: boolean
  onChange: (v: string[]) => void
}) {
  const { t } = useTranslation()
  const setAt = (i: number, v: string) => onChange(values.map((x, k) => (k === i ? v : x)))
  return (
    <div className="space-y-1">
      <label className="block font-mono text-[11px] text-foreground/70">
        {label}
        {required && <span className="ml-0.5 text-red-500">*</span>}
      </label>
      <div className="space-y-1.5">
        {values.map((v, i) => (
          <div key={i} className="flex items-center gap-1.5">
            <Input value={v} readOnly={readOnly} onChange={(e) => setAt(i, e.target.value)} className="h-8 font-mono text-[12px]" />
            {!readOnly && (
              <button
                type="button"
                onClick={() => onChange(values.filter((_, k) => k !== i))}
                className="rounded p-1 text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
              >
                <X size={13} />
              </button>
            )}
          </div>
        ))}
        {!readOnly && (
          <Button type="button" variant="outline" size="sm" className="h-7" onClick={() => onChange([...values, ''])}>
            <Plus size={12} className="mr-1" /> {t('parsingFilters.visual.addValue')}
          </Button>
        )}
      </div>
    </div>
  )
}

/** grok.patterns — list of { fieldName, pattern }. */
function PatternsField({
  value,
  readOnly,
  patternOptions,
  onPatternCreated,
  onChange,
}: {
  value: unknown
  readOnly?: boolean
  patternOptions: RegexPattern[]
  onPatternCreated?: (p: RegexPattern) => void
  onChange: (v: Array<{ fieldName: string; pattern: string }>) => void
}) {
  const { t } = useTranslation()
  const patterns: Array<{ fieldName: string; pattern: string }> = Array.isArray(value)
    ? (value as unknown[]).map((p) => {
        const r = (p ?? {}) as Record<string, unknown>
        return { fieldName: String(r.fieldName ?? ''), pattern: String(r.pattern ?? '') }
      })
    : []
  const setAt = (i: number, patch: Partial<{ fieldName: string; pattern: string }>) =>
    onChange(patterns.map((p, k) => (k === i ? { ...p, ...patch } : p)))

  return (
    <div className="space-y-1">
      <label className="block font-mono text-[11px] text-foreground/70">
        patterns<span className="ml-0.5 text-red-500">*</span>
      </label>
      <div className="space-y-1.5">
        {patterns.map((p, i) => (
          <div key={i} className="flex items-center gap-1.5">
            <Input
              value={p.fieldName}
              readOnly={readOnly}
              onChange={(e) => setAt(i, { fieldName: e.target.value })}
              placeholder="fieldName"
              className="h-8 w-[34%] font-mono text-[12px]"
            />
            <PatternInput
              value={p.pattern}
              readOnly={readOnly}
              patternOptions={patternOptions}
              onPatternCreated={onPatternCreated}
              onChange={(v) => setAt(i, { pattern: v })}
            />
            {!readOnly && (
              <button
                type="button"
                onClick={() => onChange(patterns.filter((_, k) => k !== i))}
                className="rounded p-1 text-muted-foreground hover:bg-red-500/10 hover:text-red-500"
              >
                <X size={13} />
              </button>
            )}
          </div>
        ))}
        {!readOnly && (
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="h-7"
            onClick={() => onChange([...patterns, { fieldName: '', pattern: '' }])}
          >
            <Plus size={12} className="mr-1" /> {t('parsingFilters.visual.addPattern')}
          </Button>
        )}
      </div>
    </div>
  )
}

/**
 * grok pattern field with regex-pattern autocomplete. Detects an open `{{.foo`
 * token before the caret and suggests matching named patterns; selecting one
 * inserts `{{.name}}`. A `{ }` button opens the full searchable catalog (which
 * also offers an inline "+ new pattern" create form).
 */
const TOKEN_RE = /\{\{\.?(\w*)$/

function PatternInput({
  value,
  readOnly,
  patternOptions,
  onPatternCreated,
  onChange,
}: {
  value: string
  readOnly?: boolean
  patternOptions: RegexPattern[]
  onPatternCreated?: (p: RegexPattern) => void
  onChange: (v: string) => void
}) {
  const { t } = useTranslation()
  const inputRef = useRef<HTMLInputElement>(null)
  const wrapRef = useRef<HTMLDivElement>(null)
  const caretRef = useRef<number | null>(null)
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  // Range in `value` being completed while typing a token; null when opened via the button.
  const [token, setToken] = useState<{ start: number; end: number } | null>(null)

  const closeAll = () => {
    setOpen(false)
    setToken(null)
  }

  // Close on outside click.
  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) closeAll()
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  // Restore the caret after a programmatic value change (insertion).
  useLayoutEffect(() => {
    if (caretRef.current != null && inputRef.current) {
      const pos = caretRef.current
      inputRef.current.setSelectionRange(pos, pos)
      caretRef.current = null
    }
  })

  const detectToken = (full: string, caret: number) => {
    const m = TOKEN_RE.exec(full.slice(0, caret))
    if (m) {
      setToken({ start: caret - m[0].length, end: caret })
      setQuery(m[1] ?? '')
      setOpen(true)
    } else if (token) {
      setToken(null)
      setOpen(false)
    }
  }

  const insert = (name: string) => {
    const ref = `{{.${name}}}`
    let next: string
    let caret: number
    if (token) {
      next = value.slice(0, token.start) + ref + value.slice(token.end)
      caret = token.start + ref.length
    } else {
      const at = inputRef.current?.selectionStart ?? value.length
      next = value.slice(0, at) + ref + value.slice(at)
      caret = at + ref.length
    }
    caretRef.current = caret
    onChange(next)
    closeAll()
    requestAnimationFrame(() => inputRef.current?.focus())
  }

  return (
    <div ref={wrapRef} className="relative flex-1">
      <Input
        ref={inputRef}
        value={value}
        readOnly={readOnly}
        onChange={(e) => {
          onChange(e.target.value)
          detectToken(e.target.value, e.target.selectionStart ?? e.target.value.length)
        }}
        onKeyDown={(e) => {
          if (e.key === 'Escape' && open) {
            e.stopPropagation()
            closeAll()
          }
        }}
        placeholder="pattern"
        className="h-8 w-full pr-8 font-mono text-[12px]"
      />
      {!readOnly && (
        <button
          type="button"
          title={t('parsingFilters.visual.insertPattern')}
          onClick={() => {
            if (open) {
              closeAll()
            } else {
              setToken(null)
              setQuery('')
              setOpen(true)
            }
            inputRef.current?.focus()
          }}
          className="absolute right-1 top-1/2 -translate-y-1/2 rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <Braces size={13} />
        </button>
      )}
      {open && !readOnly && (
        <div className="absolute right-0 top-[calc(100%+4px)] z-20 w-[340px] max-w-[90vw] overflow-hidden rounded-md border border-border bg-popover shadow-lg">
          <PatternMenu
            patternOptions={patternOptions}
            query={query}
            onQueryChange={token ? undefined : setQuery}
            onPick={insert}
            onPatternCreated={onPatternCreated}
          />
        </div>
      )}
    </div>
  )
}

/** Raw YAML body for processors without friendly fields (csv, trim, cast, …). */
function RawBody({
  config,
  readOnly,
  onChange,
}: {
  config: Record<string, unknown>
  readOnly?: boolean
  onChange: (config: Record<string, unknown>) => void
}) {
  const { t } = useTranslation()
  const [text, setText] = useState(() => {
    try {
      return Object.keys(config).length ? dump(config, { indent: 2, lineWidth: -1 }) : ''
    } catch {
      return ''
    }
  })
  const [err, setErr] = useState<string | null>(null)

  const commit = (next: string) => {
    setText(next)
    if (next.trim() === '') {
      setErr(null)
      onChange({})
      return
    }
    try {
      const parsed = load(next)
      if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
        setErr(null)
        onChange(parsed as Record<string, unknown>)
      } else {
        setErr(t('parsingFilters.visual.rawNotMap'))
      }
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'invalid YAML')
    }
  }

  return (
    <div className="space-y-1">
      <p className="text-[11px] text-muted-foreground">{t('parsingFilters.visual.rawHint')}</p>
      <textarea
        value={text}
        readOnly={readOnly}
        spellCheck={false}
        rows={5}
        onChange={(e) => commit(e.target.value)}
        placeholder={'source: log\nseparator: ","'}
        className="w-full rounded-md border border-input bg-background/40 p-2 font-mono text-[11px] leading-relaxed focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring read-only:opacity-80"
      />
      {err && <p className="text-[11px] text-red-500">{err}</p>}
    </div>
  )
}
