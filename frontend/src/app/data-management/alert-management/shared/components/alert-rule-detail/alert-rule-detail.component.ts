import {Component, Input, OnInit} from '@angular/core';
import {ALERT_FIELDS} from '../../../../../shared/constants/alert/alert-field.constant';
import {FILTER_OPERATORS} from '../../../../../shared/constants/filter-operators.const';
import {ElasticOperatorsEnum} from '../../../../../shared/enums/elastic-operators.enum';
import {AlertRuleType} from '../../../alert-rules/alert-rule.type';

@Component({
  selector: 'app-alert-rule-detail',
  templateUrl: './alert-rule-detail.component.html',
  styleUrls: ['./alert-rule-detail.component.scss']
})
export class AlertRuleDetailComponent implements OnInit {
  @Input() rule: AlertRuleType;

  constructor() { }

  ngOnInit() {
  }

  getFieldName(field: string): string {
    return ALERT_FIELDS.filter(value => value.field === field)[0].label;
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
