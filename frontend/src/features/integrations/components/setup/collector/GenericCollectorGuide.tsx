import { useTranslation } from 'react-i18next'
import { ExternalLink, Forward, Server, ShieldCheck } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import { FlowNode, FlowEdge } from '@/shared/components/ui/flow-diagram'
import type { Integration } from '@/features/integrations/types'

interface GenericCollectorGuideProps {
  integration: Integration
}

const SHARED = 'integrations.setup.collector.forwarder'

function GenericFlowDiagram({ source, port }: { source: string; port: string }) {
  const { t } = useTranslation()
  return (
    <div className="mt-3 flex items-stretch justify-center gap-1 sm:gap-2">
      <FlowNode
        icon={<Server size={20} className="text-muted-foreground" />}
        title={source}
        sub={t(`${SHARED}.diagram.sourceSub`)}
        tone="neutral"
      />
      <FlowEdge label={t(`${SHARED}.diagram.flow1`)} />
      <FlowNode
        icon={<Forward size={20} />}
        title={t(`${SHARED}.diagram.forwarder`)}
        sub={t(`${SHARED}.diagram.forwarderSub`, { port })}
        tone="accent"
      />
      <FlowEdge label={t(`${SHARED}.diagram.flow2`)} />
      <FlowNode
        icon={<ShieldCheck size={20} />}
        title={t(`${SHARED}.diagram.utmstack`)}
        sub={t(`${SHARED}.diagram.utmstackSub`)}
        tone="brand"
      />
    </div>
  )
}

export function GenericCollectorGuide({ integration: i }: GenericCollectorGuideProps) {
  const { t } = useTranslation()
  const port = i.defaultPort ?? '514/udp'

  return (
    <div className="space-y-4">
      <Section title={t('integrations.setup.collector.howItWorksTitle')}>
        <p className="text-sm text-foreground/90">
          {t('integrations.setup.collector.howItWorksBody', { name: i.name })}
        </p>
        <GenericFlowDiagram source={i.name} port={port} />
      </Section>

      <Section title={t('integrations.setup.collector.step1Title')} step={1}>
        <p className="mb-2 text-xs text-muted-foreground">
          {t('integrations.setup.collector.step1Hint', { name: i.name })}
        </p>
        <CodeBlock
          code={`curl -fsSL https://install.utmstack.com/collector.sh | sudo bash -s -- \\
  --token=eyJ...workspace_id=acme...`}
        />
      </Section>

      <Section title={t('integrations.setup.collector.step2Title', { name: i.name })} step={2}>
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

      <Section title={t('integrations.setup.collector.step3Title')} step={3}>
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
