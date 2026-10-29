import { Radio, AlertTriangle } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useTranslation } from 'react-i18next'
import type { Alert } from '@/features/alerts/types/alert.types'
import {useState,useEffect} from 'react'

const isExternal = (ip?: string): ip is string => !!ip && !/^(10\.|192\.168\.|172\.(1[6-9]|2\d|3[0-1])\.|127\.)/.test(ip)

/** The floating LiveOps header: brand + LIVE chip, new-alert badge and the stat pills. */
export function Header({
  replaying,
  newCount,
  openCount,
  list,
  total,
  onNewAlerts,
}: {
  replaying: boolean
  newCount: number
  openCount: number
  list: Alert[]
  total: number
  onNewAlerts: () => void
}) {
  const { t } = useTranslation()

  const [checked,setChecked] =useState(false)

  useEffect(()=>{
    setChecked(false)
  },[newCount])

  return (
    <div className="absolute z-30 flex h-10 items-center gap-3" style={{ top: 12, left: 12, right: 12 }}>
      <div className="flex shrink-0 items-center gap-2.5">
        <Radio size={20} className="text-primary" />
        <h1 className="text-[18px] font-semibold text-foreground">{t('liveOps.header')}</h1>
        <span className={cn('flex items-center gap-1.5 rounded-full px-2 py-0.5 text-[11px] font-bold tracking-wide', !replaying ? 'bg-red-500/15 text-red-600 dark:text-red-400' : 'bg-primary/15 text-primary')}>
          <span className={cn('h-2 w-2 rounded-full', !replaying ? 'lo-live-dot bg-red-500' : 'bg-primary')} />
          {replaying ? t('liveOps.replay') : t('liveOps.live')}
        </span>
        {newCount > 0 && !checked && (
          <button
            type="button"
            onClick={()=>{
              setChecked(true)
              onNewAlerts()
            }}
            className="flex items-center gap-1.5 rounded-full bg-red-500/15 px-2.5 py-0.5 text-[11px] font-bold tracking-wide text-red-600 transition-colors hover:bg-red-500/30 dark:text-red-400"
            title={t('liveOps.newAlerts')}
          >
            <AlertTriangle size={12} />
            {t('liveOps.newAlertsCount', { count: newCount })}
          </button>
        )}
      </div>
      <div className="hidden min-w-0 items-center gap-2 @min-[980px]:flex">
        {[
          [t('liveOps.statHigh'), list.filter((a) => a.severity === 'high').length],
          [t('liveOps.statOpen'), openCount],
          [t('liveOps.statAttackers'), new Set(list.map((a) => a.adversary?.ip).filter(isExternal)).size],
          list.length < total ? [t('liveOps.statTotal'), total] : null,
        ]
          .filter((s): s is [string, number] => s !== null)
          .map(([k, v]) => (
          <div key={k} className="lo-panel shrink-0 rounded-lg px-2.5 py-1">
            <span className="text-[11px] text-muted-foreground">{k} </span>
            <span className="font-mono text-[13px] font-semibold text-foreground">{v}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
