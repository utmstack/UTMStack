import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { AlertTriangle, ExternalLink, List, PanelLeftClose, PanelRightClose, Radio, SlidersHorizontal, Sparkles } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useTranslation } from 'react-i18next'
import { useSocAi, useSocAiFocus } from '@/features/soc-ai'
import { alertFocus } from '@/features/alerts/lib/alert-focus'
import type { Alert } from '@/features/alerts/types/alert.types'
import { alertsHttpService } from '@/features/alerts/services/alerts-http.service'
import { buildScene, templateFor, type Party } from '../model'
import { Stage } from '../components/Stage'
import { Timeline } from '../components/Timeline'
import { SEV_META } from '@/features/alerts/lib/alert-meta'
import '../liveops.css'

// Severity and role colours as the rest of the console shows them (Alerts page, Adversary view).
const sevBar = (severity: string) => (SEV_META[severity as keyof typeof SEV_META] ?? SEV_META.low).bar
const ROLE_DOT: Record<string, string> = {
  attacker: 'bg-red-500',
  external: 'bg-red-500',
  perimeter: 'bg-amber-500',
  cloud: 'bg-violet-500',
  internal: 'bg-primary',
}

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

function ago(iso: string | undefined, t: (k: string, o?: Record<string, unknown>) => string) {
  if (!iso) return t('liveOps.agoJust')
  const s = Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 1000))
  if (s < 60) return t('liveOps.agoJust')
  if (s < 3600) return t('liveOps.agoMin', { count: Math.round(s / 60) })
  return t('liveOps.agoH', { count: Math.round(s / 3600) })
}

/** The alert rule's description as plain text, folded to a few lines until opened. */
function Description({ text }: { text: string }) {
  const [open, setOpen] = useState(false)
  const plain = text.replace(/<br\s*\/?>/gi, ' ').replace(/<[^>]+>/g, '').replace(/\s+/g, ' ').trim()
  const long = plain.length > 200
  return (
    <div className="mt-1.5 text-[12px] leading-relaxed text-foreground/80">
      <p className={cn(!open && long && 'line-clamp-4')}>{plain}</p>
      {long && (
        <button type="button" onClick={() => setOpen((o) => !o)} className="mt-0.5 text-[11.5px] font-medium text-primary hover:underline">
          {open ? 'Less' : 'More'}
        </button>
      )}
    </div>
  )
}

