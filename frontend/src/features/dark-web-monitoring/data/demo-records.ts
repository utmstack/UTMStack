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
    compromisedData: 'email addresses',
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

/**
 * Synthetic demonstration records. No field carries a real-looking personal
 * name, phone number, or credential: identifiers are role-based
 * (`user01`, `contact`, `support`) and `password`/`hashedPassword` stay null,
 * so the UI and exports never ship PII-like or secret-like values.
 */
export const DEMO_RECORDS: DarkWebRecord[] = [
  demoRecord({ id: 'DW-001', organizationElementType: 'DOMAIN', organizationElement: 'test01.example', breachStatus: 'IN_MITIGATION', employee: true, domain: 'test01.example', email: 'user01@test01.example', username: 'user01', name: 'Employee 01', timestamp: '2026-08-07T00:00:00Z', createdAt: '2026-08-28T12:20:00Z', databaseName: 'example-forum.invalid' }),
  demoRecord({ id: 'DW-002', organizationElementType: 'DOMAIN', organizationElement: 'test01.example', breachStatus: 'IN_MITIGATION', employee: true, domain: 'test01.example', email: 'user02@test01.example', username: 'user02', name: 'Employee 02', timestamp: '2026-08-07T00:00:00Z', createdAt: '2026-08-28T12:20:00Z', databaseName: 'archive.invalid' }),
  demoRecord({ id: 'DW-003', organizationElementType: 'DOMAIN', organizationElement: 'test02.example', breachStatus: 'OPEN', employee: true, domain: 'test02.example', email: 'user03@test02.example', username: 'user03', name: 'Employee 03', timestamp: '2026-06-14T00:00:00Z', createdAt: '2026-08-26T08:14:00Z', databaseName: 'paste-archive.invalid' }),
  demoRecord({ id: 'DW-004', organizationElementType: 'DOMAIN', organizationElement: 'test03.example', breachStatus: 'OPEN', employee: false, domain: 'test03.example', email: 'contact@test03.example', username: 'contact', name: 'External Contact', timestamp: '2026-05-21T00:00:00Z', createdAt: '2026-08-25T17:05:00Z', databaseName: 'sample-forum.invalid' }),
  demoRecord({ id: 'DW-005', organizationElementType: 'DOMAIN', organizationElement: 'test01.example', breachStatus: 'OPEN', employee: true, domain: 'test01.example', email: 'user05@test01.example', username: 'user05', name: 'Employee 05', timestamp: '2026-04-02T00:00:00Z', createdAt: '2026-08-22T10:44:00Z' }),
  demoRecord({ id: 'DW-006', organizationElementType: 'EMAIL', organizationElement: 'user06@test01.example', breachStatus: 'OPEN', employee: true, domain: 'test01.example', email: 'user06@test01.example', username: 'user06', name: 'Employee 06', timestamp: '2026-03-19T00:00:00Z', createdAt: '2026-08-20T14:31:00Z' }),
  demoRecord({ id: 'DW-007', organizationElementType: 'EMAIL', organizationElement: 'user07@test02.example', breachStatus: 'COMPLETED', employee: true, domain: 'test02.example', email: 'user07@test02.example', username: 'user07', name: 'Employee 07', timestamp: '2026-02-10T00:00:00Z', createdAt: '2026-08-18T09:00:00Z' }),
  demoRecord({ id: 'DW-008', organizationElementType: 'IP', organizationElement: '192.0.2.17', breachStatus: 'OPEN', employee: true, domain: 'test01.example', email: 'user08@test01.example', username: 'user08', ipAddress: '192.0.2.17', name: 'Employee 08', timestamp: '2026-01-11T00:00:00Z', createdAt: '2026-08-16T11:12:00Z' }),
  demoRecord({ id: 'DW-009', organizationElementType: 'USER', organizationElement: 'user09', breachStatus: 'OPEN', employee: true, domain: 'test03.example', email: 'user09@test03.example', username: 'user09', name: 'Employee 09', timestamp: '2025-12-03T00:00:00Z', createdAt: '2026-08-14T15:37:00Z' }),
  demoRecord({ id: 'DW-010', organizationElementType: 'PHONE', organizationElement: '+1-202-555-0148', breachStatus: 'OPEN', employee: false, domain: 'test02.example', email: 'info@test02.example', username: 'support', phone: '+1-202-555-0148', name: 'Public Account', timestamp: '2025-11-22T00:00:00Z', createdAt: '2026-08-12T18:40:00Z' }),
  demoRecord({ id: 'DW-011', organizationElementType: 'DOMAIN', organizationElement: 'test02.example', breachStatus: 'OPEN', employee: true, domain: 'test02.example', email: 'user11@test02.example', username: 'user11', name: 'Employee 11', timestamp: '2025-10-08T00:00:00Z', createdAt: '2026-08-10T07:55:00Z' }),
  demoRecord({ id: 'DW-012', organizationElementType: 'DOMAIN', organizationElement: 'test01.example', breachStatus: 'OPEN', employee: true, domain: 'test01.example', email: 'user12@test01.example', username: 'user12', name: 'Employee 12', timestamp: '2025-09-17T00:00:00Z', createdAt: '2026-08-08T13:22:00Z' }),
]
