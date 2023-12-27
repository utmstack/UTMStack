import {Component, Input, OnInit} from '@angular/core';
import {AssetGroupType} from '../../type/asset-group.type';

@Component({
  selector: 'app-asset-group-detail',
  templateUrl: './asset-group-detail.component.html',
  styleUrls: ['./asset-group-detail.component.scss']
})
export class AssetGroupDetailComponent implements OnInit {
  @Input() group: AssetGroupType;

  constructor() {
  }

  ngOnInit() {
  }

}
