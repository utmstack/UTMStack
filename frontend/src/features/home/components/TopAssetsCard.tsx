import { useTranslation } from 'react-i18next'
import { Server } from 'lucide-react'
import { useTopAssets } from '../hooks/use-overview'
import { fmtCount } from './helpers'
import { CardHeader } from './CardHeader'
import { SkeletonRows } from './SkeletonRows'
import { EmptyState } from './EmptyState'

export function TopAssetsCard() {
  const { t } = useTranslation()
  const { items, isLoading } = useTopAssets()
  const max = Math.max(...items.map((a) => a.count), 1)
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title={t('home.assets.title')} icon={Server} action={{ label: t('home.assets.viewAll'), href: '/datasources' }} />
      <div className="grid grid-cols-[1fr_70px] gap-3 px-6 pb-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <div>{t('home.assets.asset')}</div>
        <div className="text-right">{t('home.assets.alerts')}</div>
      </div>
      <div className="border-t border-border">
        {isLoading ? (
          <div className="px-6 py-4">
            <SkeletonRows rows={5} />
          </div>
        ) : items.length === 0 ? (
          <div className="px-6 py-6">
            <EmptyState text={t('home.assets.empty')} />
          </div>
        ) : (
          items.map((a) => (
            <div
              key={a.value}
              className="grid grid-cols-[1fr_70px] items-center gap-3 border-b border-border px-6 py-2.5 text-xs last:border-b-0 hover:bg-muted/40"
            >
              <div className="min-w-0">
                <div className="truncate font-mono text-[12px]">{a.value}</div>
                <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-muted">
                  <div
                    className="h-full rounded-full bg-gradient-to-r from-rose-400 to-red-600"
                    style={{ width: `${(a.count / max) * 100}%` }}
                  />
                </div>
              </div>
              <div className="text-right font-mono tabular-nums">{fmtCount(a.count)}</div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
