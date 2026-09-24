import { useEffect, useRef, useState } from 'react'
import { ChevronDown } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { PILL_BASE as PILL, ST_META } from '../lib/alert-meta'
import { STATUS_VALUE, type AlertTag, type StatusKey } from '../types/alert.types'
import { StatusObservationModal } from './status-observation-modal'
import { TagChip } from './tag-chip'

type Variant = 'pill' | 'action'

// Name of the system tag the backend applies when an alert is closed as a false positive.
const FALSE_POSITIVE_TAG = 'False positive'

type Pending = { status: string; fp: boolean; title: string }

export function StatusChangeMenu({
  status,
  onStatus,
  variant,
  tagCatalog = [],
}: {
  status: StatusKey
  onStatus: (status: string, observation: string, fp: boolean) => void
  variant: Variant
  tagCatalog?: AlertTag[]
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [pending, setPending] = useState<Pending | null>(null)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => ref.current && !ref.current.contains(e.target as Node) && setOpen(false)
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  // Statuses that require a human-written observation before committing.
  const pickStatus = (k: StatusKey) => {
    setOpen(false)
    if (k === 'completed') {
      setPending({ status: STATUS_VALUE.completed, fp: false, title: t('alerts.status.completed') })
      return
    }
    onStatus(STATUS_VALUE[k], '', false)
  }

  const pickFalsePositive = () => {
    setOpen(false)
    setPending({ status: STATUS_VALUE.completed, fp: true, title: t('alerts.drawer.completeFalsePositive') })
  }

  return (
    <div className="relative inline-block" ref={ref}>
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation()
          setOpen((v) => !v)
        }}
        className={
          variant === 'pill'
            ? cn(PILL, ST_META[status].pill)
            : 'inline-flex items-center gap-1 rounded-md border border-border bg-card px-2 py-1 hover:bg-muted'
        }
      >
        {variant === 'pill' ? (
          <>
            <span>{t(`alerts.status.${status}`)}</span>
            <ChevronDown size={10} />
          </>
        ) : (
          <>
            {t('alerts.drawer.setStatus')} <ChevronDown size={12} />
          </>
        )}
      </button>
      {open && (
        <div
          onClick={(e) => e.stopPropagation()}
          className="absolute left-0 top-full z-30 mt-1 w-max min-w-[12rem] rounded-md border border-border bg-popover py-1 shadow-lg"
        >
          {(['open', 'in_review', 'completed'] as StatusKey[]).map((k) => (
            <button
              key={k}
              onClick={() => pickStatus(k)}
              className="flex w-full items-center px-3 py-1.5 text-left hover:bg-muted"
            >
              <span className={cn(PILL, ST_META[k].pill)}>{t(`alerts.status.${k}`)}</span>
            </button>
          ))}
          <button
            onClick={pickFalsePositive}
            title={t('alerts.drawer.completeFalsePositive')}
            aria-label={t('alerts.drawer.completeFalsePositive')}
            className="flex w-full items-center gap-1.5 border-t border-border px-3 py-1.5 text-left hover:bg-muted"
          >
            <span className={cn(PILL, ST_META.completed.pill)}>{t('alerts.status.completed')}</span>
            <span className="text-xs text-muted-foreground">+</span>
            <TagChip name={FALSE_POSITIVE_TAG} catalog={tagCatalog} size="xs" />
          </button>
        </div>
      )}
      {pending && (
        <StatusObservationModal
          title={pending.title}
          onCancel={() => setPending(null)}
          onConfirm={(observation) => {
            onStatus(pending.status, observation, pending.fp)
            setPending(null)
          }}
        />
      )}
    </div>
  )
}
