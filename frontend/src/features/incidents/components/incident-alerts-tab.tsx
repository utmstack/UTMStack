import { useState } from 'react'
import { ChevronRight, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { ConfirmDialog } from '@/shared/components/ui/confirm-dialog'
import { sevKey } from '../lib/incident-meta'
import { useIncidentAlertsTab } from '../hooks/use-incident-alerts-tab'
import { TabEmpty, TabError, TabLoader } from './ui-primitives'
import { IncidentAlertPreview } from './incident-alert-preview'

export function IncidentAlertsTab({ incidentId, onChanged }: { incidentId: string; onChanged: () => void }) {
  const { t } = useTranslation()
  const { rows, error, reload, remove, pendingDelete, deleting, cancelDelete, confirmDelete } = useIncidentAlertsTab(incidentId, onChanged)
  // One open at a time. The row used to be a link to the alerts page, which
  // meant leaving the incident to read each of its own alerts.
  const [openId, setOpenId] = useState<string | null>(null)

  if (error) return <TabError onRetry={reload} />
  if (rows === null) return <TabLoader />
  if (rows.length === 0) return <TabEmpty>{t('incidents.alerts.empty')}</TabEmpty>
  return (
    <>
      <div className="overflow-hidden rounded-lg border border-border bg-card">
        {rows.map((a) => (
          <div key={a.id} className="border-b border-border/60 last:border-b-0">
            <div
              onClick={() => setOpenId((cur) => (cur === a.alertId ? null : a.alertId))}
              className="group flex cursor-pointer items-center gap-3 px-4 py-2.5 text-xs hover:bg-muted/40"
            >
              <ChevronRight
                size={13}
                className={cn(
                  'shrink-0 text-muted-foreground/60 transition-transform',
                  openId === a.alertId && 'rotate-90 text-foreground'
                )}
              />
              <span
                className={cn(
                  'h-4 w-1 shrink-0 rounded-full',
                  a.alertSeverity === 'high' ? 'bg-red-500' : a.alertSeverity === 'medium' ? 'bg-amber-500' : 'bg-sky-500'
                )}
              />
              <span className="min-w-0 flex-1 truncate font-medium" title={a.alertName}>
                {a.alertName}
              </span>
              <span className="shrink-0 font-mono text-[10px] text-muted-foreground">
                {t(`incidents.sev.${sevKey(a.alertSeverity)}`)}
              </span>
              <button
                onClick={(e) => {
                  // The row toggles; the bin must not do both.
                  e.stopPropagation()
                  remove(a)
                }}
                title={t('incidents.alerts.remove')}
                className="shrink-0 rounded p-1 text-muted-foreground opacity-0 transition hover:bg-muted hover:text-red-500 group-hover:opacity-100"
              >
                <Trash2 size={13} />
              </button>
            </div>
            {openId === a.alertId && <IncidentAlertPreview alertId={a.alertId} />}
          </div>
        ))}
      </div>
      <ConfirmDialog
        open={pendingDelete != null}
        title={t('incidents.alerts.remove') ?? 'Remove alert'}
        body={pendingDelete ? t('incidents.alerts.removeConfirm', { name: pendingDelete.alertName }) : ''}
        confirmLabel={t('common.actions.remove') ?? t('incidents.alerts.remove') ?? undefined}
        danger
        busy={deleting}
        onClose={() => !deleting && cancelDelete()}
        onConfirm={confirmDelete}
      />
    </>
  )
}
