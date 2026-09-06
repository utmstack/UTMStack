import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Zap } from 'lucide-react'
import type { FlowCondition } from '../types/soar.types'
import { TriggerConditionsEditor } from './TriggerConditionsEditor'

interface Props {
  conditions: FlowCondition[]
  readOnly?: boolean
  onChange: (c: FlowCondition[]) => void
}

// Side panel that stands in for NodeInspector when the virtual trigger node is
// selected — hosts the alert-match conditions that used to live in a top-level
// SectionCard. Same chrome as NodeInspector so the canvas layout stays stable.
export function TriggerInspector({ conditions, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const [width, setWidth] = useState(520)

  const startResize = (e: React.MouseEvent) => {
    e.preventDefault()
    const startX = e.clientX
    const startW = width
    const onMove = (ev: MouseEvent) => {
      const next = startW + (startX - ev.clientX)
      setWidth(Math.max(280, Math.min(next, Math.min(900, window.innerWidth - 200))))
    }
    const onUp = () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
    }
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
  }

  return (
    <aside
      className="relative flex shrink-0 flex-col border-l border-border bg-card h-[100%] overflow-y-auto"
      style={{ width }}
    >
      <div
        onMouseDown={startResize}
        className="absolute left-0 top-0 z-10 h-full w-1 cursor-col-resize hover:bg-primary/40"
        title={t('soar.editor.canvas.dragToResize')}
      />
      <div className="flex items-center gap-2 border-b border-border px-3 py-2">
        <span className="flex h-6 w-6 items-center justify-center rounded bg-amber-500/15 text-amber-500">
          <Zap size={12} />
        </span>
        <div className="min-w-0">
          <div className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
            {t('soar.editor.when')}
          </div>
          <div className="truncate text-sm font-medium">{t('soar.editor.whenTitle')}</div>
        </div>
      </div>
      <div className="flex-1 space-y-3 overflow-y-auto p-3 text-xs">
        <TriggerConditionsEditor conditions={conditions} readOnly={readOnly} onChange={onChange} />
      </div>
    </aside>
  )
}
