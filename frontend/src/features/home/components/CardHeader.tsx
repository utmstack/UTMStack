import { Link } from 'react-router-dom'
import { type LucideIcon } from 'lucide-react'

export function CardHeader({
  title,
  icon: Icon,
  iconClass,
  action,
}: {
  title: string
  icon?: LucideIcon
  iconClass?: string
  action?: { label: string; href: string }
}) {
  return (
    <div className="flex items-center justify-between px-6 pt-5 pb-4">
      <h3 className="flex items-center gap-2 text-sm font-semibold">
        {Icon && <Icon size={16} strokeWidth={1.75} className={iconClass ?? 'text-muted-foreground'} />}
        {title}
      </h3>
      {action && (
        <Link to={action.href} className="text-xs font-medium text-primary hover:underline">
          {action.label}
        </Link>
      )}
    </div>
  )
}
