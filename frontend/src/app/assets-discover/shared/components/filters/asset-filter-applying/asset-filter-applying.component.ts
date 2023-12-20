import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {AssetFiltersBehavior} from '../../../behavior/asset-filters.behavior';
import {ASSETS_FIELDS_FILTERS} from '../../../const/asset-field.const';
import {STATICS_FILTERS} from '../../../const/filter-const';
import {AssetFieldFilterEnum} from '../../../enums/asset-field-filter.enum';
import {AssetMapFilterFieldEnum} from '../../../enums/asset-map-filter-field.enum';
import {AssetFilterType} from '../../../types/asset-filter.type';

@Component({
  selector: 'app-asset-filter-applying',
  templateUrl: './asset-filter-applying.component.html',
  styleUrls: ['./asset-filter-applying.component.scss']
})
export class AssetFilterApplyingComponent implements OnInit {
  @Input() assetsFilters: AssetFilterType;
  @Output() filterApplyingChange = new EventEmitter<{ prop: AssetFieldFilterEnum, values: string[] }>();
  filters: { key: string, value: any }[];
  STATIC_FILTERS = STATICS_FILTERS;

  constructor(private assetFiltersBehavior: AssetFiltersBehavior) {
  }

  ngOnInit() {
    this.assetFiltersBehavior.$assetAppliedFilter.subscribe(filters => {
      if (filters) {
        this.filters = this.setFilters();
      }
    });
  }

  anyFilterApplied() {
    return this.assetsFilters.os
      || this.assetsFilters.openPorts
      || this.assetsFilters.type
      || this.assetsFilters.status
      || this.assetsFilters.severity
      || this.assetsFilters.probe
      || this.assetsFilters.groups;
  }

  resolveFilterName(key: string) {
    switch (key) {
      case AssetMapFilterFieldEnum.STATUS:
        return ASSETS_FIELDS_FILTERS[ASSETS_FIELDS_FILTERS
          .findIndex(value => value.field === AssetFieldFilterEnum.STATUS)].label;
      case AssetMapFilterFieldEnum.OS:
        return ASSETS_FIELDS_FILTERS[ASSETS_FIELDS_FILTERS
          .findIndex(value => value.field === AssetFieldFilterEnum.OS)].label;
      case AssetMapFilterFieldEnum.PORTS:
        return ASSETS_FIELDS_FILTERS[ASSETS_FIELDS_FILTERS
          .findIndex(value => value.field === AssetFieldFilterEnum.PORTS)].label;
      case AssetMapFilterFieldEnum.SEVERITY:
        return ASSETS_FIELDS_FILTERS[ASSETS_FIELDS_FILTERS
          .findIndex(value => value.field === AssetFieldFilterEnum.SEVERITY)].label;
      case AssetMapFilterFieldEnum.TYPE:
        return ASSETS_FIELDS_FILTERS[ASSETS_FIELDS_FILTERS
          .findIndex(value => value.field === AssetFieldFilterEnum.TYPE)].label;
      case AssetMapFilterFieldEnum.PROBE:
        return ASSETS_FIELDS_FILTERS[ASSETS_FIELDS_FILTERS
          .findIndex(value => value.field === AssetFieldFilterEnum.PROBE)].label;
      case AssetMapFilterFieldEnum.GROUPS:
        return ASSETS_FIELDS_FILTERS[ASSETS_FIELDS_FILTERS
          .findIndex(value => value.field === AssetFieldFilterEnum.GROUP)].label;
    }
  }

  deleteFilter(key: string, value: string) {
    this.onDelete(key, value).then(filter => {
      this.filterApplyingChange.emit(filter);
    });
  }

  onDelete(key: string, value?: string): Promise<{ prop: AssetFieldFilterEnum, values: string[] }> {
    return new Promise<{ prop: AssetFieldFilterEnum, values: string[] }>(resolve => {
      if (value) {
        const indexVal = this.assetsFilters[key].findIndex(val => val === value);
        this.assetsFilters[key].splice(indexVal, 1);
      }
      resolve({prop: this.resolvePropFilterByKey(key), values: this.assetsFilters[key]});
    });
  }

  resolvePropFilterByKey(key: string) {
    switch (key) {
      case AssetMapFilterFieldEnum.STATUS:
        return AssetFieldFilterEnum.STATUS;
      case AssetMapFilterFieldEnum.OS:
        return AssetFieldFilterEnum.OS;
      case AssetMapFilterFieldEnum.PORTS:
        return AssetFieldFilterEnum.PORTS;
      case AssetMapFilterFieldEnum.SEVERITY:
        return AssetFieldFilterEnum.SEVERITY;
      case AssetMapFilterFieldEnum.TYPE:
        return AssetFieldFilterEnum.TYPE;
      case AssetMapFilterFieldEnum.PROBE:
        return AssetFieldFilterEnum.PROBE;
    }
  }

  setFilters() {
    return Object.keys(this.assetsFilters).map((key) => {
      return {key, value: this.assetsFilters[key]};
    });
  }
}
