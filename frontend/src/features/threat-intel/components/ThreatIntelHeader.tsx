import { Download, FileText } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'

export interface ThreatIntelHeaderProps {
  matchedCount?: number
  onRefresh?: () => void
}

export function ThreatIntelHeader({ matchedCount }: ThreatIntelHeaderProps) {
  return (
    <header className="flex flex-wrap items-center justify-between gap-3">
      <div className="text-xs text-muted-foreground">
        <span className="font-medium text-foreground">{matchedCount?.toLocaleString() || 0}</span> matched in your env
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm">
          <FileText size={14} className="mr-2" />
          Manage feeds
        </Button>
        <Button variant="default" size="sm">
          <Download size={14} className="mr-2" />
          Export IOCs
        </Button>
      </div>
    </header>
  )
}
