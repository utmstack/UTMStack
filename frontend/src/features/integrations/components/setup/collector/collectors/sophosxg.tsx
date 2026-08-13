import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide, forwarderHost } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.sophosxg'
const PORT = '7008'

function SophosXGGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="firewall-sophos-xg">
      <Section title={t(`${ROOT}.step2.title`)} step={2}>
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
  getName: () => 'SOPHOS_XG',
  matches: (n) => n.includes('sophos') && n.includes('xg'),
  sections: [],
  render: (m) => <SophosXGGuide module={m} />,
})
