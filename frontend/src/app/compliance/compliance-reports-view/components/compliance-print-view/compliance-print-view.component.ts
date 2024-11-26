import {ChangeDetectionStrategy, Component, Input, OnDestroy, OnInit, Output} from '@angular/core';
import {EMPTY, Observable, Subject} from 'rxjs';
import {catchError, concatMap, filter, map, takeUntil, tap} from 'rxjs/operators';
import { ComplianceReportType } from '../../../shared/type/compliance-report.type';
import {TimeWindowsService} from '../../../shared/components/utm-cp-section/time-windows.service';
import { HttpErrorResponse } from '@angular/common/http';
import { NgxSpinnerService } from 'ngx-spinner';
import { CpReportsService } from 'src/app/compliance/shared/services/cp-reports.service';
import { ComplianceStandardSectionType } from 'src/app/compliance/shared/type/compliance-standard-section.type';
import { UtmToastService } from 'src/app/shared/alert/utm-toast.service';
import { ExportPdfService } from 'src/app/shared/services/util/export-pdf.service';
import { SortByType } from 'src/app/shared/types/sort-by.type';
import { ActivatedRoute } from '@angular/router';


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

  reportDetail: ComplianceReportType;

  constructor(private reportsService: CpReportsService,
              private spinner: NgxSpinnerService,
              private exportPdfService: ExportPdfService,
              private toastService: UtmToastService,
              private route: ActivatedRoute) { }

  ngOnInit() {
    this.reports$ = this.route.queryParams
    .pipe(
      filter((params) => !!params['section']),
      map((params) => JSON.parse(decodeURIComponent(params['section']))),
      tap(section => this.section = section),
        concatMap(() => this.reportsService.fetchData({
          page: 0,
          size: 1000,
          standardId: this.section.standardId,
          sectionId: this.section.id
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
    report.visualization = visualization;
  }

  ngOnDestroy(): void {
    throw new Error('Method not implemented.');
  }

  onVisualizationLoaded(){

  }


}
