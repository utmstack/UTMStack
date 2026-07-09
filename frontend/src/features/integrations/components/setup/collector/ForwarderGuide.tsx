import { useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown, Forward, Server, ShieldCheck } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import { useConectionKey } from '@/features/integrations/hooks/useConnectionKey'

// Reusable scaffolding for any "device → Forwarder → UTMStack" (syslog-style)
// integration. It renders the shared parts — a plain-language intro, an animated
// data-flow diagram, the connection details, and the "turn it on" step — while
// each collector injects only its vendor-specific "Step 1" (how to point the
// device at the Forwarder) as children.
//
// Usage:
//   <ForwarderGuide source="AIX server" port="7016" sourceType="ibm-aix">
//     <Section title="Send your AIX logs…"> …config… </Section>
//   </ForwarderGuide>

const SHARED = 'integrations.setup.collector.forwarder'

// The Forwarder runs on the UTMStack host; default the address to where the user
// is browsing from (stripped of any port).
export function forwarderHost(): string {
  if (typeof window === 'undefined') return 'utmstack-host'
  const h = window.location.host
  return h.includes(':') ? h.split(':')[0] : h
}

// ── Animated data-flow diagram ────────────────────────────────────────────────

export const TONES = {
  neutral: 'border-border bg-card',
  accent: 'border-primary/40 bg-primary/[0.06] text-primary',
  brand: 'border-emerald-500/40 bg-emerald-500/[0.06] text-emerald-600 dark:text-emerald-400',
} as const

export function FlowNode({
  icon,
  title,
  sub,
  tone,
}: {
  icon: ReactNode
  title: string
  sub: string
  tone: keyof typeof TONES
}) {
  return (
    <div
      className={`flex w-[92px] shrink-0 flex-col items-center gap-1 rounded-lg border p-2.5 text-center sm:w-[112px] ${TONES[tone]}`}
    >
      {icon}
      <span className="text-[11px] font-semibold leading-tight text-foreground">{title}</span>
      <span className="text-[9px] leading-tight text-muted-foreground">{sub}</span>
    </div>
  )
}

// A connector with little packets flowing left→right (SVG animateMotion, no deps).
export function FlowEdge({ label }: { label: string }) {
  return (
    <div className="flex shrink-0 flex-col items-center justify-center">
      <span className="mb-1 text-[8px] font-medium uppercase tracking-wide text-muted-foreground">
        {label}
      </span>
      <svg viewBox="0 0 64 12" className="h-3 w-10 sm:w-16" aria-hidden="true">
        <line
          x1="2"
          y1="6"
          x2="62"
          y2="6"
          className="stroke-border"
          strokeWidth="2"
          strokeDasharray="2 3"
          strokeLinecap="round"
        />
        <circle r="2.6" className="fill-primary">
          <animateMotion dur="1.6s" repeatCount="indefinite" path="M2 6 H62" />
        </circle>
        <circle r="2.6" className="fill-primary">
          <animateMotion dur="1.6s" begin="0.8s" repeatCount="indefinite" path="M2 6 H62" />
        </circle>
      </svg>
    </div>
  )
}

function FlowDiagram({ source, port }: { source: string; port: string }) {
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
      <FlowNode
        icon={<Forward size={20} />}
        title={t(`${SHARED}.diagram.forwarder`)}
        sub={t(`${SHARED}.diagram.forwarderSub`, { port })}
        tone="accent"
      />
      <FlowEdge label={t(`${SHARED}.diagram.flow2`)} />
      <FlowNode
        icon={<ShieldCheck size={20} />}
        title={t(`${SHARED}.diagram.utmstack`)}
        sub={t(`${SHARED}.diagram.utmstackSub`)}
        tone="brand"
      />
    </div>
  )
}

// ── TLS collapsible section ───────────────────────────────────────────────────

