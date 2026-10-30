import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { adAuditHttpService } from '../services/ad-audit-http.service'
import type { Identity, IdentityDetail, IdentityListItem } from '../types/ad-user.types'

/**
 * Shared user-auditor context: the identity alias list plus all identity
 * actions. Loaded once here so Toolbar (filter dropdown) and
 * IdentityGroupPicker (assign/create/delete) share the same data.
 */
interface UserAuditorContextValue {
  identities: IdentityListItem[]
  identitiesLoaded: boolean
  reloadIdentities: () => void
  identityDetail: (id: string) => Promise<IdentityDetail>
  createIdentity: (name: string) => Promise<Identity | null>
  deleteIdentity: (id: string) => Promise<boolean>
  assignIdentity: (userId: string, identityId: string) => Promise<boolean>
}

const UserAuditorContext = createContext<UserAuditorContextValue | null>(null)

export function UserAuditorProvider({ children }: { children: ReactNode }) {
  const [identities, setIdentities] = useState<IdentityListItem[]>([])
  const [identitiesLoaded, setIdentitiesLoaded] = useState(false)

  const reloadIdentities = useCallback(() => {
    void adAuditHttpService
      .identities()
      .then((res) => setIdentities(res ?? []))
      .catch(() => setIdentities([]))
      .finally(() => setIdentitiesLoaded(true))
  }, [])

  useEffect(() => {
    reloadIdentities()
  }, [reloadIdentities])

  const createIdentity = useCallback(
    async (name: string) => {
      const res = await adAuditHttpService.createIdentity(name)
      if (res) reloadIdentities()
      return res
    },
    [reloadIdentities],
  )

  const deleteIdentity = useCallback(
    async (id: string) => {
      const ok = await adAuditHttpService.deleteIdentity(id)
      if (ok) reloadIdentities()
      return Boolean(ok)
    },
    [reloadIdentities],
  )

  const assignIdentity = useCallback(
    async (userId: string, identityId: string) => {
      const res = await adAuditHttpService.assignIdentity(userId, identityId)
      return Boolean(res)
    },
    [],
  )

  const value = useMemo<UserAuditorContextValue>(
    () => ({
      identities,
      identitiesLoaded,
      reloadIdentities,
      identityDetail: (id) => adAuditHttpService.identityDetail(id),
      createIdentity,
      deleteIdentity,
      assignIdentity,
    }),
    [identities, identitiesLoaded, reloadIdentities, createIdentity, deleteIdentity, assignIdentity],
  )

  return <UserAuditorContext.Provider value={value}>{children}</UserAuditorContext.Provider>
}

export function useUserAuditor() {
  const ctx = useContext(UserAuditorContext)
  if (!ctx) throw new Error('useUserAuditor must be used within <UserAuditorProvider>')
  return ctx
}
