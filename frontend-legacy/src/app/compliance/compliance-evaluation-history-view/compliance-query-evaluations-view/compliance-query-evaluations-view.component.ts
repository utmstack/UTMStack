import {Component, Input} from '@angular/core';
import {ComplianceStatusExtendedEnum} from '../../shared/enums/compliance-status.enum';
import {ComplianceControlEvaluationHistoryType} from '../../shared/type/compliance-control-evaluation-history.type';
import {ComplianceQueryEvaluationType} from '../../shared/type/compliance-query-evaluation.type';

@Component({
  selector: 'app-compliance-query-evaluations-view',
  templateUrl: './compliance-query-evaluations-view.component.html',
  styleUrls: ['./compliance-query-evaluations-view.component.scss']
})
export class ComplianceQueryEvaluationsViewComponent {
  @Input() evaluation: ComplianceControlEvaluationHistoryType;
  queryDetail: ComplianceQueryEvaluationType;

  ComplianceStatusExtendedEnum = ComplianceStatusExtendedEnum;
}
