import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { X } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import type { CreateModuleRequest, DataTypeOption } from '@/features/integrations/types'

interface CreateIntegrationDrawerProps {
  open: boolean
  onClose: () => void
  dataTypes: DataTypeOption[]
  onSubmit: (data: CreateModuleRequest) => Promise<void>
  isSubmitting: boolean
}

const MAX_FILE_SIZE = 3 * 1024 * 1024 // 3MB
const ALLOWED_IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/svg+xml', 'image/gif', 'image/webp']

export function CreateIntegrationDrawer({
  open,
  onClose,
  dataTypes,
  onSubmit,
  isSubmitting,
}: CreateIntegrationDrawerProps) {
  const { t } = useTranslation()
  const [moduleName, setModuleName] = useState('')
  const [prettyName, setPrettyName] = useState('')
  const [dataType, setDataType] = useState('')
  const [iconDataUrl, setIconDataUrl] = useState<string | null>(null)
  const [iconError, setIconError] = useState<string | null>(null)

  const existingDataTypes = dataTypes.map((dt) => dt.dataType.toLowerCase())
  const isDataTypeValid = dataType.trim() && !existingDataTypes.includes(dataType.toLowerCase())
  const isValid = moduleName.trim() && isDataTypeValid

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    setIconError(null)

    if (!file) return

    // Check file type
    if (!ALLOWED_IMAGE_TYPES.includes(file.type)) {
      setIconError(t('integrations.createDrawer.invalidFileType'))
      setIconDataUrl(null)
      return
    }

    // Check file size
    if (file.size > MAX_FILE_SIZE) {
      setIconError(t('integrations.createDrawer.fileTooLarge'))
      setIconDataUrl(null)
      return
    }

    const reader = new FileReader()
    reader.onload = (event) => {
      setIconDataUrl(event.target?.result as string)
      setIconError(null)
    }
    reader.onerror = () => {
      setIconError(t('integrations.createDrawer.fileReadError'))
      setIconDataUrl(null)
    }
    reader.readAsDataURL(file)
  }

  const handleSubmit = async () => {
    if (!isValid || isSubmitting) return

    try {
      await onSubmit({
        moduleName,
        dataType,
        prettyName: prettyName || undefined,
        moduleCategory: 'custom',
        moduleIcon: iconDataUrl || undefined,
      })
    } catch {
      // Error is handled by parent
    }
  }

  if (!open) return null

  return (
    <div
      className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-[820px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-5">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <h2 className="text-xl font-semibold">{t('integrations.createDrawer.title')}</h2>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>
        </header>

        <div className="flex-1 overflow-y-auto bg-muted/20 p-6">
          <form className="space-y-4" onSubmit={(e) => { e.preventDefault(); handleSubmit() }}>
            {/* Module Name */}
            <div>
              <label className="mb-1 block text-xs uppercase tracking-wider text-muted-foreground">
                {t('integrations.createDrawer.moduleName')}
              </label>
              <Input
                type="text"
                placeholder="my-custom-app"
                value={moduleName}
                onChange={(e) => setModuleName(e.target.value)}
                disabled={isSubmitting}
                className="h-9"
              />
              <p className="mt-1 text-[11px] text-muted-foreground">
                {t('integrations.createDrawer.moduleNameHint')}
              </p>
            </div>

            {/* Pretty Name */}
            <div>
              <label className="mb-1 block text-xs uppercase tracking-wider text-muted-foreground">
                {t('integrations.createDrawer.prettyName')}
              </label>
              <Input
                type="text"
                placeholder="My Custom App"
                value={prettyName}
                onChange={(e) => setPrettyName(e.target.value)}
                disabled={isSubmitting}
                className="h-9"
              />
            </div>

            {/* Data Type */}
            <div>
              <label className="mb-1 block text-xs uppercase tracking-wider text-muted-foreground">
                {t('integrations.createDrawer.dataType')}
              </label>
              <Input
                type="text"
                placeholder="custom-app-type"
                value={dataType}
                onChange={(e) => setDataType(e.target.value)}
                disabled={isSubmitting}
                className="h-9"
              />
              {dataType.trim() && !isDataTypeValid && (
                <p className="mt-1 text-[11px] text-red-500">
                  {t('integrations.createDrawer.dataTypeExists')}
                </p>
              )}
              {!dataType.trim() && (
                <p className="mt-1 text-[11px] text-muted-foreground">
                  {t('integrations.createDrawer.dataTypePlaceholder')}
                </p>
              )}
            </div>

            {/* Icon */}
            <div>
              <label className="mb-1 block text-xs uppercase tracking-wider text-muted-foreground">
                {t('integrations.createDrawer.icon')}
              </label>
              <input
                type="file"
                accept="image/*"
                onChange={handleFileUpload}
                disabled={isSubmitting}
                className="block w-full text-sm text-muted-foreground file:me-3 file:rounded-md file:border-0 file:bg-primary/10 file:px-3 file:py-2 file:text-sm file:font-medium file:text-primary hover:file:bg-primary/15"
              />
              <p className="mt-1 text-[11px] text-muted-foreground">
                {t('integrations.createDrawer.iconHint')}
              </p>
              {iconError && (
                <p className="mt-1 text-[11px] text-red-500">{iconError}</p>
              )}
              {iconDataUrl && (
                <div className="mt-3 flex items-center gap-3">
                  <img
                    src={iconDataUrl}
                    alt="icon preview"
                    className="h-12 w-12 rounded border border-border object-contain"
                  />
                  <span className="text-[11px] text-muted-foreground">
                    {t('integrations.createDrawer.iconPreview')}
                  </span>
                </div>
              )}
            </div>

            {/* Submit Buttons */}
            <div className="mt-6 flex items-center gap-2 border-t border-border pt-4">
              <Button
                type="submit"
                disabled={!isValid || isSubmitting}
                size="sm"
              >
                {isSubmitting ? t('integrations.createDrawer.creating') : t('integrations.createDrawer.create')}
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={onClose}
                disabled={isSubmitting}
              >
                {t('integrations.createDrawer.cancel')}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
