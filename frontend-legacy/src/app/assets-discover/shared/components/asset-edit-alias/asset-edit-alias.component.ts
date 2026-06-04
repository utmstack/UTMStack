import {Component, Input, OnInit} from '@angular/core';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {AssetReloadFilterBehavior} from '../../behavior/asset-reload-filter-behavior.service';
import {AssetFieldFilterEnum} from '../../enums/asset-field-filter.enum';
import {UtmNetScanService} from '../../services/utm-net-scan.service';
import {NetScanType} from '../../types/net-scan.type';

@Component({
  selector: 'app-asset-edit-alias',
  templateUrl: './asset-edit-alias.component.html',
  styleUrls: ['./asset-edit-alias.component.css']
})
export class AssetEditAliasComponent implements OnInit {
  @Input() asset: NetScanType;
  creating: any;

  constructor(private utmToastService: UtmToastService,
              private assetReloadFilterBehavior: AssetReloadFilterBehavior,
              private utmNetScanService: UtmNetScanService) {
  }

  ngOnInit() {
  }

  editAlias() {
    this.creating = true;
    this.utmNetScanService.update(this.asset)
      .subscribe(response => {
        this.utmToastService.showSuccessBottom('Asset alias created successfully');
        this.assetReloadFilterBehavior.$assetReloadFilter.next(AssetFieldFilterEnum.ALIAS);
        this.creating = false;
      }, error => {
        this.utmToastService.showError('Error creating asset',
          'Error creating asset alias, please try again');
        this.creating = false;
      });
  }
}
