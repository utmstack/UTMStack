import { ruleToForm } from '../components/rule-form'
import type { CorrelationRule } from '../services/alerting-rules-http.service'
import { ruleFormToYaml } from './rule-yaml'

export function downloadRuleYaml(rule: CorrelationRule): void {
  const yaml = ruleFormToYaml(ruleToForm(rule))
  const base = rule.relPath.split('/').pop() || `${rule.name || 'rule'}.yaml`
  const name = /\.ya?ml$/i.test(base) ? base : `${base}.yaml`
  const url = URL.createObjectURL(new Blob([yaml], { type: 'text/yaml;charset=utf-8' }))
  const a = document.createElement('a')
  a.href = url
  a.download = name
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}
