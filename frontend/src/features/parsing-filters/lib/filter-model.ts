import { load, dump } from 'js-yaml'

/** Turns a filter's relPath into a human-friendly display name: drop the
 * folder prefix and .yaml/.yml extension, then title-case the rest
 * (windows-events -> "Windows Events", cs_switch -> "Cs Switch"). */
export function displayName(relPath: string): string {
  const base = (relPath.split('/').pop() ?? relPath).replace(/\.ya?ml$/i, '')
  return base
    .split(/[-_]+/)
    .filter(Boolean)
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ')
}

/** Turns a free-form name into a flat (no subpath) yaml filename for a new
 * custom filter: lowercase, punctuation/spaces collapsed to hyphens
 * ("My Custom Filter!" -> "my-custom-filter.yaml"). */
export function sanitizeFileName(name: string): string {
  const slug = name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return slug ? `${slug}.yaml` : ''
}

/**
 * Structured model of a parsing filter, mirroring the Event Processor's pipeline
 * schema (go-sdk plugins.Config) — one dataTypes/order/steps entry per file:
 *
 *   pipeline:
 *     - dataTypes: [<string>, ...]   # a pipeline can be reused across several dataTypes
 *       order: <int32>               # position in the global pipeline order
 *       steps:
 *         - <processor>: { ...config }
 *
 * Each step is exactly one processor. We keep the processor body as a raw
 * `config` object so the round-trip (YAML ↔ model ↔ YAML) is lossless even for
 * processors the visual editor doesn't render friendly fields for — those are
 * edited as raw YAML per step.
 */

export const STEP_KINDS = [
  'grok',
  'kv',
  'json',
  'csv',
  'trim',
  'rename',
  'cast',
  'reformat',
  'delete',
  'drop',
  'add',
  'dynamic',
] as const

export type StepKind = (typeof STEP_KINDS)[number]

/** Processors with a friendly field UI. The rest fall back to raw YAML. */
export const COMMON_KINDS: StepKind[] = ['grok', 'json', 'kv', 'rename', 'drop', 'add']

export interface Step {
  kind: StepKind
  /** The processor body — the map under the processor key. */
  config: Record<string, unknown>
}

export interface FilterModel {
  dataTypes: string[]
  order: number
  steps: Step[]
}

export type ParseResult = { ok: true; model: FilterModel } | { ok: false; error: string }

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v)
}

/** Parse raw YAML into the structured model, or return a human-readable error. */
export function parseFilter(content: string): ParseResult {
  if (content.trim() === '') {
    return { ok: false, error: 'content is empty' }
  }
  let doc: unknown
  try {
    doc = load(content)
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : 'invalid YAML' }
  }
  if (!isRecord(doc)) return { ok: false, error: 'root must be a mapping with a "pipeline" key' }
  const rawPipeline = doc.pipeline
  if (!Array.isArray(rawPipeline) || rawPipeline.length === 0) {
    return { ok: false, error: '"pipeline" must be a non-empty list' }
  }
  if (rawPipeline.length > 1) {
    return { ok: false, error: `this editor only supports one pipeline entry per file (found ${rawPipeline.length})` }
  }
  const rawBlock = rawPipeline[0]
  if (!isRecord(rawBlock)) return { ok: false, error: 'the pipeline entry must be a mapping' }

  const steps: Step[] = []
  const rawSteps = rawBlock.steps
  if (rawSteps != null) {
    if (!Array.isArray(rawSteps)) return { ok: false, error: '"steps" must be a list' }
    for (const rawStep of rawSteps) {
      if (!isRecord(rawStep)) return { ok: false, error: 'each step must be a mapping' }
      const keys = Object.keys(rawStep)
      if (keys.length !== 1) {
        return { ok: false, error: `a step must have exactly one processor (found ${keys.length})` }
      }
      const kind = keys[0] as StepKind
      if (!STEP_KINDS.includes(kind)) {
        return { ok: false, error: `unknown processor "${kind}"` }
      }
      const body = rawStep[kind]
      steps.push({ kind, config: isRecord(body) ? { ...body } : {} })
    }
  }

  const rawDataTypes = rawBlock.dataTypes
  const dataTypes = Array.isArray(rawDataTypes) ? rawDataTypes.filter((d): d is string => typeof d === 'string') : []
  const order = typeof rawBlock.order === 'number' ? rawBlock.order : 0

  return { ok: true, model: { dataTypes, order, steps } }
}

/** Serialize the structured model back to YAML the Event Processor accepts. */
export function serializeFilter(model: FilterModel): string {
  const doc = {
    pipeline: [
      {
        dataTypes: model.dataTypes,
        order: model.order,
        steps: model.steps.map((s) => ({ [s.kind]: pruneEmpty(s.config) })),
      },
    ],
  }
  return dump(doc, { indent: 2, lineWidth: -1, noRefs: true, quotingType: '"' })
}

/** Drop empty-string / empty-array / null keys so the emitted YAML stays clean. */
function pruneEmpty(config: Record<string, unknown>): Record<string, unknown> {
  const out: Record<string, unknown> = {}
  for (const [k, v] of Object.entries(config)) {
    if (v == null) continue
    if (typeof v === 'string' && v.trim() === '') continue
    if (Array.isArray(v) && v.length === 0) continue
    if (isRecord(v) && Object.keys(v).length === 0) continue
    out[k] = v
  }
  return out
}

export function emptyStep(kind: StepKind): Step {
  switch (kind) {
    case 'grok':
      return { kind, config: { source: 'log', patterns: [] } }
    case 'kv':
      return { kind, config: { source: 'log' } }
    case 'json':
      return { kind, config: { source: 'log' } }
    case 'rename':
      return { kind, config: { to: '', from: [] } }
    case 'add':
      return { kind, config: { function: 'string', params: {} } }
    case 'drop':
      return { kind, config: {} }
    default:
      return { kind, config: {} }
  }
}

export function emptyModel(dataTypes: string[], order: number): FilterModel {
  return { dataTypes, order, steps: [] }
}