export function LiveOpsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const soc = useSocAi()
  const [alerts, setAlerts] = useState<Alert[] | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [hours, setHours] = useState(6)
  const [party, setParty] = useState<Party | null>(null)
  const [tick, setTick] = useState(0)
  const [error, setError] = useState('')
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

  const refresh = useCallback(() =>
    alertsHttpService
      .list({ page: 0, size: 500, filters: [{ field: 'parentId', operator: 'IS', value: '' }] })
      .then((r) => {
        setAlerts(r.data)
        setError('')
      })
      .catch((e: unknown) => setError(e instanceof Error ? e.message : String(e))), [])

  useEffect(() => {
    void refresh()
    const id = setInterval(() => void refresh(), POLL_MS)
    return () => clearInterval(id)
  }, [refresh])
  useEffect(() => {
    const id = setInterval(() => setTick((x) => x + 1), 5_000)
    return () => clearInterval(id)
  }, [])

  const list = useMemo(
    () => (alerts ?? []).map((a) => ({ ...a })).sort((a, b) => String(b['@timestamp'] ?? '').localeCompare(String(a['@timestamp'] ?? ''))),
    [alerts],
  )
  const windowAlerts = useMemo(
    () => list.filter((a) => a['@timestamp'] && new Date(a['@timestamp']).getTime() >= Date.now() - hours * 3_600_000),
    [list, hours, tick],
  )
  const alert = list.find((a) => a.id === selectedId) ?? list[list.length - 1]
  const scene = useMemo(() => (alert ? buildScene(alert) : null), [alert])

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
        <div className="absolute inset-0 flex items-center justify-center text-sm text-muted-foreground">
          {error ? <span className="flex items-center gap-2 text-red-600 dark:text-red-400"><AlertTriangle size={14} /> {t('liveOps.error', { msg: error })}</span> : t('liveOps.loading')}
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
            [t('liveOps.statLastHour'), list.filter((a) => a['@timestamp'] && new Date(a['@timestamp']).getTime() >= Date.now() - 3_600_000).length],
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
      {leftOpen ? (
        <div className="lo-panel absolute z-20 flex flex-col overflow-hidden rounded-xl" style={{ ...panelBox, left: EDGE, width: LEFT_W }}>
          <div className="flex shrink-0 items-center gap-2 border-b border-border px-3 py-2">
            <span className="text-[12px] font-semibold text-foreground">{t('liveOps.listTitle')}</span>
            <span className="text-[11px] text-muted-foreground">{t('liveOps.listShown', { count: list.length })}</span>
            <button type="button" onClick={() => setLeft(false)} className="ml-auto rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground" title={t('liveOps.listHide')}>
              <PanelLeftClose size={14} />
            </button>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto">
            {list.slice(0, 200).map((a) => {
              const sel = a.id === alert?.id
              return (
                <button
                  key={a.id}
                  type="button"
                  onClick={() => show(a.id)}
                  className={cn('relative flex w-full gap-2 border-b border-border/60 px-3 py-2 text-left transition-colors', sel ? 'bg-primary/15' : 'hover:bg-muted/60')}
                  title={sel ? 'Play this alert again' : undefined}
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
          </div>
        </div>
      ) : (
        <button type="button" onClick={() => setLeft(true)} className="lo-panel absolute z-20 flex h-10 w-10 items-center justify-center rounded-xl text-muted-foreground hover:text-foreground" style={{ top: TOP, left: EDGE }} title={t('liveOps.listShow')}>
          <List size={17} />
        </button>
      )}

      {/* Details (floating, folds away) */}
      {rightOpen ? (
        <div className="lo-panel absolute z-20 flex flex-col overflow-hidden rounded-xl" style={{ ...panelBox, right: EDGE, width: rightW }}>
          {alert && scene ? (
            <div className="min-h-0 flex-1 overflow-y-auto">
              <div className="border-b border-border px-3 py-2.5">
                <div className="flex items-center">
                  <span className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">{t('liveOps.onStage')}</span>
                  <button type="button" onClick={() => setRight(false)} className="ml-auto rounded p-1 text-muted-foreground hover:bg-muted hover:text-foreground" title={t('liveOps.hideDetails')}>
                    <PanelRightClose size={14} />
                  </button>
                </div>
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
              </div>

              <div className="border-b border-border px-3 py-2.5">
                <div className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">{t('liveOps.involved')}</div>
                <div className="mt-1.5 flex flex-col gap-1.5">
                  {scene.parties.map((p) => (
                    <div key={p.id} className={cn('flex items-center gap-2 rounded-lg border px-2 py-1.5', party?.id === p.id ? 'border-primary/60 bg-primary/10' : 'border-border bg-muted/30')}>
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
                  {['Explain this attack and tell me what to do next', 'Is this a false positive?', 'What else did this attacker do?'].map((q) => (
                    <button key={q} type="button" onClick={() => ask(q)} className="rounded-lg border border-border bg-muted/30 px-2.5 py-1.5 text-left text-[12px] text-foreground hover:border-primary/40 hover:bg-primary/10">
                      {q}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          ) : null}
        </div>
      ) : (
        <button type="button" onClick={openRight} className="lo-panel absolute z-20 flex h-10 w-10 items-center justify-center rounded-xl text-muted-foreground hover:text-foreground" style={{ top: TOP, right: EDGE }} title={t('liveOps.showDetails')}>
          <SlidersHorizontal size={16} />
        </button>
      )}

      {/* Timeline (floating) */}
      <div ref={timelineRef} className="absolute z-20" style={{ left: EDGE, right: EDGE, bottom: EDGE }}>
        <Timeline
          t={t}
          alerts={windowAlerts}
          hours={hours}
          onHours={setHours}
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
