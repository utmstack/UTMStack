export function AdversaryActivityChart({ data }: { data: number[] }) {
  const w = 600
  const h = 90
  const max = Math.max(...data, 1)
  const bw = w / data.length
  return (
    <svg viewBox={`0 0 ${w} ${h}`} className="h-20 w-full">
      {data.map((v, i) => {
        const bh = (v / max) * (h - 14)
        return (
          <rect
            key={i}
            x={i * bw + 1}
            y={h - bh - 10}
            width={bw - 2}
            height={bh}
            rx={1}
            className="fill-red-500/60"
          />
        )
      })}
      <g className="fill-muted-foreground" fontSize="9">
        {[0, 6, 12, 18, 23].map((i) => (
          <text key={i} x={i * bw + bw / 2} y={h - 1} textAnchor="middle">{`${String(i).padStart(2, '0')}h`}</text>
        ))}
      </g>
    </svg>
  )
}
