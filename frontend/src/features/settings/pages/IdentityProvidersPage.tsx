import { useMemo, useState } from 'react'
import {
  AlertCircle,
  AlertTriangle,
  Check,
  Copy,
  ExternalLink,
  Eye,
  Globe2,
  Key,
  KeyRound,
  Lock,
  MoreHorizontal,
  Pencil,
  Plus,
  RefreshCw,
  ShieldCheck,
  Trash2,
  Users,
  X,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'

/* ─── Types ────────────────────────────────────────────────────────────── */

type ProviderKind = 'azure' | 'google' | 'okta' | 'keycloak' | 'saml' | 'oidc'
type Protocol = 'saml' | 'oidc'
type ProviderStatus = 'active' | 'inactive' | 'error'
type Role = 'reader' | 'maintainer' | 'admin'

interface IdentityProvider {
  id: string
  kind: ProviderKind
  name: string
  protocol: Protocol
  status: ProviderStatus
  buttonLabel: string
  // SAML
  metadataUrl?: string
  acsUrl?: string
  entityId?: string
  // OIDC
  issuerUrl?: string
  clientId?: string
  // common
  domains: string[] // e.g. ['utmstack.com', 'contractor-x.com']
  jitProvisioning: boolean
  scimEnabled: boolean
  scimEndpoint?: string
  defaultRole: Role
  enforceMfa: boolean
  signInsLast24h: number
  signInsLast30d: number
  totalUsers: number
  lastSignIn?: string
  createdAt: string
  modifiedAt: string
  lastError?: string
}

/* ─── Mock data ────────────────────────────────────────────────────────── */

const NOW = new Date('2026-05-02T16:42:00Z').getTime()
const min = (m: number) => new Date(NOW - m * 60_000).toISOString()
const hr = (h: number) => new Date(NOW - h * 3_600_000).toISOString()
const days = (d: number) => new Date(NOW - d * 86_400_000).toISOString()

const PROVIDERS: IdentityProvider[] = [
  {
    id: 'idp-azure',
    kind: 'azure',
    name: 'Microsoft Entra ID',
    protocol: 'saml',
    status: 'active',
    buttonLabel: 'Sign in with Microsoft',
    metadataUrl: 'https://login.microsoftonline.com/00000000-0000-0000-0000-000000000000/federationmetadata/2007-06/federationmetadata.xml',
    acsUrl: 'https://utmstack.acme.com/sso/saml/azure/callback',
    entityId: 'urn:utmstack:acme',
    domains: ['utmstack.com'],
    jitProvisioning: true,
    scimEnabled: true,
    scimEndpoint: 'https://utmstack.acme.com/scim/v2',
    defaultRole: 'reader',
    enforceMfa: true,
    signInsLast24h: 142,
    signInsLast30d: 4204,
    totalUsers: 18,
    lastSignIn: min(2),
    createdAt: days(540),
    modifiedAt: days(45),
  },
  {
    id: 'idp-google',
    kind: 'google',
    name: 'Google Workspace',
    protocol: 'saml',
    status: 'active',
    buttonLabel: 'Sign in with Google',
    metadataUrl: 'https://accounts.google.com/o/saml2/idp?idpid=C00abc123',
    acsUrl: 'https://utmstack.acme.com/sso/saml/google/callback',
    entityId: 'urn:utmstack:acme',
    domains: ['utmstack.com'],
    jitProvisioning: true,
    scimEnabled: false,
    defaultRole: 'reader',
    enforceMfa: true,
    signInsLast24h: 28,
    signInsLast30d: 882,
    totalUsers: 4,
    lastSignIn: min(11),
    createdAt: days(220),
    modifiedAt: days(120),
  },
  {
    id: 'idp-okta',
    kind: 'okta',
    name: 'Okta · Contractor accounts',
    protocol: 'oidc',
    status: 'active',
    buttonLabel: 'Sign in with Okta',
    issuerUrl: 'https://contractor-x.okta.com',
    clientId: '0oa1b2c3d4e5f6g7h8i9',
    domains: ['contractor-x.com'],
    jitProvisioning: true,
    scimEnabled: true,
    scimEndpoint: 'https://utmstack.acme.com/scim/v2/okta',
    defaultRole: 'reader',
    enforceMfa: false,
    signInsLast24h: 3,
    signInsLast30d: 41,
    totalUsers: 2,
    lastSignIn: hr(8),
    createdAt: days(45),
    modifiedAt: days(15),
  },
  {
    id: 'idp-keycloak',
    kind: 'keycloak',
    name: 'Keycloak (legacy)',
    protocol: 'oidc',
    status: 'error',
    buttonLabel: 'Sign in with Keycloak',
    issuerUrl: 'https://kc.utmstack.local/realms/master',
    clientId: 'utmstack-app',
    domains: [],
    jitProvisioning: false,
    scimEnabled: false,
    defaultRole: 'reader',
    enforceMfa: false,
    signInsLast24h: 0,
    signInsLast30d: 12,
    totalUsers: 0,
    lastSignIn: days(8),
    createdAt: days(720),
    modifiedAt: days(180),
    lastError: 'Discovery endpoint unreachable · last attempt 2 hours ago',
  },
]

/* ─── Style maps ───────────────────────────────────────────────────────── */

const KIND_META: Record<ProviderKind, { label: string; tone: string; logo?: string; fallbackChar: string }> = {
  azure: { label: 'Microsoft Entra', tone: 'text-sky-500', logo: '/integrations/azure.svg', fallbackChar: 'M' },
  google: { label: 'Google Workspace', tone: 'text-amber-500', logo: '/integrations/gcp.svg', fallbackChar: 'G' },
  okta: { label: 'Okta', tone: 'text-sky-500', fallbackChar: 'O' },
  keycloak: { label: 'Keycloak', tone: 'text-red-500', fallbackChar: 'K' },
  saml: { label: 'SAML 2.0', tone: 'text-violet-500', fallbackChar: 'S' },
  oidc: { label: 'OpenID Connect', tone: 'text-fuchsia-500', fallbackChar: 'O' },
}

const STATUS_META: Record<ProviderStatus, { label: string; tone: string; dot: string; pill: string }> = {
  active: { label: 'Active', tone: 'text-emerald-500', dot: 'bg-emerald-500', pill: 'bg-emerald-500/15 text-emerald-600 ring-emerald-500/30 dark:text-emerald-300' },
  inactive: { label: 'Inactive', tone: 'text-muted-foreground', dot: 'bg-muted-foreground', pill: 'bg-muted text-muted-foreground ring-border' },
  error: { label: 'Error', tone: 'text-red-500', dot: 'bg-red-500 animate-pulse', pill: 'bg-red-500/15 text-red-600 ring-red-500/30 dark:text-red-300' },
}

const ROLE_META: Record<Role, { label: string; tone: string }> = {
  reader: { label: 'Reader', tone: 'text-emerald-500' },
  maintainer: { label: 'Maintainer', tone: 'text-sky-500' },
  admin: { label: 'Admin', tone: 'text-amber-500' },
}

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function IdentityProvidersPage() {
  const [open, setOpen] = useState<IdentityProvider | null>(null)
  const [addOpen, setAddOpen] = useState(false)
  const [allowLocal, setAllowLocal] = useState(true)
  const [enforceMfaLocal, setEnforceMfaLocal] = useState(true)
  const [sessionTimeout, setSessionTimeout] = useState<'30m' | '8h' | '24h' | '7d'>('8h')

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header onAdd={() => setAddOpen(true)} />

      <div className="mt-6">
        <StatsStrip />
      </div>

      <div className="mt-6 grid grid-cols-1 gap-5 lg:grid-cols-[1fr_360px]">
        <div className="space-y-5">
          <ProvidersSection providers={PROVIDERS} onOpen={setOpen} onAdd={() => setAddOpen(true)} />
          <AuthenticationSection
            allowLocal={allowLocal}
            onAllowLocal={setAllowLocal}
            enforceMfa={enforceMfaLocal}
            onEnforceMfa={setEnforceMfaLocal}
            sessionTimeout={sessionTimeout}
            onSessionTimeout={setSessionTimeout}
          />
        </div>

        <aside className="lg:sticky lg:top-4 lg:h-fit">
          <LoginPreview providers={PROVIDERS.filter((p) => p.status === 'active')} allowLocal={allowLocal} />
        </aside>
      </div>

      {open && <ProviderDrawer provider={open} onClose={() => setOpen(null)} />}
      {addOpen && <AddProviderDialog onClose={() => setAddOpen(false)} />}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({ onAdd }: { onAdd: () => void }) {
  return (
    <header className="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h1 className="flex items-center gap-2 text-xl font-semibold">
          <ShieldCheck size={18} strokeWidth={1.75} />
          Identity providers
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Single sign-on for your organization. Configure SAML 2.0 or OpenID Connect providers, JIT
          user provisioning, and SCIM lifecycle sync.
        </p>
      </div>
      <Button size="sm" onClick={onAdd}>
        <Plus size={14} className="mr-2" />
        Add provider
      </Button>
    </header>
  )
}

/* ─── Stats strip ──────────────────────────────────────────────────────── */

function StatsStrip() {
  const active = PROVIDERS.filter((p) => p.status === 'active').length
  const errored = PROVIDERS.filter((p) => p.status === 'error').length
  const signIns24h = PROVIDERS.reduce((s, p) => s + p.signInsLast24h, 0)
  const ssoUsers = PROVIDERS.reduce((s, p) => s + p.totalUsers, 0)
  return (
    <section className="rounded-xl border border-border bg-card">
      <div className="grid grid-cols-1 divide-y divide-border sm:grid-cols-4 sm:divide-x sm:divide-y-0">
        <StripStat label="Active providers" value={String(active)} sub={errored > 0 ? `${errored} with issues` : 'All healthy'} tone={errored > 0 ? 'text-amber-500' : 'text-emerald-500'} />
        <StripStat label="SSO users" value={String(ssoUsers)} sub="Across all providers" />
        <StripStat label="Sign-ins · 24h" value={signIns24h.toLocaleString()} sub="Successful authentications" />
        <StripStat label="MFA enforced" value="Yes" sub="On all providers and local accounts" tone="text-emerald-500" />
      </div>
    </section>
  )
}

function StripStat({
  label,
  value,
  sub,
  tone,
}: {
  label: string
  value: string
  sub: string
  tone?: string
}) {
  return (
    <div className="px-5 py-4">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className={cn('mt-1 font-mono text-2xl font-semibold tabular-nums', tone)}>{value}</div>
      <div className="mt-0.5 text-[11px] text-muted-foreground">{sub}</div>
    </div>
  )
}

/* ─── Providers section ────────────────────────────────────────────────── */

function ProvidersSection({
  providers,
  onOpen,
  onAdd,
}: {
  providers: IdentityProvider[]
  onOpen: (p: IdentityProvider) => void
  onAdd: () => void
}) {
  return (
    <Section
      title="Configured providers"
      subtitle="Each provider becomes a sign-in option on the login page."
    >
      <div className="space-y-2">
        {providers.map((p) => (
          <ProviderRow key={p.id} provider={p} onOpen={() => onOpen(p)} />
        ))}
        <button
          onClick={onAdd}
          className="flex w-full items-center justify-center gap-2 rounded-lg border border-dashed border-border bg-card px-4 py-3 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        >
          <Plus size={13} />
          Add identity provider
        </button>
      </div>
    </Section>
  )
}

function ProviderRow({ provider: p, onOpen }: { provider: IdentityProvider; onOpen: () => void }) {
  const st = STATUS_META[p.status]
  return (
    <div
      onClick={onOpen}
      className={cn(
        'group flex cursor-pointer items-center gap-3 rounded-lg border border-border bg-card px-3 py-3 text-xs transition-colors hover:bg-muted/40',
        p.status === 'inactive' && 'opacity-60'
      )}
    >
      <ProviderLogo kind={p.kind} />
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-medium">{p.name}</span>
          <span className={cn('inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium ring-1 ring-inset', st.pill)}>
            <span className={cn('h-1 w-1 rounded-full', st.dot)} />
            {st.label}
          </span>
        </div>
        <div className="mt-0.5 flex flex-wrap items-center gap-1.5 text-[11px] text-muted-foreground">
          <span className="font-mono uppercase">{p.protocol === 'saml' ? 'SAML 2.0' : 'OIDC'}</span>
          {p.domains.length > 0 && (
            <>
              <span>·</span>
              <span>{p.domains.join(', ')}</span>
            </>
          )}
          {p.jitProvisioning && (
            <>
              <span>·</span>
              <span className="inline-flex items-center gap-1">
                <Users size={10} />
                JIT
              </span>
            </>
          )}
          {p.scimEnabled && (
            <>
              <span>·</span>
              <span className="inline-flex items-center gap-1">
                <RefreshCw size={10} />
                SCIM
              </span>
            </>
          )}
          {p.enforceMfa && (
            <>
              <span>·</span>
              <span className="inline-flex items-center gap-1 text-emerald-500">
                <Lock size={10} />
                MFA
              </span>
            </>
          )}
        </div>
      </div>
      <div className="hidden text-right text-[11px] text-muted-foreground sm:block">
        <div className="font-mono tabular-nums">{p.signInsLast24h} signed in</div>
        <div className="text-[10px]">last 24h</div>
      </div>
      <div className="hidden text-right text-[11px] text-muted-foreground md:block">
        <div className="font-mono tabular-nums">{p.totalUsers} users</div>
        <div className="text-[10px]">{p.lastSignIn ? `last seen ${relativeTime(p.lastSignIn)}` : 'never used'}</div>
      </div>
      <button
        onClick={(e) => e.stopPropagation()}
        className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground opacity-0 hover:bg-background hover:text-foreground group-hover:opacity-100"
      >
        <MoreHorizontal size={14} />
      </button>
    </div>
  )
}

function ProviderLogo({ kind, size = 36 }: { kind: ProviderKind; size?: number }) {
  const meta = KIND_META[kind]
  if (meta.logo) {
    return (
      <div
        className="flex shrink-0 items-center justify-center rounded-lg border border-border bg-white p-1.5"
        style={{ width: size, height: size }}
      >
        <img src={meta.logo} alt={meta.label} className="max-h-full max-w-full object-contain" />
      </div>
    )
  }
  return (
    <div
      className={cn(
        'flex shrink-0 items-center justify-center rounded-lg bg-gradient-to-br text-sm font-semibold text-white',
        kind === 'okta' && 'from-sky-500 to-blue-600',
        kind === 'keycloak' && 'from-red-500 to-rose-600',
        kind === 'saml' && 'from-violet-500 to-purple-600',
        kind === 'oidc' && 'from-fuchsia-500 to-pink-600'
      )}
      style={{ width: size, height: size }}
    >
      {meta.fallbackChar}
    </div>
  )
}

/* ─── Authentication settings ──────────────────────────────────────────── */

function AuthenticationSection({
  allowLocal,
  onAllowLocal,
  enforceMfa,
  onEnforceMfa,
  sessionTimeout,
  onSessionTimeout,
}: {
  allowLocal: boolean
  onAllowLocal: (v: boolean) => void
  enforceMfa: boolean
  onEnforceMfa: (v: boolean) => void
  sessionTimeout: '30m' | '8h' | '24h' | '7d'
  onSessionTimeout: (v: '30m' | '8h' | '24h' | '7d') => void
}) {
  return (
    <Section
      title="Authentication settings"
      subtitle="Apply to all sign-in methods, including local accounts and SSO."
    >
      <ToggleRow
        icon={KeyRound}
        label="Allow local sign-in"
        description="When disabled, all members must sign in through an identity provider above."
        enabled={allowLocal}
        onChange={onAllowLocal}
      />
      <ToggleRow
        icon={Lock}
        label="Enforce MFA on local accounts"
        description="Members signing in with email + password must complete a second factor (TOTP / WebAuthn)."
        enabled={enforceMfa}
        onChange={onEnforceMfa}
        disabled={!allowLocal}
      />
      <Field label="Session timeout">
        <div className="grid grid-cols-4 gap-2">
          {[
            { id: '30m' as const, label: '30 min' },
            { id: '8h' as const, label: '8 hours' },
            { id: '24h' as const, label: '24 hours' },
            { id: '7d' as const, label: '7 days' },
          ].map((o) => {
            const active = sessionTimeout === o.id
            return (
              <button
                key={o.id}
                onClick={() => onSessionTimeout(o.id)}
                className={cn(
                  'rounded-lg border px-2.5 py-2 text-xs transition-colors',
                  active ? 'border-primary/40 bg-primary/5 text-primary font-medium' : 'border-border hover:bg-muted/40'
                )}
              >
                {o.label}
              </button>
            )
          })}
        </div>
      </Field>
    </Section>
  )
}

function ToggleRow({
  icon: Icon,
  label,
  description,
  enabled,
  onChange,
  disabled,
}: {
  icon: LucideIcon
  label: string
  description: string
  enabled: boolean
  onChange: (v: boolean) => void
  disabled?: boolean
}) {
  return (
    <div className={cn('flex items-start gap-3 border-b border-border/60 py-3 last:border-b-0', disabled && 'opacity-50')}>
      <Icon size={14} className="mt-0.5 shrink-0 text-muted-foreground" />
      <div className="min-w-0 flex-1">
        <div className="text-sm font-medium">{label}</div>
        <div className="text-xs text-muted-foreground">{description}</div>
      </div>
      <Toggle enabled={enabled} onChange={() => !disabled && onChange(!enabled)} />
    </div>
  )
}

function Toggle({ enabled, onChange }: { enabled: boolean; onChange: () => void }) {
  return (
    <button
      onClick={onChange}
      className={cn(
        'relative h-5 w-9 shrink-0 rounded-full transition-colors',
        enabled ? 'bg-primary' : 'bg-muted'
      )}
    >
      <span
        className={cn(
          'absolute top-0.5 h-4 w-4 rounded-full bg-white shadow transition-all',
          enabled ? 'left-4.5' : 'left-0.5'
        )}
      />
    </button>
  )
}

/* ─── Login preview ────────────────────────────────────────────────────── */

function LoginPreview({
  providers,
  allowLocal,
}: {
  providers: IdentityProvider[]
  allowLocal: boolean
}) {
  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card">
      <div className="border-b border-border px-5 py-3">
        <div className="text-[11px] uppercase tracking-wider text-muted-foreground">Preview</div>
        <div className="mt-0.5 text-sm font-semibold">Login page</div>
      </div>
      <div className="bg-muted/30 p-5">
        <div className="rounded-lg border border-border bg-card p-5 text-center">
          <div className="text-sm font-semibold">Sign in to UTMStack</div>
          <p className="mt-1 text-[11px] text-muted-foreground">
            Choose a method below.
          </p>

          <div className="mt-5 space-y-2">
            {providers.map((p) => {
              const meta = KIND_META[p.kind]
              return (
                <button
                  key={p.id}
                  className="flex w-full items-center gap-3 rounded-md border border-border bg-card px-3 py-2 text-xs text-foreground transition-colors hover:bg-muted/40"
                >
                  <ProviderLogo kind={p.kind} size={20} />
                  <span className="flex-1 text-left">{p.buttonLabel}</span>
                  <span className={cn('text-[10px]', meta.tone)}>↗</span>
                </button>
              )
            })}
            {providers.length === 0 && (
              <div className="rounded-md border border-dashed border-border bg-card px-3 py-4 text-[11px] italic text-muted-foreground">
                No SSO providers configured yet.
              </div>
            )}
          </div>

          {allowLocal && (
            <>
              <div className="my-4 flex items-center gap-2 text-[10px] text-muted-foreground">
                <span className="h-px flex-1 bg-border" />
                or
                <span className="h-px flex-1 bg-border" />
              </div>
              <div className="space-y-2 text-left">
                <input
                  type="email"
                  placeholder="Email"
                  className="h-8 w-full rounded-md border border-border bg-background/40 px-2 text-xs"
                  readOnly
                />
                <input
                  type="password"
                  placeholder="Password"
                  className="h-8 w-full rounded-md border border-border bg-background/40 px-2 text-xs"
                  readOnly
                />
                <button className="h-8 w-full rounded-md bg-foreground text-xs font-medium text-background">
                  Sign in
                </button>
              </div>
            </>
          )}

          {!allowLocal && providers.length > 0 && (
            <div className="mt-4 flex items-start gap-1.5 text-[10px] text-muted-foreground">
              <Lock size={10} className="mt-0.5 shrink-0" />
              <span>Local sign-in is disabled. SSO only.</span>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

/* ─── Provider drawer ──────────────────────────────────────────────────── */

function ProviderDrawer({ provider: p, onClose }: { provider: IdentityProvider; onClose: () => void }) {
  const st = STATUS_META[p.status]
  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[860px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <ProviderLogo kind={p.kind} size={48} />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span className="font-mono">{p.id}</span>
                  <span>·</span>
                  <span className="uppercase">{p.protocol === 'saml' ? 'SAML 2.0' : 'OIDC'}</span>
                  <span>·</span>
                  <span>created {relativeTime(p.createdAt)}</span>
                </div>
                <h2 className="mt-0.5 truncate text-xl font-semibold">{p.name}</h2>
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                  <span className={cn('inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset', st.pill)}>
                    <span className={cn('h-1 w-1 rounded-full', st.dot)} />
                    {st.label}
                  </span>
                  {p.jitProvisioning && (
                    <span className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5">
                      <Users size={10} />
                      JIT provisioning
                    </span>
                  )}
                  {p.scimEnabled && (
                    <span className="inline-flex items-center gap-1 rounded-md border border-border bg-background/40 px-1.5 py-0.5">
                      <RefreshCw size={10} />
                      SCIM
                    </span>
                  )}
                  {p.enforceMfa && (
                    <span className="inline-flex items-center gap-1 rounded-md border border-emerald-500/30 bg-emerald-500/5 px-1.5 py-0.5 text-emerald-600 dark:text-emerald-300">
                      <Lock size={10} />
                      MFA enforced
                    </span>
                  )}
                </div>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>

          {p.lastError && (
            <div className="mt-3 rounded-md border border-red-500/30 bg-red-500/5 p-3 text-xs">
              <div className="flex items-center gap-2 font-medium text-red-600 dark:text-red-300">
                <AlertTriangle size={13} />
                Connection error
              </div>
              <p className="mt-1 text-muted-foreground">{p.lastError}</p>
            </div>
          )}

          <div className="mt-4 flex flex-wrap items-center gap-2">
            <Button size="sm">
              <Pencil size={13} className="mr-1.5" />
              Edit
            </Button>
            <Button size="sm" variant="outline">
              <RefreshCw size={13} className="mr-1.5" />
              Test connection
            </Button>
            {p.status === 'active' ? (
              <Button size="sm" variant="outline">
                <Eye size={13} className="mr-1.5" />
                Disable
              </Button>
            ) : (
              <Button size="sm" variant="outline">
                <Eye size={13} className="mr-1.5" />
                Enable
              </Button>
            )}
            <Button size="sm" variant="outline" className="ml-auto text-red-500 hover:bg-red-500/10">
              <Trash2 size={13} className="mr-1.5" />
              Delete
            </Button>
          </div>
        </header>

        <div className="flex-1 space-y-4 overflow-y-auto bg-muted/20 p-6">
          {/* Connection details */}
          <Section title="Connection" subtitle="Endpoints used for authentication.">
            {p.protocol === 'saml' ? (
              <dl className="grid grid-cols-[180px_1fr] gap-y-2 text-xs">
                <Row k="Metadata URL"><CopyableMono value={p.metadataUrl ?? ''} /></Row>
                <Row k="ACS URL (callback)"><CopyableMono value={p.acsUrl ?? ''} /></Row>
                <Row k="Entity ID"><CopyableMono value={p.entityId ?? ''} /></Row>
              </dl>
            ) : (
              <dl className="grid grid-cols-[180px_1fr] gap-y-2 text-xs">
                <Row k="Issuer URL"><CopyableMono value={p.issuerUrl ?? ''} /></Row>
                <Row k="Client ID"><CopyableMono value={p.clientId ?? ''} /></Row>
                <Row k="Redirect URI"><CopyableMono value={`https://utmstack.acme.com/sso/oidc/${p.id}/callback`} /></Row>
              </dl>
            )}
          </Section>

          {/* User mapping */}
          <Section title="User mapping" subtitle="How users from this provider map onto UTMStack accounts.">
            <dl className="grid grid-cols-[180px_1fr] gap-y-2 text-xs">
              <Row k="Allowed domains">
                {p.domains.length > 0 ? (
                  <div className="flex flex-wrap gap-1.5">
                    {p.domains.map((d) => (
                      <span key={d} className="rounded-md bg-muted px-1.5 py-0.5 font-mono text-[11px]">
                        @{d}
                      </span>
                    ))}
                  </div>
                ) : (
                  <span className="italic text-muted-foreground">Any domain</span>
                )}
              </Row>
              <Row k="JIT provisioning">
                {p.jitProvisioning ? (
                  <span className="inline-flex items-center gap-1 text-emerald-500">
                    <Check size={11} strokeWidth={3} />
                    Enabled — users are auto-created on first sign-in
                  </span>
                ) : (
                  <span className="text-muted-foreground">Disabled — admins must invite users manually</span>
                )}
              </Row>
              <Row k="Default role">
                <span className={cn('font-medium', ROLE_META[p.defaultRole].tone)}>
                  {ROLE_META[p.defaultRole].label}
                </span>
              </Row>
            </dl>
          </Section>

          {/* SCIM */}
          {p.scimEnabled && (
            <Section title="SCIM provisioning" subtitle="The provider syncs users and groups into UTMStack.">
              <dl className="grid grid-cols-[180px_1fr] gap-y-2 text-xs">
                <Row k="Endpoint URL"><CopyableMono value={p.scimEndpoint ?? ''} /></Row>
                <Row k="Bearer token">
                  <span className="font-mono text-muted-foreground">scim_••••••••</span>{' '}
                  <button className="ml-2 text-primary hover:underline">Rotate</button>
                </Row>
              </dl>
            </Section>
          )}

          {/* Activity */}
          <Section title="Activity">
            <div className="grid grid-cols-3 gap-3">
              <Stat label="Sign-ins · 24h" value={p.signInsLast24h.toLocaleString()} />
              <Stat label="Sign-ins · 30d" value={p.signInsLast30d.toLocaleString()} />
              <Stat label="Total users" value={String(p.totalUsers)} />
            </div>
            {p.lastSignIn && (
              <p className="mt-3 text-[11px] text-muted-foreground">
                Last sign-in: <span className="font-mono text-foreground">{absTimestamp(p.lastSignIn)}</span>
              </p>
            )}
          </Section>
        </div>
      </div>
    </div>
  )
}

function CopyableMono({ value }: { value: string }) {
  const [copied, setCopied] = useState(false)
  return (
    <div className="flex items-center gap-2">
      <code className="flex-1 truncate rounded bg-muted/40 px-2 py-1 font-mono text-[11px]">
        {value}
      </code>
      <button
        onClick={() => {
          navigator.clipboard.writeText(value)
          setCopied(true)
          setTimeout(() => setCopied(false), 1500)
        }}
        className="flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
      >
        {copied ? <Check size={11} className="text-emerald-500" /> : <Copy size={11} />}
      </button>
    </div>
  )
}

/* ─── Add provider dialog ──────────────────────────────────────────────── */

function AddProviderDialog({ onClose }: { onClose: () => void }) {
  const [pick, setPick] = useState<ProviderKind | null>(null)
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-2xl flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div>
            <h2 className="flex items-center gap-2 text-lg font-semibold">
              <ShieldCheck size={18} />
              Add identity provider
            </h2>
            <p className="mt-1 text-xs text-muted-foreground">
              Choose your provider to start the configuration wizard.
            </p>
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>

        <div className="px-6 py-5">
          <div className="text-[11px] uppercase tracking-wider text-muted-foreground">Common providers</div>
          <div className="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-3">
            {(['azure', 'google', 'okta', 'keycloak'] as ProviderKind[]).map((k) => {
              const meta = KIND_META[k]
              return (
                <button
                  key={k}
                  onClick={() => setPick(k)}
                  className={cn(
                    'flex items-center gap-2.5 rounded-lg border p-3 text-left transition-colors',
                    pick === k ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/40'
                  )}
                >
                  <ProviderLogo kind={k} size={32} />
                  <div className="min-w-0 flex-1">
                    <div className="text-sm font-medium">{meta.label}</div>
                    <div className="text-[10px] text-muted-foreground">
                      {k === 'azure' && 'SAML 2.0'}
                      {k === 'google' && 'SAML 2.0'}
                      {k === 'okta' && 'SAML or OIDC'}
                      {k === 'keycloak' && 'OIDC'}
                    </div>
                  </div>
                </button>
              )
            })}
          </div>

          <div className="mt-5 text-[11px] uppercase tracking-wider text-muted-foreground">Generic</div>
          <div className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2">
            {(['saml', 'oidc'] as ProviderKind[]).map((k) => {
              const meta = KIND_META[k]
              return (
                <button
                  key={k}
                  onClick={() => setPick(k)}
                  className={cn(
                    'flex items-center gap-2.5 rounded-lg border p-3 text-left transition-colors',
                    pick === k ? 'border-primary/40 bg-primary/5' : 'border-border hover:bg-muted/40'
                  )}
                >
                  <ProviderLogo kind={k} size={32} />
                  <div className="min-w-0 flex-1">
                    <div className="text-sm font-medium">{meta.label}</div>
                    <div className="text-[10px] text-muted-foreground">
                      {k === 'saml' ? 'Any SAML 2.0 IdP' : 'Any OIDC IdP'}
                    </div>
                  </div>
                </button>
              )
            })}
          </div>

          {pick && (
            <div className="mt-5 rounded-md border border-border bg-background/40 p-4 text-xs">
              <div className="font-medium">Setup steps</div>
              <ol className="mt-2 list-decimal space-y-1 pl-5 text-muted-foreground">
                <li>Configure the application in <span className="text-foreground">{KIND_META[pick].label}</span>.</li>
                <li>Copy the metadata URL or issuer URL into UTMStack.</li>
                <li>Map allowed domains and pick a default role.</li>
                <li>Test the connection from the next screen.</li>
              </ol>
              <a
                className="mt-3 inline-flex items-center gap-1 text-[11px] text-primary hover:underline"
                href="#"
              >
                <ExternalLink size={11} />
                Open the {KIND_META[pick].label} setup guide
              </a>
            </div>
          )}
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
          <Button size="sm" variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button size="sm" disabled={!pick}>
            Continue
          </Button>
        </footer>
      </div>
    </div>
  )
}

/* ─── Small parts ──────────────────────────────────────────────────────── */

function Section({
  title,
  subtitle,
  children,
}: {
  title: string
  subtitle?: string
  children: React.ReactNode
}) {
  return (
    <section className="rounded-xl border border-border bg-card p-6">
      <header className="mb-4">
        <h2 className="text-sm font-semibold">{title}</h2>
        {subtitle && <p className="mt-0.5 text-xs text-muted-foreground">{subtitle}</p>}
      </header>
      {children}
    </section>
  )
}

function Field({
  label,
  hint,
  children,
}: {
  label: string
  hint?: string
  children: React.ReactNode
}) {
  return (
    <div className="border-b border-border/60 py-3 last:border-b-0">
      <label className="mb-2 block text-[11px] uppercase tracking-wider text-muted-foreground">
        {label}
      </label>
      {children}
      {hint && <p className="mt-1 text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}

function Row({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{k}</dt>
      <dd className="break-words">{children}</dd>
    </>
  )
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border border-border bg-background/40 p-3">
      <div className="text-[10px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className="mt-0.5 font-mono text-base font-semibold tabular-nums">{value}</div>
    </div>
  )
}

/* ─── Helpers ──────────────────────────────────────────────────────────── */

function relativeTime(iso: string) {
  if (!iso) return '—'
  const diff = NOW - new Date(iso).getTime()
  const m = Math.round(diff / 60_000)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  const h = Math.round(m / 60)
  if (h < 24) return `${h}h ago`
  return `${Math.round(h / 24)}d ago`
}

function absTimestamp(iso: string) {
  return new Date(iso).toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// Suppress unused-import warning
void useMemo
void Globe2
void Key
void AlertCircle
