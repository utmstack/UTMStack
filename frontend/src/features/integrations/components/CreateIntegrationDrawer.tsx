import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { X, Trash2 } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useThemeContext } from '@/app/providers/ThemeProvider'
import { categoryLabel } from '@/features/integrations/constants'
import type { CreateModuleRequest } from '@/features/integrations/types'

// When set, the drawer runs in edit mode (updates prettyName/description/icon/
// category). moduleName + dataType are identity keys and are NOT recomputed.
export interface EditTarget {
  id: string
  prettyName: string
  description: string
  category: string
  currentIcon?: string
}

interface CreateIntegrationDrawerProps {
  open: boolean
  onClose: () => void
  onSubmit: (data: CreateModuleRequest) => Promise<void>
  isSubmitting: boolean
  editing?: EditTarget | null
  // Existing identifiers (lowercased) used to reject a name whose derived
  // moduleName/dataType would collide. Excludes the module being edited.
  takenModuleNames: string[]
  takenDataTypes: string[]
  // Distinct module categories already in use (the selector's options).
  categories: string[]
}

const MAX_FILE_SIZE = 3 * 1024 * 1024 // 3MB
const ALLOWED_IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/svg+xml', 'image/gif', 'image/webp']

const fieldCls =
  'flex w-full rounded-md border border-border bg-background px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50'

