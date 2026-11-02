/**
 * Resolve which system SOAR flow answers an alert.
 *
 * System flows are named `template-platform-party-action[-token]` (see
 * `definitions/soar/`), so the relpath IS the key: no hardcoded path list.
 * `templateFor` comes from liveops (the alert's narrative template); the
 * platform is derived from the target host, same way the flow's `platform:`
 * node is chosen at execution.
 */
import { templateFor, type TemplateId } from '@/features/liveops/model'
import type { Alert } from '@/features/alerts/types/alert.types'

export type Platform = 'windows' | 'linux' | 'macos' | undefined

export function osOf(host?: string): Platform {
  const n = (host ?? '').toLowerCase()
  if (n.includes('windows') || /\b(wks|mbp)\b/.test(n)) return 'windows'
  if (n.includes('mac')) return 'macos'
  if (n) return 'linux' // any named host without a windows/macos marker → linux
  return undefined
}

/** First `-` segment of a relpath: the template token. */
export function flowTemplate(name: string): TemplateId | undefined {
  const i = name.indexOf('-')
  return (i > 0 ? name.slice(0, i) : undefined) as TemplateId | undefined
}

/** Second `-` segment of a relpath: the platform token (windows/linux/macos). */
export function flowPlatform(name: string): string | undefined {
  const parts = name.split('-')
  return parts.length > 1 ? parts[1] : undefined
}

function platformMatches(platform: string | undefined, os: Platform): boolean {
  if (os === 'windows') return platform?.startsWith('win') ?? false
  if (os === 'macos') return platform === 'darwin' || platform === 'macos'
  if (os === 'linux') return platform === 'linux'
  return true // OS unknown: accept any platform, run on the agent's host
}

/**
 * Server-side `name.contains` query so the flow list comes back pre-narrowed:
 * `${template}-${platform}` matches only relpaths of this alert's template on
 * this OS — e.g. a linux alert fetches no windows flows.
 */
export function flowNameFilter(alert: Alert): { relPath: string; systemOwner: true } {
  const template = templateFor(alert)
  const os = osOf(alert.target?.host ?? alert.adversary?.host)
  const platform = os ?? 'linux'
  return { relPath: `${template}-${platform}`, systemOwner: true }
}

/**
 * @param alert  the alert to answer
 * @param flows  flows returned for `flowNameFilter(alert)`; `flow.name` is the
 *               relpath without the .yaml suffix
 * @returns the flow relpath to trigger, or null when no system flow covers
 *          the alert's template+platform
 */
export function resolveCommonResponse(
  alert: Alert,
  flows: { relPath: string }[],
): string | null {
  const template = templateFor(alert)
  const os = osOf(alert.target?.host ?? alert.adversary?.host)
  const candidates = flows
    .filter((f) => flowTemplate(f.relPath) === template && platformMatches(flowPlatform(f.relPath), os))
    .sort((a, b) => Number(b.relPath.includes('linux')) - Number(a.relPath.includes('linux')))
  return candidates.length ? candidates[0].relPath : null
}
