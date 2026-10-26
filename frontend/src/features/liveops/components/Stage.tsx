// The Live Operations stage: the alert's scene fills the whole page as a living wall; it runs on its
// own (no player controls). The page's panels float on top, so the scene is drawn inside the space
// they leave free (`area`), with the alert's name above it and the story told in a caption below it.
// The three are sized and placed together: the scene as large as the free space allows, and the
// group centred in it, so no screen size leaves a gap between the scene and its caption.
import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { cn } from '@/shared/lib/utils'
import { SEV_META } from '@/features/alerts/lib/alert-meta'
import type { Alert } from '@/features/alerts/types/alert.types'
import type { Party, Scene } from '@/features/liveops/model'
import { H, W } from './geometry'
import { SceneDefs } from './SceneDefs'
import { SCENES } from './scenes'

const sevPill = (severity: string) => (SEV_META[severity as keyof typeof SEV_META] ?? SEV_META.low).pill

// The part of the drawing the scenes use (see the viewBox below), and its proportions.
const VIEW = { x: 16, y: 50, w: W - 32, h: H - 76 }
const ASPECT = VIEW.w / VIEW.h
/** Space between the title, the scene and the caption (px). */
const GAP = 14
/** The scene stops growing at this scale, so a very large screen does not get giant machines. */
const MAX_SCALE = 1.5

/** Height of an element, kept up to date (0 until it is measured). */
function useHeight<T extends HTMLElement>(fallback: number) {
  const ref = useRef<T>(null)
  const [h, setH] = useState(fallback)
  useLayoutEffect(() => {
    const el = ref.current
    if (!el) return
    const ro = new ResizeObserver(() => setH(el.getBoundingClientRect().height))
    ro.observe(el)
    setH(el.getBoundingClientRect().height)
    return () => ro.disconnect()
  }, [])
  return [ref, h] as const
}

/** Size of the stage itself. */
function useBox<T extends HTMLElement>() {
  const ref = useRef<T>(null)
  const [box, setBox] = useState({ width: 0, height: 0 })
  useLayoutEffect(() => {
    const el = ref.current
    if (!el) return
    const ro = new ResizeObserver(([e]) => setBox({ width: e.contentRect.width, height: e.contentRect.height }))
    ro.observe(el)
    const r = el.getBoundingClientRect()
    setBox({ width: r.width, height: r.height })
    return () => ro.disconnect()
  }, [])
  return [ref, box] as const
}

/** Scene time in ms since the alert came on stage. */
function useSceneClock() {
  const [t, setT] = useState(0)
  useEffect(() => {
    let raf = 0
    const start = performance.now()
    const tick = (now: number) => {
      setT(now - start)
      raf = requestAnimationFrame(tick)
    }
    raf = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(raf)
  }, [])
  return t
}

export interface StageArea {
  top: number
  right: number
  bottom: number
  left: number
}

