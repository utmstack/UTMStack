import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {AbstractControl, FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {Rule} from '../../../models/rule.model';
import {AfterEventFormService} from "../../../services/after-event-form.service";
import {RuleService} from '../../../services/rule.service';


@Component({
  selector: 'app-after-event',
  templateUrl: './add-after-event.component.html',
  styleUrls: ['./add-after-event.component.css']
})
export class AddAfterEventComponent implements OnInit {
  @Input() form: FormGroup;
  @Input() rule: Rule;
  @Output() remove = new EventEmitter<void>();

  constructor(private fb: FormBuilder,
              private ruleService: RuleService,
              private afterEventService: AfterEventFormService) { }

  ngOnInit() {

  }

  get with(): FormArray {
    return this.form.get('with') as FormArray;
  }

  get or(): FormArray {
    return this.form.get('or') as FormArray;
  }

  addExpression() {
    this.with.push(this.fb.group({
      field: [''],
      operator: [''],
      value: ['']
    }));
  }

  removeExpression(index: number) {
    this.with.removeAt(index);
  }

  addOr() {
    this.or.push(this.afterEventService.buildSearchRequest(
      this.afterEventService.emptySearchRequest()
    ));
  }

  removeOr(index: number) {
    this.or.removeAt(index);
  }

  asFormGroup(control: AbstractControl): FormGroup {
    return this.ruleService.asFormGroup(control);
  }
}
