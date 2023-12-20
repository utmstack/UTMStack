import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../../../shared/alert/utm-toast.service';
import {CredentialModel} from '../../../shared/model/credential.model';
import {CredentialService} from '../shared/services/credential.service';

@Component({
  selector: 'app-credential-delete',
  templateUrl: './credential-delete.component.html',
  styleUrls: ['./credential-delete.component.scss']
})
export class CredentialDeleteComponent implements OnInit {
  @Input() credential: CredentialModel;
  @Output() credentialDeleted = new EventEmitter<string>();

  constructor(
    public activeModal: NgbActiveModal,
    private credentialService: CredentialService,
    private utmToastService: UtmToastService) {
  }

  ngOnInit() {
  }

  deleteCredential() {
    this.credentialService.delete(this.credential.uuid)
      .subscribe(() => {
        this.utmToastService.showSuccessBottom('Credential deleted successfully');
        this.activeModal.close();
        this.credentialDeleted.emit('deleted');
      }, () => {
        this.utmToastService.showError('Error deleting credential',
          'Error deleting credential, please check your network and try again');
      });
  }
}