/** Mount with a key per alert: a new alert (or the same one again) starts its story from the top. */
export function Stage({
  alert,
  scene,
  templateLabel,
  area,
  selectedParty,
  onSelectParty,
  onEnded,
  blocked,
}: {
  alert: Alert
  scene: Scene
  /** The template's translated label (Live Operations chrome, i18n in the page). */
  templateLabel: string
  /** Space (px from each edge) the floating panels leave free for the scene. */
  area: StageArea
  selectedParty?: string
  onSelectParty: (p: Party) => void
  /** Called once when the story has finished (the replay of a time window moves on). */
  onEnded?: () => void
  /** The visitor blocked an address on a firewall: from that moment the scene shows the block. */
  blocked?: { ip: string; on: string }
}) {
  const t = useSceneClock()
  const ended = useRef(false)
  const [blockedAt, setBlockedAt] = useState<number | undefined>(undefined)
  useEffect(() => {
    if (blocked && blockedAt === undefined) setBlockedAt(t)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [blocked])
  useEffect(() => {
    if (!ended.current && t >= scene.duration + 1500) {
      ended.current = true
      onEnded?.()
    }
  }, [t, scene.duration, onEnded])

  const Scene = SCENES[scene.template]
  const blockCaption = blocked && blockedAt !== undefined && t >= blockedAt ? { at: blockedAt, text: `${blocked.ip} is now blocked on ${blocked.on}: the traffic stops here.` } : null
  const caption = blockCaption ?? [...scene.captions].reverse().find((c) => t >= c.at) ?? scene.captions[0]
  const capIndex = blockCaption ? scene.captions.length - 1 : scene.captions.indexOf(caption)
  // How far the story is into the current caption, for the progress bar.
  const capEnd = blockCaption ? blockCaption.at + 1 : scene.captions[capIndex + 1]?.at ?? scene.duration
  const capStart = blockCaption ? blockCaption.at : caption.at
  const capFill = blockCaption ? 1 : Math.min(1, Math.max(0, (t - capStart) / Math.max(1, capEnd - capStart)))
  const time = alert['@timestamp'] ? new Date(alert['@timestamp']) : new Date()

  // Place the title, the scene and the caption as one group inside the free space.
  const [stageRef, box] = useBox<HTMLDivElement>()
  const [titleRef, titleH] = useHeight<HTMLDivElement>(52)
  const [capRef, capH] = useHeight<HTMLDivElement>(46)
  const freeW = Math.max(0, box.width - area.left - area.right)
  const freeH = Math.max(0, box.height - area.top - area.bottom)
  const sceneW = Math.max(0, Math.min(freeW - 16, (freeH - titleH - capH - 2 * GAP) * ASPECT, VIEW.w * MAX_SCALE))
  const sceneH = sceneW / ASPECT
  const groupH = titleH + GAP + sceneH + GAP + capH
  // A little above the middle reads as centred (the eye sits high on a screen).
  const top = area.top + Math.max(0, (freeH - groupH) * 0.42)
  const sceneTop = top + titleH + GAP
  const capTop = sceneTop + sceneH + GAP
  const slide = 'left 0.35s ease, right 0.35s ease, top 0.35s ease, width 0.35s ease, height 0.35s ease'

  return (
    <div ref={stageRef} className="lo-stage absolute inset-0">
      {/* The alert on stage */}
      <div ref={titleRef} className="pointer-events-none absolute flex justify-center px-4" style={{ left: area.left, right: area.right, top, transition: slide }}>
        <div key={alert.id} className="lo-caption lo-panel flex min-w-0 max-w-full items-center gap-3 rounded-xl px-3.5 py-2">
          <span className={cn('shrink-0 rounded-md px-1.5 py-0.5 text-[10.5px] font-semibold uppercase tracking-wide ring-1 ring-inset', sevPill(String(alert.severity)))}>{alert.severity}</span>
          <div className="min-w-0">
            <div className="truncate text-[14px] font-semibold text-foreground" title={alert.name}>
              {alert.name}
            </div>
            <div className="truncate text-[11.5px] text-muted-foreground">
              {templateLabel} · {alert.technique} · {time.toLocaleString(undefined, { month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })} · {alert.dataSource}
            </div>
          </div>
        </div>
      </div>

      {/* The scene */}
      <svg
        viewBox={`${VIEW.x} ${VIEW.y} ${VIEW.w} ${VIEW.h}`}
        className="lo-scene absolute"
        style={{ left: area.left + (freeW - sceneW) / 2, top: sceneTop, width: sceneW, height: sceneH, transition: slide }}
        preserveAspectRatio="xMidYMid meet"
        role="img"
        aria-label={scene.title}
      >
        <SceneDefs />
        <Scene scene={scene} t={t} selectedId={selectedParty} onSelect={onSelectParty} blockedAt={blockedAt} />
      </svg>

      {/* The story */}
      <div ref={capRef} className="pointer-events-none absolute flex justify-center px-4" style={{ left: area.left, right: area.right, top: capTop, transition: slide }}>
        <div className="lo-panel flex max-w-[760px] items-center gap-3 rounded-xl px-4 py-2.5">
          {/* One bar per step of the story; the current one fills as it plays. */}
          <div className="flex shrink-0 gap-1">
            {scene.captions.map((_, i) => (
              <span key={i} className="relative h-1.5 w-4 overflow-hidden rounded-full bg-muted-foreground/25">
                <span className="absolute inset-y-0 left-0 rounded-full bg-primary" style={{ width: `${(i < capIndex ? 1 : i === capIndex ? capFill : 0) * 100}%` }} />
              </span>
            ))}
          </div>
          <p key={capIndex + (blockCaption ? '-blocked' : '')} className="lo-caption min-w-0 text-[13.5px] leading-snug text-foreground">
            {caption.text}
          </p>
        </div>
      </div>
    </div>
  )
}
