import {Component, OnInit} from '@angular/core';
import {NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {UtmToastService} from '../../shared/alert/utm-toast.service';
import {ModalConfirmationComponent} from '../../shared/components/utm/util/modal-confirmation/modal-confirmation.component';
import {FederationConnectionService} from './shared/services/federation-connection.service';

@Component({
  selector: 'app-connection-key',
  templateUrl: './connection-key.component.html',
  styleUrls: ['./connection-key.component.scss']
})
export class ConnectionKeyComponent implements OnInit {
  saving = false;
  enabled: boolean;
  token: string;

  constructor(private federationConnectionService: FederationConnectionService,
              private toastService: UtmToastService,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    this.getToken();
  }


  getToken() {
    this.federationConnectionService.getToken().subscribe(response => {
      this.enabled = response.body !== null && response.body !== '';
      if (this.enabled) {
        this.token = response.body;
      } else {
        this.token = null;
      }
    });
  }

  generateNewTokenModal() {
    const deleteModal = this.modalService.open(ModalConfirmationComponent, {centered: true});
    deleteModal.componentInstance.header = 'Generate new connection key';
    deleteModal.componentInstance.message = 'Some services an automations might be attached to this connection key,' +
      ' are you sure you want regenerate this token?';
    deleteModal.componentInstance.confirmBtnText = 'Generate';
    deleteModal.componentInstance.confirmBtnType = 'default';
    deleteModal.result.then((result) => {
      if (result === 'ok') {
        this.createNewToken();
      }
    });

  }

  createNewToken() {
    this.federationConnectionService.regenerateToken().subscribe(response => {
      this.token = response.body;
      this.toastService.showSuccessBottom('The connection key has been generated successfully');
    }, () => this.toastService.showError('Error generating new key', 'Error while trying to enable connection key'));
  }
}
