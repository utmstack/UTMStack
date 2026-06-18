import type { TenantResponse } from '@/features/integrations/types'
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

  // Azure and GCP are rendered by dedicated guide components (AzureGuide / GcpGuide)
  // with a richer flow diagram + step-by-step doc, so they never fall through to the
  // generic builder here.

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
