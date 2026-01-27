import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {CpReportBehavior} from '../../behavior/cp-report.behavior';
import {ComplianceStrategyEnum} from '../../enums/compliance-strategy.enum';
import {CpControlConfigService} from '../../services/cp-control-config.service';
import {ComplianceControlConfigType} from '../../type/compliance-control-config.type';
import {UtmComplianceQueryConfigType} from '../../type/compliance-query-config.type';

@Component({
  selector: 'app-utm-compliance-control-config-create',
  templateUrl: './utm-compliance-control-config-create.component.html',
  styleUrls: ['./utm-compliance-control-config-create.component.scss']
})
export class UtmComplianceControlConfigCreateComponent implements OnInit {
  @Input() control: ComplianceControlConfigType;
  @Output() controlCreated = new EventEmitter<string>();

  loading = true;
  step = 1;
  stepCompleted: number[] = [];
  creating = false;
  viewSection = false;

  complianceForm: FormGroup;
  standardSectionId: number;
  strategies = [];

  constructor(private cpControlConfigService: CpControlConfigService,
              public activeModal: NgbActiveModal,
              private cpReportBehavior: CpReportBehavior,
              private utmToastService: UtmToastService,
              public modalService: NgbModal,
              private fb: FormBuilder) {
  }

  ngOnInit() {
    this.complianceForm = this.fb.group({
      controlName: [this.control ? this.control.controlName : '' ,
        [Validators.required, Validators.minLength(10), Validators.maxLength(200)]],
      solution: [this.control ? this.control.controlSolution : '', [Validators.maxLength(2000)]],
      remediation: [this.control && this.control.controlRemediation ? this.control.controlRemediation : '', [Validators.maxLength(2000)]],
      strategy: [this.control ? this.control.controlStrategy as ComplianceStrategyEnum : null, [Validators.required]],
      queriesConfigs: this.fb.array(this.control ? this.control.queriesConfigs : [], [Validators.minLength(1)])
    });

    this.strategies = Object.keys(ComplianceStrategyEnum).map(
      key => ({
        id: key as ComplianceStrategyEnum,
        strategy: ComplianceStrategyEnum[key as keyof typeof ComplianceStrategyEnum]
      })
    );

    if (this.control) {
      this.viewSection = true;
      this.standardSectionId = this.control.standardSectionId;
    }
    this.loading = false;
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
    const controlConfigCompliance: ComplianceControlConfigType = {
      standardSectionId: this.standardSectionId,
      controlName: this.complianceForm.controls.controlName.value,
      controlSolution: this.complianceForm.controls.solution.value.replace(/\r?\n/g, '<br/>'),
      controlRemediation: this.complianceForm.controls.remediation.value.replace(/\r?\n/g, '<br/>'),
      controlStrategy: this.complianceForm.controls.strategy.value,
      queriesConfigs: this.queriesList
    };

    if (this.control) {
      controlConfigCompliance.id = this.control.id;
      this.cpControlConfigService.update(controlConfigCompliance).subscribe(() => {
        this.utmToastService.showSuccessBottom('Compliance report  edited successfully');
        this.activeModal.close();
        this.controlCreated.emit('edited');
      }, error1 => {
        this.creating = false;
        this.utmToastService.showError('Error', 'Error editing compliance report');
      });
    } else {
      this.cpControlConfigService.create(controlConfigCompliance).subscribe(() => {
        this.utmToastService.showSuccessBottom('Compliance report  created successfully');
        this.activeModal.close();
        this.cpReportBehavior.$reportUpdate.next('update');
        this.controlCreated.emit('created');
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
