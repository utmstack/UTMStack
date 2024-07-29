import {Component, OnDestroy, OnInit} from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {Subject} from 'rxjs';
import {filter, takeUntil} from 'rxjs/operators';
import {AccountService} from '../../../../core/auth/account.service';
import {User} from '../../../../core/user/user.model';
import {ThemeChangeBehavior} from '../../../behaviors/theme-change.behavior';
import {ADMIN_ROLE} from '../../../constants/global.constant';
import {AppThemeLocationEnum} from '../../../enums/app-theme-location.enum';

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.scss']
})
export class HeaderComponent implements OnInit, OnDestroy {
  user: User;
  menuActive = false;
  roleAdmin = ADMIN_ROLE;
  place = AppThemeLocationEnum;
  logoImage: string;
  altImage: string;
  destroy$: Subject<void> = new Subject();

  constructor(private accountService: AccountService,
              public sanitizer: DomSanitizer,
              private themeChangeBehavior: ThemeChangeBehavior) {
  }

  ngOnInit() {
    this.themeChangeBehavior.$themeNavbarIcon
      .pipe(
        takeUntil(this.destroy$),
        filter(icon => !!icon))
          .subscribe(icon => this.logoImage = icon);

    this.accountService.identity().then(account => {
      this.user = account;
    });
  }

  ngOnDestroy() {
    this.destroy$.next();
    this.destroy$.complete();
  }

}
