import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  ArrowRight,
  Clock,
  Globe,
  Loader2,
  Plus,
  Search,
  Server,
  Trash2,
  X,
} from 'lucide-react'
import { ApiError } from '@/shared/lib/api-client'
import { setCurrentInstanceId } from '@/shared/lib/current-instance'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { PasswordInput } from '@/shared/components/ui/password-input'
import { Topbar } from '@/shared/layouts/DashboardLayout/Topbar'
import { useAuth } from '@/features/auth'
import { useInstances, useInstanceMutations } from '../hooks/use-instances'
import type { FederationInstance, FederationInstanceInput } from '../types'

// Deterministic hue per instance so each card carries its own accent color.
function hueOf(s: string): number {
  let h = 0
  for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) % 360
  return h
}

function hostOf(url: string): string {
  try {
    return new URL(url).host
  } catch {
    return url
  }
}

// Localized "added N ago", falling back to an absolute date for older entries.
function formatAdded(iso: string): string {
  const then = new Date(iso).getTime()
  if (Number.isNaN(then)) return 'recently'
  const diff = Math.round((then - Date.now()) / 1000)
  const abs = Math.abs(diff)
  const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' })
  if (abs < 60) return rtf.format(diff, 'second')
  if (abs < 3600) return rtf.format(Math.round(diff / 60), 'minute')
  if (abs < 86400) return rtf.format(Math.round(diff / 3600), 'hour')
  if (abs < 2592000) return rtf.format(Math.round(diff / 86400), 'day')
  return new Date(iso).toLocaleDateString('en', { month: 'short', day: 'numeric', year: 'numeric' })
}

export function InstancePickerPage() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const { data: instances, isLoading } = useInstances()
  const { remove } = useInstanceMutations()
  const [showAdd, setShowAdd] = useState(false)
  const [query, setQuery] = useState('')

  const select = (id: number) => {
    setCurrentInstanceId(id)
    navigate('/home', { replace: true })
  }

  const all = instances ?? []
  const filtered = all.filter(
    (i) =>
      i.name.toLowerCase().includes(query.toLowerCase()) ||
      i.baseUrl.toLowerCase().includes(query.toLowerCase()),
  )
  return (
    <div className="relative min-h-screen overflow-hidden bg-background text-foreground">
      {/* Ambient gradient glows (echo the login page). */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0"
        style={{
          background:
            'radial-gradient(60% 50% at 50% -10%, hsl(211 95% 55% / 0.18) 0%, transparent 60%), radial-gradient(45% 45% at 100% 0%, hsl(265 80% 55% / 0.10) 0%, transparent 55%)',
        }}
      />
      <div
        aria-hidden
        className="fs-spin-slow pointer-events-none absolute -top-56 left-1/2 h-[40rem] w-[40rem] -translate-x-1/2 rounded-full opacity-[0.07] blur-3xl"
        style={{ background: 'conic-gradient(from 0deg, #1a8cff, #0ea5e9, #2563eb, #1a8cff)' }}
      />

      {/* The exact same header as the rest of the app. */}
      <Topbar />

      <main className="relative z-10 mx-auto w-full max-w-6xl px-6 pb-16 pt-6">
        {/* Hero */}
        <div className="fs-fade-up flex flex-col items-start gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight sm:text-3xl">
              Welcome back{user?.name ? `, ${user.name}` : ''}
            </h1>
            <p className="mt-1.5 text-sm text-muted-foreground">
              Select an instance to manage, or connect a new one.
            </p>
          </div>
          <div className="flex items-center gap-3">
            {all.length > 3 && (
              <div className="relative">
                <Search
                  size={15}
                  className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
                />
                <Input
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  placeholder="Search instances"
                  className="h-9 w-56 pl-9"
                />
              </div>
            )}
            <Button onClick={() => setShowAdd(true)}>
              <Plus size={16} className="mr-1.5" /> Connect instance
            </Button>
          </div>
        </div>

        {/* Stats */}
        {all.length > 0 && (
          <div className="fs-fade-up mt-6 flex flex-wrap gap-3" style={{ animationDelay: '60ms' }}>
            <StatChip label={all.length === 1 ? 'Connected instance' : 'Connected instances'} value={all.length} />
          </div>
        )}

        {/* Grid */}
        <div className="mt-8">
          {isLoading ? (
            <div className="flex justify-center py-24 text-muted-foreground">
              <Loader2 className="h-6 w-6 animate-spin" />
            </div>
          ) : all.length === 0 ? (
            <EmptyState onAdd={() => setShowAdd(true)} />
          ) : (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {filtered.map((inst, i) => (
                <InstanceCard
                  key={inst.id}
                  instance={inst}
                  index={i}
                  onSelect={() => select(inst.id)}
                  onRemove={() => remove.mutate(inst.id)}
                />
              ))}
              <button
                onClick={() => setShowAdd(true)}
                className="fs-fade-up group flex min-h-[150px] flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-border text-muted-foreground transition-colors hover:border-primary/50 hover:bg-primary/5 hover:text-foreground"
                style={{ animationDelay: `${filtered.length * 50}ms` }}
              >
                <span className="flex h-10 w-10 items-center justify-center rounded-full bg-muted transition-colors group-hover:bg-primary/15 group-hover:text-primary">
                  <Plus size={20} />
                </span>
                <span className="text-sm font-medium">Connect instance</span>
              </button>
            </div>
          )}
        </div>
      </main>

      {showAdd && <AddInstanceModal onClose={() => setShowAdd(false)} />}
    </div>
  )
}

