import {HttpErrorResponse} from '@angular/common/http';
import {AfterViewInit, Component, OnInit} from '@angular/core';
import {NgbModal, NgbModalOptions} from '@ng-bootstrap/ng-bootstrap';
import {EMPTY, Observable} from 'rxjs';
import {catchError, concatMap, filter, map} from 'rxjs/operators';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ADMIN_ROLE} from '../../shared/constants/global.constant';
import {UtmCpStandardComponent} from '../shared/components/utm-cp-standard/utm-cp-standard.component';
import {CpReportsService} from '../shared/services/cp-reports.service';
import {CpStandardSectionService} from '../shared/services/cp-standard-section.service';
import {ComplianceStandardSectionType} from '../shared/type/compliance-standard-section.type';
import {ComplianceStandardType} from '../shared/type/compliance-standard.type';

@Component({
  selector: 'app-compliance-report-viewer',
  templateUrl: './compliance-report-viewer.component.html',
  styleUrls: ['./compliance-report-viewer.component.css']
})
export class ComplianceReportViewerComponent implements OnInit, AfterViewInit {
  admin = ADMIN_ROLE;
  sections$: Observable<ComplianceStandardSectionType[]>;
  standard: ComplianceStandardType;
  selectedSection: ComplianceStandardSectionType;

  constructor(private standardSectionService: CpStandardSectionService,
              private reportsService: CpReportsService,
              private toastService: UtmToastService,
              private modalService: NgbModal) { }

  ngOnInit() {
    this.sections$ = this.standardSectionService.onRefresh$
      .pipe(filter(refresh => !!refresh),
            concatMap(() => this.standardSectionService.fetchData({
              page: 0,
              size: 1000,
              'standardId.equals': this.standard.id
            })),
            map((res) => {
              return res.body.map(s => {
                return {
                  ...s,
                  isCollapsed: false
                };
              });
            }),
            catchError((err: HttpErrorResponse) => {
              this.toastService.showError('Error',
                'Unable to retrieve the list of compliance standard sections. Please try again or contact support.');
              return EMPTY;
            }));
  }

  ngAfterViewInit(): void {
    this.manageStandards();
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
      this.standardSectionService.notifyRefresh(true);
    });
  }
}
