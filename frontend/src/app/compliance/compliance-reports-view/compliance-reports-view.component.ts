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
import {catchError, filter, map, switchMap, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {SortEvent} from '../../shared/directives/sortable/type/sort-event';
import {SortByType} from '../../shared/types/sort-by.type';
import {CpReportsService} from '../shared/services/cp-reports.service';
import {ComplianceReportType} from '../shared/type/compliance-report.type';
import {ComplianceStandardSectionType} from '../shared/type/compliance-standard-section.type';
import {ComplianceStatusEnum} from '../shared/enums/compliance-status.enum';

@Component({
  selector: 'app-compliance-reports-view',
  templateUrl: './compliance-reports-view.component.html',
  styleUrls: ['./compliance-reports-view.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ComplianceReportsViewComponent implements OnInit, OnChanges, OnDestroy {
  @Input() section: ComplianceStandardSectionType;
  @Output() pageChange = new EventEmitter<{}>();

  reports$: Observable<ComplianceReportType[]>;
  selected: number;
  fields: SortByType[];
  reportDetail: ComplianceReportType;
  loading = true;
  noData = false;
  itemsPerPage = 15;
  page = 0;
  totalItems = 0;
  sortEvent: SortEvent = {
    column: 'configReportName',
    direction: 'desc'
  };
  destroy$: Subject<void> = new Subject();
  sort = 'configReportName,desc';
  search: string;
  viewportHeight: number;
  ComplianceStatusEnum = ComplianceStatusEnum;

  constructor(private reportsService: CpReportsService,
              private toastService: UtmToastService) {
  }

  @HostListener('window:resize', ['$event'])
  onResize(event: Event) {
    this.viewportHeight = window.innerHeight;
  }

  ngOnInit() {
    this.viewportHeight = window.innerHeight;
    this.reports$ = this.reportsService.onRefresh$
      .pipe(takeUntil(this.destroy$),
        filter(reportRefresh =>
          !!reportRefresh && reportRefresh.loading),
        tap((reportRefresh) => {
          this.loading = true;
          this.selected = reportRefresh.reportSelected;
        }),
        switchMap((reportRefresh) => this.reportsService.fetchData({
          page: reportRefresh.page,
          size: this.itemsPerPage,
          standardId: this.section.standardId,
          sectionId: this.section.id,
          expandDashboard: true,
          setStatus: true,
          sort: this.sort,
          search: this.search ? this.search : null,
        })),
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
    this.reportsService.notifyRefresh({
      loading: true,
      sectionId: this.section.id,
      reportSelected: 0
    });
  }

  loadPage(pageEvent: number) {
    const page = this.page !== 0 ? this.page - 1 : this.page;
    this.reportsService.notifyRefresh({
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
    this.reportsService.notifyRefresh({
      loading: true,
      sectionId: this.section.id,
      reportSelected: 0
    });
  }

  onSearch($event: string) {
    this.search = $event;
    this.reportsService.notifyRefresh({
      loading: true,
      sectionId: this.section.id,
      reportSelected: 0
    });
  }

  getTableHeight() {
    console.log('getTableHeight:', 100 - ((350 / this.viewportHeight) * 100) + 'vh');
    return 100 - ((350 / this.viewportHeight) * 100) + 'vh';
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  protected readonly ComplianceStatus = ComplianceStatusEnum;
}
