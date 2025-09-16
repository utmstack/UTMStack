import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {AbstractControl, FormArray, FormBuilder, FormGroup} from '@angular/forms';
import {Observable, of} from 'rxjs';
import {catchError, map} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ElasticDataTypesEnum} from '../../../../shared/enums/elastic-data-types.enum';
import {FieldDataService} from '../../../../shared/services/elasticsearch/field-data.service';
import {IndexPatternService} from '../../../../shared/services/elasticsearch/index-pattern.service';
import {ElasticSearchFieldInfoType} from '../../../../shared/types/elasticsearch/elastic-search-field-info.type';
import {UtmIndexPattern} from '../../../../shared/types/index-pattern/utm-index-pattern';
import {Rule} from '../../../models/rule.model';
import {AfterEventFormService} from '../../../services/after-event-form.service';
import {RuleService} from '../../../services/rule.service';
import {getOperator} from "../../../../shared/util/query-params-to-filter.util";


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
  fields: ElasticSearchFieldInfoType[] = [] as ElasticSearchFieldInfoType[];
  allOperators = {
    keyword: [
      { label: 'filter term', value: 'filter_term' },
      { label: 'must not term', value: 'must_not_term' }
    ],
    text: [
      { label: 'filter match', value: 'filter_match' },
      { label: 'must not match', value: 'must_not_match' }
    ]
  };

  operators = [];

  constructor(private fb: FormBuilder,
              private ruleService: RuleService,
              private afterEventService: AfterEventFormService,
              private indexPatternService: IndexPatternService,
              private toastService: UtmToastService,
              private fieldDataService: FieldDataService) { }

  ngOnInit() {

    if (this.form.get('indexPattern').value) {
      this.changeIndexPattern(this.form.get('indexPattern').value);
    }

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
    this.fieldDataService.getFields(indexPattern).pipe(
      map((fields) => fields || []),
      catchError((error) => {
        this.toastService.showError('Error', 'Failed to load fields');
        return of([]);
      })
    ).subscribe(fields => this.fields = fields);
  }

  getOperators(field: ElasticSearchFieldInfoType) {
    if (!field) {return; }
    const fieldName = field.name || '';
    const hasKeyword = fieldName.includes('.keyword');
    const isNumeric = field.type === ElasticDataTypesEnum.NUMBER || field.type === ElasticDataTypesEnum.LONG
      || field.type === ElasticDataTypesEnum.FLOAT;

    return hasKeyword || isNumeric ? this.allOperators.keyword : this.allOperators.text;
  }

  onFieldChange($event: ElasticSearchFieldInfoType, index: number) {
    const control = this.with.at(index);
    control.get('operator').reset();
  }

  getOperatorByField(fieldName: string) {
    const field = this.fields.find(f => f.name === fieldName);
    //if the field doesnt appear on field list, it must be an added field ( so we'll treat it like a STRING)
    return this.getOperators(field || {name:fieldName,type:ElasticDataTypesEnum.STRING});
  }

  addField(name:string){
    return {name,type:ElasticDataTypesEnum.STRING};
  }


}
