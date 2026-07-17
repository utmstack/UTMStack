export type Proto = 'udp' | 'tcp' | 'tls' | 'http' | 'https'
export type HttpAuth = '' | 'bearer' | 'hmac'

interface ProtoPortEntry {
  udp?: string
  tcp?: string
}

// Mirrors config.ProtoPorts (syslog/network data types).
const PROTO_PORTS: Record<string, ProtoPortEntry> = {
  syslog: { udp: '7014', tcp: '7014' },
  'vmware-esxi': { udp: '7002', tcp: '7002' },
  'antivirus-esmc-eset': { udp: '7003', tcp: '7003' },
  'antivirus-kaspersky': { udp: '7004', tcp: '7004' },
  'firewall-cisco-asa': { udp: '514', tcp: '1470' },
  'firewall-cisco-firepower': { udp: '514', tcp: '1470' },
  'cisco-switch': { udp: '514', tcp: '1470' },
  'firewall-meraki': { udp: '514', tcp: '1470' },
  'firewall-fortigate-traffic': { udp: '7005', tcp: '7005' },
  'firewall-paloalto': { udp: '7006', tcp: '7006' },
  'firewall-mikrotik': { udp: '7007', tcp: '7007' },
  'firewall-sophos-xg': { udp: '7008', tcp: '7008' },
  'firewall-sonicwall': { udp: '7009', tcp: '7009' },
  'deceptive-bytes': { udp: '7010', tcp: '7010' },
  'antivirus-sentinel-one': { udp: '7012', tcp: '7012' },
  'ibm-aix': { udp: '7016', tcp: '7016' },
  'firewall-pfsense': { udp: '7017', tcp: '7017' },
  'firewall-fortiweb': { udp: '7018', tcp: '7018' },
  suricata: { udp: '7019', tcp: '7019' },
  netflow: { udp: '2055' }, // no TCP entry on the Go side — flow data is UDP-only, TLS not offered
  oracle: { udp: '7021', tcp: '7021' },
}

interface HttpSpec {
  proto: 'http' | 'https'
  port: string
  path: string
  auth: HttpAuth
  signatureHeader?: string
}

// Mirrors config.HTTPPorts (HTTP/HTTPS-native data types).
const HTTP_CATALOG: Record<string, HttpSpec> = {
  github: {
    proto: 'https',
    port: '7020',
    path: '/github',
    auth: 'hmac',
    signatureHeader: 'X-Hub-Signature-256',
  },
}

export function availableProtosFor(dataType: string): Proto[] {
  const protos: Proto[] = []
  const pp = PROTO_PORTS[dataType]
  if (pp) {
    if (pp.udp) protos.push('udp')
    if (pp.tcp) {
      protos.push('tcp')
      protos.push('tls')
    }
  }
  const http = HTTP_CATALOG[dataType]
  if (http && !protos.includes(http.proto)) protos.push(http.proto)

  return protos.length > 0 ? protos : ['udp', 'tcp', 'tls', 'http', 'https']
}

export function defaultPortFor(dataType: string, proto: Proto): string {
  const pp = PROTO_PORTS[dataType]
  if (pp) {
    if (proto === 'udp' && pp.udp) return pp.udp
    if ((proto === 'tcp' || proto === 'tls') && pp.tcp) return pp.tcp
  }
  const http = HTTP_CATALOG[dataType]
  if (http && (proto === 'http' || proto === 'https')) return http.port
  return ''
}

export function httpDefaultsFor(dataType: string): { path: string; auth: HttpAuth; signatureHeader: string } | null {
  const http = HTTP_CATALOG[dataType]
  if (!http) return null
  return { path: http.path, auth: http.auth, signatureHeader: http.signatureHeader ?? '' }
}
