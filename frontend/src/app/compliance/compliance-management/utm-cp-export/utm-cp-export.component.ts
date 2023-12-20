import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import * as moment from 'moment';
import {CpReportsService} from '../../shared/services/cp-reports.service';
import {ComplianceReportType} from '../../shared/type/compliance-report.type';

@Component({
  selector: 'app-utm-cp-export',
  templateUrl: './utm-cp-export.component.html',
  styleUrls: ['./utm-cp-export.component.scss']
})
export class UtmCpExportComponent implements OnInit {
  @Input() standardId: number;
  complianceReports: ComplianceReportType[] = [];
  creating = false;
  loadingCompliance = true;

  constructor(private cpReportsService: CpReportsService,
              public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
    this.loadingCompliance = true;
    this.getReportsExport(this.standardId);
  }

  getReportsExport(standardId?: number) {
    this.cpReportsService.queryByStandard({
      size: 10000,
      expandDashboard: true,
      standardId
    }).subscribe(response => {
      this.complianceReports = response.body;
      this.loadingCompliance = false;
    });
  }


  export() {
    this.creating = true;
    this.removeId().then(reports => {
      const dataStr = JSON.stringify(reports);
      const dataUri = 'data:application/json;charset=utf-8,' + encodeURIComponent(dataStr);
      const exportFileDefaultName = 'UTMCompliance-' + moment(new Date()).format('YYYY-MM-DD') + '.json';
      const linkElement = document.createElement('a');
      linkElement.setAttribute('href', dataUri);
      linkElement.setAttribute('download', exportFileDefaultName);
      linkElement.click();
      this.creating = false;
      this.activeModal.close();
    });
  }


  removeId(): Promise<ComplianceReportType[]> {
    return new Promise<ComplianceReportType[]>((resolve) => {
      const reports: ComplianceReportType[] = [];
      for (const vis of this.complianceReports) {
        Object.keys(vis).forEach(key => {
          if (key === 'id') {
            delete vis[key];
          }
        });
        reports.push(vis);
      }
      resolve(reports);
    });
  }
}
