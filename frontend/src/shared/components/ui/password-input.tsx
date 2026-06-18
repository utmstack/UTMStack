import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Eye, EyeOff } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Input, type InputProps } from './input'

export type PasswordInputProps = Omit<InputProps, 'type'>

/** Text input for secrets with a show/hide toggle. */
export function PasswordInput({ className, ...props }: PasswordInputProps) {
  const { t } = useTranslation()
  const [show, setShow] = useState(false)
  return (
    <div className="relative">
      <Input type={show ? 'text' : 'password'} className={cn('pr-9', className)} {...props} />
      <button
        type="button"
        tabIndex={-1}
        onClick={() => setShow((s) => !s)}
        aria-label={t(show ? 'common.hidePassword' : 'common.showPassword', {
          defaultValue: show ? 'Hide password' : 'Show password',
        })}
        className="absolute right-1.5 top-1/2 flex h-7 w-7 -translate-y-1/2 items-center justify-center rounded text-muted-foreground transition-colors hover:text-foreground"
      >
        {show ? <EyeOff size={15} strokeWidth={1.75} /> : <Eye size={15} strokeWidth={1.75} />}
      </button>
    </div>
  )
}
