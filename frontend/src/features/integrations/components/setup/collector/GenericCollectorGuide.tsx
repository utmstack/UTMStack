import { useTranslation } from 'react-i18next'
import { ExternalLink } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

interface GenericCollectorGuideProps {
  integration: Integration
}

export function GenericCollectorGuide({ integration: i }: GenericCollectorGuideProps) {
  const { t } = useTranslation()

  return (
    <div className="space-y-4">
      <Section title={t('integrations.setup.collector.howItWorksTitle')}>
        <p className="text-sm text-foreground/90">
          {t('integrations.setup.collector.howItWorksBody', { name: i.name })}
        </p>
      </Section>

      <Section title={t('integrations.setup.collector.step1Title')}>
        <p className="mb-2 text-xs text-muted-foreground">
          {t('integrations.setup.collector.step1Hint', { name: i.name })}
        </p>
        <CodeBlock
          code={`curl -fsSL https://install.utmstack.com/collector.sh | sudo bash -s -- \\
  --token=eyJ...workspace_id=acme...`}
        />
      </Section>

      <Section title={t('integrations.setup.collector.step2Title', { name: i.name })}>
        <p className="mb-2 text-xs text-muted-foreground">
          {t('integrations.setup.collector.step2Hint', { name: i.name, port: i.defaultPort ?? '514/udp' })}
        </p>
        <CodeBlock
          code={`# Example syslog target on ${i.name}:
host:        utmstack-collector.local
port:        ${i.defaultPort ?? '514/udp'}
format:      RFC 5424
facility:    local0`}
        />
        <a className="mt-2 inline-flex items-center gap-1 text-[11px] text-primary hover:underline" href="#">
          {t('integrations.setup.collector.vendorDocs')} <ExternalLink size={10} />
        </a>
      </Section>

      <Section title={t('integrations.setup.collector.step3Title')}>
        <p className="text-sm text-foreground/90">
          {t('integrations.setup.collector.step3Body', {
            name: i.name,
            source: i.name.toLowerCase().replace(/\s/g, '-'),
          })}
        </p>
      </Section>
    </div>
  )
}
