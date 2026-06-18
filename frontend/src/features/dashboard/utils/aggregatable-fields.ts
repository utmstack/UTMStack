import type { IndexProperty } from '@/features/dashboard/types'

/**
 * The OpenSearch SQL endpoint that powers widget queries cannot navigate into
 * `nested`-mapped containers — any field underneath one fails at parse time with
 * `can't resolve Symbol(namespace=FIELD_NAME, name=...)`. In UTMStack mappings the
 * parsed `log.*` (and `target.*`) sub-documents are nested, while normalized fields
 * (top-level keyword/text, plus `object`-mapped containers like `origin.*`) resolve fine.
 *
 * The properties endpoint reports a nested/unmapped container with an empty `type`
 * (an `object` container reports `"object"`), so we can detect the un-queryable
 * namespaces from the schema itself instead of hard-coding `log`/`target`.
 *
 * Additionally, this SQL implementation only registers the *base* mapped field names
 * as resolvable symbols — referencing a `.keyword` multifield directly (e.g.
 * `dataType.keyword`) fails even though the base `dataType` groups fine (SQL uses the
 * multifield under the hood). So `.keyword` duplicates are dropped from the picker.
 */
// Container fields whose leaves the SQL engine cannot navigate: `nested`, unmapped
// (empty type), AND `object`. Object sub-documents (e.g. `origin.*`, `target.*`) are
// reported as `object` but their leaves (`origin.bytesReceived`, `origin.command`)
// fail at parse time with `can't resolve Symbol(...)`, so their children are dropped
// from the picker — the visual builder must never offer a field that errors.
function containerRoots(fields: IndexProperty[]): string[] {
  const roots: string[] = []
  for (const f of fields) {
    const type = (f.type ?? '').trim().toLowerCase()
    if ((type === '' || type === 'nested' || type === 'object') && f.name) {
      roots.push(f.name)
    }
  }
  return roots
}

function isUnderContainer(name: string, roots: string[]): boolean {
  return roots.some((root) => name === root || name.startsWith(`${root}.`))
}

/**
 * Keep only fields that the SQL endpoint can actually GROUP BY / filter on:
 * drops nested-container entries and anything living underneath them.
 */
export function filterAggregatableFields(fields: IndexProperty[]): IndexProperty[] {
  const roots = containerRoots(fields)
  const names = new Set(fields.map((f) => f.name))
  return fields.filter((f) => {
    const type = (f.type ?? '').trim().toLowerCase()
    // Containers themselves are not groupable leaf fields: nested/unmapped ('') and
    // `object` (e.g. `origin`).
    if (type === '' || type === 'nested' || type === 'object') return false
    if (isUnderContainer(f.name, roots)) return false
    // Drop `.keyword` multifields when their base field is present (the base name
    // is the only resolvable symbol); keep standalone fields that happen to end in
    // `.keyword` and have no base counterpart.
    if (f.name.endsWith('.keyword') && names.has(f.name.slice(0, -'.keyword'.length))) {
      return false
    }
    return true
  })
}

/**
 * Fields that can be used in a GROUP BY / COUNT(DISTINCT …) — i.e. a dimension.
 *
 * In OpenSearch SQL a `text` field is analyzed and is NOT aggregatable: grouping
 * by it errors. A text field that has a `<name>.keyword` multifield groups fine
 * (SQL uses the keyword under the hood), so only *pure* text (no keyword sibling)
 * is excluded. Everything else (keyword, numeric, date, boolean, ip) is groupable.
 *
 * Pass the already-{@link filterAggregatableFields}'d list plus the raw mapping
 * (so the keyword siblings — which the filtered list drops — are still visible).
 */
export function groupableFields(
  aggregatable: IndexProperty[],
  rawFields: IndexProperty[]
): IndexProperty[] {
  const rawNames = new Set(rawFields.map((f) => f.name))
  return aggregatable.filter((f) => {
    const type = (f.type ?? '').trim().toLowerCase()
    if (type === 'text') return rawNames.has(`${f.name}.keyword`)
    return true
  })
}
