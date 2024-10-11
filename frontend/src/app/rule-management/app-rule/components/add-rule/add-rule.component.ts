import {Component, OnDestroy, OnInit} from '@angular/core';
import {
  FormArray,
  FormBuilder,
  FormGroup,
  Validators
} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {Observable} from 'rxjs';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {DataType, Mode, Rule, Variable} from '../../../models/rule.model';
import {DataTypeService} from '../../../services/data-type.service';
import {RuleService} from '../../../services/rule.service';

@Component({
  selector: 'app-add-rule',
  templateUrl: './add-rule.component.html',
  styleUrls: ['./add-rule.component.css'],
})
export class AddRuleComponent implements OnInit, OnDestroy {
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

  constructor(private fb: FormBuilder,
              private dataTypeService: DataTypeService,
              private ruleService: RuleService,
              private utmToastService: UtmToastService,
              public activeModal: NgbActiveModal) {
    this.initializeForm();
  }

  ngOnInit() {
    this.mode = this.rule ? 'EDIT' : 'ADD';
    this.initializeForm(this.rule);

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
      confidentiality: [rule ? rule.confidentiality : 0, [Validators.required, Validators.min(0), Validators.max(3)]],
      integrity: [rule ? rule.integrity : 0, [Validators.required, Validators.min(0), Validators.max(3)]],
      availability: [rule ? rule.availability : 0, [Validators.required, Validators.min(0), Validators.max(3)]],
      category: [rule ? rule.category : '', Validators.required],
      technique: [rule ? rule.technique : '', Validators.required],
      description: [rule ? rule.description : '', Validators.required]
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

  ngOnDestroy() {
    this.dataTypeService.resetTypes();
  }

  onChangeVariables(variables: any[]) {
    this.savedVariables = [...variables];
  }
}
