import { cn } from '@/shared/lib/utils'
import { useTranslation } from 'react-i18next'
import { Input } from '@/shared/components/ui/input'
import { Textarea } from '@/shared/components/ui/textarea'
import { Plus, X } from 'lucide-react'
import { useState, useEffect, useId, useRef } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  associationConflicts,
  controlFor,
  correlatableAttributes,
  detectIndicatorTypes,
  isScannedType,
  listEntityDefinitions,
  nextAttributes,
  nextCorrelate,
  resolveDefinition,
  attributeProblems,
  type ReportIocInput,
  type ReportAssociationInput,
} from './report'

const LABEL = 'font-mono text-[10px] uppercase tracking-widest text-muted-foreground'
const FIELD =
  'h-9 rounded-sm border border-border/60 bg-muted/30 font-mono text-sm placeholder:text-muted-foreground/40 focus:outline-none focus:ring-1 focus:ring-primary'
const TOGGLE = (selected: boolean) =>
  cn(
    'rounded-sm border px-3 py-1.5 font-mono text-[10px] uppercase tracking-wider transition-colors',
    selected
      ? 'border-primary bg-primary/10 text-primary'
      : 'border-border/60 text-muted-foreground hover:bg-muted hover:text-foreground',
  )

function FieldLabel({ htmlFor, children }: { htmlFor?: string; children: React.ReactNode }) {
  return <label htmlFor={htmlFor} className={cn(LABEL, 'block')}>{children}</label>
}
function FieldLegend({ children }: { children: React.ReactNode }) {
  return <legend className={LABEL}>{children}</legend>
}
function Frame({ children }: { children: React.ReactNode }) {
  return (
    <fieldset className="war-frame relative space-y-3 rounded-sm border border-border/50 bg-muted/20 p-3">
      {children}
    </fieldset>
  )
}

function parseTags(raw: string): string[] {
  return raw.split(',').map((s) => s.trim()).filter(Boolean)
}
function joinTags(tags: string[] | undefined): string {
  return (tags ?? []).join(', ')
}

function ReputationInput({ value, onChange }: { value: number | undefined; onChange: (v: number | undefined) => void }) {
  const { t } = useTranslation()
  const options = [
    { score: -3, key: 'threatIntel.report.reputation.alarming' },
    { score: -2, key: 'threatIntel.report.reputation.worrying' },
    { score: -1, key: 'threatIntel.report.reputation.concerning' },
    { score: 0, key: 'threatIntel.report.reputation.indefinable' },
    { score: 1, key: 'threatIntel.report.reputation.adequate' },
    { score: 2, key: 'threatIntel.report.reputation.good' },
    { score: 3, key: 'threatIntel.report.reputation.exceptional' },
  ]
  return (
    <Frame>
      <FieldLegend>{t('threatIntel.report.reputation.legend')}</FieldLegend>
      <p className="text-xs text-muted-foreground">{t('threatIntel.report.reputation.hint')}</p>
      <div className="mt-2 flex flex-wrap gap-2">
        {options.map((option) => {
          const selected = value === option.score
          return (
            <button
              key={option.score}
              type="button"
              aria-pressed={selected}
              onClick={() => onChange(selected ? undefined : option.score)}
              className={TOGGLE(selected) + (option.score < 0 && selected ? ' border-destructive bg-destructive/10 text-destructive' : '')}
            >
              {t(option.key)}
            </button>
          )
        })}
      </div>
    </Frame>
  )
}

