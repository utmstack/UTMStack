import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ReportShortNameEnum} from '../../enums/report-short-name.enum';
import {ReportService} from '../../service/report.service';
import {ReportType} from '../../type/report.type';

@Component({
  selector: 'app-report-custom-export-modal',
  templateUrl: './report-custom-export-modal.component.html',
  styleUrls: ['./report-custom-export-modal.component.css']
})
export class ReportCustomExportModalComponent implements OnInit {
  @Input() report: ReportType;
  pdfExport = false;
  reportShortNameEnum = ReportShortNameEnum;
  params = {};

  constructor(public activeModal: NgbActiveModal,
              private toastService: UtmToastService,
              private reportService: ReportService) {
  }

  ngOnInit() {
  }

  generateReport() {
    this.pdfExport = true;
    if (this.report.repHttpMethod === 'GET') {
      this.reportService.exportReportToPdfGET(this.params, this.report.repUrl).subscribe(data => {
        this.viewPdf(data.body);
      });
    } else {
      this.reportService.exportReportToPdfPOST(this.params, this.report.repUrl).subscribe(data => {
        this.viewPdf(data.body);
      });
    }
  }

  viewPdf(dat) {
    const data = new Blob([dat], {type: 'application/pdf'});
    const date = new Date();
    if (data.size > 0) {
      const reader = new FileReader();
      reader.readAsDataURL(data);
      reader.onloadend = () => {
        const base64data = reader.result.toString();
        this.activeModal.close();
        const linkElement = document.createElement('a');
        linkElement.setAttribute('href', base64data);
        linkElement.setAttribute('download', (this.report.repName + date.getDay() + '-' + date.getMonth() + '-' + date.getFullYear()));
        linkElement.click();
      };
      // const fileURL = window.URL.createObjectURL(data);
      // window.open(fileURL);
    } else {
      this.pdfExport = false;
      this.toastService.showWarning('NO DATA FOUND', 'No data found for this report');
    }
  }

  paramsIsValid(): boolean {
    return this.report.repShortName === ReportShortNameEnum.AD_USR_PWD_EXPIRED ? true : Object.keys(this.params).length > 0;
  }
}
