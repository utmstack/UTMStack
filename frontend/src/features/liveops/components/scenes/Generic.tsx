import { usePalette } from '@/features/liveops/hooks/usePalette'
import { MachineNode, Wire, Packet, AlertBell } from '..'
import { stream, type Pt } from '../geometry'
import { type SceneProps, party, sel } from './SceneProps'

export function Generic(props: SceneProps) {
  const { scene, t } = props
  const tone = usePalette()
  const src = party(scene, 'attacker') ?? party(scene, 'source')!
  const tgt = party(scene, 'target')!
  const S: Pt = { x: 220, y: 260 }
  const T: Pt = { x: 780, y: 260 }
  return (
    <g>
      <Wire a={S} b={T} tone={tone[src.role]} t={t} />
      {stream(t, 300, 1e12, 700, 1300).map((p) => (
        <Packet key={p.i} a={S} b={T} u={p.u} tone={tone[src.role]} />
      ))}
      <MachineNode party={src} at={S} t={t} state="active" selected={props.selectedId === src.id} onSelect={sel(props, src)} />
      <MachineNode party={tgt} at={T} t={t} state="targeted" selected={props.selectedId === tgt.id} onSelect={sel(props, tgt)} />
      <AlertBell at={{ x: T.x + 60, y: T.y - 110 }} t={t} start={6500} />
    </g>
  )
}
