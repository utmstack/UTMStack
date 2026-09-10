import { useTranslation } from 'react-i18next'
import { ResizableGridHeader } from '@/shared/components/ui/resizable-grid-header'
import type { useResizableColumns } from '@/shared/hooks/useResizableColumns'

export function FeedsHeader({ tableCols, startDrag }: { tableCols: string; startDrag: ReturnType<typeof useResizableColumns>['startDrag'] }) {
  const { t } = useTranslation()
  return (
    <ResizableGridHeader
      headers={['', t('threatIntel.feeds.table.name'), t('threatIntel.feeds.table.type'), t('threatIntel.feeds.table.accuracy')]}
      tableCols={tableCols}
      startDrag={startDrag}
    />
  )
}
