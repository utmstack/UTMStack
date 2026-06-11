import { useTranslation } from 'react-i18next'

export function SqlPreview({
  sql,
  rawMode,
  onToggleRaw,
  rawSql,
  onChangeRawSql,
}: {
  sql: string
  rawMode: boolean
  onToggleRaw: (next: boolean) => void
  rawSql: string
  onChangeRawSql: (next: string) => void
}) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between gap-2">
        <span className="text-xs font-medium text-foreground/80">
          {t('dashboards.editor.sqlPreview.title')}
        </span>
        <label className="flex items-center gap-1.5 text-xs text-muted-foreground">
          <input
            type="checkbox"
            checked={rawMode}
            onChange={(e) => onToggleRaw(e.target.checked)}
            className="h-3.5 w-3.5"
          />
          {t('dashboards.editor.sqlPreview.toggleRaw')}
        </label>
      </div>
      {rawMode ? (
        <textarea
          value={rawSql}
          onChange={(e) => onChangeRawSql(e.target.value)}
          spellCheck={false}
          rows={8}
          placeholder={t('dashboards.editor.sqlPreview.rawPlaceholder') ?? ''}
          className="w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-xs leading-relaxed shadow-sm focus:border-ring focus:outline-none focus:ring-1 focus:ring-ring"
        />
      ) : (
        <pre className="max-h-64 overflow-auto rounded-md border border-border bg-muted/40 px-3 py-2 font-mono text-xs leading-relaxed text-foreground/90">
          {sql || t('dashboards.editor.sqlPreview.empty')}
        </pre>
      )}
      <p className="text-[10px] text-muted-foreground">
        {t('dashboards.editor.sqlPreview.hint')}
      </p>
    </div>
  )
}
