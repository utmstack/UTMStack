import { useTranslation } from 'react-i18next'
import { Section } from '@/features/integrations/components/ui/Section'
import type { CloudGuideSection as CloudGuideSectionType } from './builders/cloudGuideBuilder'

interface CloudGuideSectionProps {
  section: CloudGuideSectionType
}

export function CloudGuideSection({ section }: CloudGuideSectionProps) {
  const { t } = useTranslation()

  return (
    <Section title={t(section.titleKey)}>
      <p className="text-sm text-foreground/90 whitespace-pre-line">
        {t(section.bodyKey)}
      </p>
      {section.image && (
        <img
          src={section.image}
          alt=""
          className="mt-3 max-w-full rounded-md border border-border"
          onError={(e) => {
            e.currentTarget.style.display = 'none'
          }}
        />
      )}
    </Section>
  )
}
