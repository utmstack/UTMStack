import { Loader2, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'

/** Shown where an alert's AI summary would be, when it has none yet. */
export function AlertAiGenerate({ generating, onGenerate }: { generating: boolean; onGenerate: () => void }) {
  const { t } = useTranslation()
  return (
    <div className="flex items-center justify-between gap-4 rounded-lg border border-dashed border-fuchsia-500/30 bg-fuchsia-500/5 p-4">
      <div className="min-w-0">
        <div className="flex items-center gap-2 text-xs font-medium">
          <Sparkles size={14} className="text-fuchsia-500" /> {t('alerts.ai.title')}
        </div>
        <p className="mt-1 text-xs text-muted-foreground">
          {generating ? t('alerts.ai.generating') : t('alerts.ai.hint')}
        </p>
      </div>
      <Button size="sm" variant="outline" onClick={onGenerate} disabled={generating} className="shrink-0">
        {generating ? <Loader2 size={13} className="mr-1.5 animate-spin" /> : <Sparkles size={13} className="mr-1.5 text-fuchsia-500" />}
        {t('alerts.ai.generate')}
      </Button>
    </div>
  )
}
