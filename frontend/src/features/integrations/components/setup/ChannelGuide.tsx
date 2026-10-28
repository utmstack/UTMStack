import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown } from 'lucide-react'
import type { ReactNode } from 'react'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import { AgentSetup } from '@/features/integrations/components/setup/AgentSetup'
import { forwarderHost } from '@/features/integrations/components/setup/collector/ForwarderGuide'
import type { Integration, IntegrationGuideChannel } from '@/features/integrations/types'

export function ChannelGuide({ integration }: { integration: Integration }) {
  const channel = integration.guideChannel
  if (!channel) return null

  switch (channel) {
    case 'syslog':
      return <SyslogDeviceGuide integration={integration} />
    case 'netflow':
      return <NetflowExporterGuide integration={integration} />
    case 'windows-agent':
      return <WindowsEventLogGuide integration={integration} />
    case 'linux-agent':
      return <LinuxJournalGuide integration={integration} />
  }
  return null
}

function GuideLayout({
  title,
  children,
}: {
  title: string
  children: ReactNode
}) {
  return (
    <div className="space-y-4">
      <Section title={title}>{children}</Section>
    </div>
  )
}

function SyslogDeviceGuide({ integration }: { integration: Integration }) {
  const { t } = useTranslation()
  const [protocol, setProtocol] = useState<'tcp' | 'udp'>('tcp')
  const host = forwarderHost()

  return (
    <GuideLayout title={`${integration.name} via Syslog`}>
      <p className="text-sm text-foreground/90">
        Configure the device to send syslog to the UTMStack Collector. TCP is recommended for chatty
        devices; the UDP receiver uses a 2,048-byte message buffer and truncates longer messages.
      </p>
      <label className="block max-w-40 text-xs text-muted-foreground">
        Protocol
        <select
          value={protocol}
          onChange={(event) => setProtocol(event.target.value as 'tcp' | 'udp')}
          className="mt-1 h-9 w-full rounded-md border border-border bg-background px-3 text-sm text-foreground"
        >
          <option value="tcp">TCP (recommended)</option>
          <option value="udp">UDP</option>
        </select>
      </label>
      <CodeBlock
        lang="config"
        code={`Collector: ${host}
Port:      7014
Protocol:  ${protocol.toUpperCase()}
Format:    RFC 5424 or RFC 3164`}
      />
      <p className="text-sm text-foreground/90">
        For a device that only supports port 514, change the Collector's syslog UDP port with:
      </p>
      <CodeBlock code="utmstack_forwarder change-port syslog udp 514" />
      <p className="text-xs text-muted-foreground">
        Only do this on a Collector that is not running Cisco integrations; they may already use UDP
        514. The command refuses a port that is busy. Changing a built-in port is not done with
        <code className="mx-1 font-mono">enable-integration</code>.
      </p>
      <p className="text-xs text-muted-foreground">
        {t('datasources.drawer.viewLogs')} will open Log Explorer for this integration's data type.
      </p>
      {integration.guideSetting && (
        <p className="text-sm text-foreground/90">{integration.guideSetting}</p>
      )}
    </GuideLayout>
  )
}

function NetflowExporterGuide({ integration }: { integration: Integration }) {
  const host = forwarderHost()
  return (
    <GuideLayout title={`${integration.name} NetFlow exporter`}>
      <p className="text-sm text-foreground/90">
        Configure the exporter to send NetFlow over UDP to the UTMStack Collector:
      </p>
      <CodeBlock
        lang="config"
        code={`Collector: ${host}
Port:      2055
Protocol:  UDP
Version:   NetFlow v9 (where supported)`}
      />
      <p className="text-sm text-foreground/90">
        Select the export version supported by the device and configure its source interface. Verify
        that flows arrive in Log Explorer using the View logs button.
      </p>
      {integration.guideSetting && (
        <p className="text-sm text-foreground/90">{integration.guideSetting}</p>
      )}
    </GuideLayout>
  )
}

function AgentChannelGuide({
  integration,
  channel,
  providerField,
}: {
  integration: Integration
  channel: Extract<IntegrationGuideChannel, 'windows-agent' | 'linux-agent'>
  providerField: string
}) {
  const { t } = useTranslation()
  const [showInstall, setShowInstall] = useState(false)
  const isWindows = channel === 'windows-agent'
  const agent: Integration = {
    id: `${integration.id}-agent-guide`,
    name: isWindows ? 'Windows Agent' : 'Linux Agent',
    moduleName: isWindows ? 'WINDOWS_AGENT' : 'LINUX_AGENT',
    dataType: isWindows ? 'wineventlog' : 'linux',
    kind: 'agents & syslog',
    status: 'available',
    description: '',
    category: 'operating-system',
    logo: integration.logo,
  }
  const filters = integration.logFilters ?? []
  const provider = filters.find((filter) => filter.field === providerField)?.value.join(', ')
  const eventChannel = filters.find((filter) => filter.field === 'log.channel')?.value.join(', ')

  return (
    <GuideLayout title={`${integration.name} via ${isWindows ? 'Windows Event Log' : 'Linux Journal'}`}>
      <p className="text-sm text-foreground/90">
        This technology sends events through the {isWindows ? 'Windows UTMStack agent' : 'Linux UTMStack agent'}.
        Install the agent on the host, then open the agent installation guide below.
      </p>
      <button
        type="button"
        onClick={() => setShowInstall((open) => !open)}
        className="flex w-full items-center justify-between rounded-md border border-border px-3 py-2 text-left text-sm font-medium hover:bg-muted/40"
      >
        <span>{isWindows ? 'Windows agent installation' : 'Linux agent installation'}</span>
        <ChevronDown size={14} className={showInstall ? 'rotate-180' : ''} />
      </button>
      {showInstall && <AgentSetup integration={agent} />}
      <Section title={isWindows ? 'Event channel and provider' : 'Journal identifier'}>
        {isWindows ? (
          <p className="text-sm text-foreground/90">
            Configure the product to write to the Windows Event Log. Expected provider:{' '}
            <code>{provider || 'product-specific'}</code>. Expected channel:{' '}
            <code>{eventChannel || 'product-specific'}</code>. Confirm both values on a sample event
            before relying on the filter.
          </p>
        ) : (
          <p className="text-sm text-foreground/90">
            Configure the product to send logs to journald or syslog. Expected journal identifier:{' '}
            <code>{provider || 'product-specific'}</code>.
          </p>
        )}
        {integration.guideSetting && (
          <p className="mt-2 text-sm text-foreground/90">{integration.guideSetting}</p>
        )}
      </Section>
      <p className="text-xs text-muted-foreground">
        {t('datasources.drawer.viewLogs')} opens Log Explorer with this integration's technology filters.
      </p>
    </GuideLayout>
  )
}

function WindowsEventLogGuide({ integration }: { integration: Integration }) {
  return (
    <AgentChannelGuide
      integration={integration}
      channel="windows-agent"
      providerField="log.providerName"
    />
  )
}

function LinuxJournalGuide({ integration }: { integration: Integration }) {
  return (
    <AgentChannelGuide
      integration={integration}
      channel="linux-agent"
      providerField="log.syslogIdentifier"
    />
  )
}
