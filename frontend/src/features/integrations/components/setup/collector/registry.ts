import type { Collector } from './types'

const registry = new Map<string, Collector>()

export function registerCollector(collector: Collector): void {
  registry.set(collector.getName().toLowerCase(), collector)
}

export function getCollector(name: string | undefined | null): Collector | undefined {
  if (!name) return undefined
  const norm = name.toLowerCase()
  const direct = registry.get(norm)
  if (direct) return direct
  for (const c of registry.values()) {
    if (c.matches?.(norm)) return c
  }
  return undefined
}
