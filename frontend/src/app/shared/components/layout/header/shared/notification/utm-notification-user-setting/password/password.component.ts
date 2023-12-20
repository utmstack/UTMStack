import {Component, OnInit} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {AccountService} from '../../../../../../../../core/auth/account.service';
import {User} from '../../../../../../../../core/user/user.model';
import {UtmToastService} from '../../../../../../../alert/utm-toast.service';
import {DEMO_URL} from '../../../../../../../constants/global.constant';
import {ContactUsComponent} from '../../../../../../contact-us/contact-us.component';


import {PasswordService} from './password.service';

@Component({
  selector: 'app-password',
  templateUrl: './password.component.html',
  styleUrls: ['password.component.scss']
})
export class PasswordComponent implements OnInit {
  doNotMatch: string;
  error: string;
  success: string;
  account: User;
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;

  constructor(
    private passwordService: PasswordService,
    private accountService: AccountService,
    public activeModal: NgbActiveModal,
    private utmToastService: UtmToastService,
    private ngbModalService: NgbModal) {
  }

  ngOnInit() {
    this.accountService.identity().then(account => {
      this.account = account;
    });
  }

  changePassword() {
    if (!window.location.href.includes(DEMO_URL)) {
      if (this.newPassword !== this.confirmPassword) {
        this.error = null;
        this.success = null;
        this.doNotMatch = 'ERROR';
      } else {
        this.doNotMatch = null;
        this.passwordService.save(this.newPassword, this.currentPassword).subscribe(
          () => {
            this.error = null;
            this.success = 'OK';
            this.activeModal.close();
            this.utmToastService.showSuccessBottom('Password updated successfully');
          },
          () => {
            this.success = null;
            this.error = 'ERROR';
          }
        );
      }
    } else {
      this.ngbModalService.open(ContactUsComponent, {centered: true});
    }
  }
}
