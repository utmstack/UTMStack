import type { ReactElement } from 'react'
import type { Scene } from '@/features/liveops/model'
import type { SceneProps } from './SceneProps'
import { BruteForce } from './BruteForce'
import { Outbound } from './Outbound'
import { Lateral } from './Lateral'
import { GoldenTicket } from './GoldenTicket'
import { Malware } from './Malware'
import { WebAttack } from './WebAttack'
import { Travel } from './Travel'
import { Scan } from './Scan'
import { CloudChange } from './CloudChange'
import { Ransomware } from './Ransomware'
import { Persistence } from './Persistence'
import { Exfil } from './Exfil'
import { DefenseEvasion } from './DefenseEvasion'
import { PrivEsc } from './PrivEsc'
import { Generic } from './Generic'

export type { SceneProps }

// One animated scene per alert template; the stage picks by scene.template.
export const SCENES: Record<Scene['template'], (p: SceneProps) => ReactElement> = {
  bruteForce: BruteForce,
  outbound: Outbound,
  lateral: Lateral,
  goldenTicket: GoldenTicket,
  malware: Malware,
  webAttack: WebAttack,
  travel: Travel,
  scan: Scan,
  cloud: CloudChange,
  ransomware: Ransomware,
  persistence: Persistence,
  exfil: Exfil,
  defenseEvasion: DefenseEvasion,
  privEsc: PrivEsc,
  generic: Generic,
}
