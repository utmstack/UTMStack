import type { TenantResponse } from '@/features/integrations/types'
import { GOOGLE_SECTIONS, GOOGLE_FIELDS } from './google'
import { AZURE_SECTIONS, AZURE_FIELDS } from './azure'
import { AWS_SECTIONS, AWS_FIELDS } from './aws'

export interface CloudGuideSection {
  id: string
  titleKey: string
  bodyKey: string
  image?: string
}

export interface CloudConfigField {
  key: string
  labelKey: string
  placeholder?: string
  secret?: boolean
  type?: 'text' | 'password' | 'textarea' | 'file'
  accept?: string
}

export interface CloudGuideContext {
  tenants: TenantResponse[]
}

export interface CloudGuideConfig {
  sections: CloudGuideSection[]
  fields: CloudConfigField[]
  tenants: TenantResponse[]
}

export function buildCloudGuide(name: string, ctx: CloudGuideContext): CloudGuideConfig {
  const normalized = (name ?? '').toLowerCase()

  if (normalized.includes('google') || normalized.includes('gcp') || normalized.includes('pubsub')) {
    return {
      sections: GOOGLE_SECTIONS,
      fields: GOOGLE_FIELDS,
      tenants: ctx.tenants,
    }
  }

  if (normalized.includes('azure')) {
    return {
      sections: AZURE_SECTIONS,
      fields: AZURE_FIELDS,
      tenants: ctx.tenants,
    }
  }

  if (
    normalized.includes('aws') ||
    normalized.includes('cloudtrail') ||
    normalized.includes('cloudwatch')
  ) {
    return {
      sections: AWS_SECTIONS,
      fields: AWS_FIELDS,
      tenants: ctx.tenants,
    }
  }

  return {
    sections: [],
    fields: [],
    tenants: ctx.tenants,
  }
}
