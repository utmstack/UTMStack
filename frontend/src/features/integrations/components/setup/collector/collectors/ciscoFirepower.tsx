import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.ciscoFirepower'
const PORT = '514'

function CiscoFirepowerGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="firewall-cisco-firepower">
      <Section title={t(`${ROOT}.step2.title`)} step={3}>
        <p className="text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step2.body`}
            values={{ port: PORT }}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'FIRE_POWER',
  matches: (n) => n.includes('firepower') || (n.includes('fire') && n.includes('power')),
  sections: [],
  render: (m) => <CiscoFirepowerGuide module={m} />,
})
