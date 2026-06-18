import { useState } from 'react'
import { useTranslation, Trans } from 'react-i18next'
import { ChevronDown, Download, Server, ShieldCheck } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { registerCollector } from '../registry'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import { useConectionKey } from '@/features/integrations/hooks/useConnectionKey'
import { forwarderHost, FlowNode, FlowEdge } from '../ForwarderGuide'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.as400'

const CONFIG_TEMPLATE = `# as400-config.yaml — edit to add or remove AS/400 servers.
# The collector watches this file and reloads automatically on save.
servers:
  - tenant:   default
    hostname: 10.0.0.5
    userId:   QSECOFR
    password: changeme`

function AS400CollectorInstall() {
  const { t } = useTranslation()
  const { key } = useConectionKey()
  const host = forwarderHost()
  const token = key.data?.connectionKey ?? '*'.repeat(7)

  const installCmd =
    `sudo bash -c "
  apt update -y && apt install wget -y && \\
  mkdir -p /opt/utmstack-as400-collector && \\
  wget --no-check-certificate -P /opt/utmstack-as400-collector \\
    https://${host}:9001/private/dependencies/collector/as400/utmstack_as400_collector_service && \\
  chmod 755 /opt/utmstack-as400-collector/utmstack_as400_collector_service && \\
  /opt/utmstack-as400-collector/utmstack_as400_collector_service install ${host} <secret>${token}</secret> yes
"`

  return (
    <Section title={t(`${ROOT}.install.title`)} step={1}>
      <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.install.body`)}</p>
      <CodeBlock code={installCmd} />
      <p className="mt-2 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
        {t(`${ROOT}.install.reuse`)}
      </p>
    </Section>
  )
}

const UNINSTALL_CMD = `sudo bash -c "
  /opt/utmstack-as400-collector/utmstack_as400_collector_service uninstall && \\
  rm -rf /opt/utmstack-as400-collector
"`

function AS400UninstallSection() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center justify-between px-4 py-3 text-left text-sm font-medium hover:bg-muted/40 transition-colors"
      >
        <span>{t(`${ROOT}.uninstall.title`)}</span>
        <ChevronDown size={14} className={cn('shrink-0 text-muted-foreground transition-transform duration-200', open && 'rotate-180')} />
      </button>
      {open && (
        <div className="space-y-3 border-t border-border px-4 pb-4 pt-3">
          <p className="text-sm text-foreground/90">{t(`${ROOT}.uninstall.body`)}</p>
          <CodeBlock code={UNINSTALL_CMD} />
        </div>
      )}
    </div>
  )
}

function AS400Guide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()

  return (
    <div className="space-y-4">
      <Section title={t(`${ROOT}.intro.title`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.intro.body`)}</p>
        <div className="mt-3 flex items-stretch justify-center gap-1 sm:gap-2">
          <FlowNode
            icon={<Server size={20} className="text-muted-foreground" />}
            title="AS/400"
            sub={t(`${ROOT}.intro.diagram.source`)}
            tone="neutral"
          />
          <FlowEdge label={t(`${ROOT}.intro.diagram.edge1`)} />
          <FlowNode
            icon={<Download size={20} />}
            title={t(`${ROOT}.intro.diagram.collector`)}
            sub={t(`${ROOT}.intro.diagram.collectorSub`)}
            tone="accent"
          />
          <FlowEdge label={t(`${ROOT}.intro.diagram.edge2`)} />
          <FlowNode
            icon={<ShieldCheck size={20} />}
            title="UTMStack"
            sub={t(`${ROOT}.intro.diagram.utmstackSub`)}
            tone="brand"
          />
        </div>
        <p className="mt-3 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
          {t(`${ROOT}.intro.note`)}
        </p>
      </Section>

      <AS400CollectorInstall />

      <Section title={t(`${ROOT}.configure.title`)} step={2}>
        <p className="mb-2 text-sm text-foreground/90">{t(`${ROOT}.configure.body`)}</p>
        <CodeBlock lang="yaml" code={CONFIG_TEMPLATE} />
        <p className="mt-2 text-[11px] text-muted-foreground">{t(`${ROOT}.configure.note`)}</p>
      </Section>

      <Section title={t(`${ROOT}.journal.title`)} step={3}>
        <p className="text-sm text-foreground/90">
          <Trans
            i18nKey={`${ROOT}.journal.body`}
            components={{ hl: <strong className="font-semibold text-primary" /> }}
          />
        </p>
      </Section>

      <AS400UninstallSection />
    </div>
  )
}

registerCollector({
  getName: () => 'AS_400',
  matches: (n) => n === 'as400' || n === 'as_400',
  sections: [],
  render: (m) => <AS400Guide module={m} />,
})
