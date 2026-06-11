import { useTranslation } from 'react-i18next'
import { Pencil, Trash2 } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import type { TenantResponse } from '@/features/integrations/types'

interface CloudTenantListProps {
  tenants: TenantResponse[]
  isDeleting?: boolean
  onEdit: (tenant: TenantResponse) => void
  onDelete: (name: string) => void
}

export function CloudTenantList({ tenants, isDeleting, onEdit, onDelete }: CloudTenantListProps) {
  const { t } = useTranslation()

  if (tenants.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        {t('integrations.setup.cloud.tenants.empty')}
      </p>
    )
  }

  return (
    <ul className="divide-y divide-border rounded-md border border-border">
      {tenants.map((tenant) => (
        <li key={tenant.name} className="flex items-center justify-between gap-2 px-3 py-2">
          <span className="truncate text-sm font-medium">{tenant.name}</span>
          <div className="flex items-center gap-1">
            <Button
              size="sm"
              variant="ghost"
              onClick={() => onEdit(tenant)}
              title={t('integrations.setup.cloud.tenants.edit')}
            >
              <Pencil size={13} />
            </Button>
            <Button
              size="sm"
              variant="ghost"
              disabled={isDeleting}
              onClick={() => onDelete(tenant.name)}
              title={t('integrations.setup.cloud.tenants.delete')}
            >
              <Trash2 size={13} />
            </Button>
          </div>
        </li>
      ))}
    </ul>
  )
}
