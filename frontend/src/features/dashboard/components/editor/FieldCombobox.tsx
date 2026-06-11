import { useId } from 'react'
import { Input } from '@/shared/components/ui/input'
import type { IndexPatternField } from '@/features/dashboard/types'

export function FieldCombobox({
  value,
  onChange,
  fields,
  placeholder,
  disabled,
}: {
  value: string
  onChange: (next: string) => void
  fields: IndexPatternField[]
  placeholder?: string
  disabled?: boolean
}) {
  const listId = useId()
  return (
    <>
      <Input
        list={listId}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled}
        className="h-9"
        autoComplete="off"
      />
      <datalist id={listId}>
        {fields.map((f) => (
          <option key={f.name} value={f.name}>
            {f.type}
          </option>
        ))}
      </datalist>
    </>
  )
}
