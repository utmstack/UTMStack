import type { TFunction } from 'i18next'

export function ConditionList({ conds, t }: { conds: { field: string; operator: string; value: string }[]; t: TFunction }) {
  return (
    <div className="mt-1.5 space-y-1">
      {conds.map((c, i) => (
        <div key={i} className="flex flex-wrap items-center gap-1.5 text-xs">
          <span className="rounded bg-background px-1.5 py-0.5 font-mono text-[11px]">{c.field}</span>
          <span className="text-[11px] text-muted-foreground">{t(`alertingRules.operator.${c.operator}`)}</span>
          {c.value && <span className="rounded bg-primary/10 px-1.5 py-0.5 font-mono text-[11px] text-primary">{c.value}</span>}
        </div>
      ))}
    </div>
  )
}
