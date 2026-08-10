import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2 } from 'lucide-react'
import { Section } from '@/features/integrations/components/ui/Section'
import { useIntegrations } from '@/features/integrations/hooks/useIntegrations'
import type { Integration, ConfigGroupResponse } from '@/features/integrations/types'
import { buildCloudGuide } from './builders/cloudGuideBuilder'
import { CloudGuideSection } from './CloudGuideSection'
import { CloudTenantList } from './CloudTenantList'
import { CloudTenantForm } from './CloudTenantForm'

interface CloudSetupProps {
  integration: Integration
}

export function CloudSetup({ integration }: CloudSetupProps) {
  const { t } = useTranslation()
  const { configGroups: tenantsQuery, deleteConfigGroup } = useIntegrations()

  const moduleName = integration.moduleName ?? ''
  const tenantList = tenantsQuery(moduleName)
  const tenants: ConfigGroupResponse[] = tenantList.data ?? []

  const config = useMemo(
    () => buildCloudGuide(integration.name, { tenants }),
    [integration.name, tenants],
  )

  const [editing, setEditing] = useState<ConfigGroupResponse | null>(null)

  const handleDelete = (name: string) => {
    deleteConfigGroup.mutate({ integration: moduleName, name })
    if (editing?.name === name) {
      setEditing(null)
    }
  }

  return (
    <div className="space-y-4">
      {config.sections.map((section) => (
        <CloudGuideSection key={section.id} section={section} />
      ))}

      <Section id="cloud-tenants-section" title={t('integrations.setup.cloud.tenants.title')}>
        <div className="space-y-3">
          {tenantList.isLoading ? (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              {t('integrations.setup.cloud.tenants.loading')}
            </div>
          ) : (
            <CloudTenantList
              tenants={config.tenants}
              isDeleting={deleteConfigGroup.isPending}
              onEdit={setEditing}
              onDelete={handleDelete}
            />
          )}

          {config.fields.length > 0 && (
            <CloudTenantForm
              moduleName={moduleName}
              fields={config.fields}
              editing={editing}
              onCancel={() => setEditing(null)}
              onSaved={() => setEditing(null)}
            />
          )}
        </div>
      </Section>
    </div>
  )
}
