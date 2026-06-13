import { useState } from 'react'
import { Activity, AlertTriangle, Crosshair, User, X, type LucideIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { TFunction } from 'i18next'
import { cn } from '@/shared/lib/utils'
import { KIND_ICON, THREAT, flagEmoji, relativeTime, threatFromSeverity } from '../lib/adversary-meta'
import type { Adversary } from '../types/adversary.types'
import { AdversaryAvatar } from './adversary-avatar'
import { AdversaryActivityChart } from './adversary-activity-chart'
import { DescRow, Section } from './ui-primitives'

type Tab = 'profile' | 'techniques' | 'targets' | 'alerts'

export function AdversaryDrawer({ a, onClose }: { a: Adversary; onClose: () => void }) {
  const { t } = useTranslation()
  const [tab, setTab] = useState<Tab>('profile')
  const maxTech = Math.max(...a.techniques.map((x) => x.count), 1)
  const maxTgt = Math.max(...a.targets.map((x) => x.hits), 1)

  return (
    <div className="fixed inset-0 z-50 flex items-stretch justify-end bg-black/40 backdrop-blur-sm" onClick={onClose}>
      <div
        className="flex w-full max-w-[820px] flex-col overflow-hidden border-l border-border bg-card shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="border-b border-border px-6 py-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex min-w-0 flex-1 items-start gap-3">
              <AdversaryAvatar a={a} large />
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                  <span>{t(`adversaries.kind.${a.kind}`)}</span>
                  {a.geo?.country && (
                    <>
                      <span>·</span>
                      <span>
                        {flagEmoji(a.geo.countryCode)} {a.geo.country}
                      </span>
                    </>
                  )}
                </div>
                <h2 className="mt-0.5 truncate font-mono text-xl font-semibold">{a.identifier}</h2>
                <div className="mt-2 flex flex-wrap items-center gap-2 text-[11px]">
                  <span
                    className={cn(
                      'inline-flex items-center gap-1.5 rounded-md px-1.5 py-0.5 font-medium ring-1 ring-inset',
                      THREAT[a.threatLevel].text,
                      THREAT[a.threatLevel].ring
                    )}
                  >
                    <span className={cn('h-1.5 w-1.5 rounded-full', THREAT[a.threatLevel].dot)} />
                    {t(`adversaries.threat.${a.threatLevel}`)}
                  </span>
                  <span className="text-muted-foreground">
                    {t('adversaries.card.alerts')}: <span className="font-mono text-foreground">{a.alertsCount}</span>
                  </span>
                </div>
              </div>
            </div>
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              <X size={16} />
            </button>
          </div>
          <nav className="mt-4 flex items-center gap-1">
            {(['profile', 'techniques', 'targets', 'alerts'] as Tab[]).map((id) => (
              <DrawerTab key={id} id={id} current={tab} onChange={setTab} />
            ))}
          </nav>
        </header>

        <div className="flex-1 space-y-4 overflow-y-auto bg-muted/20 p-6">
          {tab === 'profile' && <ProfileTab a={a} t={t} />}
          {tab === 'techniques' && <TechniquesTab a={a} t={t} maxTech={maxTech} />}
          {tab === 'targets' && <TargetsTab a={a} t={t} maxTgt={maxTgt} />}
          {tab === 'alerts' && <AlertsTab a={a} t={t} />}
        </div>
      </div>
    </div>
  )
}

function DrawerTab({ id, current, onChange }: { id: Tab; current: Tab; onChange: (t: Tab) => void }) {
  const { t } = useTranslation()
  const active = id === current
  const ICONS: Record<Tab, LucideIcon> = { profile: User, techniques: Activity, targets: Crosshair, alerts: AlertTriangle }
  const Icon = ICONS[id]
  return (
    <button
      onClick={() => onChange(id)}
      className={cn(
        'relative flex items-center gap-1.5 px-3 py-2 text-xs transition-colors',
        active ? 'text-foreground' : 'text-muted-foreground hover:text-foreground'
      )}
    >
      <Icon size={13} />
      {t(`adversaries.tabs.${id}`)}
      {active && <span className="absolute inset-x-2 -bottom-px h-0.5 rounded-full bg-primary" />}
    </button>
  )
}

function ProfileTab({ a, t }: { a: Adversary; t: TFunction }) {
  return (
    <>
      <Section title={t('adversaries.drawer.activity')}>
        <AdversaryActivityChart data={a.activity} />
      </Section>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <Section title={t('adversaries.drawer.identification')}>
          <dl className="grid grid-cols-[110px_1fr] gap-y-2 text-xs">
            <DescRow k={t('adversaries.drawer.kindLabel')}>{t(`adversaries.kind.${a.kind}`)}</DescRow>
            <DescRow k={t('adversaries.drawer.identifier')}>
              <span className="break-all font-mono">{a.identifier}</span>
            </DescRow>
            <DescRow k={t('adversaries.drawer.firstSeen')}>{relativeTime(a.firstSeen, t)}</DescRow>
            <DescRow k={t('adversaries.drawer.lastSeen')}>{relativeTime(a.lastSeen, t)}</DescRow>
          </dl>
        </Section>
        {a.geo && (a.geo.country || a.geo.aso) && (
          <Section title={t('adversaries.drawer.geolocation')}>
            <dl className="grid grid-cols-[110px_1fr] gap-y-2 text-xs">
              {a.geo.country && (
                <DescRow k={t('adversaries.drawer.country')}>
                  {flagEmoji(a.geo.countryCode)} {a.geo.country}
                </DescRow>
              )}
              {a.geo.city && <DescRow k={t('adversaries.drawer.city')}>{a.geo.city}</DescRow>}
              {a.geo.asn ? (
                <DescRow k="ASN">
                  <span className="font-mono">{a.geo.asn}</span>
                </DescRow>
              ) : null}
              {a.geo.aso && <DescRow k="ASO">{a.geo.aso}</DescRow>}
            </dl>
          </Section>
        )}
      </div>
    </>
  )
}

function TechniquesTab({ a, t, maxTech }: { a: Adversary; t: TFunction; maxTech: number }) {
  return (
    <Section title={t('adversaries.drawer.topTechniques')}>
      {a.techniques.length === 0 ? (
        <div className="text-xs text-muted-foreground">{t('adversaries.drawer.noTechniques')}</div>
      ) : (
        <ul className="space-y-2">
          {a.techniques.map((tch) => (
            <li key={tch.id} className="flex items-center gap-3 text-xs">
              <span className="w-48 shrink-0 truncate font-mono text-muted-foreground" title={tch.id}>
                {tch.id}
              </span>
              <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                <div className="h-full rounded-full bg-orange-500/70" style={{ width: `${(tch.count / maxTech) * 100}%` }} />
              </div>
              <span className="w-10 text-right font-mono tabular-nums text-muted-foreground">{tch.count}</span>
            </li>
          ))}
        </ul>
      )}
    </Section>
  )
}

function TargetsTab({ a, t, maxTgt }: { a: Adversary; t: TFunction; maxTgt: number }) {
  return (
    <Section title={t('adversaries.drawer.targets')}>
      {a.targets.length === 0 ? (
        <div className="text-xs text-muted-foreground">{t('adversaries.drawer.noTargets')}</div>
      ) : (
        <ul className="space-y-2">
          {a.targets.map((tg) => {
            const Icon = KIND_ICON[tg.kind]
            return (
              <li key={tg.label} className="flex items-center gap-3 text-xs">
                <Icon size={13} className="shrink-0 text-muted-foreground" />
                <span className="w-44 shrink-0 truncate font-mono">{tg.label}</span>
                <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-muted">
                  <div className="h-full rounded-full bg-sky-500/70" style={{ width: `${(tg.hits / maxTgt) * 100}%` }} />
                </div>
                <span className="w-10 text-right font-mono tabular-nums text-muted-foreground">{tg.hits}</span>
              </li>
            )
          })}
        </ul>
      )}
    </Section>
  )
}

function AlertsTab({ a, t }: { a: Adversary; t: TFunction }) {
  return (
    <Section title={t('adversaries.drawer.recentAlerts')}>
      {a.alerts.length === 0 ? (
        <div className="text-xs text-muted-foreground">{t('adversaries.drawer.noAlerts')}</div>
      ) : (
        <ul className="overflow-hidden rounded-lg border border-border">
          {a.alerts.map((al) => (
            <li key={al.id} className="flex items-center gap-3 border-b border-border/60 px-3 py-2 text-xs last:border-b-0">
              <span className={cn('h-3 w-1 shrink-0 rounded-full', THREAT[threatFromSeverity(al.severity)].dot)} />
              <div className="min-w-0 flex-1">
                <div className="truncate font-medium">{al.name}</div>
                <div className="mt-0.5 flex items-center gap-2 text-[11px] text-muted-foreground">
                  {al.technique && <span className="font-mono">{al.technique}</span>}
                  {al.target && <span>→ {al.target.label}</span>}
                  {al.echoes > 0 && <span>· {t('adversaries.drawer.echoes', { count: al.echoes })}</span>}
                </div>
              </div>
              <span className="shrink-0 font-mono text-[11px] text-muted-foreground">{relativeTime(al.timestamp, t)}</span>
            </li>
          ))}
        </ul>
      )}
    </Section>
  )
}
