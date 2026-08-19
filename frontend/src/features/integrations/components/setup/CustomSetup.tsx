import { useState } from 'react'
import { useTranslation, Trans } from 'react-i18next'
import { ChevronDown, Forward, Server, ShieldCheck } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import {
  ForwarderInstall,
  ForwarderUninstallSection,
  FlowNode,
  FlowEdge,
  forwarderHost,
} from '@/features/integrations/components/setup/collector/ForwarderGuide'
import { RemoteEnablePanel, type RemoteEnableSelection } from '@/features/integrations/components/setup/collector/RemoteEnablePanel'
import type { Proto } from '@/features/integrations/components/setup/collector/protoCatalog'
import type { Integration } from '@/features/integrations/types'
import { useCollectorIntegration } from '@/features/integrations/hooks/useCollectorIntegration'

const ROOT = 'integrations.setup.custom'

// A custom data type has no ProtoPorts/HTTPPorts catalog entry, so every
// protocol is offered — same set the legacy protocol tabs used to expose.
const ALL_PROTOS: Proto[] = ['udp', 'tcp', 'tls', 'http', 'https']

const FWD = '/opt/utmstack-forwarder/utmstack_forwarder'

export function CustomSetup({ integration }: { integration: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()
  const { forwarders } = useCollectorIntegration()
  // Same criterion as ForwarderGuide: skip the install step entirely once a
  // Forwarder is online, only fall back to the "none online" prompt once the
  // query has resolved.
  const hasOnlineForwarder = (forwarders.data ?? []).length > 0

  // The forwarder data-type for this custom integration. Falls back to a sample
  // slug if the catalog row has no explicit data type.
  const name = integration.dataType || integration.moduleName?.toLowerCase() || 'my-integration'

  const [selection, setSelection] = useState<RemoteEnableSelection>({ proto: 'udp', port: '7100', isMaster: false, apiKey: null })
  const { proto, port } = selection

  const isHttp = proto === 'http' || proto === 'https'
  const needsCerts = proto === 'tls' || proto === 'https'
  const enableCmd = `${FWD} enable-integration ${name} ${port} ${proto}`
  const tlsCertsCmd = `${FWD} load-tls-certs /path/to/cert.crt /path/to/key.key`
  // HTTP/HTTPS listeners bind to 127.0.0.1 by default, so the endpoint is local
  // to the Forwarder host.
  const endpoint = `${proto}://127.0.0.1:${port}/logs`

  return (
    <div className="space-y-4">
      {/* Intro + flow diagram. */}
      <Section title={t(`${ROOT}.intro.title`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.intro.body`)}</p>
        <div className="mt-3 flex items-stretch justify-center gap-1 sm:gap-2">
          <FlowNode
            icon={<Server size={20} className="text-muted-foreground" />}
            title={t(`${ROOT}.diagram.source`)}
            sub={t(`${ROOT}.diagram.sourceSub`)}
            tone="neutral"
          />
          <FlowEdge label={proto.toUpperCase()} />
          <FlowNode
            icon={<Forward size={20} />}
            title={t(`${ROOT}.diagram.forwarder`)}
            sub={`:${port}`}
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

      {/* Step 1: install the Forwarder (shared) — skipped once one is online. */}
      {!forwarders.isLoading && !hasOnlineForwarder && <ForwarderInstall source={integration.name} />}

      {/* Step 2: pick a forwarder + protocol/port and enable/disable remotely. */}
      <RemoteEnablePanel
        dataType={name}
        availableProtos={ALL_PROTOS}
        step={2}
        onSelectionChange={setSelection}
      />

      {/* Step 3: point the source at the Forwarder. */}
      <Section title={t(`${ROOT}.send.title`)} step={3}>
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
              values={{ host, port, proto: proto.toUpperCase() }}
              components={{ hl: <strong className="font-semibold text-primary" /> }}
            />
          </p>
        )}
      </Section>

      {/* Parser hint. */}
      <Section title={t(`${ROOT}.parserTitle`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.parserBody`, { dataType: name })}</p>
      </Section>

      {/* Optional — do it via CLI instead, reactive to the selection above. */}
      <ManualCommandSection needsCerts={needsCerts} tlsCertsCmd={tlsCertsCmd} enableCmd={enableCmd} />

      <ForwarderUninstallSection />
    </div>
  )
}

function ManualCommandSection({
  needsCerts,
  tlsCertsCmd,
  enableCmd,
}: {
  needsCerts: boolean
  tlsCertsCmd: string
  enableCmd: string
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center justify-between px-4 py-3 text-left text-sm font-medium hover:bg-muted/40 transition-colors"
      >
        <span>{t(`${ROOT}.manualCommand.title`)}</span>
        <ChevronDown size={14} className={cn('shrink-0 text-muted-foreground transition-transform duration-200', open && 'rotate-180')} />
      </button>
      {open && (
        <div className="space-y-3 border-t border-border px-4 pb-4 pt-3">
          <p className="text-sm text-foreground/90">{t(`${ROOT}.manualCommand.intro`)}</p>
          {needsCerts && (
            <div>
              <p className="mb-1.5 text-[11px] font-medium text-muted-foreground">{t(`${ROOT}.choose.tlsFirst`)}</p>
              <CodeBlock code={tlsCertsCmd} />
            </div>
          )}
          <p className="mb-1.5 text-[11px] font-medium text-muted-foreground">{t(`${ROOT}.choose.enableLabel`)}</p>
          <CodeBlock code={enableCmd} />
          <p className="rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
            <Trans
              i18nKey={`${ROOT}.choose.authNote`}
              components={{ code: <code className="font-mono text-foreground/80" /> }}
            />
          </p>
        </div>
      )}
    </div>
  )
}
