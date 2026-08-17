import { useMemo, useRef, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown, Forward, Server, ShieldCheck } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import { useConectionKey } from '@/features/integrations/hooks/useConnectionKey'
import { FlowNode, FlowEdge } from '@/shared/components/ui/flow-diagram'
import { RemoteEnablePanel, type RemoteEnableSelection } from './RemoteEnablePanel'
import { availableProtosFor, defaultPortFor, type Proto } from './protoCatalog'

// Reusable scaffolding for any "device → Forwarder → UTMStack" (syslog-style)
// integration. It renders the shared parts — a plain-language intro, an animated
// data-flow diagram, the connection details, the remote enable/disable panel,
// and a reactive "do it via CLI instead" fallback — while each collector
// injects only its vendor-specific device-side config as children.
//
// Usage:
//   <ForwarderGuide source="AIX server" port="7016" sourceType="ibm-aix">
//     <Section title="Send your AIX logs…"> …config… </Section>
//   </ForwarderGuide>

const SHARED = 'integrations.setup.collector.forwarder'
const LOGINPUT_PORT = '50052'

// The Forwarder runs on the UTMStack host; default the address to where the user
// is browsing from (stripped of any port).
export function forwarderHost(): string {
  if (typeof window === 'undefined') return 'utmstack-host'
  const h = window.location.host
  return h.includes(':') ? h.split(':')[0] : h
}

// ── Animated data-flow diagram ────────────────────────────────────────────────
// TONES/FlowNode/FlowEdge now live in shared/components/ui/flow-diagram.tsx
// (reused by the parsing-filters/alerting-rules pipeline diagram). Re-exported
// here so the many collector guides importing them from this module path
// don't need to change their imports.
export { TONES, FlowNode, FlowEdge } from '@/shared/components/ui/flow-diagram'

function FlowDiagram({ source, port, isMaster }: { source: string; port: string; isMaster: boolean }) {
  const { t } = useTranslation()
  return (
    <div className="mt-3 flex items-stretch justify-center gap-1 sm:gap-2">
      <FlowNode
        icon={<Server size={20} className="text-muted-foreground" />}
        title={source}
        sub={t(`${SHARED}.diagram.sourceSub`)}
        tone="neutral"
      />
      <FlowEdge label={t(`${SHARED}.diagram.flow1`)} />
      {!isMaster && (
        <>
          <FlowNode
            icon={<Forward size={20} />}
            title={t(`${SHARED}.diagram.forwarder`)}
            sub={t(`${SHARED}.diagram.forwarderSub`, { port })}
            tone="accent"
          />
          <FlowEdge label={t(`${SHARED}.diagram.flow2`)} />
        </>
      )}
      <FlowNode
        icon={<ShieldCheck size={20} />}
        title={t(`${SHARED}.diagram.utmstack`)}
        sub={t(`${SHARED}.diagram.utmstackSub`)}
        tone="brand"
      />
    </div>
  )
}

// ── Master command (POST endpoint + auth header) ─────────────────────────────

function MasterCommandSection({ selection }: { selection: RemoteEnableSelection }) {
  const { t } = useTranslation()
  const [tab, setTab] = useState<'simple' | 'batch'>('simple')
  if (!selection.apiKey) return null
  const host = typeof window === 'undefined' ? 'utmstack-host' : window.location.host
  const simpleCmd = `curl -k -X POST https://${host}:${LOGINPUT_PORT}/v1/ingest \\
  -H "Content-Type: application/json" \\
  -H "Utm-Api-Key: <YOUR_API_KEY>" \\
  -d '{
    "dataType": "<data-type>",
    "dataSource": "<data-source>",
    "timestamp": "",
    "raw": "<raw-log>"
  }'`
  const batchCmd = `curl -X POST https://${host}:${LOGINPUT_PORT}/v1/ingest \\
  -H "Content-Type: application/json" \\
  -H "Utm-Api-Key: <YOUR_API_KEY>" \\
  -d '{
    "logs": [
      {
        "dataType": "",
        "dataSource": "",
        "timestamp": "",
        "raw": ""
      },
      {
        "dataType": "",
        "dataSource": "",
        "timestamp": "",
        "raw": ""
      }
    ]
  }'`
  return (
    <Section title={t(`${SHARED}.masterHeader.title`)} step={2}>
      <p className="text-sm text-foreground/90">{t(`${SHARED}.masterHeader.body`, { name: selection.apiKey.name })}</p>
      <div className="flex gap-0 border-b border-border mb-3">
        {(['simple', 'batch'] as const).map((v) => (
          <button
            key={v}
            onClick={() => setTab(v)}
            className={cn('px-4 py-1.5 text-sm capitalize transition-colors', tab === v ? 'border-b-2 border-primary font-medium' : 'text-muted-foreground hover:text-foreground')}
          >
            {v === 'simple' ? t(`${SHARED}.masterHeader.tabSimple`) : t(`${SHARED}.masterHeader.tabBatch`)}
          </button>
        ))}
      </div>
      <CodeBlock code={tab === 'simple' ? simpleCmd : batchCmd} />
    </Section>
  )
}

