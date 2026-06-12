import { useCallback, useEffect, useState } from 'react'
import { taggingRulesHttpService as svc } from '../services/tagging-rules-http.service'
import type { TaggingRule, TaggingRuleListParams } from '../types/tagging-rule.types'

export interface UseTaggingRulesListResult {
  rules: TaggingRule[]
  total: number
  loading: boolean
  error: boolean
  refresh: () => void
}

export function useTaggingRulesList(params: TaggingRuleListParams): UseTaggingRulesListResult {
  const [rules, setRules] = useState<TaggingRule[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  // params is rebuilt every render by the page; the JSON key is what actually
  // changes the request so it's the dep we depend on (not the object identity).
  const key = JSON.stringify(params)

  const load = useCallback(async () => {
    setLoading(true)
    setError(false)
    try {
      const { data, total } = await svc.list(params)
      setRules(data ?? [])
      setTotal(total)
    } catch {
      setError(true)
      setRules([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key])

  useEffect(() => {
    void load()
  }, [load])

  return { rules, total, loading, error, refresh: () => void load() }
}
