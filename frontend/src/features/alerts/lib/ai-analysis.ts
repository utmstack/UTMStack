import type { Alert } from '../types/alert.types'
import { isAiNote } from './ai-note'

// The plugin decodes what it is sent into its own alert type, so this is the
// alert's own fields — not the inline events, history or notes, which the
// agent reads through its tools when it wants them. The analyst's notes are
// also kept out of what is sent to the model.
const ANALYSIS_FIELDS = [
  'id',
  '@timestamp',
  'name',
  'category',
  'technique',
  'description',
  'solution',
  'severity',
  'status',
  'statusObservation',
  'dataSource',
  'dataType',
  'impact',
  'impactScore',
  'adversary',
  'target',
  'tags',
  'references',
  'isIncident',
] as const satisfies readonly (keyof Alert)[]

export function aiAnalysisPayload(alert: Alert): Partial<Alert> {
  return Object.fromEntries(ANALYSIS_FIELDS.filter((k) => alert[k] !== undefined).map((k) => [k, alert[k]]))
}

export function hasAiSummary(alert: Alert | null | undefined): boolean {
  return !!alert && (isAiNote(alert.notes) || isAiNote(alert.statusObservation))
}

/**
 * The plugin answers "queued" straight away and writes its assessment onto the
 * alert when the agent finishes, so the summary is waited for by reading the
 * alert again until it carries one.
 */
export async function waitForAiSummary(
  fetchAlert: () => Promise<Alert | null>,
  { intervalMs = 3000, timeoutMs = 180_000, isCancelled = (): boolean => false } = {},
): Promise<Alert | null> {
  const deadline = Date.now() + timeoutMs
  while (Date.now() < deadline && !isCancelled()) {
    await new Promise((resolve) => setTimeout(resolve, intervalMs))
    if (isCancelled()) break
    const fresh = await fetchAlert().catch(() => null)
    if (hasAiSummary(fresh)) return fresh
  }
  return null
}
