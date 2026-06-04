import { useMemo, useState } from 'react'
import {
  AlertTriangle,
  Building2,
  CheckCircle2,
  Clock,
  Layers,
  Pencil,
  Plus,
  RefreshCw,
  Star,
  Trash2,
  X,
} from 'lucide-react'
import { toast } from 'sonner'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useWorkspaces, workspaceHttpService } from '@/features/workspace'
import type {
  CreateWorkspaceInput,
  UpdateWorkspaceInput,
  Workspace,
} from '@/features/workspace'

/* ─── Page ─────────────────────────────────────────────────────────────── */

export function WorkspacesPage() {
  // Reads come straight from the global context so the topbar selector and
  // this page never drift. Mutations call the http service directly and then
  // trigger `refresh()` to re-pull the list for everyone.
  const { workspaces, isLoading, refresh } = useWorkspaces()
  const [createOpen, setCreateOpen] = useState(false)
  const [editing, setEditing] = useState<Workspace | null>(null)
  const [deleting, setDeleting] = useState<Workspace | null>(null)

  const stats = useMemo(() => {
    const total = workspaces.length
    const defaultWs = workspaces.find((w) => w.is_default)
    const last30dCutoff = Date.now() - 30 * 86_400_000
    const recent = workspaces.filter((w) => new Date(w.created_at).getTime() > last30dCutoff).length
    return { total, defaultWs, recent }
  }, [workspaces])

  return (
    <div className="mx-auto w-full max-w-[1600px] px-6 py-6">
      <Header
        onCreate={() => setCreateOpen(true)}
        onRefresh={refresh}
        loading={isLoading}
      />

      <div className="mt-6">
        <StatsStrip total={stats.total} defaultName={stats.defaultWs?.name} recent={stats.recent} />
      </div>

      <div className="mt-6">
        <WorkspacesSection
          items={workspaces}
          loading={isLoading}
          onEdit={setEditing}
          onDelete={setDeleting}
          onCreate={() => setCreateOpen(true)}
        />
      </div>

      {createOpen && (
        <CreateDialog onClose={() => setCreateOpen(false)} onSaved={refresh} />
      )}
      {editing && (
        <EditDrawer
          workspace={editing}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null)
            refresh()
          }}
        />
      )}
      {deleting && (
        <DeleteDialog
          workspace={deleting}
          onClose={() => setDeleting(null)}
          onConfirmed={() => {
            setDeleting(null)
            refresh()
          }}
        />
      )}
    </div>
  )
}

/* ─── Header ───────────────────────────────────────────────────────────── */

function Header({
  onCreate,
  onRefresh,
  loading,
}: {
  onCreate: () => void
  onRefresh: () => void
  loading: boolean
}) {
  return (
    <header className="flex items-end justify-between gap-3">
      <div>
        <h1 className="flex items-center gap-2 text-xl font-semibold">
          <Layers size={18} strokeWidth={1.75} />
          Workspaces
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Logical tenants. Every collector, agent, integration, alert, and rule is owned by a
          workspace. Multi-workspace deployments are an Enterprise feature.
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading}>
          <RefreshCw size={14} className={cn('mr-2', loading && 'animate-spin')} />
          Refresh
        </Button>
        <Button size="sm" onClick={onCreate}>
          <Plus size={14} className="mr-2" />
          Create workspace
        </Button>
      </div>
    </header>
  )
}

/* ─── Stats strip ──────────────────────────────────────────────────────── */

function StatsStrip({
  total,
  defaultName,
  recent,
}: {
  total: number
  defaultName?: string
  recent: number
}) {
  return (
    <section className="rounded-xl border border-border bg-card">
      <div className="grid grid-cols-1 divide-y divide-border sm:grid-cols-3 sm:divide-x sm:divide-y-0">
        <StripStat
          label="Total workspaces"
          value={<span className="font-mono">{total}</span>}
          sub={total === 1 ? 'Single-tenant install' : 'Multi-tenant'}
        />
        <StripStat
          label="Default"
          value={
            <span className="inline-flex items-center gap-1.5">
              <Star size={15} strokeWidth={2} className="text-amber-500" />
              {defaultName ?? '—'}
            </span>
          }
          sub="Owns resources by default"
        />
        <StripStat
          label="Created last 30d"
          value={<span className="font-mono">{recent}</span>}
          sub="New workspaces"
        />
      </div>
    </section>
  )
}

function StripStat({
  label,
  value,
  sub,
}: {
  label: string
  value: React.ReactNode
  sub: string
}) {
  return (
    <div className="px-5 py-4">
      <div className="text-[11px] uppercase tracking-wider text-muted-foreground">{label}</div>
      <div className="mt-1 text-xl font-semibold">{value}</div>
      <div className="mt-0.5 text-[11px] text-muted-foreground">{sub}</div>
    </div>
  )
}

