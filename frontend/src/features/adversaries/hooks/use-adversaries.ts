import { useCallback, useEffect, useState } from 'react'
import type { FilterType } from '@/features/alerts/types/alert.types'
import { adversariesHttpService } from '../services/adversaries-http.service'
import type { AdversaryResponse } from '../types/adversary.types'

export interface UseAdversariesResult {
  data: AdversaryResponse[]
  loading: boolean
  error: boolean
  refresh: () => void
}

export function useAdversaries(filters: FilterType[]): UseAdversariesResult {
  const [data, setData] = useState<AdversaryResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const res = await adversariesHttpService.list(filters)
      setData(res ?? [])
    } catch {
      setError(true)
      setData([])
    } finally {
      setLoading(false)
    }
  }, [filters])

  useEffect(() => {
    void load()
  }, [load])

  return { data, loading, error, refresh: () => void load() }
}
