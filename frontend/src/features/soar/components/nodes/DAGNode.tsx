import { memo } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import { Bell, Sparkles, Terminal, Globe, Brain, Zap, GitBranch } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import type { FlowNode } from '../../types/soar.types'

/** Node data payload carried by ReactFlow — matches the domain FlowNode plus
 *  the YAML id so the canvas can round-trip back to the flow. */
export interface DAGNodeData extends FlowNode {
  nodeId: string
}

const EXECUTOR_ICONS: Record<string, typeof Terminal> = {
  shell: Terminal,
  http: Globe,
  llm_enrich: Brain,
  llm_action: Zap,
  notify: Bell,
  conditional: GitBranch,
}

const KIND_TONES = {
  executor: { border: 'border-emerald-500/40', chip: 'bg-emerald-500/15 text-emerald-500' },
  enrichment: { border: 'border-sky-500/40', chip: 'bg-sky-500/15 text-sky-500' },
} as const

/** A single DAG node laid out top-down: one target handle on top, two source
 *  handles on the bottom — green (left) for success, red (right) for error. */
export const DAGNode = memo(function DAGNode({ data, selected }: NodeProps) {
  const node = data as unknown as DAGNodeData
  const Icon = EXECUTOR_ICONS[node.executor] ?? (node.kind === 'enrichment' ? Sparkles : Terminal)
  const tone = KIND_TONES[node.kind]
  return (
    <div
      className={cn(
        'min-w-[200px] rounded-lg border-2 bg-card shadow-sm transition-shadow',
        tone.border,
        selected ? 'shadow-lg ring-2 ring-primary/50' : 'shadow-sm',
      )}
    >
      <Handle
        type="target"
        position={Position.Top}
        className="!h-3 !w-3 !border-2 !border-muted-foreground/40 !bg-background"
      />
      <div className="flex items-center gap-2 border-b border-border px-3 py-2">
        <span className={cn('flex h-6 w-6 items-center justify-center rounded', tone.chip)}>
          <Icon size={12} />
        </span>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <span className="truncate text-xs font-semibold">{node.nodeId}</span>
            <span className="rounded bg-muted px-1 py-0.5 text-[9px] uppercase text-muted-foreground">{node.kind}</span>
          </div>
          <div className="truncate text-[10px] font-mono text-muted-foreground">{node.executor}</div>
        </div>
      </div>
      {node.command && (
        <div className="px-3 py-2 font-mono text-[10px] text-muted-foreground line-clamp-2">
          {node.command}
        </div>
      )}
      {/* Success — bottom-left, green. */}
      <Handle
        id="success"
        type="source"
        position={Position.Bottom}
        style={{ left: '30%' }}
        className="!h-3 !w-3 !border-2 !border-emerald-500 !bg-emerald-500"
      />
      {/* Error — bottom-right, red. */}
      <Handle
        id="error"
        type="source"
        position={Position.Bottom}
        style={{ left: '70%' }}
        className="!h-3 !w-3 !border-2 !border-red-500 !bg-red-500"
      />
    </div>
  )
})
