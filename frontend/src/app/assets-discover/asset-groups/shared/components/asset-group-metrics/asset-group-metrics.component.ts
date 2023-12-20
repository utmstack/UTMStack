import {Component, Input, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {NgxSpinnerService} from 'ngx-spinner';
import {AssetMetricEnum} from '../../../../shared/components/asset-metrics/shared/enums/asset-metric.enum';
import {AssetGroupType} from '../../type/asset-group.type';

@Component({
  selector: 'app-asset-group-metrics',
  templateUrl: './asset-group-metrics.component.html',
  styleUrls: ['./asset-group-metrics.component.scss']
})
export class AssetGroupMetricsComponent implements OnInit {
  @Input() group: AssetGroupType;
  @Input() detailed = false;
  logAmount = 0;

  constructor(private router: Router,
              private spinner: NgxSpinnerService) {
  }

  ngOnInit() {
    if (this.group.metrics) {
      this.group.metrics = {
        alerts: this.group.metrics.hasOwnProperty(AssetMetricEnum.ALERT) ? this.group.metrics.alerts : 0,
        event: this.group.metrics.hasOwnProperty(AssetMetricEnum.EVENT) ? this.group.metrics.event : 0,
        vulnerabilities: this.group.metrics.hasOwnProperty(AssetMetricEnum.VULNERABILITIES) ? this.group.metrics.vulnerabilities : 0,
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


}
