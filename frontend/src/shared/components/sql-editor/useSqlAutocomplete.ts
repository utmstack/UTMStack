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

/** The datasets a statement can read from. */
export type SqlAutocompleteTable = string

interface UseSqlAutocompleteResult {
  suggest: (prefix: string, limit?: number) => Suggestion[]
}

export function useSqlAutocomplete(
  fields: SqlAutocompleteField[],
  tables: SqlAutocompleteTable[],
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
    for (const table of tables) {
      if (!table || seen.has(table)) continue
      seen.add(table)
      trie.insert(table, 'index')
    }
  }, [tables, trie])

  const suggest = useCallback(
    (prefix: string, limit = 20) => trie.suggest(prefix, limit),
    [trie],
  )

  return { suggest }
}
