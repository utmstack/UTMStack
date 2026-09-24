import { Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useSocAiOptional } from '../SocAiProvider'
import { useSocAiConfigured } from '../lib/useSocAiConfig'

/**
 * "Ask" inside a drawer: opens the assistant beside it. The drawer has already
 * shared its item (see useSocAiFocus), so the question needs no id. Renders
 * nothing where there is no assistant or no model provider configured.
 */
export function SocAiAskAbout() {
  const { t } = useTranslation()
  const socAi = useSocAiOptional()
  const configured = useSocAiConfigured()

  if (!socAi || !configured) return null

  return (
    <button
      type="button"
      onClick={() => socAi.openPanel('panel')}
      title={t('socAi.chat.ask')}
      className="flex h-8 items-center gap-1.5 rounded-md px-2 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
    >
      <Sparkles size={14} strokeWidth={1.75} className="text-primary" />
      {t('socAi.chat.ask')}
    </button>
  )
}
