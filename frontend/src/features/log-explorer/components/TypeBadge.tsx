import { Braces, Calendar, Globe, Hash, Tag, ToggleLeft, Type, type LucideIcon } from 'lucide-react'

const TYPE_META: Record<string, { icon: LucideIcon; color: string; label: string }> = {
  date: { icon: Calendar, color: 'text-violet-500', label: 'Date' },
  keyword: { icon: Tag, color: 'text-sky-500', label: 'Keyword' },
  text: { icon: Type, color: 'text-emerald-500', label: 'Text' },
  ip: { icon: Globe, color: 'text-fuchsia-500', label: 'IP' },
  boolean: { icon: ToggleLeft, color: 'text-rose-500', label: 'Boolean' },
}
const NUMBER_TYPES = new Set(['long', 'integer', 'short', 'byte', 'double', 'float', 'half_float', 'scaled_float'])

function typeMeta(type: string) {
  if (TYPE_META[type]) return TYPE_META[type]
  if (NUMBER_TYPES.has(type)) return { icon: Hash, color: 'text-amber-500', label: 'Number' }
  return { icon: Braces, color: 'text-muted-foreground', label: type || 'object' }
}

export function TypeBadge({ type }: { type: string }) {
  const m = typeMeta(type)
  const Icon = m.icon
  return (
    <span
      className="flex h-5 w-5 shrink-0 items-center justify-center rounded bg-muted"
      title={m.label}
    >
      <Icon size={11} className={m.color} />
    </span>
  )
}
