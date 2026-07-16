import { useTranslation } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.oracle'
const PORT = '7021'

function OracleGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()

  const rsyslogConfig = `input(type="imfile"
    File="BASEDIR/admin/*/adump/*.aud"
    Tag="oracle-audit"
    Severity="info"
    Facility="local5")

input(type="imfile"
    File="BASEDIR/diag/tnslsnr/*/listener/trace/listener.log"
    Tag="oracle-listener"
    Severity="info"
    Facility="local5")

input(type="imfile"
    File="BASEDIR/diag/rdbms/*/*/trace/alert_*.log"
    Tag="oracle-alerts"
    Severity="info"
    Facility="local5")

local5.* @<forwarder-ip>:${PORT}`

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="oracle">
      <Section title={t(`${ROOT}.step2.title`)} step={3}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step2.body`)}</p>
        <CodeBlock code="sudo touch /etc/rsyslog.d/oracle-utmstack.conf" />
        <p className="mt-3 mb-2 text-sm text-foreground/90">{t(`${ROOT}.step2.editBody`)}</p>
        <CodeBlock lang="config" code={rsyslogConfig} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step2.note`)}</p>
      </Section>
      <Section title={t(`${ROOT}.step3.title`)} step={4}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step3.body`)}</p>
        <CodeBlock code="sudo systemctl restart rsyslog" />
      </Section>
      <Section title={t(`${ROOT}.step4.title`)} step={5}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step4.body`)}</p>
        <CodeBlock lang="bash" code={`logger -p local5.info "Test Oracle Audit log message"
logger -p local5.info "Test Oracle Listener log message"
logger -p local5.info "Test Oracle Alert log message"`} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step4.note`)}</p>
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'ORACLE',
  matches: (n) => n.includes('oracle'),
  sections: [],
  render: (m) => <OracleGuide module={m} />,
})
