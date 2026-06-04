import {Component, Input, OnInit} from '@angular/core';
import {FormArray, FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../../../../shared/alert/utm-toast.service';
import {Asset} from '../../../../../../models/rule.model';
import {AssetManagerService} from '../../../../../services/asset-manager.service';

@Component({
  selector: 'app-add-asset',
  templateUrl: './add-asset.component.html',
  styleUrls: ['./add-asset.component.scss']
})
export class AddAssetComponent implements OnInit {

  @Input() asset: Asset;
  assetForm: FormGroup;
  loading = false;
  mode: 'ADD' | 'EDIT';

  constructor(public activeModal: NgbActiveModal,
              private assetManagerService: AssetManagerService,
              private formBuilder: FormBuilder,
              private utmToastService: UtmToastService) {
  }

  ngOnInit() {
    this.mode = this.asset ? 'EDIT' : 'ADD';
    this.initForm();
  }

  initForm() {
    this.assetForm = this.formBuilder.group({
      assetName: [this.asset ? this.asset.assetName : '', Validators.required],
      assetHostnameList: this.initFormArray(this.asset ? this.asset.assetHostnameList : []),
      assetIpList: this.initFormArray(this.asset ? this.asset.assetIpList : []),
      assetConfidentiality: [this.asset ? this.asset.assetConfidentiality : '', [Validators.required, Validators.min(0), Validators.max(3)]],
      assetIntegrity: [this.asset ? this.asset.assetIntegrity : '', [Validators.required, Validators.min(0), Validators.max(3)]],
      assetAvailability: [this.asset ? this.asset.assetAvailability : '', [Validators.required, Validators.min(0), Validators.max(3)]]
    });
  }

  initFormArray(values: string[]): FormArray {
    const formArray = values.map(value => this.formBuilder.control(value, Validators.required));
    return this.formBuilder.array(formArray.length > 0 ? formArray
      : [this.formBuilder.control('', Validators.required)], this.minLengthArray(1));
  }

  get assetName() {
    return this.assetForm.get('assetName');
  }

  get assetConfidentiality() {
    return this.assetForm.get('assetConfidentiality');
  }

  get assetIntegrity() {
    return this.assetForm.get('assetIntegrity');
  }

  get assetAvailability() {
    return this.assetForm.get('assetAvailability');
  }

  get assetHostnameList() {
    return this.assetForm.get('assetHostnameList') as FormArray;
  }

  get assetIpList() {
    return this.assetForm.get('assetIpList') as FormArray;
  }

  addHostname() {
    this.assetHostnameList.push(this.formBuilder.control(''));
  }

  removeHostname(index: number) {
    this.assetHostnameList.removeAt(index);
  }

  addIP() {
    this.assetIpList.push(this.formBuilder.control(''));
  }

  removeIP(index: number) {
    this.assetIpList.removeAt(index);
  }

  minLengthArray(min: number) {
    return (control: FormArray): { [key: string]: boolean } | null => {
      if (control.length >= min) {
        return null;
      }
      return { minLengthArray: true };
    };
  }

  onSubmit() {
    if (this.assetForm.valid) {
      this.loading = true;
      const formPattern: Asset = this.assetForm.value;
      const patternToSave = this.mode === 'ADD' ? formPattern : {
        ...this.asset,
        ...formPattern
      };
      this.assetManagerService
        .saveAsset(this.mode, patternToSave)
           .subscribe({
             next: response => {
               this.assetForm.reset();
               this.loading = false;
               this.utmToastService.showSuccessBottom(this.mode === 'ADD'
                       ? 'Asset saved successfully' : 'Asset edited successfully');
               this.activeModal.close(true);
             },
             error: err => {
               this.loading = false;
               this.utmToastService.showError('Error', this.mode === 'ADD'
                   ? 'Error saving asset' : 'Error editing asset');
             }
           });
    }
  }

}
