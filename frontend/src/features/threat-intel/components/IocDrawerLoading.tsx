import { X } from 'lucide-react'

interface IocDrawerLoadingProps {
  onClose: () => void
}

export function IocDrawerLoading({ onClose }: IocDrawerLoadingProps) {
  return (
    <>
      <header className="flex items-center justify-between border-b border-border px-6 py-4">
        <span className="text-sm text-muted-foreground">Loading…</span>
        <button
          onClick={onClose}
          className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <X size={16} />
        </button>
      </header>
      <div className="flex-1 bg-muted/20" />
    </>
  )
}
