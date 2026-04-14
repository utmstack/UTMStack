import {Component, Input, OnInit} from '@angular/core';
import {ComplianceEvaluationRuleLabels} from '../../../shared/enums/compliance-evaluation-rule.enum';
import {ComplianceStatusExtendedEnum} from '../../../shared/enums/compliance-status.enum';
import {ComplianceQueryEvaluationType} from '../../../shared/type/compliance-query-evaluation.type';

@Component({
  selector: 'app-compliance-query-evaluation-detail',
  templateUrl: './compliance-query-evaluation-detail.component.html',
  styleUrls: ['./compliance-query-evaluation-detail.component.css']
})
export class ComplianceQueryEvaluationDetailComponent implements OnInit {
  @Input() query: ComplianceQueryEvaluationType;
  ComplianceStatusExtendedEnum = ComplianceStatusExtendedEnum;
  ComplianceEvaluationRuleLabels = ComplianceEvaluationRuleLabels;

  constructor() { }

  ngOnInit() {

  }
}
