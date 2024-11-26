import {Component, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-modal-add-note',
  templateUrl: './modal-add-note.component.html',
  styleUrls: ['./modal-add-note.component.scss']
})
export class ModalAddNoteComponent implements OnInit {
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
