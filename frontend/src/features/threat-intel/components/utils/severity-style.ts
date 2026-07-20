import { type LucideIcon, Bug, Fingerprint, Globe2, Hash, LinkIcon, Network, Skull } from 'lucide-react'
import type { EntityType, FeedType, FeedAccuracy } from '../../domain/threat-intel.types'

export type ReputationTone = 'danger' | 'warning' | 'neutral' | 'unknown'

export const REPUTATION_STYLE: Record<ReputationTone, { bar: string; tone: string }> = {
  danger:  { bar: 'bg-red-500',    tone: 'text-red-500' },
  warning: { bar: 'bg-amber-500',  tone: 'text-amber-500' },
  neutral: { bar: 'bg-sky-500',    tone: 'text-sky-500' },
  unknown: { bar: 'bg-muted',      tone: 'text-muted-foreground' },
}

export function reputationTone(score: number): ReputationTone {
  if (score <= -2) return 'danger'
  if (score < 0) return 'warning'
  if (score === 0) return 'unknown'
  return 'neutral'
}

// Search items only give a numeric reputation. Detail responses give both the score
// and CM's string. This helper is used when only the score is available.
export function reputationLabel(score: number): string {
  if (score <= -3) return 'Alarming'
  if (score <= -1) return 'Suspicious'
  if (score === 0) return 'Indefinable'
  if (score === 1) return 'Fair'
  return 'Trustworthy'
}

export function reputationLabelKey(score: number): string {
  if (score <= -3) return 'threatIntel.reputation.alarming'
  if (score <= -1) return 'threatIntel.reputation.suspicious'
  if (score === 0) return 'threatIntel.reputation.indefinable'
  if (score === 1) return 'threatIntel.reputation.fair'
  return 'threatIntel.reputation.trustworthy'
}

const TYPE_FALLBACK = { icon: Fingerprint, labelKey: 'threatIntel.entityTypes.entity', tone: 'text-muted-foreground' } as const

const TYPE_TABLE: Partial<Record<string, { icon: LucideIcon; labelKey: string; tone: string }>> = {
  ip:       { icon: Network,     labelKey: 'threatIntel.entityTypes.ip',       tone: 'text-sky-500' },
  hostname: { icon: LinkIcon,    labelKey: 'threatIntel.entityTypes.hostname', tone: 'text-amber-500' },
  domain:   { icon: Globe2,      labelKey: 'threatIntel.entityTypes.domain',   tone: 'text-violet-500' },
  url:      { icon: LinkIcon,    labelKey: 'threatIntel.entityTypes.url',      tone: 'text-amber-500' },
  hash:     { icon: Hash,        labelKey: 'threatIntel.entityTypes.hash',     tone: 'text-fuchsia-500' },
  cve:      { icon: Bug,         labelKey: 'threatIntel.entityTypes.cve',      tone: 'text-red-500' },
  threat:   { icon: Skull,       labelKey: 'threatIntel.entityTypes.threat',   tone: 'text-red-500' },
}

export function typeMeta(type: EntityType): { icon: LucideIcon; labelKey: string; tone: string } {
  return TYPE_TABLE[type] ?? TYPE_FALLBACK
}

const FEED_TYPE_TABLE: Partial<Record<FeedType, string>> = {
  accumulative: 'text-violet-500',
}
export function feedTypeTone(type: FeedType): string {
  return FEED_TYPE_TABLE[type] ?? 'text-muted-foreground'
}

const FEED_ACCURACY_TABLE: Partial<Record<FeedAccuracy, { dot: string; labelKey: string }>> = {
  level1: { dot: 'bg-emerald-500', labelKey: 'threatIntel.feedAccuracy.level1' },
  level2: { dot: 'bg-amber-500', labelKey: 'threatIntel.feedAccuracy.level2' },
  level3: { dot: 'bg-red-500', labelKey: 'threatIntel.feedAccuracy.level3' },
}
export function feedAccuracyMeta(a: FeedAccuracy): { dot: string; labelKey: string } {
  return FEED_ACCURACY_TABLE[a] ?? { dot: 'bg-muted-foreground', labelKey: a }
}
