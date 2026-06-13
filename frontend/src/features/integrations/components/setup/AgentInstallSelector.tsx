import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { AgentInstallConfig } from '@/features/integrations/utils/agentInstallBuilder'

interface AgentInstallSelectorProps {
  config: AgentInstallConfig
  activePlatformId: string
  onSelect: (id: string) => void
  /** Resolved install command for the active platform (or a loading placeholder). */
  command: string
  /** Shell the command runs in — drives syntax highlighting (Windows = powershell). */
  lang?: 'bash' | 'powershell'
}

export function AgentInstallSelector({ config, activePlatformId, onSelect, command, lang = 'bash' }: AgentInstallSelectorProps) {
  const { t } = useTranslation()

  if (config.platforms.length === 0) {
    return null
  }

  return (
    <div className="space-y-3">
      <div>
        <span className="mb-1.5 block text-[11px] font-medium text-muted-foreground">
          {t('integrations.setup.agent.platformLabel')}
        </span>
        <div className="flex flex-wrap gap-1.5">
          {config.platforms.map((p) => (
            <button
              key={p.id}
              type="button"
              onClick={() => onSelect(p.id)}
              className={cn(
                'rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
                p.id === activePlatformId
                  ? 'bg-primary text-primary-foreground shadow-sm'
                  : 'bg-muted text-muted-foreground hover:bg-muted/70 hover:text-foreground',
              )}
            >
              {p.name}
            </button>
          ))}
        </div>
      </div>

      <CodeBlock code={command} lang={lang} />
    </div>
  )
}
