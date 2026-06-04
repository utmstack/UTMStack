import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {ALERT_STATUS_FIELD, ALERT_TIMESTAMP_FIELD} from '../../../../../../shared/constants/alert/alert-field.constant';
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
        this.filters.push(this.fb.group({
          field: field.field,
          value: extractValueFromObjectByPath(this.alert, field),
          operator: ElasticOperatorsEnum.IS,
          status: 'ACTIVE',
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
        if (typeof filter.value === 'string' && filter.value !== '-') {
          filter.value = [filter.value];
        }
        filter.value = filter.value === '-' ? null : filter.value;
      }
    }
    this.addRowToFilter.emit(filterEmit.filter(value => value.status === 'ACTIVE' && value.value !== null));
    this.activeModal.close();
  }

  getFieldToRender(field: string) {
    return this.fields[this.fields.findIndex(value => value.field === field)];
  }
}
