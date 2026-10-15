import {Component} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-tfa-disable-modal',
  templateUrl: './tfa-disable-modal.component.html',
  styleUrls: ['./tfa-disable-modal.component.css']
})
export class TfaDisableModalComponent {
  saving = false;
  password = '';

  constructor(public activeModal: NgbActiveModal) {}

  reject(): void {
    this.activeModal.dismiss('Cancel');
  }

  onDisable(): void {
    if (!this.password) {
      return;
    }
    this.saving = true;
    this.activeModal.close(this.password);
  }
}
