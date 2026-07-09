import { useTranslation } from 'react-i18next'
import { Activity } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useSystemHealth, type HealthStatus } from '../hooks/use-overview'
import { CardHeader } from './CardHeader'
import { SkeletonRows } from './SkeletonRows'

const HEALTH_META: Record<HealthStatus, { dot: string; badge: string }> = {
  up: { dot: 'bg-emerald-500', badge: 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-300' },
  degraded: { dot: 'bg-amber-500', badge: 'bg-amber-500/15 text-amber-600 dark:text-amber-300' },
  down: { dot: 'bg-red-500', badge: 'bg-red-500/15 text-red-500 dark:text-red-300' },
  unknown: { dot: 'bg-muted-foreground/50', badge: 'bg-muted text-muted-foreground' },
}

export function SystemHealthCard() {
  const { t } = useTranslation()
  const { services, isLoading } = useSystemHealth()
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title={t('home.health.title')} icon={Activity} iconClass="text-emerald-500" action={{ label: t('home.health.details'), href: '/settings/about' }} />
      <div className="divide-y divide-border">
        {isLoading ? (
          <div className="px-6 py-4">
            <SkeletonRows rows={3} />
          </div>
        ) : (
          services.map((p) => {
            const meta = HEALTH_META[p.status]
            return (
              <div key={p.name} className="flex items-center justify-between px-6 py-2.5 text-xs">
                <div className="flex items-center gap-2">
                  <span className={cn('h-2 w-2 rounded-full', meta.dot)} />
                  <span>{p.name}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-mono text-[10px] text-muted-foreground">{p.detail}</span>
                  <span className={cn('rounded-md px-1.5 py-0.5 text-[10px] font-medium leading-none', meta.badge)}>
                    {t(`about.status.${p.status}`)}
                  </span>
                </div>
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
