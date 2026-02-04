import {Rule} from '../../../../../../rule-management/models/rule.model';


export function mapRuleToYaml(rule: Rule): any {
  return {
    name: rule.name,
    dataTypes: rule.dataTypes ? rule.dataTypes.map(d => d.dataType) : [],
    impact: {
      confidentiality: rule.confidentiality,
      integrity: rule.integrity,
      availability: rule.availability
    },
    category: rule.category,
    technique: rule.technique,
    adversary: rule.adversary,
    references: rule.references,
    description: rule.description,
    where: rule.definition,
    afterEvents: rule.afterEvents && rule.afterEvents.length ? rule.afterEvents : [],
    groupBy: rule.groupBy && rule.groupBy.length ? rule.groupBy : [],
    deduplicateBy: rule.deduplicateBy && rule.deduplicateBy.length ? rule.deduplicateBy : []
  };
}
