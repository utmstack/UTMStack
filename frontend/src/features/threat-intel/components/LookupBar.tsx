import { useEffect, useState } from 'react'
import { Crosshair, Search } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useTiSearch } from '../hooks/use-ti-search'
import type { EntitySearchResponse } from '../domain/threat-intel.types'

interface LookupBarProps {
  onResults: (data: EntitySearchResponse | null) => void
}

export function LookupBar({ onResults }: LookupBarProps) {
  const [q, setQ] = useState('')
  const { mutate, isPending } = useTiSearch()


  const handleSubmit = () => {
    if (q.trim()) {
      mutate(
        { query: q },
        {
          onSuccess: (data) => {
            if (data?.kind === 'not-configured') {
              onResults(null)
            } else if (data?.kind === 'ok') {
              onResults(data.value)
            }
          },
        }
      )
    }
  }

  useEffect(()=>{
      mutate(
        { query: '*' },
        {
          onSuccess: (data) => {
            if (data?.kind === 'not-configured') {
              onResults(null)
            } else if (data?.kind === 'ok') {
              onResults(data.value)
            }
          },
        }
      )
  },[])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSubmit()
    }
  }

  return (
    <div className="rounded-xl border border-border bg-card p-4">
      <div className="flex items-center gap-2 text-[11px] uppercase tracking-wider text-muted-foreground">
        <Crosshair size={12} className="text-fuchsia-500" />
        Lookup any indicator
      </div>
      <div className="mt-2 flex items-center gap-2">
        <div className="relative flex-1">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={q}
            onChange={(e) => setQ(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Paste an IP, domain, hash, URL, or CVE — we'll check feeds and your environment"
            className="h-10 pl-9 font-mono text-sm"
          />
        </div>
        <Button onClick={handleSubmit} disabled={isPending}>
          <Crosshair size={13} className="mr-1.5" />
          {isPending ? 'Searching…' : 'Lookup'}
        </Button>
      </div>
    </div>
  )
}
