import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {AssetGroupCreateComponent} from '../../../asset-groups/asset-group-create/asset-group-create.component';
import {AssetGroupType} from '../../../asset-groups/shared/type/asset-group.type';
import {UtmAssetGroupService} from '../../services/utm-asset-group.service';
import {UtmAssetTypeService} from '../../services/utm-asset-type.service';
import {UtmNetScanService} from '../../services/utm-net-scan.service';
import {AssetType} from '../../types/asset-type';
import {NetScanType} from '../../types/net-scan.type';

@Component({
  selector: 'app-asset-create',
  templateUrl: './asset-create.component.html',
  styleUrls: ['./asset-create.component.css']
})
export class AssetCreateComponent implements OnInit {
  formAsset: FormGroup;
  @Input() asset: NetScanType;
  @Output() assetCreated = new EventEmitter<string>();
  step = 1;
  stepCompleted: number[] = [];
  creating = false;
  types: AssetType[];
  portValid: boolean;
  groups: AssetGroupType[];

  constructor(private fb: FormBuilder,
              public activeModal: NgbActiveModal,
              private assetTypeService: UtmAssetTypeService,
              private utmNetScanService: UtmNetScanService,
              private utmToastService: UtmToastService,
              private utmAssetGroupService: UtmAssetGroupService,
              private modalService: NgbModal,
              public inputClassResolve: InputClassResolve) {
  }

  ngOnInit() {
    this.getTypes();
    this.getAssetsGroups();
    this.initForm();
    if (this.asset) {
      this.utmNetScanService.findAssetByID(this.asset.id).subscribe(response => {
        this.asset = response.body;
        this.formAsset.patchValue(this.asset);
      });
    }
  }

  setFormValue() {
    this.formAsset.get('id').setValue(this.asset.id);
    this.formAsset.get('assetIp').setValue(this.asset.assetIp);
    this.formAsset.get('assetName').setValue(this.asset.assetName);
    this.formAsset.get('assetMac').setValue(this.asset.assetMac);
    this.formAsset.get('assetOs').setValue(this.asset.assetOs);
    this.formAsset.get('assetAddresses').setValue(this.asset.assetAddresses);
    this.formAsset.get('assetAlias').setValue(this.asset.assetAlias);
    this.formAsset.get('assetAlive').setValue(this.asset.assetAlive);
    this.formAsset.get('assetType').setValue(this.asset.assetType);
    this.formAsset.get('assetNotes').setValue(this.asset.assetNotes);
    this.formAsset.get('group').setValue(this.asset.group);
    this.formAsset.get('ports').setValue(this.asset.ports);
  }

  getTypes() {
    this.assetTypeService.query({size: 100}).subscribe(reponse => {
      this.types = reponse.body;
    });
  }

  getAssetsGroups() {
    this.utmAssetGroupService.query({page: 0, size: 10000}).subscribe(response => {
      this.groups = response.body;
    });
  }

  initForm() {
    this.formAsset = this.fb.group({
      id: [],
      assetIp: ['', [Validators.required,
        Validators.pattern('^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$')]],
      assetName: ['', [Validators.required]],
      assetMac: [''],
      assetOs: [''],
      assetAddresses: [],
      assetAlias: [],
      assetAlive: [true],
      assetType: [],
      assetNotes: [],
      group: [],
      softwares: [],
      ports: [],
      registeredMode: [],
      discoveredAt: [],
      assetSeverity: [],
      assetStatus: []
    });
  }


  nextStep() {
    this.stepCompleted.push(this.step);
    this.step += 1;
  }

  isCompleted(step: number) {
    return this.stepCompleted.findIndex(value => value === step) !== -1;
  }

  backStep() {
    this.stepCompleted.pop();
    this.step -= 1;
  }


  create() {
    this.creating = true;
    this.formAsset.get('id').setValue(this.asset ? this.asset.id : null);
    this.utmNetScanService.create(this.formAsset.value)
      .subscribe(response => {
        this.utmToastService.showSuccessBottom('Asset created successfully');
        this.assetCreated.emit('success');
        this.activeModal.close();
        this.creating = false;
      }, error => {
        this.utmToastService.showError('Error creating asset',
          'Error creating asset, please try again');
        this.creating = false;
      });
  }

  newGroup() {
    const modalGroup = this.modalService.open(AssetGroupCreateComponent, {centered: true});
    modalGroup.componentInstance.addGroup.subscribe(group => {
      this.formAsset.get('group').setValue(group);
    });
  }
}
