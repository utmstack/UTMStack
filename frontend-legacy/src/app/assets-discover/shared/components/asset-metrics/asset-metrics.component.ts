import {Component, Input, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {ALERT_SOURCE_HOSTNAME_FIELD} from '../../../../shared/constants/alert/alert-field.constant';
import {ChartValueSeparator} from '../../../../shared/enums/chart-value-separator';
import {ElasticOperatorsEnum} from '../../../../shared/enums/elastic-operators.enum';
import {NetScanType} from '../../types/net-scan.type';
import {AssetMetricEnum} from './shared/enums/asset-metric.enum';

@Component({
  selector: 'app-asset-metrics',
  templateUrl: './asset-metrics.component.html',
  styleUrls: ['./asset-metrics.component.scss']
})
export class AssetMetricsComponent implements OnInit {
  @Input() asset: NetScanType;
  @Input() detailed = false;
  logAmount = 0;

  constructor(private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    if (this.asset.metrics) {
      this.asset.metrics = {
        alerts: this.asset.metrics.hasOwnProperty(AssetMetricEnum.ALERT) ? this.asset.metrics.alerts : 0,
        event: this.asset.metrics.hasOwnProperty(AssetMetricEnum.EVENT) ? this.asset.metrics.event : 0,
        vulnerabilities: this.asset.metrics.hasOwnProperty(AssetMetricEnum.VULNERABILITIES) ? this.asset.metrics.vulnerabilities : 0,
      };
    }
  }

  classByMetric(metric: { key: AssetMetricEnum, value: any }) {
    switch (metric.key) {
      case AssetMetricEnum.ALERT:
        return 'text-danger-800';
      case AssetMetricEnum.EVENT:
        return 'text-info-800';
      case AssetMetricEnum.VULNERABILITIES:
        return 'text-warning-800';
      case AssetMetricEnum.FILEBEAT:
        return 'text-purple-800';
      case AssetMetricEnum.HIDS:
        return 'text-teal-800';
      case AssetMetricEnum.METRICBEAT:
        return 'text-orange-800';
      case AssetMetricEnum.WINLOGBEAT:
        return 'text-slate-800';
      default:
        return 'text-blue-800';
    }
  }

  navigateTo(key: AssetMetricEnum) {
    switch (key) {
      case AssetMetricEnum.ALERT:
        this.navigateToDataManagement(this.asset.assetName);
        break;
      case AssetMetricEnum.LOGS:
        break;
      case AssetMetricEnum.VULNERABILITIES:
        this.navigateToVulnerabilites(this.asset.assetName);
        break;
    }
  }

  navigateToDataManagement(ip: string) {
    const queryParams = {alertType: 'ALERT'};
    queryParams[ALERT_SOURCE_HOSTNAME_FIELD] = ElasticOperatorsEnum.IS + ChartValueSeparator.BUCKET_SEPARATOR + ip;
    this.navigateWithParams('/data/alert/view', queryParams);
  }

  navigateToVulnerabilites(ip: string) {
    const queryParams = {ip};
    this.navigateWithParams('/scanner/assets-discovery/assets-detail', queryParams);
  }


  navigateWithParams(route: string, queryParams: object) {
    this.spinner.show('loadingSpinner');
    this.router.navigate([route], {
      queryParams
    }).then(() => {
      this.spinner.hide('loadingSpinner');
    });
  }

  private getTotalLogs() {
    let total = 0;
    for (const key of Object.keys(this.asset.metrics)) {
      if (key !== AssetMetricEnum.ALERT && key !== AssetMetricEnum.EVENT && key !== AssetMetricEnum.VULNERABILITIES) {
        total += this.asset.metrics[key];
        return total;
      }
    }
  }
}
