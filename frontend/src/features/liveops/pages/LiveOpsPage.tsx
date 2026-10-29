import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { AlertTriangle, Loader2, Radio } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useTranslation } from 'react-i18next'
import { useSocAi, useSocAiFocus } from '@/features/soc-ai'
import { alertFocus } from '@/features/alerts/lib/alert-focus'
import type { Alert } from '@/features/alerts/types/alert.types'
import { alertsHttpService } from '@/features/alerts/services/alerts-http.service'
import { buildScene, type Party, tsMs } from '../model'
import { Stage } from '../components/Stage'
import { Timeline } from '../components/Timeline'
import { LeftSidebar } from '../components/LeftSidebar'
import { RightSidebar } from '../components/RightSidebar'
import { WINDOW_MS, type WindowKey } from '../components/Timeline'
import '../liveops.css'

// Where the floating pieces sit (px).
const EDGE = 12
const TOP = 62
const LEFT_W = 250
const RIGHT_W = 300
/** The scene never gets narrower than this, so its labels stay readable; a side panel folds instead. */
const MIN_SCENE = 620
/** How often the store is re-read while the page is open. */
const POLL_MS = 10_000

/** Width and height of an element, kept up to date. */
function useSize<T extends HTMLElement>() {
  const ref = useRef<T>(null)
  const [size, setSize] = useState({ width: 0, height: 0 })
  useEffect(() => {
    const el = ref.current
    if (!el) return
    const ro = new ResizeObserver(([e]) => setSize({ width: e.contentRect.width, height: e.contentRect.height }))
    ro.observe(el)
    return () => ro.disconnect()
  }, [])
  return [ref, size] as const
}

const isExternal = (ip?: string): ip is string => !!ip && !/^(10\.|192\.168\.|172\.(1[6-9]|2\d|3[01])\.|127\.)/.test(ip)
/** The alert rule's description as plain text, folded to a few lines until opened. */

