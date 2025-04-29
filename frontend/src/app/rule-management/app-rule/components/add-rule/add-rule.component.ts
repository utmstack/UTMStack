import {Component, OnDestroy, OnInit} from '@angular/core';
import {AbstractControl, FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {forkJoin, Observable} from 'rxjs';
import {map} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {AddRuleStepEnum, DataType, Mode, Rule, SearchRequest} from '../../../models/rule.model';
import {AfterEventFormService} from '../../../services/after-event-form.service';
import {DataTypeService} from '../../../services/data-type.service';
import {RuleService} from '../../../services/rule.service';

@Component({
  selector: 'app-add-rule',
  templateUrl: './add-rule.component.html',
  styleUrls: ['./add-rule.component.scss'],
})
export class AddRuleComponent implements OnInit, OnDestroy {
  RULE_FORM = AddRuleStepEnum;
  ruleForm: FormGroup;
  mode: Mode = 'ADD';
  loadingDataTypes = false;
  daTypeRequest: { page: number, size: number, sort: string } = {
    page: -1,
    size: 10,
    sort: 'dataType,ASC'
  };
  types$: Observable<DataType[]>;
  isSubmitting = false;
  savedVariables = [];
  rule: Rule;
  loading: false;
  currentStep: AddRuleStepEnum;
  stepCompleted: number[] = [];
  adversaryTypes = [
    {
      id: 'ORIGIN',
      name: 'origin'
    },
    {
      id: 'TARGET',
      name: 'target'
    }
];
  steps = [
    {
      id: AddRuleStepEnum.STEP0,
      label: 'Import',
      icon: 'icon-upload4'
    },
    {
      id: AddRuleStepEnum.STEP1,
      label: 'General Information',
      icon: 'icon-file-text'
    },
    {
      id: AddRuleStepEnum.STEP2,
      label: 'Conditions',
      icon: 'icon-cog'
    },
    {
      id: AddRuleStepEnum.STEP3,
      label: 'Post-Event Actions',
      icon: 'icon-loop'
    }
  ];


  constructor(private fb: FormBuilder,
              private dataTypeService: DataTypeService,
              private ruleService: RuleService,
              private utmToastService: UtmToastService,
              public activeModal: NgbActiveModal,
              private afterEventService: AfterEventFormService) {
  }

  ngOnInit() {
    this.currentStep = this.mode !== 'IMPORT' ? AddRuleStepEnum.STEP1 : AddRuleStepEnum.STEP0;

    if (this.mode !== 'IMPORT') {
      this.steps = this.steps.filter(step => step.id !== AddRuleStepEnum.STEP0);
      this.initializeForm(this.rule);
    }

    this.types$ = this.dataTypeService.type$;
    this.loadDataTypes();
  }

  onDataTypeChange(selectedDataTypes: DataType[]) {
    this.ruleForm.get('dataTypes').patchValue(selectedDataTypes);
    this.dataTypeService.resetTypes();
    this.daTypeRequest.page = -1;
    this.loadDataTypes();
  }

  saveRule() {
    if (this.ruleForm.valid) {
      const variables = this.savedVariables .length > 0 ?  this.savedVariables.map(variable => ({
        as: variable.as,
        get: variable.get,
        of_type: variable.of_type
      })) : [];
      this.isSubmitting = true;
      const rule: Rule = {
        ...this.ruleForm.value,
        dataTypes: this.getDataTypes(this.ruleForm.value.dataTypes)
      };
      rule.definition.ruleVariables = variables;
      this.ruleService.saveRule(this.mode, rule)
        .subscribe({
          next: response => {
            this.dataTypeService.resetTypes();
            this.isSubmitting = false;
            this.utmToastService.showSuccessBottom(this.mode === 'ADD'
              ? 'Rule saved successfully' : 'Rule edited successfully');
            this.activeModal.close(true);
          },
          error: err => {
            this.isSubmitting = false;
            this.utmToastService.showError('Error', this.mode === 'ADD'
              ? 'Error saving rule' : 'Error editing rule');
            console.error('Error saving rule:', err.message);
          }
        });
    } else {
      console.error('Form is invalid. Cannot save rule.');
    }
  }

  initializeForm(rule?: Rule) {
    this.ruleForm = this.fb.group({
      id: [rule ? rule.id : ''],
      dataTypes: [rule ? rule.dataTypes : '', Validators.required],
      name: [rule ? rule.name : '', Validators.required],
      adversary: [rule ? rule.adversary : '', Validators.required],
      confidentiality: [rule ? rule.confidentiality : 0, [Validators.required, Validators.min(0), Validators.max(3)]],
      integrity: [rule ? rule.integrity : 0, [Validators.required, Validators.min(0), Validators.max(3)]],
      availability: [rule ? rule.availability : 0, [Validators.required, Validators.min(0), Validators.max(3)]],
      category: [rule ? rule.category : '', Validators.required],
      technique: [rule ? rule.technique : '', Validators.required],
      description: [rule ? rule.description : '', Validators.required],
      systemOwner: [rule ? rule.systemOwner : false],
      deduplicateBy: [rule ? rule.deduplicateBy || [] : []],
      afterEvents: this.fb.array(
        rule && rule.afterEvents && rule.afterEvents.length
          ? rule.afterEvents.map(event => this.buildSearchRequest(event))
          : [this.afterEventService.buildEmptySearchRequest()]
      )
    });
    this.savedVariables = rule ? rule.definition.ruleVariables : [];
  }


  loadDataTypes() {
    this.daTypeRequest.page = this.daTypeRequest.page + 1;
    this.loadingDataTypes = true;

    this.dataTypeService.getAll(this.daTypeRequest)
      .subscribe(data => {
        this.loadingDataTypes = false;
      });
  }

  trackByFn(type: DataType) {
    return type.id;
  }

  onSearch(event: { term: string; items: any[] }) {
    this.dataTypeService.resetTypes();
    const request = {
      search: event.term
    };

    this.dataTypeService.getAll(request)
      .subscribe(data => {
        this.loadingDataTypes = false;
      });
  }

  onChangeVariables(variables: any[]) {
    this.savedVariables = [...variables];
  }

  next() {
    this.stepCompleted.push(this.currentStep);
    switch (this.currentStep) {
      case 0: this.currentStep = AddRuleStepEnum.STEP1;
              break;
      case 1: this.currentStep = AddRuleStepEnum.STEP2;
              break;
      case 2: this.currentStep = AddRuleStepEnum.STEP3;
              break;
    }
  }

  back(){
    this.stepCompleted.pop();
    switch (this.currentStep) {
      case 3: this.currentStep = AddRuleStepEnum.STEP2;
              break;
      case 2: this.currentStep = AddRuleStepEnum.STEP1;
              break;
      case 1: this.currentStep = AddRuleStepEnum.STEP0;
              break;
    }
  }

  isCompleted(step: number) {
    return this.stepCompleted.findIndex(value => value === step) !== -1;
  }

  isValidStep(step: number) {
    switch (step) {
      case AddRuleStepEnum.STEP1:
        return this.ruleForm.get('dataTypes').valid &&
         this.ruleForm.get('name').valid &&
         this.ruleForm.get('confidentiality').valid &&
         this.ruleForm.get('integrity').valid &&
         this.ruleForm.get('availability').valid &&
         this.ruleForm.get('category').valid &&
         this.ruleForm.get('technique').valid &&
         this.ruleForm.get('references').valid &&
         this.ruleForm.get('description').valid;

      case AddRuleStepEnum.STEP2:
        return this.ruleForm.get('definition').valid;

      case AddRuleStepEnum.STEP3:
        return this.ruleForm.get('afterEvents').valid;
    }
  }

  onFileChange($event: any): void {
    if ($event.length > 0 ) {
      if ($event[0].dataTypes) {
        forkJoin(
          $event[0].dataTypes.map((dt: string) =>
            this.dataTypeService.getAll({ search: dt }).pipe(
              map(res => res.body.length > 0 ? res.body[0] : null)
            )
          )
        ).subscribe(filteredDataTypes => {
          const dataTypes = filteredDataTypes.filter(dt => !!dt);
          this.rule = {
            ...$event[0],
            dataTypes: dataTypes ? dataTypes : []
          };
          this.initializeForm(this.rule);
        });
      } else {
        this.rule = {
          ...$event[0]
        };
        this.initializeForm(this.rule);
      }
    } else {
      this.mode = 'ERROR';
    }
  }

  onLoadError(){
    this.mode = 'ERROR';
  }

  getDataTypes(dataTypes: DataType[]) {
    return dataTypes.map( d => {
       if (!d.id) {
         return {
           ...d,
           dataTypeName: d.dataType,
           systemOwner: false,
           included: true
         };
       } else {
         return d;
       }
    });
  }

  buildSearchRequest(event: SearchRequest): FormGroup {
    return this.afterEventService.buildSearchRequest(event);
  }

  get afterEvents(): FormArray {
    return this.ruleForm.get('afterEvents') as FormArray;
  }

  addAfterEvent() {
    this.afterEvents.push(this.buildSearchRequest(this.afterEventService.emptySearchRequest()));
  }

  removeAfterEvent(index: number) {
    this.afterEvents.removeAt(index);
  }

  asFormGroup(control: AbstractControl): FormGroup {
    return this.ruleService.asFormGroup(control);
  }

  ngOnDestroy() {
    this.dataTypeService.resetTypes();
  }
}
