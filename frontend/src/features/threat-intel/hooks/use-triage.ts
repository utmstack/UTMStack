import { useCallback, useMemo, useRef, useState } from 'react'
import { MAX_TRIAGE_ITEMS, parseIndicators, sortRows, triageIndicators, type TriageRow } from '../lib/triage.types'
import { lookupTriageRow } from '../services/triage.service'

export interface UseTriage {
  raw: string
  setRaw: (v: string) => void
  parsed: string[]
  rows: TriageRow[]
  sorted: TriageRow[]
  counts: Partial<Record<string, number>>
  running: boolean
  done: number
  total: number
  run: () => void
  stop: () => void
  reset: () => void
}

/**
 * Bulk indicator triage. Parses the pasted block, resolves it through the
 * threat-intel lookup API with bounded concurrency, and surfaces rows as they
 * land.
 */
export function useTriage(): UseTriage {
  const [raw, setRaw] = useState('')
  const [rows, setRows] = useState<TriageRow[]>([])
  const [running, setRunning] = useState(false)
  const [done, setDone] = useState(0)
  const [total, setTotal] = useState(0)
  const abortRef = useRef<AbortController | null>(null)

  const parsed = useMemo(() => parseIndicators(raw), [raw])
  const sorted = useMemo(() => sortRows(rows), [rows])

  const counts = useMemo(() => {
    const acc: Record<string, number> = {}
    for (const row of rows) acc[row.verdict] = (acc[row.verdict] ?? 0) + 1
    return acc
  }, [rows])

  const run = useCallback(() => {
    if (parsed.length === 0) return
    const controller = new AbortController()
    abortRef.current = controller

    setRunning(true)
    setRows([])
    setDone(0)
    setTotal(Math.min(parsed.length, MAX_TRIAGE_ITEMS))

    void triageIndicators(parsed, {
      signal: controller.signal,
      lookup: (input, index, signal) => lookupTriageRow(input, index, signal),
      onRow: row => {
        setRows(prev => [...prev, row])
        setDone(n => n + 1)
      },
    }).finally(() => {
      setRunning(false)
      abortRef.current = null
    })
  }, [parsed])

  const stop = useCallback(() => {
    abortRef.current?.abort()
    setRunning(false)
  }, [])

  const reset = useCallback(() => {
    stop()
    setRaw('')
    setRows([])
    setDone(0)
    setTotal(0)
  }, [stop])

  return { raw, setRaw, parsed, rows, sorted, counts, running, done, total, run, stop, reset }
}