// ── Manual (CLI) command — collapsible, reactive to the RemoteEnablePanel ────
// selection above it. Replaces what used to be a fixed "Optional — Enable TLS
// encryption" block: the command shown here always matches whatever
// proto/port the user currently has selected, TLS/HTTPS included, instead of
// duplicating a separate static TLS section.

function ManualCommandSection({ sourceType, selection }: { sourceType: string; selection: RemoteEnableSelection }) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  const needsCerts = selection.proto === 'tls' || selection.proto === 'https'
  const loadCertsCmd = `/opt/utmstack-forwarder/utmstack_forwarder load-tls-certs /path/to/cert.crt /path/to/key.key`
  const enableCmd = `sudo /opt/utmstack-forwarder/utmstack_forwarder enable-integration ${sourceType} ${selection.port} ${selection.proto}`

  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center justify-between px-4 py-3 text-left text-sm font-medium hover:bg-muted/40 transition-colors"
      >
        <span>{t(`${SHARED}.manualCommand.title`)}</span>
        <ChevronDown size={14} className={cn('shrink-0 text-muted-foreground transition-transform duration-200', open && 'rotate-180')} />
      </button>
      {open && (
        <div className="space-y-3 border-t border-border px-4 pb-4 pt-3">
          <p className="text-sm text-foreground/90">{t(`${SHARED}.manualCommand.intro`)}</p>
          {needsCerts && (
            <>
              <p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
                {t(`${SHARED}.manualCommand.certStep`)}
              </p>
              <CodeBlock code={loadCertsCmd} />
              <p className="rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
                {t(`${SHARED}.tls.step1Note`)}
              </p>
            </>
          )}
          <p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
            {t(`${SHARED}.manualCommand.enableStep`)}
          </p>
          <CodeBlock code={enableCmd} />
        </div>
      )}
    </div>
  )
}

// ── Forwarder install section ─────────────────────────────────────────────────

function buildForwarderInstallCmd(host: string, token: string): string {
  return `sudo bash -c "
  apt update -y && apt install wget -y && \\
  mkdir -p /opt/utmstack-forwarder && \\
  wget --no-check-certificate -P /opt/utmstack-forwarder \\
    https://${host}:9001/private/dependencies/collector/forwarder/utmstack_forwarder && \\
  chmod 755 /opt/utmstack-forwarder/utmstack_forwarder && \\
  /opt/utmstack-forwarder/utmstack_forwarder install ${host} <secret>${token}</secret> yes
"`
}

export function ForwarderInstall({ source }: { source: string }) {
  const { t } = useTranslation()
  const { key } = useConectionKey()
  const token = key.data?.connectionKey ?? '*'.repeat(7)
  const installCmd = buildForwarderInstallCmd(forwarderHost(), token)

  return (
    <Section title={t(`${SHARED}.install.title`)} step={1}>
      <p className="mb-2 text-sm text-foreground/90">{t(`${SHARED}.install.body`)}</p>
      <CodeBlock code={installCmd} />
      <p className="mt-2 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
        {t(`${SHARED}.install.reuse`, { source })}
      </p>
    </Section>
  )
}

// ── Template ──────────────────────────────────────────────────────────────────

