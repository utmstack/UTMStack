import { List, PanelLeftClose } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useTranslation } from 'react-i18next'
import { templateFor, tsMs } from '../model'
import { useLiveOpsData } from '../liveops-data'
import { SEV_META } from '@/features/alerts/lib/alert-meta'
import { InfiniteScrollSentinel } from '@/shared/components/ui/infinite-scroll'

const sevBar = (severity: string) => (SEV_META[severity as keyof typeof SEV_META] ?? SEV_META.low).bar

function ago(iso: string | undefined, t: (k: string, o?: Record<string, unknown>) => string) {
  if (!iso) return t('liveOps.agoJust')
  const s = Math.max(0, Math.round((Date.now() - tsMs(iso)) / 1000))
  if (s < 60) return t('liveOps.agoJust')
  if (s < 3600) return t('liveOps.agoMin', { count: Math.round(s / 60) })
  return t('liveOps.agoH', { count: Math.round(s / 3600) })
}

export function LeftSidebar({
  open,
  panelBox,
  collapsedStyle,
  width,
  selectedId,
  onShow,
  onOpen,
  onClose,
}: {
  open: boolean
  panelBox: { top: number; bottom: number }
  collapsedStyle: { top: number; left: number }
  width: number
  selectedId: string | null
  onShow: (id: string) => void
  onOpen: () => void
  onClose: () => void
}) {
  const { t } = useTranslation()
  const { alerts, total, loading, loadMore } = useLiveOpsData()

  const list = alerts ?? []


  if (!open) {
    return (
      <button type="button" onClick={onOpen} className="lo-panel absolute z-20 flex h-10 w-10 items-center justify-center rounded-xl text-muted-foreground hover:text-foreground" style={collapsedStyle} title={t('liveOps.listShow')}>
        <List size={17} />
      </button>
    )
  }

  return (
    <div className="lo-panel absolute z-20 flex flex-col overflow-hidden rounded-xl" style={{ ...panelBox, left: collapsedStyle.left, width }}>
      <div className="flex shrink-0 items-center gap-2 border-b border-border px-3 py-2">
        <span className="text-[12px] font-semibold text-foreground">{t('liveOps.listTitle')}</span>
        <span className="text-[11px] text-muted-foreground">{t('liveOps.listShown', { count: list.length })}</span>
        <button type="button" onClick={onClose} className="ml-auto rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground" title={t('liveOps.listHide')}>
          <PanelLeftClose size={14} />
        </button>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto">
        {list.map((a) => {
          const sel = a.id === selectedId
          return (
            <button
              key={a.id}
              type="button"
              onClick={() => onShow(a.id)}
              className={cn('relative flex w-full gap-2 border-b border-border/60 px-3 py-2 text-left transition-colors', sel ? 'bg-primary/15' : 'hover:bg-muted/60')}
              title={sel ? t('liveOps.playAgain') : undefined}
            >
              <span className={cn('mt-0.5 w-1 shrink-0 self-stretch rounded-full', sevBar(String(a.severity)))} />
              <span className="min-w-0 flex-1">
                <span className="block truncate text-[12.5px] font-medium text-foreground">{a.name}</span>
                <span className="block truncate font-mono text-[10.5px] text-muted-foreground">
                  {(a.adversary?.ip ?? a.adversary?.host ?? '?')} → {a.target?.host ?? a.target?.domain ?? a.target?.ip ?? a.dataSource ?? ''}
                </span>
                <span className="block text-[10.5px] text-muted-foreground">
                  {t(`liveOps.kb.${templateFor(a)}`)} · {ago(a['@timestamp'], t)}
                </span>
              </span>
            </button>
          )
        })}
        <InfiniteScrollSentinel onReach={loadMore} hasMore={list.length < total} loading={loading} />
      </div>
    </div>
  )
}
