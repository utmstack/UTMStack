export type OrganizationElementType = 'DOMAIN' | 'EMAIL' | 'IP' | 'USER' | 'PHONE'
export type BreachStatus = 'OPEN' | 'IN_MITIGATION' | 'COMPLETED'

export interface DarkWebRecord {
  id: string
  timestamp: string
  createdAt: string
  domain: string | null
  email: string | null
  ipAddress: string | null
  username: string | null
  password: string | null
  hashedPassword: string | null
  name: string | null
  vin: string | null
  address: string | null
  phone: string | null
  databaseName: string | null
  foundIn: string | null
  source: string | null
  leakName: string | null
  breachDescription: string | null
  compromisedData: string | null
  imageUrl: string | null
  victimDomain: string | null
  distributionType: string | null
  breachId: string | null
  organizationId: number
  organizationUUID: string
  organizationElement: string
  organizationElementType: OrganizationElementType
  breachStatus: BreachStatus
  employee: boolean
  echoesCount: number
}

export interface OrganizationBreachStats {
  id: number
  orgName: string
  totalEmailsBreaches: number
  totalPhonesBreaches: number
  totalDomainBreaches: number
  totalIpsBreaches: number
  totalUsernamesBreaches: number
  totalBreaches: number
}

export type LeakSurface = 'domains' | 'emails' | 'ips' | 'usernames' | 'phones'
