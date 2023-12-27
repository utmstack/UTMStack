import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ReportFormatEnum} from '../../../../shared/enums/report-format.enum';
import {AssetHostDetailService} from '../../../assets-discovery/shared/services/asset-host-detail.service';

@Component({
  selector: 'app-scanner-export-vulnerabilities',
  templateUrl: './scanner-export-vulnerabilities.component.html',
  styleUrls: ['./scanner-export-vulnerabilities.component.scss']
})
export class ScannerExportVulnerabilitiesComponent implements OnInit {
  @Input() reportId: string;
  generateReport = false;
  fullReport = true;
  reportFormat = ReportFormatEnum.PDF;

  constructor(public activeModal: NgbActiveModal,
              private toastService: UtmToastService,
              private assetHostDetailService: AssetHostDetailService) {
  }

  ngOnInit() {
  }

  viewPdf(dat) {
    this.activeModal.close();
    const data = new Blob([dat], {type: 'application/pdf'});
    this.generateReport = false;
    if (data.size > 0) {
      const reader = new FileReader();
      reader.readAsDataURL(data);
      reader.onloadend = () => {
        const base64data = reader.result.toString();
        const linkElement = document.createElement('a');
        linkElement.setAttribute('href', base64data);
        linkElement.setAttribute('download', 'UTMStack vulnerabilities report');
        linkElement.click();
      };
      // const fileURL = window.URL.createObjectURL(data);
      // window.open(fileURL);
    } else {
      this.toastService.showWarning('NO DATA FOUND', 'No data found for this report');
    }
  }

  viewReportTaskDetail() {
    this.generateReport = true;
    this.assetHostDetailService.exportVulnerabilitiesToPdf(this.fullReport, this.reportId, this.reportFormat)
      .subscribe(pdf => {
        this.viewPdf(pdf);
      });
  }
}
