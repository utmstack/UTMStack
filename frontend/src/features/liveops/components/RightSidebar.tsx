import { useNavigate } from 'react-router-dom'
import { ExternalLink, PanelRightClose, SlidersHorizontal, Sparkles } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useTranslation } from 'react-i18next'
import {Description} from './Description'
import type { Alert } from '@/features/alerts/types/alert.types'
import type { Scene } from '../model'

const ROLE_DOT: Record<string, string> = {
  attacker: 'bg-red-500',
  external: 'bg-red-500',
  perimeter: 'bg-amber-500',
  cloud: 'bg-violet-500',
  internal: 'bg-primary',
}

/** The alert rule's description as plain text, folded to a few lines until opened. */

export function RightSidebar({
  open,
  panelBox,
  collapsedStyle,
  width,
  alert,
  scene,
  selectedParty,
  onAsk,
  onOpen,
  onClose,
}: {
  open: boolean
  panelBox: { top: number; bottom: number }
  collapsedStyle: { top: number; right: number }
  width: number
  alert: Alert | null
  scene: Scene | null
  selectedParty: string | null
  onAsk: (q: string) => void
  onOpen: () => void
  onClose: () => void
}) {
  const { t } = useTranslation()
  const navigate = useNavigate()

  if (!open) {
    return (
      <button type="button" onClick={onOpen} className="lo-panel absolute z-20 flex h-10 w-10 items-center justify-center rounded-xl text-muted-foreground hover:text-foreground" style={collapsedStyle} title={t('liveOps.showDetails')}>
        <SlidersHorizontal size={16} />
      </button>
    )
  }

  return (
    <div className="lo-panel absolute z-20 flex flex-col overflow-hidden rounded-xl" style={{ ...panelBox, right: collapsedStyle.right, width }}>
        <div className="min-h-0 flex-1 overflow-y-auto">
          <div className="border-b border-border px-3 py-2.5">
            <div className="flex items-center">
              <span className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">{t('liveOps.onStage')}</span>
              <button type="button" onClick={onClose} className="ml-auto rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground" title={t('liveOps.hideDetails')}>
                <PanelRightClose size={14} />
              </button>
            </div>
      {alert && scene ? (
        <>
            <div className="mt-0.5 text-[13px] font-semibold leading-snug text-foreground">{alert.name}</div>
            <div className="mt-0.5 text-[11.5px] text-muted-foreground">{scene.title}</div>
            {alert.description && <Description key={alert.id} text={alert.description} />}
            <div className="mt-1.5 flex flex-wrap gap-1.5 text-[11px] text-foreground/80">
              <span className="rounded border border-border px-1.5 py-0.5">{alert.status}</span>
              <span className="rounded border border-border px-1.5 py-0.5">{alert.category}</span>
              {alert.isIncident && <span className="rounded border border-red-500/40 px-1.5 py-0.5 text-red-600 dark:text-red-400">{t('liveOps.inIncident')}</span>}
            </div>
            <button type="button" onClick={() => navigate(
              `/threat-management/alerts`,
              {
                state: {
                  socaiFilters: [{ field: 'id', operator: 'IS', value: alert.id }],
                  socaiTime: '30d',
                }
              }
            )} className="mt-2 flex items-center gap-1 text-[12px] text-primary hover:underline">
              {t('liveOps.openInAlerts')} <ExternalLink size={11} />
            </button>
            </>
      ):null}
          </div>


      {alert && scene ? (

        <>
          <div className="border-b border-border px-3 py-2.5">
            <div className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">{t('liveOps.involved')}</div>
            <div className="mt-1.5 flex flex-col gap-1.5">
              {scene.parties.map((p) => (
                <div key={p.id} className={cn('flex items-center gap-2 rounded-lg border px-2 py-1.5', selectedParty === p.id ? 'border-primary/60 bg-primary/10' : 'border-border bg-muted/30')}>
                  <span className={cn('h-2 w-2 shrink-0 rounded-full', ROLE_DOT[p.role] ?? ROLE_DOT.internal)} />
                  <span className="min-w-0 flex-1">
                    <span className="block truncate text-[12px] font-medium text-foreground">
                      {p.flag ? `${p.flag} ` : ''}
                      {p.name}
                    </span>
                    {p.sub && <span className="block truncate font-mono text-[10.5px] text-muted-foreground">{p.sub}</span>}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <div className="px-3 py-2.5">
            <div className="flex items-center gap-1.5 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
              <Sparkles size={12} className="text-primary" /> {t('liveOps.askSocAi')}
            </div>
            <div className="mt-1.5 flex flex-col gap-1.5">
              {['askExplain', 'askFp', 'askAttacker'].map((k) => (
                <button key={k} type="button" onClick={() => onAsk(t(`liveOps.${k}`))} className="rounded-lg border border-border bg-muted/30 px-2.5 py-1.5 text-left text-[12px] text-foreground hover:border-primary/40 hover:bg-primary/10">
                  {t(`liveOps.${k}`)}
                </button>
              ))}
            </div>
          </div>
          </>

      ) : null}


        </div>
    </div>
  )
}
