import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide, forwarderHost } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.netflow'
const PORT = '2055'

function NetflowGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="netflow">
      <Section title={t(`${ROOT}.step2.title`)} step={3}>
        <p className="mb-2 text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step2.body`}
            values={{ port: PORT }}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
        <CodeBlock lang="config" code={`ip flow-export version 9
ip flow-export destination ${host} ${PORT}
ip flow-export source <interface>

interface <interface>
 ip flow ingress
 ip flow egress`} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step2.note`)}</p>
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'NETFLOW',
  sections: [],
  render: (m) => <NetflowGuide module={m} />,
})
