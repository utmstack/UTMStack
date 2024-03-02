import {Component, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {NgbModal, NgbPopover} from '@ng-bootstrap/ng-bootstrap';
import {NgSelectComponent} from '@ng-select/ng-select';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {AssetGroupCreateComponent} from '../../../asset-groups/asset-group-create/asset-group-create.component';
import {AssetGroupType} from '../../../asset-groups/shared/type/asset-group.type';
import {AssetReloadFilterBehavior} from '../../behavior/asset-reload-filter-behavior.service';
import {AssetFieldFilterEnum} from '../../enums/asset-field-filter.enum';
import {UtmAssetGroupService} from '../../services/utm-asset-group.service';
import {UtmNetScanService} from '../../services/utm-net-scan.service';

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
  @Output() focus = new EventEmitter<boolean>();

  @ViewChild('tagPopoverSpan') tagPopoverSpan: NgbPopover;
  @ViewChild('tagPopoverButton') tagPopoverButton: NgbPopover;
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
        this.closePopover();
      }, error => {
        this.utmToastService.showError('Error changing group',
          'Error changing group, please try again');
        this.creating = false;
        this.closePopover();
      });
  }

  getIcon(group: AssetGroupType) {
    return group ? 'icon-grid-alt' : 'icon-grid5';
  }

  addNewGroup() {
    this.closePopover();
    const modalGroup = this.modalService.open(AssetGroupCreateComponent, {centered: true});
  }

  handleClear(select: NgSelectComponent) {
    this.group = null;
    select.close();
  }

  closePopover() {
    if (this.tagPopoverSpan) {
      this.tagPopoverSpan.close();
    }

    if (this.tagPopoverButton) {
      this.tagPopoverButton.close();
    }
  }

  onHidden() {
    this.focus.emit(false);
  }

  onShown() {
    this.focus.emit(true);
  }
}
