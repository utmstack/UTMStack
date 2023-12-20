import {Component, Input, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {ITEMS_PER_PAGE} from '../../../../shared/constants/pagination.constants';
import {UtmNetScanService} from '../../services/utm-net-scan.service';
import {AssetFilterType} from '../../types/asset-filter.type';

@Component({
  selector: 'app-asset-save-report',
  templateUrl: './asset-save-report.component.html',
  styleUrls: ['./asset-save-report.component.scss']
})
export class AssetSaveReportComponent implements OnInit {
  @Input() assetFilters: AssetFilterType;
  limit: number;
  generateReport = false;

  constructor(public activeModal: NgbActiveModal,
              private utmNetScanService: UtmNetScanService,
              private toastService: UtmToastService) {
  }

  ngOnInit() {
  }

  export() {
    this.assetFilters.size = this.limit;
    this.assetFilters.page = 0;
    this.generateReport = true;
    this.utmNetScanService.exportAssetsToPdf(this.assetFilters).subscribe(data => {
      this.viewPdf(data.body);
    });
  }

  viewPdf(dat) {
    this.activeModal.close();
    const data = new Blob([dat], {type: 'application/pdf'});
    this.assetFilters.size = ITEMS_PER_PAGE;
    this.generateReport = false;
    if (data.size > 0) {
      const reader = new FileReader();
      reader.readAsDataURL(data);
      reader.onloadend = () => {
        const base64data = reader.result.toString();
        const linkElement = document.createElement('a');
        linkElement.setAttribute('href', base64data);
        linkElement.setAttribute('download', 'UTMStack Assets report');
        linkElement.click();
      };
      // const fileURL = window.URL.createObjectURL(data);
      // window.open(fileURL);
    } else {
      this.toastService.showWarning('NO DATA FOUND', 'No data found for this report');
    }
  }
}
