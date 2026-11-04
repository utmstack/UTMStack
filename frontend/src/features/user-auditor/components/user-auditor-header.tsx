import type { TFunction } from 'i18next'
import { Download } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'

export function UserAuditorHeader({ total, t }: { total: number; t: TFunction }) {
  return (
    <header className="flex flex-wrap items-center justify-between gap-3">
      <div className="text-xs text-muted-foreground">
        <span className="font-medium text-foreground">{t('userAuditor.accountsChip', { count: total })}</span>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="default" size="sm">
          <Download size={14} className="mr-2" />
          {t('userAuditor.export')}
        </Button>
      </div>
    </header>
  )
}
