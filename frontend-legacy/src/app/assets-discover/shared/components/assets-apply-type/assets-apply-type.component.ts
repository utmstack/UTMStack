import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {AssetReloadFilterBehavior} from '../../behavior/asset-reload-filter-behavior.service';
import {AssetFieldFilterEnum} from '../../enums/asset-field-filter.enum';
import {AssetTypeIconEnum} from '../../enums/asset-type-icon.enum';
import {UtmAssetTypeService} from '../../services/utm-asset-type.service';
import {UtmNetScanService} from '../../services/utm-net-scan.service';
import {AssetType} from '../../types/asset-type';

@Component({
  selector: 'app-assets-apply-type',
  templateUrl: './assets-apply-type.component.html',
  styleUrls: ['./assets-apply-type.component.scss']
})
export class AssetsApplyTypeComponent implements OnInit {
  @Input() typeFormat: 'icon' | 'button';
  @Input() showTypeLabel: boolean;
  @Input() assets: number[];
  @Input() type: AssetType;
  @Output() applyTypeEvent = new EventEmitter<string>();
  @Output() focus = new EventEmitter<boolean>();

  types: AssetType[] = [];
  loading = true;
  creating = false;

  constructor(
    private modalService: NgbModal,
    private assetTypeService: UtmAssetTypeService,
    private utmNetScanService: UtmNetScanService,
    private utmToastService: UtmToastService,
    private assetTypeChangeBehavior: AssetReloadFilterBehavior
  ) {
  }

  ngOnInit() {
    if (this.typeFormat === 'button') {
      this.getTypes();
    }
  }

  getTypes(event: Event = null) {
    if (event) {
      event.stopPropagation();
    }
    this.assetTypeService.query({size: 100}).subscribe(reponse => {
      this.types = reponse.body;
      this.loading = false;
    });
  }


  applyType() {
    this.creating = true;
    const id = this.type ? this.type.id : null;
    this.utmNetScanService.updateType({assetsIds: this.assets, assetTypeId: id})
      .subscribe(response => {
        this.utmToastService.showSuccessBottom('Type changed successfully');
        this.applyTypeEvent.emit('success');
        this.assetTypeChangeBehavior.$assetReloadFilter.next(AssetFieldFilterEnum.TYPE);
        this.creating = false;
      }, error => {
        this.utmToastService.showError('Error changing type',
          'Error changing type, please try again');
        this.creating = false;
      });
  }

  getIcon(type: AssetType) {
    if (type) {
      const typeEnum = type.typeName.replace(' ', '_').replace('-', '_');
      return AssetTypeIconEnum[typeEnum] ? AssetTypeIconEnum[typeEnum] : 'icon-display';
    } else {
      return 'icon-display';
    }
  }

  onHidden() {
    this.focus.emit(false);
  }

  onShown() {
    this.focus.emit(true);
  }
}