export function LiveOpsPage() {
  const { t } = useTranslation()
  const soc = useSocAi()
  const [alerts, setAlerts] = useState<Alert[] | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [win, setWin] = useState<WindowKey>('6h')
  const [party, setParty] = useState<Party | null>(null)
  const [tick, setTick] = useState(0)
  const [error, setError] = useState('')
  // True while the list is re-fetched after a window change; the interval polls don't set it.
  const [loading, setLoading] = useState(false)
  // Each restart of the same alert gets a new run number, so its story plays again from the top.
  const [run, setRun] = useState(0)
  const [replaying, setReplaying] = useState(false)
  // The side panels: folded or open by the visitor, or decided by the room available (null).
  const [leftPref, setLeftPref] = useState<boolean | null>(null)
  const [rightPref, setRightPref] = useState<boolean | null>(null)
  const [lastOpened, setLastOpened] = useState<'left' | 'right'>('right')
  // The room the visitor had when they last folded or opened a panel.
  const prefWidth = useRef(0)
  const [wallRef, wall] = useSize<HTMLDivElement>()
  const [timelineRef, timelineSize] = useSize<HTMLDivElement>()

  // The single source: parent alerts inside the timeline window, re-read every 30 s and
  // whenever the window changes. Counters, the alert list and the timeline all use it.
  const refresh = useCallback(
    (w: WindowKey) =>
      alertsHttpService
        .list({
          page: 0,
          size: 500,
          filters: [
            { field: 'parentId', operator: 'IS', value:''  },
            { field: '@timestamp', operator: 'IS_BETWEEN', value: [new Date(Date.now() - WINDOW_MS[w]).toISOString(),new Date(Date.now()).toISOString()] },
          ],
        })
        .then((r) => {
          setAlerts(r.data)
          setLoading(false)
          setError('')
        })
        .catch((e: unknown) => {
          setLoading(false)
          setError(e instanceof Error ? e.message : String(e))
        }),
    [],
  )

  useEffect(() => {
    setLoading(true)
    void refresh(win)
    const id = setInterval(() => void refresh(win), POLL_MS)
    return () => clearInterval(id)
  }, [win, refresh])

  useEffect(() => {
    const id = setInterval(() => setTick((x) => x + 1), 5_000)
    return () => clearInterval(id)
  }, [])

  const list = useMemo(
    () => (alerts ?? []).map((a) => ({ ...a })).sort((a, b) => tsMs(b['@timestamp']) - tsMs(a['@timestamp'])),
    [alerts],
  )
  const windowAlerts = useMemo(() => list, [list, tick])
  const alert = list.find((a) => a.id === selectedId) ?? list[list.length - 1]
  const scene = useMemo(() => (alert ? buildScene(alert, t) : null), [alert, t])

  useSocAiFocus(useMemo(() => (alert ? alertFocus(alert) : null), [alert]))

  // A new alert on stage: start fresh.
  useEffect(() => {
    setParty(null)
  }, [alert?.id])

  function setLeft(open: boolean) {
    setLeftPref(open)
    prefWidth.current = wall.width
    if (open) setLastOpened('left')
  }

  function setRight(open: boolean) {
    setRightPref(open)
    prefWidth.current = wall.width
    if (open) setLastOpened('right')
  }

  const openRight = () => setRight(true)

  function show(id: string) {
    setReplaying(false)
    if (id === alert?.id) setRun((r) => r + 1)
    else setSelectedId(id)
  }

  function ask(q: string) {
    soc.submit(q, { openPanel: true })
  }

  function onSceneEnded() {
    if (!replaying || !alert) return
    const i = windowAlerts.findIndex((a) => a.id === alert.id)
    const next = windowAlerts[i - 1]
    if (next) setSelectedId(next.id)
    else setReplaying(false)
  }

  useEffect(() => {
    if (Math.abs(wall.width - prefWidth.current) > 200) {
      setLeftPref(null)
      setRightPref(null)
    }
  }, [wall.width])
  const rightW = RIGHT_W
  const sceneWidth = (l: boolean, r: boolean) => wall.width - (l ? 2 * EDGE + LEFT_W : 64) - (r ? 2 * EDGE + rightW : 64)

  const fits = (l: boolean, r: boolean) => wall.width === 0 || sceneWidth(l, r) >= MIN_SCENE
  let rightOpen = rightPref ?? fits(false, true)
  let leftOpen = leftPref ?? fits(true, rightOpen)
  // Both panels do not fit next to a usable scene: the one opened last stays.
  if (wall.width > 0 && leftOpen && rightOpen && sceneWidth(true, true) < MIN_SCENE) {
    if (lastOpened === 'left') rightOpen = false
    else leftOpen = false
  }
  // A lone panel that would still leave too little room floats over the scene instead of pushing it.
  const roomFor = wall.width === 0 || sceneWidth(leftOpen, rightOpen) >= MIN_SCENE
  const bottom = EDGE + (timelineSize.height || 104) + 10
  const area = {
    top: TOP,
    bottom,
    left: leftOpen && roomFor ? EDGE + LEFT_W + EDGE : 64 + EDGE,
    right: rightOpen && roomFor ? EDGE + rightW + EDGE : 64 + EDGE,
  }
  const panelBox = { top: TOP, bottom }

  if (!alerts) {
    return (
      <div ref={wallRef} className="lo-wall @container relative h-full min-h-[560px] overflow-hidden text-foreground">
        <div className="absolute inset-0 flex items-center justify-center">
          {error ? <span className="flex items-center gap-2 text-red-600 dark:text-red-400"><AlertTriangle size={14} /> {t('liveOps.error', { msg: error })}</span> : (
            <span className="flex flex-col items-center gap-2.5">
              <Loader2 size={36} className="animate-spin text-primary" />
              <span className="text-[12px] font-medium text-muted-foreground">{t('liveOps.loading')}</span>
            </span>
          )}
        </div>
      </div>
    )
  }

  return (
    <div ref={wallRef} className="lo-wall @container relative h-full min-h-[560px] overflow-hidden text-foreground">
      {alert && scene ? (
        <Stage key={`${alert.id}:${run}`} alert={alert} scene={scene} templateLabel={t(`liveOps.kb.${scene.template}`)} area={area} selectedParty={party?.id} onSelectParty={setParty} onEnded={onSceneEnded} />
      ) : (
        <div className="absolute inset-0 flex items-center justify-center text-sm text-muted-foreground">{t('liveOps.empty')}</div>
      )}

      {/* Refetch in flight (window change): show it on the canvas even when it is empty. */}
      {loading && (
        <div className="absolute z-10 flex flex-col items-center justify-center gap-2.5" style={area}>
          <Loader2 size={36} className="animate-spin text-primary" />
          <span className="text-[12px] font-medium text-muted-foreground">{t('liveOps.loading')}</span>
        </div>
      )}

      {/* Header */}
      <div className="absolute z-30 flex h-10 items-center gap-3" style={{ top: EDGE, left: EDGE, right: EDGE }}>
        <div className="flex shrink-0 items-center gap-2.5">
          <Radio size={20} className="text-primary" />
          <h1 className="text-[18px] font-semibold text-foreground">{t('liveOps.header')}</h1>
          <span className={cn('flex items-center gap-1.5 rounded-full px-2 py-0.5 text-[11px] font-bold tracking-wide', !replaying ? 'bg-red-500/15 text-red-600 dark:text-red-400' : 'bg-primary/15 text-primary')}>
            <span className={cn('h-2 w-2 rounded-full', !replaying ? 'lo-live-dot bg-red-500' : 'bg-primary')} />
            {replaying ? t('liveOps.replay') : t('liveOps.live')}
          </span>
        </div>
        <div className="hidden min-w-0 items-center gap-2 @min-[980px]:flex">
          {[
            [t('liveOps.statLastHour'), list.filter((a) => a['@timestamp'] && Date.now() - tsMs(a['@timestamp']) < 3_600_000).length],
            [t('liveOps.statHigh'), list.filter((a) => a.severity === 'high').length],
            [t('liveOps.statOpen'), list.filter((a) => String(a.status).toLowerCase() === 'open').length],
            [t('liveOps.statAttackers'), new Set(list.map((a) => a.adversary?.ip).filter(isExternal)).size],
          ].map(([k, v]) => (
            <div key={k} className="lo-panel shrink-0 rounded-lg px-2.5 py-1">
              <span className="text-[11px] text-muted-foreground">{k} </span>
              <span className="font-mono text-[13px] font-semibold text-foreground">{v}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Incoming alerts (floating, folds away) */}
      <LeftSidebar
        open={leftOpen}
        panelBox={panelBox}
        collapsedStyle={{ top: TOP, left: EDGE }}
        width={LEFT_W}
        alerts={windowAlerts}
        selectedId={alert?.id ?? null}
        onShow={show}
        onOpen={() => setLeft(true)}
        onClose={() => setLeft(false)}
      />

      {/* Details (floating, folds away) */}
      <RightSidebar
        open={rightOpen}
        panelBox={panelBox}
        collapsedStyle={{ top: TOP, right: EDGE }}
        width={rightW}
        alert={alert}
        scene={scene}
        selectedParty={party?.id ?? null}
        onAsk={ask}
        onOpen={openRight}
        onClose={() => setRight(false)}
      />

      {/* Timeline (floating) */}
      <div ref={timelineRef} className="absolute z-20" style={{ left: EDGE, right: EDGE, bottom: EDGE }}>
        <Timeline
          t={t}
          alerts={windowAlerts}
          win={win}
          onWin={setWin}
          selectedId={alert?.id ?? null}
          live={!replaying}
          onSelect={show}
          onLive={() => {
            setReplaying(false)
            const newest = list[0]
            if (newest) setSelectedId(newest.id)
          }}
          replaying={replaying}
          onReplay={() => {
            if (replaying) {
              setReplaying(false)
              return
            }
            const first = [...windowAlerts].sort((a, b) => String(a['@timestamp'] ?? '').localeCompare(String(b['@timestamp'] ?? '')))[0]
            if (first) {
              setReplaying(true)
              setSelectedId(first.id)
            }
          }}
        />
      </div>
    </div>
  )
}
