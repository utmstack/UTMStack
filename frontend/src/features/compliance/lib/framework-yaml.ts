import { dump, load } from 'js-yaml'
import type { Framework, FrameworkSection, Requirement } from '../types/compliance.types'

export interface FrameworkFormState {
  key: string
  name: string
  description: string
  source: string
  sections: FrameworkSection[]
}

export type FrameworkParseResult = { ok: true; form: FrameworkFormState } | { ok: false; error: string }

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === 'object' && v !== null && !Array.isArray(v)
}
const str = (v: unknown): string => (v == null ? '' : String(v))
const strArr = (v: unknown): string[] => (Array.isArray(v) ? v.map(String) : [])

function parseRequirements(v: unknown): Requirement[] {
  if (!Array.isArray(v)) return []
  return v.map((r) => {
    const o = isRecord(r) ? r : {}
    return { id: str(o.id), name: str(o.name), satisfiedBy: strArr(o.satisfiedBy) }
  })
}
function parseSections(v: unknown): FrameworkSection[] {
  if (!Array.isArray(v)) return []
  return v.map((s) => {
    const o = isRecord(s) ? s : {}
    return { key: str(o.key), name: str(o.name), requirements: parseRequirements(o.requirements) }
  })
}

export function frameworkToForm(f?: Framework): FrameworkFormState {
  return {
    key: f?.key ?? '',
    name: f?.name ?? '',
    description: f?.description ?? '',
    source: f?.source ?? '',
    sections: (f?.sections ?? []).map((s) => ({
      key: s.key,
      name: s.name,
      requirements: (s.requirements ?? []).map((r) => ({ id: r.id, name: r.name, satisfiedBy: [...(r.satisfiedBy ?? [])] })),
    })),
  }
}

function cleanSections(sections: FrameworkSection[]): FrameworkSection[] {
  return sections
    .filter((s) => s.key.trim() || s.name.trim())
    .map((s) => ({
      key: s.key.trim(),
      name: s.name.trim(),
      requirements: (s.requirements ?? [])
        .filter((r) => r.id.trim() || r.name.trim())
        .map((r) => ({ id: r.id.trim(), name: r.name.trim(), satisfiedBy: (r.satisfiedBy ?? []).filter(Boolean) })),
    }))
}

export function formToFramework(f: FrameworkFormState): Framework {
  return {
    key: f.key.trim(),
    name: f.name.trim(),
    description: f.description.trim(),
    source: f.source.trim(),
    sections: cleanSections(f.sections),
  }
}

export function frameworkFormToYaml(f: FrameworkFormState): string {
  const doc: Record<string, unknown> = { key: f.key, name: f.name }
  if (f.description) doc.description = f.description
  if (f.source) doc.source = f.source
  const sections = cleanSections(f.sections)
  if (sections.length) doc.sections = sections
  return dump(doc, { indent: 2, lineWidth: -1, noRefs: true, quotingType: '"' })
}

export function yamlToFrameworkForm(content: string): FrameworkParseResult {
  let raw: unknown
  try {
    raw = load(content)
  } catch (e) {
    return { ok: false, error: e instanceof Error ? e.message : 'invalid YAML' }
  }
  if (Array.isArray(raw)) raw = raw[0]
  if (!isRecord(raw)) return { ok: false, error: 'root must be a framework mapping' }
  const form: FrameworkFormState = {
    key: str(raw.key),
    name: str(raw.name),
    description: str(raw.description),
    source: str(raw.source),
    sections: parseSections(raw.sections),
  }
  return { ok: true, form }
}
