import {HttpResponse} from '@angular/common/http';
import {AfterViewInit, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {EMPTY, Observable, Subject} from 'rxjs';
import {catchError, filter, map, switchMap, takeUntil, tap} from 'rxjs/operators';
import {ElasticFilterType} from '../../../../../shared/types/filter/elastic-filter.type';
import {AssetFiltersBehavior} from '../../../behavior/asset-filters.behavior';
import {AssetReloadFilterBehavior} from '../../../behavior/asset-reload-filter-behavior.service';
import {STATICS_FILTERS} from '../../../const/filter-const';
import {AssetFieldFilterEnum} from '../../../enums/asset-field-filter.enum';
import {AssetMapFilterFieldEnum} from '../../../enums/asset-map-filter-field.enum';
import {AssetsStatusEnum} from '../../../enums/assets-status.enum';
import {CollectorFieldFilterEnum} from '../../../enums/collector-field-filter.enum';
import {AssetGenericFilterService} from '../../../services/asset-generic-filter.service';
import {AssetFilterType} from '../../../types/asset-filter.type';

@Component({
  selector: 'app-asset-generic-filter',
  templateUrl: './asset-generic-filter.component.html',
  styleUrls: ['./asset-generic-filter.component.scss']
})
export class AssetGenericFilterComponent implements OnInit, AfterViewInit {
  @Input() fieldFilter: ElasticFilterType;
  @Output() filterGenericChange = new EventEmitter<{ prop: AssetFieldFilterEnum | CollectorFieldFilterEnum, values: string[] }>();
  @Input() forGroups = false;
  // fieldValues: Array<[string, number]> = [];
  fieldValues$: Observable<Array<[string, number]>>;
  loading = true;
  selected = [];
  loadingMore = false;
  searching = false;
  requestParams = {
    page: 0,
    prop: null,
    size: 6,
    forGroups: this.forGroups,
    value: null
  };
  destroy$: Subject<void> = new Subject<void>();

  constructor(private assetGenericFilterService: AssetGenericFilterService,
              private assetTypeChangeBehavior: AssetReloadFilterBehavior,
              private assetFiltersBehavior: AssetFiltersBehavior) {
  }

  ngOnInit() {
    this.requestParams = {
      ...this.requestParams,
      prop: this.fieldFilter.field
    };
    this.fieldValues$ = this.assetGenericFilterService.onRefresh$
      .pipe(
        filter(filterData => !!filterData && filterData.refresh && filterData.fieldFilter === this.fieldFilter.field),
        switchMap(() => this.assetGenericFilterService.fetchData(this.requestParams)),
        tap((response: HttpResponse<Array<[string, number]>>) => {
          this.loading = false;
          this.searching = false;
          this.loadingMore = false;
        }),
        map((response) => response.body),
        catchError (err => {
          this.loading = false;
          return EMPTY;
        }));

    this.assetFiltersBehavior.$assetFilter
      .pipe(
        takeUntil(this.destroy$),
        filter(filters => !!filters))
      .subscribe(filters => this.setValueOfFilter(filters));
    /**
     * Update type values filter on type is applied to asset
     */
    this.assetTypeChangeBehavior.$assetReloadFilter
      .pipe(takeUntil(this.destroy$))
      .subscribe(change => {
        if (change && this.fieldFilter.field === change) {
          this.requestParams.page = 0;
          this.loading = true;
          this.selected = [];
          this.getPropertyValues();
      }
    });
  }

  ngAfterViewInit(): void {
    this.getPropertyValues();
  }

  getPropertyValues() {
    this.loading = true;
    this.assetGenericFilterService.notifyRefresh({
      fieldFilter: this.fieldFilter.field,
      refresh: true
    });
  }

  onSortValuesChange($event: { orderByCount: boolean; sortAsc: boolean }) {
  }

  onScroll() {
    this.requestParams = {
      ...this.requestParams,
      size: this.requestParams.size + 6
    };
    this.loadingMore = true;
    this.getPropertyValues();
  }

  selectValue(value: string ) {
    const index = this.selected.findIndex(val => val === value);
    if (index === -1) {
      this.selected.push(value);
    } else {
      this.selected.splice(index, 1);
    }

    this.filterGenericChange.emit({prop: AssetFieldFilterEnum[this.fieldFilter.field] ?
        AssetFieldFilterEnum[this.fieldFilter.field] : CollectorFieldFilterEnum[this.fieldFilter.field], values: this.selected});

  }

  searchInValues($event: string) {
    let value = $event;
    if (this.fieldFilter.field === AssetFieldFilterEnum.ALIVE) {
      value = this.getAliveFieldValue($event);
    }
    this.requestParams.value = value;
    this.requestParams.page = 0;
    this.searching = true;
    this.getPropertyValues();
  }

  setValueOfFilter(filters: AssetFilterType) {
    const filterKey = AssetFieldFilterEnum[this.fieldFilter.field] ?
      AssetFieldFilterEnum[this.fieldFilter.field] : CollectorFieldFilterEnum[this.fieldFilter.field];

    const matchingKey = Object.keys(filters).find(key =>
      !STATICS_FILTERS.includes(key) && key === AssetMapFilterFieldEnum[filterKey]
    );

    if (matchingKey) {
      this.selected = filters[matchingKey] === null ? [] : filters[matchingKey];
    }
  }

  getValue(value: string) {
    if (this.fieldFilter.field === AssetFieldFilterEnum.ALIVE) {
      return value ? AssetsStatusEnum.CONNECTED : AssetsStatusEnum.DISCONNECTED;
    }

    return value;
  }
  getAliveFieldValue(value: string): string {
    if (value !== '' && AssetsStatusEnum.CONNECTED.toLowerCase().indexOf(value.toLowerCase()) > -1) {
      return 'true';
    } else if (value !== '' && AssetsStatusEnum.DISCONNECTED.toLowerCase().indexOf(value.toLowerCase()) > -1) {
      return 'false';
    } else {
      return '';
    }
  }

}
