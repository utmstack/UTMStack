import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide, forwarderHost } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.vmware'
const PORT = '7002'

function VMwareGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()

  // ESXi ships its host logs over TCP syslog, configured via esxcli.
  const loghostCmd = `esxcli system syslog config set --loghost='tcp://${host}:${PORT}'
esxcli system syslog reload`

  const firewallCmd = `esxcli network firewall ruleset set --ruleset-id=syslog --enabled=true
esxcli network firewall refresh`

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="vmware-esxi">
      <Section title={t(`${ROOT}.step1.title`)} step={2}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step1.body`)}</p>
        {/* ESXi syslog uses TCP. */}
        <CodeBlock code={`sudo /opt/utmstack-forwarder/utmstack_forwarder enable-integration vmware-esxi tcp`} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step1.note`)}</p>
      </Section>

      <Section title={t(`${ROOT}.step2.title`)} step={3}>
        <p className="mb-2 text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step2.body`}
            values={{ host, port: PORT }}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
        <CodeBlock code={loghostCmd} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step2.note`)}</p>
      </Section>

      <Section title={t(`${ROOT}.step3.title`)} step={4}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step3.body`)}</p>
        <CodeBlock code={firewallCmd} />
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'VMWARE',
  matches: (n) => n.includes('vmware') || n.includes('esxi'),
  sections: [],
  render: (m) => <VMwareGuide module={m} />,
})
