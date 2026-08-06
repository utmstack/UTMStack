import { useEffect, useState } from 'react'
import { ChevronDown, UserPlus } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/shared/lib/utils'
import { usersHttpService } from '@/features/team/services/team-http.service'
import type { UserListItem } from '@/features/team/types/team.types'
import { Menu } from './menu'

// Assign / reassign / clear an alert's owner. Loads the user list lazily on first
// open so the dropdown isn't fetched for every alert row.
export function AssigneeMenu({ current, onAssign }: { current?: string; onAssign: (assignee: string) => void }) {
  const { t } = useTranslation()
  const [users, setUsers] = useState<UserListItem[] | null>(null)
  const [loaded, setLoaded] = useState(false)

  useEffect(() => {
    if (!loaded) return
    let cancelled = false
    usersHttpService
      .list({ page_size: 200 })
      .then((r) => {
        if (!cancelled) setUsers(r.data ?? [])
      })
      .catch(() => {
        if (!cancelled) setUsers([])
      })
    return () => {
      cancelled = true
    }
  }, [loaded])

  const label = (u: UserListItem) => u.name?.trim() || u.email

  return (
    <div onMouseEnter={() => setLoaded(true)} onFocusCapture={() => setLoaded(true)} onClickCapture={() => setLoaded(true)}>
      <Menu
        trigger={
          <>
            <UserPlus size={13} className="mr-1.5" />
            {current ? current : t('alerts.drawer.assign')}
            <ChevronDown size={12} className="ml-1" />
          </>
        }
      >
        {users == null ? (
          <div className="px-3 py-2 text-xs text-muted-foreground">{t('alerts.drawer.loadingUsers')}</div>
        ) : (
          <>
            {users.map((u) => (
              <button
                key={u.id}
                onClick={() => onAssign(label(u))}
                className={cn(
                  'block w-full px-3 py-1.5 text-left text-sm hover:bg-muted',
                  current === label(u) && 'font-semibold text-primary'
                )}
              >
                {label(u)}
              </button>
            ))}
            {current && (
              <button
                onClick={() => onAssign('')}
                className="block w-full border-t border-border px-3 py-1.5 text-left text-sm text-muted-foreground hover:bg-muted"
              >
                {t('alerts.drawer.unassign')}
              </button>
            )}
          </>
        )}
      </Menu>
    </div>
  )
}
