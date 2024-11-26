import {HttpErrorResponse} from '@angular/common/http';
import {AfterViewInit, Component, OnDestroy, OnInit} from '@angular/core';
import {NgbModal, NgbModalOptions} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {LocalStorageService} from 'ngx-webstorage';
import {EMPTY, Observable, Subject} from 'rxjs';
import {catchError, concatMap, filter, map, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ADMIN_ROLE} from '../../shared/constants/global.constant';
import {ExportPdfService} from '../../shared/services/util/export-pdf.service';
import {UtmCpStandardComponent} from '../shared/components/utm-cp-standard/utm-cp-standard.component';
import {CpReportsService} from '../shared/services/cp-reports.service';
import {CpStandardSectionService} from '../shared/services/cp-standard-section.service';
import {ComplianceReportType} from '../shared/type/compliance-report.type';
import {ComplianceStandardSectionType} from '../shared/type/compliance-standard-section.type';
import {ComplianceStandardType} from '../shared/type/compliance-standard.type';

@Component({
  selector: 'app-compliance-report-viewer',
  templateUrl: './compliance-report-viewer.component.html',
  styleUrls: ['./compliance-report-viewer.component.css']
})
export class ComplianceReportViewerComponent implements OnInit, AfterViewInit, OnDestroy {
  admin = ADMIN_ROLE;
  sections$: Observable<ComplianceStandardSectionType[]>;
  standard: ComplianceStandardType;
  activeIndexSection = 0;
  destroy$: Subject<void> = new Subject();
  report: ComplianceReportType;
  pdfExport = false;
  action: 'reports' | 'compliance' = 'compliance';
  activeSection: ComplianceStandardSectionType = null;

  constructor(private standardSectionService: CpStandardSectionService,
              private toastService: UtmToastService,
              private modalService: NgbModal,
              private $localStorage: LocalStorageService,
              private spinner: NgxSpinnerService,
              private reportsService: CpReportsService,
              private exportPdfService: ExportPdfService) {
    this.standard = this.$localStorage.retrieve('selectedStandard');
  }

  ngOnInit() {
    this.sections$ = this.standardSectionService.onRefresh$
      .pipe(takeUntil(this.destroy$),
            filter(refresh => !!refresh && refresh.loading),
            concatMap(() => this.standardSectionService.fetchData({
              page: 0,
              size: 1000,
              'standardId.equals': this.standard.id
            })),
            map((res) => {
              return res.body.map((s, index) => {
                return {
                  ...s,
                  isCollapsed: this.activeIndexSection === index,
                  isActive: this.activeIndexSection === index
                };
              });
            }),
            tap((sections) => {
              this.activeSection = {
                ...sections[this.activeIndexSection]
              };
            }),
            catchError((err: HttpErrorResponse) => {
              this.toastService.showError('Error',
                'Unable to retrieve the list of compliance standard sections. Please try again or contact support.');
              return EMPTY;
            }));

    this.reportsService.onLoadReport$
      .pipe(takeUntil(this.destroy$),
        filter((params) => !!params))
      .subscribe((params) => this.report = params.template);
  }

  ngAfterViewInit(): void {
    if (!this.standard) {
      this.manageStandards();
    } else {
      this.standardSectionService.notifyRefresh({
        loading: true,
        activeSection: 0
      });
    }
  }

  manageStandards() {
    const options: NgbModalOptions = {
      backdrop: 'static',
      keyboard: false,
      centered: true
    };
    const modalRef = this.modalService.open(UtmCpStandardComponent, options);

    modalRef.result.then((standard) => {
      this.standard = standard;
      this.standardSectionService.notifyRefresh({
        loading: true,
        activeSection: 0
      });
    });
  }

  onChangeSectionActive(index: number) {
    this.activeIndexSection = index;
    this.standardSectionService.notifyRefresh({
      loading: true,
      activeSection: index
    });
  }

  exportToPdf() {
    this.spinner.show('buildPrintPDF');
    const url = this.getUrl();
    const fileName = this.report.associatedDashboard.name.replace(/ /g, '_');
    this.exportPdfService.getPdf(url, fileName, 'PDF_TYPE_TOKEN').subscribe(response => {
      this.spinner.hide('buildPrintPDF').then(() =>
        this.exportPdfService.handlePdfResponse(response));
    }, error => {
      this.spinner.hide('buildPrintPDF').then(() =>
        this.toastService.showError('Error', 'An error occurred while creating a PDF.'));
    });
  }

  selectAction(action: 'reports' | 'compliance' ) {
    this.action = action;
  }

  trackFn(index: number, section: ComplianceStandardSectionType) {
    return section.id;
  }

  getUrl(){
    return this.action === 'reports' ? '/dashboard/export-compliance/' + this.report.id :
     `/compliance/print-view/?section=${this.getActiveSectionParams()}`;
  }

  getActiveSectionParams(){
   return encodeURIComponent(JSON.stringify({
    standardId: this.activeSection.standardId,
    id: this.activeSection.id
   }));
  }

  ngOnDestroy(): void {
    this.reportsService.loadReport(null);
    this.destroy$.next();
    this.destroy$.complete();
    this.standardSectionService.notifyRefresh(null);
  }
}
