import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbTimeStruct} from '@ng-bootstrap/ng-bootstrap';
import {Subject} from 'rxjs';
import {map, takeUntil} from 'rxjs/operators';
import {FILTER_OPERATORS} from '../../../../../constants/filter-operators.const';
import {ElasticDataTypesEnum} from '../../../../../enums/elastic-data-types.enum';
import {ElasticOperatorsEnum} from '../../../../../enums/elastic-operators.enum';
import {ElasticSearchIndexService} from '../../../../../services/elasticsearch/elasticsearch-index.service';
import {FieldDataService} from '../../../../../services/elasticsearch/field-data.service';
import {ElasticSearchFieldInfoType} from '../../../../../types/elasticsearch/elastic-search-field-info.type';
import {ElasticFilterType} from '../../../../../types/filter/elastic-filter.type';
import {OperatorsType} from '../../../../../types/filter/operators.type';
import {TimeFilterType} from '../../../../../types/time-filter.type';
import {resolveIcon} from '../../../../../util/elastic-fields.util';
import {OperatorService} from '../shared/util/operator.service';

@Component({
  selector: 'app-elastic-filter-add',
  templateUrl: './elastic-filter-add.component.html',
  styleUrls: ['./elastic-filter-add.component.scss']
})
export class ElasticFilterAddComponent implements OnInit {
  /**
   * Index Pattern
   */
  @Input() pattern: string;
  /**
   * Event output
   */
  @Output() filterChange = new EventEmitter<ElasticFilterType>();
  /**
   * Filter to edit
   */
  @Input() filter: ElasticFilterType;

  @Input() hiddenFields: string[] = [];
  /**
   * All operators
   */
  operators: OperatorsType[] = FILTER_OPERATORS;
  operatorEnum = ElasticOperatorsEnum;
  fields: ElasticSearchFieldInfoType[] = [];
  loading = true;
  /**
   * Values of field
   */
  fieldValues: string[] = [];
  formFilter: FormGroup;
  /**
   * Determine if select value cant select multiples values or not
   */
  multiple = false;
  valueFrom: any;
  valueTo: any;
  fieldTypes = ElasticDataTypesEnum;
  time: NgbTimeStruct = {hour: 0, minute: 0, second: 0};
  loadingValues = false;
  selectableOperators = [ElasticOperatorsEnum.IS_ONE_OF,
    ElasticOperatorsEnum.IS_NOT_ONE_OF,
    ElasticOperatorsEnum.CONTAIN_ONE_OF,
    ElasticOperatorsEnum.DOES_NOT_CONTAIN_ONE_OF];
  destroy$ = new Subject<void>();

  constructor(private fb: FormBuilder,
              private fieldDataBehavior: FieldDataService,
              private elasticSearchIndexService: ElasticSearchIndexService,
              private operatorService: OperatorService) {
  }

  ngOnInit() {
    this.initFormFilter();
    // The `event` field (flat_object bag) is listed like any other field; the
    // field picker also allows typing an `event.<leaf>` dot-path (addTag), and
    // typed/selected flattened fields get the flattened operator set below.
    this.fieldDataBehavior.getFields(this.pattern)
      .pipe(takeUntil(this.destroy$),
        map(fields => fields.filter(f => !this.hiddenFields.includes(f.name))))
      .subscribe(field => {
        if (field) {
          this.fields = field;
          this.loading = false;
          if (this.filter) {
            this.setFilterEdit();
          }
        }
      });
    this.formFilter.get('field').valueChanges.subscribe(val => {
      this.getFieldValues();
    });
  }

  initFormFilter() {
    this.formFilter = this.fb.group({
      field: ['', Validators.required],
      operator: ['', Validators.required],
      value: [],
      status: ['ACTIVE', Validators.required],
    });
  }

