interface MetricProps {
  label: string
  value: string
}

export function Metric({ label, value }: MetricProps) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-mono tabular-nums">{value}</span>
    </div>
  )
}