function AttributeFields({
  definition,
  value,
  onChange,
}: {
  definition: ReturnType<typeof resolveDefinition>
  value: Record<string, string> | undefined
  onChange: (next: Record<string, string>) => void
}) {
  const { t } = useTranslation()
  const scope = useId()
  const attributes = definition?.attributes ?? []

  useEffect(() => {
    const next = nextAttributes(definition, value)
    if (next) onChange(next)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [definition?.type, value && Object.keys(value).join(',')])

  if (attributes.length === 0) return null

  const problems = attributeProblems(definition, value)
  const set = (name: string, next: string) => onChange({ ...(value ?? {}), [name]: next })

  return (
    <Frame>
      <FieldLegend>{t('threatIntel.report.attributes.legend')}</FieldLegend>
      <p className={LABEL}>{t('threatIntel.report.attributes.hint')}</p>
      <div className="space-y-3">
        {attributes.map((attribute) => {
          const id = `${scope}-${attribute.type}`
          const problem = problems[attribute.type]
          const errorId = `${id}-error`
          return (
            <div key={attribute.type} className="space-y-1.5">
              <FieldLabel htmlFor={id}>{attribute.label}</FieldLabel>
              <Input
                id={id}
                type={controlFor(attribute.dataType)}
                className={FIELD}
                value={value?.[attribute.type] ?? ''}
                placeholder={attribute.description}
                aria-invalid={problem ? true : undefined}
                aria-describedby={problem ? errorId : undefined}
                onChange={(e) => set(attribute.type, e.target.value)}
              />
              {problem && (
                <p id={errorId} className="text-sm text-destructive">
                  {problem.kind === 'number'
                    ? t('threatIntel.report.attributes.invalidNumber')
                    : problem.length === undefined
                      ? t('threatIntel.report.attributes.invalidHexAny')
                      : t('threatIntel.report.attributes.invalidHex', { length: String(problem.length) })}
                </p>
              )}
            </div>
          )
        })}
      </div>
    </Frame>
  )
}

function CorrelatePicker({
  definition,
  attributes,
  value,
  onChange,
}: {
  definition: ReturnType<typeof resolveDefinition>
  attributes: Record<string, string> | undefined
  value: string[] | undefined
  onChange: (next: string[]) => void
}) {
  const { t } = useTranslation()
  const scope = useId()
  const available = correlatableAttributes(definition, attributes)
  const [dismissed, setDismissed] = useState<string[]>([])

  useEffect(() => {
    const next = nextCorrelate(available, value, dismissed)
    if (next) onChange(next)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [available.join(','), value?.join(','), dismissed.join(',')])

  if (available.length === 0) return null

  const selected = value ?? available.filter((name) => !dismissed.includes(name))

  return (
    <Frame>
      <FieldLegend>{t('threatIntel.report.correlate.legend')}</FieldLegend>
      <p className={LABEL}>{t('threatIntel.report.correlate.hint')}</p>
      <div className="mt-2 flex flex-wrap gap-2">
        {available.map((name) => {
          const on = selected.includes(name)
          return (
            <button
              key={`${scope}-${name}`}
              type="button"
              aria-pressed={on}
              onClick={() => {
                setDismissed((d) => (on ? [...d, name] : d.filter((n) => n !== name)))
                onChange(on ? selected.filter((n) => n !== name) : [...selected, name])
              }}
              className={TOGGLE(on)}
            >
              {name}
            </button>
          )
        })}
      </div>
      {selected.length > 0 && (
        <p className={LABEL}>
          {t(selected.length === 1 ? 'threatIntel.report.correlate.count_one' : 'threatIntel.report.correlate.count_other', { count: String(selected.length) })}
        </p>
      )}
    </Frame>
  )
}

function AssociationList({
  value,
  onChange,
  parentVisibleBy,
}: {
  value: ReportAssociationInput[]
  onChange: (value: ReportAssociationInput[]) => void
  parentVisibleBy?: string[]
}) {
  const { t } = useTranslation()
  const conflicts = new Set(associationConflicts(parentVisibleBy, value))

  const update = (index: number, patch: Partial<ReportAssociationInput>) =>
    onChange(value.map((a, i) => (i === index ? { ...a, ...patch } : a)))

  return (
    <Frame>
      <FieldLegend>{t('threatIntel.report.associations.legend')}</FieldLegend>
      <div className="mt-2 space-y-3">
        {value.map((association, index) => {
          const conflicted = conflicts.has(index)
          const chosen = association.visibleBy
          return (
            <div key={index} className="space-y-2">
              <div className="flex gap-2">
                <Input
                  value={association.value}
                  placeholder={t('threatIntel.report.associations.valuePlaceholder')}
                  className={FIELD}
                  aria-invalid={conflicted ? true : undefined}
                  onChange={(e) => update(index, { value: e.target.value })}
                />
                <Input
                  value={association.type ?? ''}
                  placeholder={t('threatIntel.report.associations.typePlaceholder')}
                  className={FIELD + ' max-w-[10rem]'}
                  onChange={(e) => update(index, { type: e.target.value || undefined })}
                />
                <button
                  type="button"
                  aria-label={t('threatIntel.report.associations.remove')}
                  onClick={() => onChange(value.filter((_, i) => i !== index))}
                  className="shrink-0 rounded-sm p-1 text-muted-foreground transition-colors hover:text-destructive"
                >
                  <X className="h-4 w-4" />
                </button>
              </div>

              <div className="flex flex-wrap items-center gap-2">
                <span className={LABEL}>{t('threatIntel.report.associations.modeLabel')}</span>
                {(['association', 'aggregation'] as const).map((mode) => {
                  const on = (association.mode ?? 'association') === mode
                  return (
                    <button
                      key={mode}
                      type="button"
                      aria-pressed={on}
                      onClick={() => update(index, { mode })}
                      className={TOGGLE(on)}
                    >
                      {t(`threatIntel.report.associations.mode.${mode}`)}
                    </button>
                  )
                })}

                {chosen &&
                  chosen.map((group) => {
                    const on = true
                    return (
                      <button
                        key={group}
                        type="button"
                        aria-pressed={on}
                        onClick={() => {
                          const next = chosen.filter((g) => g !== group)
                          update(index, { visibleBy: next.length ? next : undefined })
                        }}
                        className={TOGGLE(on)}
                      >
                        {group}
                      </button>
                    )
                  })}
              </div>

              {conflicted && (
                <p className="text-sm text-destructive">
                  {t('threatIntel.report.associations.visibilityConflict')}
                </p>
              )}
            </div>
          )
        })}
        <button
          type="button"
          className="inline-flex items-center gap-1.5 rounded-sm border border-primary/40 px-3 py-1.5 font-mono text-[10px] uppercase tracking-wider text-primary transition-colors hover:bg-primary/10"
          onClick={() => onChange([...value, { value: '', mode: 'association' }])}
        >
          <Plus className="h-3 w-3" />
          {t('threatIntel.report.associations.add')}
        </button>
      </div>
    </Frame>
  )
}

export function IndicatorForm({
  value,
  onChange,
  disabled,
  typeError,
}: {
  value: ReportIocInput
  onChange: (value: ReportIocInput) => void
  disabled?: boolean
  typeError?: string
}) {
  const { t } = useTranslation()
  const inferred = value.value.trim() ? detectIndicatorTypes(value.value.trim())[0] : undefined

  const latest = useRef(value)
  latest.current = value
  const set = (patch: Partial<ReportIocInput>) => {
    latest.current = { ...latest.current, ...patch }
    onChange(latest.current)
  }

  const definitions = useQuery({
    queryKey: ['ti-report-definitions'],
    queryFn: listEntityDefinitions,
    staleTime: 1000 * 60 * 60,
  })

  const definition = resolveDefinition(definitions.data, value)
  const activeType = value.type?.trim() || inferred
  const queuesScan = isScannedType(activeType)

  const typeListId = `${useId()}-types`
  const typeErrorId = `${typeListId}-error`

  const [tagsInput, setTagsInput] = useState(() => joinTags(value.tags))
  const handleTagsChange = (raw: string) => {
    setTagsInput(raw)
    set({ tags: parseTags(raw) })
  }

  return (
    <fieldset disabled={disabled} className="space-y-4">
      <div className="war-frame relative space-y-4 rounded-sm border border-border/50 bg-muted/20 p-3">
        <div className="space-y-1.5">
          <FieldLabel htmlFor="report-value">{t('threatIntel.report.value.label')}</FieldLabel>
          <Input
            id="report-value"
            className={FIELD}
            value={value.value}
            placeholder={t('threatIntel.report.value.placeholder')}
            onChange={(e) => set({ value: e.target.value })}
          />
        </div>

        <div>
          <FieldLabel htmlFor="report-type">{t('threatIntel.report.type.label')}</FieldLabel>
          <Input
            id="report-type"
            className={FIELD}
            list={typeListId}
            value={value.type ?? ''}
            placeholder={inferred ? t('threatIntel.report.type.inferred', { type: inferred }) : t('threatIntel.report.type.placeholder')}
            aria-invalid={typeError ? true : undefined}
            aria-describedby={typeError ? typeErrorId : undefined}
            onChange={(e) => set({ type: e.target.value || undefined })}
          />
          <datalist id={typeListId}>
            {(definitions.data ?? []).map((definition) => (
              <option key={definition.type} value={definition.type}>
                {definition.label}
              </option>
            ))}
          </datalist>
          {typeError && (
            <p id={typeErrorId} className="mt-1 text-sm text-destructive">
              {typeError}
            </p>
          )}
        </div>

        <ReputationInput value={value.reputation} onChange={(reputation) => set({ reputation })} />

        <div className="space-y-1.5">
          <FieldLabel htmlFor="report-tags">{t('threatIntel.report.tags.label')}</FieldLabel>
          <Input
            id="report-tags"
            className={FIELD}
            value={tagsInput}
            placeholder={t('threatIntel.report.tags.placeholder')}
            onChange={(e) => handleTagsChange(e.target.value)}
          />
        </div>

        <div className="space-y-1.5">
          <FieldLabel htmlFor="report-description">{t('threatIntel.report.description.label')}</FieldLabel>
          <Textarea
            id="report-description"
            className={FIELD + ' h-auto py-2'}
            value={value.description ?? ''}
            rows={3}
            onChange={(e) => set({ description: e.target.value || undefined })}
          />
        </div>

        {queuesScan && <p className={LABEL}>{t('threatIntel.report.scan.notice')}</p>}
      </div>

      <AttributeFields definition={definition} value={value.attributes} onChange={(attributes) => set({ attributes })} />

      <CorrelatePicker
        key={definition?.type}
        definition={definition}
        attributes={value.attributes}
        value={value.correlate}
        onChange={(correlate) => set({ correlate })}
      />

      <AssociationList
        value={value.associations ?? []}
        onChange={(associations) => set({ associations })}
        parentVisibleBy={value.visibleBy}
      />
    </fieldset>
  )
}
