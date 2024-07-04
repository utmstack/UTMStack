import {Component, EventEmitter, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {FilterType} from '../../models/filter.type';
import {FilterService} from '../../services/filter.service';
import {Observable, Subject} from 'rxjs';
import {AssetFieldFilterEnum} from "../../../assets-discover/shared/enums/asset-field-filter.enum";
import {CollectorFieldFilterEnum} from "../../../assets-discover/shared/enums/collector-field-filter.enum";
import {AssetFilterType} from "../../../assets-discover/shared/types/asset-filter.type";
import {STATICS_FILTERS} from "../../../assets-discover/shared/const/filter-const";
import {AssetMapFilterFieldEnum} from "../../../assets-discover/shared/enums/asset-map-filter-field.enum";
import {distinctUntilChanged, filter, takeUntil, tap} from "rxjs/operators";


@Component({
    selector: 'app-generic-filter',
    templateUrl: './generic-filter.component.html',
    styleUrls: ['./generic-filter.component.css']
})
export class GenericFilterComponent<T extends FilterType, V> implements OnInit, OnDestroy {
    @Input() fieldFilter: T;
    @Input() url: string;
    @Input() forGroups = false;
    @Output() filterGenericChange = new EventEmitter<{ prop: string, values: string[]}>();

    fieldValues$: Observable<Array<[string, number]>>;
    loading = false;
    selected = [];
    loadingMore = false;
    searching = false;
    requestParams: any;
    destroy$: Subject<void> = new Subject();

    constructor(private filterService: FilterService) {
    }

    ngOnInit(): void {
      this.requestParams = {page: 0, prop: this.fieldFilter.field, size: 6, forGroups: this.forGroups};
      this.getFieldFilterValues();
      this.filterService.resetFilter
          .pipe(
              takeUntil(this.destroy$),
              filter(status => !!status)
          )
          .subscribe( status => {
              this.selected = [];
          });
    }

    getFieldFilterValues() {
        this.loading = true;
        this.fieldValues$ = this.filterService.getFieldValues(this.url, this.requestParams)
            .pipe(tap(() => this.loading = !this.loading));
    }

    onScroll() {
        this.requestParams.page += 1;
        this.loadingMore = true;
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

        this.filterGenericChange.emit({prop: this.fieldFilter.field, values: this.selected});

    }

    searchInValues($event: string) {
        this.requestParams.value = $event;
        this.requestParams.page = 0;
        this.searching = true;
        this.getFieldFilterValues();
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

    ngOnDestroy(): void {
        this.destroy$.next();
        this.destroy$.complete();
    }

}
