import { useTranslation } from 'react-i18next'
import { ExternalLink } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'

export function CustomSetup() {
  const { t } = useTranslation()

  return (
    <div className="space-y-4">
      <Section title={t('integrations.setup.custom.whenTitle')}>
        <p className="text-sm text-foreground/90">
          {t('integrations.setup.custom.whenBody')}
        </p>
      </Section>

      <Section title={t('integrations.setup.custom.endpointsTitle')}>
        <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
          <dt className="text-muted-foreground">{t('integrations.setup.custom.syslogUdp')}</dt>
          <dd className="font-mono">utmstack-collector.your-net:514</dd>
          <dt className="text-muted-foreground">{t('integrations.setup.custom.syslogTls')}</dt>
          <dd className="font-mono">utmstack-collector.your-net:6514</dd>
          <dt className="text-muted-foreground">{t('integrations.setup.custom.httpJson')}</dt>
          <dd className="font-mono">https://ingest.utmstack.com/v1/json</dd>
        </dl>
      </Section>

      <Section title={t('integrations.setup.custom.payloadTitle')}>
        <CodeBlock
          code={`POST https://ingest.utmstack.com/v1/json
Authorization: Bearer eyJ...workspace_id=acme...
Content-Type: application/json

{
  "@timestamp": "2026-05-01T16:42:00Z",
  "host": "my-app-01",
  "source": "custom-app",
  "message": "user 'jdoe' logged in",
  "user": "jdoe",
  "ip": "10.4.18.91"
}`}
        />
      </Section>

      <Section title={t('integrations.setup.custom.parserTitle')}>
        <p className="text-sm text-foreground/90">
          {t('integrations.setup.custom.parserBody')}
        </p>
        <a className="mt-2 inline-flex items-center gap-1 text-[11px] text-primary hover:underline" href="#">
          {t('integrations.setup.custom.openEditor')} <ExternalLink size={10} />
        </a>
      </Section>
    </div>
  )
}
