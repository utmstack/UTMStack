import { useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { BarChart3, MoreVertical, Pencil, Plus, Star, Trash2 } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import type { Dashboard } from '@/features/dashboard/types'

/**
 * Dashboards-as-tabs: one tab per dashboard (consumption-first, like the Log
 * Explorer tabs). The view stays clean — tabs + time range + a single options
 * (⋯) menu that holds every management action for the current dashboard plus the
 * global ones (new dashboard, visualizations library).
 */
export function DashboardTabsBar({
  dashboards,
  selectedId,
  defaultId,
  onSelect,
  onSetDefault,
  onCreate,
  onRename,
  onDelete,
  onEdit,
  right,
}: {
  dashboards: Dashboard[]
  selectedId: number | null
  defaultId: number | null
  onSelect: (id: number) => void
  onSetDefault: (id: number | null) => void
  onCreate: () => void
  onRename: (d: Dashboard) => void
  onDelete: (d: Dashboard) => void
  /** Enter edit mode for the current dashboard. Omit when it can't be edited. */
  onEdit?: () => void
  /** Consumption controls (time range, or the editor bar while editing). */
  right?: ReactNode
}) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [menuOpen, setMenuOpen] = useState(false)
  const current = dashboards.find((d) => d.id === selectedId) ?? null

  const close = () => setMenuOpen(false)

  return (
    <div className="flex items-center gap-2">
      {/* Tab strip — scrolls horizontally when dashboards overflow. */}
      <div className="flex min-w-0 flex-1 items-center gap-1 overflow-x-auto">
        {dashboards.map((d) => {
          const active = d.id === selectedId
          return (
            <button
              key={d.id}
              onClick={() => onSelect(d.id)}
              className={cn(
                'flex max-w-[200px] shrink-0 items-center gap-1.5 rounded-t-md border-b-2 px-3 py-2 text-sm transition-colors',
                active
                  ? 'border-primary font-medium text-foreground'
                  : 'border-transparent text-muted-foreground hover:text-foreground',
              )}
            >
              {defaultId === d.id && <Star size={11} className="shrink-0 fill-amber-400 text-amber-400" />}
              <span className="truncate">{d.name}</span>
            </button>
          )
        })}
      </div>

      {/* Right: consumption controls + a single options menu. */}
      <div className="flex shrink-0 items-center gap-2">
        {right}
        <div className="relative">
          <button
            onClick={() => setMenuOpen((o) => !o)}
            title={t('dashboards.tabs.options')}
            aria-label={t('dashboards.tabs.options')}
            className="flex h-9 w-9 items-center justify-center rounded-md border border-border text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            <MoreVertical size={16} />
          </button>

          {menuOpen && (
            <>
              <div className="fixed inset-0 z-10" onClick={close} />
              <div className="absolute right-0 z-20 mt-1 w-52 overflow-hidden rounded-md border border-border bg-card py-1 shadow-lg">
                {current && onEdit && (
                  <MenuItem icon={Pencil} onClick={() => { onEdit(); close() }}>
                    {t('dashboards.actions.edit')}
                  </MenuItem>
                )}
                {current && (
                  <MenuItem
                    icon={Star}
                    onClick={() => { onSetDefault(defaultId === current.id ? null : current.id); close() }}
                  >
                    {defaultId === current.id ? t('dashboards.tabs.unsetDefault') : t('dashboards.picker.setDefault')}
                  </MenuItem>
                )}
                {current && !current.systemOwner && (
                  <>
                    <MenuItem icon={Pencil} onClick={() => { onRename(current); close() }}>
                      {t('dashboards.list.rename')}
                    </MenuItem>
                    <MenuItem icon={Trash2} danger onClick={() => { onDelete(current); close() }}>
                      {t('dashboards.list.delete')}
                    </MenuItem>
                  </>
                )}
                <div className="my-1 border-t border-border" />
                <MenuItem icon={Plus} onClick={() => { onCreate(); close() }}>
                  {t('dashboards.tabs.newDashboard')}
                </MenuItem>
                <MenuItem icon={BarChart3} onClick={() => { navigate('/dashboards/visualizations'); close() }}>
                  {t('dashboards.tabs.visualizations')}
                </MenuItem>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

function MenuItem({
  icon: Icon,
  children,
  onClick,
  danger,
}: {
  icon: typeof Star
  children: ReactNode
  onClick: () => void
  danger?: boolean
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm hover:bg-muted',
        danger ? 'text-destructive' : 'text-foreground',
      )}
    >
      <Icon size={13} /> {children}
    </button>
  )
}
