import { proxyFetch } from '../console/proxy'

export interface EntityAttributeDefinition {
  type: string
  label: string
  dataType: string
  description?: string
  correlate?: boolean
}

export interface EntityDefinition {
  type: string
  label: string
  attributes?: EntityAttributeDefinition[]
  correlate?: string[]
}

export interface ReportAssociationInput {
  value: string
  type?: string
  mode?: 'association' | 'aggregation'
  tags?: string[]
  visibleBy?: string[]
  associations?: ReportAssociationInput[]
}

export interface ReportIocInput {
  type?: string
  value: string
  reputation?: number
  tags?: string[]
  description?: string
  attributes?: Record<string, string>
  associations?: ReportAssociationInput[]
  correlate?: string[]
  visibleBy?: string[]
}

export interface ReportedAssociation {
  mode: 'association' | 'aggregation'
  type: string
  value: string
  associations?: ReportedAssociation[]
}

export interface ReportIocResult {
  type: string
  value: string
  reputation: number | undefined
  tags: string[]
  associations: ReportedAssociation[]
  skippedAssociations: string[]
  message: string
}

export class TypeInferenceError extends Error {
  constructor(value: string) {
    super(value)
    this.name = 'TypeInferenceError'
  }
}

const IPV4 = /^(\d{1,3}\.){3}\d{1,3}$/
const IPV6 = /^([0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}$/
const HEX = /^[0-9a-f]+$/i

export function detectIndicatorTypes(value: string): string[] {
  const v = value.trim().toLowerCase()
  if (!v) return []
  if (IPV4.test(v)) return ['ip']
  if (IPV6.test(v)) return ['ip']
  if (HEX.test(v)) {
    switch (v.length) {
      case 32: return ['hash']
      case 40: return ['hash']
      case 56: return ['hash']
      case 64: return ['hash']
      case 96: return ['hash']
      case 128: return ['hash']
    }
  }
  if (/^cve-\d{4}-\d{4,}$/i.test(v)) return ['cve']
  if (/^https?:\/\//i.test(v)) return ['url']
  const labels = v.split('.')
  if (labels.length === 1) return ['hostname', 'email-address']
  if (labels.length === 2) return ['domain', 'hostname']
  return ['domain', 'url']
}

const NUMERIC_ATTRIBUTES: ReadonlySet<string> = new Set([
  'aba-rtn',
  'asn',
  'bank-account-nr',
  'bin',
  'breach-count',
  'cc-number',
  'accuracy-radius',
  'latitude',
  'longitude',
  'size-in-bytes',
])

const SCANNED_TYPES: ReadonlySet<string> = new Set(['ip', 'domain', 'hostname'])

export function isScannedType(type: string | undefined): boolean {
  return type ? SCANNED_TYPES.has(type) : false
}

export type AttributeControl = 'number' | 'email' | 'url' | 'date' | 'text'

export function controlFor(dataType: string): AttributeControl {
  switch (dataType) {
    case 'Integer':
    case 'Float':
      return 'number'
    case 'Email':
      return 'email'
    case 'URL':
      return 'url'
    case 'Date':
      return 'date'
    default:
      return 'text'
  }
}

export function coerceAttribute(name: string, raw: string): string | number | undefined {
  const value = raw.trim()
  if (!value) return undefined
  if (!NUMERIC_ATTRIBUTES.has(name)) return value
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : undefined
}

export type AttributeProblem = { kind: 'hex'; length?: number } | { kind: 'number' }

const HASH_LENGTHS: Record<string, number> = {
  'SHA-1': 40,
  'SHA-224': 56,
  'SHA-256': 64,
  'SHA-384': 96,
  'SHA-512': 128,
  'SHA3-224': 56,
  'SHA3-256': 64,
  'SHA3-384': 96,
  'SHA3-512': 128,
  'SHA512-224': 56,
  'SHA512-256': 64,
}

export function attributeError(definition: EntityAttributeDefinition, raw: string): AttributeProblem | null {
  const value = raw.trim().toLowerCase()
  if (!value) return null

  if (definition.dataType === 'MD5') {
    const sized = value.length === 32 || value.length === 16
    return sized && HEX.test(value) ? null : { kind: 'hex', length: 32 }
  }

  if (definition.dataType === 'Hexadecimal') {
    return HEX.test(value) ? null : { kind: 'hex' }
  }

  const length = HASH_LENGTHS[definition.dataType]
  if (length !== undefined) {
    return value.length === length && HEX.test(value) ? null : { kind: 'hex', length }
  }

  if (controlFor(definition.dataType) === 'number') {
    return Number.isFinite(Number(value)) ? null : { kind: 'number' }
  }

  return null
}

export function correlatableAttributes(
  definition: EntityDefinition | undefined,
  attributes: Record<string, string> | undefined,
): string[] {
  if (!definition?.correlate?.length || !attributes) return []
  return definition.correlate.filter(
    (name) => name !== definition.type && Boolean(attributes[name]?.trim()),
  )
}

export function nextCorrelate(
  available: readonly string[],
  value: readonly string[] | undefined,
  dismissed: readonly string[],
): string[] | null {
  const next = available.filter((name) => !dismissed.includes(name))
  if (value === undefined) return next.length ? next : null
  const unchanged = next.length === value.length && next.every((name) => value.includes(name))
  return unchanged ? null : next
}

export function nextAttributes(
  definition: EntityDefinition | undefined,
  attributes: Record<string, string> | undefined,
): Record<string, string> | null {
  const names = attributes ? Object.keys(attributes) : []
  if (names.length === 0) return null
  if (!definition) return null

  const declared = new Set((definition.attributes ?? []).map((a) => a.type))
  const kept = names.filter((name) => declared.has(name))
  if (kept.length === names.length) return null
  return Object.fromEntries(kept.map((name) => [name, attributes![name]]))
}

export function sharesVisibility(
  parent: string[] | undefined,
  child: string[] | undefined,
): boolean {
  if (parent === undefined || child === undefined) return true
  if (parent.length === 0 || child.length === 0) return false
  return parent.some((group) => child.includes(group))
}

export function resolveDefinition(
  definitions: readonly EntityDefinition[] | undefined,
  input: { type?: string; value: string },
): EntityDefinition | undefined {
  const typed = input.type?.trim()
  const value = input.value.trim()
  const type = typed || (value ? detectIndicatorTypes(value)[0] : undefined)
  if (!type) return undefined
  return definitions?.find((definition) => definition.type === type)
}

export function attributeProblems(
  definition: EntityDefinition | undefined,
  value: Record<string, string> | undefined,
): Record<string, AttributeProblem> {
  const problems: Record<string, AttributeProblem> = {}
  for (const attribute of definition?.attributes ?? []) {
    const problem = attributeError(attribute, value?.[attribute.type] ?? '')
    if (problem) problems[attribute.type] = problem
  }
  return problems
}

export function associationConflicts(
  parentVisibleBy: string[] | undefined,
  associations: readonly { visibleBy?: string[] }[],
): number[] {
  return associations
    .map((association, index) =>
      sharesVisibility(parentVisibleBy, association.visibleBy) ? -1 : index,
    )
    .filter((index) => index !== -1)
}

export function isKnownEntityType(
  type: string | undefined,
  definitions: readonly EntityDefinition[] | undefined,
): boolean {
  const candidate = type?.trim()
  if (!candidate) return true
  if (!definitions?.length) return true
  return definitions.some((definition) => definition.type === candidate)
}

interface IngestAssociation {
  mode: string
  type: string
  attributes: Record<string, string | number>
  tags?: string[]
  visibleBy?: string[]
  associations?: IngestAssociation[]
}

interface IngestEntity {
  type: string
  attributes: Record<string, string | number>
  reputation?: number
  tags?: string[]
  associations?: IngestAssociation[]
  correlate?: string[]
  visibleBy?: string[]
}

function coerceAttributes(raw: Record<string, string> | undefined): Record<string, string | number> {
  const out: Record<string, string | number> = {}
  for (const [name, value] of Object.entries(raw ?? {})) {
    const coerced = coerceAttribute(name, value)
    if (coerced !== undefined) out[name] = coerced
  }
  return out
}

function buildAssociations(
  inputs: ReportAssociationInput[] | undefined,
  depth = 0,
): { payload: IngestAssociation[]; reported: ReportedAssociation[]; skipped: string[] } {
  if (!inputs?.length || depth > 4) return { payload: [], reported: [], skipped: [] }

  const payload: IngestAssociation[] = []
  const reported: ReportedAssociation[] = []
  const skipped: string[] = []

  for (const input of inputs) {
    const value = input.value?.trim()
    if (!value) continue

    const type = input.type?.trim() || detectIndicatorTypes(value)[0]
    if (!type) {
      skipped.push(value)
      skipped.push(...buildAssociations(input.associations, depth + 1).skipped)
      continue
    }

    const mode = input.mode === 'aggregation' ? 'aggregation' : 'association'
    const nested = buildAssociations(input.associations, depth + 1)
    skipped.push(...nested.skipped)

    const associationTags = input.tags?.filter(Boolean) ?? []

    payload.push({
      mode,
      type,
      attributes: { ...coerceAttributes(undefined), [type]: value },
      ...(associationTags.length ? { tags: associationTags } : {}),
      ...(input.visibleBy ? { visibleBy: input.visibleBy } : {}),
      ...(nested.payload.length ? { associations: nested.payload } : {}),
    })
    reported.push({
      mode,
      type,
      value,
      ...(nested.reported.length ? { associations: nested.reported } : {}),
    })
  }

  return { payload, reported, skipped }
}

function clampIngestReputation(score: number): number {
  return Math.max(-3, Math.min(3, score))
}

export function buildEntityPayload(input: ReportIocInput): {
  entity: IngestEntity
  associations: ReportedAssociation[]
  skippedAssociations: string[]
} {
  const value = input.value.trim()
  const type = input.type?.trim() || detectIndicatorTypes(value)[0]
  if (!type) throw new TypeInferenceError(value)

  const attributes: Record<string, string | number> = coerceAttributes(input.attributes)
  if (input.description) attributes.text = input.description
  attributes[type] = value

  const reputation =
    input.reputation === undefined ? undefined : clampIngestReputation(input.reputation)
  const tags = input.tags?.filter(Boolean) ?? []
  const correlate = input.correlate?.filter(Boolean) ?? []
  const visibleBy = input.visibleBy?.filter(Boolean) ?? []
  const { payload, reported, skipped } = buildAssociations(input.associations)

  return {
    entity: {
      type,
      attributes,
      ...(reputation !== undefined ? { reputation } : {}),
      ...(tags.length ? { tags } : {}),
      ...(payload.length ? { associations: payload } : {}),
      ...(correlate.length ? { correlate } : {}),
      ...(visibleBy.length ? { visibleBy } : {}),
    },
    associations: reported,
    skippedAssociations: skipped,
  }
}

export async function reportIoc(input: ReportIocInput): Promise<ReportIocResult> {
  const { entity, associations, skippedAssociations } = buildEntityPayload(input)
  const value = input.value.trim()
  const tags = input.tags?.filter(Boolean) ?? []

  const ack = await proxyFetch<{ message?: string } | null>('/report/entity', {
    method: 'POST',
    body: JSON.stringify(entity),
  })

  return {
    type: entity.type,
    value,
    reputation: entity.reputation,
    tags,
    associations,
    skippedAssociations,
    message: ack?.message ?? 'accepted',
  }
}

export async function listEntityDefinitions(): Promise<EntityDefinition[]> {
  try {
    const list = await proxyFetch<EntityDefinition[]>('/report/definitions')
    return Array.isArray(list) ? list : []
  } catch {
    return []
  }
}
