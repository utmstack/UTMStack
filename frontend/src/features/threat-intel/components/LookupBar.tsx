import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import { Crosshair, Search } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

interface LookupBarProps {
  onSearch: (query: string) => void
  isPending?: boolean
}

export function LookupBar({ onSearch, isPending }: LookupBarProps) {
  const { t } = useTranslation()
  const [q, setQ] = useState('')

  const handleSubmit = () => {
    const trimmed = q.trim()
    if (trimmed) onSearch(trimmed)
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') handleSubmit()
  }

  return (
    <div className="rounded-xl border border-border bg-card p-4">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Crosshair size={12} className="text-fuchsia-500" />
        {t('threatIntel.lookup.title')}
      </div>
      <div className="mt-2 flex items-center gap-2">
        <div className="relative flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={t('threatIntel.lookup.placeholder')}
            className="h-10 pl-9 font-mono text-sm"
          />
        </div>
        <Button onClick={handleSubmit} disabled={isPending}>
          <Crosshair size={13} className="mr-1.5" />
          {isPending ? t('threatIntel.lookup.busy') : t('threatIntel.lookup.button')}
        </Button>
      </div>
    </div>
  )
}
