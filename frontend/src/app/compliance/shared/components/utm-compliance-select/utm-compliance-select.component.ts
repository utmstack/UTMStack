import {HttpResponse} from '@angular/common/http';
import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {SortEvent} from '../../../../shared/directives/sortable/type/sort-event';
import {ComplianceScheduleService} from '../../services/compliance-schedule.service';
import {CpReportsService} from '../../services/cp-reports.service';
import {CpStandardSectionService} from '../../services/cp-standard-section.service';
import {CpStandardService} from '../../services/cp-standard.service';
import {ComplianceReportType} from '../../type/compliance-report.type';
import {ComplianceScheduleType} from '../../type/compliance-schedule.type';
import {ComplianceStandardSectionType} from '../../type/compliance-standard-section.type';
import {ComplianceStandardType} from '../../type/compliance-standard.type';

@Component({
  selector: 'app-utm-compliance-select',
  templateUrl: './utm-compliance-select.component.html',
  styleUrls: ['./utm-compliance-select.component.scss']
})
export class UtmComplianceSelectComponent implements OnInit {
  @Input() idReport: number;
  @Output() reportSelected = new EventEmitter<ComplianceReportType>();
  private requestParams: any;
  private sortBy: SortEvent;
  report: ComplianceReportType;
  standards: ComplianceStandardType[] = [];
  standard: number;
  solution: string;
  standardSections: ComplianceStandardSectionType[] = [];
  section: number;
  loading = true;
  totalItems: any;
  page = 1;
  itemsPerPage = 10;
  complianceReports: ComplianceReportType[] = [];
  searching = false;

  constructor(private cpReportsService: CpReportsService,
              private utmToastService: UtmToastService,
              private cpStandardService: CpStandardService,
              private cpStandardSectionService: CpStandardSectionService,
              private complianceScheduleService: ComplianceScheduleService) {}

  ngOnInit() {
    this.requestParams = {
      page: this.page - 1,
      size: this.itemsPerPage,
      sort: this.sortBy,
      'name.contains': null
    };
    this.getStandardList();
  }

  onSortBy($event) {
  }

  getSelectedDashboard(id: number) {
    this.complianceScheduleService.find(id).subscribe(response => {
      const report: ComplianceScheduleType = response.body;
      this.selectDashboard(response.body.compliance);
    });
  }

  loadPage(page: any) {
    this.requestParams.page = page - 1;
    this.getDashboardList();
  }

  getDashboardList() {
    const query = {
      page: this.page - 1,
      size: 1000,
      sort: 'id,asc',
      'standardSectionId.equals': this.section,
      'configSolution.contains': this.solution
    };
    this.cpReportsService.query(query).subscribe(
      (res: HttpResponse<any>) => this.onSuccess(res.body, res.headers),
      (res: HttpResponse<any>) => this.onError(res.body)
    );
  }

  getSections() {
    const query = {
      page: this.page - 1,
      size: 1000,
      sort: 'id,asc',
      'standardId.equals': this.standard,
      'standardSectionName.contains': this.solution
    };
    this.cpStandardSectionService.query(query).subscribe(response => {
      this.standardSections = response.body;
      this.section = this.standardSections[0].id;
      this.getDashboardList();
      if (this.idReport) {
        this.getSelectedDashboard(this.idReport);
      }
    });
  }

  getStandardList() {
    this.cpStandardService.query({page: 0, size: 1000}).subscribe(
      (res: HttpResponse<any>) => {
        this.standards = res.body;
        this.standard = this.standards[0].id;
        this.getSections();
      },
      (res: HttpResponse<any>) => this.onError(res)
    );
  }

  onSearchDashboard($event: string) {
    this.searching = true;
    this.solution = $event;
    this.getDashboardList();
  }

  selectDashboard(report: ComplianceReportType) {
    this.report = report;
    this.idReport = report.id;
    this.reportSelected.emit(report);
  }

  private onSuccess(data, headers) {
    this.totalItems = headers.get('X-Total-Count');
    this.complianceReports = data;
    this.loading = false;
    this.searching = false;
  }

  private onError(res: any) {
    this.utmToastService.showErrorResponse('Error', res);
  }

  filterBySelect($event: {}, source: string) {
    this.getSections();
  }
}
