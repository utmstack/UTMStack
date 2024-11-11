import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {ALERT_TAGS_FIELD} from '../../../../../../shared/constants/alert/alert-field.constant';
import {ALERT_INDEX_PATTERN} from '../../../../../../shared/constants/main-index-pattern.constant';
import {ElasticDataTypesEnum} from '../../../../../../shared/enums/elastic-data-types.enum';
import {ElasticOperatorsEnum} from '../../../../../../shared/enums/elastic-operators.enum';
import {ElasticSearchIndexService} from '../../../../../../shared/services/elasticsearch/elasticsearch-index.service';
import {ElasticFilterType} from '../../../../../../shared/types/filter/elastic-filter.type';
import {UtmFieldType} from '../../../../../../shared/types/table/utm-field.type';
import {AlertFiltersBehavior} from '../../../behavior/alert-filters.behavior';
import {AlertUpdateTagBehavior} from '../../../behavior/alert-update-tag.behavior';
import {Subject} from "rxjs";
import {takeUntil} from "rxjs/operators";

@Component({
  selector: 'app-alert-generic-filter',
  templateUrl: './alert-generic-filter.component.html',
  styleUrls: ['./alert-generic-filter.component.scss']
})
export class AlertGenericFilterComponent implements OnInit, OnDestroy {
  @Output() filterGenericChange = new EventEmitter<ElasticFilterType>();
  @Input() fieldFilter: UtmFieldType;
  activeFilters: ElasticFilterType[] = [];
  fieldValues: object;
  loading = true;
  selected: any[] = [];
  search: string;
  searching = false;
  loadingMore = false;
  top = 6;
  filter: ElasticFilterType;
  sort: { orderByCount: boolean, sortAsc: boolean } = {orderByCount: true, sortAsc: false};
  destroy$: Subject<void> = new Subject<void>();

  constructor(private elasticSearchIndexService: ElasticSearchIndexService,
              private alertFiltersBehavior: AlertFiltersBehavior,
              private alertUpdateTagBehavior: AlertUpdateTagBehavior) {
  }

  ngOnInit() {
    // this.getFieldValues();
    /**
     * If filter is tags subscribe to changes to reload data on add new tag on alert
     */
    if (this.fieldFilter.field === ALERT_TAGS_FIELD) {
      this.alertUpdateTagBehavior.$tagRefresh
        .pipe(takeUntil(this.destroy$))
        .subscribe(tagUpdate => {
        if (tagUpdate) {
          this.getFieldValues();
        }
      });
    }
    /**
     * Reset all values of selected filter
     */
    this.alertFiltersBehavior.$resetFilter
      .pipe(takeUntil(this.destroy$))
      .subscribe(reset => {
      if (reset) {
        this.selected = [];
      }
    });
    this.alertFiltersBehavior.$deleteFilterValue
      .pipe(takeUntil(this.destroy$))
      .subscribe(deleteFilter => {
      if (deleteFilter) {
        const deleteField = deleteFilter.field.replace('.keyword', '');
        if (this.fieldFilter.field === deleteField) {
          const selectedIndex = this.selected.findIndex(value => value === deleteFilter.value);
          if (selectedIndex !== -1) {
            this.selected.splice(selectedIndex, 1);
          }
        }
      }
    });
    this.alertFiltersBehavior.$filters
      .pipe(takeUntil(this.destroy$))
      .subscribe((filters: ElasticFilterType[]) => {
      if (filters) {
        this.activeFilters = filters;
        this.getFieldValues();
        const index = filters.findIndex(value => value.field
            .replace('.keyword', '') === this.fieldFilter.field
          && value.operator !== ElasticOperatorsEnum.EXIST && value.operator !== ElasticOperatorsEnum.DOES_NOT_EXIST
          && value.operator !== ElasticOperatorsEnum.IS_NOT);
        if (index !== -1) {
          let values: any[] = filters[index].value;
          if (typeof values === 'string') {
            values = [values];
          }
          if (values && values.length > 0) {
            for (const val of values) {
              if (!this.selected.includes(val)) {
                this.selected.push(val);
              }
            }
          }
        }
      }
    });
  }

  getFieldValues() {
    console.log('load values');
    const field = this.setFieldKeyword();
    const filters = this.activeFilters
      .filter(value => !value.field.includes(field));
    if (this.search !== undefined && this.search !== '') {
      filters.push({
        field: this.fieldFilter.field,
        operator: ElasticOperatorsEnum.CONTAIN,
        value: this.search
      });
    }
    const req = {
      field,
      filters,
      index: ALERT_INDEX_PATTERN,
      orderByCount: this.sort.orderByCount,
      sortAsc: this.sort.sortAsc,
      top: this.top
    };
    this.elasticSearchIndexService.getValuesWithCount(req).subscribe(response => {
      this.fieldValues = response.body;
      this.loading = false;
      this.searching = false;
      this.loadingMore = false;
    });
  }

  setFieldKeyword(): string {
    return this.fieldFilter.type === ElasticDataTypesEnum.STRING ? this.fieldFilter.field + '.keyword' : this.fieldFilter.field;
  }

  valueHasData(): boolean {
    return Object.keys(this.fieldValues).length > 0;
  }

  searchInValues($event: string) {
    this.search = $event;
    this.searching = true;
    this.getFieldValues();
  }

  selectValue(value: any) {
    const index = this.selected.findIndex(val => val === value);
    if (index === -1) {
      this.selected.push(value);
    } else {
      this.selected.splice(index, 1);
    }
    this.emitCurrentFilter();
  }

  emitCurrentFilter() {
    const field = this.fieldFilter.type === ElasticDataTypesEnum.STRING ? this.fieldFilter.field + '.keyword' : this.fieldFilter.field;
    this.filter = {
      value: this.selected,
      operator: ElasticOperatorsEnum.IS_ONE_OF,
      field
    };
    this.filterGenericChange.emit(this.filter);
  }

  onScroll() {
    this.top += 10;
    this.loadingMore = true;
    this.getFieldValues();
  }

  onSortValuesChange($event: { orderByCount: boolean; sortAsc: boolean }) {
    this.sort = $event;
    this.getFieldValues();
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
