import {Component, Input, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {TaskResultParamType} from '../../../../vulnerability-scanner/shared/enums/task-result-param.type';
import {AssetSeverityEnum} from '../../enums/asset-severity.enum';
import {NetScanType} from '../../types/net-scan.type';

@Component({
  selector: 'app-asset-severity',
  templateUrl: './asset-severity.component.html',
  styleUrls: ['./asset-severity.component.scss']
})
export class AssetSeverityComponent implements OnInit {
  @Input() asset: NetScanType;
  label: string;
  background: string;

  constructor(private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    this.setAssetSeverity();
  }

  navigateToVulnerabilites() {
    this.spinner.show('loadingSpinner');
    const queryParams = {};
    queryParams[TaskResultParamType.TYPE_DATA] = TaskResultParamType.TYPE_ASSET;
    queryParams[TaskResultParamType.HOST_IP] = this.asset.assetIp;
    this.router.navigate(['/vulnerability-scanner/task-result'], {
      queryParams
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

  private setAssetSeverity() {
    switch (this.asset.assetSeverity) {
      case AssetSeverityEnum.HIGH:
        this.background = 'border-danger-800 text-danger-800';
        this.label = 'HIGH';
        break;
      case AssetSeverityEnum.MEDIUM:
        this.background = 'border-warning-800 text-warning-800';
        this.label = 'MEDIUM';
        break;
      case AssetSeverityEnum.LOW:
        this.background = 'border-info-800 text-info-800';
        this.label = 'LOW';
        break;
      case AssetSeverityEnum.LOG:
        this.background = 'border-gray-800 text-gray-800';
        this.label = 'LOG';
        break;
      case AssetSeverityEnum.UNKNOWN:
        this.background = 'border-slate-800 text-slate-800';
        this.label = 'UNKNOWN';
        break;
      default:
        this.background = 'border-slate-800 text-black-500';
        this.label = '-';
    }
  }
}
