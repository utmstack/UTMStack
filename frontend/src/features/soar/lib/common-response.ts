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

export function osOf(host?: string): 'windows' | 'macos' | undefined {
  const n = (host ?? '').toLowerCase()
  if (n.includes('windows') || /\b(wks|mbp)\b/.test(n)) return 'windows'
  if (n.includes('mac')) return 'macos'
  return undefined
}

function flowTemplate(name: string): TemplateId | undefined {
  const i = name.indexOf('-')
  return (i > 0 ? name.slice(0, i) : undefined) as TemplateId | undefined
}

function platformMatches(flowPlatform: string, os: 'windows' | 'macos' | undefined): boolean {
  if (os === 'windows') return flowPlatform.startsWith('win')
  if (os === 'macos') return flowPlatform === 'darwin' || flowPlatform === 'macos'
  return true // OS unknown: accept any platform, run on the agent's host
}

/**
 * @param alert  the alert to answer
 * @param flows  system flows (soarFlowsService.list({ systemOwner: true }));
 *               flow.name is the relpath without the .yaml suffix
 * @returns the flow relpath to trigger, or null when no system flow covers
 *          the alert's template+platform
 */
export function resolveCommonResponse(
  alert: Alert,
  flows: { name: string }[],
): string | null {
  const template = templateFor(alert)
  const os = osOf(alert.target?.host ?? alert.adversary?.host)
  const candidates = flows
    .filter((f) => flowTemplate(f.name) === template && platformMatches(f.name, os))
    // OS unknown → linux flows first (the common endpoint), then anything else
    .sort((a, b) => Number(b.name.includes('linux')) - Number(a.name.includes('linux')))
  return candidates.length ? candidates[0].name : null
}