  /**
   * Edit Filter
   */
  setFilterEdit() {
    this.formFilter.patchValue(this.filter);
    this.getOperators();
    this.isMultipleSelectValue();
    if (this.applySelectFilter()) {
      if (this.field.type === ElasticDataTypesEnum.DATE ||
        (this.field.type === ElasticDataTypesEnum.TEXT && !this.field.name.includes('.keyword'))) {
        this.fieldValues = this.filter.value;
      } else {
        this.getFieldValues();
      }
    }
    if (this.filter.operator === ElasticOperatorsEnum.IS_BETWEEN || this.filter.operator === ElasticOperatorsEnum.IS_NOT_BETWEEN) {
      this.valueFrom = this.filter.value[0];
      this.valueTo = this.filter.value[1];
    }
  }

  addFilter() {
    if (this.formFilter.get('operator').value === ElasticOperatorsEnum.IS_BETWEEN ||
      this.formFilter.get('operator').value === ElasticOperatorsEnum.IS_NOT_BETWEEN) {
      this.setValueRange();
    }

    this.filterChange.emit(this.formFilter.value);
  }

  /**
   * On operator clicked, determine if cant get field values or not
   * @param $event Operator
   */
  selectOperator($event) {
    if (this.formFilter.get('operator').value === this.operatorEnum.IS_ONE_OF
      || this.formFilter.get('operator').value === this.operatorEnum.IS_NOT_ONE_OF) {
      this.addValidatorToValue();
      // Only get values of field that are atomic(keyword, number)
      if (this.field.type === ElasticDataTypesEnum.DATE ||
        (this.field.type === ElasticDataTypesEnum.TEXT && !this.field.name.includes('.keyword'))) {
        this.fieldValues = [];
      } else {
        this.getFieldValues();
      }
    } else {
      this.cancelValidatorToValue();
    }
  }

  /**
   * Add validator to input
   */
  addValidatorToValue() {
    this.formFilter.get('value').setValidators(Validators.required);
    this.formFilter.get('value').updateValueAndValidity();
    this.formFilter.updateValueAndValidity();
  }

  /**
   * Clear validator to input
   */
  cancelValidatorToValue() {
    this.formFilter.get('value').setValidators(null);
    this.formFilter.get('value').updateValueAndValidity();
    this.formFilter.updateValueAndValidity();
  }

  getFieldValues() {
    this.loadingValues = true;
    const req = {
      page: 0,
      size: 10,
      indexPattern: this.pattern,
      keyword: this.formFilter.get('field').value
    };
    this.elasticSearchIndexService.getElasticFieldValues(req)
      .subscribe(res => {
          this.fieldValues = res.body;
          this.loadingValues = false;
        },
        error => {
          this.fieldValues = [];
          this.loadingValues = false;
        });
  }

  /**
   * Resolve the selected field to a field-info object. A typed `event.<leaf>`
   * dot-path (allowed via addTag) is not in the server field list, so synthesize
   * a FLATTENED field for it — the backend translates `event.*` to flat_object
   * queries and only a subset of operators is valid.
   */
  get field(): ElasticSearchFieldInfoType {
    const value = this.formFilter.get('field').value;
    const index = this.fields.findIndex(f => f.name === value);
    if (index !== -1) {
      return this.fields[index];
    }
    if (typeof value === 'string' && value.startsWith('event.')) {
      return {name: value, type: ElasticDataTypesEnum.FLATTENED};
    }
  }

  /**
   * Determine the type string of the selected field (handles typed event.* paths).
   */
  extractFieldDataType(): string {
    const field = this.field;
    return field ? field.type : null;
  }

  onDateRangeChange($event: TimeFilterType) {
    this.formFilter.get('value').setValue([$event.timeFrom, $event.timeTo]);
  }

  /**
   * When change field, refresh operator based on field data type, refresh multiple bar
   * @param $event Field
   */
  changeField($event) {
    // this.formFilter.get('field').setValue($event.name);
    this.formFilter.get('value').setValue(null);
    this.formFilter.get('operator').setValue(null);
    this.getOperators();
    this.isMultipleSelectValue();
  }

