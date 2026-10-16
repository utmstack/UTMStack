import { useCallback, useState, type SetStateAction } from 'react'

const EMPTY_COLUMNS: string[] = []

export type ColumnLayouts = Record<string, string[]>
export const columnScope = (dataset: string, dataType: string | null) => JSON.stringify([dataset, dataType])

/** Column layouts live in the user's persisted tab, independently for each scope. */
export function useColumnLayouts(dataset: string, dataType: string | null, initial: {
  dataset: string; dataType: string | null; columns: string[]; layouts?: ColumnLayouts
}) {
  const scope = columnScope(dataset, dataType)
  const [layouts, setLayouts] = useState<ColumnLayouts>(() => ({
    [columnScope(initial.dataset, initial.dataType)]: initial.columns,
    ...initial.layouts,
  }))
  const columns = layouts[scope] ?? EMPTY_COLUMNS
  const setScopedColumns = useCallback((key: string, update: SetStateAction<string[]>) => {
    setLayouts(current => ({ ...current, [key]: typeof update === 'function' ? update(current[key] ?? []) : update }))
  }, [])
  const setColumns = useCallback((update: SetStateAction<string[]>) => setScopedColumns(scope, update), [scope, setScopedColumns])
  return { columns, layouts, scope, setColumns, setScopedColumns }
}
