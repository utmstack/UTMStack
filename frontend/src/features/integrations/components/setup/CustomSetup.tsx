import { useState } from 'react'
import { useTranslation, Trans } from 'react-i18next'
import { Forward, Server, ShieldCheck } from 'lucide-react'
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
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.custom'

type Proto = 'udp' | 'tcp' | 'tls' | 'http' | 'https'

const PROTOCOLS: { id: Proto; label: string }[] = [
  { id: 'udp', label: 'Syslog · UDP' },
  { id: 'tcp', label: 'Syslog · TCP' },
  { id: 'tls', label: 'Syslog · TLS' },
  { id: 'http', label: 'HTTP' },
  { id: 'https', label: 'HTTPS' },
]

const FWD = '/opt/utmstack-forwarder/utmstack_forwarder'

export function CustomSetup({ integration }: { integration: Integration }) {
  const { t } = useTranslation()
  const host = forwarderHost()

  // The forwarder data-type for this custom integration. Falls back to a sample
  // slug if the catalog row has no explicit data type.
  const name = integration.dataType || integration.moduleName?.toLowerCase() || 'my-integration'

  const [proto, setProto] = useState<Proto>('udp')
  const [port, setPort] = useState('7100')

  const isHttp = proto === 'http' || proto === 'https'
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

      {/* Step 1: install the Forwarder (shared). */}
      <ForwarderInstall source={integration.name} />

      {/* Step 2: pick protocol + port → activation command. */}
      <Section title={t(`${ROOT}.choose.title`)} step={2}>
        <p className="mb-3 text-sm text-foreground/90">{t(`${ROOT}.choose.body`)}</p>

        {/* Protocol tabs */}
        <div className="mb-3 flex flex-wrap gap-1.5">
          {PROTOCOLS.map((p) => (
            <button
              key={p.id}
              type="button"
              onClick={() => setProto(p.id)}
              className={cn(
                'rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
                p.id === proto
                  ? 'bg-primary text-primary-foreground shadow-sm'
                  : 'bg-muted text-muted-foreground hover:bg-muted/70 hover:text-foreground',
              )}
            >
              {p.label}
            </button>
          ))}
        </div>

        {/* Port input */}
        <label className="mb-3 block">
          <span className="mb-1 block text-[11px] font-medium text-muted-foreground">
            {t(`${ROOT}.choose.portLabel`)}
          </span>
          <input
            value={port}
            onChange={(e) => setPort(e.target.value.replace(/[^0-9]/g, ''))}
            inputMode="numeric"
            className="h-9 w-32 rounded-md border border-border bg-background px-3 font-mono text-sm"
            placeholder="7100"
          />
        </label>

        {/* TLS and HTTPS need certificates loaded first. */}
        {(proto === 'tls' || proto === 'https') && (
          <div className="mb-3">
            <p className="mb-1.5 text-[11px] font-medium text-muted-foreground">{t(`${ROOT}.choose.tlsFirst`)}</p>
            <CodeBlock code={tlsCertsCmd} />
          </div>
        )}

        <p className="mb-1.5 text-[11px] font-medium text-muted-foreground">{t(`${ROOT}.choose.enableLabel`)}</p>
        <CodeBlock code={enableCmd} />

        {isHttp && (
          <p className="mt-2 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
            <Trans
              i18nKey={`${ROOT}.choose.authNote`}
              components={{ code: <code className="font-mono text-foreground/80" /> }}
            />
          </p>
        )}
      </Section>

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

      <ForwarderUninstallSection />
    </div>
  )
}
