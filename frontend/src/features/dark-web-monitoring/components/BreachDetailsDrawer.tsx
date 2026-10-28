import { useEffect, useRef, useState, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { AlertCircle, CalendarDays, ChevronDown, Copy, Database, Eye, EyeOff, FileText, Globe2, Info, Layers3, ShieldAlert, UserRound, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { FloatingPanel } from '@/shared/components/ui/floating-panel'
import type { BreachStatus, DarkWebRecord } from '../types'
import { STATUS_KEYS, escapeHtml, formatDateMedium, safeHtml, safeImageUrl, valueOrDash } from '../lib/format'
import { StatusPill } from './StatusPill'

interface BreachDetailsDrawerProps {
  record: DarkWebRecord
  onClose: () => void
  onStatusChange?: (status: BreachStatus) => void
}

type MitigationStepStatus = 'NOT_STARTED' | 'IN_PROGRESS' | 'COMPLETED'

const MITIGATION_STEP_KEYS = ['immediateNextSteps', 'requestDataRemoval', 'ceaseAndDesist', 'courtOrder'] as const
type MitigationStepKey = (typeof MITIGATION_STEP_KEYS)[number]
const MITIGATION_STEP_STATUSES: MitigationStepStatus[] = ['NOT_STARTED', 'IN_PROGRESS', 'COMPLETED']

// Completion is a success state: green for COMPLETED, neutral for NOT_STARTED.
const MITIGATION_STEP_STATUS_STYLES: Record<MitigationStepStatus, string> = {
  NOT_STARTED: 'bg-muted text-muted-foreground ring-border hover:bg-muted/80',
  IN_PROGRESS: 'bg-amber-500/10 text-amber-400 ring-amber-500/20 hover:bg-amber-500/20',
  COMPLETED: 'bg-green-500/10 text-green-400 ring-green-500/20 hover:bg-green-500/20',
}

const MITIGATION_STEP_IMAGES: Record<MitigationStepKey, string> = {
  immediateNextSteps:
    'data:image/svg+xml,%3Csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 80 80%27%3E%3Crect width=%2780%27 height=%2780%27 fill=%27%2310151d%27/%3E%3Crect x=%276%27 y=%278%27 width=%2714%27 height=%2714%27 rx=%273%27 fill=%27%23ef4444%27/%3E%3Crect x=%2760%27 y=%2710%27 width=%2714%27 height=%2714%27 rx=%273%27 fill=%27%2322c55e%27/%3E%3Crect x=%2710%27 y=%2756%27 width=%2714%27 height=%2714%27 rx=%273%27 fill=%27%23eab308%27/%3E%3Crect x=%2760%27 y=%2756%27 width=%2714%27 height=%2714%27 rx=%273%27 fill=%27%233b82f6%27/%3E%3Crect x=%2730%27 y=%2730%27 width=%2720%27 height=%2726%27 rx=%274%27 fill=%27%23ef4444%27/%3E%3Crect x=%2730%27 y=%2730%27 width=%2720%27 height=%2711%27 rx=%274%27 fill=%27%23f87171%27/%3E%3Cpath d=%27M34 30v-4a6 6 0 0 1 12 0v4%27 stroke=%27%23cbd5e1%27 stroke-width=%273%27 fill=%27none%27/%3E%3Ccircle cx=%2740%27 cy=%2743%27 r=%273.5%27 fill=%27%230f172a%27/%3E%3Crect x=%2738.7%27 y=%2745%27 width=%272.6%27 height=%276%27 fill=%27%230f172a%27/%3E%3C/svg%3E',
  requestDataRemoval:
    'data:image/svg+xml,%3Csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 80 80%27%3E%3Crect width=%2780%27 height=%2780%27 fill=%27%230d1219%27/%3E%3Ccircle cx=%2714%27 cy=%2714%27 r=%272%27 fill=%27%23ef4444%27/%3E%3Ccircle cx=%2730%27 cy=%279%27 r=%272%27 fill=%27%2322c55e%27/%3E%3Ccircle cx=%2748%27 cy=%2715%27 r=%272%27 fill=%27%23eab308%27/%3E%3Ccircle cx=%2764%27 cy=%2710%27 r=%272%27 fill=%27%233b82f6%27/%3E%3Ccircle cx=%2720%27 cy=%2726%27 r=%272%27 fill=%27%23a855f7%27/%3E%3Ccircle cx=%2740%27 cy=%2724%27 r=%272%27 fill=%27%23ef4444%27/%3E%3Ccircle cx=%2760%27 cy=%2726%27 r=%272%27 fill=%27%2322c55e%27/%3E%3Crect x=%274%27 y=%2766%27 width=%2772%27 height=%2710%27 rx=%273%27 fill=%27%231f2937%27/%3E%3Crect x=%2710%27 y=%2744%27 width=%2742%27 height=%2724%27 rx=%274%27 fill=%27%23374151%27/%3E%3Crect x=%2710%27 y=%2744%27 width=%2742%27 height=%276%27 rx=%273%27 fill=%27%234b5563%27/%3E%3Crect x=%2716%27 y=%2756%27 width=%278%27 height=%2710%27 rx=%272%27 fill=%27%23111827%27/%3E%3Crect x=%2738%27 y=%2756%27 width=%278%27 height=%2710%27 rx=%272%27 fill=%27%23111827%27/%3E%3Crect x=%2762%27 y=%2730%27 width=%2710%27 height=%2722%27 fill=%27%230b1220%27/%3E%3Crect x=%2746%27 y=%2732%27 width=%2720%27 height=%2710%27 rx=%272%27 fill=%27%2316a34a%27/%3E%3Crect x=%2746%27 y=%2732%27 width=%2720%27 height=%273%27 fill=%27%2322c55e%27/%3E%3Cpath d=%27M28 44v-6h12v6%27 fill=%27%230f172a%27/%3E%3C/svg%3E',
  ceaseAndDesist:
    'data:image/svg+xml,%3Csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 80 80%27%3E%3Crect width=%2780%27 height=%2780%27 fill=%27%230e131a%27/%3E%3Crect x=%2714%27 y=%2718%27 width=%2748%27 height=%2742%27 rx=%273%27 fill=%27%23f8fafc%27/%3E%3Crect x=%2714%27 y=%2718%27 width=%2748%27 height=%276%27 fill=%27%23e2e8f0%27/%3E%3Crect x=%2720%27 y=%2730%27 width=%2736%27 height=%272%27 fill=%27%2394a3b8%27/%3E%3Crect x=%2720%27 y=%2736%27 width=%2736%27 height=%272%27 fill=%27%23cbd5e1%27/%3E%3Crect x=%2720%27 y=%2742%27 width=%2724%27 height=%272%27 fill=%27%23cbd5e1%27/%3E%3Crect x=%2730%27 y=%2752%27 width=%2716%27 height=%273%27 rx=%271%27 fill=%27%23dc2626%27/%3E%3Cpath d=%27M52 46l12 12-2 6-14-14z%27 fill=%27%23f59e0b%27/%3E%3Cpath d=%27M64 58l2.6-2.6 3 3-2.6 2.6z%27 fill=%27%23111827%27/%3E%3Cpath d=%27M52 46l-1.5-5.5L62 43z%27 fill=%27%2378350f%27/%3E%3C/svg%3E',
  courtOrder:
    'data:image/svg+xml,%3Csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 80 80%27%3E%3Crect width=%2780%27 height=%2780%27 fill=%27%2311151c%27/%3E%3Crect x=%2714%27 y=%2748%27 width=%2752%27 height=%278%27 rx=%272%27 fill=%27%231f2937%27/%3E%3Cpath d=%27M40 48l-8 14h16z%27 fill=%27%23243041%27/%3E%3Cpath d=%27M32 14h20l-4 18h-12z%27 fill=%27%23d97706%27/%3E%3Cpath d=%27M32 14h8l-4 18h-4z%27 fill=%27%23f59e0b%27/%3E%3Ccircle cx=%2740%27 cy=%2736%27 r=%274%27 fill=%27%2392400e%27/%3E%3Crect x=%2737%27 y=%2738%27 width=%276%27 height=%2714%27 fill=%27%23b45309%27/%3E%3Ccircle cx=%2762%27 cy=%2720%27 r=%277%27 fill=%27%23dc2626%27/%3E%3Ccircle cx=%2762%27 cy=%2720%27 r=%274%27 fill=%27%230f172a%27/%3E%3C/svg%3E',
}

export function BreachDetailsDrawer({ record, onClose, onStatusChange }: BreachDetailsDrawerProps) {
  const { t, i18n } = useTranslation()
  const [activeTab, setActiveTab] = useState<'information' | 'echoes'>('information')
  const [showSecrets, setShowSecrets] = useState(false)
  const [stepStatuses, setStepStatuses] = useState<Record<MitigationStepKey, MitigationStepStatus>>({
    immediateNextSteps: 'NOT_STARTED',
    requestDataRemoval: 'NOT_STARTED',
    ceaseAndDesist: 'NOT_STARTED',
    courtOrder: 'NOT_STARTED',
  })

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key !== 'Escape') return
      // A dropdown inside the drawer (status pill, mitigation step) handles the
      // same Escape first; only close the drawer when nothing inner is open.
      if (document.querySelector('[data-floating-panel]')) return
      onClose()
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [onClose])

  const subject = record.organizationElement
  const subjectType = t(`darkWebMonitoring.exportDialogs.elementTypes.${record.organizationElementType}`)
  const inMitigation = record.breachStatus === 'IN_MITIGATION'
  // Only render a record-supplied image when it is a safe http(s)/data:image source.
  const evidenceUrl = safeImageUrl(record.imageUrl)
  // Interpolated record type, escaped once for the (sanitized) i18n markup below.
  const subjectTypeLabel = escapeHtml(subjectType.toLowerCase())
  const sensitiveValue = (value: string | null) =>
    value ? (showSecrets ? value : '••••••••') : '—'

  const copyId = async () => {
    try {
      await navigator.clipboard.writeText(record.id)
      toast.success(t('darkWebMonitoring.details.copied'))
    } catch {
      toast.error(t('darkWebMonitoring.details.copyFailed'))
    }
  }

  return createPortal(
    <div
      className="fixed inset-0 z-[90] flex justify-end bg-black/65"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) onClose()
      }}
    >
      <aside
        role="dialog"
        aria-modal="true"
        aria-labelledby="breach-details-title"
        className="flex h-full w-full max-w-3xl flex-col border-l border-border bg-card shadow-2xl"
      >
        <header className="flex shrink-0 items-center justify-between gap-4 border-b border-border px-5 py-4">
          <h2 id="breach-details-title" className="min-w-0 truncate font-mono text-sm font-semibold">
            {t('darkWebMonitoring.details.title', {
              status: t(`darkWebMonitoring.status.${STATUS_KEYS[record.breachStatus]}`),
              type: subjectType.toLowerCase(),
              value: subject,
            })}
          </h2>
          <button type="button" aria-label={t('darkWebMonitoring.details.close')} onClick={onClose} className="shrink-0 text-muted-foreground hover:text-foreground">
            <X size={16} />
          </button>
        </header>

        <div className="min-h-0 flex-1 overflow-y-auto px-5 py-4 font-mono text-xs">
          <div className="mb-5 flex flex-wrap items-center gap-2 border-b border-border">
            <button
              type="button"
              role="tab"
              aria-selected={activeTab === 'information'}
              onClick={() => setActiveTab('information')}
              className={`flex items-center gap-1.5 border-b-2 px-1 py-2 font-semibold ${activeTab === 'information' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground'}`}
            >
              <ShieldAlert size={13} />
              {t('darkWebMonitoring.details.information')}
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={activeTab === 'echoes'}
              onClick={() => setActiveTab('echoes')}
              className={`flex items-center gap-1.5 border-b-2 px-1 py-2 font-semibold ${activeTab === 'echoes' ? 'border-primary text-primary' : 'border-transparent text-muted-foreground'}`}
            >
              <Layers3 size={13} />
              {t('darkWebMonitoring.details.echoes')}
              <span className="rounded-full bg-muted px-2 py-0.5 text-[9px]">{record.echoesCount}</span>
            </button>
          </div>

          {activeTab === 'information' ? (
            <div role="tabpanel" className="space-y-5">
              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <div className="mb-1.5 text-[10px] text-muted-foreground">{t('darkWebMonitoring.details.id')}</div>
                  <div className="flex items-center gap-2">
                    <code className="max-w-full truncate rounded bg-muted px-2 py-1.5">{record.id}</code>
                    <button type="button" aria-label={t('darkWebMonitoring.details.copyId')} onClick={copyId} className="rounded p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground">
                      <Copy size={13} />
                    </button>
                  </div>
                </div>
                <div>
                  <div className="mb-1.5 text-[10px] text-muted-foreground">{t('darkWebMonitoring.table.status')}</div>
                  <StatusPill
                    status={record.breachStatus}
                    label={t(`darkWebMonitoring.status.${STATUS_KEYS[record.breachStatus]}`)}
                    onChange={(nextStatus) => onStatusChange?.(nextStatus)}
                  />
                </div>
              </div>

              {evidenceUrl && (
                <div className="overflow-hidden rounded-xl border border-border bg-muted">
                  <img src={evidenceUrl} alt={t('darkWebMonitoring.details.evidence')} className="max-h-80 w-full object-contain" />
                </div>
              )}

              <section>
                <h3 className="mb-2 flex items-center gap-1.5 font-semibold"><AlertCircle size={13} className="text-primary" />{t('darkWebMonitoring.details.description')}</h3>
                <p className="rounded-lg bg-background p-3 leading-relaxed text-foreground/90">
                  {valueOrDash(record.breachDescription)}
                </p>
              </section>

              <section>
                <div className="mb-2 flex items-center justify-between gap-3">
                  <h3 className="flex items-center gap-1.5 font-semibold"><ShieldAlert size={13} className="text-primary" />{t('darkWebMonitoring.details.compromisedData')}</h3>
                  <button type="button" onClick={() => setShowSecrets((current) => !current)} className="inline-flex items-center gap-1 text-[10px] text-muted-foreground hover:text-foreground">
                    {showSecrets ? <EyeOff size={12} /> : <Eye size={12} />}
                    {t(showSecrets ? 'darkWebMonitoring.details.hideSecrets' : 'darkWebMonitoring.details.revealSecrets')}
                  </button>
                </div>
                <div className="flex flex-wrap gap-1.5">
                  {(record.compromisedData ?? '').split(',').map((item) => item.trim()).filter(Boolean).map((item) => (
                    <span key={item} className="rounded-full bg-primary/10 px-2 py-1 text-[9px] text-primary">{item}</span>
                  ))}
                  {record.password && <span className="rounded-full bg-primary/10 px-2 py-1 text-[9px] text-primary">{t('darkWebMonitoring.table.password')}</span>}
                  {record.hashedPassword && <span className="rounded-full bg-primary/10 px-2 py-1 text-[9px] text-primary">{t('darkWebMonitoring.table.hashedPassword')}</span>}
                </div>
                <div className="mt-3 grid gap-3 sm:grid-cols-2">
                  {record.email && <DetailField label={t('darkWebMonitoring.table.email')} value={record.email} />}
                  {record.username && <DetailField label={t('darkWebMonitoring.table.username')} value={record.username} />}
                  {record.password && <DetailField label={t('darkWebMonitoring.table.password')} value={sensitiveValue(record.password)} />}
                  {record.hashedPassword && <DetailField label={t('darkWebMonitoring.table.hashedPassword')} value={sensitiveValue(record.hashedPassword)} />}
                  {record.phone && <DetailField label={t('darkWebMonitoring.table.phone')} value={record.phone} />}
                  {record.ipAddress && <DetailField label={t('darkWebMonitoring.table.ipAddress')} value={record.ipAddress} />}
                  {record.name && <DetailField label={t('darkWebMonitoring.table.name')} value={record.name} />}
                  {record.address && <DetailField label={t('darkWebMonitoring.details.address')} value={record.address} />}
                  {record.vin && <DetailField label={t('darkWebMonitoring.details.vin')} value={record.vin} />}
                </div>
              </section>

              <section className="grid gap-x-6 gap-y-4 sm:grid-cols-2">
                <DetailField icon={<CalendarDays size={13} />} label={t('darkWebMonitoring.details.lastSeen')} value={formatDateMedium(record.timestamp, i18n.language)} />
                <DetailField icon={<FileText size={13} />} label={t('darkWebMonitoring.table.leakName')} value={valueOrDash(record.leakName)} />
                <DetailField icon={<Database size={13} />} label={t('darkWebMonitoring.table.database')} value={valueOrDash(record.databaseName)} />
                <DetailField icon={<FileText size={13} />} label={t('darkWebMonitoring.table.foundIn')} value={valueOrDash(record.foundIn)} />
                <DetailField icon={<FileText size={13} />} label={t('darkWebMonitoring.table.source')} value={valueOrDash(record.source)} />
                <DetailField icon={<Globe2 size={13} />} label={t('darkWebMonitoring.details.victimDomain')} value={valueOrDash(record.victimDomain)} />
                <DetailField icon={<Layers3 size={13} />} label={t('darkWebMonitoring.details.distributionType')} value={valueOrDash(record.distributionType)} />
                <DetailField icon={<UserRound size={13} />} label={t('darkWebMonitoring.table.who')} value={record.employee ? t('darkWebMonitoring.employee') : t('darkWebMonitoring.other')} />
              </section>

              <section className="space-y-4">
                {record.breachStatus === 'OPEN' && (
                  <div className="flex items-start justify-between gap-4 rounded-lg border border-primary/40 bg-primary/5 p-4">
                    <div className="flex items-start gap-3">
                      <Info size={14} className="mt-0.5 shrink-0 text-primary" />
                      <div>
                        <div className="mb-1 text-[10px] font-bold tracking-[0.14em] text-primary">{t('darkWebMonitoring.details.howTo')}</div>
                        <p className="leading-relaxed text-foreground/90" dangerouslySetInnerHTML={{ __html: safeHtml(t('darkWebMonitoring.details.howToBody')) }} />
                      </div>
                    </div>
                    <button
                      type="button"
                      onClick={() => onStatusChange?.('IN_MITIGATION')}
                      className="shrink-0 rounded-md bg-primary px-4 py-2.5 text-xs font-semibold text-primary-foreground transition-colors hover:bg-primary/90"
                    >
                      {t('darkWebMonitoring.details.mitigateBreach')}
                    </button>
                  </div>
                )}

                {inMitigation && (
                  <div>
                    <h2 className="mb-3 text-lg font-bold tracking-tight">{t('darkWebMonitoring.details.mitigationStepsTitle')}</h2>
                    <div className="space-y-4">
                      {MITIGATION_STEP_KEYS.map((key) => {
                        const stepStatus = stepStatuses[key]
                        return (
                          <div key={key} className="rounded-xl border border-border bg-card p-4">
                            <div className="flex items-center justify-between gap-3">
                              <h3 className="text-sm font-bold">{t(`darkWebMonitoring.details.mitigationSteps.labels.${key}`)}</h3>
                              <StepStatusSelect
                                status={stepStatus}
                                onChange={(next) => setStepStatuses((current) => ({ ...current, [key]: next }))}
                              />
                            </div>
                            <div className="mt-3 flex items-start gap-3">
                              <img
                                src={MITIGATION_STEP_IMAGES[key]}
                                alt=""
                                aria-hidden="true"
                                className="h-20 w-20 shrink-0 rounded-lg border border-border object-cover"
                              />
                              <p className="leading-relaxed text-foreground/80">{t(`darkWebMonitoring.details.mitigationSteps.descriptions.${key}`)}</p>
                            </div>
                          </div>
                        )
                      })}
                    </div>
                  </div>
                )}

                <div>
                  <h2 className="text-xl font-bold tracking-tight">{t('darkWebMonitoring.details.whyListed')}</h2>
                  <p className="mt-3 leading-relaxed text-foreground/90" dangerouslySetInnerHTML={{ __html: safeHtml(t('darkWebMonitoring.details.whyListedBody', { type: subjectTypeLabel })) }} />
                  <ul className="mt-3 space-y-2.5">
                    {([
                      'dataBreach',
                      'phishing',
                      'forum',
                      'dnsHijack',
                      'misconfigured',
                    ] as const).map((reason) => (
                      <li key={reason} className="flex gap-2 border-b border-border/60 pb-2.5 leading-relaxed text-foreground/80" dangerouslySetInnerHTML={{ __html: safeHtml(t(`darkWebMonitoring.details.listedReasons.${reason}`, { type: subjectTypeLabel })) }} />
                    ))}
                  </ul>
                </div>
              </section>
            </div>
          ) : (
            <div role="tabpanel" className="rounded-lg border border-border bg-background px-4 py-8 text-center text-muted-foreground">
              <Layers3 size={18} className="mx-auto mb-2" />
              {record.echoesCount > 0
                ? t('darkWebMonitoring.details.echoesCount', { count: record.echoesCount })
                : t('darkWebMonitoring.details.noEchoes')}
            </div>
          )}
        </div>
      </aside>
    </div>,
    document.body,
  )
}

