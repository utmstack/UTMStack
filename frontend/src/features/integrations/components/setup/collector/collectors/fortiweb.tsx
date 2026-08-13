import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide, forwarderHost } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.fortiweb'
const PORT = '7018'

function FortiwebGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="firewall-fortiweb">
      <Section title={t(`${ROOT}.step2.title`)} step={2}>
        <p className="mb-2 text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.step2.body`}
            values={{ port: PORT }}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
        <CodeBlock lang="config" code={`config log syslog-policy
    edit "utmstack"
        config syslog-server-list
            edit 1
                set server ${host}
                set port ${PORT}
            next
        end
    next
end
config log syslogd
    set status enable
    set policy "utmstack"
end`} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step2.note`)}</p>
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'FORTIWEB',
  matches: (n) => n.includes('fortiweb'),
  sections: [],
  render: (m) => <FortiwebGuide module={m} />,
})
