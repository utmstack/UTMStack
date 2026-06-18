import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import { Filter, Loader2, Pencil, Search, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/components/ui/input'
import { useIntegrations } from '@/features/integrations/hooks/useIntegrations'
import { IntegrationsHeader } from '@/features/integrations/components/IntegrationsHeader'
import { IntegrationCard } from '@/features/integrations/components/IntegrationCard'
import { IntegrationDrawer } from '@/features/integrations/components/IntegrationDrawer'
import { AddCustomIntegrationCard } from '@/features/integrations/components/AddCustomIntegrationCard'
import { CreateIntegrationDrawer } from '@/features/integrations/components/CreateIntegrationDrawer'
import { KIND_META, categoryLabel } from '@/features/integrations/constants'
import { SYSTEM_MODULES } from '@/features/integrations/constants/systemModules'
import type { Integration, DeployKind, CreateModuleRequest } from '@/features/integrations/types'

const LOGO = (slug: string) => {
  if(!slug || slug ==''){
    return ''
  }
  if(slug.startsWith('http')){
    return slug
  }
  return `/integrations/${slug}`
}


function mapModuleToIntegration(module: any, t: TFunction): Integration {
  const categoryLower = module.moduleCategory?.toLowerCase() || ''
  // Map the raw category to a known DeployKind; unknown categories fall back to
  // 'other' (Collector group) so a new module never breaks the grid.
  const kind: DeployKind = (categoryLower in KIND_META ? categoryLower : 'other') as DeployKind

  const isSystem = !!module.isSystem
  // System modules: icon + description are owned by the frontend (per-theme icons +
  // i18n). Custom modules keep using the catalog row's icon/description.
  const meta = isSystem ? SYSTEM_MODULES[module.moduleName] : undefined
  const description = isSystem
    ? t(`integrations.modules.${module.moduleName}`, { defaultValue: module.moduleDescription || '' })
    : module.moduleDescription || ''
  // Resolve the logo + optional dark variant:
  //  - system module        → frontend-owned per-theme asset,
  //  - custom with an icon   → the user-provided icon (no dark variant),
  //  - custom without an icon → the generic custom.svg fallback (+ dark variant).
  let logo: string
  let logoDark: string | undefined
  if (meta) {
    logo = LOGO(meta.icon)
    logoDark = meta.iconDark ? LOGO(meta.iconDark) : undefined
  } else if (module.moduleIcon) {
    logo = LOGO(module.moduleIcon)
  } else {
    logo = LOGO('custom.svg')
    logoDark = LOGO('custom-dark.svg')
  }

  return {
    id: module.id.toString(),
    name: module.prettyName || module.moduleName,
    moduleName: module.moduleName,
    dataType: module.dataType,
    ingestType: module.ingestType,
    kind,
    isSystem,
    status: module.moduleActive ? 'configured' : 'available',
    description,
    category: module.moduleCategory || '',
    logo,
    logoDark,
    darkInvert: false,
    events24h: undefined,
    rate: undefined,
  }
}

// Pinned order for the three OS agents so they always lead the grid.
const AGENT_ORDER = ['WINDOWS_AGENT', 'LINUX_AGENT', 'MACOS']

// Grid sort key (lower comes first): OS agents in their pinned order, then the
// rest of the system integrations, then custom integrations.
function sortRank(i: Integration): number {
  const agentIdx = AGENT_ORDER.indexOf((i.moduleName ?? '').toUpperCase())
  if (agentIdx !== -1) return agentIdx
  return i.isSystem ? AGENT_ORDER.length : AGENT_ORDER.length + 1
}

// Delete affordance overlaid on custom integration cards. Sibling of the card
// button (not nested) to keep valid markup; stops propagation so it never opens
// the drawer.
function CustomDeleteButton({ onConfirm, pending }: { onConfirm: () => void; pending: boolean }) {
  const { t } = useTranslation()
  const [confirm, setConfirm] = useState(false)

  if (confirm) {
    return (
      <div
        className="absolute inset-x-2 top-2 z-10 flex items-center gap-1.5 rounded-md border border-red-500/30 bg-card/95 p-1.5 shadow-md backdrop-blur"
        onClick={(e) => e.stopPropagation()}
      >
        <span className="flex-1 px-1 text-left text-[11px] text-muted-foreground">{t('integrations.delete.confirm')}</span>
        <button
          onClick={(e) => { e.stopPropagation(); onConfirm() }}
          disabled={pending}
          className="rounded bg-red-500 px-2 py-1 text-[11px] font-semibold text-white hover:bg-red-600 disabled:opacity-60"
        >
          {t('integrations.delete.yes')}
        </button>
        <button
          onClick={(e) => { e.stopPropagation(); setConfirm(false) }}
          className="rounded border border-border px-2 py-1 text-[11px] hover:bg-muted"
        >
          {t('integrations.delete.no')}
        </button>
      </div>
    )
  }
  return (
    <button
      onClick={(e) => { e.stopPropagation(); setConfirm(true) }}
      title={t('integrations.delete.title')}
      className="absolute right-2 top-2 z-10 flex h-7 w-7 items-center justify-center rounded-md bg-card/80 text-muted-foreground opacity-0 ring-1 ring-border backdrop-blur transition-opacity hover:bg-red-500/10 hover:text-red-500 group-hover/card:opacity-100"
    >
      <Trash2 size={13} />
    </button>
  )
}

// Edit affordance overlaid on custom integration cards (sibling of the card
// button, left of the delete button). Opens the drawer in edit mode.
function CustomEditButton({ onClick }: { onClick: () => void }) {
  const { t } = useTranslation()
  return (
    <button
      onClick={(e) => { e.stopPropagation(); onClick() }}
      title={t('integrations.edit.title')}
      className="absolute right-11 top-2 z-10 flex h-7 w-7 items-center justify-center rounded-md bg-card/80 text-muted-foreground opacity-0 ring-1 ring-border backdrop-blur transition-opacity hover:bg-primary/10 hover:text-primary group-hover/card:opacity-100"
    >
      <Pencil size={13} />
    </button>
  )
}

export function IntegrationsPage() {
  const { t } = useTranslation()
  const integrations = useIntegrations()
  const [categoryFilter, setCategoryFilter] = useState('')
  const [customOnly, setCustomOnly] = useState(false)
  const [search, setSearch] = useState('')
  const [showOnlyConfigured, setShowOnlyConfigured] = useState(false)
  const [open, setOpen] = useState<Integration | null>(null)
  const [createDrawerOpen, setCreateDrawerOpen] = useState(false)
  const [editing, setEditing] = useState<Integration | null>(null)

  const modules = integrations.modules.data || []
  const displayList = useMemo(() => modules.map((m) => mapModuleToIntegration(m, t)), [modules, t])

  // The raw catalog row for the module being edited (for its actual stored icon).
  const editingModule = editing ? modules.find((m) => String(m.id) === editing.id) : undefined
  // Identifiers already taken (excluding the edited module), to reject name collisions.
  const takenModuleNames = modules
    .filter((m) => String(m.id) !== editing?.id)
    .map((m) => m.moduleName.toLowerCase())
  const takenDataTypes = modules
    .filter((m) => String(m.id) !== editing?.id)
    .map((m) => (m.dataType ?? '').toLowerCase())
    .filter(Boolean)
  // Distinct categories already in use (the create/edit selector's options).
  const categories = Array.from(
    new Set(modules.map((m) => m.moduleCategory).filter((c): c is string => !!c))
  ).sort()

  const filtered = useMemo(() => {
    return displayList
      .filter((i) => (categoryFilter ? i.category === categoryFilter : true))
      .filter((i) => (customOnly ? !i.isSystem : true))
      .filter((i) => (showOnlyConfigured ? i.status === 'configured' : true))
      .filter((i) =>
        search
          ? (i.name + i.description + i.category).toLowerCase().includes(search.toLowerCase())
          : true
      )
      // Ordering: the three OS agents (Linux, Windows, macOS) lead the grid, then
      // the rest of the system integrations, then custom ones (just before the
      // "+ Create" card).
      .sort((a, b) => sortRank(a) - sortRank(b))
  }, [categoryFilter, customOnly, search, showOnlyConfigured, displayList])

  const configuredCount = displayList.filter((i) => i.status === 'configured').length

  // The drawer reads the live mapped integration (by id) so its status reflects
  // enable/disable toggles immediately after the modules query refetches.
  const openLive = open ? displayList.find((d) => d.id === open.id) ?? open : null

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      {/* Loading overlay */}
      {integrations.isLoading && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-card/60 backdrop-blur-sm">
          <div className="flex items-center gap-3 rounded-lg border border-border bg-card px-4 py-3">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span className="text-sm text-muted-foreground">{t('integrations.loading')}</span>
          </div>
        </div>
      )}

      <IntegrationsHeader configured={configuredCount} total={displayList.length} />

      <div className="mt-5 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[300px] flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('integrations.searchPlaceholder')}
            className="h-10 pl-9"
          />
        </div>
        <select
          value={categoryFilter}
          onChange={(e) => setCategoryFilter(e.target.value)}
          className="h-10 rounded-md border border-border bg-background px-3 text-sm text-foreground transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        >
          <option value="">{t('integrations.allCategories')}</option>
          {categories.map((c) => (
            <option key={c} value={c}>
              {categoryLabel(c)}
            </option>
          ))}
        </select>
        <button
          onClick={() => setCustomOnly((v) => !v)}
          className={cn(
            'flex h-10 items-center gap-2 rounded-md border px-3 text-sm transition-colors',
            customOnly
              ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
              : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
          )}
        >
          {t('integrations.customOnly')}
        </button>
        <button
          onClick={() => setShowOnlyConfigured((v) => !v)}
          className={cn(
            'flex h-10 items-center gap-2 rounded-md border px-3 text-sm transition-colors',
            showOnlyConfigured
              ? 'border-primary/30 bg-primary/10 text-primary'
              : 'border-border text-muted-foreground hover:bg-muted hover:text-foreground'
          )}
        >
          <Filter size={14} />
          {t('integrations.configuredOnly')}
        </button>
      </div>

      <div className="mt-5 grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
        {filtered.map((i) => (
          <div key={i.id} className="group/card relative">
            <IntegrationCard integration={i} onOpen={() => setOpen(i)} />
            {!i.isSystem && (
              <CustomEditButton onClick={() => setEditing(i)} />
            )}
            {!i.isSystem && (
              <CustomDeleteButton
                pending={integrations.deleteModule.isPending}
                onConfirm={() =>
                  integrations.deleteModule.mutate(Number(i.id), {
                    onSuccess: () => toast.success(t('integrations.delete.deleted')),
                    onError: () => toast.error(t('integrations.delete.error')),
                  })
                }
              />
            )}
          </div>
        ))}
        <AddCustomIntegrationCard onClick={() => setCreateDrawerOpen(true)} />
        {filtered.length === 0 && (
          <div className="col-span-full rounded-xl border border-dashed border-border bg-card px-6 py-16 text-center text-sm text-muted-foreground">
            {t('integrations.noMatch')}
          </div>
        )}
      </div>

      {openLive && (
        <IntegrationDrawer
          integration={openLive}
          onClose={() => setOpen(null)}
          toggling={integrations.activateDeactivateModule.isPending}
          onToggleActive={(activate) => {
            if (!openLive.moduleName) return
            integrations.activateDeactivateModule.mutate(
              { moduleName: openLive.moduleName, activationStatus: activate },
              {
                onSuccess: () =>
                  toast.success(
                    activate ? t('integrations.toast.enabled') : t('integrations.toast.disabled')
                  ),
                onError: () => toast.error(t('integrations.toast.toggleError')),
              }
            )
          }}
        />
      )}

      <CreateIntegrationDrawer
        open={createDrawerOpen || !!editing}
        editing={
          editing
            ? {
                id: editing.id,
                prettyName: editing.name,
                description: editing.description,
                category: editing.category,
                currentIcon: editingModule?.moduleIcon || undefined,
              }
            : null
        }
        onClose={() => {
          setCreateDrawerOpen(false)
          setEditing(null)
        }}
        takenModuleNames={takenModuleNames}
        takenDataTypes={takenDataTypes}
        categories={categories}
        isSubmitting={editing ? integrations.updateModule.isPending : integrations.createModule.isPending}
        onSubmit={async (data: CreateModuleRequest) => {
          if (editing) {
            await integrations.updateModule.mutateAsync({
              id: Number(editing.id),
              data: {
                prettyName: data.prettyName,
                moduleDescription: data.moduleDescription,
                moduleIcon: data.moduleIcon,
                moduleCategory: data.moduleCategory,
              },
            })
            setEditing(null)
            toast.success(t('integrations.toast.updated'))
          } else {
            await integrations.createModule.mutateAsync(data)
            setCreateDrawerOpen(false)
            toast.success(t('integrations.toast.created'))
          }
        }}
      />
    </div>
  )
}
