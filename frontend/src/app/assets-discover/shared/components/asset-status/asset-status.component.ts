import {Component, Input, OnChanges, OnInit, SimpleChanges} from '@angular/core';
import {UtmDataInputStatus} from '../../types/data-source-input.type';
import {NetScanType} from '../../types/net-scan.type';
import {AssetsStatusEnum} from "../../enums/assets-status.enum";

@Component({
  selector: 'app-asset-status',
  templateUrl: './asset-status.component.html',
  styleUrls: ['./asset-status.component.scss']
})
export class AssetStatusComponent implements OnInit, OnChanges {
  @Input() asset: NetScanType;
  statusClass: string;
  statusLabel: string;

  constructor() {
  }

  ngOnInit() {
    this.statusClass = this.getClass();
    this.statusLabel = this.getLabel();
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes.asset) {
      this.statusClass = this.getClass();
      this.statusLabel = this.getLabel();
    }
  }

  getDisconnected(data: UtmDataInputStatus[]): number {
    if (this.asset.agent && !this.asset.assetAlive) {
      return data.length;
    }
    if (data.length > 0) {
      return data.filter(value => value.down).length;
    } else {
      return -1;
    }
  }

  getClass(): string {
    if (!this.asset.assetAlive) {
      return 'text-danger';
    } else if (this.asset.assetAlive && this.asset.agent) {
      return 'text-success';
    } else {
      return this.getDisconnected(this.asset.dataInputList) === 0 ? 'text-success'
        : this.getDisconnected(this.asset.dataInputList) > 0 ? 'text-warning-800' : 'text-muted';
    }
  }

  getLabel(): string {
    if (!this.asset.assetAlive) {
      return AssetsStatusEnum.DISCONNECTED;
    } else if (this.asset.assetAlive && this.asset.agent) {
      return AssetsStatusEnum.CONNECTED;
    } else {
      return this.getDisconnected(this.asset.dataInputList) === 0 ? 'All connected'
        : this.getDisconnected(this.asset.dataInputList) > 0 ? (this.getDisconnected(this.asset.dataInputList) + ' delayed') : 'Unknown';
    }
  }

}
