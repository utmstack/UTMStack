import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, Copy, KeyRound, TriangleAlert } from 'lucide-react'
import { Button } from '@/shared/components/ui/button'
import { Modal } from './Modal'

export function RevealModal({ name, token, onClose }: { name: string; token: string; onClose: () => void }) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)
  const copy = () => {
    navigator.clipboard.writeText(token).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  }
  return (
    <Modal onClose={onClose} title={t('apiKeys.reveal.title')} icon={KeyRound}>
      <div className="space-y-4 px-6 py-5">
        <p className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/5 px-3 py-2 text-xs text-amber-700 dark:text-amber-300">
          <TriangleAlert size={14} className="mt-0.5 shrink-0" />
          {t('apiKeys.reveal.note', { name })}
        </p>
        <div className="flex items-stretch gap-2">
          <code className="min-w-0 flex-1 truncate rounded-md border border-border bg-muted/40 px-3 py-2 font-mono text-xs">
            {token}
          </code>
          <Button variant="outline" size="sm" onClick={copy}>
            {copied ? <Check size={14} className="text-emerald-500" /> : <Copy size={14} />}
            <span className="ml-1.5">{copied ? t('apiKeys.reveal.copied') : t('apiKeys.reveal.copy')}</span>
          </Button>
        </div>
      </div>
      <footer className="flex items-center justify-end border-t border-border px-6 py-3">
        <Button size="sm" onClick={onClose}>
          {t('apiKeys.reveal.done')}
        </Button>
      </footer>
    </Modal>
  )
}
