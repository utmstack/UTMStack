import type { ReactElement } from 'react'
import type { Scene } from '@/features/liveops/model'
import type { SceneProps } from './SceneProps'
import { BruteForce } from './BruteForce'
import { Outbound } from './Outbound'
import { Lateral } from './Lateral'
import { Malware } from './Malware'
import { WebAttack } from './WebAttack'
import { Travel } from './Travel'
import { Scan } from './Scan'
import { CloudChange } from './CloudChange'
import { Generic } from './Generic'

export type { SceneProps }

// One animated scene per alert template; the stage picks by scene.template.
export const SCENES: Record<Scene['template'], (p: SceneProps) => ReactElement> = {
  bruteForce: BruteForce,
  outbound: Outbound,
  lateral: Lateral,
  malware: Malware,
  webAttack: WebAttack,
  travel: Travel,
  scan: Scan,
  cloud: CloudChange,
  generic: Generic,
}
