import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide, forwarderHost } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.sonicwall'
const PORT = '7009'

function SonicWallGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="firewall-sonicwall">
      <Section title={t(`${ROOT}.step1.title`)} step={2}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step1.body`)}</p>
        <CodeBlock code={`sudo /opt/utmstack-forwarder/utmstack_forwarder enable-integration firewall-sonicwall udp`} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step1.note`)}</p>
      </Section>

      <Section title={t(`${ROOT}.step2.title`)} step={3}>
        <p className="text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step2.body`}
            values={{ host, port: PORT }}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'SONIC_WALL',
  matches: (n) => n.includes('sonic'),
  sections: [],
  render: (m) => <SonicWallGuide module={m} />,
})
