import {Component, OnInit} from '@angular/core';
import {Observable} from 'rxjs';
import {AccountService} from '../../../../../core/auth/account.service';
import {LoginService} from '../../../../../core/login/login.service';
import {User} from '../../../../../core/user/user.model';
import {FederationModeService} from '../../../../../federation/services/federation-mode.service';
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
  federationActive$: Observable<boolean>;


  constructor(private loginService: LoginService,
              private accountService: AccountService,
              private adModuleActiveService: ActiveAdModuleActiveService,
              private federationModeService: FederationModeService) {
    this.federationActive$ = this.federationModeService.active$;
  }

  ngOnInit() {
    this.accountService.identity().then(account => {
      this.user = account;
    });
  }

}
