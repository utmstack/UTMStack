import { useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Upload, FileCheck2 } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useIntegrations } from '@/features/integrations/hooks/useIntegrations'
import type { TenantResponse } from '@/features/integrations/types'
import type { CloudConfigField } from './builders/cloudGuideBuilder'

interface CloudTenantFormProps {
  moduleName: string
  fields: CloudConfigField[]
  editing?: TenantResponse | null
  onCancel: () => void
  onSaved: () => void
}

function emptyConfig(fields: CloudConfigField[]): Record<string, string> {
  return fields.reduce<Record<string, string>>((acc, f) => {
    acc[f.key] = ''
    return acc
  }, {})
}

export function CloudTenantForm({
  moduleName,
  fields,
  editing,
  onCancel,
  onSaved,
}: CloudTenantFormProps) {
  const { t } = useTranslation()
  const { saveTenant } = useIntegrations()

  const initialConfig = useMemo(() => {
    const base = emptyConfig(fields)
    if (editing) {
      for (const key of Object.keys(base)) {
        base[key] = editing.config[key] ?? ''
      }
    }
    return base
  }, [editing, fields])

  const [name, setName] = useState(editing?.name ?? '')
  const [config, setConfig] = useState<Record<string, string>>(initialConfig)

  useEffect(() => {
    setName(editing?.name ?? '')
    setConfig(initialConfig)
  }, [editing, initialConfig])

  const isEditing = !!editing

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    saveTenant.mutate(
      { moduleName, data: { name: name.trim(), config } },
      {
        onSuccess: () => {
          setName('')
          setConfig(emptyConfig(fields))
          onSaved()
        },
        onError: (err) => {
          const fallback = isEditing
            ? t('integrations.setup.cloud.tenants.updateError')
            : t('integrations.setup.cloud.tenants.createError')
          toast.error(err instanceof Error && err.message ? err.message : fallback)
        },
      },
    )
  }

  return (
    <form className="space-y-3" onSubmit={handleSubmit}>
      {isEditing && (
        <p className="text-xs text-muted-foreground">
          {t('integrations.setup.cloud.tenants.editing', { name: editing!.name })}
        </p>
      )}

      <div>
        <label className="mb-1 block text-[11px] uppercase tracking-wider text-muted-foreground">
          {t('integrations.setup.cloud.tenants.nameLabel')}
        </label>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder={t('integrations.setup.cloud.tenants.namePlaceholder')}
          disabled={isEditing}
          className="h-9 text-sm"
          required
        />
      </div>

      {fields.map((field) => (
        <div key={field.key}>
          <label className="mb-1 block text-[11px] uppercase tracking-wider text-muted-foreground">
            {t(field.labelKey)}
          </label>
          {field.type === 'file' ? (
            <JsonFileField
              field={field}
              value={config[field.key] ?? ''}
              onChange={(value) => setConfig({ ...config, [field.key]: value })}
            />
          ) : field.type === 'textarea' ? (
            <textarea
              value={config[field.key] ?? ''}
              onChange={(e) => setConfig({ ...config, [field.key]: e.target.value })}
              placeholder={field.placeholder}
              rows={5}
              className="w-full rounded-md border border-border bg-background px-3 py-2 font-mono text-xs"
            />
          ) : (
            <Input
              type={field.secret ? 'password' : 'text'}
              value={config[field.key] ?? ''}
              onChange={(e) => setConfig({ ...config, [field.key]: e.target.value })}
              placeholder={field.placeholder}
              className="h-9 font-mono text-xs"
            />
          )}
        </div>
      ))}

      <div className="flex items-center gap-2 pt-1">
        <Button size="sm" type="submit" disabled={saveTenant.isPending}>
          {isEditing
            ? t('integrations.setup.cloud.tenants.update')
            : t('integrations.setup.cloud.tenants.add')}
        </Button>
        {isEditing && (
          <Button size="sm" type="button" variant="outline" onClick={onCancel}>
            {t('integrations.setup.cloud.tenants.cancel')}
          </Button>
        )}
      </div>
    </form>
  )
}

interface JsonFileFieldProps {
  field: CloudConfigField
  value: string
  onChange: (value: string) => void
}

function JsonFileField({ field, value, onChange }: JsonFileFieldProps) {
  const { t } = useTranslation()
  const inputRef = useRef<HTMLInputElement>(null)
  const [fileName, setFileName] = useState<string>('')
  const [error, setError] = useState<string>('')

  const hasValue = value.trim().length > 0

  const handlePick = () => inputRef.current?.click()

  const handleFile = async (file: File) => {
    setError('')
    try {
      const text = await file.text()
      const parsed = JSON.parse(text)
      onChange(JSON.stringify(parsed))
      setFileName(file.name)
    } catch {
      setError(t('integrations.setup.cloud.tenants.invalidJson'))
      onChange('')
      setFileName('')
    }
  }

  return (
    <div className="space-y-2">
      <input
        ref={inputRef}
        type="file"
        accept={field.accept ?? 'application/json,.json'}
        className="hidden"
        onChange={(e) => {
          const file = e.target.files?.[0]
          if (file) handleFile(file)
          e.target.value = ''
        }}
      />
      <div className="flex items-center gap-2">
        <Button size="sm" type="button" variant="outline" onClick={handlePick}>
          <Upload size={13} className="mr-1.5" />
          {hasValue
            ? t('integrations.setup.cloud.tenants.replaceFile')
            : t('integrations.setup.cloud.tenants.selectFile')}
        </Button>
        {hasValue && (
          <span className="flex items-center gap-1 text-xs text-muted-foreground">
            <FileCheck2 size={13} className="text-emerald-500" />
            {fileName || t('integrations.setup.cloud.tenants.fileLoaded')}
          </span>
        )}
      </div>
      {field.placeholder && !hasValue && (
        <p className="text-[11px] text-muted-foreground">{field.placeholder}</p>
      )}
      {error && <p className="text-[11px] text-destructive">{error}</p>}
    </div>
  )
}
