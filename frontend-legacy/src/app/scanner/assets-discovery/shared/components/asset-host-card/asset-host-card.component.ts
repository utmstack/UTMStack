import {Component, Input, OnInit} from '@angular/core';
import {AssetModel} from '../../../../shared/model/assets/asset.model';


@Component({
  selector: 'app-asset-host-card',
  templateUrl: './asset-host-card.component.html',
  styleUrls: ['./asset-host-card.component.scss']
})
export class AssetHostCardComponent implements OnInit {
  @Input() asset: AssetModel;
  checkbox = false;
  viewAction = false;
  pieOption: any;

  constructor() {
  }

  ngOnInit() {

  }

  toggleCheck() {
    this.checkbox = this.checkbox ? false : true;
  }

  toggleViewAction() {
    this.viewAction = this.viewAction ? false : true;
  }


  resolveSeverityFixed(): string {
    return this.asset.host.severity.value.replace('.0', '');
  }
}
