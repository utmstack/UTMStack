import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from "@angular/forms";
import {singleTermValidator, variableTemplate} from "../../../custom-validators";
import {ElasticSearchFieldInfoType} from "../../../../shared/types/elasticsearch/elastic-search-field-info.type";
import {VariableDataType} from "../../../models/rule.constant";
import {Observable} from "rxjs";
import {FieldDataService} from "../../../../shared/services/elasticsearch/field-data.service";
import {Rule} from "../../../models/rule.model";


@Component({
  selector: 'app-variables',
  templateUrl: './add-variable.component.html',
  styleUrls: ['./add-variable.component.css']
})
export class AddVariableComponent implements OnInit {
  @Input() formGroup: FormGroup;
  @Input() rule: Rule;
  @Output() variablesEmitter = new EventEmitter<any[]>();
  savedVariables = [];
  variablesDataType = VariableDataType;
  fields$: Observable<ElasticSearchFieldInfoType[]>;
  constructor(private fb: FormBuilder,
              private fieldDataService: FieldDataService) { }

  ngOnInit() {
    this.fields$ = this.getFields('log-*');
    this.savedVariables = this.rule && this.rule.definition ? this.rule.definition.ruleVariables : [];

    this.formGroup.setControl('definition', this.fb.group({
      ruleVariables: this.fb.array([this.fb.group(variableTemplate)]),
      ruleExpression: [this.rule ? this.rule.definition.ruleExpression : '', Validators.required]
    }));
  }

  getFields(indexPattern: string) {
    return this.fieldDataService.getFields(indexPattern);
  }

  get variables() {
    return this.formGroup.get('definition').get('ruleVariables') as FormArray;
  }

  addVariable() {
    this.variables.push(this.fb.group(variableTemplate));
  }

  removeVariable(index: number) {
    this.variables.removeAt(index);
  }

  onFieldChange(selectField: string) {
    const variable = this.variables.at(0);
    const varFormGroup = variable.get('as');
    if(!!variable.get('get').value){
      varFormGroup.setValidators([Validators.required, singleTermValidator()]);
      varFormGroup.updateValueAndValidity();
    } else {
      varFormGroup.setValidators([]);
    }
  }

  saveVariable(index: number): void {
    const variable = {...this.variables.at(index).value, isEditing: false};
    this.savedVariables.push(variable);
    this.variables.removeAt(index);
    this.variablesEmitter.emit(this.savedVariables);
    this.addVariable();
  }

  editVariable(index: number): void {
    const variable = this.savedVariables[index];
    if (this.variables.length === 0) {
      this.addVariable();
    }
    this.variables.at(0).patchValue(variable);
    this.savedVariables.splice(index, 1);
  }

  copyVariable(variableName: string | number) {
    const expressionControl = this.formGroup.get('definition').get('ruleExpression');
    expressionControl.setValue(expressionControl.value + ' ' + variableName);
  }

  trackByFnField(field: ElasticSearchFieldInfoType) {
    return field.name;
  }

  removeSavedVariable(index: number) {
    this.savedVariables = this.savedVariables.filter((v, i) => i !== index);
    this.variablesEmitter.emit(this.savedVariables);
  }

}
