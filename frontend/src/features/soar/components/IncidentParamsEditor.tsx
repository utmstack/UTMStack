import { useTranslation } from 'react-i18next'
import { Input } from '@/shared/components/ui/input'

interface IncidentParams {
  name?: string
  description?: string
}

interface Props {
  params: unknown
  readOnly?: boolean
  onChange: (params: IncidentParams) => void
}

// ponytail: name + description only — alert identity comes from the exec's
// AlertID and the interpolation bag on the backend.
export function IncidentParamsEditor({ params, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const p = normalize(params)

  return (
    <div className="space-y-2">
      <div>
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('soar.editor.canvas.incident.name')}
        </label>
        <Input
          value={p.name ?? ''}
          readOnly={readOnly}
          onChange={(e) => onChange({ ...p, name: e.target.value })}
          placeholder={t('soar.editor.canvas.incident.namePlaceholder')}
          className="h-8 text-xs"
        />
      </div>
      <div>
        <label className="mb-1 block text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
          {t('soar.editor.canvas.incident.description')}
        </label>
        <textarea
          value={p.description ?? ''}
          readOnly={readOnly}
          onChange={(e) => onChange({ ...p, description: e.target.value })}
          rows={4}
          placeholder={t('soar.editor.canvas.incident.descriptionPlaceholder')}
          className="w-full rounded-md border border-input bg-background px-2 py-1.5 text-[11px] focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        />
      </div>
    </div>
  )
}

function normalize(params: unknown): IncidentParams {
  if (!params || typeof params !== 'object') return {}
  return params as IncidentParams
}
