import type { BreachStatus, DarkWebRecord, OrganizationElementType } from '../types'

interface DemoRecordInput {
  id: string
  organizationElementType: OrganizationElementType
  organizationElement: string
  breachStatus: BreachStatus
  employee: boolean
  email?: string
  domain?: string
  username?: string
  password?: string
  ipAddress?: string
  phone?: string
  name?: string
  timestamp: string
  createdAt: string
  databaseName?: string
  foundIn?: string
}

function demoRecord(input: DemoRecordInput): DarkWebRecord {
  return {
    id: input.id,
    timestamp: input.timestamp,
    createdAt: input.createdAt,
    domain: input.domain ?? null,
    email: input.email ?? null,
    ipAddress: input.ipAddress ?? null,
    username: input.username ?? null,
    password: input.password ?? null,
    hashedPassword: null,
    name: input.name ?? null,
    vin: null,
    address: null,
    phone: input.phone ?? null,
    databaseName: input.databaseName ?? 'Demo source',
    foundIn: input.foundIn ?? 'multiple sources',
    source: 'dark web api',
    leakName: 'Synthetic demonstration breach',
    breachDescription: 'Synthetic demonstration record; no real account or secret is represented.',
    compromisedData: 'email addresses, credentials',
    imageUrl: null,
    victimDomain: input.organizationElement,
    distributionType: 'demo only',
    breachId: `DEMO-${input.id}`,
    organizationId: 10901,
    organizationUUID: '00000000-0000-4000-8000-000000000001',
    organizationElement: input.organizationElement,
    organizationElementType: input.organizationElementType,
    breachStatus: input.breachStatus,
    employee: input.employee,
    echoesCount: 0,
  }
}

export const DEMO_RECORDS: DarkWebRecord[] = [
  demoRecord({ id: 'DW-001', organizationElementType: 'DOMAIN', organizationElement: 'test01.example', breachStatus: 'IN_MITIGATION', employee: true, domain: 'test01.example', email: 'taylor@test01.example', username: 't.smith', password: 'Demo!4821', name: 'Taylor Smith', timestamp: '2026-08-07T00:00:00Z', createdAt: '2026-08-28T12:20:00Z', databaseName: 'example-forum.invalid' }),
  demoRecord({ id: 'DW-002', organizationElementType: 'DOMAIN', organizationElement: 'test01.example', breachStatus: 'IN_MITIGATION', employee: true, domain: 'test01.example', email: 'taylor@test01.example', username: 't.smith', password: 'Test$1902', name: 'Taylor Smith', timestamp: '2026-08-07T00:00:00Z', createdAt: '2026-08-28T12:20:00Z', databaseName: 'archive.invalid' }),
  demoRecord({ id: 'DW-003', organizationElementType: 'DOMAIN', organizationElement: 'test02.example', breachStatus: 'OPEN', employee: true, domain: 'test02.example', email: 'alex@test02.example', username: 'alex.r', password: 'Example#2026', name: 'Alex Rivera', timestamp: '2026-06-14T00:00:00Z', createdAt: '2026-08-26T08:14:00Z', databaseName: 'paste-archive.invalid' }),
  demoRecord({ id: 'DW-004', organizationElementType: 'DOMAIN', organizationElement: 'test03.example', breachStatus: 'OPEN', employee: false, domain: 'test03.example', email: 'contact@test03.example', username: 'contact', password: 'Contact123!', name: 'External Contact', timestamp: '2026-05-21T00:00:00Z', createdAt: '2026-08-25T17:05:00Z', databaseName: 'sample-forum.invalid' }),
  demoRecord({ id: 'DW-005', organizationElementType: 'DOMAIN', organizationElement: 'test01.example', breachStatus: 'OPEN', employee: true, domain: 'test01.example', email: 'morgan@test01.example', username: 'morgan.k', password: 'WelcomeDemo7', name: 'Morgan Kim', timestamp: '2026-04-02T00:00:00Z', createdAt: '2026-08-22T10:44:00Z' }),
  demoRecord({ id: 'DW-006', organizationElementType: 'EMAIL', organizationElement: 'jordan@test01.example', breachStatus: 'OPEN', employee: true, domain: 'test01.example', email: 'jordan@test01.example', username: 'jordan.p', password: 'DemoPass55', name: 'Jordan Park', timestamp: '2026-03-19T00:00:00Z', createdAt: '2026-08-20T14:31:00Z' }),
  demoRecord({ id: 'DW-007', organizationElementType: 'EMAIL', organizationElement: 'sam@test02.example', breachStatus: 'COMPLETED', employee: true, domain: 'test02.example', email: 'sam@test02.example', username: 'sam.lee', password: 'FakeCredential8', name: 'Sam Lee', timestamp: '2026-02-10T00:00:00Z', createdAt: '2026-08-18T09:00:00Z' }),
  demoRecord({ id: 'DW-008', organizationElementType: 'IP', organizationElement: '192.0.2.17', breachStatus: 'OPEN', employee: true, domain: 'test01.example', email: 'riley@test01.example', username: 'riley.d', password: 'NotARealSecret1', ipAddress: '192.0.2.17', name: 'Riley Diaz', timestamp: '2026-01-11T00:00:00Z', createdAt: '2026-08-16T11:12:00Z' }),
  demoRecord({ id: 'DW-009', organizationElementType: 'USER', organizationElement: 'casey.w', breachStatus: 'OPEN', employee: true, domain: 'test03.example', email: 'casey@test03.example', username: 'casey.w', password: 'Synthetic!22', name: 'Casey Wu', timestamp: '2025-12-03T00:00:00Z', createdAt: '2026-08-14T15:37:00Z' }),
  demoRecord({ id: 'DW-010', organizationElementType: 'PHONE', organizationElement: '+1-202-555-0148', breachStatus: 'OPEN', employee: false, domain: 'test02.example', email: 'info@test02.example', username: 'support', password: 'FakeData_301', phone: '+1-202-555-0148', name: 'Public Account', timestamp: '2025-11-22T00:00:00Z', createdAt: '2026-08-12T18:40:00Z' }),
  demoRecord({ id: 'DW-011', organizationElementType: 'DOMAIN', organizationElement: 'test02.example', breachStatus: 'OPEN', employee: true, domain: 'test02.example', email: 'avery@test02.example', username: 'avery.n', password: 'DemoOnly*76', name: 'Avery Nolan', timestamp: '2025-10-08T00:00:00Z', createdAt: '2026-08-10T07:55:00Z' }),
  demoRecord({ id: 'DW-012', organizationElementType: 'DOMAIN', organizationElement: 'test01.example', breachStatus: 'OPEN', employee: true, domain: 'test01.example', email: 'jamie@test01.example', username: 'jamie.f', password: 'ExampleSecret5', name: 'Jamie Fox', timestamp: '2025-09-17T00:00:00Z', createdAt: '2026-08-08T13:22:00Z' }),
]
