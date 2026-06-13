import { useTranslation } from 'react-i18next'
import { Plus, Puzzle } from 'lucide-react'

interface AddCustomIntegrationCardProps {
  onClick: () => void
}

export function AddCustomIntegrationCard({ onClick }: AddCustomIntegrationCardProps) {
  const { t } = useTranslation()

  return (
    <button
      onClick={onClick}
      className="group flex h-[300px] w-full flex-col items-center justify-between rounded-lg border-2 border-dashed border-border bg-gradient-to-b from-primary/[0.04] to-transparent p-5 text-center transition-all hover:border-primary/50 hover:from-primary/10 hover:shadow-md"
    >
      <div className="flex w-full flex-1 flex-col items-center justify-center gap-4">
        <div className="relative flex h-20 w-20 items-center justify-center rounded-2xl bg-primary/10 ring-1 ring-primary/20 transition-transform group-hover:scale-105">
          <Puzzle size={34} className="text-primary" strokeWidth={1.75} />
          <span className="absolute -bottom-1.5 -right-1.5 flex h-7 w-7 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-md ring-2 ring-card">
            <Plus size={16} />
          </span>
        </div>
        <div className="px-2">
          <h6 className="text-sm font-semibold text-foreground">{t('integrations.addCustom.title')}</h6>
          <p className="mt-1.5 text-[11px] leading-relaxed text-muted-foreground">
            {t('integrations.addCustom.description')}
          </p>
        </div>
      </div>

      <span className="mt-3 inline-flex shrink-0 items-center gap-1.5 rounded-md bg-primary px-4 py-1.5 text-xs font-semibold text-primary-foreground shadow-sm transition-colors group-hover:bg-primary/90">
        <Plus size={13} /> {t('integrations.addCustom.cta')}
      </span>
    </button>
  )
}
