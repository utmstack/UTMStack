import {HttpResponse} from '@angular/common/http';
import {ChangeDetectorRef, Component, OnInit, OnDestroy} from '@angular/core';
import { FormArray, FormBuilder, FormGroup, Validators } from '@angular/forms';
import {ActivatedRoute, Router} from '@angular/router';
import { Observable } from 'rxjs';
import { filter, map, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {DataType, Mode, Rule, Variable} from '../../../models/rule.model';
import {DataTypeService} from '../../../services/data-type.service';
import {RuleService} from '../../../services/rule.service';

const variableTemplate = {get: ['', Validators.required] , as: ['', Validators.required] , of_type: ['', Validators.required]};

@Component({
    selector: 'app-add-rule',
    templateUrl: './add-rule.component.html',
    styleUrls: ['./add-rule.component.css'],
})
export class AddRuleComponent implements OnInit, OnDestroy {
    ruleForm: FormGroup;
    mode: Mode = 'ADD';
    loadingDataTypes = false;
    daTypeRequest: {page: number, size: number} = {
        page: -1,
        size: 10
    };
    types$: Observable<DataType[]>;
    isSubmitting = false;
    savedVariables = [];

    constructor(private fb: FormBuilder,
                private route: ActivatedRoute,
                private dataTypeService: DataTypeService,
                private ruleService: RuleService,
                private router: Router,
                private utmToastService: UtmToastService) {
        this.initializeForm();
    }

    ngOnInit() {
        this.route.data
            .pipe(
                filter((data: {response: HttpResponse<Rule>}) => !!data.response),
                map((data: {response: HttpResponse<Rule>}) => data.response.body),
                tap((rule: Rule) => {
                    this.mode = 'EDIT';
                    this.daTypeRequest = {
                        page: -1,
                        size: 10
                    };
                    this.initializeForm(rule);
                })
            ).subscribe();

        this.types$ = this.dataTypeService.type$;
        this.loadDataTypes();
    }

    removeReference(index: number) {
        this.references.removeAt(index);
    }

    onDataTypeChange(selectedDataTypes: DataType[]) {
        this.ruleForm.get('dataTypes').patchValue(selectedDataTypes);
    }

    get references() {
        return this.ruleForm.get('references') as FormArray;
    }

    addReference() {
        this.references.push(this.fb.control('', Validators.required));
    }

    get variables() {
        return this.ruleForm.get('definition').get('ruleVariables') as FormArray;
    }

    addVariable() {
        this.variables.push(this.fb.group(variableTemplate));
    }

    removeVariable(index: number) {
        this.variables.removeAt(index);
    }

    saveRule() {
        if (this.ruleForm.valid) {
            const variables = this.savedVariables.map(variable => ({
                as: variable.as,
                get: variable.get,
                of_type: variable.of_type
            }));
            const expression =
            this.isSubmitting = true;
            const rule: Rule = {
                ...this.ruleForm.value,
            };
            rule.definition.ruleVariables = variables;
            this.ruleService.saveRule(this.mode, rule)
                .subscribe({
                    next: response => {
                        console.log('Rule saved successfully', response);
                        this.dataTypeService.resetTypes();
                        this.isSubmitting = false;
                        this.router.navigate(['/alerting-rules/rules'])
                            .then(() =>  this.utmToastService.showSuccessBottom(this.mode === 'ADD'
                                ? 'Rule saved successfully' : 'Rule edited successfully'));
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
            confidentiality: [rule ? rule.confidentiality : null, Validators.required],
            integrity: [rule ? rule.integrity : null, Validators.required],
            availability: [rule ? rule.availability : null, Validators.required],
            category: [rule ? rule.category : '', Validators.required],
            technique: [rule ? rule.technique : '', Validators.required],
            description: [rule ? rule.description : '', Validators.required],
            references: this.initReferences(rule ? rule.references : []),
            definition: this.fb.group({
                ruleVariables: rule ? this.fb.array([]) : this.fb.array([this.fb.group(variableTemplate)]),
                ruleExpression: [rule ? rule.definition.ruleExpression : '', Validators.required]
            })
        });

        this.savedVariables = rule && rule.definition ? rule.definition.ruleVariables : [];
    }

    initReferences(references: string[]): FormArray {
        const formArray = references.map(reference => this.fb.control(reference, Validators.required));
        return this.fb.array(formArray.length > 0 ? formArray : [this.fb.control('', Validators.required)], this.minLengthArray(1));
    }

    initVariables(variables: Variable[]): FormArray {
        const formArray = variables.map(variable => this.fb.group({
            get: [variable.get, Validators.required],
            as: [variable.as, Validators.required],
            of_type: [variable.of_type, Validators.required]
        }));
        return this.fb.array(formArray.length > 0 ? formArray : [this.fb.group(variableTemplate)]);
    }

    minLengthArray(min: number) {
        return (control: FormArray): { [key: string]: boolean } | null => {
            if (control.length >= min) {
                return null;
            }
            return { minLengthArray: true };
        };
    }

    loadDataTypes() {
        this.daTypeRequest.page = this.daTypeRequest.page + 1;
        this.loadingDataTypes = true;

        this.dataTypeService.getAll(this.daTypeRequest)
            .subscribe( data => {
                this.loadingDataTypes = false;
            });
    }

    trackByFn(type: DataType) {
        return type.id;
    }

    saveVariable(index: number): void {
        const variable = { ...this.variables.at(index).value, isEditing: false };
        this.savedVariables.push(variable);
        this.variables.removeAt(index);
    }

    editVariable(index: number): void {
        const variable = this.savedVariables[index];
        if (this.variables.length === 0) {
            this.addVariable();
        }
        this.variables.at(0).patchValue(variable);
        this.savedVariables.splice(index, 1);
    }

    updateVariable(index: number): void {
        this.savedVariables[index].isEditing = false;
    }

    copyVariable(variableName: string | number) {
        const expressionControl = this.ruleForm.get('definition').get('ruleExpression');
        expressionControl.setValue(expressionControl.value + variableName);
    }

    ngOnDestroy() {
        this.dataTypeService.resetTypes();
    }
}
