import { useCallback, useMemo, useState } from 'react'
import { ArrowLeft, ArrowRight, Loader2, X } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import type { Alert } from '@/features/alerts/types/alert.types'
import { STATUS_BY_VALUE, STATUS_VALUE } from '@/features/alerts/types/alert.types'
import { incidentsHttpService, IncidentsHttpError } from '../services/incidents-http.service'
import type { AlertLinkItem, CreateIncidentInput } from '../types/incident.types'
import { CreateIncidentStepDetails } from './create-incident-step-details'
import { CreateIncidentStepAlerts } from './create-incident-step-alerts'

export function CreateIncidentWizard({ onClose, onCreated }: { onClose: () => void; onCreated: () => void }) {
  const { t } = useTranslation()
  const [step, setStep] = useState<1 | 2>(1)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [alertsById, setAlertsById] = useState<Map<string, Alert>>(new Map())
  const [busy, setBusy] = useState(false)

  const onAlertsChange = useCallback((alerts: Alert[]) => {
    setAlertsById((prev) => {
      const next = new Map(prev)
      alerts.forEach((a) => next.set(a.id, a))
      return next
    })
  }, [])

  const toggle = (id: string) =>
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })

  const toggleAll = (page: Alert[]) =>
    setSelected((prev) => {
      const next = new Set(prev)
      const allOn = page.length > 0 && page.every((a) => next.has(a.id))
      if (allOn) page.forEach((a) => next.delete(a.id))
      else page.forEach((a) => next.add(a.id))
      return next
    })

  const alertList = useMemo<AlertLinkItem[]>(
    () =>
      [...selected]
        .map((id) => alertsById.get(id))
        .filter((a): a is Alert => !!a)
        .map((a) => ({
          alertId: a.id,
          alertName: a.name || a.id,
          alertSeverity: a.severity ?? 'low',
          alertStatus: STATUS_VALUE[STATUS_BY_VALUE[a.status ?? ''] ?? 'open'],
        })),
    [selected, alertsById],
  )

  const trimmedName = name.trim()
  const canProceed = trimmedName.length > 0
  const canSubmit = canProceed && selected.size > 0 && !busy

  const payload = useMemo<CreateIncidentInput>(
    () => ({
      incidentName: trimmedName,
      incidentDescription: description.trim() || undefined,
      alertList,
    }),
    [trimmedName, description, alertList],
  )

  const submit = async () => {
    if (!canSubmit) return
    setBusy(true)
    try {
      await incidentsHttpService.create(payload)
      toast.success(t('incidents.create.success'))
      onCreated()
    } catch (e) {
      toast.error(e instanceof IncidentsHttpError ? e.message : t('incidents.create.error'))
      setBusy(false)
    }
  }

  return (
    <div className="flex h-full min-h-0 w-full flex-col px-6 pb-6 pt-3">
      <header className="flex items-center justify-between border-b border-border pb-3">
        <div className="flex items-center gap-3">
          <h1 className="text-base font-semibold">{t('incidents.create.title')}</h1>
          <span className="text-xs text-muted-foreground">
            {t('incidents.create.stepIndicator', { current: step, total: 2 })}
          </span>
        </div>
        <button
          onClick={onClose}
          className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          aria-label={t('common.actions.close')}
        >
          <X size={16} />
        </button>
      </header>

      <div className="mt-4 flex min-h-0 flex-1 flex-col">
        {step === 1 ? (
          <div className="mx-auto w-full max-w-xl">
            <CreateIncidentStepDetails
              name={name}
              description={description}
              onChangeName={setName}
              onChangeDescription={setDescription}
            />
          </div>
        ) : (
          <CreateIncidentStepAlerts
            selected={selected}
            onToggle={toggle}
            onToggleAll={toggleAll}
            onAlertsChange={onAlertsChange}
          />
        )}
      </div>

      <footer className="mt-4 flex shrink-0 items-center justify-between border-t border-border pt-3">
        <div className="text-xs text-muted-foreground">
          {step === 2 && selected.size > 0 && t('incidents.create.selectedCount', { count: selected.size })}
        </div>
        <div className="flex items-center gap-2">
          {step === 1 ? (
            <>
              <Button variant="ghost" size="sm" onClick={onClose}>
                {t('common.actions.cancel')}
              </Button>
              <Button size="sm" disabled={!canProceed} onClick={() => setStep(2)}>
                {t('common.actions.next')}
                <ArrowRight size={14} className="ml-1.5" />
              </Button>
            </>
          ) : (
            <>
              <Button variant="ghost" size="sm" onClick={() => setStep(1)} disabled={busy}>
                <ArrowLeft size={14} className="mr-1.5" />
                {t('common.actions.back')}
              </Button>
              <Button size="sm" disabled={!canSubmit} onClick={submit}>
                {busy && <Loader2 size={14} className="mr-1.5 animate-spin" />}
                {t('incidents.create.submit')}
              </Button>
            </>
          )}
        </div>
      </footer>
    </div>
  )
}
