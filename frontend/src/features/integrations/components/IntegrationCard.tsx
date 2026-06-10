import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { LogoTile } from '@/features/integrations/components/ui/LogoTile'
import { KIND_META } from '@/features/integrations/constants'
import type { Integration } from '@/features/integrations/types'

interface IntegrationCardProps {
  integration: Integration
  onOpen: () => void
}

export function IntegrationCard({ integration: i, onOpen }: IntegrationCardProps) {
  const { t } = useTranslation()
  const km = KIND_META[i.kind]

  return (
    <button
      onClick={onOpen}
      className="group relative flex h-full flex-col overflow-hidden rounded-xl border border-border bg-card text-left transition-all hover:border-foreground/20 hover:shadow-md"
    >
      <div className="relative flex h-32 items-center justify-center border-b border-border bg-muted/40 p-6">
        <LogoTile src={i.logo} alt={i.name} darkInvert={i.darkInvert} size="md" />
        <span
          className={cn(
            'absolute right-3 top-3 inline-flex items-center gap-1 rounded-full bg-card/80 px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-border backdrop-blur',
            km?.tone??''
          )}
        >
          <span className={cn('inline-block h-1.5 w-1.5 rounded-full', km?.dot)} />
          {t(`integrations.kind.${i.kind}`)}
        </span>
      </div>

      <div className="flex flex-1 flex-col p-4">
        <h4 className="text-sm font-semibold leading-tight">{i.name}</h4>
        <p className="mt-1.5 line-clamp-4 flex-1 text-[11px] leading-relaxed text-muted-foreground">
          {i.description}
        </p>

        <div className="mt-3 flex items-center justify-between border-t border-border pt-3 text-[11px]">
          {i.status === 'configured' ? (
            <span className="inline-flex items-center gap-1 font-medium text-emerald-600 dark:text-emerald-400">
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
              {t('integrations.status.configured')}
              {i.rate ? (
                <span className="ml-1 font-mono tabular-nums text-muted-foreground">
                  · {t('integrations.card.eventsPerSec', { rate: i.rate.toLocaleString() })}
                </span>
              ) : null}
            </span>
          ) : (
            <span className="text-muted-foreground">{i.category}</span>
          )}
          <span className="text-primary opacity-0 transition-opacity group-hover:opacity-100">
            {i.status === 'configured' ? t('integrations.card.manage') : t('integrations.card.setup')}
          </span>
        </div>
      </div>
    </button>
  )
}
