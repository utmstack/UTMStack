import { useTranslation } from 'react-i18next'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { Section } from '@/features/integrations/components/ui/Section'
import type { Integration } from '@/features/integrations/types'

interface CloudSetupProps {
  integration: Integration
}

export function CloudSetup({ integration: i }: CloudSetupProps) {
  const { t } = useTranslation()

  return (
    <div className="space-y-4">
      <Section title={t('integrations.setup.cloud.howItWorksTitle')}>
        <p className="text-sm text-foreground/90">
          {t('integrations.setup.cloud.howItWorksBody', { name: i.name })}
        </p>
      </Section>

      <Section title={t('integrations.setup.cloud.connectionTitle')}>
        {i.cloudFields ? (
          <form className="space-y-3" onSubmit={(e) => e.preventDefault()}>
            {i.cloudFields.map((f) => (
              <div key={f.label}>
                <label className="mb-1 block text-[11px] uppercase tracking-wider text-muted-foreground">
                  {f.label}
                </label>
                <Input
                  type={f.secret ? 'password' : 'text'}
                  placeholder={f.placeholder}
                  className="h-9 font-mono text-xs"
                />
              </div>
            ))}
            <div className="flex items-center gap-2 pt-1">
              <Button size="sm" type="submit">
                {t('integrations.setup.cloud.testConnection')}
              </Button>
              <Button size="sm" type="button" variant="outline">
                {t('integrations.setup.cloud.save')}
              </Button>
            </div>
          </form>
        ) : (
          <p className="text-sm text-muted-foreground">{t('integrations.setup.cloud.noFields')}</p>
        )}
      </Section>

      <Section title={t('integrations.setup.cloud.afterTitle')}>
        <ul className="list-disc space-y-1 pl-5 text-sm">
          <li>{t('integrations.setup.cloud.after1')}</li>
          <li>{t('integrations.setup.cloud.after2')}</li>
          <li>{t('integrations.setup.cloud.after3')}</li>
        </ul>
      </Section>
    </div>
  )
}
