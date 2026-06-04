import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {AssetActiveFilterType} from '../../types/asset-active-filter.type';
import {OmpReportService} from './shared/services/omp-report.service';

@Component({
  selector: 'app-asset-save-report',
  templateUrl: './asset-save-report.component.html',
  styleUrls: ['./asset-save-report.component.scss']
})
export class AssetSaveReportComponent implements OnInit {
  saving = false;
  generateReport = false;
  @Input() filter: AssetActiveFilterType;
  @Input() type: 'asset' | 'detail' | 'task';
  limitRange = [50, 100, 200, 500, 1000, 2500];
  limit = 50;
  name: any;
  description: any;

  constructor(public activeModal: NgbActiveModal,
              private toastService: UtmToastService,
              private ompReportService: OmpReportService) {
  }

  ngOnInit() {
  }

  applyLimit(limit) {
    this.limit = limit;
  }

  viewReportAssets() {
    this.generateReport = true;
    const params = {
      modified: {
        greaterThan: this.filter['created.greaterThan'] !== undefined ?
          new Date(this.filter['created.greaterThan']).getTime() / 1000 : undefined,
        lessThan: this.filter['created.greaterThan'] !== undefined ?
          new Date(this.filter['created.lessThan']).getTime() / 1000 : undefined
      },
      hostIp: {
        contains: this.filter['hostIp.contains'],
      },
      hostOs: {
        contains: this.filter['hostOs.contains'],
      },
      hostSeverity: {
        greaterThan: this.filter['hostSeverity.greaterThan'],
        lessThan: this.filter['hostSeverity.lessThan'],
        equals: this.filter['hostSeverity.equals']
      },
      name: {
        contains: this.filter['name.contains'],
      }
    };
    Object.keys(params).forEach(value => {
      Object.keys(params[value]).forEach(child => {
        if (params[value] !== null && params[value][child] === undefined) {
          delete params[value][child];
          if (this.isEmptyObject(params[value])) {
            delete params[value];
          }
        }
      });
    });
    this.ompReportService.exportAssetsToPdf(params, this.limit).subscribe((dat) => {
      this.viewPdf(dat);
    });
  }

  viewReportAssetsDetail() {
    this.generateReport = true;
    const params = {
      host: {
        equals: this.filter.host
      },
      minQod: {
        greaterThan: this.filter['minQod.greaterThan'],
      },
    };
    this.exportResult(params);

  }

  viewReportTaskDetail() {
    this.generateReport = true;
    const params = {
      taskName: {
        equals: this.filter.task
      },
      minQod: {
        greaterThan: this.filter['minQod.greaterThan'],
      },
      reportId: {
        equals: this.filter.reportId
      },
    };
    this.exportResultTask(params);
  }

  exportResult(params) {
    this.ompReportService.exportAssetsResultToPdf(params, this.limit).subscribe((dat) => {
      this.viewPdf(dat);
    });
  }

  exportResultTask(params) {
    this.ompReportService.exportTaskResultToPdf(params, this.limit).subscribe((dat) => {
      this.viewPdf(dat);
    });
  }

  isEmptyObject(obj) {
    return (Object.getOwnPropertyNames(obj).length === 0);
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
        linkElement.setAttribute('download', this.resolveFileName());
        linkElement.click();
      };
      // const fileURL = window.URL.createObjectURL(data);
      // window.open(fileURL);
    } else {
      this.toastService.showWarning('NO DATA FOUND', 'No data found for this report');
    }
  }

  resolveFileName() {
    switch (this.type) {
      case 'asset':
        return 'UTMStack Assets report';
      case 'detail':
        return 'UTMStack Asset vulnerabilities detail report';
      case 'task':
        return 'UTMStack Task report';

    }
  }


}
