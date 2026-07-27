import type { TFunction } from 'i18next'
import { cn } from '@/shared/lib/utils'
import type { CelNode } from '../lib/cel-tree'

export function CelNodeView({ node, t, depth }: { node: CelNode; t: TFunction; depth: number }) {
  if (node.type === 'cond') {
    return (
      <div className="flex flex-wrap items-center gap-1.5 text-xs">
        {node.negate && <span className="rounded bg-red-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-red-500">not</span>}
        <span className="rounded bg-background px-1.5 py-0.5 font-mono text-[11px]">{node.field}</span>
        <span className="text-[11px] text-muted-foreground">{t(`alertingRules.celOp.${node.fn}`)}</span>
        {node.values.filter(Boolean).map((v, i) => (
          <span key={i} className="rounded bg-primary/10 px-1.5 py-0.5 font-mono text-[11px] text-primary">{v}</span>
        ))}
      </div>
    )
  }
  const connector = node.type === 'and' ? t('alertingRules.where.and') : t('alertingRules.where.or')
  return (
    <div className={cn('space-y-1.5', depth > 0 && 'rounded-md border border-border/60 bg-background/30 p-2')}>
      {node.negate && (
        <span className="inline-block rounded bg-red-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-red-500">not</span>
      )}
      {node.children.map((child, i) => (
        <div key={i}>
          {i > 0 && (
            <div className="mb-1.5 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground/70">{connector}</div>
          )}
          <CelNodeView node={child} t={t} depth={depth + 1} />
        </div>
      ))}
    </div>
  )
}
