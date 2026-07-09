import { useTranslation } from 'react-i18next'
import { BadgeCheck } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useComplianceScores } from '../hooks/use-overview'
import { CardHeader } from './CardHeader'
import { EmptyState } from './EmptyState'

export function ComplianceCard() {
  const { t } = useTranslation()
  const { items, isLoading } = useComplianceScores()
  return (
    <div className="h-full rounded-xl border border-border bg-card">
      <CardHeader title={t('home.compliance.title')} icon={BadgeCheck} iconClass="text-emerald-500" action={{ label: t('home.compliance.reports'), href: '/compliance' }} />
      <div className="px-6 pb-6">
        {isLoading ? (
          <div className="h-32 animate-pulse rounded-lg bg-muted" />
        ) : items.length === 0 ? (
          <EmptyState text={t('home.compliance.empty')} />
        ) : (
          <>
            <div
              className="grid items-end gap-3 h-32"
              style={{ gridTemplateColumns: `repeat(${items.length}, minmax(0, 1fr))` }}
            >
              {items.map((c) => (
                <div key={c.key} className="flex h-full flex-col items-center justify-end">
                  <div className="mb-1 font-mono text-[10px] tabular-nums">{c.score}%</div>
                  <div
                    className={cn(
                      'w-full rounded-t-md',
                      c.score >= 90
                        ? 'bg-gradient-to-t from-emerald-600 to-emerald-400'
                        : c.score >= 80
                          ? 'bg-gradient-to-t from-sky-600 to-sky-400'
                          : c.score >= 50
                            ? 'bg-gradient-to-t from-amber-600 to-amber-400'
                            : 'bg-gradient-to-t from-red-600 to-red-400',
                    )}
                    style={{ height: `${Math.max(c.score, 2)}%` }}
                  />
                </div>
              ))}
            </div>
            <div
              className="mt-2 grid gap-3 text-center text-[10px] uppercase tracking-wider text-muted-foreground"
              style={{ gridTemplateColumns: `repeat(${items.length}, minmax(0, 1fr))` }}
            >
              {items.map((c) => (
                <div key={c.key} className="truncate" title={c.name}>
                  {c.name}
                </div>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  )
}
