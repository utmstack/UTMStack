import { useTranslation } from 'react-i18next'
import { Sparkles } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useSocAiOptional } from '../SocAiProvider'
import { useSocAiConfigured } from '../lib/useSocAiConfig'

/**
 * Topbar entry point to the assistant. It opens the side panel and does nothing
 * else — the composer lives in the panel, not here.
 *
 * Renders nothing outside the dashboard shell (the federation pages reuse the
 * Topbar without the provider) or until a model provider is configured.
 */
export function SocAiAskButton() {
  const { t } = useTranslation()
  const socAi = useSocAiOptional()
  const configured = useSocAiConfigured()

  if (!socAi || !configured) return null

  return (
    <button
      type="button"
      onClick={socAi.togglePanel}
      aria-label={t('socAi.chat.ask')}
      title={t('socAi.chat.ask')}
      aria-expanded={socAi.open}
      className={cn(
        'flex h-9 items-center gap-1.5 rounded-md px-2.5 text-sm transition-colors',
        socAi.open
          ? 'bg-muted text-foreground'
          : 'text-muted-foreground hover:bg-muted hover:text-foreground',
      )}
    >
      <Sparkles size={16} strokeWidth={1.75} className="text-primary" />
      <span className="hidden sm:inline">{t('socAi.chat.ask')}</span>
    </button>
  )
}
