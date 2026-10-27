import { useEffect, useState, useMemo, useRef } from 'react'
import { Loader2, Search } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import { Input } from '@/shared/components/ui/input'
import { useTranslation } from 'react-i18next'
import type {
  CustomFilter,
  FilterBarLabels,
  FilterFieldDef,
  FilterOpDef,
  FilterValue,
} from './custom-filter.types'

const SELECT_CLS =
  'h-9 cursor-pointer rounded-md border border-input bg-popover px-2 text-sm text-popover-foreground transition-colors focus-visible:border-ring focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring'

export function FilterEditorPanel({
  initial,
  fields,
  operators,
  fetchValues,
  labels,
  onSubmit,
  onCancel,
}: {
  initial?: CustomFilter
  fields: FilterFieldDef[]
  operators: FilterOpDef[]
  fetchValues?: (field: string) => Promise<FilterValue[]>
  labels: FilterBarLabels
  onSubmit: (f: CustomFilter) => void
  onCancel: () => void
}) {
  const [field, setField] = useState(initial?.field ?? fields[0]?.field ?? '')
  const [operator, setOperator] = useState(initial?.operator ?? operators[0]?.id ?? 'IS')
  const [freeValue, setFreeValue] = useState(initial?.value ?? '')
  const [values, setValues] = useState<FilterValue[]>([])
  const [loadingValues, setLoadingValues] = useState(false)
  const [vq, setVq] = useState('')
  const [openFields,setOpenFields]= useState(false)

  const {t} = useTranslation()

  const [fq, setFq] = useState(t(initial?.label ?? '') )


  const op = operators.find((o) => o.id === operator)
  const needsValue = op?.needsValue ?? true

  useEffect(() => {
    if (!needsValue || !fetchValues) return
    setLoadingValues(true)
    setValues([])
    fetchValues(field)
      .then(setValues)
      .catch(() => setValues([]))
      .finally(() => setLoadingValues(false))
  }, [field, needsValue, fetchValues])

  const submit = (value: string) => {
    const fdef = fields.find((f) => f.field === field)
    if (!fdef) return
    onSubmit({ field, label: fdef.label, operator, value })
  }

  const filtered = values.filter((v) => (vq ? v.value.toLowerCase().includes(vq.toLowerCase()) : true))

  const filteredFields = useMemo(()=>fields.filter((f)=>t(f.label).includes(fq)),[fields,fq])

  const fieldSelectorRef = useRef<HTMLUListElement>(null)

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (fieldSelectorRef.current && !fieldSelectorRef.current.contains(event.target as Node)) {
        setOpenFields(false);
      }
    }

    if (openFields) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [openFields]);

  return (
    <div className="w-80 rounded-md border border-border bg-popover p-3 shadow-lg">
      <div className="space-y-2">
        <div className="relative">
          <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
          <Input
            value={fq}
            onClick={() =>setOpenFields(true)}
            onChange={(e:any) =>{ setFq(e.target.value) ;setOpenFields(true)}}
            className="h-8 pl-8 text-xs"
            autoFocus
          />
        </div>

         {openFields && filteredFields.length>0  && (
            <ul ref={fieldSelectorRef} className="absolute z-50 mt-1 w-11/12 max-h-48 overflow-auto rounded-md border-none bg-background shadow-lg">
              {filteredFields.map((f) => (
                <li
                  key={f.field}
                  onMouseDown={(e:any) => {
                    e.preventDefault();
                    e.stopPropagation()
                    setField(f.field);
                    setFq(t(f.label));
                    setOpenFields(false);
                  }}
                  className={cn(
                    "cursor-pointer px-3 py-1.5 text-xs hover:bg-accent ",
                    field === f.field && "bg-accent"
                  )}
                >
                  {t(f.label)}
                </li>
              ))}
            </ul>
          )}

        <select value={field} onChange={(e:any) => {
          const label = e.target.options[e.target.selectedIndex].text;
          setField(e.target.value.field)
          setFq(label)
        }} className={cn(SELECT_CLS, 'w-full p-0 m-0 h-0')}>
          {fields.filter((f)=>f.label.includes(fq)).map((f) => (
            <option key={f.field} value={f.field}>
              {t(f.label)}
            </option>
          ))}
        </select>


        <select value={operator} onChange={(e:any) => setOperator(e.target.value)} className={cn(SELECT_CLS, 'w-full')}>
          {operators.map((o) => (
            <option key={o.id} value={o.id}>
              {t(o.label)}
            </option>
          ))}
        </select>

        {needsValue ? (
          fetchValues ? (
            <>
              <div className="relative">
                <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground" />
                <Input
                  value={vq}
                  onChange={(e:any) => setVq(e.target.value)}
                  placeholder={t(labels.filterValues)}
                  className="h-8 pl-8 text-xs"
                  autoFocus
                />
              </div>
              <div className="max-h-48 overflow-y-auto rounded-md border border-border">
                {loadingValues ? (
                  <div className="flex items-center gap-1.5 px-3 py-3 text-xs text-muted-foreground">
                    <Loader2 className="h-3.5 w-3.5 animate-spin" /> {t(labels.loadingValues)}
                  </div>
                ) : filtered.length === 0 ? (
                  <div className="px-3 py-3 text-xs text-muted-foreground">{t(labels.noValues)}</div>
                ) : (
                  filtered.map((v) => {
                    const selected = initial?.value === v.value
                    return (
                      <button
                        key={v.value}
                        onClick={() => submit(v.value)}
                        className={cn(
                          'flex w-full items-center justify-between gap-2 px-3 py-1.5 text-left text-xs hover:bg-muted',
                          selected && 'bg-primary/10',
                        )}
                      >
                        <span className="truncate font-mono">{v.value || labels.empty}</span>
                        {v.count != null && (
                          <span className="shrink-0 font-mono text-[10px] text-muted-foreground">
                            {t(v.count.toLocaleString())}
                          </span>
                        )}
                      </button>
                    )
                  })
                )}
              </div>
              <p className="text-[10px] text-muted-foreground">{t(labels.pickValue)}</p>
            </>
          ) : (
            <>
              <Input
                value={freeValue}
                onChange={(e:any) => setFreeValue(e.target.value)}
                placeholder={labels.filterValues}
                className="h-8 text-xs"
                autoFocus
                onKeyDown={(e:any) => e.key === 'Enter' && freeValue && submit(freeValue)}
              />
              <div className="flex justify-end gap-2 pt-1">
                <Button variant="outline" size="sm" onClick={onCancel}>
                  {t(labels.cancel)}
                </Button>
                <Button size="sm" disabled={!freeValue} onClick={() => submit(freeValue)}>
                  {t(labels.addBtn)}
                </Button>
              </div>
            </>
          )
        ) : (
          <div className="flex justify-end gap-2 pt-1">
            <Button variant="outline" size="sm" onClick={onCancel}>
              {t(labels.cancel)}
            </Button>
            <Button size="sm" onClick={() => submit('')}>
              {t(labels.addBtn)}
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
