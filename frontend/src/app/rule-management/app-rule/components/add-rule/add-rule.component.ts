import { Component, OnInit } from '@angular/core';
import { FormArray, FormBuilder, FormGroup, Validators } from '@angular/forms';
import { DataType, Rule, Variable } from '../../../models/rule.model';

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

    constructor(private fb: FormBuilder) { }

    ngOnInit() {
        this.ruleForm = this.fb.group({
            /*id: [null, Validators.required],*/
            dataTypes: ['', Validators.required],
            name: ['', Validators.required],
            impact: this.fb.group({
                confidentiality: [null, Validators.required],
                integrity: [null, Validators.required],
                availability: [null, Validators.required]
            }),
            category: ['', Validators.required],
            technique: ['', Validators.required],
            description: ['', Validators.required],
            references: this.fb.array([this.fb.control('', Validators.required)], this.minLengthArray(1)),
            definition: this.fb.group({
                variables: this.fb.array([this.fb.group(variableTemplate)], this.minLengthArray(1)),
                expression: ['', Validators.required]
            })
        });
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

    minLengthArray(min: number) {
        return (control: FormArray): { [key: string]: boolean } | null => {
            if (control.length >= min) {
                return null;
            }
            return { minLengthArray: true };
        };
    }
}