  /**
   * When change operator set value to null, and set current multiple var
   * @param $event Operator
   */
  onOperatorChange($event) {
    this.formFilter.get('value').setValue(null);
    this.isMultipleSelectValue();
  }

  /**
   * Return boolean, if show input or select values
   */
  applySelectFilter(): boolean {
    const fieldSelected = this.formFilter.get('field').value;
    // if field exist
    if (this.field) {
      if (this.field.type === ElasticDataTypesEnum.TEXT ||
        this.field.type === ElasticDataTypesEnum.STRING ||
        /* flattened fields (the event bag): distinct values are unavailable by design
        (terms aggs are not supported on flattened), so multi-value ops rely on addTag */
        this.field.type === ElasticDataTypesEnum.FLATTENED) {
        /* if fields is type string or text determine if field is a keyword or not, if field is keyword return
        result of function operatorFieldSelectable() that return if current operator cant apply select or input
        */
        if (fieldSelected.includes('.keyword')) {
          return this.operatorFieldSelectable();
        } else {
          // if type of current filter is not keyword return result of validation if current operator is selectable or not
          return this.selectableOperators.includes(this.formFilter.get('operator').value);
        }
      } else if (this.field.type === ElasticDataTypesEnum.DATE) {
        return this.formFilter.get('operator').value === ElasticOperatorsEnum.IS_ONE_OF ||
          this.formFilter.get('operator').value === ElasticOperatorsEnum.IS_NOT_ONE_OF;
      } else {
        // if current field is not a date or text return result of function if field filter value cant show select or input
        return this.operatorFieldSelectable();
      }
    } else {
      return false;
    }
  }

  /**
   * Return boolean, using to determine if current operator is selectable or not
   */
  operatorFieldSelectable(): boolean {
    return this.operator === this.operatorEnum.IS || this.operator === this.operatorEnum.IS_NOT ||
      this.operator === this.operatorEnum.IS_ONE_OF || this.operator === this.operatorEnum.IS_NOT_ONE_OF ||
      this.operator === this.operatorEnum.CONTAIN_ONE_OF || this.operator === this.operatorEnum.DOES_NOT_CONTAIN_ONE_OF;
  }

  /**
   * Return current operator
   */
  get operator(): ElasticOperatorsEnum {
    return this.formFilter.get('operator').value;
  }

  applyInputFilter(): boolean {
    const field = this.field;
    return !!field && (field.type === ElasticDataTypesEnum.TEXT || field.type === ElasticDataTypesEnum.STRING);
  }

  /**
   * Return index of field selected in the server field list (typed event.* paths
   * are not in the list, so this returns -1 for them).
   */
  getIndexField(): number {
    return this.fields.findIndex(value => value.name === this.formFilter.get('field').value);
  }

  getOperators() {
    const field = this.field;
    if (field) {
      this.operators = this.operatorService.getOperators(field, this.operators);
    }
  }

  /**
   * Determine if a field select value cant choose multiple values or not
   */
  isMultipleSelectValue() {
    this.multiple = this.formFilter.get('operator').value === ElasticOperatorsEnum.IS_ONE_OF ||
      this.formFilter.get('operator').value === ElasticOperatorsEnum.IS_NOT_ONE_OF ||
      this.formFilter.get('operator').value === ElasticOperatorsEnum.CONTAIN_ONE_OF ||
      this.formFilter.get('operator').value === ElasticOperatorsEnum.DOES_NOT_CONTAIN_ONE_OF;
  }

  /**
   * Set a filter value with a range
   */
  setValueRange() {
    this.formFilter.get('value').setValue([this.valueFrom, this.valueTo]);
  }


  resolveIcon(field: ElasticSearchFieldInfoType): string {
    return resolveIcon(field);
  }
}
