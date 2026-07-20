import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import { Crosshair, Search } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'

const ENTITY_TYPES = ['ip', 'domain', 'hostname', 'url', 'md5', 'sha1', 'sha256', 'sha3-256', 'cve', 'email-address', 'threat', 'malware']

interface LookupBarProps {
  onSearch: (query: string) => void
  onLookup: (input: { type: string; value: string }) => void
  isPending?: boolean
  isLookupPending?: boolean
}

export function LookupBar({ onSearch, onLookup, isPending, isLookupPending }: LookupBarProps) {
  const { t } = useTranslation()
  const [q, setQ] = useState('')
  const [selectedType, setSelectedType] = useState('any')

  const handleSubmit = () => {
    const trimmed = q.trim()
    if (!trimmed) return
    if (selectedType === 'any') {
      onSearch(trimmed)
    } else {
      onLookup({ type: selectedType, value: trimmed })
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') handleSubmit()
  }

  const busy = isPending || isLookupPending

  return (
    <div className="rounded-xl border border-border bg-card p-4">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Crosshair size={12} className="text-fuchsia-500" />
        {t('threatIntel.lookup.title')}
      </div>
      <div className="mt-2 flex items-center gap-2">
        <select
          value={selectedType}
          onChange={(e) => setSelectedType(e.target.value)}
          className="h-10 rounded-md border border-input bg-background px-2 text-xs"
        >
          <option value="any">Any type</option>
          {ENTITY_TYPES.map((type) => (
            <option key={type} value={type}>{type}</option>
          ))}
        </select>
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
        <Button onClick={handleSubmit} disabled={busy}>
          <Crosshair size={13} className="mr-1.5" />
          {busy ? t('threatIntel.lookup.busy') : selectedType === 'any' ? t('threatIntel.lookup.button') : 'Lookup'}
        </Button>
      </div>
    </div>
  )
}
