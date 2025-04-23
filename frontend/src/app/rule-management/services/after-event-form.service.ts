// search-request-form.service.ts
import { Injectable } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Expression, SearchRequest } from '../models/rule.model';

@Injectable({
  providedIn: 'root'
})
export class AfterEventFormService {

  constructor(private fb: FormBuilder) {}

  buildExpression(expr: Expression): FormGroup {
    return this.fb.group({
      field: [expr.field || '', Validators.required],
      operator: [expr.operator || '', Validators.required],
      value: [expr.value || '', Validators.required]
    });
  }

  buildSearchRequest(event: SearchRequest): FormGroup {
    return this.fb.group({
      indexPattern: [event.indexPattern || '', Validators.required],
      with: this.fb.array(
        event.with.length
          ? event.with.map(w => this.buildExpression(w))
          : []
      ),
      or: this.fb.array(
        event.or.length
          ? event.or.map(subEvent => this.buildSearchRequest(subEvent))
          : []
      ),
      within: [event.within || ''],
      count: [event.count ? event.count : null]
    });
  }

  emptySearchRequest(): SearchRequest {
    return {
      indexPattern: '',
      with: [],
      or: [],
      within: '',
      count: null
    };
  }
}
