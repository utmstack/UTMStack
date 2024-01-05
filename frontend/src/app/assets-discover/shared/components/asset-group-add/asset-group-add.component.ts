import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {AssetGroupCreateComponent} from '../../../asset-groups/asset-group-create/asset-group-create.component';
import {AssetGroupType} from '../../../asset-groups/shared/type/asset-group.type';
import {AssetReloadFilterBehavior} from '../../behavior/asset-reload-filter-behavior.service';
import {AssetFieldFilterEnum} from '../../enums/asset-field-filter.enum';
import {UtmAssetGroupService} from '../../services/utm-asset-group.service';
import {UtmNetScanService} from '../../services/utm-net-scan.service';
import {NgSelectComponent} from "@ng-select/ng-select";

@Component({
  selector: 'app-asset-group-add',
  templateUrl: './asset-group-add.component.html',
  styleUrls: ['./asset-group-add.component.css']
})
export class AssetGroupAddComponent implements OnInit {
  @Input() typeFormat: 'icon' | 'button';
  @Input() showTypeLabel: boolean;
  @Input() assets: number[];
  @Input() group: AssetGroupType;
  @Output() applyGroupEvent = new EventEmitter<AssetGroupType>();
  groups: AssetGroupType[];
  loading = false;
  creating = false;

  constructor(
    private modalService: NgbModal,
    private utmNetScanService: UtmNetScanService,
    private utmToastService: UtmToastService,
    private utmAssetGroupService: UtmAssetGroupService,
    private assetReloadFilterBehavior: AssetReloadFilterBehavior
  ) {
  }

  ngOnInit() {
    if (this.typeFormat === 'button') {
      this.getGroups();
    }
  }

  getGroups(event: Event = null) {
    if (event) {
      event.stopPropagation();
    }
    this.utmAssetGroupService.query({page: 0, size: 1000}).subscribe(response => {
      this.groups = response.body;
      this.loading = false;
    });
  }


  applyGroup() {
    this.creating = true;
    const id = this.group ? this.group.id : null;
    this.utmNetScanService.updateGroup({assetsIds: this.assets, assetGroupId: id})
      .subscribe(response => {
        this.utmToastService.showSuccessBottom('Group changed successfully');
        this.applyGroupEvent.emit(response.body);
        this.assetReloadFilterBehavior.$assetReloadFilter.next(AssetFieldFilterEnum.GROUP);
        this.creating = false;
      }, error => {
        this.utmToastService.showError('Error changing group',
          'Error changing group, please try again');
        this.creating = false;
      });
  }

  getIcon(group: AssetGroupType) {
    return group ? 'icon-grid-alt' : 'icon-grid5';
  }

  addNewGroup() {
    const modalGroup = this.modalService.open(AssetGroupCreateComponent, {centered: true});
  }

  handleClear(select: NgSelectComponent) {
    this.group = null;
    select.close();
  }


}
