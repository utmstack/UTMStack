import {Component, Input, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {VisualizationType} from '../../../shared/chart/types/visualization.type';
import {ElasticOperatorsEnum} from '../../../shared/enums/elastic-operators.enum';
import {ElasticTimeEnum} from '../../../shared/enums/elastic-time.enum';
import {ElasticFilterCommonType} from '../../../shared/types/filter/elastic-filter-common.type';
import {TimeFilterType} from '../../../shared/types/time-filter.type';
import {ComplianceParamsEnum} from '../../shared/enums/compliance-params.enum';
import {ComplianceRequestTypeEnum} from '../../shared/enums/compliance-request-type.enum';
import {ComplianceReportType} from '../../shared/type/compliance-report.type';

@Component({
  selector: 'app-compliance-result-params',
  templateUrl: './compliance-result-params.component.html',
  styleUrls: ['./compliance-result-params.component.scss']
})
export class ComplianceResultParamsComponent implements OnInit {
  @Input() template?: ComplianceReportType;
  defaultTime: ElasticFilterCommonType = {time: ElasticTimeEnum.HOUR, last: 24, label: 'last 24 hours'};
  complianceParamsEnum = ComplianceParamsEnum;
  queryParams = {};
  dataOrigin = ComplianceParamsEnum.DATA_ORIGIN;
  page = ComplianceParamsEnum.PAGE;
  visualizationsId: number[] = [];
  visualizationParam = 'visualizationsId';

  constructor(public activeModal: NgbActiveModal,
              private router: Router) {
  }

  ngOnInit() {
    this.queryParams[this.complianceParamsEnum.TEMPLATE] = this.template.id;
    this.queryParams[this.dataOrigin] = this.template.configReportDataOrigin;
    if (this.template.configReportPageable) {
      this.queryParams[this.page] = 1;
    }
  }

  hasLimitFilter() {
    return this.template.requestParamFilters.findIndex(value => value.param === this.complianceParamsEnum.TOP) !== -1;
  }

  hasTimeFilter() {
    return this.template.configReportFilterByTime;
  }

  onTimeFilterChange($event: TimeFilterType) {
    if (this.template.configReportRequestType === ComplianceRequestTypeEnum.GET) {
      this.queryParams[this.complianceParamsEnum.FROM] = $event.timeFrom;
      this.queryParams[this.complianceParamsEnum.TO] = $event.timeTo;
    }
    this.queryParams[this.complianceParamsEnum.TIMESTAMP] = ElasticOperatorsEnum.IS_BETWEEN +
      '->' + $event.timeFrom + ',' + $event.timeTo;
  }

  generateReports() {
    this.activeModal.close();
    this.router.navigate(['compliance/template-result'], {
      queryParams: this.queryParams
    });
  }

  onVisSelected($event: VisualizationType[]) {
    $event.forEach(value => this.visualizationsId.push(value.id));
    this.queryParams[this.visualizationParam] = this.visualizationsId.toString();
  }

}
