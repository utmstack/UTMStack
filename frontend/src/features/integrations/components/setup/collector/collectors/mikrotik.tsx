import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide, forwarderHost } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.mikrotik'
const PORT = '7007'

function MikroTikGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="firewall-mikrotik">
      <Section title={t(`${ROOT}.step1.title`)} step={2}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step1.body`)}</p>
        <CodeBlock code="/opt/utmstack-forwarder/utmstack_forwarder enable-integration firewall-mikrotik udp" />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step1.note`)}</p>
      </Section>
      <Section title={t(`${ROOT}.step2.title`)} step={3}>
        <p className="mb-2 text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step2.body`}
            values={{ port: PORT }}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
        <CodeBlock lang="config" code={`/system logging action
add name=utmstack target=remote remote=${host} remote-port=${PORT} \\
    bsd-syslog=yes syslog-facility=local0 syslog-severity=auto

/system logging
add action=utmstack topics=account,error,info,warning`} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step2.note`)}</p>
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'MIKROTIK',
  matches: (n) => n.includes('mikrotik'),
  sections: [],
  render: (m) => <MikroTikGuide module={m} />,
})
