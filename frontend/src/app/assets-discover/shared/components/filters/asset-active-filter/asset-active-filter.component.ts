import {Component, Input, OnInit} from '@angular/core';
import {AssetFilterType} from '../../../types/asset-filter.type';

@Component({
    standalone: false,
    selector: 'app-asset-active-filter',
    templateUrl: './asset-active-filter.component.html',
    styleUrls: ['./asset-active-filter.component.scss'],
})
export class AssetActiveFilterComponent implements OnInit {
  @Input() assetFilters: AssetFilterType;

  constructor() {
  }

  ngOnInit() {
  }

}
