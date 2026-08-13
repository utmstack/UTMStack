import type { DeployKind, Tab } from '@/features/integrations/types'

export const KIND_META: Record<DeployKind, { label: string; tone: string; pill: string; dot: string; group: Tab }> = {
  'agents & syslog': { label: 'Agents & Syslog', tone: 'text-sky-500', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/30 dark:text-sky-300', dot: 'bg-sky-500', group: 'agents' },
  device: { label: 'Collector', tone: 'text-violet-500', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300', dot: 'bg-violet-500', group: 'collectors' },
  antivirus: { label: 'Collector', tone: 'text-violet-500', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300', dot: 'bg-violet-500', group: 'collectors' },
  other: { label: 'Collector', tone: 'text-violet-500', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300', dot: 'bg-violet-500', group: 'collectors' },
  custom: { label: 'Custom', tone: 'text-emerald-500', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300', dot: 'bg-emerald-500', group: 'custom' },
  'utmstack modules': { label: 'Collector', tone: 'text-violet-500', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/30 dark:text-violet-300', dot: 'bg-violet-500', group: 'collectors' },
  cloud: { label: 'Cloud', tone: 'text-fuchsia-500', pill: 'bg-fuchsia-500/15 text-fuchsia-600 ring-fuchsia-500/30 dark:text-fuchsia-300', dot: 'bg-fuchsia-500', group: 'cloud' },
}

// Friendly labels for known module-category slugs; unknown slugs are prettified
// from the slug itself. Shared by the category filter and the create/edit drawer.
const CATEGORY_LABELS: Record<string, string> = {
  'operating-system': 'Operating System',
  xdr: 'XDR',
  ids: 'IDS / IPS',
  siem: 'SIEM',
  devops: 'DevOps',
}

export function categoryLabel(c: string): string {
  return CATEGORY_LABELS[c] ?? c.replace(/[-_]/g, ' ').replace(/\b\w/g, (m) => m.toUpperCase())
}

// ingest_type → friendly label + subtle badge tone. The ingest type is how data
// reaches UTMStack (agent / forwarder / plugin / collector).
export const INGEST_META: Record<string, { label: string; pill: string }> = {
  agent: { label: 'Agent', pill: 'bg-sky-500/15 text-sky-600 ring-sky-500/25 dark:text-sky-300' },
  forwarder: { label: 'Collector', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/25 dark:text-violet-300' },
  plugin: { label: 'Plugin', pill: 'bg-amber-500/15 text-amber-600 ring-amber-500/25 dark:text-amber-300' },
  collector: { label: 'Collector', pill: 'bg-violet-500/15 text-violet-600 ring-violet-500/25 dark:text-violet-300' },
}

export function ingestMeta(t?: string): { label: string; pill: string } {
  const key = (t ?? '').toLowerCase()
  return (
    INGEST_META[key] ?? {
      label: key ? key.charAt(0).toUpperCase() + key.slice(1) : '',
      pill: 'bg-muted text-muted-foreground ring-border',
    }
  )
}
