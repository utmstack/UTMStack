import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../../shared/util/input-class-resolve';
import {CredentialModel} from '../../../shared/model/credential.model';
import {CredentialService} from '../shared/services/credential.service';

@Component({
  selector: 'app-credential-create',
  templateUrl: './credential-create.component.html',
  styleUrls: ['./credential-create.component.scss']
})
export class CredentialCreateComponent implements OnInit {
  @Input() credential: CredentialModel;
  @Input() type: string;
  @Output() credentialCreated = new EventEmitter<string>();
  edit = false;
  override = false;
  types = ['Username + Password'];
  formCredential: FormGroup;
  creating = false;

  constructor(
    public activeModal: NgbActiveModal,
    private credentialService: CredentialService,
    private utmToastService: UtmToastService,
    private fb: FormBuilder,
    public inputClass: InputClassResolve) {
  }

  ngOnInit() {
    this.initFormCredential();
    if (this.credential) {
      this.enablePasswordEdit();
      this.formCredential.get('login').setValue(this.credential.login);
      this.formCredential.get('allowInsecure').setValue(this.credential.allowInsecure);
      this.formCredential.get('comment').setValue(this.credential.comment);
      this.formCredential.get('name').setValue(this.credential.name);
    }
  }

  initFormCredential() {
    this.formCredential = this.fb.group({
      allowInsecure: [0],
      comment: [''],
      name: ['', Validators.required],
      password: [''],
      login: ['', Validators.required],
      type: 'up',
    });
  }

  createCredential() {
    this.creating = true;
    this.credentialService.create(this.formCredential.value).subscribe(credentialCreated => {
      this.utmToastService.showSuccessBottom('Credential created successfully');
      this.activeModal.close();
      this.credentialCreated.emit(credentialCreated.body.id);
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error creating credential',
        error1.error.statusText);
    });
  }


  editCredential() {
    this.creating = true;
    const cre = this.formCredential.value;
    cre.credentialId = this.credential.uuid;
    this.credentialService.update(cre).subscribe(credentialCreated => {
      this.utmToastService.showSuccessBottom('Credential updated successfully');
      this.activeModal.close();
      this.credentialCreated.emit('success');
    }, error1 => {
      this.creating = false;
      this.utmToastService.showError('Error updated credential',
        'Error updating credential, please check your network and try again');
    });
  }

  overridePassword() {
    this.override = this.override ? false : true;
    this.enablePasswordEdit();
  }

  enablePasswordEdit() {
    if (this.override) {
      this.formCredential.get('password').enable();
    } else {
      this.formCredential.get('password').disable();
    }
  }
}
