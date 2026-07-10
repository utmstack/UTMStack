import type { AggregationBucket } from '../../domain/threat-intel.types'

export interface HourlyBucket {
  ts: number
  count: number
}

export function fillHourlyBuckets(buckets: AggregationBucket[]): HourlyBucket[] {
  const now = Date.now()
  const hours: HourlyBucket[] = []

  // Generate 24 hourly slots ending at now, rounded to the hour.
  const roundedNow = Math.floor(now / 3600000) * 3600000
  const bucketMap = new Map(buckets.map((b) => [b.key, b.doc_count]))

  for (let i = 23; i >= 0; i--) {
    const ts = roundedNow - i * 3600000
    hours.push({ ts, count: bucketMap.get(ts) ?? 0 })
  }

  return hours
}
