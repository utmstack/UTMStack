import { useTranslation } from 'react-i18next'

export function SqlPreview({ sql }: { sql: string }) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col gap-2">
      <span className="text-xs font-medium text-foreground/80">
        {t('dashboards.editor.sqlPreview.title')}
      </span>
      <pre className="max-h-64 overflow-auto rounded-md border border-border bg-muted/40 px-3 py-2 font-mono text-xs leading-relaxed text-foreground/90">
        {sql || t('dashboards.editor.sqlPreview.empty')}
      </pre>
      <p className="text-[10px] text-muted-foreground">
        {t('dashboards.editor.sqlPreview.hint')}
      </p>
    </div>
  )
}
