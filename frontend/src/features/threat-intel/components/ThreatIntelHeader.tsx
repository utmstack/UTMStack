import { useTranslation } from 'react-i18next'
import { Download } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'

export interface ThreatIntelHeaderProps {
  matchedCount?: number
  onRefresh?: () => void
  onExport?: () => void
  isExporting?: boolean
}

export function ThreatIntelHeader({ matchedCount, onExport, isExporting }: ThreatIntelHeaderProps) {
  const { t } = useTranslation()
  const canExport = !!onExport && !!matchedCount && !isExporting
  return (
    <header className="flex flex-wrap items-center justify-between gap-3">
      <div className="text-xs text-muted-foreground">
        <span className="font-medium text-foreground">{matchedCount?.toLocaleString() || 0}</span> {t('threatIntel.header.matchedInEnv')}
      </div>
      <div className="flex items-center gap-2">
        <Button variant="default" size="sm" onClick={onExport} disabled={!canExport}>
          <Download size={14} className="mr-2" />
          {isExporting ? t('threatIntel.header.exporting') : t('threatIntel.header.export')}
        </Button>
      </div>
    </header>
  )
}
