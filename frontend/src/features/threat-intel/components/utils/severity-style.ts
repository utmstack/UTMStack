import { type LucideIcon, Bug, Fingerprint, Globe2, Hash, LinkIcon, Network, Skull } from 'lucide-react'
import type { EntityType, FeedType, FeedAccuracy } from '../../domain/threat-intel.types'

export type ReputationTone = 'danger' | 'warning' | 'neutral' | 'unknown'

export const REPUTATION_STYLE: Record<ReputationTone, { bar: string; tone: string }> = {
  danger:  { bar: 'bg-red-500',    tone: 'text-red-500' },
  warning: { bar: 'bg-amber-500',  tone: 'text-amber-500' },
  neutral: { bar: 'bg-sky-500',    tone: 'text-sky-500' },
  unknown: { bar: 'bg-muted',      tone: 'text-muted-foreground' },
}

// ponytail: CM reputation_score seen values are 0 (Indefinable) and -3 (Alarming).
// Buckets: <0 danger, 0 unknown-neutral, positive safe. Tighten when CM docs land.
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

const TYPE_FALLBACK = { icon: Fingerprint, label: 'Entity', tone: 'text-muted-foreground' } as const

const TYPE_TABLE: Partial<Record<string, { icon: LucideIcon; label: string; tone: string }>> = {
  ip:       { icon: Network,     label: 'IP',       tone: 'text-sky-500' },
  hostname: { icon: LinkIcon,    label: 'Host',     tone: 'text-amber-500' },
  domain:   { icon: Globe2,      label: 'Domain',   tone: 'text-violet-500' },
  url:      { icon: LinkIcon,    label: 'URL',      tone: 'text-amber-500' },
  hash:     { icon: Hash,        label: 'Hash',     tone: 'text-fuchsia-500' },
  cve:      { icon: Bug,         label: 'CVE',      tone: 'text-red-500' },
  threat:   { icon: Skull,       label: 'Threat',   tone: 'text-red-500' },
}

export function typeMeta(type: EntityType): { icon: LucideIcon; label: string; tone: string } {
  return TYPE_TABLE[type] ?? TYPE_FALLBACK
}

const FEED_TYPE_TABLE: Partial<Record<FeedType, string>> = {
  accumulative: 'text-violet-500',
}
export function feedTypeTone(type: FeedType): string {
  return FEED_TYPE_TABLE[type] ?? 'text-muted-foreground'
}

// ponytail: CM feed accuracy labels seen so far: "level1". Assume level1 > level2 > level3.
const FEED_ACCURACY_TABLE: Partial<Record<FeedAccuracy, { dot: string; label: string }>> = {
  level1: { dot: 'bg-emerald-500', label: 'Level 1' },
  level2: { dot: 'bg-amber-500', label: 'Level 2' },
  level3: { dot: 'bg-red-500', label: 'Level 3' },
}
export function feedAccuracyMeta(a: FeedAccuracy): { dot: string; label: string } {
  return FEED_ACCURACY_TABLE[a] ?? { dot: 'bg-muted-foreground', label: a }
}
