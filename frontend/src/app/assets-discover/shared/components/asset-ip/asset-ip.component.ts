import {Component, Input, OnInit} from '@angular/core';
import {NetScanType} from '../../types/net-scan.type';

@Component({
  selector: 'app-asset-ip',
  templateUrl: './asset-ip.component.html',
  styleUrls: ['./asset-ip.component.scss']
})
export class AssetIpComponent implements OnInit {
  @Input() asset: NetScanType;

  constructor() {
  }

  ngOnInit() {
  }

}
