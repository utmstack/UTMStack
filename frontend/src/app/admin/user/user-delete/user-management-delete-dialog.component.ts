import {Component} from '@angular/core';
import {NgbActiveModal} from '@ng-bootstrap/ng-bootstrap';
import {JhiEventManager} from 'ng-jhipster';
import {User} from '../../../core/user/user.model';
import {UserService} from '../../../core/user/user.service';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';

@Component({
  selector: 'app-ser-mgmt-delete-dialog',
  templateUrl: './user-management-delete-dialog.component.html'
})
export class UserMgmtDeleteDialogComponent {
  user: User;

  constructor(private userService: UserService,
              public activeModal: NgbActiveModal,
              private eventManager: JhiEventManager,
              private utmToast: UtmToastService) {
  }

  clear() {
    this.activeModal.dismiss('cancel');
  }

  confirmDelete(login) {
    this.userService.delete(login).subscribe(response => {
      this.eventManager.broadcast({
        name: 'userListModification',
        content: 'Deleted a user'
      });
      this.utmToast.showSuccess('User deleted successfully');
      this.activeModal.dismiss(true);
    });
  }
}
