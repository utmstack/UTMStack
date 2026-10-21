import type { TriageRow } from '../lib/triage.types'
import { verdictFor } from '../lib/triage.types'
import { threatIntelHttpService } from '@/features/threat-intel/services/threat-intel-http.service'
import { isNotFound } from '@/features/threat-intel/services/ti-errors'

const IP_RE = /^(\d{1,3}\.){3}\d{1,3}$/
const MD5_RE = /^[a-f0-9]{32}$/i
const SHA1_RE = /^[a-f0-9]{40}$/i
const SHA256_RE = /^[a-f0-9]{64}$/i
const DOMAIN_RE = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$/i
const URL_RE = /^https?:\/\/\S+/i
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const CVE_RE = /^CVE-\d{4}-\d{4,}$/i

/** Infers the CM entity type for a pasted indicator; '' if none matches. */
export function detectType(input: string): string {
  if (IP_RE.test(input)) return 'ip'
  if (MD5_RE.test(input)) return 'md5'
  if (SHA1_RE.test(input)) return 'sha1'
  if (SHA256_RE.test(input)) return 'sha256'
  if (CVE_RE.test(input)) return 'cve'
  if (URL_RE.test(input)) return 'url'
  if (EMAIL_RE.test(input)) return 'email-address'
  if (DOMAIN_RE.test(input)) return 'domain'
  return ''
}

/**
 * Resolves one pasted indicator: infer its type, ask the CM proxy for it, and
 * map the answer to a verdict.
 */
export async function lookupTriageRow(
  input: string,
  index: number,
  _signal?: AbortSignal,
): Promise<TriageRow> {
  const key = `${index}:${input}`
  const type = detectType(input)
  if (!type) {
    return { key, input, type: '', verdict: 'invalid', entity: null }
  }

  try {
    // Exact per-type lookup via the CM entity endpoint; 404 means not in corpus.
    const result = await threatIntelHttpService.entityLookup({ type, value: input })
    if (result.kind !== 'ok') {
      // Upstream not configured: the value is still recognisable, just
      // unresolvable right now.
      return { key, input, type, verdict: 'unknown', entity: null, error: 'not configured' }
    }
    const item = result.value
    return { key, input, type: item.type ?? type, verdict: verdictFor(item), entity: item }
  } catch (err) {
    if (isNotFound(err)) {
      return { key, input, type, verdict: 'notFound', entity: null }
    }
    // One failed lookup must not abort the batch — the analyst still wants
    // the other 499 verdicts, with this row visibly marked as unresolved.
    return {
      key,
      input,
      type,
      verdict: 'unknown',
      entity: null,
      error: err instanceof Error ? err.message : String(err),
    }
  }
}
