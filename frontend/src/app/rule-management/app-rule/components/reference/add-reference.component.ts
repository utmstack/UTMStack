import {Component, Input, OnInit} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from "@angular/forms";
import {urlValidator} from "../../../custom-validators";
import {Rule} from "../../../models/rule.model";

@Component({
  selector: 'app-references',
  templateUrl: './add-reference.component.html',
  styleUrls: ['./add-reference.component.css']
})
export class AddReferenceComponent implements OnInit {
  @Input() formGroup: FormGroup;
  @Input() rule: Rule;
  constructor(private fb: FormBuilder) { }

  ngOnInit() {
    this.init();
  }

  init(){
    this.formGroup.setControl('references', this.initReferences(this.rule ? this.rule.references : []));
  }

  initReferences(references: string[]): FormArray {
    const formArray = references.map(reference => this.fb.control(reference,
      [Validators.required, urlValidator]));

    return this.fb.array(formArray.length > 0 ? formArray : [this.fb.control('', [Validators.required, urlValidator])],
      this.minLengthArray(1));
  }

  minLengthArray(min: number) {
    return (control: FormArray): { [key: string]: boolean } | null => {
      if (control.length >= min) {
        return null;
      }
      return {minLengthArray: true};
    };
  }

  get references() {
    return this.formGroup.get('references') as FormArray;
  }

  addReference() {
    this.references.push(this.fb.control('', [Validators.required, urlValidator]));
  }

  removeReference(index: number) {
    this.references.removeAt(index);
  }

}
