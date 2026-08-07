import { useState } from 'react'
import { Flame, Loader2, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { SELECT_CLS } from '../lib/alert-meta'
import { useIncidentLink, type IncidentMode } from '../hooks/use-incident-link'
import type { Alert } from '../types/alert.types'

export function AlertIncidentModal({
  alerts,
  onClose,
  onDone,
}: {
  alerts: Alert[]
  onClose: () => void
  onDone: () => void
}) {
  const { t } = useTranslation()
  const [mode, setMode] = useState<IncidentMode>('new')
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [incidentId, setIncidentId] = useState('')
  const { incidents, loadingIncidents, busy, submit } = useIncidentLink({ alerts, mode, onDone })

  const canSubmit = mode === 'new' ? !!name.trim() : !!incidentId
  const onSubmit = () => {
    if (!canSubmit) return
    void submit({ name, description, incidentId })
  }

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-md flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between gap-4 border-b border-border px-5 py-4">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <Flame size={17} className="text-red-500" />
            {t('alerts.incident.titleAdd', { count: alerts.length })}
          </h2>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="flex items-center gap-1 border-b border-border px-5 pt-2">
          {(['new', 'existing'] as const).map((m) => (
            <button
              key={m}
              onClick={() => setMode(m)}
              className={cn(
                'relative px-3 py-2 text-xs transition-colors',
                mode === m ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
              )}
            >
              {m === 'new' ? t('alerts.incident.new') : t('alerts.incident.existing')}
              {mode === m && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
            </button>
          ))}
        </div>

        <div className="space-y-3 px-5 py-4">
          {mode === 'new' ? (
            <>
              <div>
                <label className="mb-1 block text-xs font-medium text-foreground/80">
                  {t('alerts.incident.nameLabel')}
                </label>
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder={t('alerts.incident.namePlaceholder')}
                  autoFocus
                />
              </div>
              <div>
                <label className="mb-1 block text-xs font-medium text-foreground/80">
                  {t('alerts.incident.descriptionLabel')}
                </label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  rows={3}
                  placeholder={t('alerts.incident.descriptionPlaceholder')}
                  className="w-full rounded-md border border-input bg-background/40 p-2 text-xs focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
              </div>
            </>
          ) : (
            <div>
              <label className="mb-1 block text-xs font-medium text-foreground/80">{t('alerts.incident.pickLabel')}</label>
              {loadingIncidents ? (
                <div className="flex items-center gap-2 py-2 text-xs text-muted-foreground">
                  <Loader2 className="h-3.5 w-3.5 animate-spin" /> {t('alerts.incident.loading')}
                </div>
              ) : (
                <select
                  value={incidentId}
                  onChange={(e) => setIncidentId(e.target.value)}
                  className={cn(SELECT_CLS, 'w-full')}
                >
                  <option value="">{t('alerts.incident.selectPlaceholder')}</option>
                  {incidents.map((i) => (
                    <option key={i.id} value={i.id}>
                      {i.incidentName}
                    </option>
                  ))}
                </select>
              )}
            </div>
          )}
          <p className="text-[11px] text-muted-foreground">{t('alerts.incident.willLink', { count: alerts.length })}</p>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-5 py-3">
          <Button variant="outline" size="sm" onClick={onClose} disabled={busy}>
            {t('alerts.incident.cancel')}
          </Button>
          <Button size="sm" disabled={!canSubmit || busy} onClick={onSubmit}>
            {busy ? '…' : mode === 'new' ? t('alerts.incident.create') : t('alerts.incident.add')}
          </Button>
        </footer>
      </div>
    </div>
  )
}
