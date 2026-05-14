import {Component, Input, OnInit} from '@angular/core';
import {ALERT_FIELDS} from '../../../../../shared/constants/alert/alert-field.constant';
import {FILTER_OPERATORS} from '../../../../../shared/constants/filter-operators.const';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {AlertRuleType} from '../../../alert-rules/alert-rule.type';
import { BehaviorSubject } from 'rxjs';
import{ElasticFilterType} from 'src/app/shared/types/filter/elastic-filter.type'

@Component({
  selector: 'app-alert-rule-detail',
  templateUrl: './alert-rule-detail.component.html',
  styleUrls: ['./alert-rule-detail.component.scss']
})
export class AlertRuleDetailComponent implements OnInit {
  @Input() rule: AlertRuleType;
  ruleCondition = new BehaviorSubject<ElasticFilterType[]>([])

  constructor() { }

  ngOnInit() {
    this.ruleCondition.next(this.rule.conditions)
  }

  getFieldName(field: string): string {
    const alert_field = ALERT_FIELDS.find(value => value.field.trim().toLowerCase() == field.trim().toLowerCase())
    return alert_field? alert_field.label : field;
  }

  getFilterName(operator: ElasticOperatorsEnum): string {
    const indxOpe = FILTER_OPERATORS.findIndex(value => value.operator === operator);
    if (indxOpe !== -1) {
      return FILTER_OPERATORS[indxOpe].name;
    } else {
      return operator;
    }
  }


}
