import { memo } from 'react'
import { Handle, Position } from '@xyflow/react'
import { Zap } from 'lucide-react'

/** The alert-match trigger — a visual-only root shown at the top of the
 *  canvas so users can see where a flow "starts". Its outgoing wires
 *  designate roots in the flow YAML. */
export const TriggerNode = memo(function TriggerNode() {
  return (
    <div className="min-w-[180px] rounded-lg border-2 border-amber-500/60 bg-amber-500/5 shadow-sm">
      <div className="flex items-center gap-2 px-3 py-2.5">
        <span className="flex h-6 w-6 items-center justify-center rounded bg-amber-500/20 text-amber-500">
          <Zap size={12} />
        </span>
        <div>
          <div className="text-xs font-semibold">Alert match</div>
          <div className="text-[10px] text-muted-foreground">flow trigger</div>
        </div>
      </div>
      <Handle
        id="trigger"
        type="source"
        position={Position.Bottom}
        className="!h-3 !w-3 !border-2 !border-amber-500 !bg-amber-500"
      />
    </div>
  )
})
