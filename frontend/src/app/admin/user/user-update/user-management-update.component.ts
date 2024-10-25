import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {User} from '../../../core/user/user.model';
import {UserService} from '../../../core/user/user.service';
import {UtmToastService} from '../../../shared/alert/utm-toast.service';
import {ContactUsComponent} from '../../../shared/components/contact-us/contact-us.component';
import {DEMO_URL} from '../../../shared/constants/global.constant';


@Component({
  selector: 'app-user-mgmt-update',
  templateUrl: './user-management-update.component.html',
  styleUrls: ['./user-managment-update.component.scss']
})
export class UserMgmtUpdateComponent implements OnInit {
  @Input() user: User;
  @Output() userChange = new EventEmitter<string>();
  languages: any[];
  authorities: any[];
  isSaving: boolean;
  creating = false;

  constructor(private userService: UserService,
              public activeModal: NgbActiveModal,
              private utmToast: UtmToastService,
              private modalService: NgbModal) {
  }

  ngOnInit() {
    this.isSaving = false;
    this.authorities = [];
    this.userService.authorities().subscribe(authorities => {
      this.authorities = authorities;
    });
    if (!this.user) {
      this.user = new User();
    }
  }

  previousState() {
    window.history.back();
  }

  save() {
    if (!window.location.href.includes(DEMO_URL)) {
      this.isSaving = true;
      if (this.user.id !== null) {
        this.userService.update(this.user)
          .subscribe(response => this.onSaveSuccess(response, 'update'),
            (error) => this.onSaveError(error, 'update'));
      } else {
        this.user.langKey = 'en';
        this.userService.create(this.user)
          .subscribe(response => this.onSaveSuccess(response, 'create'),
            (error) => this.onSaveError(error, 'create'));
      }
    } else {
      this.modalService.open(ContactUsComponent, {centered: true});
    }
  }

  addRol(roleadmin: string) {
  }

  private onSaveSuccess(result, type) {
    this.isSaving = false;
    if (type === 'update') {
      this.utmToast.showSuccess('User updated successfully');
    } else {
      this.utmToast.showSuccess('User created successfully');
    }
    this.activeModal.close();
    this.userChange.emit('changed');
  }

  private onSaveError(error, type) {
    this.isSaving = false;

    if (error.status === 400) {
      this.utmToast.showError('Error', 'Admin role removal is prohibited for the last remaining administrator user.');
    } else {
      this.utmToast.showError('Problem', 'The login or email is already in use, please check');
    }
  }
}
