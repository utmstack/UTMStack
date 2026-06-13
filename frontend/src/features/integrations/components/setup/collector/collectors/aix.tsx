import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.aix'
const PORT = '7016'

function AixGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="ibm-aix">
      <Section title={t(`${ROOT}.step1.title`)} step={2}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step1.body`)}</p>
        <CodeBlock code={`/opt/utmstack-forwarder/utmstack_forwarder enable-integration ibm-aix udp`} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step1.note`)}</p>
      </Section>
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
  getName: () => 'AIX',
  matches: (n) => n.includes('aix'),
  sections: [],
  render: (m) => <AixGuide module={m} />,
})
