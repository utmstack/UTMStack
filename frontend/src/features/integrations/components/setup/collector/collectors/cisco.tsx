import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.cisco'
const PORT = '514'

function CiscoAsaGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="firewall-cisco-asa">
      <Section title={t(`${ROOT}.step2.title`)} step={2}>
        <p className="mb-2 text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step2.body`}
            values={{ port: PORT }}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
        <CodeBlock lang="config" code={`logging enable
logging host <interface> <forwarder-ip> udp/${PORT}
logging trap informational`} />
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'CISCO',
  matches: (n) => n.includes('cisco') && !n.includes('switch') && !n.includes('meraki') && !n.includes('firepower'),
  sections: [],
  render: (m) => <CiscoAsaGuide module={m} />,
})
