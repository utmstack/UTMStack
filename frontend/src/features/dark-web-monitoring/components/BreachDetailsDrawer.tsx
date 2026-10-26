import { useEffect, useState, type ReactNode } from 'react'
import { createPortal } from 'react-dom'
import { AlertCircle, CalendarDays, Copy, Database, Eye, EyeOff, FileText, Globe2, Layers3, ShieldAlert, UserRound, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import type { BreachStatus, DarkWebRecord } from '../types'

interface BreachDetailsDrawerProps {
  record: DarkWebRecord
  onClose: () => void
}

const STATUS_STYLES: Record<BreachStatus, string> = {
  OPEN: 'bg-red-500/10 text-red-400',
  IN_MITIGATION: 'bg-amber-500/10 text-amber-400',
  COMPLETED: 'bg-blue-500/10 text-blue-400',
}
const STATUS_KEYS: Record<BreachStatus, string> = {
  OPEN: 'open',
  IN_MITIGATION: 'mitigation',
  COMPLETED: 'completed',
}

function valueOrDash(value: string | null): string {
  return value || '—'
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value))
}

export function BreachDetailsDrawer({ record, onClose }: BreachDetailsDrawerProps) {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState<'information' | 'echoes'>('information')
  const [showSecrets, setShowSecrets] = useState(false)

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', onKeyDown)
    return () => document.removeEventListener('keydown', onKeyDown)
  }, [onClose])

  const subject = record.organizationElement
  const subjectType = t(`darkWebMonitoring.exportDialogs.elementTypes.${record.organizationElementType}`)
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
                  <span className={`inline-flex rounded-full px-2 py-1 text-[10px] font-semibold ${STATUS_STYLES[record.breachStatus]}`}>
                    {t(`darkWebMonitoring.status.${STATUS_KEYS[record.breachStatus]}`)}
                  </span>
                </div>
              </div>

              {record.imageUrl && (
                <div className="overflow-hidden rounded-xl border border-border bg-muted">
                  <img src={record.imageUrl} alt={t('darkWebMonitoring.details.evidence')} className="max-h-80 w-full object-contain" />
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
                <DetailField icon={<CalendarDays size={13} />} label={t('darkWebMonitoring.details.lastSeen')} value={formatDate(record.timestamp)} />
                <DetailField icon={<FileText size={13} />} label={t('darkWebMonitoring.table.leakName')} value={valueOrDash(record.leakName)} />
                <DetailField icon={<Database size={13} />} label={t('darkWebMonitoring.table.database')} value={valueOrDash(record.databaseName)} />
                <DetailField icon={<FileText size={13} />} label={t('darkWebMonitoring.table.foundIn')} value={valueOrDash(record.foundIn)} />
                <DetailField icon={<FileText size={13} />} label={t('darkWebMonitoring.table.source')} value={valueOrDash(record.source)} />
                <DetailField icon={<Globe2 size={13} />} label={t('darkWebMonitoring.details.victimDomain')} value={valueOrDash(record.victimDomain)} />
                <DetailField icon={<Layers3 size={13} />} label={t('darkWebMonitoring.details.distributionType')} value={valueOrDash(record.distributionType)} />
                <DetailField icon={<UserRound size={13} />} label={t('darkWebMonitoring.table.who')} value={record.employee ? t('darkWebMonitoring.employee') : t('darkWebMonitoring.other')} />
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
