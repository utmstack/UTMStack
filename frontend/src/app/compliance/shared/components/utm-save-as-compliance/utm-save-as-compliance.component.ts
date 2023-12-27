import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {FILTER_OPERATORS} from '../../../../shared/constants/filter-operators.const';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {DataNatureTypeEnum} from '../../../../shared/enums/nature-data.enum';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';
import {OperatorsType} from '../../../../shared/types/filter/operators.type';
import {UtmFieldType} from '../../../../shared/types/table/utm-field.type';
import {ELASTIC_SEARCH_CSV_ENDPOINT, ELASTIC_SEARCH_ENDPOINT} from '../../constant/compliance.constant';
import {REQUEST_PARAMS_FILTER} from '../../constant/request-params-filter.constant';
import {ComplianceRequestTypeEnum} from '../../enums/compliance-request-type.enum';
import {CpReportsService} from '../../services/cp-reports.service';
import {ComplianceReportType} from '../../type/compliance-report.type';

@Component({
  selector: 'app-utm-save-as-compliance',
  templateUrl: './utm-save-as-compliance.component.html',
  styleUrls: ['./utm-save-as-compliance.component.scss']
})
export class UtmSaveAsComplianceComponent implements OnInit {
  @Input() filters: ElasticFilterType[];
  @Input() columns: UtmFieldType[];
  @Input() dataOrigen: number;
  step = 1;
  stepCompleted: number[] = [];
  creating = false;
  viewSection = false;
  standardSectionId: number;
  operators: OperatorsType[] = FILTER_OPERATORS;
  operatorEnum = ElasticOperatorsEnum;
  solution = '';


  constructor(private cpReportsService: CpReportsService,
              public activeModal: NgbActiveModal,
              private utmToastService: UtmToastService,
              public modalService: NgbModal) {
  }

  ngOnInit() {
  }


  backStep() {
    this.step -= 1;
    this.stepCompleted.pop();
  }

  nextStep() {
    this.stepCompleted.push(this.step);
    this.step += 1;
  }

  isCompleted(step: number) {
    return this.stepCompleted.findIndex(value => value === step) !== -1;
  }

  onStandardChange($event: number) {
    this.viewSection = true;
  }

  onStandardSectionChange($event: number) {
    this.standardSectionId = $event;
  }

  getFilterLabel(filter: ElasticFilterType): string {
    return filter.field + ' ' +
      this.extractOperator(filter.operator) + ' ' +
      (filter.value ? filter.value.toString().replace(',', ' and ') : '');
  }

  extractOperator(operator: string): string {
    const index = this.operators.findIndex(value => value.operator === operator);
    return this.operators[index].name;
  }

  deleteFilter(index: number) {
    this.filters.splice(index, 1);
  }

  createCompliance() {
    this.creating = true;
    const reportCompliance: ComplianceReportType = {
      columns: this.columns,
      configReportDataOrigin: this.dataOrigen,
      configReportEditable: true,
      configReportExportCsvUrl: ELASTIC_SEARCH_CSV_ENDPOINT,
      configReportFilterByTime: true,
      configReportPageable: true,
      configReportRequestType: ComplianceRequestTypeEnum.POST,
      configReportResourceUrl: ELASTIC_SEARCH_ENDPOINT,
      configSolution: this.solution.replace(/\r?\n/g, '<br/>'),
      requestBodyFilters: this.filters,
      requestParamFilters: REQUEST_PARAMS_FILTER,
      standardSectionId: this.standardSectionId
    };
    this.cpReportsService.create(reportCompliance).subscribe(() => {
      this.utmToastService.showSuccessBottom('Compliance report  created successfully');
      this.activeModal.close();
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error', 'Error creating compliance report');
    });

  }
}
