import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.kaspersky'
const IMG = '/integrations/guides/collector/kaspersky'
const PORT = '7004'

// Kaspersky Security Center console steps (mirror the legacy guide): open the
// SIEM integration settings and point them at the Forwarder.
const CONSOLE_STEPS: Array<{ key: string; img: string }> = [
  { key: 'consoleSettings', img: `${IMG}/main_page.png` },
  { key: 'siem', img: `${IMG}/integration.png` },
  { key: 'configure', img: `${IMG}/configuration.png` },
]

function KasperskyGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="antivirus-kaspersky">
      <Section title={t(`${ROOT}.step1.title`)} step={2}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step1.body`)}</p>
        <CodeBlock code={`/opt/utmstack-forwarder/utmstack_forwarder enable-integration antivirus-kaspersky udp`} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.step1.note`)}</p>
      </Section>

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
  getName: () => 'KASPERSKY',
  matches: (n) => n.includes('kaspersky'),
  sections: [],
  render: (m) => <KasperskyGuide module={m} />,
})
