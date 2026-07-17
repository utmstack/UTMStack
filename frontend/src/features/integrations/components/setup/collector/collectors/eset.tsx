import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.eset'
const IMG = '/integrations/guides/collector/eset'
const PORT = '7003'

// ESET Protect console steps (mirror the legacy guide): navigate to the Syslog
// server settings and point them at the Forwarder.
const CONSOLE_STEPS: Array<{ key: string; img: string }> = [
  { key: 'moreMenu', img: `${IMG}/main_page.png` },
  { key: 'serverConfig', img: `${IMG}/more_settings.png` },
  { key: 'advanced', img: `${IMG}/server_config.png` },
  { key: 'syslog', img: `${IMG}/syslog_server.png` },
]

function EsetGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="antivirus-esmc-eset">
      {CONSOLE_STEPS.map((s, idx) => (
        <Section key={s.key} title={t(`${ROOT}.console.${s.key}.title`)} step={idx + 3}>
          <p className="text-sm text-foreground/90">
            <Trans
              i18nKey={`${ROOT}.console.${s.key}.body`}
              values={{ port: PORT }}
              components={{ hl: <strong className="font-semibold text-primary" /> }}
            />
          </p>
          <img
            src={s.img}
            alt=""
            className="mt-3 max-w-full rounded-md border border-border"
            onError={(e) => {
              e.currentTarget.style.display = 'none'
            }}
          />
        </Section>
      ))}
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'ESET',
  matches: (n) => n.includes('eset'),
  sections: [],
  render: (m) => <EsetGuide module={m} />,
})
