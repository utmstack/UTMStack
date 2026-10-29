import type {
  Integration,
  DeployKind,
  IngestType,
  IntegrationGuideChannel,
  IntegrationLogFilter,
} from '@/features/integrations/types'

/**
 * Integrations that exist in the product but not in the backend `integrations`
 * catalog (that table is seeded once in the DB and this branch cannot write to
 * it). They are injected client-side so they show up as catalog cards with their
 * setup guides, exactly like the backend-owned ones.
 *
 * These frontend-owned demonstration entries fill gaps in the backend catalog.
 * The API row wins when one exists; otherwise the stable injected entry keeps
 * its icon, category, guide channel and log filter available in the grid.
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
  { id: '8b9c0d1e-2f3a-4b5c-8d9e-0f1a2b3c4d5e', name: 'SQL Server Event Logs', moduleName: 'WINDOWS_SQL_SERVER', dataType: 'wineventlog', kind: 'other', category: 'database', icon: 'windows.png', defaultPort: 'agent (local)' },
  { id: '9c0d1e2f-3a4b-5c6d-9e0f-1a2b3c4d5e6f', name: 'Windows File Access Auditing', moduleName: 'WINDOWS_FILE_AUDITING', dataType: 'wineventlog', kind: 'other', category: 'operating-system', icon: 'windows.png', defaultPort: 'agent (local)' },
  { id: 'ad1e2f3a-4b5c-6d7e-8f90-1a2b3c4d5e6f', name: 'NetFlow Exporter', moduleName: 'NETFLOW', dataType: 'netflow', kind: 'other', category: 'network', icon: 'netflow.svg', defaultPort: '2055/udp' },
]

interface GuideMetadata {
  channel?: IntegrationGuideChannel
  ingestType: IngestType
  dataType?: string
  forwarderDataType?: string
  defaultPort?: string
  filters?: IntegrationLogFilter[]
  filterField?: 'dataSource' | 'log.exporter'
  setting?: string
}

const SYSLOG_FORWARDER: GuideMetadata = {
  channel: 'syslog',
  ingestType: 'forwarder',
  dataType: 'syslog',
  forwarderDataType: 'syslog',
  defaultPort: '7014/tcp',
  filterField: 'dataSource',
}

const GUIDE_METADATA: Record<string, GuideMetadata> = {
  APACHE: SYSLOG_FORWARDER,
  APACHE2: SYSLOG_FORWARDER,
  ELASTICSEARCH: SYSLOG_FORWARDER,
  HAPROXY: SYSLOG_FORWARDER,
  KAFKA: SYSLOG_FORWARDER,
  KIBANA: SYSLOG_FORWARDER,
  LOGSTASH: SYSLOG_FORWARDER,
  MONGODB: SYSLOG_FORWARDER,
  MYSQL: SYSLOG_FORWARDER,
  NATS: SYSLOG_FORWARDER,
  NGINX: {
    channel: 'linux-agent',
    ingestType: 'agent',
    dataType: 'linux',
    filters: [{ field: 'log.syslogIdentifier', operator: 'IS_ONE_OF_TERMS', value: ['nginx'] }],
    setting: 'Configure Nginx access and error logs to be forwarded to the Linux journal or syslog.',
  },
  OSQUERY: {
    ingestType: 'agent',
    setting: 'Install the UTMStack agent on the host and configure it to collect the osquery results.',
  },
  POSTGRESQL: SYSLOG_FORWARDER,
  REDIS: SYSLOG_FORWARDER,
  SYSLOG: SYSLOG_FORWARDER,
  TRAEFIK: SYSLOG_FORWARDER,
  NETFLOW: {
    channel: 'netflow',
    ingestType: 'forwarder',
    dataType: 'netflow',
    forwarderDataType: 'netflow',
    defaultPort: '2055/udp',
    filterField: 'log.exporter',
  },
  JSON: {
    channel: 'syslog',
    ingestType: 'forwarder',
    dataType: 'syslog',
    forwarderDataType: 'syslog',
    defaultPort: '7014/tcp',
    filterField: 'dataSource',
    setting: 'The legacy JSON HTTP endpoint is not published in v12. This demo entry uses the supported syslog channel; send JSON text as a syslog message.',
  },
  IIS: {
    ingestType: 'agent',
    setting: 'Install the Windows UTMStack agent on the IIS host and configure it to collect IIS log files.',
  },
  WINDOWS_SQL_SERVER: {
    channel: 'windows-agent',
    ingestType: 'agent',
    filters: [
      { field: 'log.providerName', operator: 'IS_ONE_OF_TERMS', value: ['MSSQLSERVER'] },
      { field: 'log.channel', operator: 'IS_ONE_OF_TERMS', value: ['Application'] },
    ],
    setting: 'Enable SQL Server auditing and write its events to the Windows Application event log. The event provider name depends on the SQL Server instance configuration.',
  },
  WINDOWS_FILE_AUDITING: {
    channel: 'windows-agent',
    ingestType: 'agent',
    filters: [
      { field: 'log.providerName', operator: 'IS_ONE_OF_TERMS', value: ['Microsoft-Windows-Security-Auditing'] },
      { field: 'log.channel', operator: 'IS_ONE_OF_TERMS', value: ['Security'] },
    ],
    setting: 'Enable File System auditing in Advanced Audit Policy and configure auditing entries on the folders to monitor.',
  },
}

export const INJECTED_MODULES: Integration[] = SEEDS.map((s) => {
  const guide = GUIDE_METADATA[s.moduleName]
  if (!guide) {
    throw new Error(`Missing guide metadata for injected integration "${s.moduleName}"`)
  }

  return {
    id: s.id,
    name: s.name,
    moduleName: s.moduleName,
    dataType: guide.dataType ?? s.dataType,
    ingestType: guide.ingestType,
    kind: guide.ingestType === 'agent' ? 'agents & syslog' : s.kind,
    status: 'available',
    description: guide.setting ?? '',
    category: s.category,
    logo: `/integrations/${s.icon}`,
    systemOwner: true,
    guideOnly: true,
    defaultPort: guide.defaultPort ?? s.defaultPort,
    guideChannel: guide.channel,
    forwarderDataType: guide.forwarderDataType,
    logFilters: guide.filters,
    logFilterField: guide.filterField,
    guideSetting: guide.setting,
    events24h: undefined,
    rate: undefined,
    darkInvert: false,
  }
})

export function applyInjectedGuideMetadata(integration: Integration): Integration {
  const moduleName = integration.moduleName?.toUpperCase()
  const guide = moduleName ? GUIDE_METADATA[moduleName] : undefined
  if (!guide) return integration

  return {
    ...integration,
    ...(integration.dataType || !guide.dataType ? {} : { dataType: guide.dataType }),
    ...(integration.ingestType ? {} : { ingestType: guide.ingestType }),
    ...(guide.channel ? { guideChannel: guide.channel } : {}),
    ...(guide.forwarderDataType ? { forwarderDataType: guide.forwarderDataType } : {}),
    ...(integration.defaultPort || !guide.defaultPort ? {} : { defaultPort: guide.defaultPort }),
    ...(guide.filters ? { logFilters: guide.filters } : {}),
    ...(guide.filterField ? { logFilterField: guide.filterField } : {}),
    ...(guide.setting ? { guideSetting: guide.setting } : {}),
    guideOnly: true,
  }
}

export function mergeInjectedIntegrations(
  backendIntegrations: Integration[],
  injectedIntegrations: Integration[] = INJECTED_MODULES,
): Integration[] {
  const backendModuleNames = new Set(
    backendIntegrations.map((integration) => integration.moduleName?.toUpperCase()).filter(Boolean),
  )
  const normalizedBackend = backendIntegrations.map(applyInjectedGuideMetadata)
  const missingInjected = injectedIntegrations.filter(
    (integration) => !backendModuleNames.has(integration.moduleName?.toUpperCase()),
  )

  return [...normalizedBackend, ...missingInjected]
}
