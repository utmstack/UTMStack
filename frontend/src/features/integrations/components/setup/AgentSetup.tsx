import { useTranslation } from 'react-i18next'
import { Section } from '@/features/integrations/components/ui/Section'
import { AgentInstallSelector } from '@/features/integrations/components/setup/AgentInstallSelector'
import type { Integration } from '@/features/integrations/types'

interface AgentSetupProps {
  integration: Integration
}

export function AgentSetup({ integration: i }: AgentSetupProps) {
  const { t } = useTranslation()

  return (
    <div className="space-y-4">
      <Section title={t('integrations.setup.agent.howItWorksTitle')}>
        <p className="text-sm text-foreground/90">
          {t('integrations.setup.agent.howItWorksBody', { name: i.name })}
        </p>
      </Section>

      <Section title={t('integrations.setup.agent.installTitle', { name: i.name })}>
        <p className="mb-2 text-xs text-muted-foreground">
          {t('integrations.setup.agent.installHint', { name: i.name.toLowerCase() })}
        </p>
        <AgentInstallSelector integration={i} />
      </Section>

      <Section title={t('integrations.setup.agent.afterTitle')}>
        <ul className="list-disc space-y-1 pl-5 text-sm">
          <li>{t('integrations.setup.agent.after1')}</li>
          <li>{t('integrations.setup.agent.after2')}</li>
          <li>{t('integrations.setup.agent.after3')}</li>
        </ul>
      </Section>
    </div>
  )
}
