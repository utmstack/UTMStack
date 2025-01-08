import {HttpErrorResponse} from '@angular/common/http';
import {ChangeDetectionStrategy, Component, Input, OnChanges, OnDestroy, OnInit, SimpleChanges} from '@angular/core';
import {NgxSpinnerService} from 'ngx-spinner';
import {EMPTY, Observable, Subject} from 'rxjs';
import {catchError, concatMap, filter, map, switchMap, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ExportPdfService} from '../../shared/services/util/export-pdf.service';
import {SortByType} from '../../shared/types/sort-by.type';
import {CpReportsService} from '../shared/services/cp-reports.service';
import {ComplianceReportType} from '../shared/type/compliance-report.type';
import {ComplianceStandardSectionType} from '../shared/type/compliance-standard-section.type';

@Component({
  selector: 'app-compliance-reports-view',
  templateUrl: './compliance-reports-view.component.html',
  styleUrls: ['./compliance-reports-view.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ComplianceReportsViewComponent implements OnInit, OnChanges, OnDestroy {
  @Input() section: ComplianceStandardSectionType;

  reports$: Observable<ComplianceReportType[]>;
  selected: number;
  fields: SortByType[];
  reportDetail: ComplianceReportType;
  loading = false;
  noData = false;
  itemsPerPage = 15;
  page = 0;
  totalItems = 0;
  destroy$: Subject<void> = new Subject();

  constructor(private reportsService: CpReportsService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    this.reports$ = this.reportsService.onRefresh$
      .pipe(takeUntil(this.destroy$),
        filter(reportRefresh =>
          !!reportRefresh && reportRefresh.loading),
        tap((reportRefresh) => {
          this.loading = true;
          this.selected = reportRefresh.reportSelected;
        }),
        switchMap(() => this.reportsService.fetchData({
          page: this.page,
          size: this.itemsPerPage,
          standardId: this.section.standardId,
          sectionId: this.section.id,
          expandDashboard: true
        })),
        tap(res => this.totalItems = Number(res.headers.get('X-Total-Count'))),
        map((res) => {
          return res.body.filter(r => r.dashboard.length > 0 && r.dashboard.every(v => v.visualization.pattern.active));
        }),
        map((res) => {
          return res.map((r, index) => {
            return {
              ...r,
              selected: index === this.selected
            };
          });
        }),
        catchError((err: HttpErrorResponse) => {
          this.toastService.showError('Error',
            'Unable to retrieve the list of reports. Please try again or contact support.');
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

  loadPage(page: number) {
    this.page = page - 1;
    this.reportsService.notifyRefresh({
      loading: true,
      sectionId: this.section.id,
      reportSelected: 0
    });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
}
