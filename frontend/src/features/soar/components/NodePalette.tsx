import { Bell, Boxes, Brain, Globe, Sparkles, Terminal, Zap } from 'lucide-react'
import { EXECUTOR_CATALOG, type ExecutorMeta, type NodeKind } from '../types/soar.types'

const ICONS: Record<string, typeof Terminal> = {
  shell: Terminal,
  http: Globe,
  select: Boxes,
  llm_enrich: Brain,
  llm_action: Zap,
  notify: Bell,
}

/** Palette of draggable node types. Each row is one (executor, kind) pair —
 *  since some executors back both kinds (http, select via kind flag), the
 *  palette spells them out so the drag payload is unambiguous. */
export function NodePalette({ readOnly }: { readOnly?: boolean }) {
  const rows: Array<{ meta: ExecutorMeta; kind: NodeKind }> = []
  for (const meta of EXECUTOR_CATALOG) {
    for (const kind of meta.kinds) rows.push({ meta, kind })
  }

  const onDragStart = (event: React.DragEvent, meta: ExecutorMeta, kind: NodeKind) => {
    if (readOnly) return
    event.dataTransfer.setData('application/soar-node', JSON.stringify({ executor: meta.type, kind, paramsPlaceholder: meta.paramsPlaceholder }))
    event.dataTransfer.effectAllowed = 'move'
  }

  return (
    <aside className="flex w-56 shrink-0 flex-col border-r border-border bg-card">
      <div className="border-b border-border px-3 py-2">
        <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Palette</div>
        <div className="mt-0.5 text-[11px] text-muted-foreground">Drag onto the canvas</div>
      </div>
      <div className="flex-1 overflow-y-auto p-2 space-y-1">
        {rows.map(({ meta, kind }) => {
          const Icon = ICONS[meta.type] ?? Sparkles
          return (
            <div
              key={`${meta.type}:${kind}`}
              draggable={!readOnly}
              onDragStart={(e) => onDragStart(e, meta, kind)}
              className={
                'flex cursor-grab items-center gap-2 rounded-md border border-border bg-background px-2 py-1.5 text-[11px] hover:border-primary/50 hover:bg-muted active:cursor-grabbing ' +
                (readOnly ? 'cursor-not-allowed opacity-50' : '')
              }
            >
              <span
                className={
                  'flex h-6 w-6 items-center justify-center rounded ' +
                  (kind === 'enrichment' ? 'bg-sky-500/15 text-sky-500' : 'bg-emerald-500/15 text-emerald-500')
                }
              >
                <Icon size={11} />
              </span>
              <div className="min-w-0">
                <div className="truncate font-medium">{meta.label}</div>
                <div className="truncate text-[9px] uppercase text-muted-foreground">{kind}</div>
              </div>
            </div>
          )
        })}
      </div>
    </aside>
  )
}
