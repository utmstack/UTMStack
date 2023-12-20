import {Component, OnInit} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {NavBehavior} from '../../../shared/behaviors/nav.behavior';
import {Menu} from '../../../shared/types/menu/menu.model';

@Component({
  selector: 'app-menu-delete',
  templateUrl: './menu-delete-dialog.component.html',
  styleUrls: ['./menu-delete-dialog.component.scss']
})
export class MenuDeleteDialogComponent implements OnInit {
  menu: Menu;

  constructor(private activeModal: NgbActiveModal, private navBehavior: NavBehavior,) {
  }

  ngOnInit() {
  }

  deleteConfirmation() {
    this.activeModal.close(this.menu);
  }

  cancelDelete() {
    this.activeModal.dismiss('cancel');
  }

}
