import { usePalette } from '@/features/liveops/hooks/usePalette'
import { useGlow } from '@/features/liveops/hooks/useGlow'
import { MachineNode, Wire, Packet, Pill, Burst, AlertBell } from '..'
import { prog, type Pt } from '../geometry'
import { type SceneProps, party, sel } from './SceneProps'

const LAND: [number, number][][] = [
  [[-168, 66], [-140, 70], [-95, 72], [-75, 62], [-60, 50], [-80, 30], [-97, 25], [-105, 20], [-118, 32], [-125, 48], [-150, 60]],
  [[-80, 10], [-60, 8], [-35, -8], [-40, -22], [-55, -35], [-70, -52], [-75, -40], [-81, -5]],
  [[-10, 36], [0, 44], [-5, 48], [8, 55], [20, 60], [30, 70], [45, 68], [40, 45], [28, 40], [12, 38]],
  [[-17, 21], [-5, 36], [11, 37], [32, 31], [43, 12], [51, 11], [40, -15], [30, -34], [18, -34], [12, -5], [-10, 5]],
  [[40, 45], [55, 70], [100, 77], [140, 72], [170, 66], [140, 50], [122, 30], [108, 18], [100, 2], [80, 8], [72, 22], [58, 25], [48, 30]],
  [[114, -22], [130, -12], [145, -15], [153, -28], [146, -39], [130, -32], [115, -34]],
]

function inside(lon: number, lat: number, poly: [number, number][]) {
  let c = false
  for (let i = 0, j = poly.length - 1; i < poly.length; j = i++) {
    const [xi, yi] = poly[i]
    const [xj, yj] = poly[j]
    if (yi > lat !== yj > lat && lon < ((xj - xi) * (lat - yi)) / (yj - yi) + xi) c = !c
  }
  return c
}

const proj = (lon: number, lat: number): Pt => ({ x: 70 + ((lon + 180) / 360) * 860, y: 70 + ((80 - lat) / 140) * 400 })

let dots: Pt[] | null = null
function landDots() {
  if (dots) return dots
  dots = []
  for (let lat = 76; lat >= -56; lat -= 4) {
    for (let lon = -170; lon <= 178; lon += 4) {
      if (LAND.some((p) => inside(lon, lat, p))) dots.push(proj(lon, lat))
    }
  }
  return dots
}

export function Travel(props: SceneProps) {
  const { scene, t } = props
  const tone = usePalette()
  const glow = useGlow()
  const f = scene.facts as { user: string; km: number; gap: number; far?: { city?: string; country?: string; latitude?: number; longitude?: number }; home: { city: string; lat: number; lon: number } }
  const attacker = party(scene, 'attacker')!
  const cloud = party(scene, 'cloud')!
  const homeP = proj(f.home.lon, f.home.lat)
  const farP = proj(f.far?.longitude ?? 3.4, f.far?.latitude ?? 6.5)
  // The identity provider goes over open sea (Arctic, Atlantic or Indian Ocean), wherever its name
  // stays farthest from both sign-in points, so a sign-in from Europe never lands on its label.
  const spots: Pt[] = [{ x: 500, y: 92 }, { x: 412, y: 232 }, { x: 640, y: 352 }]
  const room = (c: Pt) => Math.min(...[homeP, farP].map((q) => Math.hypot(q.x - c.x, (q.y - (c.y + 40)) * 1.6)))
  const C: Pt = spots.reduce((best, c) => (room(c) > room(best) + 40 ? c : best), spots[0])
  const p1 = prog(t, 400, 2200)
  const p2 = prog(t, 3200, 5000)
  const arc = prog(t, 6200, 7600)
  return (
    <g>
      {landDots().map((d, i) => (
        <circle key={i} cx={d.x} cy={d.y} r={2.1} fill={tone.line} />
      ))}
      <Wire a={homeP} b={C} tone={tone.internal} t={t} bend={40} appear={400} />
      <Wire a={farP} b={C} tone={tone.alert} t={t} bend={40} appear={3200} />
      {p1 > 0 && p1 < 1 && <Packet a={homeP} b={C} u={p1} tone={tone.internal} bend={40} label="✓ sign-in" />}
      {p2 > 0 && p2 < 1 && <Packet a={farP} b={C} u={p2} tone={tone.alert} bend={40} label="✓ sign-in" />}
      {arc > 0 && (
        <path d={`M${homeP.x},${homeP.y} Q${(homeP.x + farP.x) / 2},${Math.min(homeP.y, farP.y) - 40} ${farP.x},${farP.y}`} fill="none" stroke={tone.highlight} strokeWidth={2} strokeDasharray="1000" strokeDashoffset={1000 * (1 - arc)} />
      )}
      {arc >= 1 && <Pill x={farP.x} y={farP.y + 52} text={`${f.km.toLocaleString()} km in ${f.gap} min`} tone={tone.highlight} />}
      <g opacity={prog(t, 200, 600)}>
        <circle cx={homeP.x} cy={homeP.y} r={9} fill={tone.internal} filter={glow} />
        <Pill x={homeP.x} y={homeP.y + 26} text={`${f.home.city} · ${f.user}`} tone={tone.internal} small />
      </g>
      <g opacity={prog(t, 3000, 3400)}>
        <circle cx={farP.x} cy={farP.y} r={9} fill={tone.alert} filter={glow} />
        <Burst at={farP} t={t} start={3200} tone={tone.alert} size={40} dur={900} />
        <Pill x={farP.x} y={farP.y + 26} text={`${attacker.flag ?? ''} ${f.far?.city ?? 'abroad'} · ${attacker.name}`} tone={tone.alert} small />
      </g>
      <MachineNode party={cloud} at={C} t={t} r={30} state={t > 5000 ? 'targeted' : 'idle'} selected={props.selectedId === cloud.id} onSelect={sel(props, cloud)} />
      <AlertBell at={{ x: C.x + 170, y: C.y - 4 }} t={t} start={9200} />
    </g>
  )
}