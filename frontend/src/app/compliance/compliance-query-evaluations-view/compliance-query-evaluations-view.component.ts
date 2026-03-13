import {Component, Input, OnDestroy, OnInit} from '@angular/core';
import {ComplianceStatusExtendedEnum} from '../shared/enums/compliance-status.enum';
import {ComplianceControlEvaluationsType} from '../shared/type/compliance-control-evaluations.type';
import {ComplianceQueryEvaluationType} from '../shared/type/compliance-query-evaluation.type';

@Component({
  selector: 'app-compliance-query-evaluations-view',
  templateUrl: './compliance-query-evaluations-view.component.html',
  styleUrls: ['./compliance-query-evaluations-view.component.scss']
})
export class ComplianceQueryEvaluationsViewComponent implements OnInit, OnDestroy {
  @Input() evaluation: ComplianceControlEvaluationsType;
  queryDetail: ComplianceQueryEvaluationType;

  ComplianceStatusExtendedEnum = ComplianceStatusExtendedEnum;

  constructor() {
  }

  ngOnDestroy(): void {

  }

  ngOnInit(): void {
     console.log('queryEval', this.evaluation);
  }

}
