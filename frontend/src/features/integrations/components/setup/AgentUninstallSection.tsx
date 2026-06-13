import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { CodeBlock } from '@/features/integrations/components/ui/CodeBlock'

// Collapsible uninstall block, closed by default — mirrors the Forwarder guide's
// uninstall section. The command is for the platform tab selected above.
export function AgentUninstallSection({
  command,
  lang = 'bash',
}: {
  command: string
  lang?: 'bash' | 'powershell'
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  return (
    <div className="overflow-hidden rounded-lg border border-border">
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center justify-between px-4 py-3 text-left text-sm font-medium hover:bg-muted/40 transition-colors"
      >
        <span>{t('integrations.setup.agent.uninstall.title')}</span>
        <ChevronDown
          size={14}
          className={cn('shrink-0 text-muted-foreground transition-transform duration-200', open && 'rotate-180')}
        />
      </button>
      {open && (
        <div className="space-y-3 border-t border-border px-4 pb-4 pt-3">
          <p className="text-sm text-foreground/90">{t('integrations.setup.agent.uninstall.body')}</p>
          <CodeBlock code={command} lang={lang} />
        </div>
      )}
    </div>
  )
}
