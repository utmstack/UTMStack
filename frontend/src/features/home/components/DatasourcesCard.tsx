import { useTranslation } from 'react-i18next'
import { Database, TrendingUp } from 'lucide-react'
import { useDatasourcesOverview } from '../hooks/use-overview'
import { fmtCount, spanLabel } from './helpers'
import { CardHeader } from './CardHeader'
import { EventsChart } from './EventsChart'

export function DatasourcesCard() {
  const { t } = useTranslation()
  const ds = useDatasourcesOverview()
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title={t('home.datasources.title')} icon={Database} action={{ label: t('home.datasources.viewAll'), href: '/datasources' }} />
      <div className="flex flex-wrap items-end gap-x-10 gap-y-4 px-6 pt-2">
        <div>
          {ds.isLoading ? (
            <div className="h-12 w-20 animate-pulse rounded bg-muted" />
          ) : (
            <div className="text-5xl font-semibold leading-none tracking-tight text-emerald-500 tabular-nums dark:text-emerald-400">
              {fmtCount(ds.activeSources)}
            </div>
          )}
          <div className="mt-2 text-xs text-muted-foreground">
            <span className="text-foreground">{t('home.datasources.ofTotal', { n: fmtCount(ds.total) })}</span>{' '}
            {t('home.datasources.datasources')}
            <br />
            {t('home.datasources.sendingData')}
          </div>
        </div>
        <div>
          <div className="flex items-baseline gap-2">
            {ds.isLoading ? (
              <div className="h-12 w-28 animate-pulse rounded bg-muted" />
            ) : (
              <span className="text-5xl font-semibold leading-none tracking-tight tabular-nums">
                {fmtCount(ds.events)}
              </span>
            )}
            <TrendingUp size={20} strokeWidth={2} className="text-emerald-500 dark:text-emerald-400" />
          </div>
          <div className="mt-2 text-xs text-muted-foreground">
            {t('home.datasources.totalEvents')}
            <br />
            {spanLabel(t, ds.from, ds.to)}
          </div>
        </div>
      </div>
      <div className="px-2 pb-2 pt-4">
        <EventsChart points={ds.points} loading={ds.isLoading} />
      </div>
    </div>
  )
}
