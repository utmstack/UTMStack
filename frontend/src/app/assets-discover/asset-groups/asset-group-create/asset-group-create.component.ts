import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {UtmAssetGroupService} from '../../shared/services/utm-asset-group.service';
import {AssetGroupType} from '../shared/type/asset-group.type';

@Component({
  selector: 'app-asset-group-create',
  templateUrl: './asset-group-create.component.html',
  styleUrls: ['./asset-group-create.component.scss']
})
export class AssetGroupCreateComponent implements OnInit {
  @Output() addGroup: EventEmitter<AssetGroupType> = new EventEmitter();
  @Input() group: AssetGroupType = {groupDescription: '', groupName: ''};
  edit: boolean;
  typing: boolean;
  exist = false;
  creating = false;
  private timer: any;

  constructor(public activeModal: NgbActiveModal,
              private utmToastService: UtmToastService,
              private utmAssetGroupService: UtmAssetGroupService) {
  }

  ngOnInit() {
    this.edit = this.group.id !== undefined;
  }

  createGroup() {
    this.utmAssetGroupService.create(this.group).subscribe(response => {
      this.addGroup.emit(response.body);
      this.activeModal.close();
      this.utmToastService.showSuccessBottom('Group created successfully');
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating group ',
        'Error creating configuration, check your network');
    });
  }

  editGroup() {
    this.utmAssetGroupService.update(this.group).subscribe(response => {
      this.addGroup.emit(response.body);
      this.utmToastService.showSuccessBottom('Group updated successfully');
      this.activeModal.close();
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error updating group ',
        'Error creating configuration, check your network');
    });
  }

  checkName() {
    this.typing = true;
    this.exist = true;
    if (this.timer) {
      clearTimeout(this.timer);
    }
    this.timer = setTimeout(() => {
      this.searchGroup();
    }, 1000);
  }

  searchGroup() {
    const req = {
      'tagName.equals': this.group.groupName
    };
    this.utmAssetGroupService.query(req).subscribe(response => {
      this.exist = response.body.length > 0 && !this.group;
      this.typing = false;
    });
  }
}
