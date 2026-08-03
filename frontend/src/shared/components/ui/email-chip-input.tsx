import { useState } from 'react'
import { ChipInput } from './chip-input'

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export interface EmailChipInputProps {
  values: string[]
  onChange: (v: string[]) => void
  placeholder?: string
  invalidMessage?: string
}

export function EmailChipInput({ values, onChange, placeholder, invalidMessage = 'Invalid email' }: EmailChipInputProps) {
  const [error, setError] = useState<string | null>(null)
  return (
    <div className="space-y-1">
      <ChipInput
        values={values}
        onChange={onChange}
        placeholder={placeholder}
        validate={(v) => (EMAIL_RE.test(v) ? null : invalidMessage)}
        onInvalid={setError}
      />
      {error && <p className="text-[11px] text-red-500">{error}</p>}
    </div>
  )
}
