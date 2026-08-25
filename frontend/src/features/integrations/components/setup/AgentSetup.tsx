import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { AgentInstallSelector } from '@/features/integrations/components/setup/AgentInstallSelector'
import { AgentUninstallSection } from '@/features/integrations/components/setup/AgentUninstallSection'
import { useConectionKey } from '@/features/integrations/hooks/useConnectionKey'
import { buildAgentInstall } from '@/features/integrations/utils/agentInstallBuilder'
import { forwarderHost } from '@/features/integrations/components/setup/collector/ForwarderGuide'
import type { Integration } from '@/features/integrations/types'

interface AgentSetupProps {
  integration: Integration
}

const AGENT_PORTS = ['443']

export function AgentSetup({ integration: i }: AgentSetupProps) {
  const { t } = useTranslation()
  const { key } = useConectionKey()

  const host = forwarderHost()

  // Always render the command even before the key loads: fall back to a masked
  // placeholder (like the AS/400 and Forwarder guides) so the user never sees a
  // bare "loading…" state. The token is masked in the UI either way.
  const token = key.data?.connectionKey ?? '*'.repeat(7)

  const config = useMemo(
    () => buildAgentInstall(i.name, { host, token }),
    [i.id, host, token],
  )

  // Platform selection is shared between the install command and the uninstall
  // section, so it lives here.
  const [platformId, setPlatformId] = useState<string>(config.platforms[0]?.id ?? '')
  const activePlatformId = config.platforms.some((p) => p.id === platformId)
    ? platformId
    : config.platforms[0]?.id ?? ''

  const installCommand = config.getCommand('install', activePlatformId)
  const uninstallCommand = config.getCommand('uninstall', activePlatformId)

  // macOS ships an Apple Silicon (arm64) build only — surface the compatibility
  // requirement so Intel-Mac users aren't left guessing.
  const isMac =
    (i.moduleName ?? '').toUpperCase() === 'MACOS' || i.name.toLowerCase().startsWith('macos')

  // Windows commands are PowerShell, the rest are bash — highlight accordingly.
  const isWindows =
    (i.moduleName ?? '').toUpperCase() === 'WINDOWS_AGENT' || i.name.toLowerCase().startsWith('windows')
  const lang = isWindows ? 'powershell' : 'bash'

  return (
    <div className="space-y-4">
      <Section title={t('integrations.setup.agent.howItWorksTitle')}>
        <p className="text-sm text-foreground/90">
          {t('integrations.setup.agent.howItWorksBody', { name: i.name })}
        </p>
      </Section>

      {/* Step 1 (legacy parity): pre-installation requirements — the agent talks
          to UTMStack over 443, the port their network already allows. */}
      <Section title={t('integrations.setup.agent.prereqTitle')} step={1}>
        <p className="mb-2 text-sm text-foreground/90">
          {t('integrations.setup.agent.prereqBody', { ports: AGENT_PORTS.join(', ') })}
        </p>
        <div className="flex flex-wrap gap-2">
          {AGENT_PORTS.map((p) => (
            <span
              key={p}
              className="rounded-md bg-muted px-2.5 py-1 font-mono text-xs font-medium text-foreground/90 ring-1 ring-inset ring-border"
            >
              {p}
            </span>
          ))}
        </div>
        {isMac && (
          <p className="mt-3 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
            {t('integrations.setup.agent.macCompat')}
          </p>
        )}
      </Section>

      {/* Step 2: install command, with platform tabs and the sensitive-info warning. */}
      <Section title={t('integrations.setup.agent.installTitle', { name: i.name })} step={2}>
        <p className="mb-2 text-xs text-muted-foreground">
          {t('integrations.setup.agent.installHint', { name: i.name.toLowerCase() })}
        </p>
        <div className="mb-3 flex gap-2 rounded-md border border-amber-500/30 bg-amber-500/[0.07] px-3 py-2 text-[11px] text-amber-700 dark:text-amber-300">
          <AlertTriangle size={14} className="mt-0.5 shrink-0" />
          <span>{t('integrations.setup.agent.sensitiveWarning')}</span>
        </div>
        <AgentInstallSelector
          config={config}
          activePlatformId={activePlatformId}
          onSelect={setPlatformId}
          command={installCommand}
          lang={lang}
        />
      </Section>

      <Section title={t('integrations.setup.agent.afterTitle')}>
        <ul className="list-disc space-y-1 pl-5 text-sm">
          <li>{t('integrations.setup.agent.after1')}</li>
          <li>{t('integrations.setup.agent.after2')}</li>
          <li>{t('integrations.setup.agent.after3')}</li>
        </ul>
      </Section>

      {/* Uninstall — collapsed by default, at the very end (like the Forwarder guide). */}
      <AgentUninstallSection command={uninstallCommand} lang={lang} />
    </div>
  )
}
