import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {Observable, Subject} from 'rxjs';
import {distinctUntilChanged, filter, map, takeUntil, tap} from 'rxjs/operators';
import {STATICS_FILTERS} from '../../../assets-discover/shared/const/filter-const';
import {AssetFieldFilterEnum} from '../../../assets-discover/shared/enums/asset-field-filter.enum';
import {AssetMapFilterFieldEnum} from '../../../assets-discover/shared/enums/asset-map-filter-field.enum';
import {CollectorFieldFilterEnum} from '../../../assets-discover/shared/enums/collector-field-filter.enum';
import {AssetFilterType} from '../../../assets-discover/shared/types/asset-filter.type';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {FilterType} from '../../models/filter.type';
import {FilterService} from '../../services/filter.service';


@Component({
    selector: 'app-generic-filter',
    templateUrl: './generic-filter.component.html',
    styleUrls: ['./generic-filter.component.css']
})
export class GenericFilterComponent<T extends UtmFieldType> implements OnInit, OnDestroy {
    @Input() fieldFilter: T;
    @Input() url: string;
    @Input() forGroups = false;

    fieldValues$: Observable<Array<[string, number]>>;
    loading = false;
    loadMore = false;
    selected = [];
    searching = false;
    requestParams: any;
    destroy$: Subject<void> = new Subject();
    filters: Array<[string, number]> = [];
    totalItems = 0;

    constructor(private filterService: FilterService) {
    }

    ngOnInit(): void {

      this.fieldValues$ = this.filterService.fieldsValues$
                            .pipe(
                                map(values => {
                                    const key = this.fieldFilter.filterField;
                                    return values.get(key);
                                })
                            );

      this.requestParams = {page: 0, prop: this.fieldFilter.filterField, size: 6, forGroups: this.forGroups};
      this.getFieldFilterValues();
      this.filterService.resetFilter
          .pipe(
              takeUntil(this.destroy$),
              filter(status => !!status)
          )
          .subscribe( status => {
              this.selected = [];
              this.filterService.onFilterChange({});
          });
    }

    getFieldFilterValues() {
        if (this.requestParams.prop) {
            this.filterService.getFieldValues(this.url, this.requestParams, this.loadMore)
                .pipe(
                    map((response) => this.loadMore ? this.filters.concat(response) : this.filters = response),
                    tap((response) => {
                        this.filters = response;
                        this.loadMore = false;
                        this.searching = false;
                    })
                ).subscribe();
        }
    }

    onScroll() {
        this.requestParams.page += 1;
        this.loadMore = true;
        this.getFieldFilterValues();
    }

    onSortValuesChange($event: { orderByCount: boolean; sortAsc: boolean }) {
    }

    selectValue(value: string) {
        const index = this.selected.findIndex(val => val === value);
        if (index === -1) {
            this.selected.push(value);
        } else {
            this.selected.splice(index, 1);
        }

        this.filterService.onFilterChange({prop: this.fieldFilter.field, values: this.selected});

    }

    searchInValues($event: string) {
        this.requestParams.value = $event;
        this.requestParams.page = 0;
        this.searching = true;
        this.getFieldFilterValues();
    }

    trackByFn(index: number , value: [string, number]) {
        return value[0];
    }

    ngOnDestroy(): void {
        this.filterService.resetFieldValues();
        this.destroy$.next();
        this.destroy$.complete();
    }

}
