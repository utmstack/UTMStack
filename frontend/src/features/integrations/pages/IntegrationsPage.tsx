import { useMemo, useState } from 'react'
import {
  Check,
  Copy,
  ExternalLink,
  Filter,
  Plus,
  Search,
  Settings as SettingsIcon,
  X,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

/* ─── Types ────────────────────────────────────────────────────────────── */

type DeployKind = 'agent' | 'collector' | 'cloud' | 'custom'
type Status = 'configured' | 'available'

interface Integration {
  id: string
  name: string
  kind: DeployKind
  status: Status
  description: string
  category: string
  logo: string
  // when true, the logo is monochrome dark and needs inverting in dark mode
  darkInvert?: boolean
  // collector-specific
  defaultPort?: string
  // cloud-specific
  cloudFields?: { label: string; placeholder: string; secret?: boolean }[]
  // counts when configured
  events24h?: number
  rate?: number
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const LOGO = (slug: string) => `/integrations/${slug}`

const INTEGRATIONS: Integration[] = [
  // ─── Agents ────────────────────────────────────────────────────────────
  { id: 'agent-windows', name: 'Windows', kind: 'agent', status: 'configured', category: 'OS / Endpoint', logo: LOGO('windows.svg'),
    description: 'Lightweight Windows agent that forwards Sysmon, Security, Application, and File Audit events to UTMStack. Auto-updates and reports its own health back to the panel.',
    events24h: 142_004, rate: 142 },
  { id: 'agent-linux', name: 'Linux', kind: 'agent', status: 'configured', category: 'OS / Endpoint', logo: LOGO('linux.svg'),
    description: 'Linux agent (Ubuntu, RHEL, Debian, CentOS, SUSE). Tails /var/log, listens to auditd and journalctl, and ships normalized events. Single curl-bash command to install.',
    events24h: 312_000, rate: 248 },
  { id: 'agent-macos', name: 'macOS', kind: 'agent', status: 'available', category: 'OS / Endpoint', logo: LOGO('macos.svg'), darkInvert: true,
    description: 'macOS agent for endpoint telemetry. Pulls unified log entries and Endpoint Security Framework (ESF) events. Universal binary for Apple Silicon and Intel Macs.' },

  // ─── Collectors ────────────────────────────────────────────────────────
  { id: 'col-vmware', name: 'VMware', kind: 'collector', status: 'available', category: 'Virtualization', logo: LOGO('vmware.svg'), defaultPort: '514/udp',
    description: 'vCenter and ESXi syslog forwarding into the UTMStack collector. Captures vMotion events, host alarms, login activity, and configuration changes.' },
  { id: 'col-kaspersky', name: 'Kaspersky', kind: 'collector', status: 'available', category: 'EDR / Antivirus', logo: LOGO('kaspersky.svg'), defaultPort: '514/tcp',
    description: 'Kaspersky Security Center syslog forwarding, parsed by the collector. Captures detections, host status, and policy changes from your KSC console.' },
  { id: 'col-eset', name: 'ESET', kind: 'collector', status: 'available', category: 'EDR / Antivirus', logo: LOGO('eset.svg'), defaultPort: '6514/tcp',
    description: 'ESET PROTECT (formerly ESMC) syslog forwarding over TLS. Detection events, scan results, and endpoint status appear under Log Explorer in real time.' },
  { id: 'col-as400', name: 'IBM AS/400', kind: 'collector', status: 'available', category: 'Mainframe', logo: LOGO('as400.svg'), defaultPort: '514/udp',
    description: 'iSeries / AS/400 audit journal forwarding via QAUDJRN. Captures user authentications, command history, object access, and system value changes.' },
  { id: 'col-oracle', name: 'Oracle Database', kind: 'collector', status: 'available', category: 'Database', logo: LOGO('oracle.svg'), defaultPort: '514/udp',
    description: 'Oracle audit trail (UNIFIED_AUDIT_TRAIL) shipped via syslog. Privileged-user activity, DDL/DML, login attempts, and policy violations end up in UTMStack.' },
  { id: 'col-mikrotik', name: 'MikroTik', kind: 'collector', status: 'available', category: 'Network', logo: LOGO('mikrotik.svg'), defaultPort: '514/udp', darkInvert: true,
    description: 'MikroTik RouterOS syslog forwarding for firewall, NAT, DHCP, and traffic events. Compatible with RouterBOARD and CHR. Tested against RouterOS 6 and 7.' },
  { id: 'col-paloalto', name: 'Palo Alto Networks', kind: 'collector', status: 'configured', category: 'Network', logo: LOGO('paloalto.svg'), defaultPort: '514/udp',
    description: 'PAN-OS firewall log forwarding for Traffic, Threat, URL, and Authentication events. Supports PA-Series, VM-Series, and Panorama. CEF and LEEF formats supported.',
    events24h: 884_204, rate: 882 },
  { id: 'col-sonicwall', name: 'SonicWall', kind: 'collector', status: 'available', category: 'Network', logo: LOGO('sonicwall.svg'), defaultPort: '514/udp',
    description: 'SonicOS firewall syslog forwarding for traffic, threat, intrusion, and authentication events. Compatible with TZ, NSa, and NSv appliances.' },
  { id: 'col-suricata', name: 'Suricata IDS', kind: 'collector', status: 'configured', category: 'IDS / IPS', logo: LOGO('suricata.png'), defaultPort: '514/tcp',
    description: 'Suricata EVE JSON output piped into the UTMStack collector. Captures alerts, flow records, anomaly events, file metadata, and TLS/SSH/DNS protocol logs.',
    events24h: 88_201, rate: 88 },
  { id: 'col-cisco-asa', name: 'Cisco ASA', kind: 'collector', status: 'available', category: 'Network', logo: LOGO('cisco.svg'), defaultPort: '514/udp', darkInvert: true,
    description: 'Cisco ASA firewall syslog with rich message normalization — ACL hits, VPN tunnels, NAT events, and authentication. Tested with ASA 9.x and Firepower 1000/2100 in ASA mode.' },
  { id: 'col-cisco-meraki', name: 'Cisco Meraki', kind: 'collector', status: 'available', category: 'Network', logo: LOGO('meraki.svg'), defaultPort: '514/udp',
    description: 'Meraki Dashboard syslog forwarding — Security, Flows, Air Marshal, and Events. Configure once per network in the Dashboard; UTMStack handles parsing automatically.' },
  { id: 'col-cisco-switch', name: 'Cisco Switch', kind: 'collector', status: 'available', category: 'Network', logo: LOGO('cisco.svg'), defaultPort: '514/udp', darkInvert: true,
    description: 'IOS / IOS-XE Catalyst switch syslog forwarding. Captures port-security violations, link state, AAA authentication, spanning-tree topology changes, and ACL drops.' },
  { id: 'col-fortigate', name: 'FortiGate', kind: 'collector', status: 'available', category: 'Network', logo: LOGO('fortigate.svg'), defaultPort: '514/udp',
    description: 'FortiOS firewall syslog (Traffic, UTM, IPS, Web Filter, Antivirus, App Control). CEF and Fortinet native formats supported. Compatible with FortiGate 6.x, 7.x.' },
  { id: 'col-sophos-firewall', name: 'Sophos Firewall', kind: 'collector', status: 'available', category: 'Network', logo: LOGO('sophos-firewall.svg'), defaultPort: '514/udp',
    description: 'Sophos XG / SG / XGS firewall syslog forwarding. Includes firewall, IPS, web filter, application control, ATP, and authentication events.' },
  { id: 'col-aix', name: 'IBM AIX', kind: 'collector', status: 'available', category: 'Unix', logo: LOGO('aix.svg'), defaultPort: '514/udp', darkInvert: true,
    description: 'IBM AIX syslog and errpt audit forwarding. Captures system events, hardware errors, security audit subsystem records, and user authentication. Tested with AIX 7.x.' },
  { id: 'col-pfsense', name: 'pfSense', kind: 'collector', status: 'configured', category: 'Network', logo: LOGO('pfsense.svg'), defaultPort: '514/udp',
    description: 'pfSense and OPNsense firewall syslog forwarding (filterlog, OpenVPN, dhcpd, IPSec, Suricata package). Auto-parsed by the collector with no extra configuration.',
    events24h: 1_204_882, rate: 1208 },
  { id: 'col-fortiweb', name: 'FortiWeb', kind: 'collector', status: 'available', category: 'Web Application Firewall', logo: LOGO('fortiweb.svg'), defaultPort: '514/udp',
    description: 'FortiWeb Web Application Firewall syslog forwarding. Includes WAF attack signatures, bot detections, machine-learning anomaly events, and policy violations.' },
  { id: 'col-firepower', name: 'Cisco Firepower', kind: 'collector', status: 'available', category: 'Network', logo: LOGO('firepower.svg'), defaultPort: '514/tcp',
    description: 'Cisco Firepower NGIPS / NGFW. Supports both eStreamer (binary) and syslog. Connection events, intrusion alerts, malware/file events, and security intelligence.' },
  { id: 'col-netflow', name: 'NetFlow / IPFIX', kind: 'collector', status: 'configured', category: 'Network', logo: LOGO('netflow.svg'), defaultPort: '2055/udp',
    description: 'NetFlow v5 / v9, IPFIX, and sFlow ingestion via the UTMStack collector. Provides full network flow visibility for traffic analysis and threat detection.',
    events24h: 4_220_044, rate: 4220 },
  { id: 'col-deceptive', name: 'Deceptive Bytes', kind: 'collector', status: 'available', category: 'Deception', logo: LOGO('deceptive-bytes.svg'), defaultPort: '514/tcp',
    description: 'Deceptive Bytes active deception platform alerts via syslog. Honeypot triggers, decoy access, and lateral-movement attempts surface immediately as critical alerts.' },
  { id: 'col-sentinelone', name: 'SentinelOne', kind: 'collector', status: 'available', category: 'EDR / Antivirus', logo: LOGO('sentinelone.svg'), defaultPort: '6514/tcp',
    description: 'SentinelOne Singularity syslog forwarding over TLS. Threat detections, behavioral indicators, deep-visibility events, and remediation actions are normalized.' },

  // ─── Cloud / In-app ────────────────────────────────────────────────────
  { id: 'cloud-bitdefender', name: 'Bitdefender', kind: 'cloud', status: 'available', category: 'EDR / Antivirus', logo: LOGO('bitdefender.svg'),
    description: 'Bitdefender GravityZone Cloud integration. UTMStack pulls EDR detections, malware events, and policy violations every minute via the GravityZone API.',
    cloudFields: [
      { label: 'API URL', placeholder: 'https://cloud.gravityzone.bitdefender.com' },
      { label: 'API key', placeholder: '••••••••', secret: true },
    ] },
  { id: 'cloud-aws', name: 'AWS', kind: 'cloud', status: 'configured', category: 'Cloud', logo: LOGO('aws.svg'),
    description: 'CloudTrail, GuardDuty, VPC Flow Logs, IAM activity, RDS audit, ECS, and Lambda — all pulled through a single STS-assumed role. Multi-region and multi-account ready.',
    cloudFields: [
      { label: 'Account ID', placeholder: '123456789012' },
      { label: 'Role ARN', placeholder: 'arn:aws:iam::...:role/UTMStackReader' },
      { label: 'External ID', placeholder: '••••••••', secret: true },
      { label: 'Regions', placeholder: 'us-east-1, us-west-2' },
    ],
    events24h: 481_022, rate: 482 },
  { id: 'cloud-gcp', name: 'Google Cloud', kind: 'cloud', status: 'available', category: 'Cloud', logo: LOGO('gcp.svg'),
    description: 'Google Cloud Audit Logs and Pub/Sub topics. Captures admin activity, data access, system events, and policy denials across your GCP projects.',
    cloudFields: [
      { label: 'Project ID', placeholder: 'my-project-001' },
      { label: 'Service account JSON', placeholder: 'Paste JSON…', secret: true },
      { label: 'Pub/Sub subscription', placeholder: 'projects/.../subscriptions/utmstack' },
    ] },
  { id: 'cloud-azure', name: 'Microsoft Azure', kind: 'cloud', status: 'configured', category: 'Cloud', logo: LOGO('azure.svg'),
    description: 'Azure AD sign-in logs, Activity logs, and Defender for Cloud alerts via the Microsoft Graph API. Captures conditional access events, MFA prompts, and risky sign-ins.',
    cloudFields: [
      { label: 'Tenant ID', placeholder: '00000000-0000-0000-0000-000000000000' },
      { label: 'Client ID', placeholder: 'app registration id' },
      { label: 'Client secret', placeholder: '••••••••', secret: true },
    ],
    events24h: 124_001, rate: 124 },
  { id: 'cloud-m365', name: 'Microsoft 365', kind: 'cloud', status: 'configured', category: 'Cloud', logo: LOGO('m365.svg'),
    description: 'M365 Unified Audit Log — Exchange, SharePoint, OneDrive, Teams, and Defender alerts. Tracks file shares, admin actions, mailbox access, and external sharing in detail.',
    cloudFields: [
      { label: 'Tenant ID', placeholder: '00000000-0000-0000-0000-000000000000' },
      { label: 'Client ID', placeholder: 'app registration id' },
      { label: 'Client secret', placeholder: '••••••••', secret: true },
    ],
    events24h: 218_204, rate: 218 },
  { id: 'cloud-sophos-central', name: 'Sophos Central', kind: 'cloud', status: 'available', category: 'EDR / Antivirus', logo: LOGO('sophos-central.svg'),
    description: 'Sophos Central API integration covering Endpoint, Server, Encryption, Mobile, and Email Security alerts. Pulled every minute, normalized into the UTMStack alert schema.',
    cloudFields: [
      { label: 'Tenant ID', placeholder: 'sophos tenant id' },
      { label: 'Client ID', placeholder: 'API client id' },
      { label: 'Client secret', placeholder: '••••••••', secret: true },
    ] },
  { id: 'cloud-github', name: 'GitHub', kind: 'cloud', status: 'available', category: 'DevOps', logo: LOGO('github.svg'), darkInvert: true,
    description: 'GitHub audit log plus Code Scanning, Secret Scanning, and Dependabot alerts. Connects via OAuth app or fine-grained PAT. Works for individual repos or whole orgs.',
    cloudFields: [
      { label: 'Organization', placeholder: 'utmstack' },
      { label: 'Personal access token', placeholder: 'ghp_••••••••', secret: true },
    ] },

  // ─── Custom ────────────────────────────────────────────────────────────
  { id: 'custom', name: 'Custom syslog / JSON', kind: 'custom', status: 'available', category: 'Custom', logo: LOGO('custom.svg'), darkInvert: true,
    description: 'Forward any structured (JSON) or syslog source to UTMStack with a free-form parser. Drop-in for proprietary appliances, custom apps, or anything not in the catalog.' },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const KIND_META: Record<DeployKind, { label: string; tone: string; pill: string; dot: string }> = {
  agent: { label: 'Agent', tone: 'text-sky-500', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300', dot: 'bg-sky-500' },
  collector: { label: 'Collector', tone: 'text-violet-500', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300', dot: 'bg-violet-500' },
  cloud: { label: 'Cloud', tone: 'text-fuchsia-500', pill: 'bg-fuchsia-500/15 text-fuchsia-600 ring-fuchsia-500/30 dark:text-fuchsia-300', dot: 'bg-fuchsia-500' },
  custom: { label: 'Custom', tone: 'text-emerald-500', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300', dot: 'bg-emerald-500' },
}

/* ─── Tabs ─────────────────────────────────────────────────────────────── */

type Tab = 'all' | 'agent' | 'collector' | 'cloud' | 'custom'

const TABS: { id: Tab; label: string; predicate: (i: Integration) => boolean }[] = [
  { id: 'all', label: 'All', predicate: () => true },
  { id: 'agent', label: 'Agents', predicate: (i) => i.kind === 'agent' },
  { id: 'collector', label: 'Collectors', predicate: (i) => i.kind === 'collector' },
  { id: 'cloud', label: 'Cloud', predicate: (i) => i.kind === 'cloud' },
  { id: 'custom', label: 'Custom', predicate: (i) => i.kind === 'custom' },
]

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function IntegrationsPage() {
  const [tab, setTab] = useState<Tab>('all')
  const [search, setSearch] = useState('')
  const [showOnlyConfigured, setShowOnlyConfigured] = useState(false)
  const [open, setOpen] = useState<Integration | null>(null)

  const filtered = useMemo(() => {
    const t = TABS.find((x) => x.id === tab) ?? TABS[0]
    return INTEGRATIONS.filter(t.predicate)
      .filter((i) => (showOnlyConfigured ? i.status === 'configured' : true))
      .filter((i) =>
        search
          ? (i.name + i.description + i.category).toLowerCase().includes(search.toLowerCase())
          : true
      )
  }, [tab, search, showOnlyConfigured])

  const configuredCount = INTEGRATIONS.filter((i) => i.status === 'configured').length

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header configured={configuredCount} total={INTEGRATIONS.length} />

      <div className="mt-5 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[300px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search integrations…"
            className="h-10 pl-9"
          />
        </div>
        <button
          onClick={() => setShowOnlyConfigured((v) => !v)}
          className={cn(
            'flex h-10 items-center gap-2 rounded-md border px-3 text-sm transition-colors',
            showOnlyConfigured
              ? 'border-primary/30 bg-primary/10 text-primary'
              : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
          )}
        >
          <Filter size={14} />
          Configured only
        </button>
      </div>

      <div className="mt-5">
        <TabsRow current={tab} onChange={setTab} />
      </div>

      <div className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
        {filtered.map((i) => (
          <IntegrationCard key={i.id} integration={i} onOpen={() => setOpen(i)} />
        ))}
        {filtered.length === 0 && (
          <div className="col-span-full rounded-xl border border-dashed border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
            No integrations match.
          </div>
        )}
      </div>

      {open && <IntegrationDrawer integration={open} onClose={() => setOpen(null)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ configured, total }: { configured: number; total: number }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">Integrations</h1>
        <p className="mt-1 text-xs text-muted-foreground">
          <span className="font-medium text-foreground">{configured}</span> of{' '}
          <span className="font-medium text-foreground">{total}</span> integrations configured.
        </p>
      </div>
      <Button size="sm" variant="outline">
        <ExternalLink size={14} className="mr-2" />
        Browse marketplace
      </Button>
    </header>
  )
}

/* ─── Tabs ─────────────────────────────────────────────────────────────── */

function TabsRow({ current, onChange }: { current: Tab; onChange: (t: Tab) => void }) {
  return (
    <div className="flex flex-wrap items-center gap-1 border-b border-border">
      {TABS.map((t) => {
        const active = current === t.id
        const count = INTEGRATIONS.filter(t.predicate).length
        return (
          <button
            key={t.id}
            onClick={() => onChange(t.id)}
            className={cn(
              'group relative flex items-center gap-2 px-3 py-2 text-xs transition-colors',
              active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <span>{t.label}</span>
            <span
              className={cn(
                'rounded-md px-1.5 py-0.5 font-mono text-[10px] tabular-nums',
                active ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'
              )}
            >
              {count}
            </span>
            {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
          </button>
        )
      })}
    </div>
  )
}

/* ─── Card ─────────────────────────────────────────────────────────────── */

function IntegrationCard({
  integration: i,
  onOpen,
}: {
  integration: Integration
  onOpen: () => void
}) {
  const km = KIND_META[i.kind]
  return (
    <button
      onClick={onOpen}
      className="group relative flex h-full flex-col overflow-hidden rounded-xl border border-border bg-card text-left transition-all hover:border-foreground/20 hover:shadow-md"
    >
      {/* Logo strip — theme-aware muted background, big logo */}
      <div className="relative flex h-32 items-center justify-center border-b border-border bg-muted/40 p-6">
        <img
          src={i.logo}
          alt={i.name}
          className={cn(
            'max-h-full max-w-[65%] object-contain',
            i.darkInvert && 'dark:brightness-0 dark:invert'
          )}
          loading="lazy"
        />
        {/* Kind badge top-right */}
        <span
          className={cn(
            'absolute right-3 top-3 inline-flex items-center gap-1 rounded-full bg-card/80 px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-border backdrop-blur',
            km.tone
          )}
        >
          <span className={cn('inline-block h-1.5 w-1.5 rounded-full', km.dot)} />
          {km.label}
        </span>
      </div>

      {/* Content */}
      <div className="flex flex-1 flex-col p-4">
        <h4 className="text-sm font-semibold leading-tight">{i.name}</h4>
        <p className="mt-1.5 line-clamp-4 flex-1 text-[11px] leading-relaxed text-muted-foreground">
          {i.description}
        </p>

        <div className="mt-3 flex items-center justify-between border-t border-border pt-3 text-[11px]">
          {i.status === 'configured' ? (
            <span className="inline-flex items-center gap-1 font-medium text-emerald-600 dark:text-emerald-400">
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
              Configured
              {i.rate ? (
                <span className="ml-1 font-mono tabular-nums text-muted-foreground">
                  · {i.rate.toLocaleString()} e/s
                </span>
              ) : null}
            </span>
          ) : (
            <span className="text-muted-foreground">{i.category}</span>
          )}
          <span className="text-primary opacity-0 transition-opacity group-hover:opacity-100">
            {i.status === 'configured' ? 'Manage →' : 'Set up →'}
          </span>
        </div>
      </div>
    </button>
  )
}

function LogoTile({
  src,
  alt,
  size,
  configured,
  darkInvert,
}: {
  src: string
  alt: string
  size: number
  configured?: boolean
  darkInvert?: boolean
}) {
  return (
    <div
      className="relative flex shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border bg-muted/40 p-2.5"
      style={{ width: size, height: size }}
    >
      <img
        src={src}
        alt={alt}
        className={cn(
          'max-h-full max-w-full object-contain',
          darkInvert && 'dark:brightness-0 dark:invert'
        )}
        loading="lazy"
      />
      {configured && (
        <span
          className="absolute -bottom-0.5 -right-0.5 flex h-3.5 w-3.5 items-center justify-center rounded-full bg-emerald-500 ring-2 ring-card"
          aria-label="Configured"
        >
          <Check size={9} strokeWidth={3} className="text-white" />
        </span>
      )}
    </div>
  )
}

/* ─── Drawer ───────────────────────────────────────────────────────────── */

function IntegrationDrawer({
  integration: i,
  onClose,
}: {
  integration: Integration
  onClose: () => void
}) {
  const km = KIND_META[i.kind]
  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[820px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-5">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <LogoTile src={i.logo} alt={i.name} size={56} darkInvert={i.darkInvert} />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span>{i.category}</span>
                  <span>·</span>
                  <span className={km.tone}>{km.label}</span>
                </div>
                <h2 className="mt-0.5 truncate text-xl font-semibold">{i.name}</h2>
                <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{i.description}</p>
                <div className="mt-2 flex items-center gap-2">
                  {i.status === 'configured' ? (
                    <span className="inline-flex items-center gap-1 rounded-md bg-emerald-500/15 px-1.5 py-0.5 text-[11px] font-medium text-emerald-600 ring-1 ring-emerald-500/30 dark:text-emerald-300">
                      <Check size={11} />
                      Configured
                    </span>
                  ) : (
                    <span className="rounded-md border border-dashed border-border px-1.5 py-0.5 text-[11px] text-muted-foreground">
                      Available
                    </span>
                  )}
                  {i.rate ? (
                    <span className="text-[11px] text-muted-foreground">
                      <span className="font-mono">{i.rate}</span> events/sec ·{' '}
                      <span className="font-mono">{(i.events24h ?? 0).toLocaleString()}</span> in 24h
                    </span>
                  ) : null}
                </div>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          <div className="mt-4 flex flex-wrap items-center gap-2">
            {i.status === 'configured' ? (
              <>
                <Button size="sm">
                  <SettingsIcon size={13} className="mr-1.5" />
                  Manage
                </Button>
                <Button size="sm" variant="outline">
                  <ExternalLink size={13} className="mr-1.5" />
                  View events
                </Button>
              </>
            ) : (
              <Button size="sm">
                <Plus size={13} className="mr-1.5" />
                {i.kind === 'agent' ? 'Get install command' : i.kind === 'collector' ? 'Set up collector' : i.kind === 'cloud' ? 'Connect' : 'Configure custom source'}
              </Button>
            )}
            <Button size="sm" variant="ghost" className="ml-auto">
              <ExternalLink size={13} className="mr-1.5" />
              Documentation
            </Button>
          </div>
        </header>

        <div className="flex-1 overflow-y-auto bg-muted/20 p-6">
          {i.kind === 'agent' && <AgentSetup integration={i} />}
          {i.kind === 'collector' && <CollectorSetup integration={i} />}
          {i.kind === 'cloud' && <CloudSetup integration={i} />}
          {i.kind === 'custom' && <CustomSetup />}
        </div>
      </div>
    </div>
  )
}

/* ─── Setup: Agent ─────────────────────────────────────────────────────── */

function AgentSetup({ integration: i }: { integration: Integration }) {
  const cmd =
    i.id === 'agent-windows'
      ? `# PowerShell (Run as Administrator)
iwr https://install.utmstack.com/agent.ps1 -OutFile install.ps1
./install.ps1 -Token "eyJ...workspace_id=acme..."`
      : i.id === 'agent-linux'
        ? `# Linux — bash
curl -fsSL https://install.utmstack.com/agent.sh | sudo bash -s -- \\
  --token=eyJ...workspace_id=acme...`
        : `# macOS — bash
curl -fsSL https://install.utmstack.com/agent-macos.sh | sudo bash -s -- \\
  --token=eyJ...workspace_id=acme...`
  return (
    <div className="space-y-4">
      <Section title="How it works">
        <p className="text-sm text-foreground/90">
          The {i.name} agent is a small daemon that runs on the host and forwards local events to
          UTMStack over an authenticated TLS connection. It auto-updates and reports its health.
        </p>
      </Section>

      <Section title={`Install on ${i.name}`}>
        <p className="mb-2 text-xs text-muted-foreground">
          Run this command on every {i.name.toLowerCase()} host you want to monitor. The token
          already encodes your current workspace.
        </p>
        <CodeBlock code={cmd} />
      </Section>

      <Section title="What you'll see after install">
        <ul className="list-disc space-y-1 pl-5 text-sm">
          <li>The host appears under Data Sources within ~60 seconds of agent start.</li>
          <li>Telemetry begins flowing into Log Explorer, scoped to your workspace.</li>
          <li>The agent self-reports CPU / memory / disk health back to UTMStack.</li>
        </ul>
      </Section>
    </div>
  )
}

/* ─── Setup: Collector ─────────────────────────────────────────────────── */

function CollectorSetup({ integration: i }: { integration: Integration }) {
  return (
    <div className="space-y-4">
      <Section title="How it works">
        <p className="text-sm text-foreground/90">
          {i.name} doesn't run a UTMStack agent. Instead, you install the UTMStack collector
          somewhere on the network that both your {i.name} device and UTMStack can reach.{' '}
          <span className="text-muted-foreground">{i.name}</span> sends syslog to the collector, the
          collector parses, normalizes, and forwards events into UTMStack.
        </p>
      </Section>

      <Section title="1 · Install the UTMStack collector">
        <p className="mb-2 text-xs text-muted-foreground">
          Run on a Linux host with network access from the {i.name} device.
        </p>
        <CodeBlock
          code={`curl -fsSL https://install.utmstack.com/collector.sh | sudo bash -s -- \\
  --token=eyJ...workspace_id=acme...`}
        />
      </Section>

      <Section title={`2 · Configure ${i.name} to send syslog`}>
        <p className="mb-2 text-xs text-muted-foreground">
          On the {i.name} device, point syslog forwarding to the collector's address on{' '}
          <span className="font-mono text-foreground">{i.defaultPort ?? '514/udp'}</span>.
        </p>
        <CodeBlock
          code={`# Example syslog target on ${i.name}:
host:        utmstack-collector.local
port:        ${i.defaultPort ?? '514/udp'}
format:      RFC 5424
facility:    local0`}
        />
        <a className="mt-2 inline-flex items-center gap-1 text-[11px] text-primary hover:underline" href="#">
          See vendor-specific instructions <ExternalLink size={10} />
        </a>
      </Section>

      <Section title="3 · Verify">
        <p className="text-sm text-foreground/90">
          Within a few minutes you should see {i.name} events under{' '}
          <span className="font-mono">source: {i.name.toLowerCase().replace(/\s/g, '-')}</span> in
          Log Explorer, and a new entry in Data Sources.
        </p>
      </Section>
    </div>
  )
}

/* ─── Setup: Cloud ─────────────────────────────────────────────────────── */

function CloudSetup({ integration: i }: { integration: Integration }) {
  return (
    <div className="space-y-4">
      <Section title="How it works">
        <p className="text-sm text-foreground/90">
          {i.name} runs entirely inside UTMStack — no agent or collector to deploy. Provide
          read-only API credentials and UTMStack will pull events on a schedule, parse them, and
          ingest them into Log Explorer.
        </p>
      </Section>

      <Section title="Connection">
        {i.cloudFields ? (
          <form className="space-y-3" onSubmit={(e) => e.preventDefault()}>
            {i.cloudFields.map((f) => (
              <div key={f.label}>
                <label className="mb-1 block text-[11px] uppercase tracking-wider text-muted-foreground">
                  {f.label}
                </label>
                <Input
                  type={f.secret ? 'password' : 'text'}
                  placeholder={f.placeholder}
                  className="h-9 font-mono text-xs"
                />
              </div>
            ))}
            <div className="flex items-center gap-2 pt-1">
              <Button size="sm" type="submit">Test connection</Button>
              <Button size="sm" type="button" variant="outline">Save</Button>
            </div>
          </form>
        ) : (
          <p className="text-sm text-muted-foreground">No connection fields configured.</p>
        )}
      </Section>

      <Section title="What you'll see after connecting">
        <ul className="list-disc space-y-1 pl-5 text-sm">
          <li>UTMStack performs an initial backfill of recent events.</li>
          <li>From then on, events are pulled every 1–5 minutes (depends on the API).</li>
          <li>The integration appears under Data Sources with current pull status and lag.</li>
        </ul>
      </Section>
    </div>
  )
}

/* ─── Setup: Custom ────────────────────────────────────────────────────── */

function CustomSetup() {
  return (
    <div className="space-y-4">
      <Section title="When to use this">
        <p className="text-sm text-foreground/90">
          When your source isn't in the catalog. Forward arbitrary syslog or JSON to UTMStack and
          define a parser to normalize fields.
        </p>
      </Section>

      <Section title="Endpoints">
        <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
          <dt className="text-muted-foreground">Syslog (UDP)</dt>
          <dd className="font-mono">utmstack-collector.your-net:514</dd>
          <dt className="text-muted-foreground">Syslog (TLS)</dt>
          <dd className="font-mono">utmstack-collector.your-net:6514</dd>
          <dt className="text-muted-foreground">HTTP / JSON</dt>
          <dd className="font-mono">https://ingest.utmstack.com/v1/json</dd>
        </dl>
      </Section>

      <Section title="JSON payload sample">
        <CodeBlock
          code={`POST https://ingest.utmstack.com/v1/json
Authorization: Bearer eyJ...workspace_id=acme...
Content-Type: application/json

{
  "@timestamp": "2026-05-01T16:42:00Z",
  "host": "my-app-01",
  "source": "custom-app",
  "message": "user 'jdoe' logged in",
  "user": "jdoe",
  "ip": "10.4.18.91"
}`}
        />
      </Section>

      <Section title="Define the parser">
        <p className="text-sm text-foreground/90">
          UTMStack will auto-detect JSON keys and propose field mappings. You can override mappings
          in the parser editor and add Grok patterns for free-form text.
        </p>
        <a className="mt-2 inline-flex items-center gap-1 text-[11px] text-primary hover:underline" href="#">
          Open parser editor <ExternalLink size={10} />
        </a>
      </Section>
    </div>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h4 className="mb-2 text-sm font-semibold">{title}</h4>
      {children}
    </div>
  )
}

function CodeBlock({ code }: { code: string }) {
  return (
    <div className="relative">
      <pre className="overflow-x-auto rounded-md bg-muted/40 p-3 font-mono text-[11px] leading-relaxed">
        {code}
      </pre>
      <button
        title="Copy"
        className="absolute right-2 top-2 flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-card hover:text-foreground"
      >
        <Copy size={11} />
      </button>
    </div>
  )
}
