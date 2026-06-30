import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { flagEmoji } from '../lib/alert-meta'
import type { Side } from '../types/alert.types'
import { Row } from './row'

export function PartyCard({ title, ep, accent }: { title: string; ep?: Side; accent?: boolean }) {
  const { t } = useTranslation()
  return (
    <div className={cn('rounded-lg border bg-card p-4', accent ? 'border-red-500/30' : 'border-border')}>
      <h4 className="mb-3 text-sm font-semibold">{title}</h4>
      {!ep ? (
        <p className="text-xs text-muted-foreground">{t('alerts.party.noData')}</p>
      ) : (
        <dl className="grid grid-cols-[80px_1fr] gap-y-2 text-xs">
          {ep.ip && <Row k={t('alerts.party.ip')}><span className="font-mono">{ep.ip}</span></Row>}
          {ep.host && <Row k={t('alerts.party.host')}><span className="font-mono">{ep.host}</span></Row>}
          {ep.user && <Row k={t('alerts.party.user')}><span className="font-mono">{ep.user}</span></Row>}
          {ep.domain && <Row k={t('alerts.party.domain')}><span className="font-mono">{ep.domain}</span></Row>}
          {ep.geolocation?.country && (
            <Row k={t('alerts.party.country')}>
              {flagEmoji(ep.geolocation.countryCode)} {ep.geolocation.country}
              {ep.geolocation.city ? ` · ${ep.geolocation.city}` : ''}
            </Row>
          )}
        </dl>
      )}
    </div>
  )
}
