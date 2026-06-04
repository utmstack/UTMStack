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
  selector: 'app-compliance-print-view',
  templateUrl: './compliance-print-view.component.html',
  styleUrls: ['./compliance-print-view.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class CompliancePrintViewComponent implements OnInit, OnDestroy {
  @Input() section: ComplianceStandardSectionType;

  reports$: Observable<ComplianceReportType[]>;
  selected: number;
  fields: SortByType[];
  preparingPrint = true;

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

  onVisualizationLoaded(){
    this.preparingPrint = false;
  }

  ngOnDestroy(): void {
    throw new Error('Method not implemented.');
  }
}