function StatChip({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex items-center gap-2 rounded-lg border border-border bg-card/60 px-3 py-2 backdrop-blur-sm">
      <span className="text-lg font-semibold tabular-nums">{value}</span>
      <span className="text-xs text-muted-foreground">{label}</span>
    </div>
  )
}

function InstanceCard({
  instance,
  index,
  onSelect,
  onRemove,
}: {
  instance: FederationInstance
  index: number
  onSelect: () => void
  onRemove: () => void
}) {
  const hue = hueOf(instance.name || instance.baseUrl)
  const initial = (instance.name || hostOf(instance.baseUrl)).charAt(0).toUpperCase()
  return (
    <div className="fs-fade-up" style={{ animationDelay: `${index * 50}ms` }}>
      <button
        onClick={onSelect}
        className="group relative flex min-h-[178px] w-full flex-col overflow-hidden rounded-xl border border-border bg-card p-5 text-left shadow-sm transition-all duration-200 hover:-translate-y-1 hover:border-primary/50 hover:shadow-lg hover:shadow-primary/5"
      >
        {/* top accent */}
        <span
          aria-hidden
          className="absolute inset-x-0 top-0 h-1 opacity-80"
          style={{
            background: `linear-gradient(90deg, hsl(${hue} 80% 55%), hsl(${(hue + 40) % 360} 80% 55%))`,
          }}
        />
        {/* tinted corner glow */}
        <span
          aria-hidden
          className="pointer-events-none absolute -right-16 -top-12 h-40 w-40 rounded-full opacity-20 blur-2xl transition-opacity duration-300 group-hover:opacity-40"
          style={{ background: `radial-gradient(circle, hsl(${hue} 80% 55%) 0%, transparent 70%)` }}
        />
        {/* faint server watermark */}
        <Server
          aria-hidden
          size={128}
          strokeWidth={1}
          className="pointer-events-none absolute -bottom-7 -right-5 text-foreground/[0.035]"
        />

        {/* header */}
        <div className="relative flex items-start justify-between">
          <span
            className="flex h-12 w-12 items-center justify-center rounded-xl text-lg font-semibold text-white shadow-sm ring-1 ring-white/10"
            style={{
              background: `linear-gradient(135deg, hsl(${hue} 72% 52%), hsl(${(hue + 40) % 360} 72% 46%))`,
            }}
          >
            {initial || <Server size={22} />}
          </span>
          <span
            role="button"
            tabIndex={0}
            aria-label="Disconnect instance"
            onClick={(e) => {
              e.stopPropagation()
              onRemove()
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.stopPropagation()
                onRemove()
              }
            }}
            className="rounded p-1.5 text-muted-foreground opacity-0 transition hover:bg-destructive/10 hover:text-destructive group-hover:opacity-100"
          >
            <Trash2 size={15} />
          </span>
        </div>

        {/* name + host */}
        <div className="relative mt-4 min-w-0 flex-1">
          <div className="truncate text-base font-semibold">{instance.name}</div>
          <div className="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
            <Globe size={12} className="shrink-0 opacity-70" />
            <span className="truncate font-mono">{hostOf(instance.baseUrl)}</span>
          </div>
        </div>

        {/* footer */}
        <div className="relative mt-4 flex items-center justify-between border-t border-border/60 pt-3">
          <span className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
            <Clock size={12} className="opacity-70" /> Added {formatAdded(instance.createdAt)}
          </span>
          <span className="flex items-center gap-1 text-xs font-medium text-primary transition-transform duration-200 group-hover:translate-x-0.5">
            Open <ArrowRight size={13} />
          </span>
        </div>
      </button>
    </div>
  )
}