interface ForwarderGuideProps {
  /** Already-translated device label, e.g. "AIX server" (used in the diagram + intro). */
  source: string
  /** Forwarder port for this data type, e.g. "7016" (illustrative default shown in the diagram). */
  port: string
  /** The dataType shown in the connection details, e.g. "ibm-aix". */
  sourceType: string
  /** Initial protocol, when the device needs something other than the catalog's first entry (e.g. ESXi/SentinelOne are TCP-only in practice). */
  defaultProto?: Proto
  /** Vendor-specific device-side config (how to point the device at the Forwarder). */
  children: ReactNode
}

export function ForwarderGuide({ source, port, sourceType, defaultProto, children }: ForwarderGuideProps) {
  const { t } = useTranslation()
  const availableProtos = useMemo(() => availableProtosFor(sourceType), [sourceType])
  const initialProto = defaultProto ?? availableProtos[0]
  const [selection, setSelection] = useState<RemoteEnableSelection>(() => ({
    proto: initialProto,
    port: defaultPortFor(sourceType, initialProto) || port,
    isMaster: false,
    apiKey: null,
  }))
  const [installOpen, setInstallOpen] = useState(false)
  const installRef = useRef<HTMLDivElement>(null)

  const handleAddCollector = () => {
    setInstallOpen(true)
    setTimeout(() => installRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 50)
  }

  return (
    <div className="space-y-4">
      <Section title={t(`${SHARED}.intro.title`)}>
        <p className="text-sm text-foreground/90">{t(`${SHARED}.intro.body`, { source })}</p>
        <FlowDiagram source={source} port={port} isMaster={selection.isMaster} />
        <p className="mt-3 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
          {t(`${SHARED}.noInstall`, { source })}
        </p>
      </Section>

      <RemoteEnablePanel
        dataType={sourceType}
        availableProtos={availableProtos}
        defaultProto={defaultProto}
        step={1}
        onSelectionChange={setSelection}
        onRequestAddCollector={handleAddCollector}
      />

      {selection.isMaster && selection.apiKey && <MasterCommandSection selection={selection} />}

      {/* Vendor-specific steps (device-side config). Hidden in master mode — it targets the forwarder, not master. */}
      {!selection.isMaster && children}

      {!selection.isMaster && (
        <div ref={installRef}>
          <ForwarderInstallSection source={source} open={installOpen} onToggle={() => setInstallOpen((o) => !o)} />
        </div>
      )}
      {!selection.isMaster && <ManualCommandSection sourceType={sourceType} selection={selection} />}
      {!selection.isMaster && <ForwarderUninstallSection />}
    </div>
  )
}

function ForwarderInstallSection({ source, open, onToggle }: { source: string; open: boolean; onToggle: () => void }) {
  const { t } = useTranslation()
  const { key } = useConectionKey()

  const token = key.data?.connectionKey ?? '*'.repeat(7)
  const installCmd = buildForwarderInstallCmd(forwarderHost(), token)

  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <button
        onClick={onToggle}
        className="flex w-full items-center justify-between px-4 py-3 text-left text-sm font-medium hover:bg-muted/40 transition-colors"
      >
        <span>{t(`${SHARED}.install.title`)}</span>
        <ChevronDown size={14} className={cn('shrink-0 text-muted-foreground transition-transform duration-200', open && 'rotate-180')} />
      </button>
      {open && (
        <div className="space-y-3 border-t border-border px-4 pb-4 pt-3">
          <p className="text-sm text-foreground/90">{t(`${SHARED}.install.body`)}</p>
          <CodeBlock code={installCmd} />
          <p className="rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
            {t(`${SHARED}.install.reuse`, { source })}
          </p>
        </div>
      )}
    </div>
  )
}

export function ForwarderUninstallSection() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  const uninstallCmd = `sudo bash -c "
  /opt/utmstack-forwarder/utmstack_forwarder uninstall && \\
  rm -rf /opt/utmstack-forwarder
"`

  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center justify-between px-4 py-3 text-left text-sm font-medium hover:bg-muted/40 transition-colors"
      >
        <span>{t(`${SHARED}.uninstall.title`)}</span>
        <ChevronDown size={14} className={cn('shrink-0 text-muted-foreground transition-transform duration-200', open && 'rotate-180')} />
      </button>
      {open && (
        <div className="space-y-3 border-t border-border px-4 pb-4 pt-3">
          <p className="text-sm text-foreground/90">{t(`${SHARED}.uninstall.body`)}</p>
          <CodeBlock code={uninstallCmd} />
        </div>
      )}
    </div>
  )
}
