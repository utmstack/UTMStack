import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import { ChannelGuide } from '@/features/integrations/components/setup/ChannelGuide'
import { ForwarderGuide } from '@/features/integrations/components/setup/collector/ForwarderGuide'
import type { RemoteEnableSelection } from './RemoteEnablePanel'
import { availableProtosFor, defaultPortFor, type Proto } from './protoCatalog'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector'

function getDefaultEndpoint(integration: Integration): { proto: Proto; port: string; protos: Proto[] } | null {
  const dataType = integration.forwarderDataType || integration.dataType
  if (!dataType) return null

  const availableProtos = availableProtosFor(dataType)
  const configuredDefault = integration.defaultPort?.match(/^(\d+)(?:\/(udp|tcp|tls|http|https))?$/i)
  const configuredProto = configuredDefault?.[2]?.toLowerCase() as Proto | undefined
  const configuredPort = Number(configuredDefault?.[1])
  const validConfiguredPort = configuredPort > 0 && configuredPort <= 65535
  const catalogProtos = availableProtos.filter((proto) => !!defaultPortFor(dataType, proto))

  if (catalogProtos.length > 0) {
    const proto = configuredProto && catalogProtos.includes(configuredProto)
      ? configuredProto
      : catalogProtos[0]
    const port = defaultPortFor(dataType, proto) || (validConfiguredPort ? String(configuredPort) : '')
    return port ? { proto, port, protos: catalogProtos } : null
  }

  if (!configuredProto || !validConfiguredPort) return null
  return { proto: configuredProto, port: String(configuredPort), protos: [configuredProto] }
}

export function GenericCollectorGuide({ integration }: { integration: Integration }) {
  const { t } = useTranslation()
  const endpoint = getDefaultEndpoint(integration)
  const [selection, setSelection] = useState<RemoteEnableSelection>(() => ({
    proto: endpoint?.proto ?? 'udp',
    port: endpoint?.port ?? '',
  }))

  if (integration.guideChannel) {
    return <ChannelGuide integration={integration} />
  }

  const forwarderDataType = integration.forwarderDataType || integration.dataType

  if (integration.ingestType?.toLowerCase() === 'forwarder' && endpoint && forwarderDataType) {
    const host = selection.collector?.ip || '<collector-host>'

    return (
      <ForwarderGuide
        source={integration.name}
        sourceType={forwarderDataType}
        port={selection.port || endpoint.port}
        defaultPort={endpoint.port}
        defaultProto={endpoint.proto}
        availableProtos={endpoint.protos}
        onSetSelection={setSelection}
      >
        <Section title={t(`${ROOT}.genericSourceTitle`, { name: integration.name })} step={2}>
          <p className="mb-2 text-sm text-foreground/90">
            {t(`${ROOT}.genericSourceBody`, { name: integration.name })}
          </p>
          <CodeBlock
            lang="config"
            code={`Host:      ${host}
Port:      ${selection.port || endpoint.port}
Protocol:  ${selection.proto.toUpperCase()}`}
          />
        </Section>
      </ForwarderGuide>
    )
  }

  return (
    <Section title={t(`${ROOT}.guideUnavailableTitle`)}>
      <p className="text-sm text-foreground/90">
        {t(`${ROOT}.guideUnavailableBody`, { name: integration.name })}
      </p>
    </Section>
  )
}
