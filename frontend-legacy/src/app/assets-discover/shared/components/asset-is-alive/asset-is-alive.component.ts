import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-asset-is-alive',
  templateUrl: './asset-is-alive.component.html',
  styleUrls: ['./asset-is-alive.component.scss']
})
export class AssetIsAliveComponent implements OnInit {
  @Input() assetAlive: boolean;
  color: string;
  icon: string;

  constructor() {
  }

  ngOnInit() {
    if (this.assetAlive) {
      this.icon = 'icon-circle2';
      this.color = 'text-success-400';
    } else {
      this.icon = 'icon-circle2';
      this.color = 'text-danger-400';
    }
  }

}
