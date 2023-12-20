import {Component, OnInit} from '@angular/core';
import {AccountService} from '../../../../../core/auth/account.service';
import {LoginService} from '../../../../../core/login/login.service';
import {User} from '../../../../../core/user/user.model';
import {ADMIN_ROLE} from '../../../../constants/global.constant';
import {ActiveAdModuleActiveService} from '../../../../services/active-modules/active-ad-module.service';

@Component({
  selector: 'app-mobile-header',
  templateUrl: './mobile-header.component.html',
  styleUrls: ['./mobile-header.component.scss']
})
export class MobileHeaderComponent implements OnInit {
  isCollapsed = false;
  user: User;
  menuActive = false;
  isAdActive: boolean;
  roleAdmin = ADMIN_ROLE;


  constructor(private loginService: LoginService,
              private accountService: AccountService,
              private adModuleActiveService: ActiveAdModuleActiveService) {
  }

  ngOnInit() {
    this.accountService.identity().then(account => {
      this.user = account;
    });
  }

}
