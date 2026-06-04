import { Component, OnInit } from '@angular/core';
import {DomSanitizer} from '@angular/platform-browser';
import {AccountService} from '../../../core/auth/account.service';
import {ThemeChangeBehavior} from '../../behaviors/theme-change.behavior';

@Component({
  selector: 'app-utm-lite-version',
  templateUrl: './utm-lite-version.component.html',
  styleUrls: ['./utm-lite-version.component.scss']
})
export class UtmLiteVersionComponent implements OnInit {
  isLogged: boolean;
  logo: string;

  constructor(private accountService: AccountService,
              public sanitizer: DomSanitizer,
              private themeChangeBehavior: ThemeChangeBehavior) {
    accountService.identity().then(account => {
      this.isLogged = account !== null;
    });
  }

  ngOnInit() {
    this.themeChangeBehavior.$themeIcon.subscribe(icon => {
      if (icon) {
        this.logo = icon;
      }
    });
  }

}
