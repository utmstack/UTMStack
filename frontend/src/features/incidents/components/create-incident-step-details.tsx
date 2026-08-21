import { useTranslation } from 'react-i18next'
import { Input } from '@/shared/components/ui/input'

export function CreateIncidentStepDetails({
  name,
  description,
  onChangeName,
  onChangeDescription,
}: {
  name: string
  description: string
  onChangeName: (v: string) => void
  onChangeDescription: (v: string) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="space-y-4">
      <div className="space-y-1.5">
        <label className="text-sm font-medium">
          {t('incidents.create.details.nameLabel')}
          <span className="ml-0.5 text-destructive">*</span>
        </label>
        <Input
          type="text"
          value={name}
          onChange={(e) => onChangeName(e.target.value)}
          placeholder={t('incidents.create.details.namePlaceholder')}
          required
        />
      </div>
      <div className="space-y-1.5">
        <label className="text-sm font-medium">
          {t('incidents.create.details.descriptionLabel')}
        </label>
        <textarea
          value={description}
          onChange={(e) => onChangeDescription(e.target.value)}
          placeholder={t('incidents.create.details.descriptionPlaceholder')}
          rows={4}
          className="flex w-full rounded-md border border-input bg-background/40 px-3 py-2 text-sm placeholder:text-muted-foreground transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:border-ring disabled:cursor-not-allowed disabled:opacity-50 resize-none"
        />
      </div>
    </div>
  )
}
