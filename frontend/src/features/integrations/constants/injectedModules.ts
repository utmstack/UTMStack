import type { Integration, DeployKind } from '@/features/integrations/types'

/**
 * Integrations that exist in the product but not in the backend `integrations`
 * catalog (that table is seeded once in the DB and this branch cannot write to
 * it). They are injected client-side so they show up as catalog cards with their
 * setup guides, exactly like the backend-owned ones.
 *
 * These are the log sources v10/v11 shipped and later dropped (the 2026-02-18
 * removal batch): web servers, databases, app stacks, messaging and proxies,
 * plus the two generic channels SYSLOG and JSON. Each keeps its own `dataType`
 * (no shared parse pipeline — see the plan) and `ingestType: 'forwarder'` so it
 * groups under Collectors and routes to the collector guide in the drawer.
 *
 * `kind` is 'other' for all of them: `KIND_META['other'].group === 'collectors'`,
 * which is what makes IntegrationDrawer render <CollectorSetup>. From there
 * `getCollector(moduleName)` picks the specific guide when one is registered
 * (IIS, SYSLOG, JSON) and falls back to the generic collector guide otherwise.
 *
 * IDs are fixed (not random per render) so React keys and references stay
 * stable across refetches.
 */
interface InjectedSeed {
  id: string
  name: string
  moduleName: string
  dataType: string
  kind: DeployKind
  category: string
  icon: string
  defaultPort: string
}

const SEEDS: InjectedSeed[] = [
  { id: '6b7f1a2c-3d4e-4f5a-9c6b-7a8d9e0f1a2b', name: 'Nginx', moduleName: 'NGINX', dataType: 'nginx', kind: 'other', category: 'web-server', icon: 'nginx.svg', defaultPort: '514/udp' },
  { id: '1c2d3e4f-5a6b-4c7d-8e9f-0a1b2c3d4e5f', name: 'Apache', moduleName: 'APACHE', dataType: 'apache', kind: 'other', category: 'web-server', icon: 'apache.svg', defaultPort: '514/udp' },
  { id: '2d3e4f5a-6b7c-4d8e-9f0a-1b2c3d4e5f6a', name: 'Apache2', moduleName: 'APACHE2', dataType: 'apache2', kind: 'other', category: 'web-server', icon: 'apache2.svg', defaultPort: '514/udp' },
  { id: '3e4f5a6b-7c8d-4e9f-0a1b-2c3d4e5f6a7b', name: 'IIS', moduleName: 'IIS', dataType: 'iis', kind: 'other', category: 'web-server', icon: 'iis.svg', defaultPort: 'filebeat/local' },
  { id: '4f5a6b7c-8d9e-4f0a-1b2c-3d4e5f6a7b8c', name: 'MySQL', moduleName: 'MYSQL', dataType: 'mysql', kind: 'other', category: 'database', icon: 'mysql.svg', defaultPort: '514/udp' },
  { id: '5a6b7c8d-9e0f-4a1b-2c3d-4e5f6a7b8c9d', name: 'PostgreSQL', moduleName: 'POSTGRESQL', dataType: 'postgresql', kind: 'other', category: 'database', icon: 'postgresql.svg', defaultPort: '514/udp' },
  { id: '6b7c8d9e-0f1a-4b2c-3d4e-5f6a7b8c9d0e', name: 'MongoDB', moduleName: 'MONGODB', dataType: 'mongodb', kind: 'other', category: 'database', icon: 'mongodb.svg', defaultPort: '514/udp' },
  { id: '7c8d9e0f-1a2b-4c3d-4e5f-6a7b8c9d0e1f', name: 'Redis', moduleName: 'REDIS', dataType: 'redis', kind: 'other', category: 'database', icon: 'redis.svg', defaultPort: '514/udp' },
  { id: '8d9e0f1a-2b3c-4d4e-5f6a-7b8c9d0e1f2a', name: 'Elasticsearch', moduleName: 'ELASTICSEARCH', dataType: 'elasticsearch', kind: 'other', category: 'database', icon: 'elasticsearch.svg', defaultPort: '514/udp' },
  { id: '9e0f1a2b-3c4d-4e5f-6a7b-8c9d0e1f2a3b', name: 'Logstash', moduleName: 'LOGSTASH', dataType: 'logstash', kind: 'other', category: 'database', icon: 'logstash.svg', defaultPort: '514/udp' },
  { id: '0f1a2b3c-4d5e-4f6a-7b8c-9d0e1f2a3b4c', name: 'Kibana', moduleName: 'KIBANA', dataType: 'kibana', kind: 'other', category: 'database', icon: 'kibana.svg', defaultPort: '514/udp' },
  { id: '1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d', name: 'Kafka', moduleName: 'KAFKA', dataType: 'kafka', kind: 'other', category: 'messaging', icon: 'kafka.svg', defaultPort: '514/udp' },
  { id: '2b3c4d5e-6f7a-4b8c-9d0e-1f2a3b4c5d6e', name: 'NATS', moduleName: 'NATS', dataType: 'nats', kind: 'other', category: 'messaging', icon: 'nats.svg', defaultPort: '514/udp' },
  { id: '3c4d5e6f-7a8b-4c9d-0e1f-2a3b4c5d6e7f', name: 'HAProxy', moduleName: 'HAPROXY', dataType: 'haproxy', kind: 'other', category: 'proxy', icon: 'haproxy.svg', defaultPort: '514/udp' },
  { id: '4d5e6f7a-8b9c-4d0e-1f2a-3b4c5d6e7f8a', name: 'Traefik', moduleName: 'TRAEFIK', dataType: 'traefik', kind: 'other', category: 'proxy', icon: 'traefik.svg', defaultPort: '514/udp' },
  { id: '5e6f7a8b-9c0d-4e1f-2a3b-4c5d6e7f8a9b', name: 'OsQuery', moduleName: 'OSQUERY', dataType: 'osquery', kind: 'other', category: 'other', icon: 'osquery.svg', defaultPort: 'agent (local)' },
  { id: '6f7a8b9c-0d1e-4f2a-3b4c-5d6e7f8a9b0c', name: 'Syslog', moduleName: 'SYSLOG', dataType: 'syslog', kind: 'other', category: 'syslog', icon: 'syslog.svg', defaultPort: '514/udp' },
  { id: '7a8b9c0d-1e2f-4a3b-4c5d-6e7f8a9b0c1d', name: 'JSON Input', moduleName: 'JSON', dataType: 'json-input', kind: 'other', category: 'other', icon: 'json.svg', defaultPort: '8080' },
]

export const INJECTED_MODULES: Integration[] = SEEDS.map((s) => ({
  id: s.id,
  name: s.name,
  moduleName: s.moduleName,
  dataType: s.dataType,
  ingestType: 'forwarder',
  kind: s.kind,
  status: 'available',
  description: '',
  category: s.category,
  logo: `/integrations/${s.icon}`,
  systemOwner: true,
  defaultPort: s.defaultPort,
  events24h: undefined,
  rate: undefined,
  darkInvert: false,
}))
