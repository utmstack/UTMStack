import { useTranslation } from 'react-i18next'
import { Crosshair } from 'lucide-react'
import { useTopTechniques } from '../hooks/use-overview'
import { fmtCount } from './helpers'
import { CardHeader } from './CardHeader'
import { SkeletonRows } from './SkeletonRows'
import { EmptyState } from './EmptyState'

export function MitreTechniquesCard() {
  const { t } = useTranslation()
  const { items, isLoading } = useTopTechniques()
  const max = Math.max(...items.map((m) => m.count), 1)
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader
        title={t('home.mitre.title')}
        icon={Crosshair}
        iconClass="text-rose-500"
        action={{ label: t('home.mitre.explore'), href: '/threat-management/alerts' }}
      />
      <div className="space-y-2.5 px-6 pb-5">
        {isLoading ? (
          <SkeletonRows rows={5} />
        ) : items.length === 0 ? (
          <EmptyState text={t('home.mitre.empty')} />
        ) : (
          items.map((m) => (
            <div key={m.value} className="grid grid-cols-[1fr_auto] items-center gap-3 text-xs">
              <div>
                <span className="font-medium">{m.value}</span>
                <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full rounded-full bg-gradient-to-r from-sky-400 to-blue-600"
                    style={{ width: `${(m.count / max) * 100}%` }}
                  />
                </div>
              </div>
              <div className="font-mono text-sm tabular-nums">{fmtCount(m.count)}</div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
