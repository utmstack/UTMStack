import {Component, Input, OnInit} from '@angular/core';
import {AssetModel} from '../../model/assets/asset.model';

@Component({
  selector: 'app-asset-card-detail',
  templateUrl: './asset-card-detail.component.html',
  styleUrls: ['./asset-card-detail.component.scss']
})
export class AssetCardDetailComponent implements OnInit {
  @Input() asset: AssetModel;

  constructor() {
  }

  ngOnInit() {
  }

  resolveAssetIP(): string {
    if (this.asset.host.detail !== null) {
      const index = this.asset.identifiers.findIndex(value => value.name === 'ip');
      if (index !== -1) {
        return this.asset.identifiers[index].value;
      }
    }
    return '-';
  }

  public resolveAssetSoCpe(): string {
    if (this.asset.host.detail !== null) {
      const index = this.asset.host.detail.findIndex(value => value.name === 'best_os_cpe');
      if (index !== -1) {
        return this.asset.host.detail[index].value;
      }
    }
    return '-';
  }
}
