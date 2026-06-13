import { useMemo } from 'react'
import { FileSearch, Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { LogResults } from '@/features/log-explorer/components/log-results'
import type { LogDocument } from '@/features/log-explorer/types/log-explorer.types'
import type { AlertEventItem } from '../types/alert.types'

/**
 * The events the engine attached inline to the alert (a sample, capped at ~11),
 * rendered with the exact same table + Fields/JSON detail as the Log Explorer.
 * The "View all related logs" button reproduces the engine's correlation search
 * (uncapped) in the Log Explorer.
 */
export function AlertRelatedEvents({
  events,
  onViewAll,
  loadingAll,
}: {
  events: AlertEventItem[]
  onViewAll: () => void
  loadingAll: boolean
}) {
  const { t } = useTranslation()
  // Each stored event IS a log document; normalise its time field to @timestamp
  // (it's stored as `timestamp`) so the shared renderer shows it.
  const docs = useMemo<LogDocument[]>(
    () =>
      events.map((e) => ({
        ...(e as Record<string, unknown>),
        '@timestamp': e['@timestamp'] || e.timestamp || e.deviceTime || '',
      })),
    [events]
  )

  return (
    <div className="space-y-3">
      <div className="flex items-start justify-between gap-3">
        <p className="text-[11px] leading-relaxed text-muted-foreground">
          {events.length > 0 ? t('alerts.related.sample') : t('alerts.related.empty')}
        </p>
        <Button size="sm" variant="outline" onClick={onViewAll} disabled={loadingAll} className="shrink-0">
          {loadingAll ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : <FileSearch size={13} className="mr-1.5" />}
          {t('alerts.related.viewAll')}
        </Button>
      </div>
      {events.length > 0 && <LogResults docs={docs} />}
    </div>
  )
}
