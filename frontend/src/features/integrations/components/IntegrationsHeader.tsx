import { useTranslation } from 'react-i18next'

interface IntegrationsHeaderProps {
  configured: number
  total: number
}

export function IntegrationsHeader({ configured, total }: IntegrationsHeaderProps) {
  const { t } = useTranslation()

  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{t('integrations.title')}</h1>
        <p className="mt-1 text-xs text-muted-foreground">
          {t('integrations.subtitle', { count: configured, total })}
        </p>
      </div>
    </header>
  )
}
