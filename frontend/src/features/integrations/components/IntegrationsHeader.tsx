import { useTranslation } from 'react-i18next'

export function IntegrationsHeader() {
  const { t } = useTranslation()

  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight">{t('integrations.title')}</h1>
      </div>
    </header>
  )
}
