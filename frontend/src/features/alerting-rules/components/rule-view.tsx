import type { TFunction } from 'i18next'
import { useDateFormat } from '@/shared/lib/datetime'
import type { CorrelationRule } from '../services/alerting-rules-http.service'
import { asList } from '../lib/as-list'
import { hasItems } from '../lib/has-items'
import { AfterEventsView } from './after-events-view'
import { DefinitionView } from './definition-view'
import { RuleDescription } from './rule-description'
import { Row } from './row'
import { Section } from './section'
import { ruleToForm } from './rule-form'

export function RuleView({ rule, t }: { rule: CorrelationRule; t: TFunction }) {
  const df = useDateFormat()
  const steps = ruleToForm(rule).correlation
  return (
    <div className="space-y-4">
      {rule.description && <Section title={t('alertingRules.view.description')}><RuleDescription text={rule.description} /></Section>}
      <Section title={t('alertingRules.view.details')}>
        <dl className="grid grid-cols-[150px_1fr] gap-y-2 text-xs">
          <Row k={t('alertingRules.table.category')}>{rule.category || '—'}</Row>
          <Row k={t('alertingRules.table.technique')}><span className="font-mono">{rule.technique || '—'}</span></Row>
          <Row k={t('alertingRules.table.adversary')}>{rule.adversary ? t(`alertingRules.adversary.${rule.adversary}`) : '—'}</Row>
          <Row k={t('alertingRules.view.impact')}><span className="font-mono">C{rule.confidentiality} · I{rule.integrity} · A{rule.availability}</span></Row>
          <Row k={t('alertingRules.view.dataTypes')}><span className="font-mono">{(rule.dataTypes ?? []).filter((d) => d.included).map((d) => d.dataType).join(', ') || '—'}</span></Row>
          {rule.ruleLastUpdate && <Row k={t('alertingRules.view.updated')}>{df.formatDateTime(rule.ruleLastUpdate)}</Row>}
        </dl>
      </Section>
      <Section title={t('alertingRules.view.definition')}>
        <DefinitionView definition={rule.definition} t={t} />
      </Section>
      {steps.length > 0 && (
        <Section title={t('alertingRules.view.correlationSteps')}>
          <AfterEventsView steps={steps} t={t} />
        </Section>
      )}
      {(hasItems(rule.groupBy) || hasItems(rule.deduplicateBy)) && (
        <Section title={t('alertingRules.view.correlation')}>
          <dl className="grid grid-cols-[150px_1fr] gap-y-2 text-xs">
            <Row k={t('alertingRules.view.groupBy')}><span className="font-mono">{asList(rule.groupBy)}</span></Row>
            <Row k={t('alertingRules.view.deduplicateBy')}><span className="font-mono">{asList(rule.deduplicateBy)}</span></Row>
          </dl>
        </Section>
      )}
    </div>
  )
}
