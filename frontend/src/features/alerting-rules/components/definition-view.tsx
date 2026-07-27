import type { TFunction } from 'i18next'
import { parseCelTree } from '../lib/cel-tree'
import { CelNodeView } from './cel-node-view'

export function DefinitionView({ definition, t }: { definition: string; t: TFunction }) {
  const tree = parseCelTree(definition)
  if (!tree) {
    return <pre className="overflow-x-auto rounded-md border border-border bg-card p-3 font-mono text-[11px] leading-relaxed">{definition || '—'}</pre>
  }
  return <CelNodeView node={tree} t={t} depth={0} />
}
