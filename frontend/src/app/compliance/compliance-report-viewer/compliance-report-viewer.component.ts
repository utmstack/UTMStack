import {HttpErrorResponse} from '@angular/common/http';
import {AfterViewInit, Component, HostListener, OnDestroy, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal, NgbModalOptions} from '@ng-bootstrap/ng-bootstrap';
import {NgxSpinnerService} from 'ngx-spinner';
import {LocalStorageService} from 'ngx-webstorage';
import {EMPTY, Observable, Subject} from 'rxjs';
import {catchError, concatMap, debounceTime, distinctUntilChanged, filter, map, takeUntil, tap} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ADMIN_ROLE} from '../../shared/constants/global.constant';
import {ExportPdfService} from '../../shared/services/util/export-pdf.service';
import {UtmCpStandardComponent} from '../shared/components/utm-cp-standard/utm-cp-standard.component';
import {ComplianceParamsEnum} from '../shared/enums/compliance-params.enum';
import {CpReportsService} from '../shared/services/cp-reports.service';
import {CpStandardSectionService} from '../shared/services/cp-standard-section.service';
import {CpStandardService} from '../shared/services/cp-standard.service';
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
  pageable = {
    page: 0,
    size: 15,
    sort: 'configReportName,desc'
  };
  viewportHeight: number;

  constructor(private standardSectionService: CpStandardSectionService,
              private toastService: UtmToastService,
              private modalService: NgbModal,
              private spinner: NgxSpinnerService,
              private reportsService: CpReportsService,
              private exportPdfService: ExportPdfService,
              private activatedRoute: ActivatedRoute,
              private standardService: CpStandardService,
              private router: Router,
              private localStorage: LocalStorageService) {

    this.standard = this.localStorage.retrieve('selectedStandard');
    this.viewportHeight = window.innerHeight;
  }

  @HostListener('window:resize', ['$event'])
  onResize(event: Event) {
    this.viewportHeight = window.innerHeight;
  }

  ngOnInit() {
    this.activatedRoute.queryParams
      .pipe(
        debounceTime(100),
        distinctUntilChanged(),
        filter((params) => params && params[ComplianceParamsEnum.STANDARD_ID]),
        map((params) => params[ComplianceParamsEnum.STANDARD_ID]),
        concatMap( standardId => this.standardService.find(standardId))
      ).pipe(
        map((response) => response.body),
        tap((standard) => {
          this.standard = standard;
          this.activeIndexSection = 0;
          this.standardSectionService.notifyRefresh({
            loading: true,
            activeSection: 0
          });
        })
    ).subscribe();

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

      this.router.navigate([], {
        relativeTo: this.activatedRoute,
        queryParams: { standardId: this.standard.id },
        queryParamsHandling: 'merge',
      });
    });
  }

  onChangeSectionActive(index: number, section: ComplianceStandardSectionType, sections: ComplianceStandardSectionType[]) {
    this.activeIndexSection = index;
    sections.forEach((sec, i) => {
      if (i === index) {
        sec.isCollapsed = true;
        sec.isActive = true;
        this.activeSection = section;
      } else {
        sec.isCollapsed = false;
        sec.isActive = false;
      }
    });
  }

  exportToPdf() {
    this.spinner.show('buildPrintPDF');
    const url = this.getUrl();
    const fileName = this.report ? this.report.associatedDashboard.name.replace(/ /g, '_') : 'Reports';
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

  getUrl() {
    if (this.action === 'reports') {
        return '/dashboard/export-compliance/' + this.report.id;
    } else {
      const section = this.getActiveSectionParams();
      return  encodeURIComponent('/compliance/print-view?section=' + section);
    }
  }

  getActiveSectionParams() {
   return encodeURIComponent(
     JSON.stringify({
        standardId: this.activeSection.standardId,
        id: this.activeSection.id,
        page: this.pageable.page,
        size: this.pageable.size,
        sort: this.pageable.sort
   }));
  }

  onPageChange($event: any) {
    this.pageable = $event;
  }

  getMenuHeight() {
    return 100 - ((150 / this.viewportHeight) * 100) + 'vh';
  }

  ngOnDestroy(): void {
    this.reportsService.loadReport(null);
    this.destroy$.next();
    this.destroy$.complete();
    this.standardSectionService.notifyRefresh(null);
  }
}
