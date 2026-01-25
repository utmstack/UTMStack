import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {CpReportBehavior} from '../../behavior/cp-report.behavior';
import {ComplianceStrategyEnum} from '../../enums/compliance-strategy.enum';
import {CpReportsConfigService} from '../../services/cp-reports-config.service';
import {UtmComplianceQueryConfigType} from '../../type/compliance-query-config.type';
import {ComplianceReportConfigType} from '../../type/compliance-report-config.type';

@Component({
  selector: 'app-utm-compliance-control-config-create',
  templateUrl: './utm-compliance-control-config-create.component.html',
  styleUrls: ['./utm-compliance-control-config-create.component.scss']
})
export class UtmComplianceControlConfigCreateComponent implements OnInit {
  @Input() report: ComplianceReportConfigType;
  @Output() reportCreated = new EventEmitter<string>();

  step = 1;
  stepCompleted: number[] = [];
  creating = false;
  viewSection = false;

  complianceForm: FormGroup;
  standardSectionId: number;
  strategies = [];

  constructor(private cpReportsConfigService: CpReportsConfigService,
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
      remediation: [this.report && this.report.configRemediation ? this.report.configRemediation : '', [Validators.maxLength(2000)]],
      strategy: [this.report ? this.report.configStrategy as ComplianceStrategyEnum : null, [Validators.required]],
      queriesConfigs: this.fb.array([], Validators.minLength(1))
    });

    this.strategies = Object.keys(ComplianceStrategyEnum).map(
      key => ({
        id: key as ComplianceStrategyEnum,
        strategy: ComplianceStrategyEnum[key as keyof typeof ComplianceStrategyEnum]
      })
    );

    if (this.report) {
      this.viewSection = true;
      this.standardSectionId = this.report.standardSectionId;
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

  createCompliance() {
    this.creating = true;
    const reportConfigCompliance: ComplianceReportConfigType = {
      standardSectionId: this.standardSectionId,
      configReportName: this.complianceForm.controls.reportName.value,
      configSolution: this.complianceForm.controls.solution.value.replace(/\r?\n/g, '<br/>'),
      configRemediation: this.complianceForm.controls.remediation.value.replace(/\r?\n/g, '<br/>'),
      configStrategy: this.complianceForm.controls.strategy.value,
      queriesConfigs: this.queriesList
    };

    //Edit
    if (this.report) {
      reportConfigCompliance.id = this.report.id;
      this.cpReportsConfigService.update(reportConfigCompliance).subscribe(() => {
        this.utmToastService.showSuccessBottom('Compliance report  edited successfully');
        this.activeModal.close();
        this.reportCreated.emit('edited');
      }, error1 => {
        this.creating = false;
        this.utmToastService.showError('Error', 'Error editing compliance report');
      });
    } else {
      this.cpReportsConfigService.create(reportConfigCompliance).subscribe(() => {
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

  get queriesConfigs(): FormArray {
    return this.complianceForm.get('queriesConfigs') as FormArray;
  }

  get queriesList() {
    return this.queriesConfigs.controls.map(c => c.value);
  }

  onQueryAdd(event: { query: UtmComplianceQueryConfigType ; index: number | null }) {
    const { query, index } = event;
    if (index === null) {
      this.queriesConfigs.push(this.fb.control(query));
    } else {
      this.queriesConfigs.at(index).setValue(query);
    }
  }

  onQueryRemove(index: number) {
    this.queriesConfigs.removeAt(index);
  }
}
