import { cn } from '@/shared/lib/utils'

export function Sparkline({ data, accent }: { data: number[]; accent: string }) {
  const w = 200
  const h = 36
  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1
  const xs = data.map((_, i) => (i * w) / (data.length - 1))
  const ys = data.map((v) => h - ((v - min) / range) * (h - 4) - 2)
  const path = data.map((_, i) => `${i === 0 ? 'M' : 'L'} ${xs[i]} ${ys[i]}`).join(' ')
  const area = `${path} L ${xs[xs.length - 1]} ${h} L ${xs[0]} ${h} Z`
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-9 w-full" preserveAspectRatio="none">
      <path d={area} className={cn('opacity-15', accent.replace('text-', 'fill-'))} />
      <path d={path} fill="none" strokeWidth="1.75" strokeLinejoin="round" className={cn('stroke-current', accent)} />
    </svg>
  )
}
