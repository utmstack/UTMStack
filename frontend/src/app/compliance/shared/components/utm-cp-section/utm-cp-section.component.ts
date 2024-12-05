import {HttpErrorResponse} from '@angular/common/http';
import {ChangeDetectionStrategy, Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {EMPTY, Observable} from 'rxjs';
import {catchError, concatMap, filter, map, tap} from 'rxjs/operators';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {CpReportsService} from '../../services/cp-reports.service';
import {ComplianceReportType} from '../../type/compliance-report.type';
import {ComplianceStandardSectionType} from '../../type/compliance-standard-section.type';

@Component({
  selector: 'app-utm-cp-section',
  templateUrl: './utm-cp-section.component.html',
  styleUrls: ['./utm-cp-section.component.css', ],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class UtmCpSectionComponent implements OnInit {

  @Input() section: ComplianceStandardSectionType;
  @Input() index: number;
  @Output() isActive: EventEmitter<number> = new EventEmitter<number>();
  @Input() loadFirst = false;
  @Input() expandable = true;

  reports$: Observable<ComplianceReportType[]>;
  selected: number;

  constructor(private reportsService: CpReportsService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
    if (this.section.isActive && this.index === 0) {
      this.reportsService.notifyRefresh({
        sectionId: this.section.id,
        loading: true,
        reportSelected: 0
      });
    }
    this.reports$ = this.reportsService.onRefresh$
      .pipe(filter(reportRefresh =>
          !!reportRefresh && reportRefresh.loading && reportRefresh.sectionId === this.section.id),
        tap((reportRefresh) => {
          this.selected = reportRefresh.reportSelected;
        }),
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
        tap((reports) => {
          if (this.loadFirst) {
            this.loadReport(reports[0]);
            this.loadFirst = false;
          }
        }),
        catchError((err: HttpErrorResponse) => {
          this.toastService.showError('Error',
            'Unable to retrieve the list of reports. Please try again or contact support.');
          return EMPTY;
        }));
  }

  loadReports() {
      this.reportsService.notifyRefresh({
        sectionId: this.section.id,
        loading: true,
        reportSelected: 0
      });
      if (!this.section.isActive) {
        this.isActive.emit(this.index);
      }
  }

  generateReport(index: number, report: ComplianceReportType) {
    if (this.section.isActive) {
      this.reportsService.notifyRefresh({
        sectionId: this.section.id,
        loading: true,
        reportSelected: index
      });
    }
    this.loadReport(report);
  }

  loadReport(report: ComplianceReportType) {
    this.reportsService.loadReport({
      template: report,
      sectionId: this.section.id,
      standardId: this.section.standardId
    });
  }
}
