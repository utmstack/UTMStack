import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {FormBuilder, FormGroup, Validators} from '@angular/forms';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {InputClassResolve} from '../../../shared/util/input-class-resolve';
import {FederationConnectionService} from '../shared/services/federation-connection.service';

@Component({
  selector: 'app-token-activate',
  templateUrl: './token-activate.component.html',
  styleUrls: ['./token-activate.component.css']
})
export class TokenActivateComponent implements OnInit {
  @Input() enable = true;
  @Output() connectionActivated = new EventEmitter<boolean>();
  tokenForm: FormGroup;
  checked = false;
  activated = false;
  validating = false;
  saving: boolean;
  enabled: boolean;

  constructor(public inputClassResolve: InputClassResolve,
              public activeModal: NgbActiveModal,
              private federationConnectionService: FederationConnectionService,
              private toastService: UtmToastService,
              private fb: FormBuilder) {
  }

  ngOnInit() {
    this.tokenForm = this.fb.group({
      activate: [true],
      name: ['', [Validators.required]],
      domain: [window.location.host]
    });
  }

  enableConnection() {
    this.saving = true;
    this.federationConnectionService.activate(this.tokenForm.value).subscribe(() => {
      this.saving = false;
      this.connectionActivated.emit(true);
      this.activeModal.close();
    }, () => {
      this.toastService.showError('Error enabling key', 'Error while trying to enable connection key');
    });
  }

}
