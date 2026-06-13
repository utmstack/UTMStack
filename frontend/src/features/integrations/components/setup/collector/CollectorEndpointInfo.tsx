import { useTranslation } from 'react-i18next'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

interface CollectorEndpointInfoProps {
  module: Integration
  port?: string
  sourceType?: string
}

export function CollectorEndpointInfo({
  module,
  port = '514/udp',
  sourceType,
}: CollectorEndpointInfoProps) {
  const { t } = useTranslation()

  const source = sourceType ?? module.moduleName?.toLowerCase() ?? module.name.toLowerCase().replace(/\s/g, '-')
  const host = typeof window !== 'undefined'
    ? window.location.host.includes(':')
      ? window.location.host.split(':')[0]
      : window.location.host
    : 'utmstack-collector.local'

  return (
    <Section title={t('integrations.setup.collector.endpoint.title')}>
      <p className="mb-2 text-xs text-muted-foreground">
        {t('integrations.setup.collector.endpoint.hint', { name: module.name })}
      </p>
      <CodeBlock
        lang="config"
        code={`host:        ${host}
port:        ${port}
source_type: ${source}`}
      />
    </Section>
  )
}
