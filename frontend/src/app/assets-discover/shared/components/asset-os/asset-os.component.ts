import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-asset-os',
  templateUrl: './asset-os.component.html',
  styleUrls: ['./asset-os.component.scss']
})
export class AssetOsComponent implements OnInit {
  @Input() assetOS: string;
  icon: string;
  color: string;

  constructor() {
  }

  ngOnInit() {
    if (this.assetOS && this.assetOS.toLowerCase().includes('windows')) {
      this.icon = 'icon-windows8';
      this.color = '';
    } else if (this.assetOS && this.assetOS.toLowerCase().includes('linux')) {
      this.icon = 'icon-tux';
      this.color = '';
    } else {
      this.icon = 'icon-question3';
      this.color = '';
    }
  }
}