function DetailField({ icon, label, value }: { icon?: ReactNode; label: string; value: string }) {
  return (
    <div className="min-w-0">
      <div className="mb-1.5 flex items-center gap-1.5 font-semibold">
        {icon && <span className="text-primary">{icon}</span>}
        {label}
      </div>
      <div className="break-words text-foreground/90">{value}</div>
    </div>
  )
}

function StepStatusSelect({
  status,
  onChange,
}: {
  status: MitigationStepStatus
  onChange: (status: MitigationStepStatus) => void
}) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const triggerRef = useRef<HTMLButtonElement>(null)

  return (
    <>
      <button
        ref={triggerRef}
        type="button"
        aria-haspopup="listbox"
        aria-expanded={open}
        onClick={(event) => {
          event.stopPropagation()
          setOpen((current) => !current)
        }}
        className={`inline-flex items-center gap-1.5 whitespace-nowrap rounded-md px-2.5 py-1.5 text-[10px] font-medium ring-1 ring-inset ${MITIGATION_STEP_STATUS_STYLES[status]}`}
      >
        {t(`darkWebMonitoring.details.mitigationSteps.status.${status}`)}
        <ChevronDown size={11} />
      </button>
      <FloatingPanel open={open} onClose={() => setOpen(false)} triggerRef={triggerRef}>
        <div role="listbox">
          {MITIGATION_STEP_STATUSES.map((option) => (
            <button
              key={option}
              type="button"
              role="option"
              aria-selected={option === status}
              disabled={option === status}
              onClick={() => {
                onChange(option)
                setOpen(false)
              }}
              className="flex w-full items-center px-3 py-1.5 text-left hover:bg-muted disabled:opacity-50 disabled:hover:bg-transparent"
            >
              <span className={`inline-flex items-center gap-1.5 whitespace-nowrap rounded-md px-2 py-1 text-[10px] font-medium ring-1 ring-inset ${MITIGATION_STEP_STATUS_STYLES[option]}`}>
                {t(`darkWebMonitoring.details.mitigationSteps.status.${option}`)}
              </span>
            </button>
          ))}
        </div>
      </FloatingPanel>
    </>
  )
}
