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
