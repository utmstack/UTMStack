import { X, type LucideIcon } from 'lucide-react'

export function Modal({
  title,
  icon: Icon,
  subtitle,
  onClose,
  children,
  footer,
}: {
  title: string
  icon: LucideIcon
  subtitle?: string
  onClose: () => void
  children: React.ReactNode
  footer: React.ReactNode
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="flex w-full max-w-lg flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-start justify-between gap-4 border-b border-border px-6 py-4">
          <div>
            <h2 className="flex items-center gap-2 text-lg font-semibold">
              <Icon size={18} />
              {title}
            </h2>
            {subtitle && <p className="mt-1 text-xs text-muted-foreground">{subtitle}</p>}
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>
        <div className="space-y-4 px-6 py-5">{children}</div>
        <footer className="flex items-center justify-end gap-2 border-t border-border px-6 py-3">
          {footer}
        </footer>
      </div>
    </div>
  )
}
