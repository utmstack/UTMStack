import {ChangeDetectionStrategy, Component, Input, OnInit} from '@angular/core';
import {EMPTY, Observable} from "rxjs";
import {ComplianceStandardSectionType} from "../../type/compliance-standard-section.type";
import {ComplianceReportType} from "../../type/compliance-report.type";
import {CpReportsService} from "../../services/cp-reports.service";
import {catchError, concatMap, filter, map} from "rxjs/operators";
import {HttpErrorResponse} from "@angular/common/http";
import {UtmToastService} from "../../../../shared/alert/utm-toast.service";

@Component({
  selector: 'app-utm-cp-section',
  templateUrl: './utm-cp-section.component.html',
  styleUrls: ['./utm-cp-section.component.css',],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class UtmCpSectionComponent implements OnInit {

  @Input() section: ComplianceStandardSectionType;
  reports$: Observable<ComplianceReportType[]>;

  constructor(private reportsService: CpReportsService,
              private toastService: UtmToastService) { }

  ngOnInit() {
    this.reports$ = this.reportsService.onRefresh$
      .pipe(filter(reportRefresh => !!reportRefresh &&
          reportRefresh.loading && reportRefresh.sectionId === this.section.id),
        concatMap(() => this.reportsService.fetchData({
          page: 0,
          size: 1000,
          standardId: this.section.standardId,
          sectionId: this.section.id
        })),
        map((res) => res.body),
        catchError((err: HttpErrorResponse) => {
          this.toastService.showError('Error',
            'Unable to retrieve the list of reports. Please try again or contact support.');
          return EMPTY;
        }));
  }
  loadReport() {
    if (this.section.isCollapsed) {
      this.reportsService.notifyRefresh({
        sectionId: this.section.id,
        loading: true
      });
    }
  }

  generateReport(report: ComplianceReportType) {
    this.reportsService.loadReport({
      template: report.id,
      sectionId: this.section.id,
      standardId: this.section.standardId
    });
  }
}
