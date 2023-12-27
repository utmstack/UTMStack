import {Component, OnInit} from '@angular/core';
import {NgbActiveModal, NgbModal} from '@ng-bootstrap/ng-bootstrap';
import {AccountService} from '../../../../../../../../core/auth/account.service';
import {User} from '../../../../../../../../core/user/user.model';
import {UtmToastService} from '../../../../../../../alert/utm-toast.service';
import {DEMO_URL} from '../../../../../../../constants/global.constant';
import {ContactUsComponent} from '../../../../../../contact-us/contact-us.component';


@Component({
  selector: 'app-settings',
  templateUrl: './settings.component.html',
  styleUrls: ['./settings.component.scss']
})
export class SettingsComponent implements OnInit {
  error: string;
  success: string;
  settingsAccount: User;
  languages: any[];

  constructor(private accountService: AccountService,
              public activeModal: NgbActiveModal,
              private utmToast: UtmToastService,
              private ngbModalService: NgbModal) {
  }

  ngOnInit() {
    this.accountService.identity().then(account => {
      this.settingsAccount = this.copyAccount(account);
    });
  }

  save() {
    if (!window.location.href.includes(DEMO_URL)) {
      this.accountService.save(this.settingsAccount).subscribe(
        () => {
          this.error = null;
          this.success = 'OK';
          this.accountService.identity(true).then(account => {
            this.settingsAccount = this.copyAccount(account);
            this.utmToast.showSuccessBottom('Account updated successfully');
            this.activeModal.close();
          });
        },
        () => {
          this.success = null;
          this.error = 'ERROR';
          this.utmToast.showError('Problem', 'Problem updating your account, please try again');
        }
      );
    } else {
      this.ngbModalService.open(ContactUsComponent, {centered: true});
    }
  }

  copyAccount(account) {
    return {
      activated: account.activated,
      email: account.email,
      firstName: account.firstName,
      langKey: account.langKey,
      lastName: account.lastName,
      login: account.login,
      imageUrl: account.imageUrl
    };
  }
}
