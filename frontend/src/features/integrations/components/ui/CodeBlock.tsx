import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Copy } from 'lucide-react'

export function CodeBlock({ code }: { code: string }) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  const handleCopy = () => {
    navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="relative">
      <pre className="overflow-x-auto rounded-md bg-muted/40 p-3 font-mono text-[11px] leading-relaxed">
        {code}
      </pre>
      <button
        title={copied ? t('integrations.codeBlock.copied') : t('integrations.codeBlock.copy')}
        onClick={handleCopy}
        className="absolute right-2 top-2 flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-card hover:text-foreground"
      >
        <Copy size={11} />
      </button>
    </div>
  )
}
