import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {ADMIN_ROLE} from '../../shared/constants/global.constant';
import {ReportCustomExportModalComponent} from '../shared/component/report-custom-export-modal/report-custom-export-modal.component';
import {UtmReportCreateComponent} from '../shared/component/utm-report-create/utm-report-create.component';
import {UtmReportSectionCreateComponent} from '../shared/component/utm-report-section-create/utm-report-section-create.component';
import {ReportParamsEnum} from '../shared/enums/report-params.enum';
import {ReportTypeEnum} from '../shared/enums/report-type.enum';
import {ReportService} from '../shared/service/report.service';
import {SectionReportService} from '../shared/service/section-report.service';
import {ReportSectionType} from '../shared/type/report-section.type';
import {ReportType} from '../shared/type/report.type';

@Component({
  selector: 'app-report-template',
  templateUrl: './report-template.component.html',
  styleUrls: ['./report-template.component.scss']
})
export class ReportTemplateComponent implements OnInit {
  loadingTemplates: any;
  admin = ADMIN_ROLE;
  sections: ReportSectionType[];
  sectionId: number;
  section: ReportSectionType;
  reports: ReportType[] = [];
  loadingMore: boolean;
  noMoreResult: boolean;
  manage: boolean;
  search: string;
  reportTypeEnum = ReportTypeEnum;

  constructor(private modalService: NgbModal,
              private reportService: ReportService,
              private router: Router,
              private sectionReportService: SectionReportService,
              private utmToastService: UtmToastService,
              private activatedRoute: ActivatedRoute) {
    activatedRoute.queryParams.subscribe(params => {
      if (params.onInitAction) {
        this.manage = true;
      }
    });
  }

  ngOnInit() {
    this.getReportSections(true);
  }

  getReportSections(initialLoad?: boolean) {
    this.sectionReportService.query({size: 500}).subscribe(response => {
      this.sections = response.body;
      if (this.sections.length > 0 && initialLoad) {
        this.sectionId = this.sections[0].id;
        this.section = this.sections[0];
        this.getReportBySection(this.sectionId);
      }
    });
  }

  getReportBySection(sectionId: number, reportName?: string) {
    const req = {
      size: 1000,
      'reportSectionId.equals': sectionId,
      'repName.contains': reportName
    };
    this.reportService.query(req).subscribe(response => {
      this.reports = response.body;
    });
  }

  onSearchFor($event: string) {
    this.search = $event;
    this.getReportBySection(this.sectionId, $event);
  }

  generateReport(report: ReportType) {
    const queryParams = {};
    switch (report.repType) {
      case ReportTypeEnum.TEMPLATE:
        queryParams[ReportParamsEnum.TEMPLATE] = report.id;
        this.router.navigate(['reports/template-result'], {
          queryParams
        });
        break;
      case ReportTypeEnum.CUSTOM_LIST:
        queryParams[ReportParamsEnum.TEMPLATE_ID] = report.id;
        queryParams[ReportParamsEnum.SECTION_ID] = report.id;
        this.router.navigate(['reports/template-view'], {
          queryParams
        });
        break;
      case ReportTypeEnum.CUSTOM_PDF:
        const modalExport = this.modalService.open(ReportCustomExportModalComponent, {centered: true});
        modalExport.componentInstance.report = report;
        break;
    }
  }

  changeSection(st: ReportSectionType) {
    this.sectionId = st.id;
    this.section = st;
    this.getReportBySection(st.id);
  }

  newSection() {
    const sectionModalCreate = this.modalService.open(UtmReportSectionCreateComponent, {centered: true});
    sectionModalCreate.componentInstance.sectionSaved.subscribe(() => {
      this.getReportSections();
    });
  }

  editSection(st: ReportSectionType) {
    const sectionModalEdit = this.modalService.open(UtmReportSectionCreateComponent, {centered: true});
    sectionModalEdit.componentInstance.section = st;
    sectionModalEdit.componentInstance.sectionSaved.subscribe(() => {
      this.getReportSections();
    });
  }

  deleteSection(st: ReportSectionType) {
    const deleteSection = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteSection.componentInstance.header = 'Delete section report';
    deleteSection.componentInstance.message = 'All reports in this section will be deleted. ' +
      'Are you sure that you want to delete this section?';
    deleteSection.componentInstance.confirmBtnText = 'Delete';
    deleteSection.componentInstance.confirmBtnIcon = 'icon-ungroup';
    deleteSection.componentInstance.confirmBtnType = 'delete';
    deleteSection.result.then(() => {
      this.deleteSectionReport(st);
    });
  }

  deleteSectionReport(section: ReportSectionType) {
    this.sectionReportService.delete(section.id).subscribe(() => {
      this.utmToastService.showSuccessBottom('Section report deleted successfully');
      const index = this.sections.findIndex(value => value.id === section.id);
      if (section.id === this.sectionId) {
        this.sectionId = this.sections[0].id;
        this.getReportBySection(this.sectionId);
      }
      this.sections.splice(index, 1);
    }, error1 => {
      this.utmToastService.showError('Error', 'Error deleting section report');
    });
  }


  createReport() {
    const modalCreateReport = this.modalService.open(UtmReportCreateComponent, {centered: true});
    modalCreateReport.componentInstance.reportCreated.subscribe(() => {
      this.getReportBySection(this.sectionId);
    });
  }

  editReport(rp: ReportType) {
    const modalEditReport = this.modalService.open(UtmReportCreateComponent, {centered: true});
    modalEditReport.componentInstance.report = rp;
    modalEditReport.componentInstance.reportCreated.subscribe(() => {
      this.getReportBySection(this.sectionId);
    });
  }

  deleteReport(rp: ReportType) {
    const deleteReport = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteReport.componentInstance.header = 'Delete report';
    deleteReport.componentInstance.message = 'Are you sure that you want to delete this report?';
    deleteReport.componentInstance.confirmBtnText = 'Delete';
    deleteReport.componentInstance.confirmBtnIcon = 'icon-clipboard4';
    deleteReport.componentInstance.confirmBtnType = 'delete';
    deleteReport.result.then(() => {
      this.deleteAction(rp);
    });
  }

  deleteAction(rp: ReportType) {
    this.reportService.delete(rp.id).subscribe(() => {
      this.utmToastService.showSuccessBottom('Report deleted successfully');
      const index = this.reports.findIndex(value => value.id === rp.id);
      this.reports.splice(index, 1);
    }, error1 => {
      this.utmToastService.showError('Error', 'Error deleting report');
    });
  }

  navigateToManage() {
    this.router.navigate(['/reports/templates'], {queryParams: {onInitAction: 'manageReport'}});
    this.manage = true;
  }

  navigateToView() {
    this.router.navigate(['/reports/templates']);
    this.manage = false;
  }
}
