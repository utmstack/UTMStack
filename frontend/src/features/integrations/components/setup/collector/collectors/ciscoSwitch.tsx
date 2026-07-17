import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.ciscoSwitch'
const PORT = '514'

function CiscoSwitchGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="cisco-switch">
      <Section title={t(`${ROOT}.step2.title`)} step={3}>
        <p className="mb-2 text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step2.body`}
            values={{ port: PORT }}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
        <CodeBlock lang="config" code={`logging on
logging host <forwarder-ip>
logging trap informational`} />
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'CISCO_SWITCH',
  matches: (n) => n.includes('cisco') && n.includes('switch'),
  sections: [],
  render: (m) => <CiscoSwitchGuide module={m} />,
})
