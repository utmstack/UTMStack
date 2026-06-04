import {Component, OnInit} from '@angular/core';
import {NgxSpinnerService} from 'ngx-spinner';
import {AccountService} from '../../../../../core/auth/account.service';
import {RestartApiBehavior} from '../../../../behaviors/restart-api.behavior';
import {AppRestartApiService} from './app-restart-api.service';
import {AppStatusType} from './app-status.type';

@Component({
  selector: 'app-app-restart-api',
  templateUrl: './app-restart-api.component.html',
  styleUrls: ['./app-restart-api.component.scss']
})
export class AppRestartApiComponent implements OnInit {
  show = false;
  status: AppStatusType;
  interval: any;

  constructor(private appRestartApiService: AppRestartApiService,
              private accountService: AccountService,
              private spinnerService: NgxSpinnerService,
              private restartApiBehavior: RestartApiBehavior) {
  }

  ngOnInit() {
    setTimeout(() => {
      if (this.accountService.isAuthenticated()) {
        this.getAppStatus();
        this.restartApiBehavior.$restartSignal.subscribe(value => {
          if (value) {
            this.getAppStatus();
          }
        });
      }
    }, 30000);
  }

  getAppStatus() {
    this.appRestartApiService.appStatus().subscribe(response => {
      this.status = response.body;
      this.show = this.status.restartRequired;
      this.spinnerService.hide('disconnectedSpinner');
      this.spinnerService.hide('restartingApi');
      clearInterval(this.interval);
    });
  }


  restartNow() {
    this.appRestartApiService.restartApi().subscribe(value => console.log('Restarting app, please wait'));
    this.spinnerService.hide('disconnectedSpinner');
    this.spinnerService.show('restartingApi');
    this.checkForRestart();
  }

  checkForRestart() {
    this.interval = setInterval(() => this.getAppStatus(), 60000);
  }
}
