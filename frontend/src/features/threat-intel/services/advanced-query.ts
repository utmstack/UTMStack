import type {
  AdvancedSearchRequest,
  AdvancedCondition,
  AdvancedAggregation,
} from '../domain/threat-intel.types'

export function mergeAdvancedRequests(
  ...frags: (AdvancedSearchRequest | undefined)[]
): AdvancedSearchRequest {
  const must: AdvancedCondition[] = []
  const should: AdvancedCondition[] = []
  const mustNot: AdvancedCondition[] = []
  const filter: AdvancedCondition[] = []
  let minShouldMatch: number | undefined
  const aggs: Record<string, AdvancedAggregation> = {}
  for (const f of frags) {
    if (!f) continue
    if (f.query?.must) must.push(...f.query.must)
    if (f.query?.should) should.push(...f.query.should)
    if (f.query?.must_not) mustNot.push(...f.query.must_not)
    if (f.query?.filter) filter.push(...f.query.filter)
    if (f.query?.minimum_should_match !== undefined) minShouldMatch = f.query.minimum_should_match
    if (f.aggs) Object.assign(aggs, f.aggs)
  }
  const query: NonNullable<AdvancedSearchRequest['query']> = {}
  if (must.length) query.must = must
  if (should.length) {
    query.should = should
    query.minimum_should_match = minShouldMatch ?? 1
  }
  if (mustNot.length) query.must_not = mustNot
  if (filter.length) query.filter = filter
  return {
    query: Object.keys(query).length ? query : undefined,
    aggs: Object.keys(aggs).length ? aggs : undefined,
  }
}
