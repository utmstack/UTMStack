import type { TFunction } from 'i18next'
import { ConditionList } from './condition-list'
import type { AfterStep } from './rule-form'

export function AfterEventsView({ steps, t }: { steps: AfterStep[]; t: TFunction }) {
  return (
    <div className="space-y-2">
      {steps.map((s, i) => (
        <div key={i} className="rounded-md border border-border bg-card p-3">
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-[11px] text-muted-foreground">
            <span>{t('alertingRules.editor.indexPattern')}: <span className="font-mono text-foreground">{s.indexPattern || '—'}</span></span>
            <span>{t('alertingRules.editor.within')}: <span className="font-mono text-foreground">{s.within || '—'}</span></span>
            <span>{t('alertingRules.editor.count')} ≥ <span className="font-mono text-foreground">{s.count}</span></span>
          </div>
          {s.with.length > 0 && <ConditionList conds={s.with} t={t} />}
          {s.or.map((g, gi) => (
            <div key={gi} className="mt-1.5 border-t border-border/60 pt-1.5">
              <span className="text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">{t('alertingRules.where.or')}</span>
              <ConditionList conds={g.with} t={t} />
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}
