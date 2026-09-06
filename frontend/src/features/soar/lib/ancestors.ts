import type { FlowNode } from '../types/soar.types'

/** One ancestor entry, ready for the InsertFieldMenu and edge labels. */
export interface AncestorContribution {
  nodeId: string
  executor: string
  /** Field names known statically from the node's YAML. Empty when the
   *  runtime output shape isn't declarable (`http`, `llm_enrich` — depends on
   *  the remote response or the prompt). */
  fields: string[]
}

/** For a given node id, return every enrichment ancestor reachable by walking
 *  upstream through `onSuccess` / `onError` edges. Ordered nearest-first so
 *  the InsertFieldMenu shows direct parents at the top. Excludes executor
 *  ancestors — they don't contribute output. */
export function enrichmentAncestors(nodes: Record<string, FlowNode>, target: string): AncestorContribution[] {
  const reverse = buildReverseIndex(nodes)
  const out: AncestorContribution[] = []
  const seen = new Set<string>()
  const queue: string[] = [...(reverse.get(target) ?? [])]
  while (queue.length > 0) {
    const id = queue.shift()!
    if (seen.has(id)) continue
    seen.add(id)
    const n = nodes[id]
    if (!n) continue
    if (n.kind === 'enrichment') {
      // llm_enrich output is always normalized to {"result": ...} by the
      // backend, so "result" is statically known; other executors' output
      // shapes stay runtime-dependent (empty fields = user types the path).
      const fields = n.executor === 'llm_enrich' ? ['result'] : []
      out.push({ nodeId: id, executor: n.executor, fields })
    }
    for (const parent of reverse.get(id) ?? []) queue.push(parent)
  }
  return out
}

/** Reverse the DAG so we can walk from a node up to its parents. */
function buildReverseIndex(nodes: Record<string, FlowNode>): Map<string, string[]> {
  const rev = new Map<string, string[]>()
  const push = (child: string, parent: string) => {
    const list = rev.get(child) ?? []
    if (!list.includes(parent)) list.push(parent)
    rev.set(child, list)
  }
  for (const [id, n] of Object.entries(nodes)) {
    for (const c of n.onSuccess ?? []) push(c, id)
    for (const c of n.onError ?? []) push(c, id)
  }
  return rev
}

