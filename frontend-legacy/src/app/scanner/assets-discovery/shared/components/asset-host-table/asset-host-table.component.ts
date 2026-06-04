import {Component, Input, OnInit} from '@angular/core';
import {Router} from '@angular/router';
import {AssetModel} from '../../../../shared/model/assets/asset.model';

@Component({
  selector: 'app-asset-host-table',
  templateUrl: './asset-host-table.component.html',
  styleUrls: ['./asset-host-table.component.scss']
})
export class AssetHostTableComponent implements OnInit {
  @Input() assets: AssetModel[];
  @Input() loading: boolean;
  assetSelected: AssetModel[] = [];

  constructor(private router: Router) {
  }

  ngOnInit() {
  }

  addToSelected(asset: AssetModel) {
    const index = this.assetSelected.findIndex(value => value.uuid === asset.uuid);
    if (index === -1) {
      this.assetSelected.push(asset);
    } else {
      this.assetSelected.splice(index, 1);
    }
  }

  isSelected(asset: AssetModel) {
    return this.assetSelected.findIndex(value => value.uuid === asset.uuid) > -1;
  }

  viewDetail(asset: AssetModel) {
    this.router.navigate(['/scanner/assets-discovery/assets-detail'],
      {queryParams: {reportId: asset.uuid}});
    // {queryParams: {reportId: '67684681-f08b-43a0-bfed-26ad26c02de7'}});
  }

}
