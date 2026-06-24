import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide, forwarderHost } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.sentinelOne'
const IMG = '/integrations/guides/collector/sentinelone'
const PORT = '7012'

// SentinelOne management console steps (mirror the legacy guide): open Settings,
// then the Syslog integration, and point it at the Forwarder.
const CONSOLE_STEPS: Array<{ key: string; img: string }> = [
  { key: 'settings', img: `${IMG}/main_page.png` },
  { key: 'syslog', img: `${IMG}/sentinelone_integration.png` },
]

function SentinelOneGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="antivirus-sentinel-one">
      <Section title={t(`${ROOT}.step1.title`)} step={2}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step1.body`)}</p>
        {/* SentinelOne streams CEF syslog over TCP. */}
        <CodeBlock code={`sudo /opt/utmstack-forwarder/utmstack_forwarder enable-integration antivirus-sentinel-one tcp`} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step1.note`)}</p>
      </Section>

      {CONSOLE_STEPS.map((s, idx) => (
        <Section key={s.key} title={t(`${ROOT}.console.${s.key}.title`)} step={idx + 3}>
          <p className="text-sm text-foreground/90">
            <Trans
              i18nKey={`${ROOT}.console.${s.key}.body`}
              values={{ host, port: PORT }}
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
  getName: () => 'SENTINEL_ONE',
  matches: (n) => n.includes('sentinel'),
  sections: [],
  render: (m) => <SentinelOneGuide module={m} />,
})
