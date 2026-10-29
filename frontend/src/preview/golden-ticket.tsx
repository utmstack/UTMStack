// Standalone preview for the new Live Operations "Golden Ticket" scene.
// Not part of the app routes: it renders the real Stage + scene from a synthetic
// alert so you can see the animation without a backend. Open /preview.html.
import { useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { useTranslation } from 'react-i18next'
import '@/app/styles/index.css'
import '@/features/liveops/liveops.css'
import i18n, { SUPPORTED_LANGUAGES } from '@/shared/i18n'
import { ThemeProvider } from '@/app/providers/ThemeProvider'
import { Stage } from '@/features/liveops/components/Stage'
import { buildScene, type Party } from '@/features/liveops/model'
import type { Alert } from '@/features/alerts/types/alert.types'

// A compromised workstation forging tickets at the domain controller —
// the canonical golden-ticket shape. Mirrors the wineventlog example alert.
const ALERT: Alert = {
  id: 'preview-golden-ticket',
  '@timestamp': new Date().toISOString(),
  name: 'Golden Ticket Attack Detection',
  category: 'Credential Access',
  severity: 'high',
  status: 'Open',
  technique: 'T1558.001 - Steal or Forge Kerberos Tickets: Golden Ticket',
  dataSource: 'SRV-DJ-61',
  dataType: 'wineventlog',
  adversary: { host: 'wks-0457', ip: '192.168.67.85', user: 'gsmith', domain: 'northwind1.gob.ec' },
  target: { host: 'dc01', domain: 'northwind1.gob.ec', user: 'gsmith' },
  events: [],
}

export function GoldenTicketPreview() {
  const { t } = useTranslation()
  const [run, setRun] = useState(0)
  const [party, setParty] = useState<Party | null>(null)
  const [lang, setLang] = useState('en')
  const scene = buildScene(ALERT, t)

  useEffect(() => {
    if (lang !== i18n.language) void i18n.changeLanguage(lang)
  }, [lang])

  const area = { top: 56, bottom: 64, left: 64, right: 64 }

  return (
    <div className="relative h-screen w-screen overflow-hidden bg-background text-foreground">
      <Stage
        key={run}
        alert={ALERT}
        scene={scene}
        templateLabel={t('liveOps.kb.goldenTicket')}
        area={area}
        selectedParty={party?.id}
        onSelectParty={setParty}
      />
      <div className="absolute left-4 top-4 z-40 flex items-center gap-2">
        <button
          onClick={() => setRun((r) => r + 1)}
          className="rounded-md border border-border bg-card px-3 py-1.5 text-sm font-medium hover:bg-accent"
        >
          ↻ Replay
        </button>
        <select
          value={lang}
          onChange={(e) => setLang(e.target.value)}
          className="rounded-md border border-border bg-card px-2 py-1.5 text-sm"
        >
          {SUPPORTED_LANGUAGES.map((l) => (
            <option key={l.code} value={l.code}>
              {l.label}
            </option>
          ))}
        </select>
      </div>
    </div>
  )
}

createRoot(document.getElementById('root')!).render(
  <ThemeProvider>
    <GoldenTicketPreview />
  </ThemeProvider>,
)
