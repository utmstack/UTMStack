import {Component, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-modal-confirmation',
  templateUrl: './modal-confirmation.component.html',
  styleUrls: ['./modal-confirmation.component.scss']
})
export class ModalConfirmationComponent implements OnInit {
  header: string;
  message: string;
  confirmBtnText: string;
  confirmBtnIcon: string;
  confirmBtnType: 'delete' | 'default';
  textDisplay: string;
  textType: 'warning' | 'danger';
  hideBtnCancel = false;

  constructor(private activeModal: NgbActiveModal) {
  }

  ngOnInit() {
  }

  confirm() {
    this.activeModal.close('ok');
  }

  cancel() {
    this.activeModal.dismiss('cancel');
  }
}
