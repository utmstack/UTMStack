import { useCommonResponse } from '@/features/soar/hooks/use-common-response'
import { Zap } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import type { Alert } from '../types/alert.types'

/**
 * "Run the common response" action for an alert: resolves the matching system
 * SOAR flow and triggers it on the alert's agent.
 *
 * `extended` (liveops right sidebar) renders the labelled button showing the
 * flow name; the reduced form (alerts table row) is a small icon button whose
 * tooltip carries the flow name.
 */
export function AlertResponseButton({ alert, extended }: { alert: Alert; extended?: boolean }) {
  const { t } = useTranslation()
  const { flow, loading, trigger } = useCommonResponse(alert)

  const run = async () => {
    if (!flow) return
    try {
      await trigger(flow.relPath, alert)
      toast.success(t('liveOps.responseDone'))
    } catch {
      toast.error(t('liveOps.responseError'))
    }
  }

  const label = loading
    ? t('liveOps.responseFinding')
    : flow
      ? flow.name
      : t('liveOps.responseNone')
  const disabled = !flow || loading

  if (!extended) {
    return (
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation()
          run()
        }}
        disabled={disabled}
        title={label}
        aria-label={label}
        className="mx-auto flex h-7 w-7 items-center justify-center rounded text-muted-foreground/60 transition hover:bg-background hover:text-primary disabled:pointer-events-none disabled:opacity-40"
      >
        <Zap size={13} />
      </button>
    )
  }

  return (
    <button
      type="button"
      disabled={disabled}
      onClick={run}
      title={flow ? flow.name : t('liveOps.responseNone')}
      className="flex w-full items-center gap-1.5 rounded-lg border border-primary/40 bg-primary/10 px-2.5 py-1.5 text-left text-[12px] font-medium text-foreground hover:bg-primary/20 disabled:opacity-40"
    >
      <Zap size={12} className="text-primary" />
      {label}
    </button>
  )
}