/* ─── Workspaces section ───────────────────────────────────────────────── */

function WorkspacesSection({
  items,
  loading,
  onEdit,
  onDelete,
  onCreate,
}: {
  items: Workspace[]
  loading: boolean
  onEdit: (w: Workspace) => void
  onDelete: (w: Workspace) => void
  onCreate: () => void
}) {
  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card">
      <header className="flex items-center justify-between border-b border-border px-5 py-3">
        <div>
          <h2 className="text-sm font-semibold">All workspaces</h2>
          <p className="mt-0.5 text-xs text-muted-foreground">
            Click a row to edit. The default workspace cannot be deleted.
          </p>
        </div>
      </header>

      <div className="grid grid-cols-[1fr_180px_140px_220px_60px] gap-3 border-b border-border bg-muted/30 px-5 py-2.5 text-[11px] uppercase tracking-wider text-muted-foreground">
        <div>Name</div>
        <div>Slug</div>
        <div>Created</div>
        <div>Description</div>
        <div className="text-right" />
      </div>

      {loading && items.length === 0 ? (
        <div className="px-5 py-16 text-center text-sm text-muted-foreground">Loading…</div>
      ) : items.length === 0 ? (
        <div className="px-5 py-16 text-center">
          <Building2 size={28} strokeWidth={1.5} className="mx-auto mb-3 text-muted-foreground/60" />
          <div className="text-sm font-medium">No workspaces yet</div>
          <div className="mt-1 text-xs text-muted-foreground">
            Create your first workspace to start scoping resources.
          </div>
          <Button size="sm" className="mt-4" onClick={onCreate}>
            <Plus size={14} className="mr-2" />
            Create workspace
          </Button>
        </div>
      ) : (
        items.map((w) => (
          <WorkspaceRow key={w.id} workspace={w} onEdit={onEdit} onDelete={onDelete} />
        ))
      )}
    </section>
  )
}

function WorkspaceRow({
  workspace,
  onEdit,
  onDelete,
}: {
  workspace: Workspace
  onEdit: (w: Workspace) => void
  onDelete: (w: Workspace) => void
}) {
  return (
    <div
      className="grid grid-cols-[1fr_180px_140px_220px_60px] gap-3 border-b border-border px-5 py-3 text-xs transition-colors last:border-b-0 hover:bg-muted/30"
      onClick={() => onEdit(workspace)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault()
          onEdit(workspace)
        }
      }}
    >
      <div className="flex min-w-0 items-center gap-2">
        <Building2 size={14} className="shrink-0 text-muted-foreground" />
        <span className="truncate text-sm font-medium">{workspace.name}</span>
        {workspace.is_default && (
          <span className="inline-flex items-center gap-1 rounded-md bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium text-amber-600 ring-1 ring-inset ring-amber-500/30 dark:text-amber-300">
            <Star size={9} strokeWidth={3} />
            Default
          </span>
        )}
      </div>
      <div className="truncate font-mono text-[11px] text-muted-foreground">{workspace.slug}</div>
      <div className="truncate text-[11px] text-muted-foreground">
        {formatDate(workspace.created_at)}
      </div>
      <div className="truncate text-[11px] text-muted-foreground">
        {workspace.description || <span className="italic">No description</span>}
      </div>
      <div className="flex items-center justify-end gap-1" onClick={(e) => e.stopPropagation()}>
        <Button
          variant="ghost"
          size="icon"
          className="h-7 w-7"
          onClick={() => onEdit(workspace)}
          title="Edit"
        >
          <Pencil size={13} />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          className={cn(
            'h-7 w-7',
            workspace.is_default
              ? 'cursor-not-allowed text-muted-foreground/40'
              : 'text-red-500 hover:bg-red-500/10 hover:text-red-500',
          )}
          onClick={() => !workspace.is_default && onDelete(workspace)}
          disabled={workspace.is_default}
          title={workspace.is_default ? 'Cannot delete default workspace' : 'Delete'}
        >
          <Trash2 size={13} />
        </Button>
      </div>
    </div>
  )
}

/* ─── Create dialog ────────────────────────────────────────────────────── */

