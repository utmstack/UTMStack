import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {AssetFieldFilterEnum} from '../../../enums/asset-field-filter.enum';
import {UtmNetScanService} from '../../../services/utm-net-scan.service';

@Component({
  selector: 'app-asset-filter-is-alive',
  templateUrl: './asset-filter-is-alive.component.html',
  styleUrls: ['./asset-filter-is-alive.component.scss']
})
export class AssetFilterIsAliveComponent implements OnInit {
  @Output() filterAliveChange = new EventEmitter<{ prop: AssetFieldFilterEnum, values: boolean }>();
  alive = null;
  alives: Array<[string, number]> = [];
  assetFieldFilterEnum = AssetFieldFilterEnum;

  constructor(private utmNetScanService: UtmNetScanService) {
  }

  ngOnInit() {
    this.getAliveStatus();
  }

  getAliveStatus() {
    this.utmNetScanService.getFieldValues({page: 0, size: 5, forGroups: false, prop: this.assetFieldFilterEnum.ALIVE})
      .subscribe(response => {
        this.alives = response.body;
      });
  }

  getTotal() {
    return this.alives.reduce((a, b) => a + (b[1] || 0), 0);
  }

  setAlive(param) {
    this.alive = param;
    this.filterAliveChange.emit({prop: AssetFieldFilterEnum.ALIVE, values: this.alive});
  }
}
