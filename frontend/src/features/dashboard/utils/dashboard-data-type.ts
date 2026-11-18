import type { Visualization } from '@/features/dashboard/types'
import { normalizeSpecDataType, parseSpec } from './spec'

/**
 * A dashboard's data type is the one all of its widgets share — a mixed
 * dashboard (or one with no typed widgets) has none.
 */
export function dashboardDataType(visualizations: Visualization[]): string | null {

  const types = new Set<string>()
  for (const v of visualizations) {
    const spec = parseSpec(v.spec)
    if (!spec) continue
    const dataType = normalizeSpecDataType(spec).dataType
    if (dataType){
      types.add(dataType)
    }else{
      const catched = spec.filters?.find(f=>f.field=='dataType')?.value as string
      if(catched) types.add(catched)
    }
  }

  return types.size === 1 ? [...types][0] : null
}
