import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import { useConectionKey } from '@/features/integrations/hooks/useConnectionKey'
import { buildAgentInstall, type AgentAction } from '@/features/integrations/utils/agentInstallBuilder'
import type { Integration } from '@/features/integrations/types'

interface AgentInstallSelectorProps {
  integration: Integration
}

export function AgentInstallSelector({ integration }: AgentInstallSelectorProps) {
  const { t } = useTranslation()
  const { key, isLoading } = useConectionKey()

  const host = window.location.host.includes(':')
    ? window.location.host.split(':')[0]
    : window.location.host


  const config = useMemo(
     () => buildAgentInstall(integration.name, { host, token: key.data?.connectionKey ?? '' }),
    [integration.id, host, key.data],
  )

  const [action, setAction] = useState<AgentAction>('install')
  const [platformId, setPlatformId] = useState<string>(config.platforms[0]?.id ?? '')

  if (config.platforms.length === 0) {
    return null
  }



  const activePlatformId = config.platforms.some(p => p.id === platformId)
    ? platformId
    : config.platforms[0].id

  const command = isLoading || !key.data
    ? t('integrations.setup.agent.loadingToken')
    : config.getCommand(action, activePlatformId)

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <label className="block">
          <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
            {t('integrations.setup.agent.actionLabel')}
          </span>
          <select
            value={action}
            onChange={(e) => setAction(e.target.value as AgentAction)}
            className="h-9 w-full appearance-none rounded-md border border-border bg-background px-3 text-sm"
          >
            {config.actions.map((a) => (
              <option key={a.id} value={a.id}>
                {t(a.labelKey)}
              </option>
            ))}
          </select>
        </label>

        <label className="block">
          <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
            {t('integrations.setup.agent.platformLabel')}
          </span>
          <select
            value={activePlatformId}
            onChange={(e) => setPlatformId(e.target.value)}
            className="h-9 w-full appearance-none rounded-md border border-border bg-background px-3 text-sm"
          >
            {config.platforms.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </label>
      </div>

      <CodeBlock code={command} />
    </div>
  )
}
