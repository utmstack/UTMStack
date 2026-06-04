import {Component, EventEmitter, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {AssetSoftwareService} from '../../services/asset-software.service';
import {NetScanSoftwares} from '../../types/net-scan-softwares';

@Component({
  selector: 'app-asset-software-add',
  templateUrl: './asset-software-add.component.html',
  styleUrls: ['./asset-software-add.component.css']
})
export class AssetSoftwareAddComponent implements OnInit {
  @Output() softwareCreated = new EventEmitter<NetScanSoftwares>();
  formSoftware: FormGroup;
  creating: boolean;

  constructor(public activeModal: NgbActiveModal,
              public inputClassResolve: InputClassResolve,
              private modalService: NgbModal,
              private fb: FormBuilder,
              private utmToastService: UtmToastService,
              private assetSoftwareService: AssetSoftwareService) {
  }

  ngOnInit() {
    this.formSoftware = this.fb.group({
      softDeveloper: ['', Validators.required],
      softName: ['', Validators.required],
      softVersion: ['', Validators.required],
    });
  }

  addSoftware() {
    this.assetSoftwareService.create(this.formSoftware.value).subscribe(response => {
      this.softwareCreated.emit(response.body);
      this.activeModal.close();
      this.utmToastService.showSuccessBottom('Software created successfully');
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating Software ',
        'Error creating Software, check your network');
    });

  }
}
