import { useCallback, useEffect, useState } from 'react'
import { adversariesHttpService } from '../services/adversaries-http.service'
import { toAdversaries } from '../lib/adversary'
import type { Adversary } from '../types/adversary.types'

export interface UseAdversariesResult {
  adversaries: Adversary[]
  loading: boolean
  error: boolean
  refresh: () => void
}

/**
 * Loads the alerts grouped by adversary and runs them through the client-side
 * view-model transform (threat level, activity histogram, target/technique
 * aggregates).
 */
export function useAdversaries(): UseAdversariesResult {
  const [adversaries, setAdversaries] = useState<Adversary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const res = await adversariesHttpService.list([])
      setAdversaries(toAdversaries(res ?? []))
    } catch {
      setError(true)
      setAdversaries([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  return { adversaries, loading, error, refresh: () => void load() }
}
