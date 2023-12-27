import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbDate, NgbTimeStruct} from '@ng-bootstrap/ng-bootstrap';
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
  timeValue: NgbDate;
  time: NgbTimeStruct = {hour: 0, minute: 0, second: 0};
  loadingValues = false;

  constructor(private fb: FormBuilder,
              private fieldDataBehavior: FieldDataService,
              private elasticSearchIndexService: ElasticSearchIndexService) {
  }

  ngOnInit() {
    this.initFormFilter();
    this.fieldDataBehavior.getFields(this.pattern).subscribe(field => {
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
    this.elasticSearchIndexService.getElasticFieldValues(req).subscribe(res => {
      this.fieldValues = res.body;
      this.loadingValues = false;
    });
  }

  /**
   * Return field data type
   */
  extractFieldDataType(): string {
    const field = this.formFilter.get('field').value;
    const index = this.fields.findIndex(value => value.name === field);
    return this.fields[index].type;
  }

  onDateRangeChange($event: TimeFilterType) {
    this.formFilter.get('value').setValue([$event.timeFrom, $event.timeTo]);
  }

  /**
   * When change field, refresh operator based on field data type, refresh multiple bar
   * @param $event Field
   */
  changeField($event) {
    this.formFilter.get('field').setValue($event.name);
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
        this.field.type === ElasticDataTypesEnum.STRING) {
        /* if fields is type string or text determine if field is a keyword or not, if field is keyword return
        result of function operatorFieldSelectable() that return if current operator cant apply select or input
        */
        if (fieldSelected.includes('.keyword')) {
          return this.operatorFieldSelectable();
        } else {
          // if type of current filter is not keyword return result of validation if current operator is selectable or not
          return (this.formFilter.get('operator').value === ElasticOperatorsEnum.IS_ONE_OF ||
            this.formFilter.get('operator').value === ElasticOperatorsEnum.IS_NOT_ONE_OF);
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
      this.operator === this.operatorEnum.IS_ONE_OF || this.operator === this.operatorEnum.IS_NOT_ONE_OF;
  }

  /**
   * Return current operator
   */
  get operator(): ElasticOperatorsEnum {
    return this.formFilter.get('operator').value;
  }

  applyInputFilter(): boolean {
    const field = this.formFilter.get('field').value;
    const index = this.fields.findIndex(value => value.name === field);
    if (index !== -1) {
      return (this.fields[index].type === ElasticDataTypesEnum.TEXT || this.fields[index].type === ElasticDataTypesEnum.STRING);
    } else {
      return false;
    }
  }

  /**
   * Return index of field selected
   */
  getIndexField(): number {
    return this.fields.findIndex(value => value.name === this.formFilter.get('field').value);
  }

  /**
   * return current field selected
   */
  get field(): ElasticSearchFieldInfoType {
    const field = this.formFilter.get('field').value;
    const index = this.fields.findIndex(value => value.name === field);
    if (index !== -1) {
      return this.fields[index];
    }
  }

  /**
   * Return operators based on field type
   */
  getOperators() {
    const index = this.getIndexField();
    if (index !== -1) {
      const fieldType = this.fields[index].type;
      if (fieldType === ElasticDataTypesEnum.TEXT || fieldType === ElasticDataTypesEnum.STRING) {
        if (!this.field.name.includes('.keyword')) {
          this.operators = FILTER_OPERATORS.filter(value =>
            value.operator !== ElasticOperatorsEnum.IS_BETWEEN &&
            value.operator !== ElasticOperatorsEnum.IS_NOT_BETWEEN);
        } else {
          this.operators = FILTER_OPERATORS.filter(value =>
            value.operator !== ElasticOperatorsEnum.IS_BETWEEN &&
            value.operator !== ElasticOperatorsEnum.CONTAIN &&
            value.operator !== ElasticOperatorsEnum.DOES_NOT_CONTAIN &&
            value.operator !== ElasticOperatorsEnum.IS_NOT_BETWEEN);
        }

      } else if (fieldType === ElasticDataTypesEnum.LONG ||
        fieldType === ElasticDataTypesEnum.NUMBER || fieldType === ElasticDataTypesEnum.DATE) {
        this.operators = FILTER_OPERATORS.filter(value =>
          value.operator !== ElasticOperatorsEnum.CONTAIN &&
          value.operator !== ElasticOperatorsEnum.DOES_NOT_CONTAIN);
      } else {
        this.operators = FILTER_OPERATORS;
      }
    }
  }

  /**
   * Determine if a field select value cant choose multiple values or not
   */
  isMultipleSelectValue() {
    this.multiple = this.formFilter.get('operator').value === ElasticOperatorsEnum.IS_ONE_OF ||
      this.formFilter.get('operator').value === ElasticOperatorsEnum.IS_NOT_ONE_OF;
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
