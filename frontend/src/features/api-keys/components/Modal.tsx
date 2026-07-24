import { X, type LucideIcon } from 'lucide-react'

export function Modal({
  title,
  icon: Icon,
  onClose,
  children,
}: {
  title: string
  icon: LucideIcon
  onClose: () => void
  children: React.ReactNode
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-lg flex-col overflow-hidden rounded-xl border border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between gap-4 border-b border-border px-6 py-4">
          <h2 className="flex items-center gap-2 text-base font-semibold">
            <Icon size={17} strokeWidth={1.75} />
            {title}
          </h2>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </header>
        {children}
      </div>
    </div>
  )
}
