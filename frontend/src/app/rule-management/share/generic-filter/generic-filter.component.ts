import {AfterViewInit, Component, Input, OnDestroy, OnInit} from '@angular/core';
import {Observable, Subject} from 'rxjs';
import {filter, map, switchMap, takeUntil, tap} from 'rxjs/operators';
import {UtmFieldType} from '../../../shared/types/table/utm-field.type';
import {FilterService} from '../../services/filter.service';


@Component({
  selector: 'app-generic-filter',
  templateUrl: './generic-filter.component.html',
  styleUrls: ['./generic-filter.component.css']
})
export class GenericFilterComponent<T extends UtmFieldType> implements OnInit, AfterViewInit, OnDestroy {
  @Input() fieldFilter: T;
  @Input() url: string;
  @Input() forGroups = false;

  fieldValues$: Observable<Array<[string, number]>>;
  loading = true;
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
    this.requestParams = {page: 0, prop: this.fieldFilter.filterField, size: 6, forGroups: this.forGroups};
    this.fieldValues$ = this.filterService.onRefresh$
      .pipe(
        filter(refresh => !!refresh &&
          (refresh === 'ALL' || refresh === this.fieldFilter.filterField)),
        switchMap(() => this.getFieldFilterValues()),
        tap(() => this.loading = false),
      );
    this.filterService.resetFilter
      .pipe(
        takeUntil(this.destroy$),
        filter(status => !!status)
      )
      .subscribe(status => {
        this.selected = [];
        this.filterService.onFilterChange({});
      });
  }

  ngAfterViewInit(): void {
    this.filterService.notifyRefresh(this.fieldFilter.filterField);
  }

  getFieldFilterValues() {
    return this.filterService.fetchData({
      url: this.url,
      params: this.requestParams
    })
      .pipe(
        tap((response) => {
          this.searching = false;
        }));
  }

  onScroll() {
    this.requestParams.size += 6;
    console.log(this.fieldFilter.filterField);
    this.filterService.notifyRefresh(this.fieldFilter.filterField);
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
    this.filterService.notifyRefresh(this.fieldFilter.filterField);
  }

  ngOnDestroy(): void {
    this.filterService.resetFieldValues();
    this.destroy$.next();
    this.destroy$.complete();
  }

}
