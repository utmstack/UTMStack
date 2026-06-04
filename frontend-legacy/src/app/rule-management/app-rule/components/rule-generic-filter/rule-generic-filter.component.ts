import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {errorHandler} from '@angular/platform-browser/src/browser';
import { of } from 'rxjs';
import {catchError} from 'rxjs/operators';
import {AssetFiltersBehavior} from '../../../../assets-discover/shared/behavior/asset-filters.behavior';
import {
  AssetReloadFilterBehavior
} from '../../../../assets-discover/shared/behavior/asset-reload-filter-behavior.service';
import {STATICS_FILTERS} from '../../../../assets-discover/shared/const/filter-const';
import {AssetFieldFilterEnum} from '../../../../assets-discover/shared/enums/asset-field-filter.enum';
import {AssetMapFilterFieldEnum} from '../../../../assets-discover/shared/enums/asset-map-filter-field.enum';
import {CollectorFieldFilterEnum} from '../../../../assets-discover/shared/enums/collector-field-filter.enum';
import {UtmNetScanService} from '../../../../assets-discover/shared/services/utm-net-scan.service';
import {AssetFilterType} from '../../../../assets-discover/shared/types/asset-filter.type';
import {ElasticFilterType} from '../../../../shared/types/filter/elastic-filter.type';


@Component({
  selector: 'app-rule-generic-filter',
  templateUrl: './rule-generic-filter.component.html',
  styleUrls: ['./rule-generic-filter.component.scss']
})
export class RuleGenericFilterComponent implements OnInit {
  @Input() fieldFilter: ElasticFilterType;
  @Output() filterGenericChange = new EventEmitter<{ prop: AssetFieldFilterEnum | CollectorFieldFilterEnum, values: string[] }>();
  @Input() forGroups = false;
  fieldValues: Array<[string, number]> = [];
  loading = true;
  selected = [];
  loadingMore = false;
  searching = false;
  requestParams: any;

  constructor(private utmNetScanService: UtmNetScanService,
              private assetTypeChangeBehavior: AssetReloadFilterBehavior,
              private assetFiltersBehavior: AssetFiltersBehavior) {
  }

  ngOnInit() {
    this.requestParams = {page: 0, prop: this.fieldFilter.field, size: 6, forGroups: this.forGroups};
    this.getPropertyValues();
    this.assetFiltersBehavior.$assetFilter.subscribe(filters => {
      if (filters) {
        this.setValueOfFilter(filters);
      }
    });
    /**
     * Update type values filter on type is applied to asset
     */
    this.assetTypeChangeBehavior.$assetReloadFilter.subscribe(change => {
      if (change && this.fieldFilter.field === change) {
        this.requestParams.page = 0;
        this.fieldValues = [];
        this.loading = true;
        this.getPropertyValues();
      }
    });
  }

  getPropertyValues() {
    this.utmNetScanService.getFieldValues(this.requestParams)
        .pipe(
            catchError((err) => {
              return of(new HttpResponse({ body: [] }));
            })
        )
        .subscribe(response => {
      this.fieldValues = this.fieldValues.concat(response.body);
      this.loading = false;
      this.searching = false;
      this.loadingMore = false;
    });
  }

  onSortValuesChange($event: { orderByCount: boolean; sortAsc: boolean }) {
  }

  onScroll() {
    this.requestParams.page += 1;
    this.loadingMore = true;
    this.getPropertyValues();
  }

  selectValue(value: string) {
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
    this.requestParams.value = $event;
    this.requestParams.page = 0;
    this.searching = true;
    this.fieldValues = [];
    this.getPropertyValues();
  }

  setValueOfFilter(filters: AssetFilterType) {
    for (const key of Object.keys(filters)) {
      const filterKey = AssetFieldFilterEnum[this.fieldFilter.field] ?
        AssetFieldFilterEnum[this.fieldFilter.field] : CollectorFieldFilterEnum[this.fieldFilter.field];
      if (!STATICS_FILTERS.includes(key)
        && key === AssetMapFilterFieldEnum[filterKey]) {
        this.selected = filters[key] === null ? [] : filters[key];
      }
    }
  }

}