function CreateDialog({ onClose, onSaved }: { onClose: () => void; onSaved: () => void }) {
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [slugTouched, setSlugTouched] = useState(false)
  const [description, setDescription] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const derivedSlug = useMemo(() => deriveSlug(name), [name])
  const effectiveSlug = slugTouched ? slug : derivedSlug

  const onNameChange = (v: string) => {
    setName(v)
  }

  const onSubmit = async () => {
    const payload: CreateWorkspaceInput = {
      name: name.trim(),
      description: description.trim() || undefined,
    }
    if (slugTouched && slug.trim()) {
      payload.slug = slug.trim()
    }
    setSubmitting(true)
    try {
      const created = await workspaceHttpService.create(payload)
      toast.success(`Workspace "${created.name}" created`)
      onSaved()
      onClose()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not create workspace')
    } finally {
      setSubmitting(false)
    }
  }

  const valid = name.trim().length >= 2

  return (
    <ModalShell title="Create workspace" onClose={onClose}>
      <div className="space-y-4">
        <Field label="Name" hint="Human-readable name shown in the workspace selector.">
          <Input
            value={name}
            onChange={(e) => onNameChange(e.target.value)}
            placeholder="e.g. Acme Corp"
            autoFocus
            className="h-9 text-sm"
          />
        </Field>
        <Field
          label="Slug"
          hint="URL-safe identifier (lowercase letters, digits, hyphens). Auto-derived from the name."
        >
          <Input
            value={effectiveSlug}
            onChange={(e) => {
              setSlug(e.target.value)
              setSlugTouched(true)
            }}
            placeholder="auto"
            className="h-9 font-mono text-xs"
          />
        </Field>
        <Field label="Description" hint="Optional context — shown in the workspaces list.">
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            placeholder="What is this workspace for?"
            className="w-full rounded-md border border-border bg-background/40 px-2.5 py-2 text-sm focus:bg-card focus:outline-none focus:ring-1 focus:ring-ring"
          />
        </Field>
      </div>
      <div className="mt-6 flex items-center justify-end gap-2 border-t border-border pt-4">
        <Button variant="outline" size="sm" onClick={onClose} disabled={submitting}>
          Cancel
        </Button>
        <Button size="sm" onClick={onSubmit} disabled={!valid || submitting}>
          {submitting ? 'Creating…' : 'Create workspace'}
        </Button>
      </div>
    </ModalShell>
  )
}

/* ─── Edit drawer ──────────────────────────────────────────────────────── */

