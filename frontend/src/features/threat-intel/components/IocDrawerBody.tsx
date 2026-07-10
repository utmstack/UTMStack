import { useTranslation } from 'react-i18next'
import { ExternalLink, Globe2, MapPin, MoreHorizontal, Search, Shield, Sparkles, X } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/components/ui/button'
import type { EntityDetail, EntityRelation } from '../domain/threat-intel.types'
import { REPUTATION_STYLE, reputationTone, typeMeta } from './utils/severity-style'
import { absTimestamp, relativeTime } from './utils/time-format'
import { Section } from './Section'
import { KV } from './KV'
import { Metric } from './Metric'

interface IocDrawerBodyProps {
  detail: EntityDetail
  relations: EntityRelation[]
  onClose: () => void
}

export function IocDrawerBody({ detail, relations, onClose }: IocDrawerBodyProps) {
  const { t } = useTranslation()
  const a = detail.attributes
  const tone = reputationTone(a.reputation_score)
  const rep = REPUTATION_STYLE[tone]
  const typeMeta_ = typeMeta(a.type)
  const TIcon = typeMeta_.icon
  const associations = detail.latest_associations ?? []
  const merged = relations.length > 0 ? relations : associations
  const meta = detail.metadata
  const hasMeta = !!(meta && (meta.asn || meta.aso || meta.city || meta.country))

  return (
    <>
      <header className="border-b border-border px-6 py-4">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
              <TIcon size={11} className={typeMeta_.tone} />
              <span>{t(typeMeta_.labelKey)}</span>
              <span>·</span>
              <span className={cn('font-medium', rep.tone)}>{a.reputation}</span>
              <span>·</span>
              <span>{t('threatIntel.drawer.accuracy')} {a.accuracy}</span>
            </div>
            <h2 className="mt-1.5 break-all font-mono text-base font-semibold leading-snug">
              {a.value}
            </h2>
            {a.description && (
              <p className="mt-1.5 line-clamp-2 text-[11px] text-muted-foreground">{a.description}</p>
            )}
            {a.tags.length > 0 && (
              <div className="mt-2 flex flex-wrap items-center gap-1.5 text-[11px]">
                {a.tags.map((tag) => (
                  <span key={tag} className="rounded-md bg-muted px-1.5 py-0.5">
                    {tag}
                  </span>
                ))}
              </div>
            )}
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
          >
            <X size={16} />
          </button>
        </div>

        <div className="mt-4 _flex flex-wrap items-center gap-2 hidden">
          <Button size="sm">
            <Shield size={13} className="mr-1.5" />
            {t('threatIntel.drawer.actions.block')}
          </Button>
          <Button size="sm" variant="outline">
            <Search size={13} className="mr-1.5" />
            {t('threatIntel.drawer.actions.viewEvents')}
          </Button>
          <Button size="sm" variant="outline">
            <ExternalLink size={13} className="mr-1.5" />
            {t('threatIntel.drawer.actions.externalLookup')}
          </Button>
          <Button size="sm" variant="ghost" className="ml-auto">
            <MoreHorizontal size={13} />
          </Button>
        </div>
      </header>

      <div className="flex-1 overflow-y-auto bg-muted/20">
        <div className="grid grid-cols-3 gap-4 p-6">
          <div className="col-span-2 space-y-4">
            <Section title={t('threatIntel.drawer.sections.timeline')}>
              <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
                <KV k={t('threatIntel.drawer.fields.firstSeen')}>
                  <span className="font-mono">{absTimestamp(a.first_seen)}</span>
                </KV>
                <KV k={t('threatIntel.drawer.fields.lastSeen')}>
                  <span className="font-mono">{absTimestamp(a.last_seen)}</span>
                </KV>
              </dl>
            </Section>

            {hasMeta && (
              <Section title={t('threatIntel.drawer.sections.networkOrigin')}>
                <dl className="grid grid-cols-[140px_1fr] gap-y-2 text-xs">
                  {meta.country && <KV k={t('threatIntel.drawer.fields.country')}><span>{meta.country}</span></KV>}
                  {meta.city && <KV k={t('threatIntel.drawer.fields.city')}><span>{meta.city}</span></KV>}
                  {meta.aso && <KV k={t('threatIntel.drawer.fields.aso')}><span>{meta.aso}</span></KV>}
                  {meta.asn > 0 && <KV k={t('threatIntel.drawer.fields.asn')}><span className="font-mono">{meta.asn}</span></KV>}
                </dl>
              </Section>
            )}

            {merged.length > 0 && (
              <Section title={t('threatIntel.drawer.sections.associationsWithCount', { count: merged.length })}>
                <ul className="space-y-1.5">
                  {merged.slice(0, 12).map((r) => {
                    const rt = typeMeta(r.type)
                    const RIcon = rt.icon
                    const rTone = reputationTone(r.reputation_score)
                    return (
                      <li
                        key={r.id}
                        className="flex items-center justify-between gap-3 rounded border border-border bg-background/40 px-3 py-2 text-xs"
                      >
                        <div className="flex min-w-0 items-center gap-2">
                          <RIcon size={11} className={rt.tone} />
                          <span className="text-[10px] uppercase tracking-wider text-muted-foreground">
                            {t(rt.labelKey)}
                          </span>
                          <span className="truncate font-mono">{r.value}</span>
                        </div>
                        <span className={cn('shrink-0 text-[11px]', REPUTATION_STYLE[rTone].tone)}>
                          {r.reputation}
                        </span>
                      </li>
                    )
                  })}
                  {merged.length > 12 && (
                    <li className="px-3 py-1 text-[11px] text-muted-foreground">
                      {t('threatIntel.drawer.moreCount', { count: merged.length - 12 })}
                    </li>
                  )}
                </ul>
              </Section>
            )}

            {detail.geolocations?.length > 0 && (
              <Section title={t('threatIntel.drawer.sections.geolocationsWithCount', { count: detail.geolocations.length })}>
                <ul className="space-y-1.5">
                  {detail.geolocations.slice(0, 6).map((g) => (
                    <li
                      key={g.object}
                      className="flex items-center justify-between gap-3 rounded border border-border bg-background/40 px-3 py-2 text-xs"
                    >
                      <div className="flex min-w-0 items-center gap-2">
                        <Globe2 size={11} className="text-violet-500" />
                        <span className="truncate font-mono">{g.object}</span>
                      </div>
                      <div className="flex items-center gap-2 text-[11px] text-muted-foreground">
                        <MapPin size={10} />
                        <span>
                          {g.city}
                          {g.country ? `, ${g.country}` : ''}
                        </span>
                      </div>
                    </li>
                  ))}
                  {detail.geolocations.length > 6 && (
                    <li className="px-3 py-1 text-[11px] text-muted-foreground">
                      {t('threatIntel.drawer.moreCount', { count: detail.geolocations.length - 6 })}
                    </li>
                  )}
                </ul>
              </Section>
            )}
          </div>

          <aside className="col-span-1 space-y-4">
            <div className="rounded-lg border border-fuchsia-500/30 bg-gradient-to-br from-fuchsia-500/10 via-violet-500/5 to-transparent p-4">
              <div className="mb-2 flex items-center gap-2 text-xs font-medium">
                <Sparkles size={13} className="text-fuchsia-500" />
                {t('threatIntel.drawer.aiBrief')}
              </div>
              <p className="text-xs leading-relaxed">
                {a.description || t('threatIntel.drawer.noDescription')}
              </p>
              <p className="mt-2 text-[11px] text-muted-foreground">
                {t('threatIntel.drawer.lastObserved', { when: relativeTime(a.last_seen) })}
              </p>
            </div>

            <div className="rounded-lg border border-border bg-card p-4">
              <div className="mb-2 text-[11px] uppercase tracking-wider text-muted-foreground">
                {t('threatIntel.drawer.fields.reputation')}
              </div>
              <div className="space-y-2 text-xs">
                <Metric label={t('threatIntel.drawer.fields.current')} value={`${a.reputation} (${a.reputation_score})`} />
                <Metric label={t('threatIntel.drawer.fields.best')} value={`${a.best_reputation} (${a.best_reputation_score})`} />
                <Metric label={t('threatIntel.drawer.fields.worst')} value={`${a.worst_reputation} (${a.worst_reputation_score})`} />
                <Metric label={t('threatIntel.drawer.fields.accuracy')} value={`${a.accuracy} (${a.accuracy_score})`} />
              </div>
            </div>
          </aside>
        </div>
      </div>
    </>
  )
}
