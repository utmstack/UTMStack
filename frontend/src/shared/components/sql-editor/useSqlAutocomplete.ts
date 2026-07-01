import { useCallback, useEffect, useMemo } from 'react'

import {
  AutocompleteTrie,
  Suggestion,
  createAutocompleteTrie,
} from './autocomplete-trie.service'
import { SQL_KEYWORDS } from './sql-keywords'

/** Minimal shape needed for field completion. */
export interface SqlAutocompleteField {
  name: string
}

/** Minimal shape needed for index-pattern completion. */
export interface SqlAutocompletePattern {
  pattern: string
}

interface UseSqlAutocompleteResult {
  suggest: (prefix: string, limit?: number) => Suggestion[]
}

export function useSqlAutocomplete(
  fields: SqlAutocompleteField[],
  patterns: SqlAutocompletePattern[],
): UseSqlAutocompleteResult {
  const trie = useMemo<AutocompleteTrie>(() => {
    const t = createAutocompleteTrie()
    for (const kw of SQL_KEYWORDS) t.insert(kw, 'sql')
    return t
  }, [])

  useEffect(() => {
    trie.clearTag('field')
    const seen = new Set<string>()
    for (const f of fields) {
      const base = f.name.endsWith('.keyword') ? f.name.slice(0, -'.keyword'.length) : f.name
      if (!base || seen.has(base)) continue
      seen.add(base)
      trie.insert(base, 'field')
    }
  }, [fields, trie])

  useEffect(() => {
    trie.clearTag('index')
    const seen = new Set<string>()
    for (const p of patterns) {
      if (!p.pattern || seen.has(p.pattern)) continue
      seen.add(p.pattern)
      trie.insert(p.pattern, 'index')
    }
  }, [patterns, trie])

  const suggest = useCallback(
    (prefix: string, limit = 20) => trie.suggest(prefix, limit),
    [trie],
  )

  return { suggest }
}