function EditDrawer({
  workspace,
  onClose,
  onSaved,
}: {
  workspace: Workspace
  onClose: () => void
  onSaved: () => void
}) {
  const [name, setName] = useState(workspace.name)
  const [description, setDescription] = useState(workspace.description ?? '')
  const [submitting, setSubmitting] = useState(false)

  const dirty = name.trim() !== workspace.name || description.trim() !== (workspace.description ?? '')
  const valid = name.trim().length >= 2

  const onSubmit = async () => {
    const payload: UpdateWorkspaceInput = {}
    if (name.trim() !== workspace.name) payload.name = name.trim()
    if (description.trim() !== (workspace.description ?? '')) payload.description = description.trim()
    setSubmitting(true)
    try {
      await workspaceHttpService.update(workspace.id, payload)
      toast.success('Workspace updated')
      onSaved()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not update workspace')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/50 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-[640px] flex-col overflow-hidden border-l border-border bg-card shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-3 border-b border-border px-6 py-4">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <Building2 size={16} className="text-muted-foreground" />
              <h2 className="truncate text-base font-semibold">{workspace.name}</h2>
              {workspace.is_default && (
                <span className="inline-flex items-center gap-1 rounded-md bg-amber-500/15 px-1.5 py-0.5 text-[10px] font-medium text-amber-600 ring-1 ring-inset ring-amber-500/30 dark:text-amber-300">
                  <Star size={9} strokeWidth={3} />
                  Default
                </span>
              )}
            </div>
            <div className="mt-1 flex items-center gap-3 text-[11px] text-muted-foreground">
              <span className="font-mono">{workspace.slug}</span>
              <span className="inline-flex items-center gap-1">
                <Clock size={11} />
                Created {formatDate(workspace.created_at)}
                {workspace.created_by && ` · by ${workspace.created_by}`}
              </span>
            </div>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X size={16} />
          </Button>
        </header>

        <div className="flex-1 overflow-y-auto px-6 py-5">
          <div className="space-y-4">
            <Field label="Name">
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="h-9 text-sm"
              />
            </Field>
            <Field label="Slug" hint="Slug is immutable — create a new workspace if you need a different one.">
              <Input
                value={workspace.slug}
                disabled
                className="h-9 cursor-not-allowed font-mono text-xs"
              />
            </Field>
            <Field label="Description">
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={4}
                className="w-full rounded-md border border-border bg-background/40 px-2.5 py-2 text-sm focus:bg-card focus:outline-none focus:ring-1 focus:ring-ring"
              />
            </Field>
            <Field label="Last updated">
              <div className="text-xs text-muted-foreground">{formatDate(workspace.updated_at)}</div>
            </Field>
          </div>
        </div>

        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-4">
          <Button variant="outline" size="sm" onClick={onClose} disabled={submitting}>
            Cancel
          </Button>
          <Button size="sm" onClick={onSubmit} disabled={!dirty || !valid || submitting}>
            {submitting ? 'Saving…' : 'Save changes'}
          </Button>
        </footer>
      </div>
    </div>
  )
}

/* ─── Delete dialog ────────────────────────────────────────────────────── */

function DeleteDialog({
  workspace,
  onClose,
  onConfirmed,
}: {
  workspace: Workspace
  onClose: () => void
  onConfirmed: () => void
}) {
  const [confirmText, setConfirmText] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const matches = confirmText.trim() === workspace.slug

  const onSubmit = async () => {
    setSubmitting(true)
    try {
      await workspaceHttpService.remove(workspace.id)
      toast.success(`Workspace "${workspace.name}" deleted`)
      onConfirmed()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Could not delete workspace')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <ModalShell title="Delete workspace" onClose={onClose} tone="danger">
      <div className="flex items-start gap-3 rounded-md border border-red-500/30 bg-red-500/10 px-3 py-2.5 text-xs text-red-600 dark:text-red-300">
        <AlertTriangle size={16} className="mt-0.5 shrink-0" />
        <div>
          <div className="font-medium">This action cannot be undone.</div>
          <p className="mt-1 text-[11px] text-red-700/80 dark:text-red-300/80">
            Deleting <span className="font-mono">{workspace.slug}</span> will eventually drop every
            collector, agent, integration, and rule scoped to it. This release does not yet support
            migrating resources to another workspace before deletion.
          </p>
        </div>
      </div>

      <div className="mt-4">
        <label className="mb-1.5 block text-[11px] uppercase tracking-wider text-muted-foreground">
          Type <span className="font-mono">{workspace.slug}</span> to confirm
        </label>
        <Input
          value={confirmText}
          onChange={(e) => setConfirmText(e.target.value)}
          placeholder={workspace.slug}
          className="h-9 font-mono text-xs"
          autoFocus
        />
      </div>

      <div className="mt-6 flex items-center justify-end gap-2 border-t border-border pt-4">
        <Button variant="outline" size="sm" onClick={onClose} disabled={submitting}>
          Cancel
        </Button>
        <Button
          size="sm"
          onClick={onSubmit}
          disabled={!matches || submitting}
          className="bg-red-600 text-white hover:bg-red-700"
        >
          {submitting ? 'Deleting…' : 'Delete workspace'}
        </Button>
      </div>
    </ModalShell>
  )
}

/* ─── Shared primitives ────────────────────────────────────────────────── */

function ModalShell({
  title,
  onClose,
  tone = 'default',
  children,
}: {
  title: string
  onClose: () => void
  tone?: 'default' | 'danger'
  children: React.ReactNode
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
      onClick={onClose}
    >
      <div
        className="w-full max-w-[520px] rounded-xl border border-border bg-card shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between border-b border-border px-5 py-3">
          <h2
            className={cn(
              'flex items-center gap-2 text-base font-semibold',
              tone === 'danger' && 'text-red-500 dark:text-red-300',
            )}
          >
            {tone === 'danger' ? (
              <Trash2 size={16} />
            ) : (
              <CheckCircle2 size={16} className="text-muted-foreground" />
            )}
            {title}
          </h2>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X size={16} />
          </Button>
        </header>
        <div className="px-5 py-4">{children}</div>
      </div>
    </div>
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
    <div>
      <label className="mb-1.5 block text-[11px] uppercase tracking-wider text-muted-foreground">
        {label}
      </label>
      {children}
      {hint && <p className="mt-1 text-[11px] text-muted-foreground">{hint}</p>}
    </div>
  )
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: '2-digit' })
}

// Mirror the backend's slug derivation so the user sees an accurate preview
// before submitting. Backend re-derives + validates — this is just UX.
function deriveSlug(name: string): string {
  let out = ''
  let prevHyphen = false
  for (const ch of name.toLowerCase()) {
    if ((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')) {
      out += ch
      prevHyphen = false
    } else if (!prevHyphen && out.length > 0) {
      out += '-'
      prevHyphen = true
    }
  }
  out = out.replace(/^-+|-+$/g, '')
  return out.length > 64 ? out.slice(0, 64).replace(/-+$/g, '') : out
}
