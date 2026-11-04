import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { proxyFetch } from '../console/proxy'
import type { ReportIocResult, ReportedAssociation } from '../lib/report'

export const pollPlan = { intervalMs: 5000, timeoutMs: 120000 } as const

export type ReceiptStatus = 'pending' | 'live' | 'processing'

export function receiptStatus(input: { found: boolean; elapsedMs: number }): ReceiptStatus {
  if (input.found) return 'live'
  return input.elapsedMs >= pollPlan.timeoutMs ? 'processing' : 'pending'
}

interface SearchItem {
  id: string
  type: string
  attributes: Record<string, string>
}

async function findEntity(value: string, type: string): Promise<string | null> {
  try {
    const item = await proxyFetch<SearchItem>(`/entity/lookup`, {
      method: 'POST',
      body: JSON.stringify({ type, value }),
    })
    return item?.id ?? null
  } catch {
    return null
  }
}

function reputationKey(score: number | undefined): string | undefined {
  const map: Record<number, string> = {
    [-3]: 'threatIntel.report.reputation.alarming',
    [-2]: 'threatIntel.report.reputation.worrying',
    [-1]: 'threatIntel.report.reputation.concerning',
    0: 'threatIntel.report.reputation.indefinable',
    1: 'threatIntel.report.reputation.adequate',
    2: 'threatIntel.report.reputation.good',
    3: 'threatIntel.report.reputation.exceptional',
  }
  return score === undefined ? undefined : map[score]
}

export function SubmissionReceipt({
  result,
  onEntityView,
  onReportAnother,
}: {
  result: ReportIocResult
  onEntityView?: (id: string) => void
  onReportAnother: () => void
}) {
  const { t } = useTranslation()
  const [entityId, setEntityId] = useState<string | null>(null)
  const [elapsedMs, setElapsedMs] = useState(0)

  useEffect(() => {
    let cancelled = false
    const startedAt = Date.now()

    const tick = async () => {
      if (cancelled) return
      const elapsed = Date.now() - startedAt
      setElapsedMs(elapsed)
      if (elapsed >= pollPlan.timeoutMs) return

      const id = await findEntity(result.value, result.type)
      if (cancelled) return
      if (id) {
        setEntityId(id)
        return
      }
      window.setTimeout(tick, pollPlan.intervalMs)
    }

    const timer = window.setTimeout(tick, pollPlan.intervalMs)
    return () => {
      cancelled = true
      window.clearTimeout(timer)
    }
  }, [result.value, result.type])

  const status = receiptStatus({ found: Boolean(entityId), elapsedMs })
  const repKey = reputationKey(result.reputation)

  return (
    <div className="space-y-4 rounded-xl border border-border bg-card p-4">
      <h2 className="text-[11px] uppercase tracking-wider text-muted-foreground">
        {t('threatIntel.report.receipt.title')}
      </h2>

      <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
        <dt className="text-muted-foreground">{t('threatIntel.report.type.label')}</dt>
        <dd>{result.type}</dd>
        <dt className="text-muted-foreground">{t('threatIntel.report.value.label')}</dt>
        <dd className="break-all">{result.value}</dd>
        <dt className="text-muted-foreground">{t('threatIntel.report.reputation.legend')}</dt>
        <dd>{result.reputation === undefined ? t('threatIntel.report.receipt.noVerdict') : repKey ? t(repKey) : result.reputation}</dd>
        {result.tags.length > 0 && (
          <>
            <dt className="text-muted-foreground">{t('threatIntel.report.tags.label')}</dt>
            <dd>{result.tags.join(', ')}</dd>
          </>
        )}
        {result.associations.length > 0 && (
          <>
            <dt className="text-muted-foreground">{t('threatIntel.report.associations.legend')}</dt>
            <dd>
              <AssociationTree associations={result.associations} />
            </dd>
          </>
        )}
      </dl>

      {result.skippedAssociations.length > 0 && (
        <p className="rounded-md border border-amber-300 bg-amber-50 p-2 text-xs text-amber-700">
          {t('threatIntel.report.receipt.skipped')}: {result.skippedAssociations.join(', ')}
        </p>
      )}

      <p className="text-sm">
        {status === 'live' && entityId ? (
          <Button
            type="button"
            variant="link"
            size="sm"
            className="h-auto px-0 underline"
            onClick={() => onEntityView?.(entityId)}
          >
            {t('threatIntel.report.receipt.live')}
          </Button>
        ) : status === 'processing' ? (
          t('threatIntel.report.receipt.processing')
        ) : (
          t('threatIntel.report.receipt.pending')
        )}
      </p>

      <div className="flex justify-end">
        <Button type="button" variant="outline" size="sm" onClick={onReportAnother}>
          {t('threatIntel.report.receipt.another')}
        </Button>
      </div>
    </div>
  )
}

function AssociationTree({ associations }: { associations: ReportedAssociation[] }) {
  return (
    <ul className="space-y-0.5">
      {associations.map((association, index) => (
        <li key={`${association.type}:${association.value}:${index}`}>
          <span className="break-all">{association.value}</span>
          <span className="text-muted-foreground"> · {association.type}</span>
          {association.associations && association.associations.length > 0 && (
            <div className="ms-3 border-l border-border/60 pl-2">
              <AssociationTree associations={association.associations} />
            </div>
          )}
        </li>
      ))}
    </ul>
  )
}