// Derive a URL-safe slug from the integration name.
function slugify(s: string): string {
  return s
    .trim()
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

export function CreateIntegrationDrawer({
  open,
  onClose,
  onSubmit,
  isSubmitting,
  editing,
  takenModuleNames,
  takenDataTypes,
  categories,
}: CreateIntegrationDrawerProps) {
  const { t } = useTranslation()
  const { theme } = useThemeContext()
  const isEdit = !!editing

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [category, setCategory] = useState('')
  const [iconDataUrl, setIconDataUrl] = useState<string | null>(null)
  const [iconRemoved, setIconRemoved] = useState(false)
  const [iconError, setIconError] = useState<string | null>(null)

  // Seed the form when the drawer opens (prefilled for edit, empty for create).
  useEffect(() => {
    if (!open) return
    setName(editing?.prettyName ?? '')
    setDescription(editing?.description ?? '')
    setCategory(editing?.category ?? '')
    setIconDataUrl(null)
    setIconRemoved(false)
    setIconError(null)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, editing?.id])

  // moduleName + dataType are derived from the name on create.
  const slug = slugify(name)
  const computedModuleName = slug ? slug.toUpperCase().replace(/-/g, '_') : ''
  const nameTaken =
    !isEdit &&
    !!slug &&
    (takenDataTypes.includes(slug) || takenModuleNames.includes(computedModuleName.toLowerCase()))

  const isValid = isEdit
    ? !!name.trim()
    : !!name.trim() && !!slug && !!category && !nameTaken

  // Icon preview resolution: uploaded > existing (edit, not removed) > default.
  const defaultIcon = theme === 'dark' ? '/integrations/custom-dark.svg' : '/integrations/custom.svg'
  const keepsExisting = isEdit && !!editing?.currentIcon && !iconRemoved && !iconDataUrl
  const previewIcon = iconDataUrl ?? (keepsExisting ? editing!.currentIcon! : defaultIcon)
  const showingDefault = !iconDataUrl && !keepsExisting

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    setIconError(null)
    if (!file) return
    if (!ALLOWED_IMAGE_TYPES.includes(file.type)) {
      setIconError(t('integrations.createDrawer.invalidFileType'))
      return
    }
    if (file.size > MAX_FILE_SIZE) {
      setIconError(t('integrations.createDrawer.fileTooLarge'))
      return
    }
    const reader = new FileReader()
    reader.onload = (event) => {
      setIconDataUrl(event.target?.result as string)
      setIconRemoved(false)
      setIconError(null)
    }
    reader.onerror = () => setIconError(t('integrations.createDrawer.fileReadError'))
    reader.readAsDataURL(file)
  }

  const removeIcon = () => {
    setIconDataUrl(null)
    setIconRemoved(true)
    setIconError(null)
  }

  const handleSubmit = async () => {
    if (!isValid || isSubmitting) return
    // undefined → no change / none; '' → explicitly cleared; dataUrl → new icon.
    const submitIcon = iconDataUrl ? iconDataUrl : iconRemoved ? '' : undefined
    try {
      await onSubmit({
        moduleName: computedModuleName,
        dataType: slug,
        prettyName: name.trim(),
        moduleDescription: description.trim() || undefined,
        moduleCategory: category,
        moduleIcon: submitIcon,
      })
    } catch {
      // Error handled by parent.
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
            <h2 className="text-xl font-semibold">
              {isEdit ? t('integrations.createDrawer.editTitle') : t('integrations.createDrawer.title')}
            </h2>
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
            {/* Integration Name → pretty name (+ derived identifier) */}
            <div>
              <label className="mb-1 block text-xs uppercase tracking-wider text-muted-foreground">
                {t('integrations.createDrawer.name')}
              </label>
              <Input
                type="text"
                placeholder={t('integrations.createDrawer.namePlaceholder')}
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isSubmitting}
                className="h-9"
              />
              {!isEdit && slug && !nameTaken && (
                <p className="mt-1 text-[11px] text-muted-foreground">
                  {t('integrations.createDrawer.identifierHint', { id: slug })}
                </p>
              )}
              {!isEdit && nameTaken && (
                <p className="mt-1 text-[11px] text-red-500">
                  {t('integrations.createDrawer.nameTaken')}
                </p>
              )}
            </div>

            {/* Description */}
            <div>
              <label className="mb-1 block text-xs uppercase tracking-wider text-muted-foreground">
                {t('integrations.createDrawer.description')}
              </label>
              <textarea
                rows={3}
                placeholder={t('integrations.createDrawer.descriptionPlaceholder')}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={isSubmitting}
                className={fieldCls}
              />
            </div>

            {/* Category */}
            <div>
              <label className="mb-1 block text-xs uppercase tracking-wider text-muted-foreground">
                {t('integrations.createDrawer.category')}
              </label>
              <select
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                disabled={isSubmitting}
                className={`${fieldCls} h-9 py-0`}
              >
                <option value="" disabled>
                  {t('integrations.createDrawer.categoryPlaceholder')}
                </option>
                {categories.map((c) => (
                  <option key={c} value={c}>
                    {categoryLabel(c)}
                  </option>
                ))}
              </select>
            </div>

            {/* Icon — default shown until the user uploads one; removable */}
            <div>
              <label className="mb-1 block text-xs uppercase tracking-wider text-muted-foreground">
                {t('integrations.createDrawer.icon')}
              </label>
              <div className="flex items-center gap-4">
                <img
                  src={previewIcon}
                  alt="icon preview"
                  className="h-14 w-14 shrink-0 rounded-md border border-border bg-muted/30 object-contain p-1"
                />
                <div className="min-w-0 flex-1">
                  <input
                    type="file"
                    accept="image/*"
                    onChange={handleFileUpload}
                    disabled={isSubmitting}
                    className="block w-full text-sm text-muted-foreground file:me-3 file:rounded-md file:border-0 file:bg-primary/10 file:px-3 file:py-2 file:text-sm file:font-medium file:text-primary hover:file:bg-primary/15"
                  />
                  <p className="mt-1 text-[11px] text-muted-foreground">
                    {showingDefault
                      ? t('integrations.createDrawer.defaultIconNote')
                      : t('integrations.createDrawer.iconHint')}
                  </p>
                </div>
                {!showingDefault && (
                  <button
                    type="button"
                    onClick={removeIcon}
                    disabled={isSubmitting}
                    title={t('integrations.createDrawer.removeIcon')}
                    className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground ring-1 ring-border hover:bg-red-500/10 hover:text-red-500 disabled:opacity-50"
                  >
                    <Trash2 size={14} />
                  </button>
                )}
              </div>
              {iconError && <p className="mt-1 text-[11px] text-red-500">{iconError}</p>}
            </div>

            {/* Actions */}
            <div className="mt-6 flex items-center gap-2 border-t border-border pt-4">
              <Button type="submit" disabled={!isValid || isSubmitting} size="sm">
                {isEdit
                  ? isSubmitting ? t('integrations.createDrawer.saving') : t('integrations.createDrawer.save')
                  : isSubmitting ? t('integrations.createDrawer.creating') : t('integrations.createDrawer.create')}
              </Button>
              <Button type="button" variant="outline" size="sm" onClick={onClose} disabled={isSubmitting}>
                {t('integrations.createDrawer.cancel')}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
