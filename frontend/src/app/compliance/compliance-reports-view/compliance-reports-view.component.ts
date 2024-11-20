import {HttpErrorResponse} from '@angular/common/http';
import {ChangeDetectionStrategy, Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';
import {EMPTY, Observable} from "rxjs";
import {catchError, concatMap, filter, map, tap} from 'rxjs/operators';
import {SortByType} from '../../shared/types/sort-by.type';
import {CpReportsService} from '../shared/services/cp-reports.service';
import {ComplianceReportType} from '../shared/type/compliance-report.type';
import {ComplianceStandardSectionType} from '../shared/type/compliance-standard-section.type';
import {NgxSpinnerService} from "ngx-spinner";
import {ExportPdfService} from "../../shared/services/util/export-pdf.service";
import {UtmToastService} from "../../shared/alert/utm-toast.service";

@Component({
  selector: 'app-compliance-reports-view',
  templateUrl: './compliance-reports-view.component.html',
  styleUrls: ['./compliance-reports-view.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ComplianceReportsViewComponent implements OnInit, OnChanges {
  @Input() section: ComplianceStandardSectionType;

  reports$: Observable<ComplianceReportType[]>;
  selected: number;
  fields: SortByType[];

  reportDetail: ComplianceReportType;

  constructor(private reportsService: CpReportsService,
              private spinner: NgxSpinnerService,
              private exportPdfService: ExportPdfService,
              private toastService: UtmToastService) { }

  ngOnInit() {
    this.reports$ = this.reportsService.onRefresh$
      .pipe(filter(reportRefresh =>
          !!reportRefresh && reportRefresh.loading),
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
          console.log(res.body);
          return res.body.map((r, index) => {
            return {
              ...r,
              selected: index === this.selected
            };
          });
        }),
        tap((reports) => {
          /*if (this.loadFirst) {
            console.log('load first', this.loadFirst);
            this.loadReport(reports[0]);
            this.loadFirst = false;
          }*/
        }),
        catchError((err: HttpErrorResponse) => {
          this.toastService.showError('Error',
            'Unable to retrieve the list of reports. Please try again or contact support.');
          return EMPTY;
        }));
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes.section && changes.section.currentValue) {
      this.reportsService.notifyRefresh({
        loading: true,
        sectionId: changes.section.currentValue.id,
        reportSelected: 0
      });
    }
  }

  onStatusChange(status: boolean, report: ComplianceReportType) {
    report.status = status;
  }

  exportToPdf() {
    this.spinner.show('buildPrintPDF');
    const url = '/dashboard/export-compliance/' + this.reportDetail.id;
    const fileName = this.reportDetail.associatedDashboard.name.replace(/ /g, '_');
    this.exportPdfService.getPdf(url, fileName, 'PDF_TYPE_TOKEN').subscribe(response => {
      this.spinner.hide('buildPrintPDF').then(() =>
        this.exportPdfService.handlePdfResponse(response));
    }, error => {
      this.spinner.hide('buildPrintPDF').then(() =>
        this.toastService.showError('Error', 'An error occurred while creating a PDF.'));
    });
  }
}
