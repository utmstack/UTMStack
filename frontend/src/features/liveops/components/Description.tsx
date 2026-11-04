import { useState } from 'react'
import { cn } from '@/shared/lib/utils'

export function Description({ text }: { text: string }) {
  const [open, setOpen] = useState(false)
  const plain = text.replace(/<br\s*\/?>/gi, ' ').replace(/<[^>]+>/g, '').replace(/\s+/g, ' ').trim()
  const long = plain.length > 200
  return (
    <div className="mt-1.5 text-[12px] leading-relaxed text-foreground/80">
      <p className={cn(!open && long && 'line-clamp-4')}>{plain}</p>
      {long && (
        <button type="button" onClick={() => setOpen((o) => !o)} className="mt-0.5 text-[11.5px] font-medium text-primary hover:underline">
          {open ? 'Less' : 'More'}
        </button>
      )}
    </div>
  )
}
