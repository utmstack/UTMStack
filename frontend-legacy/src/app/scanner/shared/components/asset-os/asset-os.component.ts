import {Component, Input, OnInit} from '@angular/core';
import {AssetModel} from '../../model/assets/asset.model';
import {AssetSoResolverService} from '../../providers/asset-so-resolver.service';

@Component({
  selector: 'app-asset-os',
  templateUrl: './asset-os.component.html',
  styleUrls: ['./asset-os.component.scss']
})
export class AssetOsComponent implements OnInit {
  @Input() asset: AssetModel;
  @Input() position: 'left' | 'right' | 'bottom' | 'top';

  constructor(public assetSoResolverService: AssetSoResolverService) {
  }

  ngOnInit() {
    if (!this.position) {
      this.position = 'left';
    }
  }

}
