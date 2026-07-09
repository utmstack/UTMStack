export function IconBtn({ children, label }: { children: React.ReactNode; label: string }) {
  return (
    <button
      aria-label={label}
      title={label}
      className="flex h-7 w-7 items-center justify-center rounded-md hover:bg-muted hover:text-foreground"
    >
      {children}
    </button>
  )
}
