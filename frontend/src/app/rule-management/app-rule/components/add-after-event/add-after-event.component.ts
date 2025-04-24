import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {AbstractControl, FormArray, FormBuilder, FormGroup} from '@angular/forms';
import {Rule} from '../../../models/rule.model';
import {AfterEventFormService} from '../../../services/after-event-form.service';
import {RuleService} from '../../../services/rule.service';
import {HttpResponse} from '@angular/common/http';
import {IndexPatternService} from '../../../../shared/services/elasticsearch/index-pattern.service';
import {Observable, of} from 'rxjs';
import {UtmIndexPattern} from '../../../../shared/types/index-pattern/utm-index-pattern';
import {catchError, map} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {FieldDataService} from '../../../../shared/services/elasticsearch/field-data.service';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';


@Component({
  selector: 'app-after-event',
  templateUrl: './add-after-event.component.html',
  styleUrls: ['./add-after-event.component.css']
})
export class AddAfterEventComponent implements OnInit {
  @Input() form: FormGroup;
  @Input() rule: Rule;
  @Output() remove = new EventEmitter<void>();
  patterns$: Observable<UtmIndexPattern[]>;
  fields$: Observable<ElasticSearchFieldInfoType[]>;
  operators =  [
    {label: 'equals', value: 'eq'},
    {label: 'not equals', value: 'neq'}
  ];

  constructor(private fb: FormBuilder,
              private ruleService: RuleService,
              private afterEventService: AfterEventFormService,
              private indexPatternService: IndexPatternService,
              private toastService: UtmToastService,
              private fieldDataService: FieldDataService) { }

  ngOnInit() {

    this.patterns$ = this.indexPatternService.query(
      {
        page: 0,
        size: 1000,
        sort: 'id,asc',
        'isActive.equals': true,
      }
    ).pipe(
      map((res: HttpResponse<UtmIndexPattern[]>) => res.body),
      catchError((err: HttpResponse<UtmIndexPattern>) => {
        this.toastService.showError('Error', 'Failed to load index patterns');
        return of([] as UtmIndexPattern[]);
      })
    );
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

  changeIndexPattern(indexPattern: string) {
    this.fields$ = this.fieldDataService.getFields(indexPattern).pipe(
      map((fields) => fields || []),
      catchError((error) => {
        this.toastService.showError('Error', 'Failed to load fields');
        return of([]);
      })
    );
  }
}
