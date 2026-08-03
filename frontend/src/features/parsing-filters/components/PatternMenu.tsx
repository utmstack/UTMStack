import { useTranslation } from 'react-i18next'
import { Lock } from 'lucide-react'
import { Input } from '@/shared/components/ui/input'
import type { RegexPattern } from '@/features/regex-patterns/services/regex-patterns-http.service'

/**
 * The inner body of a regex-pattern picker: a searchable list of named patterns
 * and selection. The parent owns the positioned popover container; this renders
 * only its contents so it can be reused both from the visual grok field and the
 * YAML content editor.
 *
 * Patterns are a shared, read-only vocabulary seeded by the backend, so this
 * picker only selects — there is no authoring here.
 *
 * `query` is controlled by the parent. Pass `onQueryChange` to render a search
 * box; omit it when the query is driven externally (e.g. an open `{{.foo` token).
 */
export function PatternMenu({
  patternOptions,
  query,
  onQueryChange,
  onPick,
}: {
  patternOptions: RegexPattern[]
  query: string
  onQueryChange?: (q: string) => void
  onPick: (name: string) => void
}) {
  const { t } = useTranslation()

  const filtered = patternOptions
    .filter((p) => p.patternId.toLowerCase().includes(query.toLowerCase()))
    .slice(0, 50)

  return (
    <>
      {onQueryChange && (
        <div className="border-b border-border p-1.5">
          <Input
            autoFocus
            value={query}
            onChange={(e) => onQueryChange(e.target.value)}
            placeholder={t('parsingFilters.visual.patternSearch')}
            className="h-7 text-[12px]"
          />
        </div>
      )}
      <div className="max-h-[220px] overflow-y-auto py-1">
        {filtered.length === 0 ? (
          <div className="px-3 py-4 text-center text-[11px] text-muted-foreground">
            {t('parsingFilters.visual.patternNone')}
          </div>
        ) : (
          filtered.map((p) => (
            <button
              key={p.patternId}
              type="button"
              onClick={() => onPick(p.patternId)}
              className="flex w-full flex-col gap-0.5 px-3 py-1.5 text-left hover:bg-muted"
            >
              <span className="flex items-center gap-1.5">
                <span className="font-mono text-[12px] text-foreground">{`{{.${p.patternId}}}`}</span>
                {p.systemOwner ? (
                  <span className="inline-flex items-center gap-0.5 rounded bg-violet-500/15 px-1 py-0 text-[9px] font-medium text-violet-500">
                    <Lock size={8} />
                    {t('parsingFilters.system')}
                  </span>
                ) : (
                  <span className="rounded bg-sky-500/15 px-1 py-0 text-[9px] font-medium text-sky-500">
                    {t('parsingFilters.user')}
                  </span>
                )}
              </span>
              {p.patternDefinition && (
                <span className="truncate font-mono text-[10px] text-muted-foreground">{p.patternDefinition}</span>
              )}
            </button>
          ))
        )}
      </div>
    </>
  )
}
