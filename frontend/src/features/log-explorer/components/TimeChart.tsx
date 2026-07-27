import type { ChartView } from '../types/log-explorer.types'
import { chartTimeLabel } from './log-explorer.constants'

export function TimeChart({ data }: { data: ChartView }) {
  const values = data.values
  const max = Math.max(1, ...values)
  const w = 1200
  const h = 300
  const n = values.length || 1
  const slot = w / n
  const bw = Math.max(1, slot - 3)
  return (
    <div>
      <svg viewBox={`0 0 ${w} ${h}`} className="h-[300px] w-full" preserveAspectRatio="none">
        {values.map((v, i) => {
          if (v <= 0) return null
          const bh = Math.max(2, (v / max) * h)
          const x = i * slot + (slot - bw) / 2
          return (
            <rect key={i} x={x} y={h - bh} width={bw} height={bh} rx={2} className="fill-primary/45">
              <title>{`${data.categories[i]}: ${v.toLocaleString()}`}</title>
            </rect>
          )
        })}
      </svg>
      {data.categories.length > 1 && (
        <div className="mt-2 flex justify-between text-[10px] text-muted-foreground">
          <span className="font-mono">{chartTimeLabel(data.categories[0])}</span>
          <span className="font-mono">{chartTimeLabel(data.categories[data.categories.length - 1])}</span>
        </div>
      )}
    </div>
  )
}
