import { useTiUsage } from './use-ti-usage'

export function useTiConfigStatus(): { isConfigured: boolean | undefined; isLoading: boolean } {
  const q = useTiUsage()
  if (q.isLoading) return { isConfigured: undefined, isLoading: true }
  if (q.data?.kind === 'not-configured') return { isConfigured: false, isLoading: false }
  if (q.data?.kind === 'ok') return { isConfigured: true, isLoading: false }
  // request errored (non-503) — treat as configured but broken; the individual hooks handle their own errors
  return { isConfigured: true, isLoading: false }
}
