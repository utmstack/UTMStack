import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { useMutation, useQuery } from '@tanstack/react-query'
import { toast } from 'sonner'
import { Button } from '@/shared/components/ui/button'
import { IocDrawer } from '../components/IocDrawer'
import { NotConfiguredState } from '../components/NotConfiguredState'
import { useTiConfigStatus } from '../hooks/use-ti-config-status'
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
  const { isConfigured } = useTiConfigStatus()
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

  if (!isConfigured) return <NotConfiguredState />

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
    <div className="w-full overflow-y-auto px-6 pb-6 pt-3">
      <div className="  space-y-4">
        <header className="flex flex-wrap items-baseline gap-3">
          <h1 className="text-xl font-semibold">{t('threatIntel.report.title')}</h1>
          <span className="text-xs text-muted-foreground">{t('threatIntel.report.subtitle')}</span>
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
            className="space-y-4"
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

            <div className="flex justify-end">
              <Button
                type="submit"
                disabled={
                  !input.value.trim() ||
                  typeUnknown ||
                  hasAttributeProblem ||
                  hasVisibilityConflict ||
                  submit.isPending
                }
              >
                {submit.isPending ? t('threatIntel.report.submitting') : t('threatIntel.report.submit')}
              </Button>
            </div>
          </form>
        )}

        <IocDrawer id={openIoc} onClose={() => setOpenIoc(null)} />
      </div>
    </div>
  )
}
