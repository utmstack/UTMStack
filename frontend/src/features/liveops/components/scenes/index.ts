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
import { Ransomware } from './Ransomware'
import { Persistence } from './Persistence'
import { Exfil } from './Exfil'
import { DefenseEvasion } from './DefenseEvasion'
import { PrivEsc } from './PrivEsc'
import { UserExecution } from './UserExecution'
import { SoftwareDeploy } from './SoftwareDeploy'
import { Phishing } from './Phishing'
import { ProcessInjection } from './ProcessInjection'
import { CredDump } from './CredDump'
import { DataDestruction } from './DataDestruction'
import { ToolDownload } from './ToolDownload'
import { C2Tunnel } from './C2Tunnel'
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
  ransomware: Ransomware,
  persistence: Persistence,
  exfil: Exfil,
  defenseEvasion: DefenseEvasion,
  privEsc: PrivEsc,
  userExec: UserExecution,
  softwareDeploy: SoftwareDeploy,
  phishing: Phishing,
  processInjection: ProcessInjection,
  credDump: CredDump,
  dataDestruction: DataDestruction,
  toolDownload: ToolDownload,
  c2tunnel: C2Tunnel,
  generic: Generic,
}
