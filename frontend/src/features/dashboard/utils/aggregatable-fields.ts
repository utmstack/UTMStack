import { isArrayType, isContainerType, normalizeType } from '@/features/dashboard/utils/field-types'
import type { IndexProperty } from '@/features/dashboard/types'

/**
 * Which of a dataset's fields a widget can be built on.
 *
 * The store describes a record's own columns plus the paths inside its
 * sub-documents: `origin` is reported as a JSON container and
 * `origin.geolocation.country` as a string beside it. The container is not a
 * value — nothing can be grouped by "origin" — while its paths are exactly what
 * a widget asks about, so the containers are dropped and their paths kept.
 */
export function filterAggregatableFields(fields: IndexProperty[]): IndexProperty[] {
  return fields.filter((f) => !!f.name && !isContainerType(f.type))
}

/**
 * The fields a chart can break down by.
 *
 * A list column (`errors`, `references`) holds several values per record, so
 * grouping by it groups by the whole list — a bar per combination, which is
 * never the question being asked. The raw record text is dropped for the same
 * reason: every record is its own group.
 */
export function groupableFields(aggregatable: IndexProperty[]): IndexProperty[] {
  return aggregatable.filter((f) => {
    if (isArrayType(f.type)) return false
    if (f.name === 'raw') return false
    return normalizeType(f.type) !== ''
  })
}
