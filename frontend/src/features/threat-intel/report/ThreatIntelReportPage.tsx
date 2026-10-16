import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery } from '@tanstack/react-query'
import { toast } from 'sonner'
import { IocDrawer } from '../components/IocDrawer'
import { IndicatorForm } from './indicator-form'
import { SubmissionReceipt } from './submission-receipt'
import {
  TypeInferenceError,
  attributeProblems,
  associationConflicts,
  isKnownEntityType,
  listEntityDefinitions,
  reportIoc,
  resolveDefinition,
  type ReportIocInput,
  type ReportIocResult,
} from './report'

const EMPTY: ReportIocInput = { value: '' }

export function ThreatIntelReportPage() {
  const { t } = useTranslation()
  const [input, setInput] = useState<ReportIocInput>(EMPTY)
  const [result, setResult] = useState<ReportIocResult | null>(null)
  const [openIoc, setOpenIoc] = useState<string | null>(null)

  const submit = useMutation({
    mutationFn: reportIoc,
    onSuccess: setResult,
    onError: (e) => {
      if (!(e instanceof TypeInferenceError)) toast.error(e instanceof Error ? e.message : String(e))
    },
  })

  const definitions = useQuery({
    queryKey: ['ti-report-definitions'],
    queryFn: listEntityDefinitions,
    staleTime: 1000 * 60 * 60,
  })

  const typeInferenceFailed = submit.error instanceof TypeInferenceError
  const typeUnknown = !isKnownEntityType(input.type, definitions.data)

  const definition = resolveDefinition(definitions.data, input)
  const hasAttributeProblem = Object.keys(attributeProblems(definition, input.attributes)).length > 0
  const hasVisibilityConflict = associationConflicts(input.visibleBy, input.associations ?? []).length > 0

  const submitError =
    submit.error && !typeInferenceFailed
      ? t('threatIntel.report.error.generic')
      : null

  return (
    <div className="flex h-full flex-col overflow-y-auto px-6 py-6">
      <h1 className="text-lg font-semibold">{t('nav.threat-intel.report')}</h1>
      <div className="mx-auto mt-4 w-full max-w-2xl space-y-6">
      <header className="flex flex-wrap items-baseline gap-3">
        <h1 className="text-xl font-bold tracking-widest text-foreground sm:text-2xl">
          {t('threatIntel.report.title').toUpperCase()}
        </h1>
        <span className="font-mono text-[10px] uppercase tracking-widest text-muted-foreground">
          {t('threatIntel.report.subtitle')}
        </span>
      </header>

      {result ? (
        <SubmissionReceipt
          result={result}
          onEntityView={setOpenIoc}
          onReportAnother={() => {
            setResult(null)
            setInput({ ...EMPTY })
            submit.reset()
          }}
        />
      ) : (
        <form
          onSubmit={(e) => {
            e.preventDefault()
            submit.mutate(input)
          }}
          className="space-y-6"
        >
          <IndicatorForm
            value={input}
            onChange={(next) => {
              setInput(next)
              if (submit.error) submit.reset()
            }}
            disabled={submit.isPending}
            typeError={
              typeInferenceFailed
                ? t('threatIntel.report.type.inferenceFailed')
                : typeUnknown
                  ? t('threatIntel.report.type.unknown', { type: input.type?.trim() ?? '' })
                  : undefined
            }
          />

          {submitError && <p className="text-sm text-destructive">{submitError}</p>}

          <button
            type="submit"
            className="inline-flex items-center justify-center gap-1.5 rounded-sm border border-primary/40 px-3 py-1.5 font-mono text-[10px] uppercase tracking-wider text-primary transition-colors hover:bg-primary/10 disabled:pointer-events-none disabled:opacity-40"
            disabled={
              !input.value.trim() ||
              typeUnknown ||
              hasAttributeProblem ||
              hasVisibilityConflict ||
              submit.isPending
            }
          >
            {submit.isPending ? t('threatIntel.report.submitting') : t('threatIntel.report.submit')}
          </button>
        </form>
      )}

      </div>
      <IocDrawer id={openIoc} onClose={() => setOpenIoc(null)} />
    </div>
  )
}
