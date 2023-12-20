import {Component, Input, OnInit} from '@angular/core';
import {VsSeverityResolverService} from '../../../../../shared/services/scan/vs-severity-resolver.service';
import {AssetActiveFilterType} from '../../../types/asset-active-filter.type';

@Component({
  selector: 'app-asset-report-active-filter',
  templateUrl: './asset-report-active-filter.component.html',
  styleUrls: ['./asset-report-active-filter.component.scss']
})
export class AssetReportActiveFilterComponent implements OnInit {
  @Input() filterActive: AssetActiveFilterType;

  constructor(public severityResolverService: VsSeverityResolverService) {
  }

  ngOnInit() {
  }

  resolveSeverity(): string {
    if (this.filterActive['hostSeverity.equals'] !== undefined) {
      if (this.filterActive['hostSeverity.equals'] !== '""') {
        return this.filterActive['hostSeverity.equals'] + '';
      } else {
        return -99 + '';
      }
    } else if (this.filterActive['hostSeverity.greaterThan'] !== undefined) {
      return (this.filterActive['hostSeverity.greaterThan'] + 1) + '';
    }
  }
}
