import {Component, Input, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {CpReportBehavior} from '../../shared/behavior/cp-report.behavior';
import {CpStandardSectionBehavior} from '../../shared/behavior/cp-standard-section.behavior';
import {UtmComplianceCreateComponent} from '../../shared/components/utm-compliance-create/utm-compliance-create.component';
import {CpReportsService} from '../../shared/services/cp-reports.service';
import {ComplianceReportType} from '../../shared/type/compliance-report.type';
import {ComplianceStandardSectionType} from '../../shared/type/compliance-standard-section.type';
import {UtmCpReportDeleteComponent} from './utm-cp-report-delete/utm-cp-report-delete.component';

@Component({
  selector: 'app-utm-cp-reports',
  templateUrl: './utm-cp-reports.component.html',
  styleUrls: ['./utm-cp-reports.component.scss']
})
export class UtmCpReportsComponent implements OnInit {
  section: ComplianceStandardSectionType;
  @Input() manage: boolean;
  complianceReports: ComplianceReportType[] = [];
  loadingTemplates = true;
  noMoreResult: boolean;
  page = 1;
  solution: string;
  loadingMore: false;
  showDetailFor = 0;

  constructor(private cpReportsService: CpReportsService,
              public cpStandardSectionBehavior: CpStandardSectionBehavior,
              private cpReportBehavior: CpReportBehavior,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    this.cpStandardSectionBehavior.$standardSection.subscribe(section => {
      if (section) {
        this.section = section;
        this.getReports();
      } else {
        this.complianceReports = [];
        this.section = null;
        this.loadingTemplates = false;
      }
    });

    this.cpReportBehavior.$reportUpdate.subscribe(update => {
      if (update) {
        this.getReports();
      }
    });

  }

  getReports() {
    const query = {
      page: this.page - 1,
      size: 1000,
      sort: 'id,asc',
      'standardSectionId.equals': this.section.id,
      'configSolution.contains': this.solution
    };
    this.complianceReports = [];
    this.cpReportsService.query(query).subscribe(response => {
      this.complianceReports = response.body;
      this.loadingTemplates = false;
    });
  }

  onSearchFor($event: string) {
    this.solution = $event;
    this.getReports();
  }

  onScroll() {
    // TODO
  }

  deleteSection(report: ComplianceReportType) {
    const modal = this.modalService.open(UtmCpReportDeleteComponent, {centered: true});
    modal.componentInstance.report = report;
    modal.componentInstance.reportDelete.subscribe(() => {
      this.getReports();
    });
  }

  editReport(report: ComplianceReportType) {
    const reportModal = this.modalService.open(UtmComplianceCreateComponent, {centered: true});
    reportModal.componentInstance.report = report;
    reportModal.componentInstance.reportCreated.subscribe(() => {
      this.getReports();
    });
  }

  toggleDetail(id: number) {
    this.showDetailFor = this.showDetailFor === id ? 0 : id;
  }
}
