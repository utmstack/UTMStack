import { Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import type { AiNote } from '../lib/ai-note'

export function AlertAiAssessment({ note }: { note: AiNote }) {
  const { t } = useTranslation()
  const riskTone =
    note.risk === 'high' ? 'text-red-500' : note.risk === 'medium' ? 'text-amber-500' : 'text-emerald-500'
  return (
    <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-4">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2 text-xs font-medium">
          <Sparkles size={14} className="text-fuchsia-500" /> {t('alerts.ai.title')}
        </div>
        <div className="flex items-center gap-2 text-sm">
          {note.score != null && <span className="font-semibold tabular-nums">{t('alerts.ai.score', { score: note.score })}</span>}
          {note.risk && (
            <span className={cn('text-[10px] font-semibold uppercase tracking-wider', riskTone)}>
              {t('alerts.ai.risk', { risk: t(`alerts.severity.${note.risk}`) })}
            </span>
          )}
        </div>
      </div>
      <dl className="mt-3 space-y-2.5 text-xs">
        {note.sections.map((s, i) => (
          <div key={i}>
            {s.label && <dt className="font-medium text-foreground/80">{s.label}</dt>}
            <dd className="leading-relaxed text-muted-foreground">{s.value}</dd>
          </div>
        ))}
      </dl>
    </div>
  )
}
