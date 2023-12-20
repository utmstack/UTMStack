import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {CpReportsService} from '../../services/cp-reports.service';
import {ComplianceReportType} from '../../type/compliance-report.type';

@Component({
  selector: 'app-utm-report-info-view',
  templateUrl: './utm-report-info-view.component.html',
  styleUrls: ['./utm-report-info-view.component.scss']
})
export class UtmReportInfoViewComponent implements OnInit {
  @Input() complianceReports: ComplianceReportType[];
  @Input() standardId: number;
  @Output() complianceReportsChange = new EventEmitter<ComplianceReportType[]>();
  loadingTemplates = true;
  noMoreResult: boolean;
  page = 1;
  solution: string;
  loadingMore: false;
  reports: ComplianceReportType[];

  constructor(private cpReportsService: CpReportsService) {
  }

  ngOnInit() {
    this.reports = this.complianceReports.slice();
    if (!this.complianceReports) {
      this.complianceReports = [];
      // this.getReports();
    } else {
      this.loadingTemplates = false;
    }
  }


  // onScroll() {
  //   this.page += 1;
  //   this.getReports();
  // }

  deleteSection(index: number) {
    this.complianceReports.splice(index, 1);
    this.complianceReportsChange.emit(this.complianceReports);
  }

  getReports() {
    const query = {
      page: 0,
      size: 10000,
      sort: 'id,asc',
      standardId: this.standardId,
      'configSolution.contains': this.solution,
    };
    this.complianceReports = [];
    this.cpReportsService.queryByStandard(query).subscribe(response => {
      this.complianceReports = response.body;
      this.loadingTemplates = false;
      this.complianceReportsChange.emit(this.complianceReports);
    });
  }

  onSearchFor($event: string) {
    this.solution = $event;
    if ($event) {
      this.reports = this.reports.filter(value => value.configSolution.toLowerCase().includes($event.toLowerCase()));
    } else {
      this.reports = this.complianceReports.slice();
    }
  }

}
