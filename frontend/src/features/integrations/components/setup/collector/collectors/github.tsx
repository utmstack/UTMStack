import { useTranslation, Trans } from 'react-i18next'
import { registerCollector } from '../registry'
import { ForwarderGuide, HttpsCertsSection, forwarderHost } from '../ForwarderGuide'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.github'
const PORT = '7020'
const IMG = '/integrations/guides/github'

function GithubGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()
  const webhookUrl = `https://${host}:${PORT}/github`

  return (
    <ForwarderGuide source={t(`${ROOT}.source`)} port={PORT} sourceType="github" hideTLS>
      <HttpsCertsSection step={2} port={PORT} />

      <Section title={t(`${ROOT}.step1.title`)} step={3}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.step1.body`)}</p>
        <CodeBlock code="sudo /opt/utmstack-forwarder/utmstack_forwarder enable-integration github https" />
        <p className="mt-2 rounded-md bg-amber-500/10 border border-amber-500/30 px-3 py-2 text-[11px] text-amber-700 dark:text-amber-400">
          {t(`${ROOT}.step1.tokenNote`)}
        </p>
      </Section>

      <Section title={t(`${ROOT}.certWarning.title`)} variant="warning">
        <p className="text-sm text-foreground/90">{t(`${ROOT}.certWarning.body`)}</p>
      </Section>

      <Section title={t(`${ROOT}.step2.title`)} step={4}>
        <p className="mb-3 text-sm text-foreground/90">{t(`${ROOT}.step2.body`)}</p>
        <img
          src={`${IMG}/webhook.png`}
          alt="GitHub webhook settings"
          className="mb-3 w-full rounded-md border border-border"
        />
        <div className="mb-3 space-y-2 rounded-lg border border-border bg-muted/30 p-3">
          <div className="flex flex-col gap-0.5">
            <span className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
              {t(`${ROOT}.step2.payloadUrl`)}
            </span>
            <code className="break-all font-mono text-xs text-primary">{webhookUrl}</code>
          </div>
          <div className="flex flex-col gap-0.5">
            <span className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
              {t(`${ROOT}.step2.contentType`)}
            </span>
            <code className="font-mono text-xs text-foreground">application/json</code>
          </div>
          <div className="flex flex-col gap-0.5">
            <span className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
              {t(`${ROOT}.step2.secret`)}
            </span>
            <span className="text-xs text-foreground/80">
              <Trans
                i18nKey={`${ROOT}.step2.secretHint`}
                components={{ hl: <strong className="font-semibold text-primary" /> }}
              />
            </span>
          </div>
          <div className="flex flex-col gap-0.5">
            <span className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
              {t(`${ROOT}.step2.events`)}
            </span>
            <span className="text-xs text-foreground/80">{t(`${ROOT}.step2.eventsHint`)}</span>
          </div>
        </div>
        <img
          src={`${IMG}/github-webhook.png`}
          alt="GitHub webhook form"
          className="w-full rounded-md border border-border"
        />
      </Section>
    </ForwarderGuide>
  )
}

registerCollector({
  getName: () => 'GITHUB',
  sections: [],
  render: (m) => <GithubGuide module={m} />,
})
