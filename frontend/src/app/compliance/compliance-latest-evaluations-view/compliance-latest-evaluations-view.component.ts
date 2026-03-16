import {HttpErrorResponse} from '@angular/common/http';
import {
  ChangeDetectionStrategy,
  Component,
  EventEmitter, HostListener,
  Input,
  OnChanges,
  OnDestroy,
  OnInit,
  Output
} from '@angular/core';
import {EMPTY, Observable, Subject} from 'rxjs';
import {catchError, distinctUntilChanged, filter, map, switchMap, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {TimezoneFormatService} from '../../shared/services/utm-timezone.service';
import {DatePipeDefaultOptions} from '../../shared/types/date-pipe-default-options';
import {ComplianceStatusExtendedEnum} from '../shared/enums/compliance-status.enum';
import {CpControlConfigService} from '../shared/services/cp-control-config.service';
import {ComplianceControlLatestEvaluationType} from '../shared/type/compliance-control-latest-evaluation.type';
import {ComplianceStandardSectionType} from '../shared/type/compliance-standard-section.type';

@Component({
  selector: 'app-compliance-latest-evaluations-view',
  templateUrl: './compliance-latest-evaluations-view.component.html',
  styleUrls: ['./compliance-latest-evaluations-view.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ComplianceLatestEvaluationsViewComponent implements OnInit, OnChanges, OnDestroy {
  @Input() section: ComplianceStandardSectionType;
  @Output() pageChange = new EventEmitter<{}>();

  controls$: Observable<ComplianceControlLatestEvaluationType[]>;
  selected: number;
  controlDetail: ComplianceControlLatestEvaluationType;
  loading = true;
  noData = false;
  itemsPerPage = 15;
  page = 0;
  totalItems = 0;
  sortEvent: SortEvent = {
    column: 'controlName',
    direction: 'desc'
  };
  destroy$: Subject<void> = new Subject();
  sort = 'controlName,desc';
  search: string;
  viewportHeight: number;
  dateFormat$: Observable<DatePipeDefaultOptions>;

  ComplianceStatusExtendedEnum = ComplianceStatusExtendedEnum;

  constructor(private controlsService: CpControlConfigService,
              private toastService: UtmToastService,
              private timezoneFormatService: TimezoneFormatService) {
  }

  @HostListener('window:resize', ['$event'])
  onResize(event: Event) {
    this.viewportHeight = window.innerHeight;
  }

  ngOnInit() {
    this.viewportHeight = window.innerHeight;
    this.dateFormat$ = this.timezoneFormatService.getDateFormatSubject();
    this.controls$ = this.controlsService.onRefresh$
      .pipe(
        takeUntil(this.destroy$),
        distinctUntilChanged((prev, curr) =>
          prev &&
          curr &&
          prev.loading === curr.loading &&
          prev.page === curr.page &&
          prev.reportSelected === curr.reportSelected &&
          prev.sectionId === curr.sectionId
        ),
        filter(reportRefresh =>
          !!reportRefresh && reportRefresh.loading
        ),
        tap((reportRefresh) => {
          this.loading = true;
          this.selected = reportRefresh.reportSelected;
        }),
        switchMap((reportRefresh) => {
          return this.controlsService.fetchData({
            page: reportRefresh.page,
            size: this.itemsPerPage,
            sectionId: this.section.id,
            sort: this.sort,
            search: this.search ? this.search : null,
          });
        }),
        tap(res => this.totalItems = Number(res.headers.get('X-Total-Count'))),
        map((res) => {
          return res.body.map((r, index) => {
            return {
              ...r,
              selected: index === this.selected
            };
          });
        }),
        catchError((err: HttpErrorResponse) => {
          this.toastService.showError('Error',
            'Unable to retrieve the list of reports. Please try again or contact support.');
          this.loading = false;
          return EMPTY;
        }),
        tap((data) => {
          this.loading = false;
          this.noData = data.length === 0;
        }));
  }

  ngOnChanges(): void {
    this.page = 0;
    this.controlsService.notifyRefresh({
      loading: true,
      sectionId: this.section.id,
      reportSelected: 0
    });
  }

  loadPage(pageEvent: number) {
    const page = this.page !== 0 ? this.page - 1 : this.page;
    this.controlsService.notifyRefresh({
      loading: true,
      sectionId: this.section.id,
      reportSelected: 0,
      page
    });
    this.pageChange.emit({
      page,
      size: this.itemsPerPage,
      sort: this.sort
    });
  }

  onSortBy(sort: SortEvent) {
    this.sort = `${sort.column},${sort.direction}`;
    this.controlsService.notifyRefresh({
      loading: true,
      sectionId: this.section.id,
      reportSelected: 0
    });
  }

  onSearch($event: string) {
    this.search = $event;
    this.controlsService.notifyRefresh({
      loading: true,
      sectionId: this.section.id,
      reportSelected: 0
    });
  }

  getTableHeight() {
    return 100 - ((350 / this.viewportHeight) * 100) + 'vh';
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