function EmptyState({ onAdd }: { onAdd: () => void }) {
  return (
    <div className="fs-fade-up flex flex-col items-center justify-center rounded-2xl border border-dashed border-border bg-card/40 py-20 text-center">
      <span className="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10 text-primary">
        <Server size={28} />
      </span>
      <h3 className="mt-4 text-lg font-medium">No instances yet</h3>
      <p className="mt-1 max-w-sm text-sm text-muted-foreground">
        Connect your first UTMStack instance to start managing it from one place.
      </p>
      <Button onClick={onAdd} className="mt-5">
        <Plus size={16} className="mr-1.5" /> Connect instance
      </Button>
    </div>
  )
}

function AddInstanceModal({ onClose }: { onClose: () => void }) {
  const { create } = useInstanceMutations()
  const [name, setName] = useState('')
  const [baseUrl, setBaseUrl] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [tlsSkipVerify, setTlsSkipVerify] = useState(false)

  const err = create.error instanceof ApiError ? create.error.message : null

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    const input: FederationInstanceInput = {
      name: name.trim(),
      baseUrl: baseUrl.trim(),
      apiKey,
      tlsSkipVerify,
    }
    create.mutate(input, { onSuccess: onClose })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-background/70 backdrop-blur-sm" onClick={onClose} />
      <div className="fs-fade-up relative z-10 w-full max-w-md rounded-xl border border-border bg-card p-6 shadow-2xl">
        <div className="mb-5 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Connect an instance</h2>
          <button
            onClick={onClose}
            aria-label="Close"
            className="rounded p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          >
            <X size={18} />
          </button>
        </div>
        <form onSubmit={submit} className="space-y-4">
          <Field label="Name">
            <Input autoFocus value={name} onChange={(e) => setName(e.target.value)} placeholder="Client A" />
          </Field>
          <Field label="Base URL">
            <Input
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
              placeholder="https://client-a.example.com"
            />
          </Field>
          <Field label="Admin API key">
            <PasswordInput
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="Utm-Api-Key value"
            />
          </Field>
          <label className="flex items-center gap-2 text-xs text-muted-foreground">
            <input
              type="checkbox"
              checked={tlsSkipVerify}
              onChange={(e) => setTlsSkipVerify(e.target.checked)}
            />
            Skip TLS verification (self-signed certificates)
          </label>
          {err && <p className="text-xs text-destructive">{err}</p>}
          <p className="text-[11px] text-muted-foreground">
            The instance must be enterprise and the key valid — the connection is verified before
            saving.
          </p>
          <div className="flex justify-end gap-2 pt-1">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={create.isPending || !name.trim() || !baseUrl.trim() || !apiKey}
            >
              {create.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />} Connect
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1.5">
      <label className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
        {label}
      </label>
      {children}
    </div>
  )
}
