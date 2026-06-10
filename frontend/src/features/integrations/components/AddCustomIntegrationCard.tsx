import { useTranslation } from 'react-i18next'
import { Plus } from 'lucide-react'

interface AddCustomIntegrationCardProps {
  onClick: () => void
}

export function AddCustomIntegrationCard({ onClick }: AddCustomIntegrationCardProps) {
  const { t } = useTranslation()

  return (
    <button
      onClick={onClick}
      className="group relative flex h-full flex-col items-center justify-center overflow-hidden rounded-xl border-2 border-dashed border-border bg-card/50 text-center transition-all hover:border-foreground/40 hover:bg-card hover:shadow-md"
    >
      <div className="flex flex-col items-center gap-3">
        <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
          <Plus size={24} className="text-primary" />
        </div>
        <div>
          <h4 className="text-sm font-semibold text-foreground">{t('integrations.addCustom.title')}</h4>
          <p className="mt-1 text-xs text-muted-foreground">{t('integrations.addCustom.description')}</p>
        </div>
      </div>
    </button>
  )
}
