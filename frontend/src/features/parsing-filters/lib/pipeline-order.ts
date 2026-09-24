// The name the engine matches on: the file's base name without its extension.
export function pipelineIdentity(relPath: string): string {
  const base = relPath.split('/').pop() ?? relPath
  return base.replace(/\.disabled$/, '').replace(/\.[^.]+$/, '')
}

// The order is saved as the tenant's whole sequence, and a view may show only
// some of it (a filter, a search, the pages loaded so far). Moving is therefore
// done on the full sequence: the two pipelines exchange places and everything
// else stays where it is. Null when either is not in it, e.g. deleted since.
export function swapInSequence(sequence: string[], a: string, b: string): string[] | null {
  const i = sequence.indexOf(a)
  const j = sequence.indexOf(b)
  if (i < 0 || j < 0) return null
  const next = [...sequence]
  ;[next[i], next[j]] = [next[j], next[i]]
  return next
}