function TLSSection({ sourceType, port }: { sourceType: string; port: string }) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  const loadCmd = `/opt/utmstack-forwarder/utmstack_forwarder load-tls-certs /path/to/cert.crt /path/to/key.key`
  const enableCmd = `/opt/utmstack-forwarder/utmstack_forwarder enable-integration ${sourceType} ${port} tls`

  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center justify-between px-4 py-3 text-left text-sm font-medium hover:bg-muted/40 transition-colors"
      >
        <span>{t(`${SHARED}.tls.title`)}</span>
        <ChevronDown size={14} className={cn('shrink-0 text-muted-foreground transition-transform duration-200', open && 'rotate-180')} />
      </button>
      {open && (
        <div className="space-y-3 border-t border-border px-4 pb-4 pt-3">
          <p className="text-sm text-foreground/90">{t(`${SHARED}.tls.intro`)}</p>
          <p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">{t(`${SHARED}.tls.step1`)}</p>
          <CodeBlock code={loadCmd} />
          <p className="rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">{t(`${SHARED}.tls.step1Note`)}</p>
          <p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">{t(`${SHARED}.tls.step2`)}</p>
          <CodeBlock code={enableCmd} />
        </div>
      )}
    </div>
  )
}

// ── HTTPS certificates step ───────────────────────────────────────────────────

// Required (not collapsible) cert-loading step for collectors whose listener is
// natively HTTPS — same command CustomSetup shows for the tls protocol.
export function HttpsCertsSection({ step, port }: { step: number; port: string }) {
  const { t } = useTranslation()
  const loadCmd = `/opt/utmstack-forwarder/utmstack_forwarder load-tls-certs /path/to/cert.crt /path/to/key.key`

  return (
    <Section title={t(`${SHARED}.httpsCerts.title`)} step={step}>
      <p className="mb-2 text-sm text-foreground/90">{t(`${SHARED}.httpsCerts.body`, { port })}</p>
      <CodeBlock code={loadCmd} />
      <p className="mt-2 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
        {t(`${SHARED}.tls.step1Note`)}
      </p>
    </Section>
  )
}

// ── Forwarder install section ─────────────────────────────────────────────────

export function ForwarderInstall({ source }: { source: string }) {
  const { t } = useTranslation()
  const { key } = useConectionKey()
  const host = forwarderHost()

  // Token is always shown as •••••••  in the UI; the real value is copied on click.
  // Falls back to "*******" if the key hasn't loaded or the request failed.
  const token = key.data?.connectionKey ?? '*'.repeat(7)
  const installCmd = `sudo bash -c "
  apt update -y && apt install wget -y && \\
  mkdir -p /opt/utmstack-forwarder && \\
  wget --no-check-certificate -P /opt/utmstack-forwarder \\
    https://${host}:9001/private/dependencies/collector/forwarder/utmstack_forwarder && \\
  chmod 755 /opt/utmstack-forwarder/utmstack_forwarder && \\
  /opt/utmstack-forwarder/utmstack_forwarder install ${host} <secret>${token}</secret> yes
"`

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
  /** Forwarder port for this data type, e.g. "7016". */
  port: string
  /** OpenSearch source_type / dataType shown in the connection details, e.g. "ibm-aix". */
  sourceType: string
  /** Vendor-specific "Step 1" — how to point the device at the Forwarder. */
  children: ReactNode
  /** Hide the TLS collapsible section (e.g. for integrations that are natively HTTPS). */
  hideTLS?: boolean
}

export function ForwarderGuide({ source, port, sourceType, children, hideTLS }: ForwarderGuideProps) {
  const { t } = useTranslation()

  return (
    <div className="space-y-4">
      <Section title={t(`${SHARED}.intro.title`)}>
        <p className="text-sm text-foreground/90">{t(`${SHARED}.intro.body`, { source })}</p>
        <FlowDiagram source={source} port={port} />
        <p className="mt-3 rounded-md bg-muted/40 px-3 py-2 text-[11px] text-muted-foreground">
          {t(`${SHARED}.noInstall`, { source })}
        </p>
      </Section>

      <ForwarderInstall source={source} />

      {/* Vendor-specific steps (enable listener + device-side config). */}
      {children}

      {!hideTLS && <TLSSection sourceType={sourceType} port={port} />}
      <ForwarderUninstallSection />
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
