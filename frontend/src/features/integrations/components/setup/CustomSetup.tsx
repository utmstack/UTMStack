import { useState } from 'react'
import { useTranslation, Trans } from 'react-i18next'
import {  Forward, Server, ShieldCheck } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import {
  ForwarderGuide,
  FlowNode,
  FlowEdge,
} from '@/features/integrations/components/setup/collector/ForwarderGuide'
import type { Integration } from '@/features/integrations/types'
import {  type RemoteEnableSelection } from './collector/RemoteEnablePanel'
const ROOT = 'integrations.setup.custom'

// A custom data type has no ProtoPorts/HTTPPorts catalog entry, so every
// protocol is offered — same set the legacy protocol tabs used to expose.


export function CustomSetup({ integration }: { integration: Integration }) {
  const { t } = useTranslation()
  const name = integration.dataType || integration.moduleName?.toLowerCase() || 'my-integration'

  const [selection, setSelection] = useState<RemoteEnableSelection>({ proto: 'udp', port: '7100' })
  const { proto, port } = selection

  const isHttp = proto === 'http' || proto === 'https'
  // HTTP/HTTPS listeners bind to 127.0.0.1 by default, so the endpoint is local
  // to the Forwarder host.
  const endpoint = `${proto}://127.0.0.1:${port}/logs`

  return (
    <div className="space-y-4">
    <ForwarderGuide port={selection.port} source={`${ROOT}`}  sourceType={integration.dataType!}

    onSetSelection={setSelection}

    header={(
      <Section title={t(`${ROOT}.intro.title`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.intro.body`)}</p>
        <div className="mt-3 flex items-stretch justify-center gap-1 sm:gap-2">
          <FlowNode
            icon={<Server size={20} className="text-muted-foreground" />}
            title={t(`${ROOT}.diagram.source`)}
            sub={t(`${ROOT}.diagram.sourceSub`)}
            tone="neutral"
          />
          <FlowEdge label={selection.proto.toUpperCase()} />
          <FlowNode
            icon={<Forward size={20} />}
            title={t(`${ROOT}.diagram.forwarder`)}
            sub={`:${selection.port}`}
            tone="accent"
          />
          <FlowEdge label={t(`${ROOT}.diagram.ingests`)} />
          <FlowNode
            icon={<ShieldCheck size={20} />}
            title="UTMStack"
            sub={t(`${ROOT}.diagram.utmstackSub`)}
            tone="brand"
          />
        </div>
        <p className="mt-3 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
          {t(`${ROOT}.intro.note`)}
        </p>
      </Section>
    )}
    >

      <Section title={t(`${ROOT}.send.title`)} step={2}>
        {isHttp ? (
          <>
            <p className="mb-2 text-sm text-foreground/90">
              <Trans
                i18nKey={`${ROOT}.send.bodyHttp`}
                values={{ endpoint }}
                components={{ hl: <strong className="font-semibold text-primary" /> }}
              />
            </p>
            <CodeBlock
              lang="bash"
              code={`curl -k -X POST "${endpoint}" \\
  -H "Content-Type: application/json" \\
  -d '{"message":"user jdoe logged in","host":"app-01","user":"jdoe"}'`}
            />
            <p className="mt-2 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
              {t(`${ROOT}.send.httpBindNote`)}
            </p>
          </>
        ) : (
          <p className="text-sm text-foreground/90">
            <Trans
              i18nKey={`${ROOT}.send.bodyNetwork`}
              values={{ host:selection.collector?.ip, port:selection.port, proto: selection.proto.toUpperCase() }}
              components={{ hl: <strong className="font-semibold text-primary" /> }}
            />
          </p>
        )}
      </Section>

      {/* Parser hint. */}
      <Section title={t(`${ROOT}.parserTitle`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.parserBody`, { dataType: name })}</p>
      </Section>

    </ForwarderGuide>

    </div>
  )
}

