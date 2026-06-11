import type { ReactNode } from 'react'
import type { Integration } from '@/features/integrations/types'

export interface CollectorSection {
  id: string
  titleKey: string
  bodyKey: string
  image?: string
}

export interface Collector {
  getName(): string
  matches?(name: string): boolean
  sections: CollectorSection[]
  render(module: Integration): ReactNode
}
