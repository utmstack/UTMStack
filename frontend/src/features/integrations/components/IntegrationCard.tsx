import { useTranslation } from 'react-i18next'
import { Puzzle } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useThemeContext } from '@/app/providers/ThemeProvider'
import { ingestMeta } from '@/features/integrations/constants'
import type { Integration } from '@/features/integrations/types'

interface IntegrationCardProps {
  integration: Integration
  onOpen: () => void
}

export function IntegrationCard({ integration: i, onOpen }: IntegrationCardProps) {
  const { t } = useTranslation()
  const { theme } = useThemeContext()
  const configured = i.status === 'configured'
  const logo = theme === 'dark' && i.logoDark ? i.logoDark : i.logo

  return (
    <button
      onClick={onOpen}
      className="group relative flex h-[300px] w-full flex-col items-center justify-between rounded-lg border border-border bg-card p-5 text-center shadow-sm transition-all hover:border-foreground/20 hover:shadow-md"
    >
      {i.ingestType && (
        <span
          className={cn(
            'absolute left-2.5 top-2.5 rounded px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide ring-1',
            ingestMeta(i.ingestType).pill,
          )}
        >
          {ingestMeta(i.ingestType).label}
        </span>
      )}
      <div className="flex w-full flex-1 flex-col items-center justify-center overflow-hidden">
        {/* Logo — bare, centered, large (mirrors the legacy module card). */}
        <div className="flex h-24 w-full shrink-0 items-center justify-center px-3 py-2">
          <img
            src={logo}
            alt={i.name}
            loading="lazy"
            className="h-full w-auto max-w-[80%] object-contain"
          />
        </div>

        <h6 className="mt-3 line-clamp-2 w-full text-base font-semibold leading-tight">{i.name}</h6>

        {i.description ? (
          <p className="mt-2 line-clamp-4 w-full text-xs leading-relaxed text-muted-foreground">
            {i.description}
          </p>
        ) : (
          <p className="mt-2 w-full text-xs text-muted-foreground/70">{i.category}</p>
        )}
      </div>

      {/* Status button (visual; the whole card opens the drawer). */}
      <span
        className={cn(
          'mt-3 inline-flex shrink-0 items-center gap-1.5 rounded-md px-4 py-1.5 text-xs font-semibold transition-colors',
          configured
            ? 'bg-emerald-500/15 text-emerald-600 group-hover:bg-emerald-500/25 dark:text-emerald-400'
            : 'bg-primary/10 text-primary group-hover:bg-primary/20',
        )}
      >
        <Puzzle size={13} />
        {configured ? t('integrations.card.enabled') : t('integrations.card.setupShort')}
        {configured && i.rate ? (
          <span className="ml-1 font-mono tabular-nums opacity-80">
            · {t('integrations.card.eventsPerSec', { rate: i.rate.toLocaleString() })}
          </span>
        ) : null}
      </span>
    </button>
  )
}
