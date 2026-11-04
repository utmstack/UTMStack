import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Pill, Burst, AlertBell, Zone } from '..'
import { prog, type Pt } from '../geometry'
import { type SceneProps, party } from './SceneProps'

// A grid of data items: each one is destroyed in turn (fades to a crossed
// empty slot), no encryption, no note — the data is simply gone.
export function DataDestruction(props: SceneProps) {
  const { scene, t, tr } = props
  const tone = usePalette()
  const f = scene.facts as { items: number }
  const host = party(scene, 'host')!
  const H0: Pt = { x: 150, y: 260 }
  const per = Math.min(900, 3600 / f.items)
  const last = f.items - 1
  const rows = Math.ceil(f.items / 2)
  const cell = (i: number) => ({ x: 380 + (i % 2) * 180, y: rows === 1 ? 260 : 140 + Math.floor(i / 2) * (270 / (rows - 1)) })
  const gone = (i: number) => t > 600 + i * per + 600
  const all = t > 600 + last * per + 900
  return (
    <g>
      <Zone x={40} y={70} w={280} h={380} label={tr('liveOps.scHost', { name: host.name })} tone={tone.internal} />
      <Zone x={350} y={70} w={614} h={380} label={tr('liveOps.scData')} tone={tone.muted} />
      {Array.from({ length: f.items }, (_, i) => {
        const o = prog(t, 300 + i * 250, 800 + i * 250)
        if (o <= 0) return null
        const r = cell(i)
        const dead = gone(i)
        return (
          <g key={i} opacity={o}>
            <rect x={r.x} y={r.y - 32} width={140} height={64} rx={8} fill={dead ? 'none' : tone.surface} stroke={dead ? tone.line : tone.textMuted} strokeWidth={1.2} strokeDasharray={dead ? '5 5' : undefined} />
            <text x={r.x + 12} y={r.y + 5} fill={dead ? tone.textFaint : tone.textSoft} fontSize={11.5} fontFamily="JetBrains Mono, monospace" fontWeight={600}>
              {dead ? tr('liveOps.scGone') : tr('liveOps.scDataItem', { i: i + 1 })}
            </text>
            {dead && <line x1={r.x} y1={r.y - 32} x2={r.x + 140} y2={r.y + 32} stroke={tone.alert} strokeWidth={2.2} opacity={prog(t, 600 + i * per + 600, 900 + i * per + 600)} />}
          </g>
        )
      })}
      {all && (
        <g opacity={prog(t, 600 + last * per + 900, 600 + last * per + 1500)}>
          <Burst at={cell(last / 2)} t={t} start={600 + last * per + 600} tone={tone.alert} size={80} dur={900} />
          <Pill x={620} y={420} text={tr('liveOps.scNoRecovery')} tone={tone.alert} />
        </g>
      )}
      <MachineNode party={host} at={H0} t={t} state={all ? 'compromised' : 'active'} selected={props.selectedId === host.id} />
      <AlertBell at={{ x: H0.x + 70, y: H0.y - 118 }} t={t} start={8600} />
    </g>
  )
}
