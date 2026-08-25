import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { AlertTriangle, ChevronDown, Boxes, Download, ShieldCheck } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { registerCollector } from '../registry'
import { Section } from '@/features/integrations/components/ui/Section'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'
import { useConectionKey } from '@/features/integrations/hooks/useConnectionKey'
import { forwarderHost, FlowNode, FlowEdge } from '../ForwarderGuide'
import type { Integration } from '@/features/integrations/types'

const ROOT = 'integrations.setup.collector.utmstack'

// The self-monitoring collector installs the `utmstack_collector` binary on the
// UTMStack host itself, where it tails the platform's own Docker container logs.
// It only runs on Ubuntu or Red Hat-family hosts (the two supported install OSes).
const OSES = [
  { id: 'ubuntu', label: 'Ubuntu', pkg: 'apt' },
  { id: 'redhat', label: 'CentOS / Red Hat', pkg: 'dnf' },
] as const

type OsId = (typeof OSES)[number]['id']

function installCmd(pkg: string, host: string, token: string): string {
  return `sudo bash -c "
  ${pkg} update -y && ${pkg} install wget -y && \\
  mkdir -p /opt/utmstack-collector && \\
  wget --no-check-certificate -P /opt/utmstack-collector \\
    https://${host}/private/dependencies/collector/utmstack_collector && \\
  chmod 755 /opt/utmstack-collector/utmstack_collector && \\
  /opt/utmstack-collector/utmstack_collector install ${host} <secret>${token}</secret> yes
"`
}

const UNINSTALL_CMD = `sudo bash -c "
  /opt/utmstack-collector/utmstack_collector uninstall && \\
  (systemctl stop UTMStackCollector 2>/dev/null || true) && \\
  (systemctl disable UTMStackCollector 2>/dev/null || true) && \\
  rm -rf /opt/utmstack-collector && \\
  rm -f /etc/systemd/system/UTMStackCollector.service && \\
  (systemctl daemon-reload 2>/dev/null || true)
"`

function OsTabs({ active, onSelect }: { active: OsId; onSelect: (id: OsId) => void }) {
  return (
    <div className="mb-3 flex gap-1.5">
      {OSES.map((os) => (
        <button
          key={os.id}
          onClick={() => onSelect(os.id)}
          className={cn(
            'rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
            active === os.id
              ? 'bg-primary/15 text-primary ring-1 ring-inset ring-primary/30'
              : 'text-muted-foreground hover:bg-muted/60',
          )}
        >
          {os.label}
        </button>
      ))}
    </div>
  )
}

function UninstallSection() {
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

function UtmstackGuide({ module: _module }: { module: Integration }) {
  const { t } = useTranslation()
  const { key } = useConectionKey()
  const host = forwarderHost()
  const token = key.data?.connectionKey ?? '*'.repeat(7)

  const [os, setOs] = useState<OsId>('ubuntu')
  const pkg = OSES.find((o) => o.id === os)?.pkg ?? 'apt'

  return (
    <div className="space-y-4">
      <Section title={t(`${ROOT}.intro.title`)}>
        <p className="text-sm text-foreground/90">{t(`${ROOT}.intro.body`)}</p>
        <div className="mt-3 flex items-stretch justify-center gap-1 sm:gap-2">
          <FlowNode
            icon={<Boxes size={20} className="text-muted-foreground" />}
            title={t(`${ROOT}.intro.diagram.source`)}
            sub={t(`${ROOT}.intro.diagram.sourceSub`)}
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
      </Section>

      <Section title={t(`${ROOT}.requirements.title`)} variant="warning">
        <p className="text-sm text-foreground/90">{t(`${ROOT}.requirements.body`)}</p>
      </Section>

      <Section title={t(`${ROOT}.install.title`)} step={1}>
        <p className="mb-3 text-sm text-foreground/90">{t(`${ROOT}.install.body`)}</p>
        <div className="mb-3 flex gap-2 rounded-md border border-amber-500/30 bg-amber-500/[0.07] px-3 py-2 text-[11px] text-amber-700 dark:text-amber-300">
          <AlertTriangle size={14} className="mt-0.5 shrink-0" />
          <span>{t(`${ROOT}.install.sensitiveWarning`)}</span>
        </div>
        <OsTabs active={os} onSelect={setOs} />
        <CodeBlock lang="bash" code={installCmd(pkg, host, token)} />
      </Section>

      <Section title={t(`${ROOT}.after.title`)}>
        <ul className="list-disc space-y-1 pl-5 text-sm">
          <li>{t(`${ROOT}.after.item1`)}</li>
          <li>{t(`${ROOT}.after.item2`)}</li>
          <li>{t(`${ROOT}.after.item3`)}</li>
        </ul>
      </Section>

      <UninstallSection />
    </div>
  )
}

registerCollector({
  getName: () => 'UTMSTACK',
  matches: (n) => n === 'utmstack',
  sections: [],
  render: (m) => <UtmstackGuide module={m} />,
})
