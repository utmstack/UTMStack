import { useTranslation } from 'react-i18next'
import { Download, Info } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'

export interface ThreatIntelHeaderProps {
  matchedCount?: number
  onRefresh?: () => void
  onExport?: () => void
  isExporting?: boolean
  noInstanceIocs?: boolean
}

export function ThreatIntelHeader({ matchedCount, onExport, isExporting, noInstanceIocs }: ThreatIntelHeaderProps) {
  const { t } = useTranslation()
  const canExport = !!onExport && !!matchedCount && !isExporting
  return (
    <header className="flex flex-wrap items-center justify-between gap-3">
      <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
        {!noInstanceIocs && (
        <span>
          <span className="font-medium text-foreground">{matchedCount?.toLocaleString() || 0}</span> {t('threatIntel.header.matchedInEnv')}
        </span>
        )}
        {noInstanceIocs && (
          <span
            className="inline-flex items-center gap-1 rounded-md border border-amber-500/40 bg-amber-500/10 px-2 py-0.5 text-[11px] font-medium text-amber-600 dark:text-amber-400"
            title={t('threatIntel.header.noInstanceIocsHint', {
              defaultValue: 'No indicators observed in your alerts — showing generic feed data.',
            })}
          >
            <Info size={12} />
            {t('threatIntel.header.noInstanceIocs', {
              defaultValue: 'No instance indicators — showing generic data',
            })}
          </span>
        )}
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
