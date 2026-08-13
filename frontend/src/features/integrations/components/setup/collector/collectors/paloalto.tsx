import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide, forwarderHost } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.paloalto'
const PORT = '7006'

function PaloAltoGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="firewall-paloalto">
      <Section title={t(`${ROOT}.step2.title`)} step={2}>
        <p className="text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step2.body`}
            values={{ host, port: PORT }}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
      </Section>

      <Section title={t(`${ROOT}.step3.title`)} step={4}>
        <p className="text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step3.body`}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'PALO_ALTO',
  matches: (n) => n.includes('palo'),
  sections: [],
  render: (m) => <PaloAltoGuide module={m} />,
})
