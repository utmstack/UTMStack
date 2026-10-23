import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {ALERT_STATUS_FIELD, ALERT_TIMESTAMP_FIELD} from '../../../../../../shared/constants/alert/alert-field.constant';
import {ElasticDataTypesEnum} from '../../../../../../shared/enums/elastic-data-types.enum';
import {ElasticOperatorsEnum} from '../../../../../../shared/enums/elastic-operators.enum';
import {ElasticFilterType} from '../../../../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../../../../shared/types/table/utm-field.type';
import {extractValueFromObjectByPath} from '../../../../../../shared/util/get-value-object-from-property-path.util';
import {AlertFiltersBehavior} from '../../../behavior/alert-filters.behavior';

@Component({
  selector: 'app-row-to-filters',
  templateUrl: './row-to-filters.component.html',
  styleUrls: ['./row-to-filters.component.scss']
})
export class RowToFiltersComponent implements OnInit {
  @Input() alert: any;
  @Input() fields: UtmFieldType[];
  @Input() icon: string;
  operatorEnum = ElasticOperatorsEnum;
  /**
   * Return ElasticFilterType[] from row
   */
  @Output() addRowToFilter = new EventEmitter<ElasticFilterType[]>();
  formRowFilters: FormGroup;

  constructor(private activeModal: NgbActiveModal,
              private alertFiltersBehavior: AlertFiltersBehavior,
              private fb: FormBuilder) {
  }

  get filters() {
    return this.formRowFilters.controls.filters as FormArray;
  }

  ngOnInit() {
    this.initFormRow();
    this.addValuesToFilter();
  }

  addValuesToFilter() {
    for (const field of this.fields) {
      if (field.visible && field.field !== ALERT_TIMESTAMP_FIELD) {
        const value = extractValueFromObjectByPath(this.alert, field);
        // Composite columns (type OBJECT) hold a nested object that cannot be
        // expressed as a single filter value. They are offered UNCHECKED so
        // the default "Apply filters" selection never includes them (#2968).
        const isComposite = field.type === ElasticDataTypesEnum.OBJECT;
        this.filters.push(this.fb.group({
          field: field.field,
          value,
          operator: ElasticOperatorsEnum.IS,
          status: isComposite ? 'INACTIVE' : 'ACTIVE',
          label: field.label
        }));
      }
    }
  }

  initFormRow() {
    this.formRowFilters = this.fb.group({
      filters: this.fb.array([], Validators.required),
    });
  }

  cancelDelete() {
    this.activeModal.dismiss('cancel');
  }

  addToSelected(field: UtmFieldType) {
    field.visible = !field.visible;
  }

  applyFilter() {
    const filterEmit: ElasticFilterType[] = this.filters.value;
    for (const filter of filterEmit) {
      if (!(filter.field.includes(ALERT_STATUS_FIELD) ||
        filter.field.includes(ALERT_TIMESTAMP_FIELD))) {
        filter.operator = filter.operator === ElasticOperatorsEnum.IS ? ElasticOperatorsEnum.IS_ONE_OF : ElasticOperatorsEnum.IS_NOT_ONE_OF;
        filter.value = this.sanitizeFilterValue(filter.value);
      }
    }
    this.addRowToFilter.emit(filterEmit.filter(value => value.status === 'ACTIVE' && value.value !== null));
    this.activeModal.close();
  }

  /**
   * Normalise a row value into a shape the backend accepts for IS_ONE_OF /
   * IS_NOT_ONE_OF (a list, never a bare object/null/undefined - see
   * FilterType.validate). Returns null to drop filters that have no usable
   * value: placeholder '-', booleans, composite (object) columns and empty
   * collections.
   */
  private sanitizeFilterValue(value: any): any {
    if (value === null || value === undefined || value === '-') {
      return null;
    }
    if (typeof value === 'object') {
      return Array.isArray(value) && value.length > 0 ? value.map(v => String(v)) : null;
    }
    if (typeof value === 'boolean') {
      return null;
    }
    return [String(value)];
  }

  getFieldToRender(field: string) {
    return this.fields[this.fields.findIndex(value => value.field === field)];
  }
}
