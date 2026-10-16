import { useTranslation } from 'react-i18next'
import { Input } from '@/shared/components/ui/input'
import { Textarea } from '@/shared/components/ui/textarea'
import { Button } from '@/shared/components/ui/button'
import { Plus, X,Check } from 'lucide-react'
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

const LABEL = 'block text-[11px] uppercase tracking-wider text-muted-foreground'

function Card({ children }: { children: React.ReactNode }) {
  return (
    <fieldset className=" rounded-xl border border-border bg-card p-4">
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

const VISIBILITY_GROUPS = ['public']

function VisibilityPicker({ value, onChange }: { value: string[]; onChange: (v: string[]) => void }) {
  const { t } = useTranslation()
  return (
    <Card>
      <h3 className="text-sm font-medium text-foreground">{t('threatIntel.report.visibility.legend')}</h3>
      <div className="mt-2 flex flex-wrap gap-2">
        {VISIBILITY_GROUPS.map((group) => {
          const selected = value.includes(group)
          return (
            <Button
              key={group}
              type="button"
              size="sm"
              aria-pressed={selected}
              variant={selected ? 'default' : 'outline'}
              onClick={() =>
                onChange(selected ? value.filter((g) => g !== group) : [...value, group])
              }
            >
              {group}
            </Button>
          )
        })}
      </div>
    </Card>
  )
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
    <Card>
      <h3 className="text-sm font-medium text-foreground">{t('threatIntel.report.reputation.legend')}</h3>
      <p className="text-xs text-muted-foreground">{t('threatIntel.report.reputation.hint')}</p>
      <div className="mt-2 flex flex-wrap gap-2">
        {options.map((option) => {
          const selected = value === option.score
          return (
            <Button
              key={option.score}
              type="button"
              size="sm"
              aria-pressed={selected}
              onClick={() => onChange(selected ? undefined : option.score)}

              variant={selected && option.score < 0 ? 'destructive' : selected && option.score==0 ? 'secondary': selected?'default' : 'outline'}
            >
              {selected && (<Check size="20" className="scale-0.1 mr-3" />)}
              {t(option.key)}  {option.score}
            </Button>
          )
        })}
      </div>
    </Card>
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
    <Card>
      <h3 className={LABEL}>{t('threatIntel.report.attributes.legend')}</h3>
      <p className="text-xs text-muted-foreground">{t('threatIntel.report.attributes.hint')}</p>
      <div className="space-y-3">
        {attributes.map((attribute) => {
          const id = `${scope}-${attribute.type}`
          const problem = problems[attribute.type]
          const errorId = `${id}-error`
          return (
            <div key={attribute.type} className="space-y-1.5">
              <label htmlFor={id} className={LABEL}>{attribute.label}</label>
              <Input
                id={id}
                type={controlFor(attribute.dataType)}
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
    </Card>
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
    <Card>
      <h3 className={LABEL}>{t('threatIntel.report.correlate.legend')}</h3>
      <p className="text-xs text-muted-foreground">{t('threatIntel.report.correlate.hint')}</p>
      <div className="mt-2 flex flex-wrap gap-2">
        {available.map((name) => {
          const on = selected.includes(name)
          return (
            <Button
              key={`${scope}-${name}`}
              type="button"
              size="sm"
              aria-pressed={on}
              variant={on ? 'secondary' : 'outline'}
              onClick={() => {
                setDismissed((d) => (on ? [...d, name] : d.filter((n) => n !== name)))
                onChange(on ? selected.filter((n) => n !== name) : [...selected, name])
              }}
            >
              {name}
            </Button>
          )
        })}
      </div>
      {selected.length > 0 && (
        <p className="text-xs text-muted-foreground">
          {t(selected.length === 1 ? 'threatIntel.report.correlate.count_one' : 'threatIntel.report.correlate.count_other', { count: String(selected.length) })}
        </p>
      )}
    </Card>
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
    <Card>
      <h3 className="text-sm font-medium text-foreground">{t('threatIntel.report.associations.legend')}</h3>
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
                  aria-invalid={conflicted ? true : undefined}
                  onChange={(e) => update(index, { value: e.target.value })}
                />
                <Input
                  value={association.type ?? ''}
                  placeholder={t('threatIntel.report.associations.typePlaceholder')}
                  className="max-w-[10rem]"
                  onChange={(e) => update(index, { type: e.target.value || undefined })}
                />
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  aria-label={t('threatIntel.report.associations.remove')}
                  onClick={() => onChange(value.filter((_, i) => i !== index))}
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>

              <div className="flex flex-wrap items-center gap-2">
                <span className="text-[11px] uppercase tracking-wider text-muted-foreground">
                  {t('threatIntel.report.associations.modeLabel')}
                </span>
                {(['association', 'aggregation'] as const).map((mode) => {
                  const on = (association.mode ?? 'association') === mode
                  return (
                    <Button
                      key={mode}
                      type="button"
                      size="sm"
                      aria-pressed={on}
                      variant={on ? 'secondary' : 'outline'}
                      onClick={() => update(index, { mode })}
                    >
                      {t(`threatIntel.report.associations.mode.${mode}`)}
                    </Button>
                  )
                })}

                {chosen &&
                  chosen.map((group) => {
                    const on = true
                    return (
                      <Button
                        key={group}
                        type="button"
                        size="sm"
                        aria-pressed={on}
                        variant={on ? 'secondary' : 'outline'}
                        onClick={() => {
                          const next = chosen.filter((g) => g !== group)
                          update(index, { visibleBy: next.length ? next : undefined })
                        }}
                      >
                        {group}
                      </Button>
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
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => onChange([...value, { value: '', mode: 'association' }])}
        >
          <Plus className="mr-1.5 h-3 w-3" />
          {t('threatIntel.report.associations.add')}
        </Button>
      </div>
    </Card>
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
  const [visibleBySeeded, setVisibleBySeeded] = useState(false)
  const handleTagsChange = (raw: string) => {
    setTagsInput(raw)
    set({ tags: parseTags(raw) })
  }

  useEffect(() => {
    if (visibleBySeeded) return
    if (value.visibleBy === undefined) {
      setVisibleBySeeded(true)
      set({ visibleBy: ['public'] })
    }
  }, [value.visibleBy, visibleBySeeded, set])

  return (
    <fieldset disabled={disabled} className="space-y-4">
      <Card>
        <div className="space-y-1.5">
          <label htmlFor="report-value" className={LABEL}>{t('threatIntel.report.value.label')}</label>
          <Input
            id="report-value"
            dir="ltr"
            value={value.value}
            placeholder={t('threatIntel.report.value.placeholder')}
            onChange={(e) => set({ value: e.target.value })}
          />
        </div>

        <div>
          <label htmlFor="report-type" className={LABEL}>{t('threatIntel.report.type.label')}</label>
          <Input
            id="report-type"
            dir="ltr"
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

        <div className="space-y-1.5">
          <label htmlFor="report-tags" className={LABEL}>{t('threatIntel.report.tags.label')}</label>
          <Input
            id="report-tags"
            dir="ltr"
            value={tagsInput}
            placeholder={t('threatIntel.report.tags.placeholder')}
            onChange={(e) => handleTagsChange(e.target.value)}
          />
        </div>

        <div className="space-y-1.5">
          <label htmlFor="report-description" className={LABEL}>{t('threatIntel.report.description.label')}</label>
          <Textarea
            id="report-description"
            value={value.description ?? ''}
            rows={3}
            onChange={(e) => set({ description: e.target.value || undefined })}
          />
        </div>

        {queuesScan && (
          <p className="text-xs text-muted-foreground">{t('threatIntel.report.scan.notice')}</p>
        )}
      </Card>

      <AttributeFields definition={definition} value={value.attributes} onChange={(attributes) => set({ attributes })} />

      <CorrelatePicker
        key={definition?.type}
        definition={definition}
        attributes={value.attributes}
        value={value.correlate}
        onChange={(correlate) => set({ correlate })}
      />

      <ReputationInput value={value.reputation} onChange={(reputation) => set({ reputation })} />

      <AssociationList
        value={value.associations ?? []}
        onChange={(associations) => set({ associations })}
        parentVisibleBy={value.visibleBy}
      />

      <VisibilityPicker
        value={value.visibleBy ?? []}
        onChange={(visibleBy) => set({ visibleBy })}
      />
    </fieldset>
  )
}
