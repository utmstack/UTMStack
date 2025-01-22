import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {FILTER_OPERATORS} from '../../../../shared/constants/filter-operators.const';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';
import {OperatorsType} from '../../../../shared/types/filter/operators.type';
import {CpReportBehavior} from '../../behavior/cp-report.behavior';
import {ComplianceTypeEnum} from '../../enums/compliance-type.enum';
import {CpReportsService} from '../../services/cp-reports.service';
import {ComplianceReportType} from '../../type/compliance-report.type';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';

@Component({
  selector: 'app-utm-compliance-create',
  templateUrl: './utm-compliance-create.component.html',
  styleUrls: ['./utm-compliance-create.component.scss']
})
export class UtmComplianceCreateComponent implements OnInit {
  @Input() report: ComplianceReportType;
  @Output() reportCreated = new EventEmitter<string>();
  step = 1;
  stepCompleted: number[] = [];
  creating = false;
  viewSection = false;
  standardSectionId: number;
  operators: OperatorsType[] = FILTER_OPERATORS;
  operatorEnum = ElasticOperatorsEnum;
  solution = '';
  dashboardId: number;
  name: string;
  complianceForm: FormGroup;


  constructor(private cpReportsService: CpReportsService,
              public activeModal: NgbActiveModal,
              private cpReportBehavior: CpReportBehavior,
              private utmToastService: UtmToastService,
              public modalService: NgbModal,
              private fb: FormBuilder) {
  }

  ngOnInit() {
    this.complianceForm = this.fb.group({
      reportName: [this.report ? this.report.configReportName : '' ,
        [Validators.required, Validators.minLength(10), Validators.maxLength(200)]],
      solution: [this.report ? this.report.configSolution : '', [Validators.maxLength(2000)]],
      remediation: [this.report && this.report.configRemediation ? this.report.configRemediation : '', [Validators.maxLength(2000)]]
    });

    if (this.report) {
      this.viewSection = true;
      this.standardSectionId = this.report.standardSectionId;
      this.dashboardId = this.report.dashboardId;
    }
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


  createCompliance() {
    this.creating = true;
    const reportCompliance: ComplianceReportType = {
      // columns: this.columns,
      // configReportDataOrigin: this.dataOrigen,
      configReportEditable: true,
      // configReportExportCsvUrl: ELASTIC_SEARCH_CSV_ENDPOINT,
      // configReportFilterByTime: true,
      // configReportPageable: true,
      dashboardId: this.dashboardId,
      // configReportRequestType: ComplianceRequestTypeEnum.POST,
      // configReportResourceUrl: ELASTIC_SEARCH_ENDPOINT,
      configSolution: this.complianceForm.controls.solution.value.replace(/\r?\n/g, '<br/>'),
      configRemediation: this.complianceForm.controls.remediation.value.replace(/\r?\n/g, '<br/>'),
      // requestBodyFilters: this.filters,
      // requestParamFilters: REQUEST_PARAMS_FILTER,
      configType: ComplianceTypeEnum.TEMPLATE,
      standardSectionId: this.standardSectionId,
      configReportName: this.complianceForm.controls.reportName.value,
    };
    if (this.report) {
      reportCompliance.id = this.report.id;
      this.cpReportsService.update(reportCompliance).subscribe(() => {
        this.utmToastService.showSuccessBottom('Compliance report  edited successfully');
        this.activeModal.close();
        this.reportCreated.emit('edited');
      }, error1 => {
        this.creating = false;
        this.utmToastService.showError('Error', 'Error editing compliance report');
      });
    } else {
      this.cpReportsService.create(reportCompliance).subscribe(() => {
        this.utmToastService.showSuccessBottom('Compliance report  created successfully');
        this.activeModal.close();
        this.cpReportBehavior.$reportUpdate.next('update');
        this.reportCreated.emit('created');
      }, error1 => {
        this.creating = false;
        this.utmToastService.showError('Error', 'Error creating compliance report');
      });
    }
  }

  onDashboardSelected($event: number) {
    this.dashboardId = $event;
  }
}
