import { HttpErrorResponse } from '@angular/common/http';
import {ChangeDetectionStrategy, Component, Input, OnDestroy, OnInit, Output} from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import {EMPTY, Observable} from 'rxjs';
import {catchError, concatMap, filter, map, takeUntil, tap} from 'rxjs/operators';
import { CpReportsService } from 'src/app/compliance/shared/services/cp-reports.service';
import { ComplianceStandardSectionType } from 'src/app/compliance/shared/type/compliance-standard-section.type';
import { UtmToastService } from 'src/app/shared/alert/utm-toast.service';
import { SortByType } from 'src/app/shared/types/sort-by.type';
import { ComplianceReportType } from '../../../shared/type/compliance-report.type';


@Component({
    standalone: false,
    selector: 'app-compliance-print-view',
    templateUrl: './compliance-print-view.component.html',
    styleUrls: ['./compliance-print-view.component.css'],
    changeDetection: ChangeDetectionStrategy.OnPush,
})
export class CompliancePrintViewComponent implements OnInit, OnDestroy {
  @Input() section: ComplianceStandardSectionType;

  reports$: Observable<ComplianceReportType[]>;
  selected: number;
  fields: SortByType[];
  preparingPrint = true;
  visualizationCount = 0;
  loadedCount = 0;

  constructor(private reportsService: CpReportsService,
              private toastService: UtmToastService,
              private route: ActivatedRoute) { }

  ngOnInit() {
    this.reports$ = this.route.queryParams
    .pipe(
      filter((params) => !!params.section),
      map((params) => JSON.parse(decodeURIComponent(params.section))),
      tap(params => this.section = params),
        concatMap((params) => this.reportsService.fetchData({
          page: params.page,
          size: params.size,
          standardId: this.section.standardId,
          sectionId: this.section.id,
          expandDashboard: true,
          setStatus: true,
          sort: params.sort,
        })),
        map((res) => {
          return res.body.map((r, index) => {
            return {
              ...r,
              selected: index === this.selected
            };
          });
        }),
        tap((reports) => {
          // Count the evidence visualizations that will actually render: one per
          // report whose dashboard has a TABLE/LIST viz (that's what
          // app-compliance-status picks and what the template renders, firing
          // onVisualizationRun once each). r.visualization is not populated yet
          // at this point (it's set later by the template), so count from
          // r.dashboard — the field the backend actually returns.
          this.visualizationCount = reports.reduce((acc, r) => {
            const hasEvidence = (r.dashboard || []).some(v =>
              v && v.visualization && (v.visualization.chartType === 'TABLE_CHART' || v.visualization.chartType === 'LIST_CHART'));
            return acc + (hasEvidence ? 1 : 0);
          }, 0);
          if (this.visualizationCount === 0) {
            this.preparingPrint = false;
          }
        }),
        catchError((err: HttpErrorResponse) => {
          this.toastService.showError('Error',
            'Unable to retrieve the list of reports. Please try again or contact support.');
          return EMPTY;
        }));
  }

  onVisualizationChange(visualization: any, report: ComplianceReportType) {
    visualization.visualization = {
      ...visualization.visualization,
      page: {
        size: 5,
        page: 0
      }
    };
    report.visualization = visualization;
  }

  onVisualizationRun(){
    this.loadedCount += 1;
    if (this.loadedCount >= this.visualizationCount) {
      this.preparingPrint = false;
    }
  }

  ngOnDestroy(): void {
    throw new Error('Method not implemented.');
  }
}
