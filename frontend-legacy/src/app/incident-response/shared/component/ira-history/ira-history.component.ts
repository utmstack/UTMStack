import {Component, Input, OnInit} from '@angular/core';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';
import {IncidentResponseRuleHistoryService} from '../../services/incident-response-rule-history.service';
import {IncidentRuleType} from '../../type/incident-rule.type';
import {IraHistoryType} from '../../type/ira-history.type';

@Component({
  selector: 'app-ira-history',
  templateUrl: './ira-history.component.html',
  styleUrls: ['./ira-history.component.css']
})
export class IraHistoryComponent implements OnInit {
  @Input() incidentRule: IncidentRuleType;
  incidentRulesHistory: IraHistoryType[] = [];

  loading = true;

  constructor(private incidentResponseRuleHistoryService: IncidentResponseRuleHistoryService) {
  }

  ngOnInit() {
    this.getHistory();
  }

  getHistory() {
    this.incidentResponseRuleHistoryService.query({page: 0, size: 1000, 'ruleId.equals': this.incidentRule.id, sort: 'createdDate,desc'})
      .subscribe(response => {
        if (response.body) {
          this.incidentRulesHistory = response.body;
          this.loading = false;
        }
      });
  }

  describeChange(currentRule: IraHistoryType): string {
    const index = this.incidentRulesHistory.findIndex(value => value.id === currentRule.id);
    const rule = this.incidentRulesHistory[index].previousState;
    let nextRule;
    const changes: string[] = [];
    if (this.incidentRulesHistory[index + 1]) {
      nextRule = this.incidentRulesHistory[index + 1].previousState;

      if (rule.name !== nextRule.name) {
        changes.push(`changed the name from "${nextRule.name}" to "${rule.name}"`);
      }
      if (rule.description !== nextRule.description) {
        changes.push(`changed the description from "${nextRule.description}" to "${rule.description}"`);
      }
      // Compare conditions array
      if (JSON.stringify(rule.conditions) !== JSON.stringify(nextRule.conditions)) {
        changes.push(`changed the conditions`);
      }
      if (rule.command !== nextRule.command) {
        changes.push(`changed the command from <code>${nextRule.command}</code> to <code>${rule.command}</code>`);
      }
      if (rule.active !== nextRule.active) {
        changes.push(`changed the active status from "${rule.active.toString() === 'false' ? 'Inactive' : 'Active'}"
         to "${nextRule.active.toString() === 'false' ? 'Inactive' : 'Active'}"`);
      }
      if (JSON.stringify(rule.excludedAgents) !== JSON.stringify(nextRule.excludedAgents)) {
        changes.push(`changed the excluded agents`);
      }
      if (rule.agentPlatform !== nextRule.agentPlatform) {
        changes.push(`changed the agent platform from "${nextRule.agentPlatform}" to "${rule.agentPlatform}"`);
      }
      if (changes.length > 0) {
        return `The user <b>${rule.lastModifiedBy}</b> made the following changes: ${changes.join(', ')}.`;
      }
    } else {
      return `The user <b>${rule.createdBy}</b> has created a new rule in the system.
              The rule, named ${rule.name}, carries the following description: ${rule.description}.
              It operates under the following conditions: ${JSON.stringify(rule.conditions)}.
              The command it follows is: <code>${rule.command}</code>.
              The status of the rule is currently ${rule.active ? 'Active' : 'Inactive'},
              and it applies to the ${rule.agentPlatform} platform.
              With exluded agents: ${rule.excludedAgents}.`;
    }
  }

  areFilterTypesEqual(a: ElasticFilterType[], b: ElasticFilterType[]): boolean {
    if (a.length !== b.length) {
      return false;
    }

    for (let i = 0; i < a.length; i++) {
      if (a[i].label !== b[i].label || a[i].field !== b[i].field || a[i].value !== b[i].value) {
        return false;
      }
    }

    return true;
  }

}
