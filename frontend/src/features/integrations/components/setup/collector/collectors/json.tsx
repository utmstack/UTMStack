import { useTranslation } from 'react-i18next'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import { registerCollector } from '../registry'

const ROOT = 'integrations.setup.collector.json'

function JsonEndpoint() {
  const { t } = useTranslation()
  const host = typeof window !== 'undefined'
    ? window.location.host.split(':')[0]
    : 'utmstack.local'

  return (
    <Section title={t(`${ROOT}.endpoint.title`)}>
      <p className="mb-2 text-xs text-muted-foreground">{t(`${ROOT}.endpoint.hint`)}</p>
      <CodeBlock
        code={`POST https://${host}:8080/v1/logs
Content-Type: application/json
Authorization: Bearer <federation-token>

{
  "source": "my-app",
  "events": [ { "ts": "2026-06-10T12:34:56Z", "level": "info", "msg": "hello" } ]
}`}
      />
    </Section>
  )
}

registerCollector({
  getName: () => 'JSON',
  sections: [
    {
      id: 'overview',
      titleKey: `${ROOT}.sections.overview.title`,
      bodyKey: `${ROOT}.sections.overview.body`,
    },
    {
      id: 'auth',
      titleKey: `${ROOT}.sections.auth.title`,
      bodyKey: `${ROOT}.sections.auth.body`,
    },
  ],
  render: () => <JsonEndpoint />,
})
