import {HttpResponse} from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { FormArray, FormBuilder, FormGroup, Validators } from '@angular/forms';
import {ActivatedRoute} from '@angular/router';
import {map, tap} from 'rxjs/operators';
import {DataType, Mode, Rule, Variable} from '../../../models/rule.model';

const variableTemplate = {get: ['', Validators.required] , as: ['', Validators.required] , of_type: ['', Validators.required]};

@Component({
    selector: 'app-add-rule',
    templateUrl: './add-rule.component.html',
    styleUrls: ['./add-rule.component.css']
})
export class AddRuleComponent implements OnInit {
    ruleForm: FormGroup;
    types: DataType[] = [
        { id: 1, name: 'Linux' }
    ];
    mode: Mode = 'ADD';

    constructor(private fb: FormBuilder,
                private route: ActivatedRoute) {
        this.initializeForm();
    }

    ngOnInit() {
        this.route.data
            .pipe(
                map((data: {response: HttpResponse<Rule>}) => data.response.body),
                tap((rule: Rule) => {
                    this.mode = 'EDIT';
                    this.initializeForm(rule);
                })
            ).subscribe();
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
        return this.ruleForm.get('definition').get('variables') as FormArray;
    }

    addVariable() {
        this.variables.push(this.fb.group(variableTemplate));
    }

    removeVariable(index: number) {
        this.variables.removeAt(index);
    }

    saveRule() {
        if (this.ruleForm.valid) {
            const newRule: Rule = this.ruleForm.value;
            console.log('Saving rule:', JSON.stringify(newRule));
        } else {
            console.error('Form is invalid. Cannot save rule.');
        }
    }

    addType(newType: DataType): DataType | null {
        const exists = this.types.find(type => type.name.toLowerCase() === newType.name.toLowerCase());
        if (!exists) {
            const newDataType: DataType = { id: this.types.length + 1, name: newType.name };
            this.types = [...this.types, newDataType];
            return newDataType;
        }
        return null;
    }

    initializeForm(rule?: Rule) {
        this.ruleForm = this.fb.group({
            dataTypes: [rule ? rule.data_types : '', Validators.required],
            name: [rule ? rule.name : '', Validators.required],
            impact: this.fb.group({
                confidentiality: [rule ? rule.impact.confidentiality : null, Validators.required],
                integrity: [rule ? rule.impact.integrity : null, Validators.required],
                availability: [rule ? rule.impact.availability : null, Validators.required]
            }),
            category: [rule ? rule.category : '', Validators.required],
            technique: [rule ? rule.technique : '', Validators.required],
            description: [rule ? rule.description : '', Validators.required],
            references: this.initReferences(rule ? rule.references : []),
            definition: this.fb.group({
                variables: this.initVariables(rule ? rule.definition.variables : []),
                expression: [rule ? rule.definition.expression : '', Validators.required]
            })
        });
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
        return this.fb.array(formArray.length > 0 ? formArray : [this.fb.group(variableTemplate)], this.minLengthArray(1));
    }

    minLengthArray(min: number) {
        return (control: FormArray): { [key: string]: boolean } | null => {
            if (control.length >= min) {
                return null;
            }
            return { minLengthArray: true };
        };
    }
}
