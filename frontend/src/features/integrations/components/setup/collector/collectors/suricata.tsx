import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide, forwarderHost } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.suricata'
const PORT = '7019'

const EVE_LOG = `outputs:
  - eve-log:
      enabled: yes
      filetype: syslog
      facility: local5
      types:
        - alert
        - dns
        - http
        - tls
        - flow`

function SuricataGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()

  // Forward the local5 facility (where Suricata writes its eve JSON) to the
  // Forwarder. "@" = UDP, "@@" = TCP in rsyslog syntax.
  const rsyslogRule = `local5.*  @${host}:${PORT}`

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="suricata">
      <Section title={t(`${ROOT}.step1.title`)} step={2}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step1.body`)}</p>
        <CodeBlock code={`/opt/utmstack-forwarder/utmstack_forwarder enable-integration suricata udp`} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step1.note`)}</p>
      </Section>

      <Section title={t(`${ROOT}.step2.title`)} step={3}>
        <p className="mb-2 text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step2.body`}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
        <CodeBlock lang="yaml" code={EVE_LOG} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step2.note`)}</p>
      </Section>

      <Section title={t(`${ROOT}.step3.title`)} step={4}>
        <p className="mb-2 text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step3.body`}
            values={{ port: PORT }}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
        <CodeBlock lang="config" code={rsyslogRule} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step3.note`)}</p>
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'SURICATA',
  matches: (n) => n.includes('suricata'),
  sections: [],
  render: (m) => <SuricataGuide module={m} />,
})
