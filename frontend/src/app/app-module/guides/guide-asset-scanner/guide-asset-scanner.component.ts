import {Component, Input, OnInit} from '@angular/core';
import {UtmModulesEnum} from '../../shared/enum/utm-module.enum';

@Component({
  selector: 'app-guide-asset-scanner',
  templateUrl: './guide-asset-scanner.component.html',
  styleUrls: ['./guide-asset-scanner.component.css']
})
export class GuideAssetScannerComponent implements OnInit {
  @Input() integrationId: number;
  module = UtmModulesEnum;
  @Input() serverId: number;
  configValidity: boolean;


  constructor() { }

  ngOnInit() {
  }

  configValidChange($event: boolean) {
    this.configValidity = !$event;
  }

}
